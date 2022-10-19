package log

import "github.com/sirupsen/logrus"

func GetLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	return logger
}
