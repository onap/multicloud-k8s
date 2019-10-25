package logutils

import (
	log "github.com/sirupsen/logrus"
)

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.JSONFormatter{})
}

func WithFields(msg string, fkey string, fvalue string) {
	log.WithFields(log.Fields{fkey: fvalue}).Error(msg)
}

