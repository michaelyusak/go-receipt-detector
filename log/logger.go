package log

import (
	"github.com/sirupsen/logrus"
)

var (
	logLevels = map[string]logrus.Level{
		"DEBUG": logrus.DebugLevel,
	}
)

func Init(logLevel string) {
	logrusLogLevel, ok := logLevels[logLevel]
	if !ok {
		panic("[log][logger][Init] invalid log level")
	}

	logrus.SetLevel(logrusLogLevel)

	if logrusLogLevel == logrus.DebugLevel {
		logrus.SetReportCaller(true)
	}
}
