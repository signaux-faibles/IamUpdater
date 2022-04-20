package main

import (
	"context"
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

func main() {
	fields := logger.DataForMethod("main")
	conf := config.InitConfig("config.toml", "config.d")

	logger.ConfigureWith(*conf.Logger)
	logger.Info("START", fields)

	clientId := conf.Stock.Target
	kc, err := NewKeycloakContext(conf.Access)
	if err != nil {
		logger.Panic(err)
	}

	// realmName conf
	kc.SaveMasterRealm(*conf.Realm)

	// clients conf
	kc.SaveClients(conf.Clients)

	// loading desired state for users, composites roles
	users, compositeRoles, err := loadExcel(conf.Stock.Filename)
	if err != nil {
		logger.Panic(err)
	}
	// gather roles, newRoles are created before users, oldRoles are deleted after users
	logger.Infof("checking roles and creating new ones")
	newRoles, oldRoles := neededRoles(compositeRoles, users).compare(kc.GetClientRoles()[clientId])

	i, err := kc.CreateClientRoles(clientId, newRoles)
	if err != nil {
		logger.ErrorE("failed creating new roles", fields, err)
	}
	logger.Infof("%d roles created", i)

	// check and adjust composite roles
	if err = kc.ComposeRoles(clientId, compositeRoles); err != nil {
		logger.Panic(err)
	}

	// checking users
	missing, obsolete, update, current := users.Compare(kc)

	if err = kc.CreateUsers(missing.GetNewGocloakUsers(), users, clientId); err != nil {
		logger.Panic(err)
	}

	// disable obsolete users
	if err = kc.DisableUsers(obsolete, clientId); err != nil {
		logger.Panic(err)
	}
	// enable existing but disabled users
	if err = kc.EnableUsers(update); err != nil {
		logger.Panic(err)
	}

	// make sure every on has correct roles
	if err = kc.UpdateCurrentUsers(current, users, clientId); err != nil {
		logger.Panic(err)
	}

	// delete old roles
	if len(oldRoles) > 0 {
		fields.AddArray("toDelete", oldRoles)
		logger.Info("removing unused roles", fields)
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
	logger.Info("DONE", fields)
}
