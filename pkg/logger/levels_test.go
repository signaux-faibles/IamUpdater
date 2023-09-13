package logger

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"keycloakUpdater/v2/pkg/structs"
)

func Test_level_Trace(t *testing.T) {
	ass := assert.New(t)

	traceloggerConfig := structs.LoggerConfig{
		Filename:        createTempFilename(t),
		Level:           "trace",
		TimestampFormat: time.DateTime,
	}
	ConfigureWith(traceloggerConfig)
	Trace("message de Trace", nil)
	Debug("message de Debug", nil)

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(traceloggerConfig.Filename)
	ass.NoError(err)
	ass.Contains(string(logsFromFile), "level="+levelTraceName)
	ass.Contains(string(logsFromFile), "level="+slog.LevelDebug.String())
}

func Test_level_Notice(t *testing.T) {
	ass := assert.New(t)

	noticeloggerConfig := structs.LoggerConfig{
		Filename:        createTempFilename(t),
		Level:           levelNoticeName,
		TimestampFormat: time.DateTime,
	}
	ConfigureWith(noticeloggerConfig)
	Info("message d'Info", nil)
	Notice("message de Notice", nil)
	Warn("message de Warn", nil)

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(noticeloggerConfig.Filename)
	ass.NoError(err)
	ass.NotContains(string(logsFromFile), "level="+slog.LevelInfo.String())
	ass.Contains(string(logsFromFile), "level="+levelNoticeName)
	ass.Contains(string(logsFromFile), "level="+slog.LevelWarn.String())
}
