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

func (cc *RealmConfigurator) GetConfig() map[string]interface{} {
	return cc.configuration
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
	updateBoolParam(rc, "brute_force_protected", func(param *bool) {
		rc.toConfigure.BruteForceProtected = param
	})
	updateStringParam(rc, "display_name", func(param *string) {
		rc.toConfigure.DisplayName = param
	})
	updateStringParam(rc, "display_name_html", func(param *string) {
		rc.toConfigure.DisplayNameHTML = param
	})
	updateStringParam(rc, "email_theme", func(param *string) {
		rc.toConfigure.EmailTheme = param
	})
	updateStringParam(rc, "login_theme", func(param *string) {
		rc.toConfigure.LoginTheme = param
	})
	updateIntParam(rc, "minimum_quick_login_wait_seconds", func(param *int) {
		rc.toConfigure.MinimumQuickLoginWaitSeconds = param
	})
	updateBoolParam(rc, "remember_me", func(param *bool) {
		rc.toConfigure.RememberMe = param
	})
	updateBoolParam(rc, "reset_password_allowed", func(param *bool) {
		rc.toConfigure.ResetPasswordAllowed = param
	})
	updateMapOfStringsParam(rc, "smtp_server", func(param *map[string]string) {
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
