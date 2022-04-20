package main

import (
	"fmt"
	"testing"

	"github.com/cnf/structhash"
)

func Test_readExcel(t *testing.T) {
	t.Log("test de lecture excel")
	users, userMap, err := loadExcel("./userBase.xlsx")
	if err != nil {
		t.Fatalf("ne peut lire le fichier d'exemple: %s", err.Error())
	}
	hashUsers := fmt.Sprintf("%x", structhash.Md5(users, 1))
	if hashUsers != "d2578b3813f40abdb0e5d57fd801c810" {
		fmt.Println(hashUsers)
		t.Fatalf("la lecture du fichier renvoie des données utilisateurs différentes")
	}
	hashUserMap := fmt.Sprintf("%x", structhash.Md5(userMap, 1))
	if hashUserMap != "0fc072173fd22e567dbe26c474ea2547" {
		fmt.Println(hashUserMap)
		t.Fatalf("la lecture du fichier renvoie un mapping différent")
	}
}
