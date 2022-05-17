package main

import (
	"fmt"
	"os"

	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

func main() {
	if len(os.Args) == 1 {
		keycloak()
		wekan()
	} else if len(os.Args) == 2 {
		if os.Args[1] == "keycloak" {
			keycloak()
		} else if os.Args[1] == "wekan" {
			wekan()
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

func keycloak() {
	conf, err := config.InitConfig("/workspace/config.toml")
	if err != nil {
		panic(err)
	}

	logger.ConfigureWith(*conf.Logger)

	clientId := conf.Stock.ClientForRoles
	kc, err := NewKeycloakContext(conf.Access)
	if err != nil {
		logger.Panic(err)
	}
	if err = UpdateAll(&kc, clientId, conf.Realm, conf.Clients, conf.Stock.UsersAndRolesFilename, conf.Access.Username); err != nil {
		panic(err)
	}
}

func wekan() {
	fmt.Println("Wekan boilerplate not (yet) implemented")
}
