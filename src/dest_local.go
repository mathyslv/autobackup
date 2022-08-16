package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

type BackupDestinationLocal struct {
	target    *BackupTarget
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

func (d *BackupDestinationLocal) runBackup(t *BackupTarget) {
	localDestPath := d.Directory + "/" + filepath.Base(t.Archive)
	if handleErr(os.MkdirAll(d.Directory, os.ModePerm)) {
		return
	}
	_, err := copyFile(t.Archive, localDestPath)
	if !handleErr(err) {
		log.Infoln(getDestLogPrefix(d), "Backup saved")
	}
}

func (d *BackupDestinationLocal) cleanOldBackups(t *BackupTarget) {
	var files []os.FileInfo
	handleErr(filepath.Walk(d.Directory, func(path string, info os.FileInfo, err error) error {
		if handleErr(err) {
			return err
		}
		if info.IsDir() {
			return nil
		}
		match, _ := regexp.Match("gitlab-mathyslv_[0-9]{8}_[0-9]{6}.tar.gz", []byte(info.Name()))
		if match {
			files = append(files, info)
		}
		return nil
	}))
	if len(files) > 0 {
		if t.Config.KeepOnly >= len(files) {
			return
		}
		sort.SliceStable(files, func(i, j int) bool {
			return files[i].ModTime().Before(files[j].ModTime())
		})
	}
	for i := 0; i < len(files)-t.Config.KeepOnly; i++ {
		handleErr(os.Remove(filepath.Join(d.Directory, files[i].Name())))
		log.Debugf("%s Removed old backup file '%s'\n", getDestLogPrefix(d), filepath.Join(d.Directory, files[i].Name()))
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
