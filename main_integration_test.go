//go:build integration
// +build integration

package main

import (
	"context"
	"github.com/Nerzal/gocloak/v11"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"github.com/stretchr/testify/assert"
	"os"
	"strconv"
	"testing"
	"time"
)

var kc KeycloakContext
var conf structs.Config

func TestMain(m *testing.M) {
	fields := logger.DataForMethod("TestMain")

	var err error
	if conf, err = config.InitConfig("test/resources/test_config_v1.toml", "test/resources/test_config.d"); err != nil {
		panic(err)
	}
	// configure logger
	logger.ConfigureWith(*conf.Logger)

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Panicf("Could not connect to docker: %s", err)
	}
	// pulls an image, creates a container based on it and runs it
	keycloakContainerName := "keycloakUpdater-ti-" + strconv.Itoa(time.Now().Nanosecond())
	fields.AddAny("container", keycloakContainerName)
	logger.Info("trying start keycloak", fields)

	keycloak, err := pool.RunWithOptions(&dockertest.RunOptions{
		Name:       keycloakContainerName,
		Repository: "jboss/keycloak",
		Tag:        "16.1.1",
		Env:        []string{"KEYCLOAK_USER=" + conf.Access.Username, "KEYCLOAK_PASSWORD=" + conf.Access.Password},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})
	if err != nil {
		kill(keycloak)
		logger.ErrorE("Could not start keycloak", fields, err)
	}
	// container stops after 60 seconds
	if err = keycloak.Expire(120); err != nil {
		kill(keycloak)
		logger.ErrorE("Could not set expiration on container keycloak", fields, err)
	}
	keycloakPort := keycloak.GetPort("8080/tcp")
	fields.AddAny("port", keycloakPort)
	logger.Info("keycloak started", fields)
	//exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		kc, err = Init("http://localhost:"+keycloakPort, conf.Access.Realm, conf.Access.Username, conf.Access.Password)
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

func TestAllIntegration(t *testing.T) {
	asserte := assert.New(t)

	//var conf structs.Config
	var err error
	//if conf, err = config.InitConfig("test/resources/test_config.toml", "test/resources/test_config.d"); err != nil {
	//	panic(err)
	//}
	// configure logger
	//logger.ConfigureWith(*conf.Logger)

	// update all
	if err = UpdateAll(&kc, conf.Stock.Target, conf.Realm, conf.Clients, conf.Stock.Filename, conf.Access.Username); err != nil {
		panic(err)
	}

	// assertions about realm
	asserte.Equal("master", *kc.Realm.Realm)
	asserte.Equal("Signaux Faibles", *kc.Realm.DisplayName)
	asserte.Equal("signaux-faibles", *kc.Realm.EmailTheme)
	asserte.Equal("signaux-faibles", *kc.Realm.LoginTheme)
	asserte.NotNil(kc.Realm.SMTPServer)
	asserte.Equal(5, *kc.Realm.MinimumQuickLoginWaitSeconds)
	asserte.True(*kc.Realm.RememberMe)

	// assertions about clients
	// in config, 2 clients are configured "signauxfaibles" and "another"
	configuredClients := []string{"signauxfaibles", "another"}
	clientsMap := getConfiguredClients(configuredClients)

	clientSF := clientsMap["signauxfaibles"]
	asserte.NotNil(clientSF)
	asserte.True(*clientSF.PublicClient)
	asserte.Contains(*clientSF.RedirectURIs, "https://signaux-faibles.beta.gouv.fr/*", "https://localhost:8080/*")
	asserte.Len(kc.ClientRoles["signauxfaibles"], 144)

	clientAnother := clientsMap["another"]
	asserte.NotNil(clientAnother)
	asserte.False(*clientAnother.PublicClient)

	user, err := kc.GetUser("raphael.squelbut@shodo.io")
	asserte.Nil(err)
	asserte.NotNil(user)

	err = logUser(clientSF, user)
	asserte.Nil(err)
}

func TestSecondPassage(t *testing.T) {
	asserte := assert.New(t)
	// voir le fichier
	disabledUser, err := kc.GetUser("raphael.squelbut@shodo.io")
	asserte.Nil(err)
	asserte.NotNil(disabledUser)

	// in config, 2 clients are configured "signauxfaibles" and "another"
	configuredClients := []string{"signauxfaibles", "another"}
	clientSF := getConfiguredClients(configuredClients)["signauxfaibles"]

	if conf, err = config.InitConfig("test/resources/test_config_v2.toml", ""); err != nil {
		panic(err)
	}
	// configure logger
	logger.ConfigureWith(*conf.Logger)

	// update all
	err = UpdateAll(&kc, conf.Stock.Target, conf.Realm, conf.Clients, conf.Stock.Filename, conf.Access.Username)
	if err != nil {
		panic(err)
	}
	asserte.Len(kc.ClientRoles["signauxfaibles"], 25)
	err = logUser(clientSF, disabledUser)
	asserte.NotNil(err)
	apiError, ok := err.(*gocloak.APIError)
	asserte.True(ok)
	asserte.Equal(400, apiError.Code)
}

func contains(array []string, item string) bool {
	if array == nil {
		return false
	}
	for _, s := range array {
		if item == s {
			return true
		}
	}
	return false
}

func getConfiguredClients(configuredClients []string) map[string]gocloak.Client {
	// in config, 2 clients are configured "signauxfaibles" and "another"
	clientsMap := make(map[string]gocloak.Client, len(kc.ClientRoles))
	for _, client := range kc.Clients {
		if contains(configuredClients, *client.ClientID) {
			clientsMap[*client.ClientID] = *client
		}
	}
	return clientsMap
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
