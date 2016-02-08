package logger

import (
	"errors"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/latam-airlines/crane/configuration"
)

var logger *log.Logger

func Configure(config configuration.Loggging, debug bool) error {
	if logger == nil {
		logger = log.New()
	}

	var err error

	if logger.Level, err = log.ParseLevel(config.Level); err != nil {
		return err
	}

	if debug {
		logger.Level = log.DebugLevel
	}

	switch config.Formatter {
	case "text":
		formatter := new(log.TextFormatter)
		formatter.ForceColors = config.Colored
		formatter.FullTimestamp = true
		logger.Formatter = formatter
		break
	case "json":
		formatter := new(log.JSONFormatter)
		logger.Formatter = formatter
		break
	default:
		return errors.New("Formato de lo log desconocido")
	}

	switch config.Output {
	case "console":
		logger.Out = os.Stdout
		break
	case "file":
		logFile, err := os.OpenFile("crane.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			logger.Warnln("Error al abrir el archivo")
		}
		logger.Out = logFile
		break
	default:
		return errors.New("Output de logs desconocido")
	}

	return nil
}

func Instance() *log.Logger {
	return logger
}
