package main

import (
	"fmt"
	"testing"

	"github.com/cnf/structhash"
	"github.com/stretchr/testify/assert"
)

func Test_readExcel(t *testing.T) {
	ass := assert.New(t)
	users, rolesMap, err := loadExcel("./userBase.xlsx")
	ass.Nil(err)

	hashUsers := fmt.Sprintf("%x", structhash.Md5(users, 1))
	ass.Equal("dc752fde23cf5ebfea42301c74cab527", hashUsers)

	hashRolesMap := fmt.Sprintf("%x", structhash.Md5(rolesMap, 1))
	ass.Equal("0fc072173fd22e567dbe26c474ea2547", hashRolesMap)
}
