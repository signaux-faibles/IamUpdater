//go:build integration
// +build integration

// nolint:errcheck
package main

import (
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

//func TestWekan_WekanUpdate(t *testing.T) {
//  req := require.New(t)
//  req.Nil(globalWekan.AssertHasAdmin(context.Background()))
//  adminUser, _ := globalWekan.GetUserFromUsername(ctx, globalWekan.AdminUsername())
//  req.False(adminUser.LoginDisabled)
//  absentUser, err := globalWekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
//  req.IsType(libwekan.UnknownUserError{}, err)
//  req.Empty(absentUser)
//  err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/1.xlsx")
//  req.Nil(err)
//  insertedRaphaelUser, _ := globalWekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
//  req.Equal("raphael.squelbut@shodo.io", string(insertedRaphaelUser.Username))
//
//  notInsertedJohnUser, err := globalWekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
//  req.Empty(notInsertedJohnUser)
//  req.IsType(libwekan.UnknownUserError{}, err)
//
//  insertedHerbertUser, _ := globalWekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
//  req.NotEmpty(insertedHerbertUser)
//
//  bfcBoard, _ := globalWekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
//  req.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
//  req.True(bfcBoard.UserIsActiveMember(insertedHerbertUser))
//  req.True(bfcBoard.UserIsActiveMember(adminUser))
//
//  nordBoard, _ := globalWekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
//  req.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
//  req.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
//  req.True(nordBoard.UserIsActiveMember(adminUser))
//
//  adminUser, _ = globalWekan.GetUserFromUsername(ctx, libwekan.Username(globalWekan.AdminUsername()))
//  req.False(adminUser.LoginDisabled)
//
//  //
//  // effectue les mêmes tests avec un autre fichier excel
//  //
//  err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/2.xlsx")
//  req.Nil(err)
//
//  notInsertedJohnUser, err = globalWekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
//  req.Empty(notInsertedJohnUser)
//  req.IsType(libwekan.UnknownUserError{}, err)
//
//  insertedRaphaelUser, _ = globalWekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
//  insertedHerbertUser, _ = globalWekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
//  insertedFranckUser, _ := globalWekan.GetUserFromUsername(context.Background(), "franck.michael@pantheon.fr")
//
//  bfcBoard, _ = globalWekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
//  req.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
//  req.False(bfcBoard.UserIsActiveMember(insertedHerbertUser))
//  req.False(bfcBoard.UserIsActiveMember(insertedFranckUser))
//  req.True(bfcBoard.UserIsActiveMember(adminUser))
//
//  nordBoard, _ = globalWekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
//  req.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
//  req.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
//  req.False(nordBoard.UserIsActiveMember(insertedFranckUser))
//  req.True(bfcBoard.UserIsActiveMember(adminUser))
//}

func TestWekan_ManageUsers_withoutScopeWekan(t *testing.T) {
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)
	usersWithoutScopeWekan := make(Users)
	usernameDeTest := Username("no_wekan")
	usersWithoutScopeWekan = Users{
		usernameDeTest: User{
			scope:  []string{"not_wekan"},
			email:  usernameDeTest,
			boards: []string{"tableau-crp-bfc"},
		},
	}

	err := ManageUsers(wekan, usersWithoutScopeWekan)
	ass.Nil(err)
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.IsType(libwekan.UnknownUserError{}, actualErr)
	ass.Empty(actualUser)
}

func TestWekan_ManageUsers_withScopeWekan(t *testing.T) {
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	usersWithScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"wekan"},
			email: usernameDeTest,
			//boards: []string{"tableau-crp-bfc"},
		},
	}

	err := ManageUsers(wekan, usersWithScopeWekan)
	ass.Nil(err)
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.Nil(actualErr)
	ass.NotEmpty(actualUser)
}

func TestWekan_ManageUsers_removeScopeWekan(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	usersWithScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"wekan"},
			email: usernameDeTest,
		},
	}
	ManageUsers(wekan, usersWithScopeWekan)

	// THEN
	usersWithoutScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"not_wekan"},
			email: usernameDeTest,
			//boards: []string{"tableau-crp-bfc"},
		},
	}
	err := ManageUsers(wekan, usersWithoutScopeWekan)
	ass.Nil(err)
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.Nil(actualErr)
	ass.NotEmpty(actualUser)
	ass.True(actualUser.LoginDisabled)
}

func TestWekan_ManageBoardsMembers_withoutBoard(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
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
	ManageUsers(wekan, usersWithoutBoards)

	// THEN
	err := ManageBoardsMembers(wekan, usersWithoutBoards)
	ass.Nil(err)
	actualUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	actualBFCBoard, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	ass.False(actualBFCBoard.UserIsMember(actualUser))
}

func TestWekan_ManageBoardsMembers_withBoard(t *testing.T) {
	// WHEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
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
	ManageUsers(wekan, usersWithBoards)

	// THEN
	err := ManageBoardsMembers(wekan, usersWithBoards)
	ass.Nil(err)
	actualUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	actualBFCBoard, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	ass.True(actualBFCBoard.UserIsActiveMember(actualUser))
}

func TestWekan_ManageBoardsMembers_removeFromBoard(t *testing.T) {
	wekan := restoreMongoDumpInDatabase(mongodb, "", t)
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	usernameEnPlus := Username("dummy_user")
	boardDeTest := "tableau-crp-bfc"
	usersWithBoards := Users{
		usernameDeTest: User{
			scope:  []string{"wekan"},
			email:  usernameDeTest,
			boards: []string{boardDeTest},
		},
		usernameEnPlus: User{
			scope:  []string{"wekan"},
			email:  usernameEnPlus,
			boards: []string{boardDeTest},
		},
	}
	ManageUsers(wekan, usersWithBoards)
	ManageBoardsMembers(wekan, usersWithBoards)
	usersWithoutBoards := Users{
		usernameDeTest: User{
			scope:  []string{"wekan"},
			email:  usernameDeTest,
			boards: []string{},
		},
		usernameEnPlus: User{
			scope:  []string{"wekan"},
			email:  usernameEnPlus,
			boards: []string{boardDeTest},
		},
	}
	ManageUsers(wekan, usersWithoutBoards)

	// THEN
	err := ManageBoardsMembers(wekan, usersWithoutBoards)
	ass.Nil(err)
	actualUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	actualBFCBoard, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	ass.True(actualBFCBoard.UserIsMember(actualUser))
	ass.False(actualBFCBoard.UserIsActiveMember(actualUser))
}

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
