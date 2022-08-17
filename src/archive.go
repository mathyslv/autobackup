package main

import (
	"archive/tar"
	"compress/gzip"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

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
	//log.Debugf("Adding file %s to archive\n", header.Name)
}

func getArchiveExt(t *BackupTarget) string {
	log.Warnln(t.Config.Format)
	switch t.Config.Format {
	case "tar.gz", "compressed":
		return ".tar.gz"
	default:
		return ".unknown"
	}
}

func buildArchive(t *BackupTarget) {
	archiveBasepath := filepath.Join(t.TmpWorkdir, t.Name)
	if t.Config.DateSuffix {
		archiveBasepath += "_" + time.Now().Format("02012006_150405")
	}
	t.Archive = archiveBasepath + t.Ext
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
	log.Infof("[%s] Created archive '%s'\n", t.Name, filepath.Base(t.Archive))
}
