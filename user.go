package main

import (
	"errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"gopkg.in/yaml.v3"
	"os"
	"sort"
	"strings"

	"github.com/Nerzal/gocloak/v11"
)

// User is the definition of an user in excel state
type User struct {
	Niveau            string
	Email             string
	Prenom            string
	Nom               string
	Segment           string
	Fonction          string
	Employeur         string
	Goup              string
	Scope             []string
	AccesGeographique string `yaml:"accesGeographique"`
}

// Users is the collection of wanted users
type Users map[string]User

func (user User) roles() Roles {
	var roles Roles
	if user.Niveau == "a" {
		roles = append(roles, "urssaf", "dgefp", "bdf")
	}
	if user.Niveau == "b" {
		roles = append(roles, "dgefp")
	}
	if user.Niveau != "0" {
		roles = append(roles, "score", "detection", "pge")
		if user.AccesGeographique != "" {
			roles = append(roles, user.AccesGeographique)
		}
		if !(len(user.Scope) == 1 && user.Scope[0] == "") {
			roles = append(roles, user.Scope...)
		}
	}
	return roles
}

func (users Users) addUser(input User) Users {
	user := input.format()
	users[user.Email] = user
	return users
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
func (users Users) Compare(kc KeycloakContext) ([]gocloak.User, []gocloak.User, []gocloak.User, []gocloak.User) {
	var missing []User
	var enable []gocloak.User
	var obsolete []gocloak.User
	var current []gocloak.User

	for _, u := range users {
		kcu, err := kc.GetUser(u.Email)
		if err != nil {
			missing = append(missing, u)
		}
		if err == nil && !*kcu.Enabled {
			enable = append(enable, kcu)
		}
	}

	for _, kcu := range kc.Users {
		if _, ok := users[strings.ToLower(*kcu.Username)]; !ok {
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
	if user.Goup != "" {
		attributes["goup_path"] = []string{user.Goup}
	}
	attributes["fonction"] = []string{user.Fonction}
	attributes["employeur"] = []string{user.Employeur}

	if user.Segment != "" {
		attributes["segment"] = []string{user.Segment}
	}
	return gocloak.User{
		Username:      &user.Email,
		Email:         &user.Email,
		EmailVerified: &t,
		FirstName:     &user.Prenom,
		LastName:      &user.Nom,
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

func (user User) format() User {
	r := User{
		Niveau:            strings.ToLower(user.Niveau),
		Email:             strings.Trim(strings.ToLower(user.Email), " "),
		Nom:               strings.ToUpper(user.Nom),
		Prenom:            strings.ToTitle(user.Prenom),
		Segment:           strings.ToUpper(user.Segment),
		Fonction:          user.Fonction,
		Employeur:         strings.ToUpper(user.Employeur),
		Goup:              user.Goup,
		Scope:             user.Scope,
		AccesGeographique: user.AccesGeographique,
	}
	return r
}

func loadUsers(usersFilename []string) (Users, error) {
	users := make(Users)
	fields := logger.DataForMethod("loadUsers")
	if len(usersFilename) <= 0 {
		logger.Error("pas de fiche utilisateur", fields)
		return nil, errors.New("pas de fiche utilisateur à charger")
	}
	for _, f := range usersFilename {
		fields.AddAny("file", f)
		logger.Debug("charge la fiche", fields)
		current := User{}
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		if err = yaml.Unmarshal(data, &current); err != nil {
			return nil, err
		}
		users.addUser(current)
	}
	return users, nil
}
