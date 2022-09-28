//go:build integration
// +build integration

package main

import (
	"context"
	"testing"

	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
)

func TestWekan_WekanUpdate(t *testing.T) {
	ass := assert.New(t)
	adminUser, _ := wekan.AdminUser(context.Background())
	ass.False(adminUser.LoginDisabled)
	absentUser, err := wekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	ass.IsType(libwekan.UnknownUserError{}, err)
	ass.Empty(absentUser)
	err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/1.xlsx")
	ass.Nil(err)
	insertedRaphaelUser, _ := wekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	ass.Equal("raphael.squelbut@shodo.io", string(insertedRaphaelUser.Username))

	notInsertedJohnUser, err := wekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
	ass.Empty(notInsertedJohnUser)
	ass.IsType(libwekan.UnknownUserError{}, err)

	insertedHerbertUser, _ := wekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
	ass.NotEmpty(insertedHerbertUser)

	bfcBoard, _ := wekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
	ass.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
	ass.True(bfcBoard.UserIsActiveMember(insertedHerbertUser))
	ass.True(bfcBoard.UserIsActiveMember(adminUser))

	nordBoard, _ := wekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
	ass.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
	ass.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
	ass.True(bfcBoard.UserIsActiveMember(adminUser))

	adminUser, _ = wekan.AdminUser(context.Background())
	ass.False(adminUser.LoginDisabled)
	err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/2.xlsx")
	ass.Nil(err)

	notInsertedJohnUser, err = wekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
	ass.Empty(notInsertedJohnUser)
	ass.IsType(libwekan.UnknownUserError{}, err)

	insertedRaphaelUser, _ = wekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	insertedHerbertUser, _ = wekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
	insertedFranckUser, _ := wekan.GetUserFromUsername(context.Background(), "franck.michael@pantheon.fr")

	bfcBoard, _ = wekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
	ass.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
	ass.False(bfcBoard.UserIsActiveMember(insertedHerbertUser))
	ass.False(bfcBoard.UserIsActiveMember(insertedFranckUser))
	ass.True(bfcBoard.UserIsActiveMember(adminUser))

	nordBoard, _ = wekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
	ass.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
	ass.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
	ass.False(nordBoard.UserIsActiveMember(insertedFranckUser))
	ass.True(bfcBoard.UserIsActiveMember(adminUser))
}

func TestWekan_ListBoards(t *testing.T) {
	ass := assert.New(t)
	boards := excelUsers1.listBoards()
	ass.Len(boards["tableau-crp-bfc"], 2)
	ass.Len(boards["tableau-codefi-nord"], 1)
	ass.Len(boards, 2)
}
