//go:build integration

package main

import (
	"context"
	"testing"

	"github.com/jaswdr/faker"
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"keycloakUpdater/v2/pkg/logger"
	"keycloakUpdater/v2/pkg/structs"
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

func TestWekan_ManageUsers_whenAlreadyExists(t *testing.T) {
	ass := assert.New(t)
	logger.ConfigureWith(structs.LoggerConfig{Filename: "/dev/null", Level: "ERROR"})
	wekan := restoreMongoDumpInDatabase(mongodb, "", t, "")
	// GIVEN
	email := "legacy@tests.unitaires"

	// construction d'un user legacy
	user := libwekan.BuildUser(email, "LU", "User LEGACY")
	user.Emails[0].Verified = false
	user.AuthenticationMethod = "password"
	user.Services = libwekan.UserServices{
		Password: libwekan.UserServicesPassword{
			Bcrypt: fake.BinaryString().BinaryString(512),
		},
	}
	user.Username = libwekan.Username(user.Profile.Fullname)
	// insertion en base de user legacy
	ctx := context.Background()
	err := wekan.InsertUser(ctx, user)
	require.NoError(t, err)

	// on simule 3 users dont 1 qui a un email existant en base
	username1 := Username(fake.Internet().Email())
	legacyUsername := Username(email)
	username3 := Username(fake.Internet().Email())
	excelUser := Users{
		username1: User{
			scope: []string{"wekan"},
			email: username1,
		},
		legacyUsername: User{
			scope:  []string{"wekan"},
			email:  legacyUsername,
			prenom: "user",
			nom:    "legacy",
		},
		username3: User{
			scope: []string{"wekan"},
			email: username3,
		},
	}

	err = pipeline.StopAfter(wekan, excelUser, stageManageUsers)
	require.NoError(t, err)
	// WHEN

	// THEN
	// le 1er et le 3e user sont en bases
	// le second qui a un email déjà présent non
	actualUser, actualErr := wekan.GetUserFromUsername(ctx, libwekan.Username(username1))
	ass.NoError(actualErr)
	ass.NotEmpty(actualUser)

	actualUser, actualErr = wekan.GetUserFromUsername(ctx, libwekan.Username(legacyUsername))
	ass.Empty(actualUser)
	ass.Error(actualErr)

	actualUser, actualErr = wekan.GetUserFromUsername(ctx, libwekan.Username(username3))
	ass.NoError(actualErr)
	ass.NotEmpty(actualUser)
}
