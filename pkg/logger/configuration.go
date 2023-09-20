package logger

import (
	"log/slog"
	"os"
	"slices"

	"github.com/samber/slog-formatter"
	"github.com/samber/slog-multi"
)

func configFormatters() slogmulti.Middleware {
	formattingMiddleware := slogformatter.NewFormatterHandler(
		errorFormatter(),
		keycloakUserFormatter(),
		clientFormatter(),
		singleRoleFormatter(),
		manyRolesFormatter(),
		wekanBoardLabelFormatter(),
		wekanUserUpdateFormatter(),
	)
	return formattingMiddleware
}

func composeReplaceAttrs(
	funcs ...func(args []string, a slog.Attr) slog.Attr) func(args []string, a slog.Attr) slog.Attr {
	if len(funcs) == 0 {
		return nil
	}
	if len(funcs) == 1 {
		return funcs[0]
	}
	composition := func(args []string, a slog.Attr) slog.Attr {
		attr := funcs[0](args, a)
		return funcs[1](args, attr)
	}
	if len(funcs) == 2 {
		return composition
	}
	composedArray := slices.Insert(funcs[2:], 0, composition)
	return composeReplaceAttrs(composedArray...)
}

func configFileHandler(logFilename string, timeFormat string) *slog.TextHandler {
	var err error
	var file *os.File
	if file, err = os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644); err != nil {
		slog.Error("erreur à l'ouverture du fichier de log", slog.String("filename", logFilename), slog.Any("error", err))
		panic(err)
	}
	return slog.NewTextHandler(file, &slog.HandlerOptions{
		Level:       loglevel,
		ReplaceAttr: composeReplaceAttrs(customizeTimeFormat(timeFormat), customizeLogLevelNames),
	})
}

func configLogLevel(configLogLevel string) {
	var err error
	level := loglevel.Level()
	if level, err = parseLogLevel(configLogLevel); err != nil {
		slog.Warn("erreur de configuration sur le log level", slog.String("valeur", configLogLevel))
	}
	loglevel.Set(level)
}
