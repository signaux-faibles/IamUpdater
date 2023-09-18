package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/signaux-faibles/libwekan"

	"keycloakUpdater/v2/pkg/logger"
)

var GENUINEUSERSELECTOR = []func(wekan libwekan.Wekan, user libwekan.User) bool{
	isOauth2User,
	IsAdminUser,
}

// checkNativeUsers apporte des logs permettant de garder un œil sur les utilisateurs gérés manuellement
func checkNativeUsers(wekan libwekan.Wekan, _ Users) error {
	ctx := context.Background()
	logContext := logger.ContextForMethod(checkNativeUsers)
	logger.Info("inventaire des comptes standards", logContext)
	wekanUsers, err := wekan.GetUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range wekanUsers {
		if !user.LoginDisabled && user.AuthenticationMethod != "oauth2" && user.Username != wekan.AdminUsername() {
			logContext.AddAny("username", user.Username)
			boards, err := wekan.SelectBoardsFromMemberID(ctx, user.ID)
			if err != nil {
				return err
			}

			activeUserBoards := selectSlice(boards, func(board libwekan.Board) bool { return board.UserIsActiveMember(user) && board.Slug != "templates" })
			activeUserBoardSlugs := mapSlice(activeUserBoards, func(board libwekan.Board) libwekan.BoardSlug { return board.Slug })
			logContext.AddAny("boards", activeUserBoardSlugs)
			logger.Info("utilisateur standard actif", logContext)
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
	logContext := logger.ContextForMethod(insertUsers)
	logger.Info("> traite les inscriptions des utilisateurs", logContext)
	logContext.AddInt("population", len(users))
	logger.Info(">> inscrit les nouveaux utilisateurs", logContext)
	if err := wekan.AssertPrivileged(ctx); err != nil {
		return err
	}
	for _, user := range users {
		logger.Notice(">>> crée l'utilisateur", logContext.AddAny("username", user.Username))
		err := wekan.InsertUser(ctx, user)
		if err != nil {
			logger.Error("erreur Wekan pendant la création des utilisateurs", logContext, err)
			return err
		}
	}
	return nil
}

func ensureUsersAreEnabled(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	logContext := logger.ContextForMethod(ensureUsersAreEnabled)
	logger.Info(">> active des utilisateurs réinscrits", logContext.Clone().AddInt("population", len(users)))
	if err := wekan.AssertPrivileged(ctx); err != nil {
		return err
	}
	for _, user := range users {
		logger.Debug(">>> examine le statut de l'utilisateur", logContext.AddAny("username", user.Username))
		err := wekan.EnableUser(ctx, user)
		if err == (libwekan.NothingDoneError{}) {
			continue
		}
		if err != nil {
			logger.Error("erreur Wekan pendant la radiation d'un utilisateur", logContext, err)
			return err
		}
		logger.Notice(">>> active l'utilisateur", logContext)
	}
	return nil
}

func ensureUsersAreDisabled(ctx context.Context, wekan libwekan.Wekan, users libwekan.Users) error {
	logContext := logger.ContextForMethod(ensureUsersAreDisabled)
	logger.Info(">> radie les utilisateurs absents", logContext.Clone().AddInt("population", len(users)))
	for _, user := range users {
		logContext.AddAny("username", user.Username)
		logger.Debug(">>> examine le statut de l'utilisateur", logContext)
		err := wekan.DisableUser(ctx, user)
		if err == (libwekan.NothingDoneError{}) {
			continue
		}
		if err != nil {
			logger.Error("erreur Wekan pendant l'examen du statut des utilisateurs", logContext, err)
			return err
		}
		logger.Notice(">>> désactive l'utilisateur", logContext)
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
