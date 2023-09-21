package main

import (
	"context"

	"github.com/Nerzal/gocloak/v13"
	"github.com/pkg/errors"

	"keycloakUpdater/v2/pkg/logger"
	"keycloakUpdater/v2/pkg/structs"
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
	logContext := logger.ContextForMethod(Init).
		AddString("path", hostname).
		AddString("realm", realm).
		AddString("user", username)

	logger.Debug("initialize KeycloakContext", logContext.Clone().AddString("status", "START"))
	kc := KeycloakContext{}
	kc.API = gocloak.NewClient(hostname)
	var err error
	ctx := context.Background()
	logger.Trace("récupère le token d'admin", logContext)
	kc.JWT, err = kc.API.LoginAdmin(ctx, username, password, realm)
	if err != nil {
		return KeycloakContext{}, err
	}

	// fetch Realm
	logger.Trace("récupère le realm", logContext)
	kc.Realm, err = kc.API.GetRealm(ctx, kc.JWT.AccessToken, realm)
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Trace("synchronise les clients", logContext)
	err = kc.refreshClients()
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Trace("synchronise les utilisateurs", logContext)
	err = kc.refreshUsers()
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Trace("synchronise les rôles du Realm", logContext)
	kc.Roles, err = kc.API.GetRealmRoles(ctx, kc.JWT.AccessToken, realm, gocloak.GetRoleParams{})
	if err != nil {
		return KeycloakContext{}, err
	}

	logger.Trace("synchronise les rôles clients", logContext)
	err = kc.refreshClientRoles()
	if err != nil {
		return KeycloakContext{}, err
	}
	logger.Debug("initialize KeycloakContext", logContext.Clone().AddString("status", "END"))
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
	fields := logger.ContextForMethod(kc.CreateClientRoles)

	defer func() {
		if err := kc.refreshClientRoles(); err != nil {
			logger.Error("error refreshing client roles", fields, err)
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
	logContext := logger.ContextForMethod(kc.CreateUsers).AddString("clientId", clientName)
	for _, user := range users {
		userLogContext := logContext.Clone().AddUser(user)
		logger.Notice("crée l'utilisateur Keycloak", userLogContext)
		u, err := kc.API.CreateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), user)
		if err != nil {
			logger.Error("erreur keycloak pendant la création de l'utilisateur", userLogContext, err)
			return err
		}

		configRoles := userMap[Username(*user.Username)].getRoles()
		roles := kc.FindKeycloakRoles(clientName, configRoles)
		userLogContext.AddRoles(roles)
		if roles != nil {
			logger.Notice("ajoute les rôles à l'utilisateur", userLogContext)
			if err = kc.AddClientRolesToUser(internalID, u, roles); err != nil {
				logger.Error("erreur pendant l'ajout des rôles à l'utilisateur", userLogContext, err)
				return err
			}
		} else {
			logger.Warn("pas de rôle à ajouter au nouvel utilisateur", userLogContext)
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
	fields := logger.ContextForMethod(kc.disableUser)
	disabled := false
	u.Enabled = &disabled
	fields.AddUser(u)
	logger.Info("disabling user", fields)
	err := kc.API.UpdateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), u)
	if err != nil {
		logger.Error("error disabling user", fields, err)
		return err
	}
	roles, err := kc.API.GetClientRolesByUserID(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalClientID, *u.ID)
	if err != nil {
		logger.Error("failed to retrieve roles for user", fields, err)
	}
	var ro []gocloak.Role
	for _, r := range roles {
		ro = append(ro, *r)
	}
	fields.AddArray("roles", rolesFromGocloakRoles(roles))
	logger.Info("remove roles from user", fields)
	err = kc.API.DeleteClientRolesFromUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalClientID, *u.ID, ro)
	if err != nil {
		logger.Error("failed to remove roles", fields, err)
		return err
	}
	return nil
}

// EnableUsers enables users and adds roles
func (kc *KeycloakContext) EnableUsers(users []gocloak.User) error {
	fields := logger.ContextForMethod(kc.EnableUsers)
	t := true
	for _, user := range users {
		fields.AddUser(user)
		logger.Info("enabling user", fields)
		user.Enabled = &t
		err := kc.API.UpdateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), user)
		if err != nil {
			logger.Error("failed to enable user", fields, err)
		}
	}
	err := kc.refreshUsers()
	return err
}

// UpdateCurrentUsers sets client roles on specified users according userMap
func (kc KeycloakContext) UpdateCurrentUsers(users []gocloak.User, userMap Users, clientName string) error {
	logContext := logger.ContextForMethod(kc.UpdateCurrentUsers)
	accountInternalID, err := kc.GetInternalIDFromClientID("account")
	if err != nil {
		return err
	}
	internalID, err := kc.GetInternalIDFromClientID(clientName)
	if err != nil {
		return err
	}

	for _, user := range users {
		logContext.AddUser(user)
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
			logger.Info("met à jour l'utilisateur et ses attributs", logContext.AddAny("update", update))
			err := kc.API.UpdateUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), update)
			if err != nil {
				logger.Error("erreur pendant la mise à jour de l'utilisateur", logContext, err)
				return err
			}
		}

		novel, old := userMap[Username(*user.Username)].getRoles().compare(rolesFromGocloakRoles(roles))
		if len(old) > 0 {
			oldRolesLogContext := logContext.Clone().AddArray("oldRoles", old)
			logger.Info("retire les rôles inutilisés", oldRolesLogContext)
			err = kc.API.DeleteClientRolesFromUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), internalID, *user.ID, kc.FindKeycloakRoles(clientName, old))
			if err != nil {
				logger.Error("failed to delete roles", oldRolesLogContext, err)
			}
		}

		if len(novel) > 0 {
			novelRolesLogContext := logContext.Clone().AddArray("novelRoles", novel)
			logger.Info("adding missing roles", novelRolesLogContext)
			keycloakRoles := kc.FindKeycloakRoles(clientName, novel)
			err = kc.AddClientRolesToUser(internalID, *user.ID, keycloakRoles)
			if err != nil {
				logger.Error("failed to add roles", novelRolesLogContext, err)
			}
		}

		if len(accountRoles) > 0 {
			accountRolesLogContext := logContext.Clone().AddArray("accountRoles", accountRoles)
			logger.Info("disabling account management", accountRolesLogContext)
			err = kc.API.DeleteClientRolesFromUser(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), accountInternalID, *user.ID, kc.FindKeycloakRoles("account", accountRoles))
			if err != nil {
				logger.Error("failed to disable management", accountRolesLogContext, err)
			}
		}
	}

	return nil
}

// SaveMasterRealm update master Realm
func (kc *KeycloakContext) SaveMasterRealm(input gocloak.RealmRepresentation) {
	logContext := logger.ContextForMethod(kc.SaveMasterRealm)
	id := "master"
	input.ID = &id
	input.Realm = &id
	logger.Info("met à jour le Realm", logContext.AddString("realm", id))
	if err := kc.API.UpdateRealm(context.Background(), kc.JWT.AccessToken, input); err != nil {
		logger.Panic("Erreur pendant la mise à jour du Realm ", logContext, err)
	}
	kc.refreshRealm(*input.Realm)
}

func (kc *KeycloakContext) refreshRealm(realmName string) {
	logContext := logger.ContextForMethod(kc.refreshRealm)
	logger.Debug("refresh Realm", logContext.AddString("realm", realmName))
	realm, err := kc.API.GetRealm(context.Background(), kc.JWT.AccessToken, realmName)
	if err != nil {
		logger.Panic("Erreur pendant la récupération du Realm", logContext, err)
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
	logContext := logger.ContextForMethod(kc.saveClient).AddClient(input)
	id, found := kc.GetQuietlyInternalIDFromClientID(*input.ClientID)
	// need client creation
	if !found {
		logger.Info("crée le client Keycloak", logContext)
		createdId, err := kc.API.CreateClient(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), input)
		if err != nil {
			return errors.WithStack(err)
		}
		logContext.AddAny("id", createdId)
		return nil
	}
	// update client
	input.ID = &id
	if err := kc.API.UpdateClient(context.Background(), kc.JWT.AccessToken, kc.getRealmName(), input); err != nil {
		logger.Info("update client", logContext)
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
