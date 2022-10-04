package structs

import (
	"github.com/Nerzal/gocloak/v11"
)

type Stock struct {
	ClientsAndRealmFolder string
	ClientForRoles        string
	UsersAndRolesFilename string
}

type Config struct {
	Access  *Access                      `toml:"access"`
	Stock   *Stock                       `toml:"stock"`
	Logger  *LoggerConfig                `toml:"logger"`
	Realm   *gocloak.RealmRepresentation `toml:"realm"`
	Clients []*gocloak.Client            `toml:"clients"`
	Mongo   *Mongo                       `toml:"mongodb"`
	Wekan   *Wekan                       `toml:"wekan"`
}

type Mongo struct {
	Url      string `toml:"url"`
	Database string `toml:"database"`
}

type Wekan struct {
	AdminUsername string `toml:"adminUserName"`
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
