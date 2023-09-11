package logger

import (
	"context"
	"log"
	"log/slog"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/pkg/errors"
	slogmulti "github.com/samber/slog-multi"

	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
)

var loglevel *slog.LevelVar

func init() {
	loglevel = new(slog.LevelVar)
	loglevel.Set(slog.LevelInfo)

	handler := slog.NewJSONHandler(log.Default().Writer(), &slog.HandlerOptions{
		Level: loglevel,
	})
	parentLogger := slog.New(handler)
	buildInfo, _ := debug.ReadBuildInfo()
	sha1 := findBuildSetting(buildInfo.Settings, "vcs.revision")
	appLogger := parentLogger.With(
		slog.Group("app", slog.String("sha1", sha1)),
	)
	slog.SetDefault(appLogger)
	//logger = appLogger
}

func ConfigureWith(config structs.LoggerConfig) {
	configLogLevel(config.Level)
	fileHandler := configFileHandler(config.Filename)
	formatters := configFormatters(config.TimestampFormat)
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

func Debug(msg string, data map[string]interface{}) {
	logWithContext(slog.LevelDebug, msg, data, nil)
}

func Info(msg string, data map[string]interface{}) {
	logWithContext(slog.LevelInfo, msg, data, nil)
}

func Warn(msg string, data map[string]interface{}) {
	logWithContext(slog.LevelWarn, msg, data, nil)
}

func WarnE(msg string, data map[string]interface{}, err error) {
	logWithContext(slog.LevelWarn, msg, data, err)
}

func Error(msg string, data map[string]interface{}) {
	logWithContext(slog.LevelWarn, msg, data, nil)
}

func ErrorE(msg string, data map[string]interface{}, err error) {
	logWithContext(slog.LevelWarn, msg, data, err)
}

func Errorf(msg string, args ...interface{}) {
	slog.Error(msg, args...)
}

func Panicf(msg string, args ...interface{}) {
	Errorf(msg, args)
	panic(msg)
}

func Panic(err error) {
	Panicf(err.Error())
}

func logWithContext(level slog.Level, msg string, data map[string]interface{}, err error) {
	var logCtx []any
	for k, v := range data {
		logCtx = append(logCtx, slog.Any(k, v))
	}
	if err != nil {
		logCtx = append(logCtx, slog.String("error", err.Error()))
	}
	slog.Log(context.Background(), level, msg, logCtx...)
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
	default:
		return slog.LevelInfo, errors.New("log level inconnu : '" + logLevel + "'")
	}
}
