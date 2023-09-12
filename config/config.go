package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"github.com/signaux-faibles/keycloakUpdater/v2/structs"
)

func InitConfig(configFilename string) (structs.Config, error) {
	return initConfig(configFilename, false)
}

func OverrideConfig(original structs.Config, overridingFilename string) structs.Config {
	if overridingFilename == "" {
		slog.Debug("pas de surcharge de configuration")
		return original
	}
	slog.Info("surcharge de configuration", slog.String("filename", overridingFilename))
	overridingConfig, err := initConfig(overridingFilename, true)
	if err != nil {
		slog.Error(
			"erreur pendant la récupération de la surcharge de configuration",
			slog.String("filename", overridingFilename),
			slog.Any("error", err))
		return original
	}
	return merge(original, overridingConfig)
}

func initConfig(configFilename string, quietly bool) (structs.Config, error) {
	var conf structs.Config
	//var err error
	//var meta toml.MetaData
	filenames := getAllConfigFilenames(configFilename)
	slog.Info("config files", slog.Any("filenames", filenames))
	allConfig := readAllConfigFiles(filenames)
	for _, current := range allConfig {
		conf = merge(conf, current)
	}
	if !quietly && conf.Stock == nil {
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
		slog.Error(
			"error pendant la lecture des du fichier de configuration",
			slog.String("filename", filename),
			slog.Any("error", err),
		)
		panic(err)
	}
	r = append(r, filename)
	var files []os.DirEntry
	config := extractConfig(filename)
	if config.Stock == nil {
		return r
	}
	folder := config.Stock.ClientsAndRealmFolder
	if folder == "" {
		slog.Warn("no configuration folder is defined")
		return r
	}
	stockFilename := config.Stock.UsersAndRolesFilename
	if stockFilename != "" {
		if _, err = os.ReadFile(stockFilename); err != nil {
			slog.Error(
				"error pendant la lecture des du fichier stock",
				slog.String("filename", stockFilename),
				slog.Any("error", err),
			)
			panic(err)
		}
	}
	if files, err = os.ReadDir(folder); err != nil {
		slog.Error(
			"error pendant la lecture des clients Keycloak",
			slog.Any("folder", folder),
			slog.Any("error", err),
		)
		panic(err)
	}
	for _, f := range files {
		filename := folder + "/" + f.Name()
		if !strings.HasSuffix(filename, ".toml") {
			slog.Debug("ignore le fichier de configuration", slog.String("filename", filename))
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
		slog.Error(
			"error pendant le décodage du fichier de configuration Toml",
			slog.Any("filename", filename),
			slog.Any("error", err),
		)
		panic(err)
	}
	if meta.Undecoded() != nil {
		for _, key := range meta.Undecoded() {
			slog.Warn(
				"Attention : la clé du fichier de configuration n'est pas utilisée",
				slog.String("clé", key.String()),
				slog.String("filename", filename),
			)
		}
	}
	return conf
}
