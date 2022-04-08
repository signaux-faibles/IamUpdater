package main

import (
	"fmt"
	"github.com/Nerzal/gocloak/v11"
	"log"
)

type RealmConfigurator struct {
	toConfigure   *gocloak.RealmRepresentation
	configuration map[string]interface{}
}

// GetConfig implements Configurator
func (rc *RealmConfigurator) GetConfig() map[string]interface{} {
	return rc.configuration
}

func NewRealmConfigurator(name string, config map[string]interface{}) RealmConfigurator {
	r :=
		RealmConfigurator{
			toConfigure: &gocloak.RealmRepresentation{
				Realm: &name,
			},
			configuration: config,
		}
	return r
}

func (rc *RealmConfigurator) Configure(kc KeycloakContext) {
	if rc.toConfigure == nil {
		log.Fatal("client to configure is nil")
	}
	updateBoolParam(rc, "bruteForceProtected", func(param *bool) {
		rc.toConfigure.BruteForceProtected = param
	})
	updateStringParam(rc, "displayName", func(param *string) {
		rc.toConfigure.DisplayName = param
	})
	updateStringParam(rc, "displayNameHtml", func(param *string) {
		rc.toConfigure.DisplayNameHTML = param
	})
	updateStringParam(rc, "emailTheme", func(param *string) {
		rc.toConfigure.EmailTheme = param
	})
	updateStringParam(rc, "loginTheme", func(param *string) {
		rc.toConfigure.LoginTheme = param
	})
	updateIntParam(rc, "minimumQuickLoginWaitSeconds", func(param *int) {
		rc.toConfigure.MinimumQuickLoginWaitSeconds = param
	})
	updateBoolParam(rc, "rememberMe", func(param *bool) {
		rc.toConfigure.RememberMe = param
	})
	updateBoolParam(rc, "resetPasswordAllowed", func(param *bool) {
		rc.toConfigure.ResetPasswordAllowed = param
	})
	updateMapOfStringsParam(rc, "smtpServer", func(param *map[string]string) {
		rc.toConfigure.SMTPServer = param
	})

	// listing des clés non utilisées
	// il faut prévenir l'utilisateur
	for k := range rc.configuration {
		log.Printf("%s - CAUTION : param '%s' is not yet implemented !!", rc, k)
	}
	kc.Realm = rc.toConfigure
	kc.RefreshRealm()
	log.Printf("%s has been configured and updated", rc)
}

func (rc RealmConfigurator) String() string {
	return fmt.Sprintf("ClientConfigurator %s", *rc.toConfigure.Realm)
}
