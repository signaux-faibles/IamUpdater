package config

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/imdario/mergo"

	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
)

func merge(first structs.Config, second structs.Config) structs.Config {
	allClients := concatClients(first.Clients, second.Clients)
	err := mergo.Merge(&first, second, mergo.WithOverride)
	if err != nil {
		logger.Error("erreur pendant le merging de la configuration", logger.ContextForMethode(merge), err)
	}
	first.Clients = allClients
	return first
}

func concatClients(first []*gocloak.Client, second []*gocloak.Client) []*gocloak.Client {
	r := make([]*gocloak.Client, 0)
	if first != nil {
		r = append(r, first[:]...)
	}
	if second != nil {
		r = append(r, second[:]...)
	}
	return r
}
