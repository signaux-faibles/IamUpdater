package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"keycloakUpdater/v2/pkg/structs"
)

func Test_InitConfig(t *testing.T) {
	ass := assert.New(t)
	currentConfigFile := "test_config.toml"
	config, err := InitConfig(currentConfigFile)
	ass.NoError(err)
	ass.NotNil(config)

	expectedKeycloak := structs.Keycloak{
		Address:  "http://localhost:8080",
		Username: "tuadmin",
		Password: "tupwd",
		Realm:    "master",
	}
	ass.Equal(expectedKeycloak, *config.Keycloak)

	expectedStock := structs.Stock{
		ClientsAndRealmFolder: "../test/sample/test_config.d",
		ClientForRoles:        "signauxfaibles",
		UsersAndRolesFilename: "../test/sample/userBase.xlsx",
		BoardsConfigFilename:  "",
		MaxChangesToAccept:    0,
	}
	ass.Equal(expectedStock, *config.Stock)

	expectedLogger := structs.LoggerConfig{
		Filename:        "roles-test.log",
		Level:           "TRACE",
		TimestampFormat: "2006-01-02 15:04:05",
	}
	ass.Equal(expectedLogger, *config.Logger)

	ass.Equal("Signaux Faibles", *config.Realm.DisplayName)
	ass.Equal("<div class=\"kc-logo-text\"><span>Signaux Faibles</span></div>", *config.Realm.DisplayNameHTML)
	ass.Equal("signaux-faibles", *config.Realm.EmailTheme)
	smtp := *config.Realm.SMTPServer
	ass.NotNil(smtp)
	ass.Equal("noreply@localhost", smtp["from"])
	ass.Equal("Authentification Signaux Faibles", smtp["fromDisplayName"])

	ass.Len(config.Clients, 2)
	// ouais c'est la honte
	// la prochaine fois ça sera mieux ;)
	clientSF := *config.Clients[1]
	ass.Equal("signauxfaibles", *clientSF.ClientID)
	ass.Equal("signauxfaibles", *clientSF.Name)
	ass.Contains(*clientSF.RedirectURIs, "https://signaux-faibles.beta.gouv.fr/*")
	ass.Contains(*clientSF.RedirectURIs, "https://localhost:8080/*")
}

func Test_OverrideConfig(t *testing.T) {
	ass := assert.New(t)
	currentConfigFile := "test_config.toml"
	original, err := InitConfig(currentConfigFile)
	ass.NoError(err)
	ass.NotNil(original)
	overrided := OverrideConfig(original, "test_overriding_config.toml")

	expectedKeycloak := structs.Keycloak{
		Address:  "http://anotherkeycloak:8080",
		Username: "tuadmin",
		Password: "tupwd",
		Realm:    "master",
	}
	ass.Equal(expectedKeycloak, *overrided.Keycloak)

	expectedStock := structs.Stock{
		ClientsAndRealmFolder: "../test/sample/test_config.d",
		ClientForRoles:        "signauxfaibles",
		UsersAndRolesFilename: "empty_test_file.txt",
		BoardsConfigFilename:  "",
		MaxChangesToAccept:    0,
	}
	ass.Equal(expectedStock, *overrided.Stock)

}

func Test_getAllConfigFilenames(t *testing.T) {
	assertions := assert.New(t)
	currentConfigFile := "test_config.toml"
	expected := []string{
		currentConfigFile,
		"../test/sample/test_config.d/another.toml",
		"../test/sample/test_config.d/realm_master.toml",
		"../test/sample/test_config.d/client_signauxfaibles.toml",
	}

	// using the function
	mydir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(mydir)

	actual := getAllConfigFilenames(currentConfigFile)
	assertions.ElementsMatch(expected, actual)
}
