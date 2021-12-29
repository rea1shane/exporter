package util

import (
	formatter "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func GetLogger() *log.Logger {
	logger := log.New()
	logger.SetFormatter(&formatter.Formatter{
		TimestampFormat: "2006-01-02 | 15:04:05",
		FieldsOrder:     []string{"name", "duration_seconds"},
		HideKeys:        true,
	})
	return logger
}
