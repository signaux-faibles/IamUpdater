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
	realm gocloak.RealmRepresentation,
	clients []*gocloak.Client,
	filename string,
	currentUsername string,
) error {
	fields := logger.DataForMethod("UpdateAll")
	var err error
	var currentUser gocloak.User
	if currentUser, err = kc.GetUser(currentUsername); err != nil {
		return errors.Wrap(err, "current user not found : "+currentUsername)
	}

	logger.Info("START", fields)

	// realmName conf
	kc.SaveMasterRealm(realm)

	// clients conf
	if err = kc.SaveClients(clients); err != nil {
		return errors.Wrap(err, "error when saving clients")
	}

	// loading desired state for users, composites roles
	users, compositeRoles, err := loadExcel(filename)
	if err != nil {
		logger.Panic(err)
	}
	// gather roles, newRoles are created before users, oldRoles are deleted after users
	logger.Info("checking roles and creating new ones", fields)
	neededRoles := neededRoles(compositeRoles, users)
	newRoles, oldRoles := neededRoles.compare(kc.GetClientRoles()[clientId])

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
	missing, obsolete, update, current := users.Compare(*kc)

	if err = kc.CreateUsers(missing.GetNewGocloakUsers(), users, clientId); err != nil {
		logger.Panic(err)
	}

	// disable obsolete users
	obsolete = removeUser(obsolete, currentUser)
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

func removeUser(users []gocloak.User, toRemove gocloak.User) []gocloak.User {
	if users == nil {
		return nil
	}
	for index, current := range users {
		if current.Username == toRemove.Username {
			return append(users[:index], users[index+1:]...)
		}
	}
	return users
}
