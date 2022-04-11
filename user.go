package main

import (
	"errors"
	"log"
	"sort"
	"strings"

	"github.com/Nerzal/gocloak/v11"
)

// User is the definition of an user in excel state
type User struct {
	niveau            string
	email             string
	prenom            string
	nom               string
	segment           string
	fonction          string
	employeur         string
	goup              string
	scope             []string
	accesGeographique string
}

// Users is the collection of wanted users
type Users map[string]User

func (user User) roles() Roles {
	var roles Roles
	if user.niveau == "a" {
		roles = append(roles, "urssaf", "dgefp", "bdf")
	}
	if user.niveau == "b" {
		roles = append(roles, "dgefp")
	}
	if user.niveau != "0" {
		roles = append(roles, "score", "detection")
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
func (kc KeycloakContext) GetUser(username string) (gocloak.User, error) {
	for _, u := range kc.Users {
		if u != nil && u.Username != nil && strings.EqualFold(*u.Username, username) {
			return *u, nil
		}
	}
	return gocloak.User{}, errors.New("user does not exists")
}

// Compare returns missing, obsoletes, disabled users from kc.Users from []user
func (users Users) Compare(kc KeycloakContext) (UserSlice, []gocloak.User, []gocloak.User, []gocloak.User) {
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
		if *kcu.Username == kc.username {
			log.Printf("avoid user from context : %s", kc.username)
		}
		if _, ok := users[strings.ToLower(*kcu.Username)]; !ok {
			if *kcu.Enabled {
				obsolete = append(obsolete, *kcu)
			}
		} else {
			current = append(current, *kcu)
		}
	}
	return missing, obsolete, enable, current
}

// UserSlice is a transparent name
type UserSlice []User

// GetNewGocloakUsers returns an array of gocloak.User objects to create
func (users UserSlice) GetNewGocloakUsers() []gocloak.User {
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
	return gocloak.User{
		Username:      &user.email,
		Email:         &user.email,
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
