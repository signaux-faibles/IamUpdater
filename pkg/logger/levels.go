package logger

import (
	"log/slog"
	"strings"

	"github.com/pkg/errors"
)

const (
	levelTrace      = slog.Level(-8)
	levelNotice     = slog.Level(2)
	levelTraceName  = "TRACE"
	levelNoticeName = "NOTICE"
)

var levelNames = map[slog.Leveler]string{
	levelTrace:  levelTraceName,
	levelNotice: levelNoticeName,
}

func customizeLogLevelNames(_ []string, a slog.Attr) slog.Attr {
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

func parseLogLevel(logLevel string) (slog.Leveler, error) {
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
	case levelTraceName:
		return levelTrace, nil
	case levelNoticeName:
		return levelNotice, nil
	default:
		return slog.LevelWarn, errors.New("log level inconnu : '" + logLevel + "'")
	}
}
