package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_WekanUpdate(t *testing.T) {
	ass := assert.New(t)

	err := WekanUpdate(mongoUrl, "wekan", "signaux.faibles")

	ass.Nil(err)
}
