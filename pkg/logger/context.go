package logger

import (
	"log/slog"
	"reflect"
	"runtime"
	"strings"

	"github.com/Nerzal/gocloak/v13"
)

type LogContext map[string]slog.Attr

func ContextForMethod(method interface{}) *LogContext {
	methodName := runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()
	context := make(LogContext)
	return (&context).AddString("method", methodName)
}

func (d *LogContext) AddAny(key string, value any) *LogContext {
	(*d)[key] = slog.Any(key, value)
	return d
}

func (d *LogContext) AddString(key string, value string) *LogContext {
	(*d)[key] = slog.String(key, value)
	return d
}

func (d *LogContext) AddInt(key string, value int) *LogContext {
	(*d)[key] = slog.Int(key, value)
	return d
}

func (d *LogContext) AddArray(key string, any []string) *LogContext {
	return d.AddString(key, strings.Join(any, ", "))
}

func (d *LogContext) AddClient(input gocloak.Client) *LogContext {
	return d.AddAny("clientId", input)
}

func (d *LogContext) AddRole(input gocloak.Role) *LogContext {
	return d.AddString("role", role2string(input))
}

func (d *LogContext) AddRoles(all []gocloak.Role) *LogContext {
	var val string
	if all == nil {
		val = ""
	} else {
		val = strings.Join(toStrings(all, role2string), ", ")
	}
	return d.AddString("roles", val)
}

func (d *LogContext) AddUser(user gocloak.User) *LogContext {
	return d.AddAny("user", user)
}

func (d *LogContext) Remove(key string) *LogContext {
	delete(*d, key)
	return d
}

func (d *LogContext) addError(err error) *LogContext {
	return d.AddAny("error", err)
}
