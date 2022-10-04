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
	fields := logger.DataForMethod("main()")

	clientId := conf.Stock.ClientForRoles
	kc, err := NewKeycloakContext(conf.Access)
	if err != nil {
		logger.Panic(err)
	}

	// loading desired state for users, composites roles
	logger.Info("lecture du fichier excel stock", fields)
	users, compositeRoles, err := loadExcel(conf.Stock.UsersAndRolesFilename)
	if err != nil {
		logger.Panic(err)
	}

	if err = UpdateKeycloak(
		&kc,
		clientId,
		conf.Realm,
		conf.Clients,
		users,
		compositeRoles,
		Username(conf.Access.Username),
		AcceptedChanges,
	); err != nil {
		panic(err)
	}

	err = WekanUpdate(
		conf.Mongo.Url,
		conf.Mongo.Database,
		conf.Wekan.AdminUsername,
		users,
		conf.Wekan.SlugDomainRegexp,
	)

	if err != nil {
		panic(err)
	}
}
