package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_intersect(t *testing.T) {
	ass := assert.New(t)
	firstArray := []string{"Albert", "Arthur", "Abbigaelle", "Barbara"}
	secondArray := []string{"Abbigaelle", "Bambi", "Björn", "Barbara"}
	both, a, b := intersect(firstArray, secondArray)

	ass.Exactly([]string{"Abbigaelle", "Barbara"}, both)
	ass.Exactly([]string{"Albert", "Arthur"}, a)
	ass.Exactly([]string{"Bambi", "Björn"}, b)
}
