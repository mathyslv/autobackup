package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func _handleErrLevel(fn func(string, ...interface{}), err error, args ...interface{}) bool {
	if err != nil {
		if len(args) > 0 {
			format := args[0].(string)
			fn("%s: %s\n", fmt.Sprintf(format, args[1:]...), err.Error())
		} else {
			fn("Error: %s\n", err.Error())
		}
	}
	return err != nil
}

func handleErr(err error, args ...interface{}) bool {
	return _handleErrLevel(log.Errorf, err, args...)
}

func handleFatalErr(err error, args ...interface{}) bool {
	return _handleErrLevel(log.Fatalf, err, args...)
}

func handleWarnErr(err error, args ...interface{}) bool {
	return _handleErrLevel(log.Warnf, err, args...)
}

func handleInfoErr(err error, args ...interface{}) bool {
	return _handleErrLevel(log.Infof, err, args...)
}
