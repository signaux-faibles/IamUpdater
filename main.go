package main

import (
	"fmt"
	"os"

	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
)

func main() {
	conf, err := config.InitConfig("/workspace/config.toml")
	if err != nil {
		panic(err)
	}

	logger.ConfigureWith(*conf.Logger)

	if len(os.Args) == 1 {
		keycloak(conf)
		wekan(conf)
	} else if len(os.Args) == 2 {
		if os.Args[1] == "keycloak" {
			keycloak(conf)
		} else if os.Args[1] == "wekan" {
			wekan(conf)
		} else {
			usage()
		}
	} else {
		usage()
	}
}

func usage() {
	fmt.Println("usage: keycloakUpdater [keycloak|wekan]")
}

func keycloak(conf structs.Config) {
	clientId := conf.Stock.ClientForRoles
	kc, err := NewKeycloakContext(conf.Access)
	if err != nil {
		logger.Panic(err)
	}
	if err = UpdateAll(&kc, clientId, conf.Realm, conf.Clients, conf.Stock.UsersAndRolesFilename, conf.Access.Username); err != nil {
		panic(err)
	}
}

func wekan(conf structs.Config) {
	fmt.Println("Wekan boilerplate not (yet) implemented")
}
