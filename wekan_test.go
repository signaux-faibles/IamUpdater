package main

import (
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWekan_ListBoards(t *testing.T) {
	// WHEN
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

	// THEN
	boards := users.inferBoardsMember()
	ass.Len(boards["tableau1"], 1)
	ass.Len(boards["tableau2"], 2)
	ass.Len(boards, 2)
}

func TestWekan_AddBoards_whenNewBoard(t *testing.T) {
	// WHEN
	ass := assert.New(t)
	boardsMembers := BoardsMembers{
		"tableau1": Users{},
	}
	addedBoardSlug := libwekan.BoardSlug("tableau2")
	boards := []libwekan.Board{{Slug: addedBoardSlug}}

	// THEN
	addedBoardsMembers := boardsMembers.addBoards(boards)
	_, ok := addedBoardsMembers[addedBoardSlug]
	ass.True(ok)
	ass.Len(addedBoardsMembers, 2)
}

func TestWekan_AddBoards_whenExistingBoard(t *testing.T) {
	// WHEN
	ass := assert.New(t)
	addedBoardSlug := libwekan.BoardSlug("tableau")
	boards := []libwekan.Board{{Slug: addedBoardSlug}}
	testUser := User{email: "testUser"}
	boardsMembers := BoardsMembers{
		addedBoardSlug: Users{
			testUser.email: testUser,
		},
	}

	// THEN
	addedBoardsMembers := boardsMembers.addBoards(boards)
	_, ok := addedBoardsMembers[addedBoardSlug]
	ass.True(ok)
	ass.Equal(boardsMembers, addedBoardsMembers)
	ass.Len(addedBoardsMembers, 1)
	ass.Equal(testUser.email, addedBoardsMembers[addedBoardSlug][testUser.email].email)
}

func TestWekan_AddBoards_whenEmptyBoardsMembers(t *testing.T) {
	// WHEN
	boardsMembers := make(BoardsMembers)
	addedBoardSlug := libwekan.BoardSlug("tableau")
	boards := []libwekan.Board{{Slug: addedBoardSlug}}
	require.Len(t, boardsMembers, 0)

	// THEN
	addedBoardsMembers := boardsMembers.addBoards(boards)
	_, ok := addedBoardsMembers[addedBoardSlug]
	assert.True(t, ok)
}

func TestWekan_AddBoards_whenNilBoardsMembers(t *testing.T) {
	// WHEN
	var boardsMembers BoardsMembers
	addedBoardSlug := libwekan.BoardSlug("tableau")
	boards := []libwekan.Board{{Slug: addedBoardSlug}}
	require.Nil(t, boardsMembers)

	// THEN
	addedBoardsMembers := boardsMembers.addBoards(boards)
	_, ok := addedBoardsMembers[addedBoardSlug]
	assert.True(t, ok)
}

func TestWekan_AddBoards_whenEmptyBoards(t *testing.T) {
	// WHEN
	boardsMembers := BoardsMembers{
		"tableau1": Users{},
	}
	boards := []libwekan.Board{}
	require.Empty(t, boards)

	// THEN
	addedBoardsMembers := boardsMembers.addBoards(boards)
	assert.Len(t, addedBoardsMembers, 1)
}

func TestWekan_AddBoards_whenNilBoards(t *testing.T) {
	// WHEN
	boardsMembers := BoardsMembers{
		"tableau1": Users{},
	}
	var boards []libwekan.Board
	require.Nil(t, boards)

	// THEN
	addedBoardsMembers := boardsMembers.addBoards(boards)
	assert.Len(t, addedBoardsMembers, 1)
}
