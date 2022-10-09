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

	withOauth2Func := isOauth2([]libwekan.Username{libwekan.Username(wekan.AdminUsername())})
	withOauth2FromWekan := selectSlice(fromWekan, withOauth2Func)
	creations, enable, disable := wekanUsersfromConfig.ListWekanChanges(withOauth2FromWekan)

	fields := logger.DataForMethod("insertUsers")
	logger.Info("> traite les inscriptions des utilisateurs", fields)
	fields.AddAny("population", len(creations))
	logger.Info(">> inscrit les nouveaux utilisateurs", fields)
	err = insertUsers(context.Background(), wekan, creations)
	if err != nil {
		return err
	}

	fields = logger.DataForMethod("EnableUsers")
	fields.AddAny("population", len(enable))
	logger.Info(">> active des utilisateurs réinscrits", fields)
	err = ensureUsersAreEnabled(context.Background(), wekan, enable)
	if err != nil {
		return err
	}

	fields = logger.DataForMethod("DisableUsers")
	fields.AddAny("population", len(disable))
	logger.Info(">> radie les utilisateurs absents", fields)
	return ensureUsersAreDisables(context.Background(), wekan, disable)
}

func insertUsers(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.DataForMethod("InsertUser")
	if err := wekan.AssertPrivileged(ctx); err != nil {
		return err
	}

	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Info(">>> crée l'utilisateur", fields)
		err := wekan.InsertUser(ctx, user)
		if err != nil {
			logger.Error(err.Error(), fields)
			return err
		}
	}
	return nil
}

func ensureUsersAreEnabled(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.DataForMethod("EnableUser")
	if err := wekan.AssertPrivileged(ctx); err != nil {
		return err
	}
	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Debug(">>> examine le statut de l'utilisateur", fields)
		err := wekan.EnableUser(ctx, user)
		if err == (libwekan.NothingDoneError{}) {
			continue
		}
		if err != nil {
			logger.Error(err.Error(), fields)
			return err
		}
		logger.Info(">>> active l'utilisateur", fields)
	}
	return nil
}

func ensureUsersAreDisables(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.DataForMethod("DisableUser")
	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Debug(">>> examine le statut de l'utilisateur", fields)
		err := wekan.DisableUser(ctx, user)
		if err == (libwekan.NothingDoneError{}) {
			continue
		}
		if err != nil {
			logger.Error(err.Error(), fields)
			return err
		}
		logger.Info(">>> désactive l'utilisateur", fields)
	}
	return nil
}

func (users Users) buildWekanUsers() libwekan.Users {
	var wekanUsers libwekan.Users
	for _, user := range users {
		initials := firstChar(user.prenom) + firstChar(user.nom)
		fullname := fmt.Sprintf("%s %s", strings.ToUpper(user.nom), user.prenom)
		wekanUser := libwekan.BuildUser(string(user.email), initials, fullname)
		wekanUsers = append(wekanUsers, wekanUser)
	}
	return wekanUsers
}

func wekanUsernamesSelect(users libwekan.Users, usernames []Username) libwekan.Users {
	wekanUsersMap := toMap(users)
	var filteredUsers libwekan.Users
	for _, username := range usernames {
		filteredUsers = append(filteredUsers, wekanUsersMap[username])
	}
	return filteredUsers
}

func toMap(users libwekan.Users) map[Username]libwekan.User {
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

func isOauth2(exceptions []libwekan.Username) func(libwekan.User) bool {
	return func(user libwekan.User) bool {
		return user.AuthenticationMethod == "oauth2" || contains(exceptions, user.Username)
	}
}

// checkNativeUsers apporte des logs permettant de garder un œil sur les utilisateurs gérés manuellement
func checkNativeUsers(wekan libwekan.Wekan, _ Users) error {
	ctx := context.Background()
	fields := logger.DataForMethod("checkNativeUsers")
	logger.Info("inventaire des comptes standards", fields)
	wekanUsers, err := wekan.GetUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range wekanUsers {
		if !user.LoginDisabled && user.AuthenticationMethod != "oauth2" && user.Username != wekan.AdminUsername() {
			fields.AddAny("username", user.Username)
			boards, err := wekan.SelectBoardsFromMemberID(ctx, user.ID)
			if err != nil {
				return err
			}

			activeUserBoards := selectSlice(boards, func(board libwekan.Board) bool { return board.UserIsActiveMember(user) && board.Slug != "templates" })
			activeUserBoardSlugs := mapSlice(activeUserBoards, func(board libwekan.Board) libwekan.BoardSlug { return board.Slug })
			fields.AddAny("boards", activeUserBoardSlugs)
			logger.Warn("utilisateur standard actif", fields)
		}
	}
	return nil
}
