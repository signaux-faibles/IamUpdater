package logger

import (
	"log/slog"
	"strings"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/samber/slog-formatter"
	slogmulti "github.com/samber/slog-multi"
)

const defaultTimestampFormat = "2006-01-02 15:04:05"

func addFormattersToHandler(formatters slogmulti.Middleware, handler slog.Handler) slog.Handler {
	return slogmulti.Pipe(formatters).Handler(handler)
}

func timeFormatter(input string) slogformatter.Formatter {
	format := defaultTimestampFormat
	if len(input) > 0 {
		format = input
	}
	return slogformatter.TimeFormatter(format, time.UTC)
}

func errorFormatter() slogformatter.Formatter {
	return slogformatter.ErrorFormatter("error")
}

func clientFormatter() slogformatter.Formatter {
	return slogformatter.FormatByType(func(client gocloak.Client) slog.Value {
		return slog.StringValue(*client.ClientID)
	})
}

func userFormatter() slogformatter.Formatter {
	return slogformatter.FormatByType(func(user gocloak.User) slog.Value {
		return slog.StringValue(*user.Username)
	})
}

func singleRoleFormatter() slogformatter.Formatter {
	return slogformatter.FormatByType(func(role gocloak.Role) slog.Value {
		return slog.StringValue(role2string(role))
	})
}

func manyRolesFormatter() slogformatter.Formatter {
	return slogformatter.FormatByType(func(roles []gocloak.Role) slog.Value {
		var val string
		if roles == nil {
			val = ""
		} else {
			val = strings.Join(toStrings(roles, role2string), ", ")
		}
		return slog.StringValue(val)
	})
}

func toStrings[T any](array []T, toString func(T) string) []string {
	y := make([]string, len(array))
	for i, v := range array {
		y[i] = toString(v)
	}
	return y
}

func role2string(role gocloak.Role) string { return *role.Name }
