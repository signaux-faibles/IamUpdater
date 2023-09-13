package main

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strconv"

	"github.com/Nerzal/gocloak/v13"
	"github.com/pkg/errors"

	"keycloakUpdater/v2/pkg/logger"
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
	logContext := logger.ContextForMethod(UpdateKeycloak).AddString("client", clientId)

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
		return errors.New("trop de modifications utilisateurs.")
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
	if i > 0 {
		slices.Sort(newRoles)
		logger.Notice("rôles créés", logContext.Clone().AddAny("size", i).AddArray("roles", newRoles))
	} else {
		logger.Info("pas de rôle à créer", logContext)
	}

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
	logContext := logger.ContextForMethod(areYouSureTooApplyChanges)
	logger.Notice("utilisateurs à rajouter/supprimer/activer", logContext.Clone().AddInt("nombre", changes))
	logger.Info("utilisateurs à conserver", logContext.Clone().AddInt("nombre", keeps))
	if keeps < 1 {
		logger.Warn("aucun utilisateur à conserver -> Refus de prendre en compte les changements.", logContext)
		return false
	}
	if acceptedChanges <= 0 {
		logger.Info("tous les changements sont acceptés", logContext.Clone().AddInt("changements", changes))
		return true
	}
	if changes > acceptedChanges {
		logger.Warn(
			"trop de changements à prendre en compte.",
			logContext.Clone().AddInt("max", acceptedChanges).AddInt("current", changes),
		)
		return false
	}
	// pas trop de modif
	return true
}
