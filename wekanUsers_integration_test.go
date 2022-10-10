//go:build integration
// +build integration

// nolint:errcheck
package main

import (
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWekan_ManageUsers_withoutScopeWekan(t *testing.T) {
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
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

	err := pipeline.StopAfter(wekan, usersWithoutScopeWekan.selectScopeWekan(), stageManageUsers)
	ass.NoError(err)
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.IsType(libwekan.UserNotFoundError{}, actualErr)
	ass.Empty(actualUser)
}

func TestWekan_ManageUsers_withScopeWekan(t *testing.T) {
	ass := assert.New(t)
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	// GIVEN
	usernameDeTest := Username("wekan_user")
	usersWithScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"wekan"},
			email: usernameDeTest,
		},
	}

	// WHEN
	err := pipeline.StopAfter(wekan, usersWithScopeWekan, stageManageUsers)

	// THEN
	ass.NoError(err)
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.Nil(actualErr)
	ass.NotEmpty(actualUser)
}

func TestWekan_ManageUsers_removeScopeWekan(t *testing.T) {
	// GIVEN
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	usersWithScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"wekan"},
			email: usernameDeTest,
		},
	}
	pipeline.StopAfter(wekan, usersWithScopeWekan, stageManageUsers)

	usersWithoutScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"not_wekan"},
			email: usernameDeTest,
			//boards: []string{"tableau-crp-bfc"},
		},
	}
	// WHEN
	err := pipeline.StopAfter(wekan, usersWithoutScopeWekan.selectScopeWekan(), stageManageUsers)
	ass.NoError(err)

	// THEN
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.Nil(actualErr)
	ass.NotEmpty(actualUser)
	ass.True(actualUser.LoginDisabled)
}
