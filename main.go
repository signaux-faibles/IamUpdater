package main

import (
	"context"
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"strings"
)

func main() {
	log := logger.InfoLogger()

	conf := config.InitConfig("config.toml", "config.d")

	log.ConfigureWith(*conf.Logger)
	log.Infof("START")

	clientId := conf.Stock.Target
	kc, err := NewKeycloakContext(conf.Access, *log)
	if err != nil {
		panic(err)
	}

	// realmName conf
	kc.SaveMasterRealm(*conf.Realm)
	log.Infof("master Realm has been configured and updated")

	// clients conf
	kc.SaveClients(conf.Clients)

	// loading desired state for users, composites roles
	users, compositeRoles, err := loadExcel(conf.Stock.Filename)
	if err != nil {
		log.Panicf("error loading stock file")
	}
	// gather roles, newRoles are created before users, oldRoles are deleted after users
	log.Info("checking roles and creating new ones")
	newRoles, oldRoles := neededRoles(compositeRoles, users).compare(kc.GetClientRoles()[clientId])

	i, err := kc.CreateClientRoles(clientId, newRoles)
	if err != nil {
		log.Panicf("failed creating new roles: %s", err.Error())
	}
	log.Infof("%d roles created", i)

	// check and adjust composite roles
	if err = kc.ComposeRoles(clientId, compositeRoles); err != nil {
		log.Error(err)
	}

	// checking users
	missing, obsolete, update, current := users.Compare(kc)

	if err = kc.CreateUsers(missing.GetNewGocloakUsers(), users, clientId); err != nil {
		log.Panic(err)
	}

	// disable obsolete users
	if err = kc.DisableUsers(obsolete, clientId); err != nil {
		log.Panic(err)
	}
	// enable existing but disabled users
	if err = kc.EnableUsers(update); err != nil {
		log.Panic(err)
	}

	// make sure every on has correct roles
	if err = kc.UpdateCurrentUsers(current, users, clientId); err != nil {
		panic(err)
	}

	// delete old roles
	if len(oldRoles) > 0 {
		log.Infof("removing unused roles: %s", strings.Join(oldRoles, ", "))
		internalID, err := kc.GetInternalIDFromClientID(clientId)
		if err != nil {
			panic(err)
		}
		for _, role := range oldRoles.GetKeycloakRoles(clientId, kc) {
			err = kc.API.DeleteClientRole(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *role.Name)
			if err != nil {
				panic(err)
			}
		}
	}
	log.Infof("DONE")
}
