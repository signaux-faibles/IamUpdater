package main

//import (
//	"golang.org/x/exp/constraints"
//	"strings"
//	"testing"
//)
//
//type StructureComparable struct {
//	label string
//}
//
//func (s *StructureComparable) index() int {
//	return len(s.label)
//}
//
//func Test_ordonne(t *testing.T) {
//	tableau1 := []StructureComparable{{"fff"}, {"zzz"}, {"aaa"}}
//	ordonne(tableau1)
//}

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
