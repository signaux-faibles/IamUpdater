//go:build integration
// +build integration

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

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var kc KeycloakContext

// var wekan libwekan.Wekan
var signauxfaibleClientID = "signauxfaibles"
var cwd, _ = os.Getwd()
var mongoUrl string
var excelUsers Users
var excelUserMap map[string]Roles

const keycloakAdmin = "ti_admin"
const keycloakPassword = "pwd"

func TestMain(m *testing.M) {
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Panicf("Could not connect to docker: %s", err)
	}
	mongo := startWekanDB(pool)
	keycloak := startKeycloak(pool)
	excelUsers, excelUserMap, err = loadExcel("./userBase.xlsx")
	if err != nil {
		logger.Panicf("Could not read excel test cases")
	}

	code := m.Run()
	kill(keycloak)
	kill(mongo)
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
		logger.ErrorE("N'arrive pas à démarrer Mongo: %s", fields, err)
	}
	err = restoreMongoDump(mongodb)
	if err != nil {
		logger.ErrorE("Erreur lors de la restauration du dump : %s", fields, err)
	}
	//wekan, err := libwekan.Connect(context.TODO(), mongoUrl, "wekan", "root")
	//if err != nil {
	//	logger.ErrorE("Erreur lors de la connexion à wekan", fields, err)
	//}
	//users, err := wekan.GetUsers(context.TODO())
	//if err != nil {
	//	logger.ErrorE("Erreur lors de la récupération des users", fields, err)
	//}
	//logger.Infof("nombre de users :%v", users)
	logger.Info("Mongo est prêt et restauré", fields)
	return mongodb
}

func restoreMongoDump(mongodb *dockertest.Resource) error {
	fields := logger.DataForMethod("restoreMongoDump")
	var b bytes.Buffer
	output := bufio.NewWriter(&b)

	options := dockertest.ExecOptions{
		StdOut: output,
		StdErr: output,
	}
	//command := fmt.Sprintf("mongorestore --uri %s /dump", mongoUrl)
	//fields.AddAny("command", command)
	//logger.Info("Restaure le dump", fields)
	//if exitCode, err := mongodb.Exec([]string{"/bin/bash", "-c", command}, options); err != nil {
	if exitCode, err := mongodb.Exec([]string{"/bin/bash", "-c", "mongorestore  --uri mongodb://root:password@localhost/ /dump"}, options); err != nil {
		fields.AddAny("exitCode", exitCode)
		logger.ErrorE("Erreur lors de la restauration du dump", fields, err)
		return err
	}
	_, err := mongodb.Exec([]string{"/bin/bash", "-c", "mongo mongodb://root:password@localhost/wekan --authenticationDatabase admin --eval 'printjson(db.users.find({}).toArray())'"}, options)
	if err != nil {
		return err
	}
	return output.Flush()
}
