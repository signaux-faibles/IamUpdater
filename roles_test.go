package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoles_add_on_nil_roles_dont_throw_error(t *testing.T) {
	ass := assert.New(t)
	var roles Roles
	roles.add()
	ass.Nil(roles)

	roles.add("any")
	ass.Contains(roles, "any")
}

func TestRoles_add_nil_role(t *testing.T) {
	ass := assert.New(t)
	roles := Roles{"first"}
	roles.add()
	ass.Len(roles, 1)
	ass.Contains(roles, "first")
}

func TestRoles_add_same_role(t *testing.T) {
	ass := assert.New(t)
	sameRole := "same role"
	roles := Roles{sameRole}
	roles.add(sameRole)
	ass.Len(roles, 1)
	ass.Contains(roles, sameRole)
}

func TestRoles_add_many_roles(t *testing.T) {
	ass := assert.New(t)
	first := "premier role"
	second := "deuxieme role"
	third := "troisieme role"
	roles := Roles{first}
	roles.add(second, third)
	ass.Len(roles, 3)
	ass.Contains(roles, first, second, third)
}
