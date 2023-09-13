package logger

import (
	"log/slog"
	"strings"

	"github.com/pkg/errors"
)

const (
	LevelTrace      = slog.Level(-8)
	LevelNotice     = slog.Level(2)
	LevelTraceName  = "TRACE"
	LevelNoticeName = "NOTICE"
)

var levelNames = map[slog.Leveler]string{
	LevelTrace:  LevelTraceName,
	LevelNotice: LevelNoticeName,
}

func customizeLogLevelNames() func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.LevelKey {
			level := a.Value.Any().(slog.Level)
			levelLabel, exists := levelNames[level]
			if !exists {
				levelLabel = level.String()
			}
			a.Value = slog.StringValue(levelLabel)
		}
		return a
	}
}

func parseLogLevel(logLevel string) (slog.Level, error) {
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
		// custom levels
	case LevelTraceName:
		return LevelTrace, nil
	case LevelNoticeName:
		return LevelNotice, nil
	default:
		return slog.LevelWarn, errors.New("log level inconnu : '" + logLevel + "'")
	}
}
