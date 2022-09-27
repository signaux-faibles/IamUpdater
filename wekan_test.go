package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_intersect(t *testing.T) {
	ass := assert.New(t)
	arrayWithA := []string{"A", "AA", "AB", "BA"}
	arrayWithB := []string{"B", "BB", "AB", "BA"}
	both, a, b := intersect(arrayWithA, arrayWithB)

	ass.Exactly([]string{"AB", "BA"}, both)
	ass.Exactly([]string{"A", "AA"}, a)
	ass.Exactly([]string{"B", "BB"}, b)
}
