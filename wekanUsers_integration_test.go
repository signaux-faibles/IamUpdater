//go:build integration
// +build integration

package main

import (
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
