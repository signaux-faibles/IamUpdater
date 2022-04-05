package main

import (
	"github.com/Nerzal/gocloak/v11"
	"github.com/spf13/viper"
	"log"
)

func ConfigureClients(kc KeycloakContext) {
	rawConfig := viper.Get("clients")
	clients := rawConfig.([]interface{})
	for _, client := range clients {
		clientConfig := client.(map[string]interface{})
		configureClient(kc, clientConfig)
	}
}

func configureClient(kc KeycloakContext, config map[string]interface{}) {
	clientId := getString(config["clientId"])
	clientToConfigure := gocloak.Client{}
	if existingClient := kc.getClientByClientId(*clientId); existingClient.ID != nil {
		clientToConfigure = existingClient
	}
	log.Printf("configure client %s", *clientId)
	clientToConfigure.ClientID = clientId
	clientToConfigure.Name = clientId
	clientToConfigure.DirectAccessGrantsEnabled = getBool(config["directAccessGrantsEnabled"])

	kc.updateClient(clientToConfigure)
}

func getString(value interface{}) *string {
	if value == nil {
		return nil
	}
	r := value.(string)
	return &r
}

func getBool(value interface{}) *bool {
	if value == nil {
		return nil
	}
	r := value.(bool)
	return &r
}
