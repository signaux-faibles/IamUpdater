//go:build integration

// nolint:errcheck
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gosimple/slug"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/signaux-faibles/libwekan"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"keycloakUpdater/v2/pkg/logger"
	"keycloakUpdater/v2/pkg/structs"
)

var kc KeycloakContext

var signauxfaibleClientID = "signauxfaibles"

var mongoUrl string
var mongodb *dockertest.Resource

var ctx = context.Background()

const keycloakAdmin = "ti_admin"
const keycloakPassword = "pwd"

func TestMain(m *testing.M) {
	logContext := logger.ContextForMethod(TestMain)
	var err error
	pool, err := dockertest.NewPool("")
	pool.MaxWait = time.Minute * 2
	logger.ConfigureWith(
		structs.LoggerConfig{
			Filename: "/dev/null",
		})
	if err != nil {
		logger.Panic("erreur pendant la connection à Docker", logContext, err)
	}

	keycloak := startKeycloak(pool)
	mongodb = startWekanDB(pool)

	code := m.Run()

	kill(keycloak)
	// kill(mongodb)
	// You can't defer this because os.Exit doesn't care for defer

	os.Exit(code)
}

func kill(resource *dockertest.Resource) {
	logContext := logger.ContextForMethod(kill)
	if resource == nil {
		return
	}
	if err := resource.Close(); err != nil {
		logger.Panic("erreur pendant la purge des ressources Docker", logContext, err)
	}
}

func startKeycloak(pool *dockertest.Pool) *dockertest.Resource {
	if os.Getenv("DISABLE_KEYCLOAK") == "yes" {
		return nil
	}
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	logContext := logger.ContextForMethod(startKeycloak)

	// pulls an image, creates a container based on it and runs it
	keycloakContainerName := "keycloakUpdater-ti-" + strconv.Itoa(time.Now().Nanosecond())
	logContext.AddString("container", keycloakContainerName)
	logger.Info("Démarre keycloak", logContext)

	keycloak, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       keycloakContainerName,
			Repository: "ghcr.io/signaux-faibles/conteneurs/keycloak",
			Tag:        "v23.0",
			Env: []string{
				"KEYCLOAK_ADMIN=" + keycloakAdmin,
				"KEYCLOAK_ADMIN_PASSWORD=" + keycloakPassword,
				"DB_VENDOR=h2",
			},
			Cmd: []string{
				"start-dev",
				"--http-relative-path=/auth",
				"--spi-login-protocol-openid-connect-legacy-logout-redirect-uri=true"},
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
		logger.Panic("Could not start keycloak", logContext, err)
	}
	// container stops after 120 seconds
	if err = keycloak.Expire(600); err != nil {
		kill(keycloak)
		logger.Error("Could not set expiration on container keycloak", logContext, err)
	}

	logger.Info("keycloak a démarré avec l'admin", logContext.AddAny("name", keycloakAdmin))
	keycloakPort := keycloak.GetPort("8080/tcp")
	logContext.AddAny("port", keycloakPort)
	//exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		kc, err = Init("http://localhost:"+keycloakPort+"/auth", "master", keycloakAdmin, keycloakPassword)
		if err != nil {
			logger.Trace("keycloak n'est pas prêt", logContext.AddAny("error", err))
			return err
		}
		return nil
	}); err != nil {
		logger.Panic("erreur pendant la connexion à Keycloak", logContext, err)
	}
	logger.Info("keycloak est prêt", logContext)
	return keycloak
}

func startWekanDB(pool *dockertest.Pool) *dockertest.Resource {
	dir, _ := os.Getwd()
	logContext := logger.ContextForMethod(startWekanDB)
	mongodbContainerName := "mongodb-ti-" + strconv.Itoa(time.Now().Nanosecond())
	mongodb, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       mongodbContainerName,
			Repository: "mongo",
			Tag:        "5.0",
			Env: []string{
				// username and password for mongodb superuser
				"MONGO_INITDB_ROOT_USERNAME=root",
				"MONGO_INITDB_ROOT_PASSWORD=password",
			},
			Mounts: []string{dir + "/test/resources/dump_wekan/:/dump/"},
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
		logger.Error("erreur pendant le démarrage de Mongo", logContext, err)
		kill(mongodb)
	}
	// container stops after 120 seconds
	if err = mongodb.Expire(600); err != nil {
		kill(mongodb)
		logger.Error("Could not set expiration on container mongoDb", logContext, err)
	}
	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		logger.Info("Mongo n'est pas encore prêt", logContext)
		var err error
		mongoUrl = fmt.Sprintf("mongodb://root:password@localhost:%s", mongodb.GetPort("27017/tcp"))

		var db *mongo.Client
		db, err = mongo.Connect(
			context.TODO(),
			options.Client().ApplyURI(mongoUrl),
		)
		if err != nil {
			return err
		}
		return db.Ping(context.TODO(), nil)
	}); err != nil {
		kill(mongodb)
		panic("N'arrive pas à démarrer Mongo")
	}

	logger.Info("Mongo est prêt", logContext)
	return mongodb
}

func restoreMongoDumpInDatabase(mongodb *dockertest.Resource, suffix string, t *testing.T, slugDomainRegexp string) libwekan.Wekan {
	databasename := t.Name() + suffix
	logContext := logger.ContextForMethod(restoreMongoDumpInDatabase).AddAny("database", databasename)
	var output bytes.Buffer
	outputWriter := bufio.NewWriter(&output)

	dockerOptions := dockertest.ExecOptions{
		StdOut: outputWriter,
		StdErr: outputWriter,
	}
	command := fmt.Sprintf("mongorestore  --uri mongodb://root:password@localhost/ /dump --nsFrom \"wekan.*\" --nsTo \"%s.*\"", databasename)
	logContext.AddAny("command", command)
	logger.Info("Restaure le dump", logContext)
	if exitCode, err := mongodb.Exec([]string{"/bin/bash", "-c", command}, dockerOptions); err != nil {
		logContext.AddAny("exitCode", exitCode)
		logger.Error("Erreur lors de la restauration du dump", logContext, err)
		require.NoError(t, err)
	}
	err := outputWriter.Flush()
	require.NoError(t, err)
	wekan, err := initWekan(mongoUrl, databasename, "signaux.faibles", slugDomainRegexp)
	require.NoError(t, err)
	return wekan
}

func createTempFilename(t *testing.T) string {
	return fmt.Sprint(t.TempDir(), os.PathSeparator, slug.Make(t.Name()))
}
