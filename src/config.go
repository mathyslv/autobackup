package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func parseConfigDestinations(key string, t *BackupTarget) {
	for _, destination := range t.Config.Destinations {
		if parseConfigFn, ok := parseConfigFnMap[destination]; ok {
			parseConfigFn(key+"."+destination, t)
		} else {
			log.Warnf("[%s] Unknown backup destination '%s'\n", t.Name, destination)
			continue
		}
		t.DestinationConfig[len(t.DestinationConfig)-1].setTarget(t)
	}
}

func parseConfig() []*BackupTarget {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatalf("Config file '%s' not found\n", viper.ConfigFileUsed())
		} else {
			log.Fatalf("Config file was found but another error was produced : %s\n", err.Error())
		}
	}
	log.Infof("Configuration file : '%s'\n", viper.ConfigFileUsed())
	autobackupConfig := viper.AllSettings()

	var backupTargets []*BackupTarget
	for key, _ := range autobackupConfig {
		backupTarget := new(BackupTarget)
		handleFatalErr(viper.UnmarshalKey(key, &backupTarget.Config), "unable to decode into struct\n")
		backupTarget.Config.Path = parseTilde(backupTarget.Config.Path)
		backupTarget.Name = key
		parseConfigDestinations(key, backupTarget)
		backupTargets = append(backupTargets, backupTarget)
	}
	return backupTargets
}
