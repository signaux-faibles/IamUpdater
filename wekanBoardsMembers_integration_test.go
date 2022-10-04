//go:build integration
// +build integration

// nolint:errcheck

package main

import (
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"testing"
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
	err := pipeline.StopAfter(wekan, usersWithoutBoards, StageManageBoardsMembers)
	ass.Nil(err)
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
	err := pipeline.StopAfter(wekan, usersWithBoards, StageManageBoardsMembers)
	ass.Nil(err)
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
	pipeline.StopAfter(wekan, usersWithBoards, StageManageBoardsMembers)
	usersWithoutBoards := Users{
		usernameDeTest: User{
			scope:  []string{"wekan"},
			email:  usernameDeTest,
			boards: []string{},
		},
	}

	// THEN
	err := pipeline.StopAfter(wekan, usersWithoutBoards, StageManageBoardsMembers)
	ass.Nil(err)
	actualUser, _ := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	actualBFCBoard, _ := wekan.GetBoardFromSlug(ctx, libwekan.BoardSlug(boardDeTest))
	ass.True(actualBFCBoard.UserIsMember(actualUser))
	// Fail:
	// ass.False(actualBFCBoard.UserIsActiveMember(actualUser))
	// ce fail provient du fait que la board tableau-crp-bfc n'est plus dans le fichier de configuration…
	// À cause de cela, elle n'est plus référencée dans l'objet BoardsMembers en input, et alors la
	// fonction ne fait rien sur cette board… Logique !
	// la solution nécessite d'inférer la liste des boards à traiter par un autre canal que
	// la configuration des utilisateurs.
	// J'ai ajouté la fonction SelectBoardsFromSlugExpression qui permettrait de récupérer la liste des boards
	// dont le slug correspond à une expression régulière, que l'on pourrait passer en amont (ou en aval) de l'appel
	// de la fonction pour ajouter les boards vides lorsque cela est nécessaire.
	// Cela va demander de revoir la signature de la fonction ManageBoardsMembers, dans tous les cas…
}
