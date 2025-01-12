package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

const LogFileName = "gcm.log"

var Log = logrus.New()

func InitLogger(verbose bool) {

	// Create a log file
	file, err := os.OpenFile(GetLogFilePath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logrus.Fatalf("Failed to open log file: %v", err)
	}

	// Set the output of the logger to the file
	Log.SetOutput(file)

	Log.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	if verbose {
		Log.SetLevel(logrus.DebugLevel)
		Log.Debugln("Verbose (debug) logging enabled")
	} else {
		Log.SetLevel(logrus.InfoLevel)
	}
}

func GetLogFilePath() string {
	path, err := filepath.Abs(LogFileName)
	if err != nil {
		return LogFileName
	}
	return path
}
