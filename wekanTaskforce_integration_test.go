//go:build integration
// +build integration

package main

import (
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_ManageBoardsLabelsTaskforce_whenEverythingsFine(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)

	usernameDeTest := Username("wekanUser")
	labelDeTest := "label_test"
	boardDeTest := "tableau-crp-bfc"

	users := Users{
		usernameDeTest: User{
			scope:     []string{"wekan"},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{labelDeTest},
		},
	}

	boardLabel := libwekan.NewBoardLabel(labelDeTest, "red")
	board, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	wekan.InsertBoardLabel(ctx, board, boardLabel)
	ManageUsers(wekan, users)
	ManageBoardsMembers(wekan, users)

	//THEN
	err := ManageBoardsLabelsTaskforce(wekan, users)
	ass.Nil(err)
	rules, err := wekan.SelectRulesFromBoardID(ctx, board.ID)
	ass.Nil(err)
	ass.Len(rules, 1)
}

func Test_ManageBoardsLabelsTaskforce_withoutScopeWekan(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)
	req := require.New(t)

	usernameDeTest := Username("no_wekan")
	labelDeTest := "label_test"
	boardDeTest := "tableau-crp-bfc"

	usersWithoutScopeWekan := Users{
		usernameDeTest: User{
			scope:     []string{"not_wekan"},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{labelDeTest},
		},
	}
	boardLabel := libwekan.NewBoardLabel(labelDeTest, "red")
	board, err := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	require.Nil(t, err)
	err = wekan.InsertBoardLabel(ctx, board, boardLabel)
	require.Nil(t, err)
	ManageUsers(wekan, usersWithoutScopeWekan)
	ManageBoardsMembers(wekan, usersWithoutScopeWekan)

	// THEN
	err = ManageBoardsLabelsTaskforce(wekan, usersWithoutScopeWekan)
	ass.Nil(err)
	rules, err := wekan.SelectRulesFromBoardID(ctx, board.ID)
	req.Nil(err)
	ass.Len(rules, 0)
}

func Test_ManageBoardsLabelsTaskforce_withUserWithoutTaskForce(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)
	req := require.New(t)
	usernameDeTest := Username("with_wekan")
	labelDeTest := "red"
	boardDeTest := "tableau-crp-bfc"
	usersWithTaskForce := Users{
		usernameDeTest: User{
			scope:     []string{"wekan"},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{},
		},
	}
	boardLabel := libwekan.NewBoardLabel(labelDeTest, labelDeTest)
	board, err := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	req.Nil(err)
	err = wekan.InsertBoardLabel(ctx, board, boardLabel)
	req.Nil(err)
	ManageUsers(wekan, usersWithTaskForce)
	ManageBoardsMembers(wekan, usersWithTaskForce)

	// THEN
	err = ManageBoardsLabelsTaskforce(wekan, usersWithTaskForce)
	ass.Nil(err)
	rules, _ := wekan.SelectRulesFromBoardID(ctx, board.ID)
	ass.Len(rules, 0)
}

func Test_ManageBoardsLabelsTaskforce_withRuleRemoval(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)

	usernameDeTest := Username("wekanUser")
	labelDeTest := "label_test"
	boardDeTest := "tableau-crp-bfc"

	users := Users{
		usernameDeTest: User{
			scope:     []string{"wekan"},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{labelDeTest},
		},
	}

	boardLabel := libwekan.NewBoardLabel(labelDeTest, "red")
	board, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	wekan.InsertBoardLabel(ctx, board, boardLabel)
	ManageUsers(wekan, users)
	ManageBoardsMembers(wekan, users)
	ManageBoardsLabelsTaskforce(wekan, users)

	//THEN
	users = Users{
		usernameDeTest: User{
			scope:     []string{"wekan"},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{},
		},
	}
	err := ManageBoardsLabelsTaskforce(wekan, users)
	ass.Nil(err)
	rules, _ := wekan.SelectRulesFromBoardID(ctx, board.ID)
	ass.Len(rules, 0)
}

func Test_ManageBoardsLabelsTaskforce_withScopeWekanRemoval(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)

	usernameDeTest := Username("wekanUser")
	labelDeTest := "label_test"
	boardDeTest := "tableau-crp-bfc"

	users := Users{
		usernameDeTest: User{
			scope:     []string{"wekan"},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{labelDeTest},
		},
	}

	boardLabel := libwekan.NewBoardLabel(labelDeTest, "red")
	board, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	wekan.InsertBoardLabel(ctx, board, boardLabel)
	ManageUsers(wekan, users)
	ManageBoardsMembers(wekan, users)
	ManageBoardsLabelsTaskforce(wekan, users)

	//THEN
	users = Users{
		usernameDeTest: User{
			scope:     []string{},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{labelDeTest},
		},
	}
	err := ManageBoardsLabelsTaskforce(wekan, users)
	ass.Nil(err)
	rules, _ := wekan.SelectRulesFromBoardID(ctx, board.ID)
	ass.Len(rules, 0)
}

func Test_ManageBoardsLabelsTaskforce_withBoardDeTestRemoval(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)

	usernameDeTest := Username("wekanUser")
	labelDeTest := "label_test"
	boardDeTest := "tableau-crp-bfc"

	users := Users{
		usernameDeTest: User{
			scope:     []string{"wekan"},
			email:     usernameDeTest,
			boards:    []string{},
			taskforce: []string{labelDeTest},
		},
	}

	boardLabel := libwekan.NewBoardLabel(labelDeTest, "red")
	board, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	wekan.InsertBoardLabel(ctx, board, boardLabel)
	ManageUsers(wekan, users)
	ManageBoardsMembers(wekan, users)
	ManageBoardsLabelsTaskforce(wekan, users)

	//THEN
	users = Users{
		usernameDeTest: User{
			scope:     []string{},
			email:     usernameDeTest,
			boards:    []string{boardDeTest},
			taskforce: []string{labelDeTest},
		},
	}
	err := ManageBoardsLabelsTaskforce(wekan, users)
	ass.Nil(err)
	rules, _ := wekan.SelectRulesFromBoardID(ctx, board.ID)
	ass.Len(rules, 0)
}
