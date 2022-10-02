package main

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v11"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

// Roles is role collection in []string with some handy functions attached
type Roles []string
type CompositeRoles map[string]Roles

func (roles *Roles) add(role string) {
	if roles != nil {
		for _, r := range *roles {
			if role == r {
				return
			}
		}
	}
	*roles = append(*roles, role)
}

func (roles Roles) contains(role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func (roles Roles) compare(otherRoles Roles) (Roles, Roles) {
	var toCreate Roles
	for _, r := range roles {
		if !otherRoles.contains(r) {
			toCreate.add(r)
		}
	}

	var toDelete Roles
	for _, r := range otherRoles {
		if !roles.contains(r) {
			toDelete.add(r)
		}
	}
	return toCreate, toDelete
}

func rolesFromGocloakRoles(roles []*gocloak.Role) Roles {
	var r Roles
	for _, i := range roles {
		r.add(*i.Name)
	}
	return r
}

func neededRoles(compositeRoles map[string]Roles, users Users) Roles {
	var neededRoles Roles
	for composite, roles := range compositeRoles {
		neededRoles.add(composite)
		for _, role := range roles {
			neededRoles.add(role)
		}
	}
	for _, user := range users {
		for _, role := range user.roles() {
			neededRoles.add(role)
		}
	}
	return neededRoles
}

// GetRoleFromRoleName resolves gocloak role object from a name
func (kc KeycloakContext) GetRoleFromRoleName(clientID string, role string) *gocloak.Role {
	for _, r := range kc.ClientRoles[clientID] {
		if r.Name != nil && *r.Name == role {
			return r
		}
	}
	return nil
}

// GetKeycloakRoles retrieves existing gocloak roles from Roles array
func (roles Roles) GetKeycloakRoles(clientName string, kc KeycloakContext) []gocloak.Role {
	var gocloakRoles []gocloak.Role
	for _, r := range roles {
		role := kc.GetRoleFromRoleName(clientName, r)
		if role != nil {
			gocloakRoles = append(gocloakRoles, *role)
		}
	}
	return gocloakRoles
}

// ComposeRoles writes roles composition to keycloak server
func (kc KeycloakContext) ComposeRoles(clientID string, compositeRoles map[string]Roles) error {
	fields := logger.DataForMethod("kc.ComposeRoles")
	fields.AddAny("clientId", clientID)
	// Add known roles
	for role, roles := range compositeRoles {
		fields.AddAny("role", role)
		gocloakRole := kc.GetRoleFromRoleName(clientID, role)
		if gocloakRole == nil {
			logger.Warn("role doesn't exists", fields)
			continue
		}
		gocloakRoles := roles.GetKeycloakRoles(clientID, kc)
		fields.AddRoles(gocloakRoles)
		if len(gocloakRoles) != len(roles) {
			message := fmt.Sprintf("only %d on %d roles exist, some roles may not be used in user base", len(gocloakRoles), len(roles))
			logger.Warn(message, fields)
			if len(gocloakRoles) == 0 {
				logger.Warn("no composite roles to send, discarding", fields)
				continue
			}
		}
		logger.Info("add composite roles", fields)
		err := kc.API.AddClientRoleComposite(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), *gocloakRole.ID, gocloakRoles)
		if err != nil {
			logger.WarnE("error from keycloak", fields, err)
		}
	}

	// Clean composite roles
	internalID, err := kc.GetInternalIDFromClientID(clientID)
	if err != nil {
		logger.WarnE("can't resolve client", fields, err)
	}

	for _, r := range kc.ClientRoles[clientID] {
		fields.AddRole(*r)
		composingRoles, err := kc.API.GetCompositeClientRolesByRoleID(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *r.ID)
		if err != nil {
			logger.ErrorE("error when searching composite client role", fields, err)
		}
		wantedRoles := compositeRoles[*r.Name]
		var deleteRoles []gocloak.Role
		for _, c := range composingRoles {
			if !wantedRoles.contains(*c.Name) {
				deleteRoles = append(deleteRoles, *c)
			}
		}
		if len(deleteRoles) != 0 {
			fields.AddRoles(deleteRoles)
			logger.Info("removing composing role(s)", fields)
			if err = kc.API.DeleteClientRoleComposite(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), *r.ID, deleteRoles); err != nil {
				logger.ErrorE("Error deleting client role composite", fields, err)
			}
		}
	}
	return nil
}
