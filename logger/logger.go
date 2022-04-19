package logger

import (
	"github.com/mattn/go-colorable"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
	"os"

	"github.com/snowzach/writerhook"
)

var logger *logrus.Logger

func init() {
	logger = logrus.New()
	// formatter
	consoleFormatter := &logrus.TextFormatter{
		PadLevelText:  true,
		ForceColors:   true,
		FullTimestamp: true,
		//TimestampFormat: ,
	}

	logger.SetLevel(logrus.InfoLevel)
	logger.SetOutput(colorable.NewColorableStdout())
	logger.SetFormatter(consoleFormatter)
}

func ConfigureWith(config structs.LoggerConfig) {
	//log.ReportCaller = true
	var err error

	// level
	var logLevel logrus.Level
	if logLevel, err = logrus.ParseLevel(config.Level); err != nil {
		logger.Info("bad log level '%s' : %s", config.Level, err)
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)
	// formatter
	consoleFormatter := &logrus.TextFormatter{
		PadLevelText:    true,
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: config.TimestampFormat,
	}
	fileFormater := &logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: config.TimestampFormat,
		PadLevelText:    true,
	}
	var hook logrus.Hook

	if config.Rotation {
		hook = rotateFileHook(config.Filename, logLevel, fileFormater)
	} else {
		hook = simpleFileHook(config.Filename, logLevel, fileFormater)
	}

	logger.SetOutput(colorable.NewColorableStdout())
	logger.SetFormatter(consoleFormatter)

	logger.AddHook(hook)
}

func Debugf(msg string, args ...interface{}) {
	logger.Debugf(msg, args...)
}

func Infof(msg string, args ...interface{}) {
	logger.Infof(msg, args...)
}

func Warnf(msg string, args ...interface{}) {
	logger.Warnf(msg, args...)
}

func Errorf(msg string, args ...interface{}) {
	logger.Errorf(msg, args...)
}

func Panicf(msg string, args ...interface{}) {
	logger.Panicf(msg, args...)
}

func Panic(err error) {
	logger.Panic(err)
}

func rotateFileHook(filename string, logLevel logrus.Level, fileFormater logrus.Formatter) logrus.Hook {

	hook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   filename,
		MaxSize:    50, // megabytes
		MaxBackups: 99, // amouts
		MaxAge:     1,  //days
		Level:      logLevel,
		Formatter:  fileFormater,
	})
	if err != nil {
		logger.Fatalf("Failed to initialize file rotate hook: %v", err)
	}
	return hook
}

func simpleFileHook(filename string, logLevel logrus.Level, fileFormater logrus.Formatter) logrus.Hook {
	var hook logrus.Hook
	var err error
	var file *os.File

	if file, err = os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err != nil {
		logger.Fatalf("Failed to initialize file hook: %v", err)
	}

	if hook, err = writerhook.NewWriterHook(file, logLevel, fileFormater); err != nil {
		logger.Fatalf("Failed to initialize file hook: %v", err)
	}
	return hook
}
