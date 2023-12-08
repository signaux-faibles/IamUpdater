package logger

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func Test_parseLogLevel(t *testing.T) {
	type args struct {
		logLevel string
	}
	tests := []struct {
		name string
		args args
		want slog.Leveler
	}{
		{"parse niveau TRACE", args{"TRACE"}, levelTrace},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseLogLevel(tt.args.logLevel)
			require.NoError(t, err)
			assert.Equalf(t, tt.want, got, "parseLogLevel(%v)", tt.args.logLevel)
		})
	}
}
