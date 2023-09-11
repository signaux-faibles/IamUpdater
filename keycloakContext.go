package main

import (
	"context"
	"github.com/Nerzal/gocloak/v13"
	"github.com/pkg/errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
)

// KeycloakContext carry keycloak state
type KeycloakContext struct {
	API         *gocloak.GoCloak
	JWT         *gocloak.JWT
	Realm       *gocloak.RealmRepresentation
	Clients     []*gocloak.Client
	Users       []*gocloak.User
	Roles       []*gocloak.Role
	ClientRoles map[string][]*gocloak.Role
}

func NewKeycloakContext(access *structs.Keycloak) (KeycloakContext, error) {
	init, err := Init(access.Address, access.Realm, access.Username, access.Password)
	return init, err
}

// Init provides a connected keycloak context object
func Init(hostname, realm, username, password string) (KeycloakContext, error) {
	fields := logger.DataForMethod("Init")
	fields.AddAny("path", hostname)
	fields.AddAny("realm", realm)
	fields.AddAny("user", username)

	logger.Info("initialize KeycloakContext [START]", fields)
	kc := KeycloakContext{}
	kc.API = gocloak.NewClient(hostname)
	var err error
	ctx := context.Background()
	logger.Debug("récupère le token d'admin", fields)
	kc.JWT, err = kc.API.LoginAdmin(ctx, username, password, realm)
	if err != nil {
		return KeycloakContext{}, err
	}

	// fetch Realm
	logger.Debug("récupère le realm", fields)
	kc.Realm, err = kc.API.GetRealm(ctx, kc.JWT.AccessToken, realm)
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Debug("synchronise les clients", fields)
	err = kc.refreshClients()
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Debug("synchronise les utilisateurs", fields)
	err = kc.refreshUsers()
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Debug("synchronise les rôles du Realm", fields)
	kc.Roles, err = kc.API.GetRealmRoles(ctx, kc.JWT.AccessToken, realm, gocloak.GetRoleParams{})
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Debug("synchronise les rôles clients", fields)
	err = kc.refreshClientRoles()
	if err != nil {
		return KeycloakContext{}, err
	}
	logger.Info("initialize KeycloakContext [DONE]", fields)
	return kc, nil
}

// GetRoles returns realm roles in []string
func (kc KeycloakContext) GetRoles() Roles {
	var roles Roles
	for _, r := range kc.Roles {
		if r != nil && r.Name != nil {
			roles.add(*r.Name)
		}
	}
	return roles
}

// CreateClientRoles creates a bunch of roles in a client from a []string
func (kc *KeycloakContext) CreateClientRoles(clientID string, roles Roles) (int, error) {
	fields := logger.DataForMethod("kc.CreateClientRoles")

	defer func() {
		if err := kc.refreshClientRoles(); err != nil {
			logger.ErrorE("error refreshing client roles", fields, err)
			panic(err)
		}
	}()

	internalClientID, err := kc.GetInternalIDFromClientID(clientID)
	if err != nil {
		return 0, errors.Errorf("kc.CreateClientRoles, %s: no such client", clientID)
	}

	i := 0
	for _, role := range roles {
		if kc.GetClientRoles()[clientID].contains(role) {
			return i, errors.Errorf("kc.CreateClientRoles, %s: role already exists", role)
		}
		kcRole := gocloak.Role{
			Name: &role,
		}
		_, err := kc.API.CreateClientRole(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalClientID, kcRole)
		if err != nil {
			return i, errors.Errorf("kc.CreateClientRoles, %s: could not create roles, %s", role, err.Error())
		}
		i++
	}
	return i, nil
}

// GetInternalIDFromClientID resolves internal client ID from human readable clientID
func (kc KeycloakContext) GetInternalIDFromClientID(clientID string) (string, error) {
	for _, c := range kc.Clients {
		if c != nil && c.ClientID != nil {
			if *c.ClientID == clientID {
				return *c.ID, nil
			}
		}
	}
	return "", errors.Errorf("kc.GetInternalIDFromClientID %s: no such clientID", clientID)
}

// GetQuietlyInternalIDFromClientID resolves internal client ID from human readable clientID
func (kc KeycloakContext) GetQuietlyInternalIDFromClientID(clientID string) (string, bool) {
	id, err := kc.GetInternalIDFromClientID(clientID)
	if err != nil {
		return "", false
	}
	return id, true
}

// GetClientRoles returns realm roles in map[string][]string
func (kc KeycloakContext) GetClientRoles() CompositeRoles {
	clientRoles := make(CompositeRoles)
	for n, c := range kc.ClientRoles {
		for _, r := range c {
			clientRoles.addRole(n, *r.Name)
		}
	}
	return clientRoles
}

func (kc *KeycloakContext) refreshClients() error {
	var err error
	clients, err := kc.API.GetClients(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), gocloak.GetClientsParams{})
	kc.Clients = clients
	return err
}

// refreshUsers pulls user base from keycloak server
func (kc *KeycloakContext) refreshUsers() error {
	var err error
	max := 100000
	kc.Users, err = kc.API.GetUsers(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), gocloak.GetUsersParams{
		Max: &max,
	})
	return err
}

func (kc *KeycloakContext) refreshClientRoles() error {
	kc.ClientRoles = make(map[string][]*gocloak.Role)
	for _, c := range kc.Clients {
		if c != nil && c.ClientID != nil {
			roles, err := kc.API.GetClientRoles(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), *c.ID, gocloak.GetRoleParams{})
			if err != nil {
				return err
			}
			kc.ClientRoles[*c.ClientID] = roles
		}
	}
	return nil
}

// CreateUsers sends a slice of gocloak Users to keycloak
func (kc *KeycloakContext) CreateUsers(users []gocloak.User, userMap Users, clientName string) error {
	internalID, err := kc.GetInternalIDFromClientID(clientName)
	if err != nil {
		return err
	}
	for _, user := range users {
		fields := logger.DataForMethod("kc.CreateUsers")
		fields.AddUser(user)
		logger.Info("creating user", fields)
		u, err := kc.API.CreateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), user)
		if err != nil {
			logger.WarnE("unable to create user", fields, err)
		}

		configRoles := userMap[Username(*user.Username)].getRoles()
		roles := kc.FindKeycloakRoles(clientName, configRoles)
		fields.AddRoles(roles)
		if roles != nil {
			logger.Info("adding roles to user", fields)
			if err = kc.AddClientRolesToUser(internalID, u, roles); err != nil {
				logger.ErrorE("error adding client roles", fields, err)
				return err
			}
		} else {
			logger.Warn("no role to add to user", fields)
		}
	}

	err = kc.refreshUsers()
	return err
}

func (kc *KeycloakContext) AddClientRolesToUser(internalClientId, userID string, roles []gocloak.Role) error {
	return kc.API.AddClientRolesToUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalClientId, userID, roles)
}

// DisableUsers disables users and deletes every roles of users
func (kc *KeycloakContext) DisableUsers(users []gocloak.User, clientName string) error {
	internalID, err := kc.GetInternalIDFromClientID(clientName)
	if err != nil {
		return err
	}
	for _, u := range users {
		if err = kc.disableUser(u, internalID); err != nil {
			return err
		}
	}
	err = kc.refreshUsers()
	return err
}

func (kc *KeycloakContext) disableUser(u gocloak.User, internalClientID string) error {
	fields := logger.DataForMethod("kc.disableUser")
	disabled := false
	u.Enabled = &disabled
	fields.AddUser(u)
	logger.Info("disabling user", fields)
	err := kc.API.UpdateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), u)
	if err != nil {
		logger.WarnE("error disabling user", fields, err)
		return err
	}
	roles, err := kc.API.GetClientRolesByUserID(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalClientID, *u.ID)
	if err != nil {
		logger.WarnE("failed to retrieve roles for user", fields, err)
	}
	var ro []gocloak.Role
	for _, r := range roles {
		ro = append(ro, *r)
	}
	fields.AddArray("roles", rolesFromGocloakRoles(roles))
	logger.Info("remove roles from user", fields)
	err = kc.API.DeleteClientRolesFromUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalClientID, *u.ID, ro)
	if err != nil {
		logger.WarnE("failed to remove roles", fields, err)
		return err
	}
	return nil
}

// EnableUsers enables users and adds roles
func (kc *KeycloakContext) EnableUsers(users []gocloak.User) error {
	fields := logger.DataForMethod("kc.EnableUsers")
	t := true
	for _, user := range users {
		fields.AddUser(user)
		logger.Info("enabling user", fields)
		user.Enabled = &t
		err := kc.API.UpdateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), user)
		if err != nil {
			logger.WarnE("failed to enable user", fields, err)
		}
	}
	err := kc.refreshUsers()
	return err
}

// UpdateCurrentUsers sets client roles on specified users according userMap
func (kc KeycloakContext) UpdateCurrentUsers(users []gocloak.User, userMap Users, clientName string) error {
	fields := logger.DataForMethod("kc.UpdateCurrentUsers")
	accountInternalID, err := kc.GetInternalIDFromClientID("account")
	if err != nil {
		return err
	}
	internalID, err := kc.GetInternalIDFromClientID(clientName)
	if err != nil {
		return err
	}

	for _, user := range users {
		fields.AddUser(user)
		roles, err := kc.API.GetClientRolesByUserID(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *user.ID)
		if err != nil {
			return err
		}
		accountPRoles, err := kc.API.GetClientRolesByUserID(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), accountInternalID, *user.ID)
		if err != nil {
			return err
		}
		accountRoles := rolesFromGocloakRoles(accountPRoles)

		u := userMap[Username(*user.Username)]
		ug := u.ToGocloakUser()
		if user.LastName != nil && u.nom != *user.LastName ||
			user.LastName != nil && u.prenom != *user.FirstName ||
			!compareAttributes(user.Attributes, ug.Attributes) {

			update := gocloak.User{
				ID:         user.ID,
				FirstName:  &u.prenom,
				LastName:   &u.nom,
				Attributes: ug.Attributes,
			}
			logger.Info("updating user name and attributes", fields)
			err := kc.API.UpdateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), update)
			if err != nil {
				logger.WarnE("failed to update user names", fields, err)
				return err
			}
		}

		novel, old := userMap[Username(*user.Username)].getRoles().compare(rolesFromGocloakRoles(roles))
		if len(old) > 0 {
			fields.AddArray("oldRoles", old)
			logger.Info("deleting unused roles", fields)
			err = kc.API.DeleteClientRolesFromUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *user.ID, kc.FindKeycloakRoles(clientName, old))
			if err != nil {
				logger.WarnE("failed to delete roles", fields, err)
			}
			fields.Remove("oldRoles")
		}

		if len(novel) > 0 {
			fields.AddArray("novelRoles", novel)
			logger.Info("adding missing roles", fields)
			keycloakRoles := kc.FindKeycloakRoles(clientName, novel)
			err = kc.AddClientRolesToUser(internalID, *user.ID, keycloakRoles)
			if err != nil {
				logger.WarnE("failed to add roles", fields, err)
			}
			fields.Remove("novelRoles")
		}

		if len(accountRoles) > 0 {
			fields.AddArray("accountRoles", accountRoles)
			logger.Info("disabling account management", fields)
			err = kc.API.DeleteClientRolesFromUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), accountInternalID, *user.ID, kc.FindKeycloakRoles("account", accountRoles))
			if err != nil {
				logger.WarnE("failed to disable management", fields, err)
			}
			fields.Remove("accountRoles")
		}
	}

	return nil
}

// SaveMasterRealm update master Realm
func (kc *KeycloakContext) SaveMasterRealm(input gocloak.RealmRepresentation) {
	fields := logger.DataForMethod("kc.SaveMasterRealm")
	id := "master"
	input.ID = &id
	input.Realm = &id
	logger.Info("update realm", fields)
	if err := kc.API.UpdateRealm(context.Background(), kc.JWT.AccessToken, input); err != nil {
		logger.ErrorE("Error when updating Realm ", fields, err)
		panic(err)
	}

	kc.refreshRealm(*input.Realm)
}

func (kc *KeycloakContext) refreshRealm(realmName string) {
	logger.Debugf("refresh Realm")
	realm, err2 := kc.API.GetRealm(context.Background(), kc.JWT.AccessToken, realmName)
	if err2 != nil {
		logger.Errorf("Error when fetching Realm : +%v", err2)
		panic(err2)
	}
	kc.Realm = realm
}

// SaveClients save clients then refresh clients list
func (kc *KeycloakContext) SaveClients(input []*gocloak.Client) error {
	for _, client := range input {
		if err := kc.saveClient(*client); err != nil {
			return errors.WithStack(err)
		}
	}
	err := kc.refreshClients()
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func (kc KeycloakContext) saveClient(input gocloak.Client) error {
	fields := logger.DataForMethod("kc.saveClient")
	fields.AddClient(input)
	//kc.refreshClients()
	id, found := kc.GetQuietlyInternalIDFromClientID(*input.ClientID)
	// need client creation
	if !found {
		logger.Info("create client", fields)
		createdId, err := kc.API.CreateClient(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), input)
		if err != nil {
			return errors.WithStack(err)
		}
		fields.AddAny("id", createdId)
		return nil
	}
	// update client
	input.ID = &id
	if err := kc.API.UpdateClient(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), input); err != nil {
		logger.Info("update client", fields)
		return errors.Wrap(err, "error updating client")
	}
	return nil
}

func (kc KeycloakContext) getRealmName() string {
	return *kc.Realm.Realm
}

func (kc KeycloakContext) getClient(clientID string) (*gocloak.Client, bool) {
	client, ok := kc.getClients(clientID)[clientID]
	return client, ok
}

func (kc KeycloakContext) getClients(clientIDs ...string) map[string]*gocloak.Client {
	clientsMap := make(map[string]*gocloak.Client, len(kc.Clients))
	for _, client := range kc.Clients {
		if contains(clientIDs, *client.ClientID) {
			clientsMap[*client.ClientID] = client
		}
	}
	return clientsMap
}
