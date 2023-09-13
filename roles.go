package main

import (
	"context"
	"fmt"

	"github.com/Nerzal/gocloak/v13"

	"keycloakUpdater/v2/pkg/logger"
)

// Roles is role collection in []string with some handy functions attached
type Roles []string
type CompositeRoles map[string]Roles

func (roles *Roles) add(toAdd ...string) {
	if roles != nil && len(toAdd) > 0 {
		for _, current := range toAdd {
			if !contains(*roles, current) {
				*roles = append(*roles, current)
			}
		}
	}
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

func neededRoles(compositeRoles CompositeRoles, users Users) Roles {
	var neededRoles Roles
	for composite, roles := range compositeRoles {
		neededRoles.add(composite)
		for _, role := range roles {
			neededRoles.add(role)
		}
	}
	for _, user := range users {
		for _, role := range user.getRoles() {
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

// FindKeycloakRoles retrieves existing gocloak roles from Roles array
func (kc KeycloakContext) FindKeycloakRoles(clientName string, roles Roles) []gocloak.Role {
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
func (kc KeycloakContext) ComposeRoles(clientID string, compositeRoles CompositeRoles) error {
	fields := logger.ContextForMethod(kc.ComposeRoles)
	fields.AddString("clientId", clientID)
	// Add known roles
	for role, roles := range compositeRoles {
		fields.AddAny("role", role)
		gocloakRole := kc.GetRoleFromRoleName(clientID, role)
		if gocloakRole == nil {
			logger.Warn("role doesn't exists", fields)
			continue
		}

		gocloakRoles := kc.FindKeycloakRoles(clientID, roles)
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
			logger.Error("error from keycloak", fields, err)
		}
	}

	// Clean composite roles
	internalID, err := kc.GetInternalIDFromClientID(clientID)
	if err != nil {
		logger.Error("can't resolve client", fields, err)
	}

	for _, r := range kc.ClientRoles[clientID] {
		fields.AddRole(*r)
		composingRoles, err := kc.API.GetCompositeClientRolesByRoleID(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *r.ID)
		if err != nil {
			logger.Error("error when searching composite client role", fields, err)
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
				logger.Error("Error deleting client role composite", fields, err)
			}
		}
	}
	return nil
}

func (compositeRoles CompositeRoles) addRole(key, role string) {
	roles := compositeRoles[key]
	roles.add(role)
	compositeRoles[key] = roles
}
