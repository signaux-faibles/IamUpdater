package config

import (
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"

	"keycloakUpdater/v2/pkg/logger"
	"keycloakUpdater/v2/pkg/structs"
)

func InitConfig(configFilename string) (structs.Config, error) {
	return initConfig(configFilename, false)
}

func OverrideConfig(original structs.Config, overridingFilename string) structs.Config {
	if overridingFilename == "" {
		logger.Debug("pas de surcharge de configuration", nil)
		return original
	}
	logContext := logger.ContextForMethod(OverrideConfig)
	logger.Info("surcharge de configuration", logContext.AddString("filename", overridingFilename))
	overridingConfig, err := initConfig(overridingFilename, true)
	if err != nil {
		logger.Error(
			"erreur pendant la récupération de la surcharge de configuration",
			logContext.AddAny("filename", overridingFilename),
			err)
		return original
	}
	return merge(original, overridingConfig)
}

func initConfig(configFilename string, quietly bool) (structs.Config, error) {
	var conf structs.Config
	//var err error
	//var meta toml.MetaData
	filenames := getAllConfigFilenames(configFilename)

	logger.Info(
		"config files",
		logger.ContextForMethod(initConfig).AddArray("filenames", filenames),
	)
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
	logContext := logger.ContextForMethod(getAllConfigFilenames)
	// checking file exist
	var err error
	if _, err = os.Open(filename); err != nil {
		logger.Panic(
			"erreur pendant la lecture des du fichier de configuration",
			logContext.AddAny("filename", filename),
			err,
		)
	}
	r = append(r, filename)
	var files []os.DirEntry
	config := extractConfig(filename)
	if config.Stock == nil {
		return r
	}
	folder := config.Stock.ClientsAndRealmFolder
	if folder == "" {
		logger.Warn("Attention : aucun répertoire de configuration n'est défini", logContext)
		return r
	}
	stockFilename := config.Stock.UsersAndRolesFilename
	if stockFilename != "" {
		if _, err = os.ReadFile(stockFilename); err != nil {
			logger.Panic(
				"erreur pendant la lecture des du fichier stock",
				logContext.AddAny("filename", stockFilename),
				err,
			)
		}
	}
	if files, err = os.ReadDir(folder); err != nil {
		logger.Panic(
			"erreur pendant la lecture des clients Keycloak",
			logContext.AddAny("folder", folder),
			err,
		)
	}
	for _, f := range files {
		filename := folder + "/" + f.Name()
		if !strings.HasSuffix(filename, ".toml") {
			logger.Debug("ignore le fichier de configuration", logContext.AddAny("filename", filename))
			continue
		}
		r = append(r, filename)
	}
	return r
}

func extractConfig(filename string) structs.Config {
	logContext := logger.ContextForMethod(extractConfig)
	var conf structs.Config
	var err error
	var meta toml.MetaData
	if meta, err = toml.DecodeFile(filename, &conf); err != nil {
		logger.Panic(
			"error pendant le décodage du fichier de configuration Toml",
			logContext.AddAny("filename", filename),
			err,
		)
	}
	if meta.Undecoded() != nil {
		for _, key := range meta.Undecoded() {
			logger.Warn(
				"Attention : la clé du fichier de configuration n'est pas utilisée",
				logContext.AddAny("clé", key.String()).AddAny("filename", filename),
			)
		}
	}
	return conf
}
