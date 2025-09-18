package log

import (
	"github.com/sirupsen/logrus"
)

var (
	logLevels = map[string]logrus.Level{
		"DEBUG": logrus.DebugLevel,
		"INFO":  logrus.InfoLevel,
		"WARN":  logrus.WarnLevel,
		"ERROR": logrus.ErrorLevel,
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
