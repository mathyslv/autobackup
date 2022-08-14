package main

import (
	"archive/tar"
	"compress/gzip"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func parseTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		dirname, _ := os.UserHomeDir()
		path = filepath.Join(dirname, path[2:])
	}
	return path
}

func addFileToArchive(f string, t *BackupTarget, tw *tar.Writer) {
	fileHandle, err := os.Open(f)
	handleFatalErr(err, "Cannot open file")
	info, err := fileHandle.Stat()
	handleFatalErr(err, "Cannot stat file")
	header, err := tar.FileInfoHeader(info, info.Name())
	if t.Config.PreserveAbsoluteHierarchy {
		header.Name = f
	} else {
		header.Name = strings.ReplaceAll(f, t.Config.Path, "")
		if header.Name[0] == '/' {
			header.Name = header.Name[1:]
		}
	}
	handleFatalErr(tw.WriteHeader(header))
	_, err = io.Copy(tw, fileHandle)
	handleFatalErr(err, "Cannot copy content from file")
	handleFatalErr(fileHandle.Close(), "Cannot close file")
	log.Debugf("Adding file %s to archive\n", header.Name)
}

func buildArchive(t *BackupTarget) {
	t.Archive = filepath.Join(t.TmpWorkdir, t.Name+".tar.gz")
	fileWriter, err := os.Create(t.Archive)
	handleFatalErr(err, "Error when creating file")
	defer func() { handleFatalErr(fileWriter.Close(), "Error when closing file writer") }()

	gzipWriter := gzip.NewWriter(fileWriter)
	defer func() { handleFatalErr(gzipWriter.Close(), "Error when closing gzip writer") }()

	tarWriter := tar.NewWriter(gzipWriter)
	defer func() { handleFatalErr(tarWriter.Close(), "Error when closing tar writer") }()

	for _, file := range t.Files {
		addFileToArchive(file, t, tarWriter)
	}

}

func processBackupTarget(t *BackupTarget) {
	createBackupTargetTempWorkdir(t)
	t.Files = listBackupTargetFiles(t)
	buildArchive(t)
	for _, destination := range t.DestinationConfig {
		destination.runBackup(t)
	}
}

func launchBackupTargetCron(c *cron.Cron, t *BackupTarget) cron.EntryID {
	/*entryId, err := c.AddFunc(t.Config.Cron, func() {
		processBackupTarget(t)
	})*/
	processBackupTarget(t)
	//handleFatalErr(err, "Cannot create cron job : %s\n", err)
	log.Debugf("[%s] Cron : '%s' created\n", t.Name, t.Config.Cron)
	//return entryId
	return 0
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

func createBackupTargetTempWorkdir(target *BackupTarget) {
	dir, err := ioutil.TempDir(os.TempDir(), "autobackup_"+target.Name+"_")
	handleFatalErr(err, "Cannot create temporary working directory")
	log.Debugf("[%s] Created temporary working directory %s\n", target.Name, dir)
	target.TmpWorkdir = dir
}

func deleteBackupTargetTempWorkdir(target *BackupTarget) {
	err := os.RemoveAll(target.TmpWorkdir)

	if err != nil {
		handleFatalErr(err, "[%s] Error when deleting temporary working directory", target.Name)
	} else {
		log.Debugf("[%s] Deleted temporary working directory\n", target.TmpWorkdir)
	}
}

func main() {

	log.SetLevel(log.DebugLevel)

	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("$HOME/.config/autobackup")
	viper.AddConfigPath(".")

	backupTargets := parseConfig()

	cronRunner := cron.New()

	defer func() {
		for _, target := range backupTargets {
			deleteBackupTargetTempWorkdir(&target)
		}
	}()

	for _, backupTarget := range backupTargets {
		log.Infof("Processing backup target '%s'\n", backupTarget.Name)
		for _, destination := range backupTarget.Config.Destinations {
			if !stringInSlice(destination, DestinationList) {

			}
		}
		launchBackupTargetCron(cronRunner, &backupTarget)
	}

	cronRunner.Start()
	time.Sleep(5 * time.Minute)
}
