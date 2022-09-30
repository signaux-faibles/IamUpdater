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
	req.Nil(wekan.AssertHasAdmin(context.Background()))
	adminUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(wekan.AdminUsername()))
	req.False(adminUser.LoginDisabled)
	absentUser, err := wekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	req.IsType(libwekan.UnknownUserError{}, err)
	req.Empty(absentUser)
	err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/1.xlsx")
	req.Nil(err)
	insertedRaphaelUser, _ := wekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	req.Equal("raphael.squelbut@shodo.io", string(insertedRaphaelUser.Username))

	notInsertedJohnUser, err := wekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
	req.Empty(notInsertedJohnUser)
	req.IsType(libwekan.UnknownUserError{}, err)

	insertedHerbertUser, _ := wekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
	req.NotEmpty(insertedHerbertUser)

	bfcBoard, _ := wekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
	req.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
	req.True(bfcBoard.UserIsActiveMember(insertedHerbertUser))
	req.True(bfcBoard.UserIsActiveMember(adminUser))

	nordBoard, _ := wekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
	req.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
	req.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
	req.True(nordBoard.UserIsActiveMember(adminUser))

	adminUser, _ = wekan.GetUserFromUsername(ctx, libwekan.Username(wekan.AdminUsername()))
	req.False(adminUser.LoginDisabled)

	//
	// effectue les mêmes tests avec un autre fichier excel
	//
	err = WekanUpdate(mongoUrl, "wekan", "signaux.faibles", "test/resources/wekanUpdate_states/2.xlsx")
	req.Nil(err)

	notInsertedJohnUser, err = wekan.GetUserFromUsername(context.Background(), "john.doe@zone51.gov.fr")
	req.Empty(notInsertedJohnUser)
	req.IsType(libwekan.UnknownUserError{}, err)

	insertedRaphaelUser, _ = wekan.GetUserFromUsername(context.Background(), "raphael.squelbut@shodo.io")
	insertedHerbertUser, _ = wekan.GetUserFromUsername(context.Background(), "herbert.leonard@pantheon.fr")
	insertedFranckUser, _ := wekan.GetUserFromUsername(context.Background(), "franck.michael@pantheon.fr")

	bfcBoard, _ = wekan.GetBoardFromSlug(context.Background(), "tableau-crp-bfc")
	req.True(bfcBoard.UserIsActiveMember(insertedRaphaelUser))
	req.False(bfcBoard.UserIsActiveMember(insertedHerbertUser))
	req.False(bfcBoard.UserIsActiveMember(insertedFranckUser))
	req.True(bfcBoard.UserIsActiveMember(adminUser))

	nordBoard, _ = wekan.GetBoardFromSlug(context.Background(), "tableau-codefi-nord")
	req.False(nordBoard.UserIsActiveMember(insertedRaphaelUser))
	req.True(nordBoard.UserIsActiveMember(insertedHerbertUser))
	req.False(nordBoard.UserIsActiveMember(insertedFranckUser))
	req.True(bfcBoard.UserIsActiveMember(adminUser))
}

func TestWekan_ListBoards(t *testing.T) {
	ass := assert.New(t)
	boards := excelUsers1.listBoards()
	ass.Len(boards["tableau-crp-bfc"], 3)
	ass.Len(boards["tableau-codefi-nord"], 1)
	ass.Len(boards, 2)
}
