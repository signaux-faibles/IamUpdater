package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/signaux-faibles/libwekan"
	"strings"
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

func ManageBoardsMembers(wekan libwekan.Wekan, fromConfig Users) error {
	wekanBoardsMembers := fromConfig.selectScopeWekan().inferBoardsMember()
	for boardSlug, boardMembers := range wekanBoardsMembers {
		err := SetMembers(wekan, boardSlug, boardMembers)
		if err != nil {
			return err
		}
	}
	return nil
}

func ManageBoardsLabelsTaskforce(wekan libwekan.Wekan, fromConfig Users) error {
	return errors.New("not implemented")
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

func SetMembers(wekan libwekan.Wekan, boardSlug libwekan.BoardSlug, boardMembers Users) error {
	board, err := wekan.GetBoardFromSlug(context.Background(), boardSlug)
	if err != nil {
		return err
	}
	currentMembersIDs := mapSlice(board.Members, func(member libwekan.BoardMember) libwekan.UserID { return member.UserID })

	// globalWekan.AdminUser() est membre de toutes les boards, ajoutons le ici pour ne pas risquer de l'oublier dans les utilisateurs
	wantedMembersUsernames := []libwekan.Username{wekan.AdminUsername()}
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

	// globalWekan.AdminUser() est administrateur de toutes les boards, appliquons la règle
	return wekan.EnsureUserIsBoardAdmin(context.Background(), board.ID, libwekan.UserID(wekan.AdminID()))
}

func (users Users) inferBoardsMember() BoardsMembers {
	wekanBoardsUserSlice := make(map[libwekan.BoardSlug][]User)
	for _, user := range users {
		for _, boardSlug := range user.boards {
			if boardSlug != "" {
				boardSlug := libwekan.BoardSlug(boardSlug)
				wekanBoardsUserSlice[boardSlug] = append(wekanBoardsUserSlice[boardSlug], user)
			}
		}
	}

	wekanBoardsUsers := make(BoardsMembers)
	for boardSlug, userSlice := range wekanBoardsUserSlice {
		wekanBoardsUsers[boardSlug] = mapifySlice(userSlice, func(user User) Username { return user.email })
	}
	return wekanBoardsUsers
}

type BoardsMembers map[libwekan.BoardSlug]Users

func (boardsMembers BoardsMembers) AddBoards(boards []libwekan.Board) BoardsMembers {
	if boardsMembers == nil {
		boardsMembers = make(BoardsMembers)
	}
	for _, b := range boards {
		if _, ok := boardsMembers[b.Slug]; !ok {
			boardsMembers[b.Slug] = make(Users)
		}
	}
	return boardsMembers
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
