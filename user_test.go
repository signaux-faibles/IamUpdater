package main

import (
	"github.com/stretchr/testify/assert"
	"io/fs"
	"testing"
)

func Test_loadUsers(t *testing.T) {
	assertions := assert.New(t)
	t.Log("test de chargement des users")
	filenames := []string{
		"test/sample/users/john_doe.yml",
		"test/sample/users/raphael_squelbut.yml",
		"test/sample/users/un_mec_pas_de_l_urssaf.yml",
	}
	users, err := loadUsers(filenames)

	assertions.Nil(err)
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

func Test_loadNoUsers(t *testing.T) {
	assertions := assert.New(t)
	t.Log("test de chargement avec aucun users")
	var filenames []string
	users, err := loadUsers(filenames)

	assertions.Nil(users)
	assertions.NotNil(err)
	assertions.EqualError(err, "pas de fiche utilisateur à charger")
}

func Test_loadUsers_noSuchFile(t *testing.T) {
	assertions := assert.New(t)
	t.Log("test de chargement des users")
	fichierInexistant := "test/sample/users/fichier_inexistant.yml"
	filenames := []string{
		fichierInexistant,
	}
	users, err := loadUsers(filenames)

	assertions.Nil(users)
	assertions.NotNil(err)
	assertions.Error(err)
	assertions.ErrorAs(err, &fs.ErrNotExist)
}

func Test_loadUsers_nonYamlFile(t *testing.T) {
	assertions := assert.New(t)
	t.Log("test de chargement des users")
	invalidYamlFile := "test/sample/users/invalid_yml_file.yml"
	filenames := []string{
		"test/sample/users/admin.yml",
		invalidYamlFile,
	}
	users, err := loadUsers(filenames)
	expectedMessage := "problème avec le contenu de la fiche utilisateur " + invalidYamlFile + ": yaml: mapping values are not allowed in this context"
	assertions.Nil(users)
	assertions.NotNil(err)
	assertions.Error(err)
	assertions.EqualError(err, expectedMessage)
}
