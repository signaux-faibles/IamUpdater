package main

import (
	"fmt"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"sort"
	"strings"

	"github.com/Nerzal/gocloak/v13"
)

type Username string

// User is the definition of an user in excel state
type User struct {
	niveau            string
	email             Username
	prenom            string
	nom               string
	segment           string
	fonction          string
	employeur         string
	goup              string
	scope             []string
	accesGeographique string
	boards            []string
	taskforces        []string
}

// Users is the collection of wanted users
type Users map[Username]User

var habilitations = CompositeRoles{
	"a": []string{"bdf", "detection", "dgefp", "pge", "score", "urssaf"},
	"b": []string{"detection", "dgefp", "pge", "score"},
}

func (user User) getRoles() Roles {
	fields := logger.DataForMethod("getRoles")
	fields.AddAny("user", user)
	var roles Roles
	// TODO should return MisconfiguredUserError
	if user.niveau == "" {
		logger.Warn("aucun niveau", fields)
	}
	roles = habilitations[user.niveau]
	if strings.EqualFold("a", user.niveau) {
		roles.add("urssaf", "dgefp", "bdf")
	}
	if strings.EqualFold("b", user.niveau) {
		roles.add("dgefp")
	}
	if strings.EqualFold("a", user.niveau) || strings.EqualFold("b", user.niveau) {
		roles.add("score", "detection", "pge")
		if user.accesGeographique != "" {
			roles.add(user.accesGeographique)
		}
	}
	if !(len(user.scope) == 1 && user.scope[0] == "") {
		roles.add(user.scope...)
	}
	return roles
}

// GetUser resolves existing user from its username
func (kc KeycloakContext) GetUser(username Username) (gocloak.User, error) {
	for _, u := range kc.Users {
		if u != nil && u.Username != nil && strings.EqualFold(*u.Username, string(username)) {
			return *u, nil
		}
	}
	return gocloak.User{},
		fmt.Errorf(
			"l'utilisateur '%s' n'existe pas dans le contexte Keycloak",
			username,
		)
}

// Compare returns missing, obsoletes, disabled users from kc.Users from []user
func (users Users) Compare(kc KeycloakContext) ([]gocloak.User, []gocloak.User, []gocloak.User, []gocloak.User) {
	var missing []User
	var enable []gocloak.User
	var obsolete []gocloak.User
	var current []gocloak.User

	for _, u := range users {
		kcu, err := kc.GetUser(u.email)
		if err != nil {
			missing = append(missing, u)
		}
		if err == nil && !*kcu.Enabled {
			enable = append(enable, kcu)
		}
	}

	for _, kcu := range kc.Users {
		if _, ok := users[Username(strings.ToLower(*kcu.Username))]; !ok {
			if *kcu.Enabled {
				obsolete = append(obsolete, *kcu)
			}
		} else {
			current = append(current, *kcu)
		}
	}
	return toGocloakUsers(missing), obsolete, enable, current
}

func toGocloakUsers(users []User) []gocloak.User {
	var u []gocloak.User
	for _, user := range users {
		u = append(u, user.ToGocloakUser())
	}
	return u
}

// ToGocloakUser creates a new gocloak.User object from User specification
func (user User) ToGocloakUser() gocloak.User {
	t := true
	attributes := make(map[string][]string)
	if user.goup != "" {
		attributes["goup_path"] = []string{user.goup}
	}
	attributes["fonction"] = []string{user.fonction}
	attributes["employeur"] = []string{user.employeur}

	if user.segment != "" {
		attributes["segment"] = []string{user.segment}
	}
	email := string(user.email)
	return gocloak.User{
		Username:      &email,
		Email:         &email,
		EmailVerified: &t,
		FirstName:     &user.prenom,
		LastName:      &user.nom,
		Enabled:       &t,
		Attributes:    &attributes,
	}
}

func compareAttributes(a *map[string][]string, b *map[string][]string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	for k := range *b {
		if _, ok := (*a)[k]; !ok {
			return false
		}
	}
	for k, attribA := range *a {
		attribB, ok := (*b)[k]
		if !ok {
			return false
		}
		sort.Strings(attribB)
		sort.Strings(attribA)
		if strings.Join(attribA, "\t") != strings.Join(attribB, "\t") {
			return false
		}
	}
	return true
}
