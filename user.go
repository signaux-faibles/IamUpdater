package main

import (
	"errors"
	"sort"
	"strings"

	"github.com/Nerzal/gocloak/v11"
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

func (user User) roles() Roles {
	var roles Roles
	if user.niveau == "" {
		// TODO should return MisconfiguredUserError
	}
	if strings.EqualFold("a", user.niveau) {
		roles = append(roles, "urssaf", "dgefp", "bdf")
	}
	if strings.EqualFold("b", user.niveau) {
		roles = append(roles, "dgefp")
	}
	if user.niveau != "0" {
		roles = append(roles, "score", "detection", "pge")
		if user.accesGeographique != "" {
			roles = append(roles, user.accesGeographique)
		}
		if !(len(user.scope) == 1 && user.scope[0] == "") {
			roles = append(roles, user.scope...)
		}
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
	return gocloak.User{}, errors.New("user does not exists")
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
