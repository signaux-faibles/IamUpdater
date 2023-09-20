//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fake faker.Faker

func init() {
	fake = faker.New()
}

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
	err := pipeline.StopAfter(wekan, usersWithScopeWekan, stageManageUsers)
	ass.NoError(err)

	usersWithoutScopeWekan := Users{
		usernameDeTest: User{
			scope: []string{"not_wekan"},
			email: usernameDeTest,
			//boards: []string{"tableau-crp-bfc"},
		},
	}
	// WHEN
	err = pipeline.StopAfter(wekan, usersWithoutScopeWekan.selectScopeWekan(), stageManageUsers)
	ass.NoError(err)

	// THEN
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(usernameDeTest))
	ass.Nil(actualErr)
	ass.NotEmpty(actualUser)
	ass.True(actualUser.LoginDisabled)
}

func TestWekan_ManageUsers_withScopeWekanAlreadyExists(t *testing.T) {
	ass := assert.New(t)
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	// GIVEN

	// construction d'un user legacy
	user := libwekan.BuildUser("legacy@tests.unitaires", "LU", "User LEGACY")
	user.AuthenticationMethod = "password"
	user.Services = libwekan.UserServices{
		Password: libwekan.UserServicesPassword{
			Bcrypt: fake.BinaryString().BinaryString(512),
		},
	}
	user.Username = libwekan.Username(user.Profile.Fullname)
	// insertion en base
	ctx := context.Background()
	err := wekan.InsertUser(ctx, user)
	require.NoError(t, err)

	legacy := Username("legacy@tests.unitaires")
	legacyUser := Users{
		legacy: User{
			scope:  []string{"wekan"},
			email:  legacy,
			prenom: "user",
			nom:    "legacy",
		},
	}

	err = pipeline.StopAfter(wekan, legacyUser, stageManageUsers)
	//require.NoError(t, err)
	// WHEN

	// THEN
	ass.NoError(err)
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, "legacy@tests.unitaires")
	ass.Nil(actualErr)
	ass.NotEmpty(actualUser)
}
