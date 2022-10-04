//go:build integration
// +build integration

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

	err := pipeline.StopAfter(wekan, usersWithoutScopeWekan, StageManageUsers)
	ass.Nil(err)
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.IsType(libwekan.UnknownUserError{}, actualErr)
	ass.Empty(actualUser)
}

func TestWekan_ManageUsers_withScopeWekan(t *testing.T) {
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	ass := assert.New(t)
	usernameDeTest := Username("wekan_user")
	usersWithScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"wekan"},
			email: usernameDeTest,
		},
	}

	err := pipeline.StopAfter(wekan, usersWithScopeWekan, StageManageUsers)
	ass.Nil(err)
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
	pipeline.StopAfter(wekan, usersWithScopeWekan, StageManageUsers)

	usersWithoutScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"not_wekan"},
			email: usernameDeTest,
			//boards: []string{"tableau-crp-bfc"},
		},
	}
	// WHEN
	err := pipeline.StopAfter(wekan, usersWithoutScopeWekan, StageManageUsers)
	ass.Nil(err)

	// THEN
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.Nil(actualErr)
	ass.NotEmpty(actualUser)
	ass.True(actualUser.LoginDisabled)
}
