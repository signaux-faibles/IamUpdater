package config

import (
	"github.com/Nerzal/gocloak/v13"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_merge(t *testing.T) {
	ass := assert.New(t)
	wantedAccess := &structs.Keycloak{}
	wantedRealm := &gocloak.RealmRepresentation{}
	clientA := gocloak.Client{}
	clientB := gocloak.Client{}
	clientC := gocloak.Client{}
	wantedClients := []*gocloak.Client{&clientA, &clientB, &clientC}
	configA := structs.Config{
		Keycloak: wantedAccess,
		Realm:    nil,
		Clients:  []*gocloak.Client{&clientA},
	}
	configB := structs.Config{
		Keycloak: nil,
		Realm:    wantedRealm,
		Clients:  []*gocloak.Client{&clientB, &clientC},
	}
	type args struct {
		first  structs.Config
		second structs.Config
	}
	tests := []struct {
		name string
		args args
		want structs.Config
	}{
		{name: "merge Configs", args: args{first: configA, second: configB}, want: structs.Config{
			Keycloak: wantedAccess,
			Realm:    wantedRealm,
			Clients:  wantedClients,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := merge(tt.args.first, tt.args.second)
			ass.EqualValues(tt.want, got)
		})
	}
}

func Test_concatClients(t *testing.T) {
	assertions := assert.New(t)
	clientA := gocloak.Client{}
	clientB := gocloak.Client{}
	clientC := gocloak.Client{}
	type args struct {
		first  []*gocloak.Client
		second []*gocloak.Client
	}
	tests := []struct {
		name string
		args args
		want []*gocloak.Client
	}{
		{name: "only first", args: args{first: []*gocloak.Client{&clientA, &clientB}, second: nil}, want: []*gocloak.Client{&clientA, &clientB}},
		{name: "only second", args: args{first: nil, second: []*gocloak.Client{&clientC}}, want: []*gocloak.Client{&clientC}},
		{name: "both", args: args{first: []*gocloak.Client{&clientA, &clientB}, second: []*gocloak.Client{&clientC}}, want: []*gocloak.Client{&clientA, &clientB, &clientC}},
		{name: "empty but not nil", args: args{first: nil, second: nil}, want: []*gocloak.Client{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := concatClients(tt.args.first, tt.args.second)
			assertions.ElementsMatch(tt.want, r)
		})
	}
}
