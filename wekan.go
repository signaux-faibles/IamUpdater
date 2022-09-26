package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/signaux-faibles/libwekan"
)

type WekanBoardsUsers map[BoardSlug][]User
type WekanUsers libwekan.Users
type UserSlice []User

func WekanUpdate(url, database, admin, filename string) error {
	wekan, err := libwekan.Connect(context.TODO(), url, database, libwekan.Username(admin))
	if err != nil {
		return err
	}

	users, _, err := loadExcel(filename)
	enable, disable, both, err := users.NeededChanges(wekan)
	if err != nil {
		return err
	}

	err = enable.EnableUsers(wekan)
	if err != nil {
		return err
	}

	disable.DisableUsers(wekan)
	if err != nil {
		return err
	}

	both.EnableUsers(wekan)
	if err != nil {
		return err
	}

	targetBoardsMembers := users.listBoards()
	for boardSlug, boardMembers := range wantedBoards {

	}
	return err
}

func SetMembers(wekan libwekan.Wekan, boardSlug BoardSlug, members UserSlice) {

}

func (wekanUsers WekanUsers) CreateUsers(wekan libwekan.Wekan) error {
	for _, wekanUser := range wekanUsers {
		_, err := wekan.InsertUser(context.Background(), wekanUser)
		if _, ok := err.(libwekan.UserAlreadyExistsError); ok {
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (wekanUsers WekanUsers) DisableUsers(wekan libwekan.Wekan) error {
	for _, wekanUser := range wekanUsers {
		_, err := wekan.DisableUser(context.Background(), wekanUser)
		if err != nil {
			return err
		}
	}
	return nil
}

func (wekanUsers WekanUsers) EnableUsers(wekan libwekan.Wekan) error {
	for _, wekanUser := range wekanUsers {
		_, err := wekan.EnableUser(context.Background(), wekanUser)
		if err != nil {
			return err
		}
	}
	return nil
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

func (users Users) NeededChanges(wekan libwekan.Wekan) (creations WekanUsers, enable WekanUsers, disable WekanUsers, err error) {
	wekanUsers, err := wekan.GetUsers(context.Background())
	if err != nil {
		return nil, nil, nil, err
	}

	both, onlyWekan, onlyUsers := intersect(WekanUsers(wekanUsers).Usernames(), users.Usernames())
	enable = WekanUsers(wekanUsers).UsernamesFilter(both)
	disable = WekanUsers(wekanUsers).UsernamesFilter(onlyWekan)
	creations = users.UsernamesFilter(onlyUsers).BuildWekanUsers()

	return creations, enable, disable, nil
}

func (users UserSlice) BuildWekanUsers() WekanUsers {
	var wekanUsers WekanUsers
	for _, user := range users {
		initials := user.prenom[0:1] + user.nom[0:1]
		fullname := fmt.Sprintf("%s %s", strings.ToUpper(user.nom), user.prenom)
		wekanUser := libwekan.BuildUser(string(user.email), initials, fullname)
		wekanUsers = append(wekanUsers, wekanUser)
	}
	return wekanUsers
}

func (users WekanUsers) UsernamesFilter(usernames []Username) WekanUsers {
	wekanUsersMap := users.Map()
	var filteredUsers WekanUsers
	for _, username := range usernames {
		filteredUsers = append(filteredUsers, wekanUsersMap[username])
	}
	return filteredUsers
}

func (users WekanUsers) Map() map[Username]libwekan.User {
	wekanUsersMap := make(map[Username]libwekan.User)
	for _, user := range users {
		wekanUsersMap[Username(user.Username)] = user
	}
	return wekanUsersMap
}

func (users Users) UsernamesFilter(usernames []Username) UserSlice {
	var filteredUsers UserSlice
	for _, username := range usernames {
		filteredUsers = append(filteredUsers, users[username])
	}
	return filteredUsers
}

func (users Users) Usernames() []Username {
	var usernames []Username
	for _, user := range users {
		usernames = append(usernames, user.email)
	}
	return usernames
}

func (users WekanUsers) Usernames() []Username {
	var usernames []Username
	for _, user := range users {
		usernames = append(usernames, Username(user.Username))
	}
	return usernames
}

func intersect(usernamesA []Username, usernamesB []Username) (both []Username, onlyA []Username, onlyB []Username) {
	for _, usernameA := range usernamesA {
		foundBoth := false
		for _, usernameB := range usernamesB {
			if usernameA == usernameB {
				both = append(both, usernameA)
				foundBoth = true
			}
		}
		if !foundBoth {
			onlyA = append(onlyA, usernameA)
		}
	}

	for _, usernameB := range usernamesB {
		foundBoth := false
		for _, usernameA := range usernamesA {
			if usernameA == usernameB {
				foundBoth = true
			}
		}
		if !foundBoth {
			onlyB = append(onlyB, usernameB)
		}
	}
	return both, onlyA, onlyB
}
