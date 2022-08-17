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
)

type BackupDestinationGcp struct {
	target       *BackupTarget
	ready        bool
	client       *storage.Client
	bucketHandle *storage.BucketHandle
	Credentials  string `mapstructure:"credentials"`
	Folder       string `mapstructure:"folder"`
	Bucket       string `mapstructure:"bucket"`
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

func (d *BackupDestinationGcp) init() bool {
	return false

	client, err := storage.NewClient(context.TODO(), option.WithCredentialsFile(d.Credentials))
	if handleErr(err) {
		d.ready = false
		return false
	}
	d.bucketHandle = client.Bucket(d.Bucket)
	d.ready = true
	return d.ready
}

func (d *BackupDestinationGcp) runBackup() {
	client, err := storage.NewClient(context.TODO(), option.WithCredentialsFile(d.Credentials))
	if handleErr(err) {
		return
	}
	bucketHandle := client.Bucket(d.Bucket)
	objectName := filepath.Base(d.target.Archive)
	if len(d.Folder) > 0 {
		objectName = filepath.Join(d.Folder, objectName)
	}
	obj := bucketHandle.Object(objectName)
	bucketWriter := obj.NewWriter(context.TODO())
	archiveHandle, err := os.Open(d.target.Archive)
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

func (d *BackupDestinationGcp) buildBackupsList(ctx context.Context) ([]BackupItem, error) {
	var backupItems []BackupItem

	query := &storage.Query{}
	err := query.SetAttrSelection([]string{"Name"})
	if err != nil {
		return nil, err
	}

	objectIterator := d.bucketHandle.Objects(ctx, query)
	for {
		objectAttrs, err := objectIterator.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, err
		}
		if doesBackupNameMatch(d.target, objectAttrs.Name) {
			backupItems = append(backupItems, BackupItem{
				Name: objectAttrs.Name,
				Date: objectAttrs.Created,
			})
		}
	}
	return backupItems, nil
}

func (d *BackupDestinationGcp) cleanOldBackups() {
	backupItems, err := d.buildBackupsList(context.TODO())

	if handleErr(err) {
		return
	}
	sortBackups(d.target, backupItems)

	for i := 0; i < len(backupItems)-d.target.Config.KeepOnly; i++ {
		objectHandle := d.bucketHandle.Object(backupItems[i].Name)
		if handleErr(objectHandle.Delete(context.TODO())) {
			return
		}
		log.Debugf("%s Removed old backup object '%s'\n", getDestLogPrefix(d), backupItems[i].Name)
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

func (d *BackupDestinationGcp) isReady() bool {
	return d.ready
}
