package logger

import (
	"fmt"
	"github.com/Nerzal/gocloak/v11"
	"strings"
)

type Data map[string]interface{}

func DataForMethod(method string) Data {
	fields := map[string]interface{}{
		"method": method,
	}
	return fields
}

func (d Data) AddUser(user gocloak.User) {
	d["user"] = *user.Username
}

func (d Data) AddError(err error) {
	d["error"] = err
}

func (d Data) AddArray(key string, any []interface{}) {
	var s []string
	for _, v := range any {
		s = append(s, fmt.Sprintf("%v", v))
	}
	d.AddStringArray(key, s)
}

func (d Data) AddStringArray(key string, any []string) {
	d[key] = strings.Join(any, ", ")
}

func (d Data) AddAny(key string, any interface{}) {
	d[key] = any
}

func (d Data) Remove(key string) {
	delete(d, key)
}

func (d Data) AddClient(input gocloak.Client) {
	d["clientId"] = *input.ClientID
}

func (d Data) AddRole(input gocloak.Role) {
	d["role"] = *input.Name
}

func ToInterfaces[T any](array []T) []interface{} {
	y := make([]interface{}, len(array))
	for i, v := range array {
		y[i] = v
	}
	return y
}
