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
	rc.updateBoolParam("brute_force_protected", func(param *bool) {
		rc.toConfigure.BruteForceProtected = param
	})
	rc.updateStringParam("display_name", func(param *string) {
		rc.toConfigure.DisplayName = param
	})
	rc.updateStringParam("display_name_html", func(param *string) {
		rc.toConfigure.DisplayNameHTML = param
	})
	rc.updateStringParam("email_theme", func(param *string) {
		rc.toConfigure.EmailTheme = param
	})
	rc.updateStringParam("login_theme", func(param *string) {
		rc.toConfigure.LoginTheme = param
	})
	rc.updateIntParam("minimum_quick_login_wait_seconds", func(param *int) {
		rc.toConfigure.MinimumQuickLoginWaitSeconds = param
	})
	rc.updateBoolParam("remember_me", func(param *bool) {
		rc.toConfigure.RememberMe = param
	})
	rc.updateBoolParam("reset_password_allowed", func(param *bool) {
		rc.toConfigure.ResetPasswordAllowed = param
	})
	rc.updateMapOfStrings("smtp_server", func(param *map[string]string) {
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

func (rc *RealmConfigurator) updateBoolParam(key string, setter func(*bool)) {
	val, ok := rc.configuration[key]
	if !ok {
		return
	}
	delete(rc.configuration, key)
	t := val.(bool)
	setter(&t)
}

func (rc *RealmConfigurator) updateMapOfStrings(key string, setter func(*map[string]string)) {
	val, ok := rc.configuration[key]
	if !ok {
		return
	}
	delete(rc.configuration, key)
	r := map[string]string{}
	t := val.(map[string]interface{})
	for k, v := range t {
		r[k] = v.(string)
	}
	setter(&r)
}

func (rc *RealmConfigurator) updateIntParam(key string, setter func(param *int)) {
	val, ok := rc.configuration[key]
	if !ok {
		return
	}
	delete(rc.configuration, key)
	r := int(val.(int64))
	setter(&r)
}

func (rc *RealmConfigurator) updateStringParam(key string, setter func(*string)) {
	val, ok := rc.configuration[key]
	if !ok {
		return
	}
	delete(rc.configuration, key)
	r := val.(string)
	setter(&r)
}

func (rc *RealmConfigurator) updateStringArrayParam(key string, setter func(*[]string)) {
	val, ok := rc.configuration[key]
	if !ok {
		return
	}
	delete(rc.configuration, key)
	vals := val.([]interface{})
	var t []string
	for _, uri := range vals {
		t = append(t, uri.(string))
	}
	setter(&t)
}

func (rc RealmConfigurator) String() string {
	return fmt.Sprintf("ClientConfigurator %s", *rc.toConfigure.Realm)
}
