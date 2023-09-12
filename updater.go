package main

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strconv"

	"github.com/Nerzal/gocloak/v13"
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
	maxChangesToAccept int,
) error {
	fields := logger.DataForMethod("UpdateAll")

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

	logger.Info("START", fields)
	logger.Info("accepte "+strconv.Itoa(maxChangesToAccept)+" changements pour les users", fields)

	// checking users
	logger.Info("checking users", fields)
	missing, obsolete, update, current := users.Compare(*kc)
	changes := len(missing) + len(obsolete) + len(update)
	keeps := len(current)
	if sure := areYouSureTooApplyChanges(changes, keeps, maxChangesToAccept); !sure {
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
		slog.Error("erreur pendant l'écriture des nouveaux rôles", slog.Any("error", err))
		panic(err)
	}
	slog.Info("roles created", slog.Int("size", i))

	// check and adjust composite roles
	if err = kc.ComposeRoles(clientId, compositeRoles); err != nil {
		slog.Error("erreur pendant l'écriture des rôles composés", slog.Any("error", err))
		panic(err)
	}

	if err = kc.CreateUsers(missing, users, clientId); err != nil {
		slog.Error("erreur pendant la création des utilisateurs", slog.Any("error", err))
		panic(err)
	}

	// disable obsolete users
	if err = kc.DisableUsers(obsolete, clientId); err != nil {
		slog.Error("erreur pendant la désactivation des utilisateurs", slog.Any("error", err))
		panic(err)
	}
	// enable existing but disabled users
	if err = kc.EnableUsers(update); err != nil {
		slog.Error("erreur pendant l'activation des utilisateurs", slog.Any("error", err))
		panic(err)
	}

	// make sure every on has correct roles
	if err = kc.UpdateCurrentUsers(current, users, clientId); err != nil {
		slog.Error("erreur pendant la mise à jour des utilisateurs", slog.Any("error", err))
		panic(err)
	}

	// delete old roles
	if len(oldRoles) > 0 {
		sort.Strings(oldRoles)
		fields.AddArray("toDelete", oldRoles)
		logger.Info("removing unused roles", fields)
		fields.Remove("toDelete")
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
	logger.Info("DONE", fields)
	return nil
}

func areYouSureTooApplyChanges(changes, keeps, acceptedChanges int) bool {
	fields := logger.DataForMethod("areYouSureTooApplyChanges")
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
