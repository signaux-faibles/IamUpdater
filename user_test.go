package main

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestUser_roles_with_niveau_a(t *testing.T) {
	ass := assert.New(t)
	user := User{niveau: "a"}
	actual := user.getRoles()
	sort.Strings(actual)
	expected := []string{"score", "detection", "pge", "urssaf", "dgefp", "bdf"}
	sort.Strings(expected)
	ass.ElementsMatch(actual, expected)
}

func TestUser_roles_with_niveau_b(t *testing.T) {
	ass := assert.New(t)
	user := User{niveau: "b"}
	actual := user.getRoles()
	expected := []string{"score", "detection", "pge", "dgefp"}
	ass.ElementsMatch(actual, expected)
}

func TestUser_roles_with_scopes(t *testing.T) {
	ass := assert.New(t)
	scopes := []string{"first", "second"}
	user := User{niveau: "0", scope: scopes}
	actual := user.getRoles()

	ass.Contains(actual, scopes[0])
	ass.Contains(actual, scopes[1])
}

func TestUser_roles_with_acces_geographique_and_niveau_a(t *testing.T) {
	ass := assert.New(t)
	accessGeographique := "any where"
	user := User{niveau: "A", accesGeographique: accessGeographique}
	actual := user.getRoles()
	ass.Contains(actual, accessGeographique)
}

func TestUser_roles_with_acces_geographique_and_niveau_0(t *testing.T) {
	ass := assert.New(t)
	accessGeographique := "any where"
	user := User{niveau: "0", accesGeographique: accessGeographique}
	actual := user.getRoles()
	ass.NotContains(actual, accessGeographique)
}
