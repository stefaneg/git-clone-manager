package logger

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

type CustomFormatter struct {
	logrus.TextFormatter
}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	if entry.Level == logrus.InfoLevel {
		entry.Message = fmt.Sprintf("%s\n", entry.Message)
		return []byte(entry.Message), nil
	}
	return f.TextFormatter.Format(entry)
}

func InitLogger(verbose bool) {
	Log.SetFormatter(&CustomFormatter{logrus.TextFormatter{}})
	if verbose {
		Log.SetLevel(logrus.DebugLevel)
		Log.Debugln("Verbose (debug) logging enabled")
	} else {
		Log.SetLevel(logrus.InfoLevel)
	}
}
