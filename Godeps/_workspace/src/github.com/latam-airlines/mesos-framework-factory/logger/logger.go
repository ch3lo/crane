package logger

import log "github.com/Sirupsen/logrus"

var logger *log.Logger

func Setup(l *log.Logger) {
	logger = l
}

func Instance() *log.Logger {
	if logger == nil {
		logger = log.New()
	}

	return logger
}
