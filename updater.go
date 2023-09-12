package main

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/pkg/errors"

	"keycloakUpdater/v2/logger"
)

func UpdateKeycloak(
	kc *KeycloakContext,
	clientId string,
	realm *gocloak.RealmRepresentation,
	clients []*gocloak.Client,
	users Users,
	compositeRoles CompositeRoles,
	configuredUsername Username,
	maxChangesToAccept int,
) error {
	logContext := logger.ContextForMethod(UpdateKeycloak)

	if _, exists := users[configuredUsername]; !exists {
		return errors.Errorf(
			"l'utilisateur passé dans la configuration n'est pas présent dans le fichier d'habilitations: %s",
			configuredUsername,
		)
	}

	if _, err := kc.GetUser(configuredUsername); err != nil {
		return errors.Wrap(
			err,
			fmt.Sprintf(
				"l'utilisateur passé dans la configuration n'existe pas dans Keycloak : %s",
				configuredUsername,
			),
		)
	}

	logger.Info("START", logContext)
	logger.Info("accepte "+strconv.Itoa(maxChangesToAccept)+" changements pour les users", logContext)

	// checking users
	logger.Info("checking users", logContext)
	missing, obsolete, update, current := users.Compare(*kc)
	changes := len(missing) + len(obsolete) + len(update)
	keeps := len(current)
	if sure := areYouSureTooApplyChanges(changes, keeps, maxChangesToAccept); !sure {
		return errors.New("Trop de modifications utilisateurs.")
	}

	// gather roles, newRoles are created before users, oldRoles are deleted after users
	logger.Info("checking roles", logContext)
	neededRoles := neededRoles(compositeRoles, users)
	newRoles, oldRoles := neededRoles.compare(kc.GetClientRoles()[clientId])

	logger.Info("starting keycloak configuration", logContext)
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
		logger.Panic("erreur pendant l'écriture des nouveaux rôles", logContext, err)
	}
	logger.Info("roles created", logContext.AddAny("size", i))

	// check and adjust composite roles
	if err = kc.ComposeRoles(clientId, compositeRoles); err != nil {
		logger.Panic("erreur pendant l'écriture des rôles composés", logContext, err)
	}

	if err = kc.CreateUsers(missing, users, clientId); err != nil {
		logger.Panic("erreur pendant la création des utilisateurs", logContext, err)
	}

	// disable obsolete users
	if err = kc.DisableUsers(obsolete, clientId); err != nil {
		logger.Panic("erreur pendant la désactivation des utilisateurs", logContext, err)
	}
	// enable existing but disabled users
	if err = kc.EnableUsers(update); err != nil {
		logger.Panic("erreur pendant l'activation des utilisateurs", logContext, err)
	}

	// make sure every on has correct roles
	if err = kc.UpdateCurrentUsers(current, users, clientId); err != nil {
		logger.Error("erreur pendant la mise à jour des utilisateurs", logContext, err)
	}

	// delete old roles
	if len(oldRoles) > 0 {
		sort.Strings(oldRoles)
		logContext.AddArray("toDelete", oldRoles)
		logger.Info("removing unused roles", logContext)
		logContext.Remove("toDelete")
		internalID, err := kc.GetInternalIDFromClientID(clientId)
		if err != nil {
			panic(err)
		}
		for _, role := range kc.FindKeycloakRoles(clientId, oldRoles) {
			err = kc.API.DeleteClientRole(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *role.Name)
			if err != nil {
				panic(err)
			}
		}
		err = kc.refreshClientRoles()
		if err != nil {
			panic(err)
		}
	}
	logger.Info("DONE", logContext)
	return nil
}

func areYouSureTooApplyChanges(changes, keeps, acceptedChanges int) bool {
	fields := logger.ContextForMethod(areYouSureTooApplyChanges)
	logger.Info("nombre d'utilisateurs à rajouter/supprimer/activer : "+strconv.Itoa(changes), fields)
	logger.Info("nombre d'utilisateurs à conserver : "+strconv.Itoa(keeps), fields)
	if keeps < 1 {
		fmt.Println("Aucun utilisateur à conserver -> Refus de prendre en compte les changements.")
		return false
	}
	if acceptedChanges <= 0 {
		fmt.Println("Tous les changements sont acceptés (acceptedChanges: " + strconv.Itoa(changes) + ")")
		return true
	}
	if changes > acceptedChanges {
		fmt.Println("Trop de changements à prendre en compte. (Max : " + strconv.Itoa(acceptedChanges) + ")")
		return false
	}
	// pas trop de modif
	return true
}
