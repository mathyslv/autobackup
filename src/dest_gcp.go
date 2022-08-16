package main

import (
	"cloud.google.com/go/storage"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type BackupDestinationGcp struct {
	target      *BackupTarget
	Credentials string `mapstructure:"credentials"`
	Folder      string `mapstructure:"folder"`
	Bucket      string `mapstructure:"bucket"`
}

func NewBackupDestinationGcp() *BackupDestinationGcp {
	return &BackupDestinationGcp{}
}

func parseConfigGcpDestination(unmarshalKey string, t *BackupTarget) {
	gcpDest := NewBackupDestinationGcp()
	handleFatalErr(
		viper.UnmarshalKey(unmarshalKey, &gcpDest),
		"Cannot parse gcp backup destination %s\n",
		unmarshalKey)
	if len(gcpDest.Credentials) > 0 {
		gcpDest.Credentials = parseTilde(gcpDest.Credentials)
	}
	t.DestinationConfig = append(t.DestinationConfig, gcpDest)
}

func (d *BackupDestinationGcp) runBackup(t *BackupTarget) {
	client, err := storage.NewClient(context.TODO(), option.WithCredentialsFile(d.Credentials))
	if handleErr(err) {
		return
	}
	bucketHandle := client.Bucket(d.Bucket)
	objectName := filepath.Base(t.Archive)
	if len(d.Folder) > 0 {
		objectName = filepath.Join(d.Folder, objectName)
	}
	obj := bucketHandle.Object(objectName)
	bucketWriter := obj.NewWriter(context.TODO())
	archiveHandle, err := os.Open(t.Archive)
	defer func() {
		handleErr(archiveHandle.Close())
	}()
	if handleErr(err) {
		return
	}
	_, err = io.Copy(bucketWriter, archiveHandle)
	if handleErr(err) {
		return
	}
	if handleErr(bucketWriter.Close()) {
		return
	}
	log.Infof("%s Backup uploaded to bucket %s\n", getDestLogPrefix(d), d.Bucket)
}

func (d *BackupDestinationGcp) cleanOldBackups(t *BackupTarget) {

	client, err := storage.NewClient(context.TODO(), option.WithCredentialsFile(d.Credentials))
	if handleErr(err) {
		return
	}
	bucketHandle := client.Bucket(d.Bucket)
	query := &storage.Query{}
	if handleErr(query.SetAttrSelection([]string{"Name"})) {
		return
	}
	objectIterator := bucketHandle.Objects(context.TODO(), query)

	var objectsAttrs []*storage.ObjectAttrs
	for {
		objectAttrs, err := objectIterator.Next()
		if err == iterator.Done {
			break
		} else if handleErr(err) {
			return
		}
		match, _ := regexp.Match("gitlab-mathyslv_[0-9]{8}_[0-9]{6}.tar.gz", []byte(objectAttrs.Name))
		if match {
			objectsAttrs = append(objectsAttrs, objectAttrs)
		}
	}

	if len(objectsAttrs) > 0 {
		if t.Config.KeepOnly >= len(objectsAttrs) {
			return
		}
		sort.SliceStable(objectsAttrs, func(i, j int) bool {
			return objectsAttrs[i].Created.Before(objectsAttrs[j].Created)
		})
	}

	for i := 0; i < len(objectsAttrs)-t.Config.KeepOnly; i++ {
		objectHandle := bucketHandle.Object(objectsAttrs[i].Name)
		if handleErr(objectHandle.Delete(context.TODO())) {
			return
		}
		log.Debugf("%s Removed old backup object '%s'\n", getDestLogPrefix(d), objectsAttrs[i].Name)
	}
	log.Infoln(getDestLogPrefix(d), "Cleaned old backups")
}

func (d *BackupDestinationGcp) getName() string {
	return "gcp"
}

func (d *BackupDestinationGcp) getTarget() *BackupTarget {
	return d.target
}

func (d *BackupDestinationGcp) setTarget(ptr *BackupTarget) {
	d.target = ptr
}
