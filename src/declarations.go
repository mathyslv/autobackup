package main

import (
	"context"
	"time"
)

type BackupTargetConfig struct {
	Type                      string   `mapstructure:"type"`
	Path                      string   `mapstructure:"path"`
	Format                    string   `mapstructure:"format"`
	Frequency                 string   `mapstructure:"frequency"`
	Cron                      string   `mapstructure:"cron"`
	KeepOnly                  int      `mapstructure:"keep_only"`
	Replace                   bool     `mapstructure:"replace"`
	DateSuffix                bool     `mapstructure:"date_suffix"`
	ExcludeVcs                bool     `mapstructure:"exclude_vcs"`
	PreserveAbsoluteHierarchy bool     `mapstructure:"preserve_absolute_hierarchy"`
	Destinations              []string `mapstructure:"destinations"`
	ExcludeDirs               []string `mapstructure:"exclude_dirs"`
}

type BackupItem struct {
	Name string
	Date time.Time
}

type BackupDestination interface {
	init() bool
	isReady() bool
	runBackup()
	buildBackupsList(context.Context) ([]BackupItem, error)
	cleanOldBackups()
	setTarget(*BackupTarget)
	getName() string
	getTarget() *BackupTarget
}

type BackupTarget struct {
	Name              string
	TmpWorkdir        string
	Archive           string
	Ext               string
	Files             []string
	Config            BackupTargetConfig
	DestinationConfig []BackupDestination
}

var parseConfigFnMap = map[string]func(string, *BackupTarget){
	"local": parseConfigLocalDestination,
	"aws":   parseConfigAwsDestination,
	"gcp":   parseConfigGcpDestination,
}
