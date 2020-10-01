package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	// initialisation et connexion à keycloak
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	kc, err := NewKeycloakContext(
		viper.GetString("realm"),
		viper.GetString("host"),
		viper.GetString("username"),
		viper.GetString("password"),
	)
	if err != nil {
		panic(err)
	}

	// loading desired state for users, composites roles
	users, compositeRoles, err := loadExcel()
	if err != nil {
		panic(err)
	}

	// gather roles, newRoles are created before users, oldRoles are deleted after users
	log.Println("checking roles and creating new ones")
	newRoles, oldRoles := neededRoles(users).compare(kc.GetClientRoles()[viper.GetString("client")])

	i, err := kc.CreateClientRoles(viper.GetString("client"), newRoles)
	if err != nil {
		log.Printf("failed creating new roles: %s", err.Error())
		panic(err)
	}
	log.Printf("%d roles created", i)

	// check and adjust composite roles
	err = kc.ComposeRoles(
		viper.GetString("client"),
		compositeRoles,
	)

	if err != nil {
		fmt.Println(err)
	}
	// checking users
	missing, obsolete, update, current := users.Compare(kc)

	err = kc.CreateUsers(missing.GetNewGocloakUsers(), users, viper.GetString("client"))
	if err != nil {
		panic(err)
	}

	// disable obsolete users
	err = kc.DisableUsers(obsolete, viper.GetString("client"))
	if err != nil {
		panic(err)
	}
	// enable existing but disabled users
	err = kc.EnableUsers(update, users, viper.GetString("client"))
	if err != nil {
		panic(err)
	}
	// make sure every on has correct roles
	err = kc.UpdateCurrentUsers(current, users, viper.GetString("client"))
	if err != nil {
		panic(err)
	}

	// delete old roles
	if len(oldRoles) > 0 {
		log.Printf("removing unused roles: %s", strings.Join(oldRoles, ", "))
		internalID, err := kc.GetInternalIDFromClientID(viper.GetString("client"))
		if err != nil {
			panic(err)
		}
		for _, role := range oldRoles.GetKeycloakRoles(viper.GetString("client"), kc) {
			err = kc.API.DeleteClientRole(context.Background(), kc.JWT.AccessToken, kc.realm, internalID, *role.Name)
			if err != nil {
				panic(err)
			}
		}
	}
}
