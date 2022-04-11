package config

import (
	"github.com/BurntSushi/toml"
	"github.com/Nerzal/gocloak/v11"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

type Config struct {
	Access  *access                      `toml:"access"`
	Realm   *gocloak.RealmRepresentation `toml:"realm"`
	Clients []*gocloak.Client            `toml:"clients"`
}

type access struct {
	Address       string
	Username      string
	Password      string
	Realm         string
	DefaultClient string
	UsersFile     string
}

func InitConfig(configFilename, configFolder string) Config {
	var conf Config
	//var err error
	//var meta toml.MetaData
	filenames := getAllConfigFilenames(configFilename, configFolder)
	log.Printf("config files : %s", filenames)
	allConfig := readAllConfigFiles(filenames)
	for _, current := range allConfig {
		conf = merge(conf, current)
	}
	return conf
}

func (c Config) GetAddress() string {
	return c.Access.Address
}

func (c Config) GetDefaultClient() string {
	return c.Access.DefaultClient
}

func (c Config) GetRealm() string {
	return c.Access.Realm
}

func (c Config) GetUsername() string {
	return c.Access.Username
}

func (c Config) GetPassword() string {
	return c.Access.Password
}

func (c Config) GetUsersFile() string {
	return c.Access.UsersFile
}

func readAllConfigFiles(filenames []string) []Config {
	var r = make([]Config, 0)
	for _, filename := range filenames {
		config := extractConfig(filename)
		r = append(r, config)
	}
	return r
}

func getAllConfigFilenames(filename, folder string) []string {
	var r = make([]string, 0)
	// checking file exist
	var file *os.File
	var err error
	if file, err = os.Open(filename); err != nil {
		log.Panicf("error reading clients config file : %s", err)
	}
	log.Print(file)
	r = append(r, filename)
	var files []fs.FileInfo
	if files, err = ioutil.ReadDir(folder); err != nil {
		log.Panicf("error reading clients config folder : %s", err)
	}
	for _, f := range files {
		filename := folder + "/" + f.Name()
		if !strings.HasSuffix(filename, ".toml") {
			log.Printf("ignore config file %s", filename)
			continue
		}
		r = append(r, filename)
	}
	return r
}

func extractConfig(filename string) Config {
	var conf Config
	var err error
	var meta toml.MetaData
	if meta, err = toml.DecodeFile(filename, &conf); err != nil {
		log.Panicf("error decoding toml config file '%s': %s", filename, err)
	}
	if meta.Undecoded() != nil {
		for _, key := range meta.Undecoded() {
			log.Printf("Caution : key '%s' from config file '%s' is not used", key, filename)
		}
	}
	return conf
}

func merge(first Config, second Config) Config {
	r := Config{Clients: make([]*gocloak.Client, 0)}
	r.Access = mergeAccess(first.Access, second.Access)
	r.Realm = mergeRealm(first.Realm, second.Realm)
	r.Clients = mergeClients(first.Clients, second.Clients)
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

func mergeAccess(first *access, second *access) *access {
	if first != nil {
		return first
	}
	return second
}
