package main

import (
	"context"
	"fmt"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
	"strings"
)

// ManageUsers
// - objectif de traiter les utilisateurs Wekan
// - création des utilisateurs inconnus dans Wekan
// - désactivation des utilisateurs superflus
func ManageUsers(wekan libwekan.Wekan, fromConfig Users) error {
	wekanUsersfromConfig := fromConfig.selectScopeWekan()
	addAdmin(wekanUsersfromConfig, wekan)

	fromWekan, err := wekan.GetUsers(context.TODO())
	if err != nil {
		return err
	}

	creations, enable, disable := wekanUsersfromConfig.ListWekanChanges(fromWekan)

	fields := logger.DataForMethod("InsertUsers")
	fields.AddAny("population", len(creations))
	logger.Info("inscription des nouveaux utilisateurs", fields)
	err = InsertUsers(context.Background(), wekan, creations)
	if err != nil {
		return err
	}

	fields = logger.DataForMethod("EnableUsers")
	fields.AddAny("population", len(enable))
	logger.Info("réactivation des utilisateurs inactifs", fields)
	err = EnableUsers(context.Background(), wekan, enable)
	if err != nil {
		return err
	}

	fields = logger.DataForMethod("DisableUsers")
	fields.AddAny("population", len(disable))
	logger.Info("désactivation des utilisateurs supprimés", fields)
	return DisableUsers(context.Background(), wekan, disable)
}

func InsertUsers(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.DataForMethod("InsertUser")
	if err := wekan.AssertPrivileged(ctx); err != nil {
		return err
	}

	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Debug("insertion de l'utilisateur", fields)
		err := wekan.InsertUser(ctx, user)
		if err != nil {
			logger.Error(err.Error(), fields)
			return err
		}
	}
	return nil
}

func EnableUsers(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.DataForMethod("EnableUser")
	if err := wekan.AssertPrivileged(ctx); err != nil {
		return err
	}

	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Debug("activation de l'utilisateur", fields)
		err := wekan.EnableUser(ctx, user)
		if err != nil {
			logger.Error(err.Error(), fields)
			return err
		}
	}
	return nil
}

func DisableUsers(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.DataForMethod("DisableUser")
	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Debug("désactivation de l'utilisateur", fields)
		err := wekan.DisableUser(ctx, user)
		if err != nil {
			logger.Error(err.Error(), fields)
			return err
		}
	}
	return nil
}

func (users Users) BuildWekanUsers() libwekan.Users {
	var wekanUsers libwekan.Users
	for _, user := range users {
		initials := firstChar(user.prenom) + firstChar(user.nom)
		fullname := fmt.Sprintf("%s %s", strings.ToUpper(user.nom), user.prenom)
		wekanUser := libwekan.BuildUser(string(user.email), initials, fullname)
		wekanUsers = append(wekanUsers, wekanUser)
	}
	return wekanUsers
}

func WekanUsernamesSelect(users libwekan.Users, usernames []Username) libwekan.Users {
	wekanUsersMap := Map(users)
	var filteredUsers libwekan.Users
	for _, username := range usernames {
		filteredUsers = append(filteredUsers, wekanUsersMap[username])
	}
	return filteredUsers
}

func Map(users libwekan.Users) map[Username]libwekan.User {
	wekanUsersMap := make(map[Username]libwekan.User)
	for _, user := range users {
		wekanUsersMap[Username(user.Username)] = user
	}
	return wekanUsersMap
}

func UsernamesSelect(users Users, usernames []Username) Users {
	isInUsername := func(username Username, void User) bool {
		return contains(usernames, username)
	}
	return selectMap(users, isInUsername)
}

func (users Users) selectScopeWekan() Users {
	wekanUsers := make(Users)
	for username, user := range users {
		if contains(user.scope, "wekan") {
			wekanUsers[username] = user
		}
	}
	return wekanUsers
}

func (users Users) Usernames() []Username {
	var usernames []Username
	for username := range users {
		usernames = append(usernames, username)
	}
	return usernames
}
