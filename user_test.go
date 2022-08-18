package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_loadUsers(t *testing.T) {
	assertions := assert.New(t)
	t.Log("test de chargement des users")
	filenames := []string{
		"test/users/john_doe.yml",
		"test/users/raphael_squelbut.yml",
		"test/users/un_mec_pas_de_l_urssaf.yml",
	}
	users, err := loadUsers(filenames)
	if err != nil {
		t.Fatalf(err.Error())
	}
	assertions.Len(users, 3)
	assertions.NotNil(users["john.doe@zone51.gov.fr"])
	assertions.NotNil(users["ti_admin"])

	pasUrssaf := users["quelqun@pasdelurssaf.fr"]
	assertions.NotNil(pasUrssaf)
	assertions.NotContains(pasUrssaf.roles(), "urssaf")

	raphael := users["raphael.squelbut@shodo.io"]
	assertions.NotNil(raphael)
	assertions.Contains(raphael.roles(), "France entière")
	assertions.Contains(raphael.roles(), "urssaf")
}
