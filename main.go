package main

import (
	"flag"
	"fmt"

	"github.com/pkg/errors"

	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

var overridingConfigFilename string

func init() {
	const (
		emptyOverridingFilename = ""
		usage                   = "chemin vers le fichier de configuration"
	)
	flag.StringVar(&overridingConfigFilename, "config", emptyOverridingFilename, usage)
	flag.StringVar(&overridingConfigFilename, "c", emptyOverridingFilename, usage+" (shorthand)")

}

func main() {
	flag.Parse()
	conf, err := config.InitConfig("./config.toml")
	config.OverrideConfig(conf, overridingConfigFilename)
	if err != nil {
		panic(err)
	}

	logger.ConfigureWith(*conf.Logger)
	logContext := logger.ContextForMethod(main)

	// loading desired state for users, composites roles
	logger.Info("lecture du fichier excel stock", logContext)
	users, compositeRoles, err := loadExcel(conf.Stock.UsersAndRolesFilename)
	if err != nil {
		logger.Panic("erreur pendant la lecture du fichier Excel", logContext, err)
	}

	if conf.Keycloak != nil {
		clientId := conf.Stock.ClientForRoles
		kc, err := NewKeycloakContext(conf.Keycloak)
		if err != nil {
			logger.Panic("erreur pendant l'initialisation du contexte Keycloak'", logContext, err)
		}

		if err = UpdateKeycloak(
			&kc,
			clientId,
			conf.Realm,
			conf.Clients,
			users,
			compositeRoles,
			Username(conf.Keycloak.Username),
			conf.Stock.MaxChangesToAccept,
		); err != nil {
			logger.Error("erreur pendant la mise à jour de Keycloak", logContext, err)
		}
	}

	if conf.Mongo != nil && conf.Wekan != nil {
		err = WekanUpdate(
			conf.Mongo.Url,
			conf.Mongo.Database,
			conf.Wekan.AdminUsername,
			users,
			conf.Wekan.SlugDomainRegexp,
		)
	}

	if err != nil {
		logger.Error("le traitement s'est terminé de façon anormale", logContext, err)
		fmt.Println("======= Détail de l'erreur")
		printErrChain(err, 0)
	} else {
		logger.Info("le traitement s'est terminé correctement", logContext)
	}
}

func printErrChain(err error, i int) {
	if err != nil {
		fmt.Printf("%d: %+v\n", i, err)
		printErrChain(errors.Unwrap(err), i+1)
	}
}
