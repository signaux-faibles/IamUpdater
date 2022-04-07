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
	cc.updateStringParam("admin_url", func(param *string) {
		cc.toConfigure.AdminURL = param
	})
	cc.updateBoolParam("authorization_services_enabled", func(b *bool) {
		cc.toConfigure.AuthorizationServicesEnabled = b
	})
	cc.updateBoolParam("bearer_only", func(b *bool) {
		cc.toConfigure.BearerOnly = b
	})
	cc.updateBoolParam("direct_access_grants_enabled", func(b *bool) {
		cc.toConfigure.DirectAccessGrantsEnabled = b
	})
	cc.updateBoolParam("implicit_flow_enabled", func(b *bool) {
		cc.toConfigure.ImplicitFlowEnabled = b
	})
	cc.updateBoolParam("public_client", func(b *bool) {
		cc.toConfigure.PublicClient = b
	})
	cc.updateStringArrayParam("redirect_uris", func(param *[]string) {
		cc.toConfigure.RedirectURIs = param
	})
	cc.updateStringParam("root_url", func(param *string) {
		cc.toConfigure.RootURL = param
	})
	cc.updateBoolParam("service_accounts_enabled", func(b *bool) {
		cc.toConfigure.ServiceAccountsEnabled = b
	})
	cc.updateStringArrayParam("web_origins", func(param *[]string) {
		cc.toConfigure.WebOrigins = param
	})
	cc.updateMapOfStrings("attributes", func(param *map[string]string) {
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

func (cc *ClientConfigurator) updateBoolParam(key string, setter func(*bool)) {
	val, ok := cc.configuration[key]
	if !ok {
		return
	}
	delete(cc.configuration, key)
	t := val.(bool)
	setter(&t)
}

func (cc *ClientConfigurator) updateMapOfStrings(key string, setter func(*map[string]string)) {
	val, ok := cc.configuration[key]
	if !ok {
		return
	}
	delete(cc.configuration, key)
	r := map[string]string{}
	t := val.(map[string]interface{})
	for k, v := range t {
		r[k] = v.(string)
	}
	setter(&r)
}

func (cc *ClientConfigurator) updateStringParam(key string, setter func(*string)) {
	val, ok := cc.configuration[key]
	if !ok {
		return
	}
	delete(cc.configuration, key)
	r := val.(string)
	setter(&r)
}

func (cc *ClientConfigurator) updateStringArrayParam(key string, setter func(*[]string)) {
	val, ok := cc.configuration[key]
	if !ok {
		return
	}
	delete(cc.configuration, key)
	vals := val.([]interface{})
	var t []string
	for _, uri := range vals {
		t = append(t, uri.(string))
	}
	setter(&t)
}

func (cc ClientConfigurator) String() string {
	return fmt.Sprintf("ClientConfigurator %s", *cc.toConfigure.ClientID)
}
