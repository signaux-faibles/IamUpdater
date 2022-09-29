package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/signaux-faibles/libwekan"
)

func WekanUpdate(url, database, admin, filename string) error {
	wekan, err := libwekan.Init(context.TODO(), url, database, libwekan.Username(admin))
	if err != nil {
		return err
	}
	if err := wekan.Ping(); err != nil {
		return err
	}

	wekanAdminUser, err := wekan.AdminUser(context.Background())
	if err != nil {
		return err
	}

	usersFromExcel, _, err := loadExcel(filename)
	if err != nil {
		return err
	}

	// TODO: améliorer ça, objectif de s'assurer que l'admin reste bien actif
	usersFromExcel[Username(wekanAdminUser.Username)] = User{
		email: Username(wekanAdminUser.Username),
		scope: []string{"wekan"},
	}

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

	wekanBoardsMembers := usersFromExcel.listBoards()
	for boardSlug, boardMembers := range wekanBoardsMembers {
		err := SetMembers(wekan, boardSlug, boardMembers)
		if err != nil {
			return nil
		}
	}
	return nil
}

func SetMembers(wekan libwekan.Wekan, boardSlug libwekan.BoardSlug, boardMembers Users) error {
	board, err := wekan.GetBoardFromSlug(context.Background(), boardSlug)
	if err != nil {
		return err
	}
	currentMembersIDs := mapSlice(board.Members, func(member libwekan.BoardMember) libwekan.UserID { return member.UserId })

	admin, err := wekan.AdminUser(context.Background())
	if err != nil {
		return err
	}
	// wekan.AdminUser() est membre de toutes les boards, ajoutons le ici pour ne pas risquer de l'oublier dans les utilisateurs
	wantedMembersUsernames := []libwekan.Username{admin.Username}
	for username := range boardMembers {
		wantedMembersUsernames = append(wantedMembersUsernames, libwekan.Username(username))
	}
	wantedMembers, err := wekan.GetUsersFromUsernames(context.Background(), wantedMembersUsernames)
	if err != nil {
		return err
	}
	wantedMembersIDs := mapSlice(wantedMembers, func(user libwekan.User) libwekan.UserID { return user.ID })

	alreadyBoardMember, wantedInactiveBoardMember, ongoingBoardMember := intersect(currentMembersIDs, wantedMembersIDs)
	for _, userID := range alreadyBoardMember {
		err := wekan.EnsureUserIsActiveBoardMember(context.Background(), board.ID, userID)
		if err != nil {
			return err
		}
	}
	for _, userID := range ongoingBoardMember {
		err := wekan.EnsureUserIsActiveBoardMember(context.Background(), board.ID, userID)
		if err != nil {
			return err
		}
	}
	for _, userID := range wantedInactiveBoardMember {
		err := wekan.EnsureUserIsInactiveBoardMember(context.Background(), board.ID, userID)
		if err != nil {
			return err
		}
	}

	// wekan.AdminUser() est administrateur de toutes les boards, appliquons la règle
	return wekan.EnsureUserIsBoardAdmin(context.Background(), board.ID, admin.ID)
}

func (users Users) listBoards() map[libwekan.BoardSlug]Users {
	wekanBoardsUserSlice := make(map[libwekan.BoardSlug][]User)
	for _, user := range users {
		for _, boardSlug := range user.boards {
			if boardSlug != "" {
				boardSlug := libwekan.BoardSlug(boardSlug)
				wekanBoardsUserSlice[boardSlug] = append(wekanBoardsUserSlice[boardSlug], user)
			}
		}
	}

	wekanBoardsUsers := make(map[libwekan.BoardSlug]Users)
	for boardSlug, userSlice := range wekanBoardsUserSlice {
		wekanBoardsUsers[boardSlug] = mapifySlice(userSlice, func(user User) Username { return user.email })
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
