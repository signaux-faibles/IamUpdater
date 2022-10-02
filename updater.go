package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Nerzal/gocloak/v11"
	"github.com/pkg/errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

func UpdateKeycloak(
	kc *KeycloakContext,
	clientId string,
	realm *gocloak.RealmRepresentation,
	clients []*gocloak.Client,
	users Users,
	compositeRoles CompositeRoles,
	configuredUsername Username,
	acceptedChanges int,
) error {
	fields := logger.DataForMethod("UpdateAll")

	if _, exists := users[configuredUsername]; !exists {
		return errors.Errorf("configured user is not in stock file: %s", configuredUsername)
	}

	if _, err := kc.GetUser(configuredUsername); err != nil {
		return errors.Wrap(err, string("configured user does not exist in keycloak : "+configuredUsername))
	}

	logger.Info("START", fields)
	logger.Info("accepte "+strconv.Itoa(acceptedChanges)+"changements pour les users", fields)

	// checking users
	logger.Info("checking users", fields)
	missing, obsolete, update, current := users.Compare(*kc)
	changes := len(missing) + len(obsolete) + len(update)
	keeps := len(current)
	if sure := areYouSureTooApplyChanges(changes, keeps, acceptedChanges); !sure {
		return errors.New("Trop de modifications utilisateurs.")
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
	if err := kc.SaveClients(clients); err != nil {
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

func areYouSureTooApplyChanges(changes, keeps, acceptedChanges int) bool {
	fields := logger.DataForMethod("areYouSureTooApplyChanges")
	logger.Info("nombre d'utilisateurs à rajouter/supprimer/activer : "+strconv.Itoa(changes), fields)
	logger.Info("nombre d'utilisateurs à conserver : "+strconv.Itoa(keeps), fields)
	condition1 := changes > acceptedChanges
	if condition1 {
		fmt.Println("Nombre d'utilisateurs à rajouter/supprimer/activer : " + strconv.Itoa(changes))
	}
	condition2 := keeps < 1
	if condition2 {
		fmt.Println("Nombre d'utilisateurs à conserver : " + strconv.Itoa(keeps))
	}
	if condition1 || condition2 {
		fmt.Println("Voulez vous continuez ? (t/F) :")
		var reader = bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		logger.Info("application des modifications : "+input, fields)
		if sure, err := strconv.ParseBool(input); err == nil {
			return sure
		}
		// par defaut
		return false
	}
	// pas trop de modif
	return true
}
