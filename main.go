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
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	clientId := viper.GetString("access.client")

	kc, err := NewKeycloakContext(
		viper.GetString("access.address"),
		viper.GetString("access.realm"),
		viper.GetString("access.username"),
		viper.GetString("access.password"))
	if err != nil {
		panic(err)
	}

	// realm configuration
	ConfigureRealm(&kc)

	// clients configuration
	ConfigureClients(kc)
	if err = kc.refreshClients(); err != nil {
		log.Fatalf("error refreshing clients : %s", err)
	}

	// loading desired state for users, composites roles
	excelFileName := viper.GetString("users.file")
	users, compositeRoles, err := loadExcel(excelFileName)
	if err != nil {
		panic(err)
	}
	// gather roles, newRoles are created before users, oldRoles are deleted after users
	log.Println("checking roles and creating new ones")
	newRoles, oldRoles := neededRoles(compositeRoles, users).compare(kc.GetClientRoles()[clientId])

	i, err := kc.CreateClientRoles(clientId, newRoles)
	if err != nil {
		log.Printf("failed creating new roles: %s", err.Error())
		panic(err)
	}
	log.Printf("%d roles created", i)

	// check and adjust composite roles
	err = kc.ComposeRoles(
		clientId,
		compositeRoles,
	)
	if err != nil {
		fmt.Println(err)
	}

	// checking users
	missing, obsolete, update, current := users.Compare(kc)

	err = kc.CreateUsers(missing.GetNewGocloakUsers(), users, clientId)
	if err != nil {
		panic(err)
	}

	// disable obsolete users
	err = kc.DisableUsers(obsolete, clientId)
	if err != nil {
		panic(err)
	}
	// enable existing but disabled users
	err = kc.EnableUsers(update)
	if err != nil {
		panic(err)
	}

	// make sure every on has correct roles
	err = kc.UpdateCurrentUsers(current, users, clientId)
	if err != nil {
		panic(err)
	}

	// delete old roles
	if len(oldRoles) > 0 {
		log.Printf("removing unused roles: %s", strings.Join(oldRoles, ", "))
		internalID, err := kc.GetInternalIDFromClientID(clientId)
		if err != nil {
			panic(err)
		}
		for _, role := range oldRoles.GetKeycloakRoles(clientId, kc) {
			err = kc.API.DeleteClientRole(context.Background(), kc.JWT.AccessToken, kc.realm, internalID, *role.Name)
			if err != nil {
				panic(err)
			}
		}
	}
}
