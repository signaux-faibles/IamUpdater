package config

import (
	"github.com/BurntSushi/toml"
	"github.com/Nerzal/gocloak/v13"
	"github.com/pkg/errors"
	"github.com/signaux-faibles/keycloakUpdater/v2/logger"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"os"
	"strings"
)

func InitConfig(configFilename string) (structs.Config, error) {
	var conf structs.Config
	//var err error
	//var meta toml.MetaData
	filenames := getAllConfigFilenames(configFilename)
	logger.Infof("config files : %s", filenames)
	allConfig := readAllConfigFiles(filenames)
	for _, current := range allConfig {
		conf = merge(conf, current)
	}
	if conf.Stock == nil {
		return structs.Config{}, errors.Errorf("[config] Stock is not configured")
	}
	return conf, nil
}

func readAllConfigFiles(filenames []string) []structs.Config {
	var r = make([]structs.Config, 0)
	for _, filename := range filenames {
		config := extractConfig(filename)
		r = append(r, config)
	}
	return r
}

func getAllConfigFilenames(filename string) []string {
	var r = make([]string, 0)
	// checking file exist
	var err error
	if _, err = os.Open(filename); err != nil {
		logger.Panicf("error reading clients config file : %s", err)
	}
	r = append(r, filename)
	var files []os.DirEntry
	config := extractConfig(filename)
	folder := config.Stock.ClientsAndRealmFolder
	if folder == "" {
		logger.Warnf("no configuration folder is defined")
		return r
	}
	stockFilename := config.Stock.UsersAndRolesFilename
	if stockFilename != "" {
		if _, err = os.ReadFile(stockFilename); err != nil {
			logger.Panicf("error reading stock file '%s' : %s", stockFilename, err)
		}
	}
	if files, err = os.ReadDir(folder); err != nil {
		logger.Panicf("error reading clients config folder : %s", err)
	}
	for _, f := range files {
		filename := folder + "/" + f.Name()
		if !strings.HasSuffix(filename, ".toml") {
			logger.Debugf("ignore config file %s", filename)
			continue
		}
		r = append(r, filename)
	}
	return r
}

func extractConfig(filename string) structs.Config {
	var conf structs.Config
	var err error
	var meta toml.MetaData
	if meta, err = toml.DecodeFile(filename, &conf); err != nil {
		logger.Panicf("error decoding toml config file '%s': %s", filename, err)
	}
	if meta.Undecoded() != nil {
		for _, key := range meta.Undecoded() {
			logger.Warnf("Caution : key '%s' from config file '%s' is not used", key, filename)
		}
	}
	return conf
}

func merge(first structs.Config, second structs.Config) structs.Config {
	r := structs.Config{Clients: make([]*gocloak.Client, 0)}
	r.Stock = firstNonNil(first.Stock, second.Stock)
	r.Logger = firstNonNil(first.Logger, second.Logger)
	r.Keycloak = mergeAccess(first.Keycloak, second.Keycloak)
	r.Realm = mergeRealm(first.Realm, second.Realm)
	r.Clients = mergeClients(first.Clients, second.Clients)
	r.Mongo = firstNonNil(first.Mongo, second.Mongo)
	r.Wekan = firstNonNil(first.Wekan, second.Wekan)
	return r
}

func mergeClients(first []*gocloak.Client, second []*gocloak.Client) []*gocloak.Client {
	r := make([]*gocloak.Client, 0)
	if first != nil {
		r = append(r, first[:]...)
	}
	if second != nil {
		r = append(r, second[:]...)
	}
	return r
}

func mergeRealm(first *gocloak.RealmRepresentation, second *gocloak.RealmRepresentation) *gocloak.RealmRepresentation {
	if first != nil {
		return first
	}
	return second
}

func firstNonNil[T any](first *T, second *T) *T {
	if first != nil {
		return first
	}
	return second
}

func mergeAccess(first *structs.Keycloak, second *structs.Keycloak) *structs.Keycloak {
	if first != nil {
		return first
	}
	return second
}
