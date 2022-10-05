package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUser_roles_with_niveau_a(t *testing.T) {
	ass := assert.New(t)
	user := User{niveau: "a"}
	actual := user.roles()

	expected := []string{"score", "detection", "pge", "urssaf", "dgefp", "bdf"}
	ass.ElementsMatch(actual, expected)
}

func TestUser_roles_with_niveau_b(t *testing.T) {
	ass := assert.New(t)
	user := User{niveau: "b"}
	actual := user.roles()

	expected := []string{"score", "detection", "pge", "dgefp"}
	ass.ElementsMatch(actual, expected)
}

func TestUser_roles_with_scopes(t *testing.T) {
	ass := assert.New(t)
	scopes := []string{"first", "second"}
	user := User{scope: scopes}
	actual := user.roles()

	ass.Contains(actual, scopes[0])
	ass.Contains(actual, scopes[1])
}

func TestUser_roles_with_acces_geographique(t *testing.T) {
	ass := assert.New(t)
	accessGeographique := "any where"
	user := User{accesGeographique: accessGeographique}
	actual := user.roles()
	ass.Contains(actual, accessGeographique)
}
