package main

import (
	"fmt"
	"testing"

	"github.com/cnf/structhash"
	"github.com/spf13/viper"
)

func Test_readExcel(t *testing.T) {
	t.Log("test de lecture excel")
	viper.Set("base", "./userBase.xlsx")
	users, userMap, err := loadExcel()
	if err != nil {
		t.Fatalf("ne peut lire le fichier d'exemple: %s", err.Error())
	}
	hashUsers := fmt.Sprintf("%x", structhash.Md5(users, 1))
	if hashUsers != "12b815bf6a2eb9775e602f5b545edb98" {
		fmt.Println(hashUsers)
		t.Fatalf("la lecture du fichier renvoie des données utilisateurs différentes")
	}
	hashUserMap := fmt.Sprintf("%x", structhash.Md5(userMap, 1))
	if hashUserMap != "0fc072173fd22e567dbe26c474ea2547" {
		fmt.Println(hashUserMap)
		t.Fatalf("la lecture du fichier renvoie un mapping différent")
	}
}
