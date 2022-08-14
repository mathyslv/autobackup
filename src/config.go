package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func parseConfigDestinations(key string, t *BackupTarget) {
	for _, destination := range t.Config.Destinations {
		unmarshalKey := fmt.Sprintf("%s.%s", key, destination)
		switch destination {
		case "local":
			var local BackupDestinationLocal
			handleFatalErr(
				viper.UnmarshalKey(unmarshalKey, &local),
				"Cannot parse backup destination %s\n",
				unmarshalKey)
			local.Directory = parseTilde(local.Directory)
			t.DestinationConfig = append(t.DestinationConfig, local)
		case "aws":
			var aws BackupDestinationAws
			handleFatalErr(
				viper.UnmarshalKey(unmarshalKey, &aws),
				"Cannot parse backup destination %s\n",
				unmarshalKey)
			t.DestinationConfig = append(t.DestinationConfig, aws)
		default:
			log.Warnf("[%s] Unknown backup destination '%s'\n", t.Name, destination)
		}
	}
}

func parseConfig() []BackupTarget {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
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
		log.Infof("path = %s\n", backupTarget.Config.Path)
		backupTarget.Config.Path = parseTilde(backupTarget.Config.Path)
		backupTarget.Name = key
		parseConfigDestinations(key, &backupTarget)
		backupTargets = append(backupTargets, backupTarget)
	}
	return backupTargets
}
