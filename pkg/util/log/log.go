package log

import "github.com/sirupsen/logrus"

var (
	logger = logrus.New()
)

func init() {
	logger.SetLevel(logrus.DebugLevel)
}

func GetLogger() *logrus.Logger {
	return logger
}
