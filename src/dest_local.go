package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

type BackupDestinationLocal struct {
	target    *BackupTarget
	ready     bool
	Directory string `mapstructure:"directory"`
}

func NewBackupDestinationLocal() *BackupDestinationLocal {
	return &BackupDestinationLocal{}
}

func parseConfigLocalDestination(unmarshalKey string, t *BackupTarget) {
	localDest := NewBackupDestinationLocal()
	handleFatalErr(
		viper.UnmarshalKey(unmarshalKey, &localDest),
		"Cannot parse local backup destination %s\n",
		unmarshalKey)
	localDest.Directory = parseTilde(localDest.Directory)
	t.DestinationConfig = append(t.DestinationConfig, localDest)
}

func (d *BackupDestinationLocal) init() bool {
	d.ready = true
	return d.ready
}

func (d *BackupDestinationLocal) runBackup() {
	localDestPath := d.Directory + "/" + filepath.Base(d.target.Archive)
	if handleErr(os.MkdirAll(d.Directory, os.ModePerm), getDestLogPrefix(d)) {
		return
	}
	_, err := copyFile(d.target.Archive, localDestPath)
	if !handleErr(err, getDestLogPrefix(d)) {
		log.Infoln(getDestLogPrefix(d), "Backup saved")
	}
}

func (d *BackupDestinationLocal) buildBackupsList(_ context.Context) ([]BackupItem, error) {
	var backupItems []BackupItem

	err := filepath.Walk(d.Directory, func(path string, info os.FileInfo, err error) error {
		if handleErr(err, getDestLogPrefix(d)) {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if doesBackupNameMatch(d.target, info.Name()) {
			backupItems = append(backupItems, BackupItem{
				Name: path,
				Date: info.ModTime(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return backupItems, nil
}

func (d *BackupDestinationLocal) cleanOldBackups() {
	backupItems, err := d.buildBackupsList(context.TODO())

	if handleErr(err, getDestLogPrefix(d)) {
		return
	}
	sortBackups(d.target, backupItems)
	for i := 0; i < len(backupItems)-d.target.Config.KeepOnly; i++ {
		handleErr(os.Remove(backupItems[i].Name), getDestLogPrefix(d))
		log.Debugf("%s Removed old backup file '%s'\n", getDestLogPrefix(d), backupItems[i].Name)
	}
	log.Infoln(getDestLogPrefix(d), "Cleaned old backups")
}

func (d *BackupDestinationLocal) getName() string {
	return "local"
}

func (d *BackupDestinationLocal) setTarget(ptr *BackupTarget) {
	d.target = ptr
}

func (d *BackupDestinationLocal) getTarget() *BackupTarget {
	return d.target
}

func (d *BackupDestinationLocal) isReady() bool {
	return d.ready
}
