package main

import (
	"context"
	"fmt"
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"log"
	"strings"
)

func main() {
	conf := config.InitConfig("config.toml", "config.d")
	// initialisation et connexion à keycloak
	clientId := conf.GetDefaultClient()

	log.Printf("default clientId : %s", clientId)

	kc, err := NewKeycloakContext(
		conf.GetAddress(),
		conf.GetRealm(),
		conf.GetUsername(),
		conf.GetPassword(),
	)
	if err != nil {
		panic(err)
	}

	// realmName conf
	kc.SaveMasterRealm(*conf.Realm)
	log.Println("master Realm has been configured and updated")

	// clients conf
	kc.SaveClients(conf.Clients)

	// loading desired state for users, composites roles
	users, compositeRoles, err := loadExcel(conf.GetUsersFile())
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
