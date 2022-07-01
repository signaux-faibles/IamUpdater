package main

import (
	"context"
	"github.com/Nerzal/gocloak/v11"
	"github.com/pkg/errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"strconv"
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
	logger.Info("checking users", fields)
	missing, obsolete, update, current := users.Compare(*kc)
	if sure := areYouSureForUsers(len(missing)+len(obsolete)+len(update), len(current)); !sure {
		return errors.Errorf("trop de modifications utilisateurs")
	}

	// gather roles, newRoles are created before users, oldRoles are deleted after users
	logger.Info("checking roles", fields)
	neededRoles := neededRoles(compositeRoles, users)
	newRoles, oldRoles := neededRoles.compare(kc.GetClientRoles()[clientId])

	logger.Info("starting keycloak configuration", fields)
	// realmName conf
	if realm != nil {
		kc.SaveMasterRealm(*realm)
	}

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

	if err = kc.CreateUsers(missing, users, clientId); err != nil {
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

func areYouSureForUsers(nbChanges, nbUnchanges int) bool {
	fields := logger.DataForMethod("areYouSureForUsers")
	logger.Info("nombre d'utilisateurs à rajouter/supprimer/activer : "+strconv.Itoa(nbChanges), fields)
	logger.Info("nombre d'utilisateurs à mettre à jour au niveau des rôles : "+strconv.Itoa(nbUnchanges), fields)
	return true
}
