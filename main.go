package main

import (
	"github.com/signaux-faibles/keycloakUpdater/v2/config"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
)

func main() {
	conf, err := config.InitConfig("config.toml", "config.d")
	if err != nil {
		panic(err)
	}

	logger.ConfigureWith(*conf.Logger)

	clientId := conf.Stock.Target
	kc, err := NewKeycloakContext(conf.Access)
	if err != nil {
		logger.Panic(err)
	}

	if err = UpdateAll(&kc, clientId, *conf.Realm, conf.Clients, conf.Stock.Filename); err != nil {

		panic(err)
	}

}
