//go:build integration
// +build integration

// nolint:errcheck

package main

import (
	"testing"

	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
)

func TestWekan_ManageBoardsMembers_withoutBoard(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	boardDeTest := "tableau-crp-bfc"
	usersWithoutBoards := Users{
		usernameDeTest: User{
			scope:  []string{"wekan"},
			email:  usernameDeTest,
			boards: []string{},
		},
	}

	// THEN
	err := pipeline.StopAfter(wekan, usersWithoutBoards, stageManageBoardsMembers)
	ass.NoError(err)
	actualUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	actualBFCBoard, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	ass.False(actualBFCBoard.UserIsMember(actualUser))
}

func TestWekan_ManageBoardsMembers_withBoard(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	boardDeTest := "tableau-crp-bfc"
	usersWithBoards := Users{
		usernameDeTest: User{
			scope:  []string{"wekan"},
			email:  usernameDeTest,
			boards: []string{boardDeTest},
		},
	}

	// THEN
	err := pipeline.StopAfter(wekan, usersWithBoards, stageManageBoardsMembers)
	ass.NoError(err)
	actualUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	actualBFCBoard, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	ass.True(actualBFCBoard.UserIsActiveMember(actualUser))
}

func TestWekan_ManageBoardsMembers_removeFromBoard(t *testing.T) {
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	boardDeTest := "tableau-crp-bfc"
	usersWithBoards := Users{
		usernameDeTest: User{
			scope:  []string{"wekan"},
			email:  usernameDeTest,
			boards: []string{boardDeTest},
		},
	}
	pipeline.StopAfter(wekan, usersWithBoards, stageManageBoardsMembers)
	usersWithoutBoards := Users{
		usernameDeTest: User{
			scope:  []string{"wekan"},
			email:  usernameDeTest,
			boards: []string{},
		},
	}

	// THEN
	err := pipeline.StopAfter(wekan, usersWithoutBoards, stageManageBoardsMembers)
	ass.NoError(err)
	actualUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	actualBFCBoard, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	ass.True(actualBFCBoard.UserIsMember(actualUser))
	ass.False(actualBFCBoard.UserIsActiveMember(actualUser))
}
