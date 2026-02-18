package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

func New() *logrus.Logger {
	log := logrus.New()

	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	log.SetOutput(os.Stdout)
	log.SetLevel(logrus.InfoLevel)

	return log
}
