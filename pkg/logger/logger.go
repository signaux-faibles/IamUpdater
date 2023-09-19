package logger

import (
	"context"
	"log"
	"log/slog"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	slogmulti "github.com/samber/slog-multi"

	"keycloakUpdater/v2/pkg/structs"
)

var loglevel *slog.LevelVar

func init() {
	loglevel = new(slog.LevelVar)
	loglevel.Set(slog.LevelInfo)

	handler := slog.NewJSONHandler(log.Default().Writer(), &slog.HandlerOptions{
		Level:       loglevel,
		ReplaceAttr: composeReplaceAttrs(customizeLogLevelNames, customizeTimeFormat(time.DateTime)),
	})
	parentLogger := slog.New(handler)
	buildInfo, _ := debug.ReadBuildInfo()
	sha1 := findBuildSetting(buildInfo.Settings, "vcs.revision")
	appLogger := parentLogger.With(
		slog.Group("app", slog.String("sha1", sha1)),
	)
	slog.SetDefault(appLogger)
}

func ConfigureWith(config structs.LoggerConfig) {
	configLogLevel(config.Level)
	fileHandler := configFileHandler(config.Filename, config.TimestampFormat)
	formatters := configFormatters()
	formattedFileHandler := addFormattersToHandler(formatters, fileHandler)

	defaultHandler := addFormattersToHandler(formatters, slog.Default().Handler())
	combinedHandlers := slogmulti.Fanout(formattedFileHandler, defaultHandler)
	slog.SetDefault(slog.New(combinedHandlers))
	slog.Info("configuration des loggers effectuée", slog.Group(
		"config",
		slog.String("level", config.Level),
		slog.String("filename", config.Filename),
		slog.String("timeFormat", config.TimestampFormat),
	))
}

func Trace(msg string, data *LogContext) {
	logWithContext(levelTrace, msg, data, nil)
}

func Debug(msg string, data *LogContext) {
	logWithContext(slog.LevelDebug, msg, data, nil)
}

func Info(msg string, data *LogContext) {
	logWithContext(slog.LevelInfo, msg, data, nil)
}

func Notice(msg string, data *LogContext) {
	logWithContext(levelNotice, msg, data, nil)
}

func Warn(msg string, data *LogContext) {
	logWithContext(slog.LevelWarn, msg, data, nil)
}

func Error(msg string, data *LogContext, err error) {
	logWithContext(slog.LevelError, msg, data, err)
}

func Panic(msg string, data *LogContext, err error) {
	Error(msg, data, err)
	panic(err)
}

func logWithContext(level slog.Level, msg string, data *LogContext, err error) {
	var logCtx []slog.Attr
	if data != nil {
		for _, v := range *data {
			logCtx = append(logCtx, v)
		}
	}
	if err != nil {
		logCtx = append(logCtx, slog.Any("error", err))
	}
	slog.LogAttrs(context.Background(), level, msg, logCtx...)
}

func findBuildSetting(settings []debug.BuildSetting, search string) string {
	retour := "NOT FOUND"
	slices.SortFunc(settings, func(s1 debug.BuildSetting, s2 debug.BuildSetting) int {
		return strings.Compare(s1.Key, s2.Key)
	})
	index, found := slices.BinarySearchFunc(settings, search, func(input debug.BuildSetting, searched string) int {
		return strings.Compare(input.Key, searched)
	})
	if found {
		retour = settings[index].Value
	}
	return retour
}
