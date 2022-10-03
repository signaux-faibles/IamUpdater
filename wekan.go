package main

import (
	"context"
	"github.com/signaux-faibles/libwekan"
)

func WekanUpdate(url, database, admin, filename string) error {
	wekan, err := initWekan(url, database, admin)
	if err != nil {
		return err
	}

	allUsersFromExcel, _, err := loadExcel(filename)
	if err != nil {
		return err
	}

	err = ManageUsers(wekan, allUsersFromExcel)
	if err != nil {
		return err
	}

	err = ManageBoardsMembers(wekan, allUsersFromExcel)
	if err != nil {
		return err
	}

	return ManageBoardsLabelsTaskforce(wekan, allUsersFromExcel)
}

func addAdmin(usersFromExcel Users, wekan libwekan.Wekan) {
	usersFromExcel[Username(wekan.AdminUsername())] = User{
		email: Username(wekan.AdminUsername()),
		scope: []string{"wekan"},
	}
}

func initWekan(url string, database string, admin string) (libwekan.Wekan, error) {
	wekan, err := libwekan.Init(context.Background(), url, database, libwekan.Username(admin))
	if err != nil {
		return libwekan.Wekan{}, err
	}

	if err := wekan.Ping(context.Background()); err != nil {
		return libwekan.Wekan{}, err
	}
	err = wekan.AssertHasAdmin(context.Background())
	if err != nil {
		return libwekan.Wekan{}, err
	}
	return wekan, nil
}

func (users Users) ListWekanChanges(wekanUsers libwekan.Users) (
	creations libwekan.Users,
	enable libwekan.Users,
	disable libwekan.Users,
) {
	wekanUsernames := mapSlice(wekanUsers, usernameFromWekanUser)
	configUsernames := users.Usernames()

	both, onlyWekan, notInWekan := intersect(wekanUsernames, configUsernames)
	creations = UsernamesSelect(users, notInWekan).BuildWekanUsers()
	enable = WekanUsernamesSelect(wekanUsers, both)
	disable = WekanUsernamesSelect(wekanUsers, onlyWekan)

	return creations, enable, disable
}

func usernameFromWekanUser(user libwekan.User) Username {
	return Username(user.Username)
}

func firstChar(s string) string {
	if len(s) > 0 {
		return s[0:1]
	}
	return ""
}
