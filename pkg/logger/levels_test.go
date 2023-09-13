package logger

import (
	"context"
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
	slog.Log(context.Background(), LevelTrace, "message de Trace")
	slog.Log(context.Background(), slog.LevelDebug, "message de Debug")

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(traceloggerConfig.Filename)
	ass.NoError(err)
	ass.Contains(string(logsFromFile), "level="+LevelTraceName)
	ass.Contains(string(logsFromFile), "level="+slog.LevelDebug.String())
}

func Test_level_Notice(t *testing.T) {
	ass := assert.New(t)

	noticeloggerConfig := structs.LoggerConfig{
		Filename:        createTempFilename(t),
		Level:           LevelNoticeName,
		TimestampFormat: time.DateTime,
	}
	ConfigureWith(noticeloggerConfig)
	slog.Log(context.Background(), slog.LevelInfo, "message d'Info")
	slog.Log(context.Background(), LevelNotice, "message de Notice")
	slog.Log(context.Background(), slog.LevelWarn, "message de Warn")

	var logsFromFile []byte
	var err error
	logsFromFile, err = os.ReadFile(noticeloggerConfig.Filename)
	ass.NoError(err)
	ass.NotContains(string(logsFromFile), "level="+slog.LevelInfo.String())
	ass.Contains(string(logsFromFile), "level="+LevelNoticeName)
	ass.Contains(string(logsFromFile), "level="+slog.LevelWarn.String())
}
