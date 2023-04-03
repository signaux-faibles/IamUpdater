package config

import (
	"fmt"
	"github.com/Nerzal/gocloak/v13"
	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

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

func Test_merge(t *testing.T) {
	assertions := assert.New(t)
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
			assertions.EqualValues(tt.want, got)
		})
	}
}

func Test_mergeKeycloak(t *testing.T) {
	anAccess := structs.Keycloak{}
	anotherAccess := structs.Keycloak{}
	type args struct {
		first  *structs.Keycloak
		second *structs.Keycloak
	}
	tests := []struct {
		name string
		args args
		want *structs.Keycloak
	}{
		{name: "first is chosen", args: args{first: &anAccess, second: nil}, want: &anAccess},
		{name: "second is chosen", args: args{first: nil, second: &anAccess}, want: &anAccess},
		{name: "first is chosen", args: args{first: &anAccess, second: &anotherAccess}, want: &anAccess},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeAccess(tt.args.first, tt.args.second); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeRealm(t *testing.T) {
	aRealm := gocloak.RealmRepresentation{}
	anotherRealm := gocloak.RealmRepresentation{}
	type args struct {
		first  *gocloak.RealmRepresentation
		second *gocloak.RealmRepresentation
	}
	tests := []struct {
		name string
		args args
		want *gocloak.RealmRepresentation
	}{
		{name: "first is chosen", args: args{first: &aRealm, second: nil}, want: &aRealm},
		{name: "second is chosen", args: args{first: nil, second: &aRealm}, want: &aRealm},
		{name: "first is chosen", args: args{first: &aRealm, second: &anotherRealm}, want: &aRealm},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeRealm(tt.args.first, tt.args.second); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeRealm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeClients(t *testing.T) {
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
			r := mergeClients(tt.args.first, tt.args.second)
			assertions.ElementsMatch(tt.want, r)
		})
	}
}
