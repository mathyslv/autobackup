package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func parseConfigLocalDestination(unmarshalKey string, t *BackupTarget) {
	localDest := NewBackupDestinationLocal()
	handleFatalErr(
		viper.UnmarshalKey(unmarshalKey, &localDest),
		"Cannot parse local backup destination %s\n",
		unmarshalKey)
	localDest.Directory = parseTilde(localDest.Directory)
	t.DestinationConfig = append(t.DestinationConfig, localDest)
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

func parseConfigDestinations(key string, t *BackupTarget) {
	for _, destination := range t.Config.Destinations {
		unmarshalKey := fmt.Sprintf("%s.%s", key, destination)
		switch destination {
		case "local":
			parseConfigLocalDestination(unmarshalKey, t)
		case "aws":
			parseConfigAwsDestination(unmarshalKey, t)
		default:
			log.Warnf("[%s] Unknown backup destination '%s'\n", t.Name, destination)
			continue
		}
		t.DestinationConfig[len(t.DestinationConfig)-1].setTarget(t)
	}
}

func parseConfig() []BackupTarget {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Config file '%s' not found\n", viper.ConfigFileUsed())
		} else {
			log.Fatalf("Config file was found but another error was produced : %s\n", err.Error())
		}
	}
	log.Infof("Configuration file : '%s'\n", viper.ConfigFileUsed())
	autobackupConfig := viper.AllSettings()
	var backupTargets []BackupTarget
	for key, _ := range autobackupConfig {
		var backupTarget BackupTarget
		handleFatalErr(viper.UnmarshalKey(key, &backupTarget.Config), "unable to decode into struct\n")
		backupTarget.Config.Path = parseTilde(backupTarget.Config.Path)
		backupTarget.Name = key
		parseConfigDestinations(key, &backupTarget)
		backupTargets = append(backupTargets, backupTarget)
	}
	return backupTargets
}
