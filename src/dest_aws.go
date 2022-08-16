package main

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type BackupDestinationAws struct {
	target      *BackupTarget
	Credentials string `mapstructure:"credentials"`
	Config      string `mapstructure:"config"`
	Folder      string `mapstructure:"folder"`
	Bucket      string `mapstructure:"bucket"`
}

func NewBackupDestinationAws() *BackupDestinationAws {
	return &BackupDestinationAws{}
}

func parseConfigAwsDestination(unmarshalKey string, t *BackupTarget) {
	awsDest := NewBackupDestinationAws()
	handleFatalErr(
		viper.UnmarshalKey(unmarshalKey, &awsDest),
		"Cannot parse aws backup destination %s\n",
		unmarshalKey)
	if len(awsDest.Credentials) > 0 {
		awsDest.Credentials = parseTilde(awsDest.Credentials)
	}
	if len(awsDest.Config) > 0 {
		awsDest.Config = parseTilde(awsDest.Config)
	}
	t.DestinationConfig = append(t.DestinationConfig, awsDest)
}

func (d *BackupDestinationAws) runBackup(t *BackupTarget) {
	var sharedCredentialsFiles []string
	var sharedConfigFiles []string

	if len(d.Credentials) > 0 {
		sharedCredentialsFiles = []string{d.Credentials}
	}
	if len(d.Config) > 0 {
		sharedConfigFiles = []string{d.Config}
	}

	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithSharedCredentialsFiles(sharedCredentialsFiles),
		config.WithSharedConfigFiles(sharedConfigFiles),
	)
	if handleErr(err) {
		return
	}

	client := s3.NewFromConfig(cfg)
	log.Infof("%s Upload an object to the bucket '%s'\n", getDestLogPrefix(d), d.Bucket)
	stat, err := os.Stat(t.Archive)
	if handleErr(err) {
		return
	}
	file, err := os.Open(t.Archive)
	if handleErr(err) {
		return
	}
	objectKey := filepath.Base(t.Archive)
	if len(d.Folder) > 0 {
		objectKey = filepath.Join(d.Folder, objectKey)
	}
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:        aws.String(d.Bucket),
		Key:           aws.String(objectKey),
		Body:          file,
		ContentLength: stat.Size(),
	})
	if handleErr(file.Close()) {
		return
	}
	handleErr(err)
	log.Infof("%s Backup uploaded to bucket %s\n", getDestLogPrefix(d), d.Bucket)
}

func (d *BackupDestinationAws) cleanOldBackups(t *BackupTarget) {
	log.Infof("%s Clean old backups", getDestLogPrefix(d))
}

func (d *BackupDestinationAws) getName() string {
	return "aws"
}

func (d *BackupDestinationAws) getTarget() *BackupTarget {
	return d.target
}

func (d *BackupDestinationAws) setTarget(ptr *BackupTarget) {
	d.target = ptr
}
