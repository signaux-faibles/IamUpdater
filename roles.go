package main

import (
	"context"
	"log"

	"github.com/Nerzal/gocloak/v7"
)

// Roles is role collection in []string with some handy functions attached
type Roles []string

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
	var create Roles
	for _, r := range roles {
		if !otherRoles.contains(r) {
			create.add(r)
		}
	}

	var delete Roles
	for _, r := range otherRoles {
		if !roles.contains(r) {
			delete.add(r)
		}
	}
	return create, delete
}

func rolesFromGocloakRoles(roles []gocloak.Role) Roles {
	var r Roles
	for _, i := range roles {
		r.add(*i.Name)
	}
	return r
}

func rolesFromGocloakPRoles(roles []*gocloak.Role) Roles {
	var r Roles
	for _, i := range roles {
		r.add(*i.Name)
	}
	return r
}

func neededRoles(users Users) Roles {
	var roles Roles
	for _, user := range users {
		for _, role := range user.roles() {
			roles.add(role)
		}
	}
	return roles
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
	// Add known roles
	for role, roles := range compositeRoles {
		gocloakRole := kc.GetRoleFromRoleName(clientID, role)
		if gocloakRole == nil {
			log.Printf("composeRoles - %s: role doesn't exists", role)
			continue
		}
		gocloakRoles := roles.GetKeycloakRoles(clientID, kc)
		if len(gocloakRoles) != len(roles) {
			log.Printf("composeRoles - %s: only %d on %d roles exist, some roles may not be used in user base", role, len(gocloakRoles), len(roles))
			if len(gocloakRoles) == 0 {
				log.Printf("composeRoles - %s: no composite roles to send, discarding", role)
				continue
			}
		}
		err := kc.API.AddClientRoleComposite(context.Background(), kc.JWT.AccessToken, kc.realm, *gocloakRole.ID, gocloakRoles)
		if err != nil {
			log.Printf("composeRoles - %s: error from keycloak, %s", role, err.Error())
		}
	}

	// Clean composite roles
	internalID, err := kc.GetInternalIDFromClientID(clientID)
	if err != nil {
		log.Printf("composeRoles - %s: can't resolve client", clientID)
	}

	for _, r := range kc.ClientRoles[clientID] {
		composingRoles, err := kc.API.GetCompositeClientRolesByRoleID(context.Background(), kc.JWT.AccessToken, kc.realm, internalID, *r.ID)
		if err != nil {
			log.Println(err.Error())
		}
		wantedRoles := compositeRoles[*r.Name]
		var deleteRoles []gocloak.Role
		for _, c := range composingRoles {
			if !wantedRoles.contains(*c.Name) {
				deleteRoles = append(deleteRoles, *c)
			}
		}
		if len(deleteRoles) != 0 {
			log.Printf("composeRoles - %s: removing %d composing role(s)", *r.Name, len(deleteRoles))
			kc.API.DeleteClientRoleComposite(context.Background(), kc.JWT.AccessToken, kc.realm, *r.ID, deleteRoles)
		}
	}
	return nil
}
