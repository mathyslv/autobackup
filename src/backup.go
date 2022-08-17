package main

import (
	"regexp"
	"sort"
)

func sortBackups(t *BackupTarget, backups []BackupItem) {
	if len(backups) > 0 {
		if t.Config.KeepOnly >= len(backups) {
			return
		}
		sort.SliceStable(backups, func(i, j int) bool {
			return backups[i].Date.Before(backups[j].Date)
		})
	}
}

func doesBackupNameMatch(t *BackupTarget, backupName string) bool {
	reg := t.Name
	if t.Config.DateSuffix {
		reg += "_[0-9]{8}_[0-9]{6}"
	}
	reg += t.Ext
	match, _ := regexp.Match(reg, []byte(backupName))
	return match
}
