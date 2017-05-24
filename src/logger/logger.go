package logger

import (
	"log"
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	once   sync.Once
	logger *logrus.Logger
)

type nullFormatter struct {
}

func initLogger() {
	once.Do(func() {
		// Logging setup
		logger = logrus.New()

		log.SetOutput(logger.Writer())
	})
}

func Instance() *logrus.Logger {
	initLogger()

	return logger
}

// Don't pass logs to stdout
func (nullFormatter) Format(e *logrus.Entry) ([]byte, error) {
	return []byte{}, nil
}
