package main

import (
	"fmt"
	"github.com/Nerzal/gocloak/v11"
	"log"
)

type ClientConfigurator struct {
	toConfigure   *gocloak.Client
	configuration map[string]interface{}
}

// GetConfig implements Configurator
func (cc *ClientConfigurator) GetConfig() map[string]interface{} {
	return cc.configuration
}

func NewClient(clientID string, config map[string]interface{}) ClientConfigurator {
	r :=
		ClientConfigurator{
			toConfigure:   &gocloak.Client{Name: &clientID, ClientID: &clientID},
			configuration: config,
		}
	return r
}

func ExistingClient(client *gocloak.Client, config map[string]interface{}) ClientConfigurator {
	r :=
		ClientConfigurator{
			toConfigure:   client,
			configuration: config,
		}
	return r
}

func (cc *ClientConfigurator) Configure(kc KeycloakContext) {
	if cc.toConfigure == nil {
		log.Fatal("client to configure is nil")
	}
	updateStringParam(cc, "adminUrl", func(param *string) {
		cc.toConfigure.AdminURL = param
	})
	updateBoolParam(cc, "authorizationServicesEnabled", func(b *bool) {
		cc.toConfigure.AuthorizationServicesEnabled = b
	})
	updateBoolParam(cc, "bearerOnly", func(b *bool) {
		cc.toConfigure.BearerOnly = b
	})
	updateBoolParam(cc, "directAccessGrantsEnabled", func(b *bool) {
		cc.toConfigure.DirectAccessGrantsEnabled = b
	})
	updateBoolParam(cc, "implicitFlowEnabled", func(b *bool) {
		cc.toConfigure.ImplicitFlowEnabled = b
	})
	updateBoolParam(cc, "publicClient", func(b *bool) {
		cc.toConfigure.PublicClient = b
	})
	updateStringArrayParam(cc, "redirectUris", func(param *[]string) {
		cc.toConfigure.RedirectURIs = param
	})
	updateStringParam(cc, "rootUrl", func(param *string) {
		cc.toConfigure.RootURL = param
	})
	updateBoolParam(cc, "serviceAccountsEnabled", func(b *bool) {
		cc.toConfigure.ServiceAccountsEnabled = b
	})
	updateStringArrayParam(cc, "webOrigins", func(param *[]string) {
		cc.toConfigure.WebOrigins = param
	})
	updateMapOfStringsParam(cc, "attributes", func(param *map[string]string) {
		cc.toConfigure.Attributes = param
	})
	// listing des clés non utilisées
	// il faut prévenir l'utilisateur
	for k := range cc.configuration {
		log.Printf("%s - CAUTION : param '%s' is not yet implemented !!", cc, k)
	}
	kc.updateClient(*cc.toConfigure)
	log.Printf("%s has been configured and updated", cc)
}

func (cc ClientConfigurator) String() string {
	return fmt.Sprintf("ClientConfigurator %s", *cc.toConfigure.ClientID)
}
