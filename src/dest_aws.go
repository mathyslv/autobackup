package main

import (
	log "github.com/sirupsen/logrus"
)

type BackupDestinationAws struct {
	Credentials string
	Bucket      string
}

func (d BackupDestinationAws) runBackup(target *BackupTarget) {
	log.Debugf("[%s][aws][trigger] Push backup to bucket %s\n", target.Name, d.Bucket)
}

func (d BackupDestinationAws) getName() string {
	return "aws"
}
