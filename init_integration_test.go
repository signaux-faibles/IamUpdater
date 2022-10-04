//go:build integration
// +build integration

// nolint:errcheck
package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest/v3/docker"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/libwekan"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var kc KeycloakContext

var signauxfaibleClientID = "signauxfaibles"
var cwd, _ = os.Getwd()
var mongoUrl string
var mongodb *dockertest.Resource

var ctx = context.Background()

const keycloakAdmin = "ti_admin"
const keycloakPassword = "pwd"

func TestMain(m *testing.M) {
	var err error
	pool, err := dockertest.NewPool("")
	pool.MaxWait = time.Minute * 2

	if err != nil {
		logger.Panicf("Could not connect to docker: %s", err)
	}

	var keycloak *dockertest.Resource
	if !(os.Getenv("DISABLE_KEYCLOAK") == "yes") {
		keycloak = startKeycloak(pool)
	}
	mongodb = startWekanDB(pool)

	// excelUsers1, _, err := loadExcel("test/resources/wekanUpdate_states/1.xlsx")
	if err != nil {
		logger.Panicf("Could not read excel test cases")
	}

	//excelUsers2, _, err = loadExcel("test/resources/wekanUpdate_states/2.xlsx")
	if err != nil {
		logger.Panicf("Could not read excel test cases")
	}

	code := m.Run()

	if !(os.Getenv("DISABLE_KEYCLOAK") == "yes") {
		kill(keycloak)
	}
	kill(mongodb)
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

func startKeycloak(pool *dockertest.Pool) *dockertest.Resource {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	fields := logger.DataForMethod("startKeycloak")

	// pulls an image, creates a container based on it and runs it
	keycloakContainerName := "keycloakUpdater-ti-" + strconv.Itoa(time.Now().Nanosecond())
	fields.AddAny("container", keycloakContainerName)
	logger.Info("Démarre keycloak", fields)

	keycloak, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Name:       keycloakContainerName,
			Repository: "ghcr.io/signaux-faibles/conteneurs/keycloak",
			Tag:        "v1.0.0",
			Env: []string{
				"KEYCLOAK_USER=" + keycloakAdmin,
				"KEYCLOAK_PASSWORD=" + keycloakPassword,
				"DB_VENDOR=h2",
			},
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
	// container stops after 120 seconds
	if err = keycloak.Expire(240); err != nil {
		kill(keycloak)
		logger.ErrorE("Could not set expiration on container keycloak", fields, err)
	}

	logger.Infof("keycloak a démarré avec l'admin %v", keycloakAdmin)
	keycloakPort := keycloak.GetPort("8080/tcp")
	fields.AddAny("port", keycloakPort)
	logger.Info("keycloak started", fields)
	//exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		var err error
		kc, err = Init("http://localhost:"+keycloakPort, "master", keycloakAdmin, keycloakPassword)
		if err != nil {
			logger.Info("keycloak n'est pas prêt", fields)
			return err
		}
		return nil
	}); err != nil {
		logger.Panicf("Could not connect to keycloak: %s", err)
	}
	logger.Info("keycloak est prêt", fields)
	return keycloak
}

func startWekanDB(pool *dockertest.Pool) *dockertest.Resource {
	fields := logger.DataForMethod("startWekanDB")
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
			Mounts: []string{cwd + "/test/resources/dump_wekan/:/dump/"},
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
		fmt.Println(err.Error())
		kill(mongodb)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		logger.Info("Mongo n'est pas encore prêt", fields)
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
	//err = restoreMongoDump(mongodb)
	//if err != nil {
	//	logger.ErrorE("Erreur lors de la restauration du dump : %s", fields, err)
	//}
	//globalWekan, err = libwekan.Init(context.Background(), mongoUrl, "wekan", "signaux.faibles")
	//if err != nil {
	//	logger.ErrorE("Erreur lors de la creation de l'objet libwekan.Wekan : %s", fields, err)
	//}
	logger.Info("Mongo est prêt", fields)
	return mongodb
}

//func restoreMongoDump(mongodb *dockertest.Resource) error {
//	fields := logger.DataForMethod("restoreMongoDump")
//	var output bytes.Buffer
//	outputWriter := bufio.NewWriter(&output)
//
//	dockerOptions := dockertest.ExecOptions{
//		StdOut: outputWriter,
//		StdErr: outputWriter,
//	}
//	command := "mongorestore  --uri mongodb://root:password@localhost/ /dump"
//	fields.AddAny("command", command)
//	logger.Info("Restaure le dump", fields)
//	if exitCode, err := mongodb.Exec([]string{"/bin/bash", "-c", command}, dockerOptions); err != nil {
//		fields.AddAny("exitCode", exitCode)
//		logger.ErrorE("Erreur lors de la restauration du dump", fields, err)
//		return err
//	}
//	return outputWriter.Flush()
//}

func restoreMongoDumpInDatabase(mongodb *dockertest.Resource, suffix string, t *testing.T, slugDomainRegexp string) libwekan.Wekan {
	databasename := t.Name() + suffix
	fields := logger.DataForMethod("restoreMongoDump")
	fields.AddAny("database", databasename)
	var output bytes.Buffer
	outputWriter := bufio.NewWriter(&output)

	dockerOptions := dockertest.ExecOptions{
		StdOut: outputWriter,
		StdErr: outputWriter,
	}
	command := fmt.Sprintf("mongorestore  --uri mongodb://root:password@localhost/ /dump --nsFrom \"wekan.*\" --nsTo \"%s.*\"", databasename)
	fields.AddAny("command", command)
	logger.Info("Restaure le dump", fields)
	if exitCode, err := mongodb.Exec([]string{"/bin/bash", "-c", command}, dockerOptions); err != nil {
		fields.AddAny("exitCode", exitCode)
		logger.ErrorE("Erreur lors de la restauration du dump", fields, err)
		require.Nil(t, err)
	}
	err := outputWriter.Flush()
	require.Nil(t, err)
	wekan, err := initWekan(mongoUrl, databasename, "signaux.faibles", slugDomainRegexp)
	require.Nil(t, err)
	return wekan
}
