package logger

import (
	"log/slog"
	"os"

	"github.com/samber/slog-formatter"
	"github.com/samber/slog-multi"
)

func configFormatters(timeFormat string) slogmulti.Middleware {
	formattingMiddleware := slogformatter.NewFormatterHandler(
		timeFormatter(timeFormat),
		errorFormatter(),
		userFormatter(),
		clientFormatter(),
		singleRoleFormatter(),
		manyRolesFormatter(),
	)
	return formattingMiddleware
}

func configFileHandler(logFilename string) *slog.TextHandler {
	var err error
	var file *os.File
	if file, err = os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err != nil {
		slog.Error("erreur à l'ouverture du fichier de log", slog.String("filename", logFilename), slog.Any("error", err))
		panic(err)
	}
	return slog.NewTextHandler(file, &slog.HandlerOptions{
		Level: loglevel,
	})
}

func configLogLevel(configLogLevel string) {
	var err error
	var level = loglevel.Level()
	if level, err = parseLogLevel(configLogLevel); err != nil {
		logger.Warn("erreur de configuration sur le log level", slog.String("valeur", configLogLevel))
	}
	loglevel.Set(level)
}
