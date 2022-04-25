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
	//logger.ReportCaller = true
	var err error

	// level
	var logLevel logrus.Level
	if logLevel, err = logrus.ParseLevel(config.Level); err != nil {
		logger.Infof("bad log level '%s' : %s", config.Level, err)
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	// console
	logger.SetOutput(colorable.NewColorableStdout())
	consoleFormatter := &logrus.TextFormatter{
		PadLevelText:    true,
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: config.TimestampFormat,
	}
	logger.SetFormatter(consoleFormatter)
	//consoleFormatter := &SimpleFormatter{}
	logger.SetFormatter(consoleFormatter)

	// file
	fileFormatter := &logrus.TextFormatter{
		DisableColors:   true,
		TimestampFormat: config.TimestampFormat,
		PadLevelText:    true,
	}
	var hook logrus.Hook
	if config.Filename != "" {
		if config.Rotation {
			hook = rotateFileHook(config.Filename, logLevel, fileFormatter)
		} else {
			hook = simpleFileHook(config.Filename, logLevel, fileFormatter)
		}
		logger.AddHook(hook)
	}
}

func Debugf(msg string, args ...interface{}) {
	logger.Debugf(msg, args...)
}

func Debug(msg string, data map[string]interface{}) {
	logger.WithFields(data).Debug(msg)
}

func Infof(msg string, args ...interface{}) {
	logger.Infof(msg, args...)
}

func Info(msg string, data map[string]interface{}) {
	logger.WithFields(data).Info(msg)
}

func Warnf(msg string, args ...interface{}) {
	logger.Warnf(msg, args...)
}

func Warn(msg string, data map[string]interface{}) {
	logger.WithFields(data).Warning(msg)
}

func WarnE(msg string, data map[string]interface{}, err error) {
	Data(data).AddError(err)
	Warn(msg, data)
}

func Error(msg string, data map[string]interface{}) {
	logger.WithFields(data).Error(msg)
}

func ErrorE(msg string, data map[string]interface{}, err error) {
	Data(data).AddError(err)
	Error(msg, data)
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
