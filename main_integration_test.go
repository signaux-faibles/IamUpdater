//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"github.com/Nerzal/gocloak/v11"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strconv"
	"testing"
	"time"
)

var kc KeycloakContext
var signauxfaibleClientID = "signauxfaibles"

const keycloakAdmin = "ti_admin"
const keycloakPassword = "pwd"

func TestMain(m *testing.M) {
	fields := logger.DataForMethod("TestMain")

	var err error

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Panicf("Could not connect to docker: %s", err)
	}
	// pulls an image, creates a container based on it and runs it
	keycloakContainerName := "keycloakUpdater-ti-" + strconv.Itoa(time.Now().Nanosecond())
	fields.AddAny("container", keycloakContainerName)
	logger.Info("trying start keycloak", fields)

	keycloak, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       keycloakContainerName,
			Repository: "ghcr.io/signaux-faibles/conteneurs/keycloak",
			Tag:        "v1.0.0",
			Env:        []string{"KEYCLOAK_USER=" + keycloakAdmin, "KEYCLOAK_PASSWORD=" + keycloakPassword},
		},
		func(config *docker.HostConfig) {
			// set AutoRemove to true so that stopped container goes away by itself
			config.AutoRemove = true
			config.RestartPolicy = docker.RestartPolicy{
				Name: "no",
			}
		},
	)
	if err != nil {
		kill(keycloak)
		logger.ErrorE("Could not start keycloak", fields, err)
	}
	// container stops after 60 seconds
	if err = keycloak.Expire(120); err != nil {
		kill(keycloak)
		logger.ErrorE("Could not set expiration on container keycloak", fields, err)
	}
	logger.Infof("keycloak started with username %v", keycloakAdmin)
	keycloakPort := keycloak.GetPort("8080/tcp")
	fields.AddAny("port", keycloakPort)
	logger.Info("keycloak started", fields)
	//exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		kc, err = Init("http://localhost:"+keycloakPort, "master", keycloakAdmin, keycloakPassword)
		if err != nil {
			logger.Info("keycloak is not ready", fields)
			return err
		}
		return nil
	}); err != nil {
		logger.Panicf("Could not connect to keycloak: %s", err)
	}

	code := m.Run()
	kill(keycloak)
	// You can't defer this because os.Exit doesn't care for defer

	os.Exit(code)
}

func kill(resource *dockertest.Resource) {
	if resource == nil {
		return
	}
	if err := resource.Close(); err != nil {
		logger.Panicf("Could not purge resource: %s", err)
	}
}

func TestKeycloakConfiguration_access_username_should_be_present_in_stock_file(t *testing.T) {
	asserte := assert.New(t)

	testUser := "ti_admin"
	testFilenames := []string{
		"test/users/john_doe.yml",
		"test/users/raphael_squelbut.yml",
		"test/users/un_mec_pas_de_l_urssaf.yml",
	}

	// erreur in configuration : access.username should be in usersFolder
	err := UpdateAll(
		&kc,
		"peuimporte",
		nil,
		nil,
		testFilenames,
		"regions_et_departements.csv",
		testUser,
		10,
	)

	asserte.NotNil(err)
	expectedMessage := fmt.Sprintf("configured user is not in users folder (%s) : %s", testFilenames, testUser)
	asserte.EqualError(err, expectedMessage)
}

func TestKeycloakInitialisation(t *testing.T) {
	asserte := assert.New(t)
	var conf structs.Config
	var err error
	if conf, err = config.InitConfig("test/resources/initialisation/test_config.toml"); err != nil {
		panic(err)
	}
	//configure logger
	logger.ConfigureWith(*conf.Logger)

	usersFilenames := []string{
		"test/users/admin.yml",
		"test/users/john_doe.yml",
		"test/users/raphael_squelbut.yml",
		"test/users/un_mec_pas_de_l_urssaf.yml",
	}

	// update all
	if err = UpdateAll(
		&kc,
		conf.Stock.DefaultClient,
		conf.Realm,
		conf.Clients,
		usersFilenames,
		"regions_et_departements.csv",
		conf.Access.Username,
		10,
	); err != nil {
		t.Fatalf("erreur pendant l'update : %v", err)
	}

	// assertions about realm
	asserte.Equal("master", *kc.Realm.Realm)
	asserte.Equal("Signaux Faibles", *kc.Realm.DisplayName)
	asserte.Equal("signaux-faibles", *kc.Realm.EmailTheme)
	asserte.Equal("signaux-faibles", *kc.Realm.LoginTheme)
	asserte.NotNil(kc.Realm.SMTPServer)
	asserte.Equal(5, *kc.Realm.MinimumQuickLoginWaitSeconds)
	asserte.True(*kc.Realm.RememberMe)

	clientSF, found := kc.getClient(signauxfaibleClientID)
	asserte.True(found)
	asserte.NotNil(clientSF)
	asserte.True(*clientSF.PublicClient)
	asserte.Contains(*clientSF.RedirectURIs, "https://signaux-faibles.beta.gouv.fr/*", "https://localhost:8080/*")
	asserte.Len(kc.ClientRoles[signauxfaibleClientID], 145)

	clientAnother, found := kc.getClient("another")
	asserte.True(found)
	asserte.NotNil(clientAnother)
	asserte.False(*clientAnother.PublicClient)
	asserte.Len(kc.ClientRoles["another"], 1) // il y a au minimum 1 rôle pour 1 client

	user, err := kc.GetUser("raphael.squelbut@shodo.io")
	asserte.Nil(err)
	asserte.NotNil(user)

	emptyUser := gocloak.User{}
	if user != emptyUser {
		err = logUser(*clientSF, user)
		asserte.Nil(err)
	}
}

// TestClientSignauxFaiblesExists teste l'existence du Client "signauxfaibles" par l'API
func TestClientSignauxFaiblesExists(t *testing.T) {
	asserte := assert.New(t)

	searchClientRequest := gocloak.GetClientsParams{
		ClientID: &signauxfaibleClientID,
	}
	clientSG, err := kc.API.GetClients(context.Background(), kc.JWT.AccessToken, *kc.Realm.Realm, searchClientRequest)
	if err != nil {
		t.Fatalf("error getting keycloak users by role pge : %v", err)
	}
	//clientSG := searchClientByName(t, "signauxfaibles")
	actual, found := kc.getClient(signauxfaibleClientID)
	asserte.True(found)
	asserte.Len(clientSG, 1)
	asserte.Contains(clientSG, actual)
}

func TestRolesExistences(t *testing.T) {

	asserte := assert.New(t)
	rolesToTest := []string{"urssaf", "dgefp", "bdf", "score", "detection", "pge"}

	for _, role := range rolesToTest {

		searchClientRolePGE := gocloak.GetRoleParams{
			Search: &role,
		}

		clientSG, _ := kc.getClient(signauxfaibleClientID)
		rolesFromAPI, err := kc.API.GetClientRoles(
			context.Background(),
			kc.JWT.AccessToken,
			*kc.Realm.Realm,
			*clientSG.ID,
			searchClientRolePGE,
		)
		if err != nil {
			t.Fatalf("error getting client roles : %v", err)
		}
		asserte.Lenf(rolesFromAPI, 1, "erreur pour le rôle %v", role)
		expected := rolesFromAPI[0]

		// on compare les résultats de l'API avec le contenu de l'objet KeycloakContext
		clientRolesFromContext := kc.ClientRoles[signauxfaibleClientID]
		asserte.Contains(clientRolesFromContext, expected)

		actual := kc.GetRoleFromRoleName(signauxfaibleClientID, role)
		asserte.Equalf(expected, actual, "erreur pour le rôle %v", role)
	}
}

func TestRolesAssignedToAll(t *testing.T) {
	asserte := assert.New(t)
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

		// il y a actuellement 2 users dans le fichier de provisionning excel
		// les 2 doivent avoir le rôle urssaf
		asserte.Len(usersFromAPI, 3)
	}
}

func TestKeycloak_should_not_update_when_too_many_changes(t *testing.T) {
	asserte := assert.New(t)
	var err error
	var conf structs.Config

	if conf, err = config.InitConfig("test/resources/update/test_config.toml"); err != nil {
		panic(err)
	}

	usersFilenames := []string{
		"test/users/admin.yml",
		"test/users/john_doe_v2.yml",
		"test/users/raphael_squelbut_v2.yml",
	}

	// configure logger
	logger.ConfigureWith(*conf.Logger)

	stdin := readStdin("false\n")
	// update all
	actual := UpdateAll(
		&kc,
		conf.Stock.DefaultClient,
		conf.Realm,
		conf.Clients,
		usersFilenames,
		"regions_et_departements.csv",
		conf.Access.Username,
		4,
	)
	os.Stdin = stdin
	asserte.EqualError(actual, "Trop de modifications utilisateurs.")
}

func TestKeycloakUpdate(t *testing.T) {
	asserte := assert.New(t)
	var conf structs.Config

	// voir le fichier
	// le user raphael.squelbut@shodo.io a été créé au test précédent
	disabledUser, err := kc.GetUser("raphael.squelbut@shodo.io")
	asserte.Nil(err)
	asserte.NotNil(disabledUser)

	clientSF, found := kc.getClient(signauxfaibleClientID)
	asserte.True(found)

	// le user doit encore pouvoir se loguer
	// avant l'exécution de l'update
	err = logUser(*clientSF, disabledUser)
	asserte.Nil(err)

	if conf, err = config.InitConfig("test/resources/update/test_config.toml"); err != nil {
		panic(err)
	}
	// configure logger
	logger.ConfigureWith(*conf.Logger)
	usersFilenames := []string{
		"test/users/admin.yml",
		"test/users/john_doe_v2.yml",
		"test/users/raphael_squelbut_v2.yml",
	}
	// update all
	err = UpdateAll(
		&kc,
		conf.Stock.DefaultClient,
		conf.Realm,
		conf.Clients,
		usersFilenames,
		"test/resources/regions_et_departements_faux.csv",
		conf.Access.Username,
		10,
	)
	if err != nil {
		panic(err)
	}

	// des rôles ont été supprimés dans le fichier de rôles
	asserte.Len(kc.ClientRoles[signauxfaibleClientID], 26)

	// on vérifie
	err = logUser(*clientSF, disabledUser)
	asserte.NotNil(err)
	apiError, ok := err.(*gocloak.APIError)
	asserte.True(ok)
	asserte.Equal(400, apiError.Code)
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
