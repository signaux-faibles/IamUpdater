package logger

import (
	"github.com/mattn/go-colorable"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"github.com/sirupsen/logrus"
	"github.com/snowzach/rotatefilehook"
)

type Logger struct{ *logrus.Logger }

func InfoLogger() *Logger {
	return NewLogger(logrus.InfoLevel)
}

func NewLogger(level logrus.Level) *Logger {
	log := logrus.New()
	// formatter
	consoleFormatter := &logrus.TextFormatter{
		PadLevelText:  true,
		ForceColors:   true,
		FullTimestamp: true,
		//TimestampFormat: ,
	}

	log.SetLevel(level)
	log.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(consoleFormatter)

	return &Logger{log}
}

func (log *Logger) ConfigureWith(config structs.LoggerConfig) {
	//log.ReportCaller = true
	var err error

	// level
	var logLevel logrus.Level
	if logLevel, err = logrus.ParseLevel(config.Level); err != nil {
		log.Info("bad log level '%s' : %s", config.Level, err)
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)

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

	rotateFileHook, err := rotatefilehook.NewRotateFileHook(rotatefilehook.RotateFileConfig{
		Filename:   config.Filename,
		MaxSize:    50, // megabytes
		MaxBackups: 3,  // amouts
		MaxAge:     28, //days
		Level:      logLevel,
		Formatter:  fileFormater,
	})

	if err != nil {
		log.Fatalf("Failed to initialize file rotate hook: %v", err)
	}

	log.SetOutput(colorable.NewColorableStdout())
	log.SetFormatter(consoleFormatter)

	log.AddHook(rotateFileHook)
}
