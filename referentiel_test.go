package main

import (
  "github.com/stretchr/testify/assert"
  "testing"
)

func Test_loadReferentiel(t *testing.T) {
  assertions := assert.New(t)
  t.Log("test de lecture des rôles géographiques")
  referentiel, err := loadReferentiel("referentiel.csv")
  assertions.Nil(err)
  assertions.Len(referentiel, 35)

  rolesLaReunion := referentiel["La Réunion"]
  assertions.NotNil(rolesLaReunion)
  assertions.Len(rolesLaReunion, 2)

  rolesAuvergne := referentiel["Auvergne"]
  assertions.NotNil(rolesAuvergne)
  assertions.Len(rolesAuvergne, 4)

  //rolesLaReunion := referentiel["La Réunion"]
  //assertions.NotNil(rolesLaReunion)
  //assertions.Len(rolesLaReunion,2)
  //users, userMap, err := loadExcel("./userBase.xlsx")
  //if err != nil {
  //  t.Fatalf("ne peut lire le fichier d'exemple: %s", err.Error())
  //}
  //hashUsers := fmt.Sprintf("%x", structhash.Md5(users, 1))
  //if hashUsers != "d2578b3813f40abdb0e5d57fd801c810" {
  //  fmt.Println(hashUsers)
  //  t.Fatalf("la lecture du fichier renvoie des données utilisateurs différentes")
  //}
  //hashUserMap := fmt.Sprintf("%x", structhash.Md5(userMap, 1))
  //if hashUserMap != "0fc072173fd22e567dbe26c474ea2547" {
  //  fmt.Println(hashUserMap)
  //  t.Fatalf("la lecture du fichier renvoie un mapping différent")
  //}
}
