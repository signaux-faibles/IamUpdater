package logger

import (
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

func (d Data) AddArray(key string, any []string) {
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
	d["role"] = role2string(input)
}

func (d Data) AddRoles(all []gocloak.Role) {
	if all == nil {
		return
	}
	d["roles"] = strings.Join(toStrings(all, role2string), ", ")
}

func toStrings[T any](array []T, toString func(T) string) []string {
	y := make([]string, len(array))
	for i, v := range array {
		y[i] = toString(v)
	}
	return y
}

func role2string(role gocloak.Role) string { return *role.Name }
