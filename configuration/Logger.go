package configuration

import (
	"github.com/alecthomas/log4go"
	"fmt"
	"os"
	"strings"
)

const LOG_FORMAT = "%D %T %L %M"
const LOG_FILENAME = "logs/estreamer.log"

var logger log4go.Logger

func InitializeLogger(logLevel string) {
	currentPath, _ := os.Getwd()
	if isExistDirectory(fmt.Sprintf("%s/logs", currentPath)) == false {
		os.MkdirAll(fmt.Sprintf("%s/logs", currentPath), os.FileMode(644))
	}

	logger = log4go.NewLogger()

	fileLogWriter := log4go.NewFileLogWriter(fmt.Sprintf("%s/%s", currentPath, LOG_FILENAME), true)
	if fileLogWriter != nil {
		fileLogWriter.SetFormat(LOG_FORMAT)
		fileLogWriter.SetRotateSize(10 * 1024 * 1024)
		logger.AddFilter("file", getLogLevel(logLevel), fileLogWriter)
	}

	/*
	consoleLogWriter := log4go.NewConsoleLogWriter()
	consoleLogWriter.SetFormat(LOG_FORMAT)
	logger.AddFilter("console", getLogLevel(logLevel), consoleLogWriter)
	*/
}

func isExistDirectory(directory string) bool {
	_, err := os.Open(directory)
	if err != nil {
		return false
	}

	return true
}

func getLogLevel(logLevel string) log4go.Level {
	switch strings.ToLower(logLevel) {
	case "debug":
		return log4go.DEBUG

	case "info":
		return log4go.INFO

	case "error":
		return log4go.ERROR

	case "warning":
		return log4go.WARNING

	default:
		return log4go.INFO
	}
}

func WriteDebug(format string, a ...interface{}) {
	logger.Debug(format, a...)
}

func WriteInfo(format string, a ...interface{}) {
	logger.Info(format, a...)
}

func WriteError(err error, format string, a ...interface{}) {
	if err != nil {
		logger.Error("%s: %s", fmt.Sprintf(format, a...), err.Error())
	} else {
		logger.Error(format, a...)
	}
}

func WriteWarning(format string, a ...interface{}) {
	logger.Warn(format, a...)
}

func CloseLogger() {
	if logger != nil {
		logger.Close()
	}
}


