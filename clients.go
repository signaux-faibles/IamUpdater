package main

import (
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
	cc.updateStringParam("adminUrl", func(param *string) {
		cc.toConfigure.AdminURL = param
	})
	cc.updateBoolParam("authorizationServicesEnabled", func(b *bool) {
		cc.toConfigure.AuthorizationServicesEnabled = b
	})
	cc.updateBoolParam("bearerOnly", func(b *bool) {
		cc.toConfigure.BearerOnly = b
	})
	cc.updateBoolParam("directAccessGrantsEnabled", func(b *bool) {
		cc.toConfigure.DirectAccessGrantsEnabled = b
	})
	cc.updateBoolParam("implicitFlowEnabled", func(b *bool) {
		cc.toConfigure.ImplicitFlowEnabled = b
	})
	cc.updateBoolParam("publicClient", func(b *bool) {
		cc.toConfigure.PublicClient = b
	})
	cc.updateStringArrayParam("redirectUris", func(param *[]string) {
		cc.toConfigure.RedirectURIs = param
	})
	cc.updateStringParam("rootUrl", func(param *string) {
		cc.toConfigure.RootURL = param
	})
	cc.updateBoolParam("serviceAccountsEnabled", func(b *bool) {
		cc.toConfigure.ServiceAccountsEnabled = b
	})
	cc.updateStringArrayParam("webOrigins", func(param *[]string) {
		cc.toConfigure.WebOrigins = param
	})
	cc.updateMapOfStrings("attributes", func(param *map[string]string) {
		cc.toConfigure.Attributes = param
	})
	kc.updateClient(*cc.toConfigure)
}

func (cc *ClientConfigurator) updateBoolParam(key string, setter func(*bool)) {
	val, ok := cc.configuration[key]
	if !ok {
		return
	}
	t := val.(bool)
	setter(&t)
}

func (cc *ClientConfigurator) updateMapOfStrings(key string, setter func(*map[string]string)) {
	val, ok := cc.configuration[key]
	if !ok {
		return
	}
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
	r := val.(string)
	setter(&r)
}

func (cc *ClientConfigurator) updateStringArrayParam(key string, setter func(*[]string)) {
	val, ok := cc.configuration[key]
	if !ok {
		return
	}
	vals := val.([]interface{})
	var t []string
	for _, uri := range vals {
		t = append(t, uri.(string))
	}
	setter(&t)
}
