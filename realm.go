package main

import (
	"github.com/spf13/viper"
	"log"
)

func ConfigureRealm(kc *KeycloakContext) {
	log.Println("configure realm...")
	// configure login
	*kc.Realm.RememberMe = true
	*kc.Realm.ResetPasswordAllowed = true

	// configure email
	smtp := map[string]string{
		"host":            viper.GetString("email.host"),
		"port":            viper.GetString("email.port"),
		"from":            viper.GetString("email.from.address"),
		"fromDisplayName": viper.GetString("email.from.label"),
	}
	*kc.Realm.SMTPServer = smtp

	// configure display
	displayname := viper.GetString("realm.description.displayname")
	kc.Realm.DisplayName = &displayname
	displaynameHTML := viper.GetString("realm.description.displaynameHTML")
	kc.Realm.DisplayNameHTML = &displaynameHTML

	// configure theme
	theme := viper.GetString("realm.description.theme")
	kc.Realm.LoginTheme = &theme
	kc.Realm.EmailTheme = &theme

	// configure security
	*kc.Realm.BruteForceProtected = true
	*kc.Realm.MinimumQuickLoginWaitSeconds = 5
	log.Println("update realm")
	kc.RefreshRealm()
	log.Println("configure realm [OK]")
}
