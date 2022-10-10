//go:build integration
// +build integration

// nolint:errcheck
package main

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v11"
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestKeycloakConfiguration_access_username_should_be_present_in_stock_file(t *testing.T) {
	ass := assert.New(t)

	testUser := "ti_admin"

	//users, _, _ := loadExcel(testFilename)
	users := Users{keycloakAdmin + "estAbsent": User{}}

	// erreur in configuration : access.username should be in usersAndRolesFilename
	err := UpdateKeycloak(
		&kc,
		"peuimporte",
		nil,
		nil,
		users,
		nil,
		Username(testUser),
		10,
	)

	expectedError := fmt.Sprintf("configured user is not in stock file: %s", testUser)

	ass.NotNil(err)
	ass.EqualError(err, expectedError)
}

func TestKeycloakInitialisation(t *testing.T) {
	ass := assert.New(t)
	var conf structs.Config
	var err error
	if conf, err = config.InitConfig("test/resources/initialisation/test_config.toml"); err != nil {
		panic(err)
	}
	users := TEST_USERS
	compositeRoles := referentiel.toRoles()
	//configure logger
	logger.ConfigureWith(*conf.Logger)

	// update all
	if err = UpdateKeycloak(
		&kc,
		conf.Stock.ClientForRoles,
		conf.Realm,
		conf.Clients,
		users,
		compositeRoles,
		Username(conf.Keycloak.Username),
		10,
	); err != nil {
		t.Fatalf("erreur pendant l'update : %v", err)
	}

	// assertions about realm
	ass.Equal("master", *kc.Realm.Realm)
	ass.Equal("Signaux Faibles", *kc.Realm.DisplayName)
	ass.Equal("signaux-faibles", *kc.Realm.EmailTheme)
	ass.Equal("signaux-faibles", *kc.Realm.LoginTheme)
	ass.NotNil(kc.Realm.SMTPServer)
	ass.Equal(5, *kc.Realm.MinimumQuickLoginWaitSeconds)
	ass.True(*kc.Realm.RememberMe)

	clientSF, found := kc.getClient(signauxfaibleClientID)
	ass.True(found)
	ass.NotNil(clientSF)
	ass.True(*clientSF.PublicClient)
	ass.Contains(*clientSF.RedirectURIs, "https://signaux-faibles.beta.gouv.fr/*", "https://localhost:8080/*")
	ass.Len(kc.ClientRoles[signauxfaibleClientID], 145)

	clientAnother, found := kc.getClient("another")
	ass.True(found)
	ass.NotNil(clientAnother)
	ass.False(*clientAnother.PublicClient)
	ass.Len(kc.ClientRoles["another"], 1) // il y a au minimum 1 rôle pour 1 client

	user, err := kc.GetUser("raphael.squelbut@shodo.io")
	ass.Nil(err)
	ass.NotNil(user)

	emptyUser := gocloak.User{}
	if user != emptyUser {
		err = logUser(*clientSF, user)
		ass.Nil(err)
	}
}

// TestClientSignauxFaiblesExists teste l'existence du Client "signauxfaibles" par l'API
func TestClientSignauxFaiblesExists(t *testing.T) {
	ass := assert.New(t)

	searchClientRequest := gocloak.GetClientsParams{
		ClientID: &signauxfaibleClientID,
	}
	clientSG, err := kc.API.GetClients(context.Background(), kc.JWT.AccessToken, *kc.Realm.Realm, searchClientRequest)
	if err != nil {
		t.Fatalf("error getting keycloak users by role pge : %v", err)
	}
	//clientSG := searchClientByName(t, "signauxfaibles")
	actual, found := kc.getClient(signauxfaibleClientID)
	ass.True(found)
	ass.Len(clientSG, 1)
	ass.Contains(clientSG, actual)
}

func TestRolesExistences(t *testing.T) {
	ass := assert.New(t)
	rolesToTest := []string{"dgefp", "bdf", "score", "detection", "urssaf", "pge"}

	for _, role := range rolesToTest {

		roleRequest := gocloak.GetRoleParams{
			Search: &role,
		}

		clientSG, _ := kc.getClient(signauxfaibleClientID)
		rolesFromAPI, err := kc.API.GetClientRoles(
			context.Background(),
			kc.JWT.AccessToken,
			*kc.Realm.Realm,
			*clientSG.ID,
			roleRequest,
		)
		if err != nil {
			t.Fatalf("error getting client roles : %v", err)
		}
		ass.Lenf(rolesFromAPI, 1, "erreur pour le rôle %v", role)
		expected := rolesFromAPI[0]

		// on compare les résultats de l'API avec le contenu de l'objet KeycloakContext
		clientRolesFromContext := kc.ClientRoles[signauxfaibleClientID]
		ass.Contains(clientRolesFromContext, expected)

		actual := kc.GetRoleFromRoleName(signauxfaibleClientID, role)
		ass.Equalf(expected, actual, "erreur pour le rôle %v", role)
	}
}

func TestRolesAssignedToAll(t *testing.T) {
	ass := assert.New(t)
	clientSG, _ := kc.getClient(signauxfaibleClientID)
	rolesToTest := []string{"score", "detection", "pge"}

	for _, role := range rolesToTest {
		roleUrssaf := kc.GetRoleFromRoleName(signauxfaibleClientID, role)

		usersFromAPI, err := kc.API.GetUsersByClientRoleName(
			context.Background(),
			kc.JWT.AccessToken,
			*kc.Realm.Realm,
			*clientSG.ID,
			*roleUrssaf.Name,
			gocloak.GetUsersByRoleParams{},
		)
		if err != nil {
			t.Fatalf("erreur lors de la récupération des users payant le rôle %v : %v", role, err)
		}

		// il y a actuellement 3 users dans le fichier de provisionning excel
		// les 3 doivent avoir le rôle urssaf
		ass.Len(usersFromAPI, 3)
	}
}

func TestKeycloak_should_not_update_when_too_many_changes(t *testing.T) {
	ass := assert.New(t)
	var err error
	var conf structs.Config

	if conf, err = config.InitConfig("test/resources/update/test_config.toml"); err != nil {
		panic(err)
	}
	// configure logger
	logger.ConfigureWith(*conf.Logger)

	//users, compositeRoles, err := loadExcel(conf.Stock.UsersAndRolesFilename)
	users := ANOTHER_USERS
	compositeRoles := referentiel.toRoles()

	stdin := readStdin("false")
	// update all
	os.Stdin = stdin
	actual := UpdateKeycloak(
		&kc,
		conf.Stock.ClientForRoles,
		conf.Realm,
		conf.Clients,
		users,
		compositeRoles,
		Username(conf.Keycloak.Username),
		4,
	)
	ass.EqualError(actual, "Trop de modifications utilisateurs.")
}

func TestKeycloakUpdate(t *testing.T) {
	ass := assert.New(t)
	var conf structs.Config

	// voir le fichier
	// le user raphael.squelbut@shodo.io a été créé au test précédent
	disabledUser, err := kc.GetUser("raphael.squelbut@shodo.io")
	ass.Nil(err)
	ass.NotNil(disabledUser)

	clientSF, found := kc.getClient(signauxfaibleClientID)
	ass.True(found)

	// le user doit encore pouvoir se loguer
	// avant l'exécution de l'update
	err = logUser(*clientSF, disabledUser)
	ass.Nil(err)

	if conf, err = config.InitConfig("test/resources/update/test_config.toml"); err != nil {
		panic(err)
	}

	//users, compositeRoles, err := loadExcel(conf.Stock.UsersAndRolesFilename)
	users := ANOTHER_USERS
	compositeRoles := referentiel.toRoles()
	newCompositeRoles := make(CompositeRoles)
	acceptedCompositesRoles := []string{"Bretagne"}
	//// WILL REMOVE SOME MAP ENTRIES
	for _, name := range acceptedCompositesRoles {
		newCompositeRoles[name] = compositeRoles[name]
	}

	// configure logger
	logger.ConfigureWith(*conf.Logger)

	// update all
	err = UpdateKeycloak(
		&kc,
		conf.Stock.ClientForRoles,
		conf.Realm,
		conf.Clients,
		users,
		newCompositeRoles,
		Username(conf.Keycloak.Username),
		10,
	)
	if err != nil {
		panic(err)
	}

	// des rôles ont été supprimés dans le fichier de rôles
	ass.Len(kc.ClientRoles[signauxfaibleClientID], 26)

	// on vérifie
	err = logUser(*clientSF, disabledUser)
	ass.NotNil(err)
	apiError, ok := err.(*gocloak.APIError)
	ass.True(ok)
	ass.Equal(400, apiError.Code)
}

func readStdin(message string) *os.File {
	r, w, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	origStdin := os.Stdin
	os.Stdin = r

	w.WriteString(message)
	return origStdin
}

func logUser(client gocloak.Client, user gocloak.User) error {
	// try connecting a user

	// 1. need client secret
	clientSecret, err := kc.API.RegenerateClientSecret(context.Background(), kc.JWT.AccessToken, *kc.Realm.Realm, *client.ID)
	if err != nil {
		return err
	}
	// 2. set password for user
	err = kc.API.SetPassword(context.Background(), kc.JWT.AccessToken, *user.ID, *kc.Realm.Realm, "abcd", false)
	if err != nil {
		return err
	}
	// 3. log user
	_, err = kc.API.Login(context.Background(), *client.ClientID, *clientSecret.Value, *kc.Realm.Realm, *user.Username, "abcd")
	if err != nil {
		return err
	}
	return nil
}

var TEST_USERS = Users{
	"john.doe@zone51.gov.fr":    User{"A", "john.doe@zone51.gov.fr", "John", "Doe", "LISTENS THE WIND", "Recouvrement et accompagnement des entreprises", "PENTAGON", "", nil, "Alsace", nil, nil},
	"raphael.squelbut@shodo.io": User{"A", "raphael.squelbut@shodo.io", "Raphaël", "SQUELBUT", "sf", "Développeur", "SIGNAUX FAIBLES", "", []string{"wekan"}, "France entière", nil, nil},
	"quelqun@pasdelurssaf.fr":   User{"B", "quelqun@pasdelurssaf.fr", "quelqun", "pasdelurssaf", "", "Un mec pas de l’URSSAF", "", "", nil, "77", nil, nil},
	keycloakAdmin:               ADMIN,
}

var ADMIN = User{"0", keycloakAdmin, "", "admin_name", "", "", "", "", nil, "", nil, nil}

var ANOTHER_USERS = Users{
	"raphael.squelbut@shodo.fr": User{"A", "raphael.squelbut@shodo.fr", "Raphaël", "SQUELBUT", "sf", "Développeur", "SIGNAUX FAIBLES", "", []string{"wekan"}, "France entière", nil, nil},
	"john.doux@beat.gouv.fr":    User{"A", "john.doux@beat.gouv.fr", "John", "Doux", "LISTENS THE WIND", "Recouvrement et accompagnement des entreprises", "PENTAGON", "", nil, "Lorraine", nil, nil},
	"random@beat.gouv.fr":       User{"B", "random@beat.gouv.fr", "John", "Random", "LISTENS THE WIND", "Recouvrement et accompagnement des entreprises", "PENTAGON", "", nil, "Guadeloupe", nil, nil},
	keycloakAdmin:               ADMIN,
}
