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
	ass.Equal("ed87c55be7c75632fbf8dd5e9df3e94e", hashUsers)

	hashRolesMap := fmt.Sprintf("%x", structhash.Md5(rolesMap, 1))
	ass.Equal("0fc072173fd22e567dbe26c474ea2547", hashRolesMap)
}
