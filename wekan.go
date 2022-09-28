package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/signaux-faibles/libwekan"
)

type WekanBoardsUsers map[BoardSlug][]User
type UserSlice []User

func WekanUpdate(url, database, admin, filename string) error {
	wekan, err := libwekan.Connect(context.TODO(), url, database, libwekan.Username(admin))
	//wekan.AdminUser(context.TODO())
	if err != nil {
		return err
	}

	usersFromExcel, _, err := loadExcel(filename)
	wekanUsers, err := wekan.GetUsers(context.TODO())
	if err != nil {
		return err
	}
	creations, enable, disable, err := usersFromExcel.WekanSelect().ListWekanChanges(wekanUsers)
	if err != nil {
		return err
	}

	err = wekan.CreateUsers(context.Background(), creations)
	if err != nil {
		return err
	}

	err = wekan.EnableUsers(context.Background(), enable)
	if err != nil {
		return err
	}

	err = wekan.DisableUsers(context.Background(), disable)
	if err != nil {
		return err
	}

	targetBoardsMembers := usersFromExcel.listBoards()
	for boardSlug, boardMembers := range targetBoardsMembers {
		SetMembers(wekan, boardSlug, boardMembers)
	}
	return err
}

func SetMembers(wekan libwekan.Wekan, boardSlug BoardSlug, boardMembers UserSlice) {

}

func (users Users) listBoards() WekanBoardsUsers {
	wekanBoardsUsers := make(WekanBoardsUsers)
	for _, user := range users {
		for _, board := range user.boards {
			if board != "" {
				boardSlug := BoardSlug(board)
				wekanBoardsUsers[boardSlug] = append(wekanBoardsUsers[boardSlug], user)
			}
		}
	}
	return wekanBoardsUsers
}

func (users Users) ListWekanChanges(wekanUsers libwekan.Users) (
	creations libwekan.Users,
	enable libwekan.Users,
	disable libwekan.Users,
	err error,
) {

	if err != nil {
		return nil, nil, nil, err
	}

	wekanUsernames := mapSlice(wekanUsers, usernameFromWekanUser)

	both, onlyWekan, notInWekan := intersect(wekanUsernames, users.Usernames())
	creations = UsernamesSelect(users, notInWekan).BuildWekanUsers()
	enable = WekanUsernamesSelect(wekanUsers, both)
	disable = WekanUsernamesSelect(wekanUsers, onlyWekan)

	return creations, enable, disable, nil
}

func usernameFromWekanUser(user libwekan.User) Username {
	return Username(user.Username)
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

func firstChar(s string) string {
	if len(s) > 0 {
		return s[0:1]
	}
	return ""
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

func (users Users) WekanSelect() Users {
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
	for _, user := range users {
		usernames = append(usernames, user.email)
	}
	return usernames
}
