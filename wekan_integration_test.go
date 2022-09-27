//go:build integration
// +build integration

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_WekanUpdate(t *testing.T) {
	ass := assert.New(t)

	err := WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/sample/userBase.xlsx")

	ass.Nil(err)
}

func Test_ListBoards(t *testing.T) {
	ass := assert.New(t)
	boards := excelUsers.listBoards()
	ass.Len(boards["tableau-crp-bfc"], 2)
	ass.Len(boards["tableau-codefi-nord"], 1)
	count := 0
	for range boards {
		count += 1
	}
	ass.Equal(count, 2)
}
