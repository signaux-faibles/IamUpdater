package config

import (
  "github.com/BurntSushi/toml"
  "github.com/pkg/errors"
  "github.com/signaux-faibles/keycloakUpdater/v2/logger"
  "github.com/signaux-faibles/keycloakUpdater/v2/structs"
  "os"
  "strings"
)

func InitConfig(configFilename string) (structs.Config, error) {
  return initConfig(configFilename, false)
}

func OverrideConfig(original structs.Config, overridingFilename string) structs.Config {
  if overridingFilename == "" {
    logger.Infof("pas de surcharge de configuration")
    return original
  }
  logger.Infof("surcharge de configuration : %s", overridingFilename)
  overridingConfig, err := initConfig(overridingFilename, true)
  if err != nil {
    logger.Errorf(
      "erreur pendant la récupération de la surcharge de configuration '%s' : %s",
      overridingFilename,
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
  logger.Infof("config files : %s", filenames)
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
    logger.Panicf("error reading clients config file : %s", err)
  }
  r = append(r, filename)
  var files []os.DirEntry
  config := extractConfig(filename)
  if config.Stock == nil {
    return r
  }
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
