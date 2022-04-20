package main

import (
	"github.com/ory/dockertest/v3"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"os"
	"testing"
)

var kc KeycloakContext

func TestMain(m *testing.M) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Panicf("Could not connect to docker: %s", err)
	}
	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("ghcr.io/signaux-faibles/conteneurs/keycloak", "v1.0.0", []string{"KEYCLOAK_USER=kcadmin", "KEYCLOAK_PASSWORD=kcpwd"})
	if err != nil {
		kill(resource)
		logger.Panicf("Could not start resource: %s", err)
	}

	//exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		kc, err = Init("http://localhost:"+resource.GetPort("8080/tcp"), "master", "kcadmin", "kcpwd")
		if err != nil {
			logger.Infof("keycloak is not ready")
			return err
		}
		return nil
	}); err != nil {
		logger.Panicf("Could not connect to keycloak: %s", err)
	}

	code := m.Run()
	kill(resource)
	// You can't defer this because os.Exit doesn't care for defer

	os.Exit(code)
}

func kill(resource *dockertest.Resource) {
	if err := resource.Close(); err != nil {
		logger.Panicf("Could not purge resource: %s", err)
	}
}

func TestSomething(t *testing.T) {
	fields := logger.DataForMethod("TestSomething")
	fields.AddArray("nbClients", logger.ToStrings(kc.Clients))
	logger.Info("coucou", fields)
}

func TestAnotherThing(t *testing.T) {
	fields := logger.DataForMethod("TestSomething")
	fields.AddArray("nbUsers", logger.ToStrings(kc.Users))
	logger.Info("coucou", fields)
}
