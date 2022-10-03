package main

import (
	"context"
	"fmt"
	"github.com/signaux-faibles/libwekan"
	"strings"
)

func ManageUsers(wekan libwekan.Wekan, fromConfig Users) error {
	wekanUsersfromConfig := fromConfig.selectScopeWekan()
	addAdmin(wekanUsersfromConfig, wekan)

	fromWekan, err := wekan.GetUsers(context.TODO())
	if err != nil {
		return err
	}

	creations, enable, disable := wekanUsersfromConfig.ListWekanChanges(fromWekan)

	err = wekan.CreateUsers(context.Background(), creations)
	if err != nil {
		return err
	}

	err = wekan.EnableUsers(context.Background(), enable)
	if err != nil {
		return err
	}

	return wekan.DisableUsers(context.Background(), disable)
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
	return mapSelect(users, isInUsername)
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
