package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWekan_ListBoards(t *testing.T) {
	ass := assert.New(t)

	users := Users{
		"user1": User{
			email:  "user1",
			boards: []string{"tableau1", "tableau2"},
		},
		"user2": User{
			email:  "user2",
			boards: []string{"tableau2"},
		},
		"user3": User{
			email:  "user3",
			boards: []string{""},
		},
	}

	boards := users.listBoards()
	ass.Len(boards["tableau1"], 1)
	ass.Len(boards["tableau2"], 2)
	ass.Len(boards, 2)
}
