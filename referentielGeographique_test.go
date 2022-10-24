package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReferentielGeographique_toRoles(t *testing.T) {
	ass := assert.New(t)

	actual := referentiel.toRoles()

	ass.Contains(actual, franceEntiere)
	ass.Len(actual[franceEntiere], 101)
	ass.Equal(actual[franceEntiere][0], "01")
	ass.Equal(actual[franceEntiere][100], "976")
}
