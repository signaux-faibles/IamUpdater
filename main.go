package main

import (
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

const AcceptedChanges int = 10

func main() {
	conf, err := config.InitConfig("./config.toml")
	if err != nil {
		panic(err)
	}

	logger.ConfigureWith(*conf.Logger)

	clientId := conf.Stock.ClientForRoles
	kc, err := NewKeycloakContext(conf.Access)
	if err != nil {
		logger.Panic(err)
	}
	if err = UpdateAll(
		&kc,
		clientId,
		conf.Realm,
		conf.Clients,
		conf.Stock.UsersAndRolesFilename,
		Username(conf.Access.Username),
		AcceptedChanges,
	); err != nil {
		panic(err)
	}
	//WekanUpdate(")
}
