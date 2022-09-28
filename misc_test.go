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
	"testing"

	"github.com/stretchr/testify/assert"
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

func Test_mapSelect(t *testing.T) {
	ass := assert.New(t)
	m := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
	}
	expected := map[string]int{
		"a": 1,
		"b": 2,
	}
	m1 := mapSelect(m, func(i string, e int) bool {
		return e < 3
	})
	ass.Equal(expected, m1)
	m2 := mapSelect(m, func(i string, b int) bool {
		return i == "a" || i == "b"
	})
	ass.Equal(expected, m2)
}

func Test_mapSlice(t *testing.T) {
	ass := assert.New(t)
	array := []string{"a", "aa", "aaa", "aaaa"}
	result := mapSlice(array, func(e string) int { return len(e) })
	ass.Equal([]int{1, 2, 3, 4}, result)
}

func Test_mapifySlice(t *testing.T) {
	ass := assert.New(t)
	array := []string{"a", "aa", "aaa", "aaaa"}
	expected := map[int]string{
		1: "a",
		2: "aa",
		3: "aaa",
		4: "aaaa",
	}
	result := mapifySlice(array, func(e string) int { return len(e) })
	ass.Equal(expected, result)
}

func Test_contains(t *testing.T) {
	ass := assert.New(t)
	array := []string{"a", "b", "c", "d"}
	ass.False(contains(array, "e"))
	ass.True(contains(array, "a"))
}
