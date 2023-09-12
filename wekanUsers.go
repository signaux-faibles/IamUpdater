package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/signaux-faibles/libwekan"

	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

var GENUINEUSERSELECTOR = []func(wekan libwekan.Wekan, user libwekan.User) bool{
	isOauth2User,
	IsAdminUser,
}

// checkNativeUsers apporte des logs permettant de garder un œil sur les utilisateurs gérés manuellement
func checkNativeUsers(wekan libwekan.Wekan, _ Users) error {
	ctx := context.Background()
	fields := logger.ContextForMethod(checkNativeUsers)
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

// manageUsers
// - objectif de traiter les utilisateurs Wekan
// - création des utilisateurs inconnus dans Wekan
// - désactivation des utilisateurs superflus
func manageUsers(wekan libwekan.Wekan, fromConfig Users) error {
	// l'admin wekan n'est pas dans le fichier de configuration source, ajoutons le
	addAdmin(fromConfig, wekan)

	fromWekan, err := selectWekanUsers(wekan)
	if err != nil {
		return err
	}

	toCreate, toEnable, toDisable := fromConfig.ListWekanChanges(fromWekan)

	if err := insertUsers(context.Background(), wekan, toCreate); err != nil {
		return err
	}

	if err := ensureUsersAreEnabled(context.Background(), wekan, toEnable); err != nil {
		return err
	}

	return ensureUsersAreDisabled(context.Background(), wekan, toDisable)
}

func selectWekanUsers(wekan libwekan.Wekan) (libwekan.Users, error) {
	users, err := wekan.GetUsers(context.TODO())
	genuineUsers := selectSlice(users, selectGenuineUserFunc(wekan))
	return genuineUsers, err
}

func insertUsers(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.ContextForMethod(insertUsers)
	logger.Info("> traite les inscriptions des utilisateurs", fields)
	fields.AddAny("population", len(users))
	logger.Info(">> inscrit les nouveaux utilisateurs", fields)
	if err := wekan.AssertPrivileged(ctx); err != nil {
		return err
	}

	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Info(">>> crée l'utilisateur", fields)
		err := wekan.InsertUser(ctx, user)
		if err != nil {
			logger.Error("erreur Wekan pendant la création des utilisateurs", fields, err)
			return err
		}
	}
	return nil
}

func ensureUsersAreEnabled(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.ContextForMethod(ensureUsersAreEnabled)
	fields.AddAny("population", len(users))
	logger.Info(">> active des utilisateurs réinscrits", fields)
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
			logger.Error("erreur Wekan pendant la radiation des utilisateurs", fields, err)
			return err
		}
		logger.Info(">>> active l'utilisateur", fields)
	}
	return nil
}

func ensureUsersAreDisabled(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	fields := logger.ContextForMethod(ensureUsersAreDisabled)
	fields.AddAny("population", len(users))
	logger.Info(">> radie les utilisateurs absents", fields)
	for _, user := range users {
		fields.AddAny("username", user.Username)
		logger.Debug(">>> examine le statut de l'utilisateur", fields)
		err := wekan.DisableUser(ctx, user)
		if err == (libwekan.NothingDoneError{}) {
			continue
		}
		if err != nil {
			logger.Error("erreur Wekan pendant l'examen du statut des utilisateurs", fields, err)
			return err
		}
		logger.Info(">>> désactive l'utilisateur", fields)
	}
	return nil
}

func (users Users) ListWekanChanges(wekanUsers libwekan.Users) (
	toCreate libwekan.Users,
	toEnable libwekan.Users,
	toDisable libwekan.Users,
) {
	configUsers := users.buildWekanUsers()
	wekanUsernames := mapSlice(wekanUsers, libwekan.User.GetUsername)
	configUsernames := mapSlice(keys(users), Username.toWekanUsername)

	both, onlyWekan, onlyConfig := intersect(wekanUsernames, configUsernames)

	toCreate = selectSlice(configUsers, acceptUserWithUsernameIn(onlyConfig))
	toEnable = selectSlice(wekanUsers, acceptUserWithUsernameIn(both))
	toDisable = selectSlice(wekanUsers, acceptUserWithUsernameIn(onlyWekan))

	return toCreate, toEnable, toDisable
}

func acceptUserWithUsernameIn(usernames []libwekan.Username) func(libwekan.User) bool {
	return func(user libwekan.User) bool {
		return contains(usernames, user.Username)
	}
}

func (users Users) buildWekanUsers() libwekan.Users {
	return mapSlice(values(users), User.buildWekanUser)
}

func (user User) buildWekanUser() libwekan.User {
	initials := firstChar(user.prenom) + firstChar(user.nom)
	fullname := fmt.Sprintf("%s %s", strings.ToUpper(user.nom), user.prenom)
	return libwekan.BuildUser(string(user.email), initials, fullname)
}

func (username Username) toWekanUsername() libwekan.Username {
	return libwekan.Username(username)
}

func (users Users) selectScopeWekan() Users {
	hasScope := func(user User) bool { return contains(user.scope, "wekan") }
	return selectMapByValue(users, hasScope)
}

// addAdmin modifie l'objet Users en place car c'est une map !
func addAdmin(users Users, wekan libwekan.Wekan) {
	users[Username(wekan.AdminUsername())] = User{
		email: Username(wekan.AdminUsername()),
		scope: []string{"wekan"},
	}
}

func IsAdminUser(wekan libwekan.Wekan, user libwekan.User) bool {
	return user.Username == wekan.AdminUsername()
}

func isOauth2User(_ libwekan.Wekan, user libwekan.User) bool {
	return user.AuthenticationMethod == "oauth2"
}

func selectGenuineUserFunc(wekan libwekan.Wekan) func(user libwekan.User) bool {
	return func(user libwekan.User) bool {
		for _, fn := range GENUINEUSERSELECTOR {
			if fn(wekan, user) {
				return true
			}
		}
		return false
	}
}

func firstChar(s string) string {
	if len(s) > 0 {
		return s[0:1]
	}
	return ""
}
