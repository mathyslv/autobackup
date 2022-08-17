package main

import (
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createBackupTargetTempWorkdir(target *BackupTarget) {
	dir, err := ioutil.TempDir(os.TempDir(), "autobackup_"+target.Name+"_")
	handleFatalErr(err, "Cannot create temporary working directory")
	log.Debugf("[%s] Created temporary working directory %s\n", target.Name, dir)
	target.TmpWorkdir = dir
}

func deleteBackupTargetTempWorkdir(t *BackupTarget) {
	err := os.RemoveAll(t.TmpWorkdir)
	if err != nil {
		handleFatalErr(err, "[%s] Error when deleting temporary working directory", t.Name)
	} else {
		log.Debugf("[%s] Deleted temporary working directory\n", t.Name)
	}
}

func processBackupTarget(t *BackupTarget) {
	createBackupTargetTempWorkdir(t)
	defer deleteBackupTargetTempWorkdir(t)
	t.Files = listBackupTargetFiles(t)
	buildArchive(t)
	for _, d := range t.DestinationConfig {
		d.runBackup()
		if t.Config.KeepOnly > 0 {
			d.cleanOldBackups()
		}
	}
}

func launchBackupTargetCron(c *cron.Cron, t *BackupTarget) cron.EntryID {
	entryId, err := c.AddFunc(t.Config.Cron, func() {
		processBackupTarget(t)
	})
	//processBackupTarget(t)
	handleFatalErr(err, "Cannot create cron job : %s\n", err)
	log.Infof("[%s] Backup target successfully configured", t.Name)
	return entryId
}

func listBackupTargetFiles(target *BackupTarget) []string {
	var files []string

	err := filepath.Walk(target.Config.Path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() ||
				info.Name() == ".git" ||
				strings.Contains(path, ".git/") {
				return nil
			}
			if len(target.Config.ExcludeDirs) > 0 {
				for _, excludedDir := range target.Config.ExcludeDirs {
					if len(excludedDir) == 0 {
						continue
					}
					if excludedDir[len(excludedDir)-1] != '/' {
						excludedDir += "/"
					}
					if strings.Contains(path, excludedDir) {
						return nil
					}
				}
			}
			files = append(files, path)
			return nil
		})
	handleFatalErr(err, "Cannot iterate through files at %s", target.Config.Path)
	return files
}

func main() {

	log.SetLevel(log.DebugLevel)
	log.SetReportCaller(true)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/autobackup")
	viper.AddConfigPath(".")

	backupTargets := parseConfig()

	cronRunner := cron.New()

	for _, backupTarget := range backupTargets {
		log.Infof("Processing backup target '%s'\n", backupTarget.Name)

		backupTarget.Ext = getArchiveExt(backupTarget)

		//nextTime := cronexpr.MustParse(backupTarget.Config.Cron).Next(time.Now())
		//log.Infof("[%s] Next tick of %s in %dh%d (%s)", backupTarget.Name, backupTarget.Config.Cron, int(nextTime.Sub(time.Now()).Hours()), int(nextTime.Sub(time.Now()).Minutes())%60, nextTime.Format("15:04 02/01/2006"))

		var validIndex int
		for _, d := range backupTarget.DestinationConfig {
			if d.init() {
				backupTarget.DestinationConfig[validIndex] = d
				validIndex++
			} else {
				log.Warnf("%s Destination removed because initialization failed\n", getDestLogPrefix(d))
			}
		}
		for invalidIndex := validIndex; invalidIndex < len(backupTarget.DestinationConfig); invalidIndex++ {
			backupTarget.DestinationConfig[invalidIndex] = nil
		}
		backupTarget.DestinationConfig = backupTarget.DestinationConfig[:validIndex]
		launchBackupTargetCron(cronRunner, backupTarget)
	}

	cronRunner.Start()
	log.Infoln("Ready")
	time.Sleep(time.Duration(1<<63 - 1))
}
