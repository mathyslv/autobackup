package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

type BackupDestinationLocal struct {
	Directory string `mapstructure:"directory"`
	Name      string
}

func (d BackupDestinationLocal) runBackup(target *BackupTarget) {
	localDestPath := d.Directory + "/" + filepath.Base(target.Archive)
	if handleErr(os.MkdirAll(d.Directory, os.ModePerm)) {
		return
	}
	_, err := copyFile(target.Archive, localDestPath)
	if !handleErr(err) {
		log.Infof("[%s][local][trigger] Save backup to %s\n", target.Name, localDestPath)
	}
}

/* func launchLocalBackup(cronRunner *cron.Cron, target *BackupTarget) cron.EntryID {
	log.Debugf("[%s][local] Directory : '%s'\n", target.Name, target.DestinationConfig.Local.Directory)
	entryId, err := cronRunner.AddFunc(target.Config.Cron, func() { runLocalBackup(target) })
	if err != nil {
		log.Fatalf("Cannot create cron job : %s\n", err.Error())
	}
	log.Debugf("[%s][local] Cron : '%s' created\n", target.Name, target.Config.Cron)
	return entryId
} */

func (d BackupDestinationLocal) getName() string {
	return "aws"
}
