package main

import (
	"context"
	"github.com/Nerzal/gocloak/v11"
	"github.com/pkg/errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

func UpdateAll(
	kc *KeycloakContext,
	clientId string,
	realm *gocloak.RealmRepresentation,
	clients []*gocloak.Client,
	filename string,
	configuredUsername string,
) error {
	fields := logger.DataForMethod("UpdateAll")

	if _, err := kc.GetUser(configuredUsername); err != nil {
		return errors.Wrap(err, "configured user does not exist in keycloak : "+configuredUsername)
	}

	logger.Info("START", fields)

	// loading desired state for users, composites roles
	logger.Info("loading excel stock file", fields)
	users, compositeRoles, err := loadExcel(filename)
	if err != nil {
		return err
	}
	if _, exists := users[configuredUsername]; !exists {
		return errors.Errorf("configured user is not in stock file (%s) : %s", filename, configuredUsername)
	}

	// checking users
	missing, obsolete, update, current := users.Compare(*kc)

	// realmName conf
	if realm != nil {
		kc.SaveMasterRealm(*realm)
	}

	// gather roles, newRoles are created before users, oldRoles are deleted after users
	logger.Info("checking roles and creating new ones", fields)
	neededRoles := neededRoles(compositeRoles, users)
	newRoles, oldRoles := neededRoles.compare(kc.GetClientRoles()[clientId])

	// clients conf
	if err = kc.SaveClients(clients); err != nil {
		return errors.Wrap(err, "error when saving clients")
	}

	i, err := kc.CreateClientRoles(clientId, newRoles)
	if err != nil {
		logger.ErrorE("failed creating new roles", fields, err)
	}
	logger.Infof("%d roles created", i)

	// check and adjust composite roles
	if err = kc.ComposeRoles(clientId, compositeRoles); err != nil {
		logger.Panic(err)
	}

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
		fields.Remove("toDelete")
		internalID, err := kc.GetInternalIDFromClientID(clientId)
		if err != nil {
			panic(err)
		}
		for _, role := range oldRoles.GetKeycloakRoles(clientId, *kc) {
			err = kc.API.DeleteClientRole(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *role.Name)
			if err != nil {
				panic(err)
			}
		}
		_ = kc.refreshClientRoles()
	}
	logger.Info("DONE", fields)
	return nil
}

func doesObsoletesContainsConfiguredUser(excelUsers []gocloak.User, currentUser gocloak.User) []gocloak.User {
	if excelUsers == nil {
		return nil
	}
	for index, current := range excelUsers {
		if current.Username == currentUser.Username {
			logger.Panicf("currentUser %v is not in stock file", currentUser)
			return append(excelUsers[:index], excelUsers[index+1:]...)
		}
	}
	return excelUsers
}
