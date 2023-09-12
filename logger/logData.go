package logger

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/Nerzal/gocloak/v13"
)

type LogContext map[string]interface{}

func ContextForMethod(method string) LogContext {
	fields := map[string]interface{}{
		"method": method,
	}
	return fields
}

func ContextForMethode(method interface{}) LogContext {
	methodName := runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()
	fields := map[string]interface{}{
		"method": methodName,
	}
	return fields
}

func (d LogContext) AddUser(user gocloak.User) LogContext {
	d["user"] = *user.Username
	return d
}

func (d LogContext) AddError(err error) LogContext {
	d["error"] = err
	return d
}

func (d LogContext) removeError() LogContext {
	delete(d, "error")
	return d
}

func (d LogContext) AddArray(key string, any []string) LogContext {
	d[key] = strings.Join(any, ", ")
	return d
}

func (d LogContext) AddAny(key string, any interface{}) LogContext {
	d[key] = any
	return d
}

func (d LogContext) Remove(key string) LogContext {
	delete(d, key)
	return d
}

func (d LogContext) AddClient(input gocloak.Client) LogContext {
	d["clientId"] = *input.ClientID
	return d
}

func (d LogContext) AddRole(input gocloak.Role) LogContext {
	d["role"] = role2string(input)
	return d
}

func (d LogContext) AddRoles(all []gocloak.Role) LogContext {
	var val string
	if all == nil {
		val = ""
	} else {
		val = strings.Join(toStrings(all, role2string), ", ")
	}
	d["roles"] = val
	return d
}
