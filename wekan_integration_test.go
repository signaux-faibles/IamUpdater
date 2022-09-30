package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
)

func TestWekan_WekanUpdate(t *testing.T) {
	req := require.New(t)
	req.Nil(globalWekan.AssertHasAdmin(context.Background()))
	adminUser, _ := globalWekan.GetUserFromUsername(ctx, globalWekan.AdminUsername())
	req.False(adminUser.LoginDisabled)
	absentUser, err := globalWekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	req.IsType(libwekan.UnknownUserError{}, err)
	req.Empty(absentUser)
	err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/1.xlsx")
	req.Nil(err)
	insertedRaphaelUser, _ := globalWekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	req.Equal("raphael.squelbut@shodo.io", string(insertedRaphaelUser.Username))

	notInsertedJohnUser, err := globalWekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
	req.Empty(notInsertedJohnUser)
	req.IsType(libwekan.UnknownUserError{}, err)

	insertedHerbertUser, _ := globalWekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
	req.NotEmpty(insertedHerbertUser)

	bfcBoard, _ := globalWekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
	req.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
	req.True(bfcBoard.UserIsActiveMember(insertedHerbertUser))
	req.True(bfcBoard.UserIsActiveMember(adminUser))

	nordBoard, _ := globalWekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
	req.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
	req.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
	req.True(nordBoard.UserIsActiveMember(adminUser))

	adminUser, _ = globalWekan.GetUserFromUsername(ctx, libwekan.Username(globalWekan.AdminUsername()))
	req.False(adminUser.LoginDisabled)

	//
	// effectue les mêmes tests avec un autre fichier excel
	//
	err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/2.xlsx")
	req.Nil(err)

	notInsertedJohnUser, err = globalWekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
	req.Empty(notInsertedJohnUser)
	req.IsType(libwekan.UnknownUserError{}, err)

	insertedRaphaelUser, _ = globalWekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	insertedHerbertUser, _ = globalWekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
	insertedFranckUser, _ := globalWekan.GetUserFromUsername(context.Background(), "franck.michael@pantheon.fr")

	bfcBoard, _ = globalWekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
	req.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
	req.False(bfcBoard.UserIsActiveMember(insertedHerbertUser))
	req.False(bfcBoard.UserIsActiveMember(insertedFranckUser))
	req.True(bfcBoard.UserIsActiveMember(adminUser))

	nordBoard, _ = globalWekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
	req.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
	req.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
	req.False(nordBoard.UserIsActiveMember(insertedFranckUser))
	req.True(bfcBoard.UserIsActiveMember(adminUser))
}

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
	actualUser, actualErr := globalWekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
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
