package structs

import (
	"github.com/Nerzal/gocloak/v11"
)

type Stock struct {
	ClientsAndRealmFolder string
	DefaultClient         string
	UsersFolder           string
}

type Config struct {
	Access  *Access                      `toml:"access"`
	Stock   *Stock                       `toml:"stock"`
	Logger  *LoggerConfig                `toml:"logger"`
	Realm   *gocloak.RealmRepresentation `toml:"realm"`
	Clients []*gocloak.Client            `toml:"clients"`
}

type Access struct {
	Address  string
	Username string
	Password string
	Realm    string
}

type LoggerConfig struct {
	Filename        string
	Level           string
	TimestampFormat string
	Rotation        bool
}
