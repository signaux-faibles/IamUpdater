package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	initConfig()
	// initialisation et connexion à keycloak

	clientId := viper.GetString("access.defaultClient")

	log.Printf("default clientId : %s", clientId)

	realmName := viper.GetString("access.realm")

	kc, err := NewKeycloakContext(
		viper.GetString("access.address"),
		realmName,
		viper.GetString("access.username"),
		viper.GetString("access.password"))
	if err != nil {
		panic(err)
	}

	// realmName config
	rawConfig := viper.GetStringMap("realm")
	masterConfigurator := NewRealmConfigurator(realmName, rawConfig)
	masterConfigurator.Configure(kc)

	// clients config
	clients := readClientConfigurations(kc)
	for _, client := range clients {
		client.Configure(kc)
	}
	if err = kc.refreshClients(); err != nil {
		log.Fatalf("error refreshing clients : %s", err)
	}

	// loading desired state for users, composites roles
	excelFileName := viper.GetString("users.file")
	users, compositeRoles, err := loadExcel(excelFileName)
	if err != nil {
		panic(err)
	}
	// gather roles, newRoles are created before users, oldRoles are deleted after users
	log.Println("checking roles and creating new ones")
	newRoles, oldRoles := neededRoles(compositeRoles, users).compare(kc.GetClientRoles()[clientId])

	i, err := kc.CreateClientRoles(clientId, newRoles)
	if err != nil {
		log.Printf("failed creating new roles: %s", err.Error())
		panic(err)
	}
	log.Printf("%d roles created", i)

	// check and adjust composite roles
	err = kc.ComposeRoles(
		clientId,
		compositeRoles,
	)
	if err != nil {
		fmt.Println(err)
	}

	// checking users
	missing, obsolete, update, current := users.Compare(kc)

	err = kc.CreateUsers(missing.GetNewGocloakUsers(), users, clientId)
	if err != nil {
		panic(err)
	}

	// disable obsolete users
	err = kc.DisableUsers(obsolete, clientId)
	if err != nil {
		panic(err)
	}
	// enable existing but disabled users
	err = kc.EnableUsers(update)
	if err != nil {
		panic(err)
	}

	// make sure every on has correct roles
	err = kc.UpdateCurrentUsers(current, users, clientId)
	if err != nil {
		panic(err)
	}

	// delete old roles
	if len(oldRoles) > 0 {
		log.Printf("removing unused roles: %s", strings.Join(oldRoles, ", "))
		internalID, err := kc.GetInternalIDFromClientID(clientId)
		if err != nil {
			panic(err)
		}
		for _, role := range oldRoles.GetKeycloakRoles(clientId, kc) {
			err = kc.API.DeleteClientRole(context.Background(), kc.JWT.AccessToken, kc.realm, internalID, *role.Name)
			if err != nil {
				panic(err)
			}
		}
	}
}

func initConfig() {
	viper.AddConfigPath("config.d")
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	err := viper.MergeInConfig()
	if err != nil {
		log.Fatalf("error reading config toml : %s", err)
	}
	files, err := ioutil.ReadDir("config.d")
	if err != nil {
		log.Fatalf("error reading clients config folder : %s", err)
	}
	for _, f := range files {
		filename := f.Name()
		if !strings.HasSuffix(filename, ".toml") {
			log.Printf("ignore config file %s", filename)
			continue
		}
		clientID, _, _ := strings.Cut(filename, ".toml")
		viper.SetConfigName(clientID)
		err := viper.MergeInConfig()
		if err != nil {
			log.Fatalf("error reading config file %s : %s", filename, err)
		}
	}
}

func readClientConfigurations(kc KeycloakContext) []ClientConfigurator {
	var r []ClientConfigurator

	// read fom main.toml
	rawConfig := viper.GetStringMap("client")
	for name, rawClient := range rawConfig {
		clientConfig := mainToConfig(rawClient)
		log.Printf("read config for client %s.....", name)
		if client := kc.getClientByClientId(name); client != nil {
			r = append(r, ExistingClient(client, clientConfig))
		} else {
			r = append(r, NewClient(name, clientConfig))
		}
		log.Printf("read config for client %s [OK]", name)
	}
	return r
}

func mainToConfig(raw interface{}) map[string]interface{} {
	array := raw.([]interface{})
	if len(array) != 1 {
		log.Fatalf("error in config -> %s", array)
	}
	r := array[0].(map[string]interface{})
	return r
}
