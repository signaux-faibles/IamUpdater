package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_loadReferentiel(t *testing.T) {
	assertions := assert.New(t)
	t.Log("test de lecture des rôles géographiques")
	referentiel, err := loadReferentiel("regions_et_departements.csv")
	assertions.Nil(err)
	assertions.Len(referentiel, 35)

	rolesLaReunion := referentiel["La Réunion"]
	assertions.NotNil(rolesLaReunion)
	assertions.Len(rolesLaReunion, 2)

	rolesAuvergne := referentiel["Auvergne"]
	assertions.NotNil(rolesAuvergne)
	assertions.Len(rolesAuvergne, 4)
}
