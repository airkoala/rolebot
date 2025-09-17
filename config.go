package main

import (
  "errors"
  "log"
  "os"
  "strings"

  "github.com/pelletier/go-toml/v2"
)

// [[groups]]
// heading = "Other Group"
// description = "lipsum lipsum ..."
// multiple = false
// roles = [
//   1416011758459682887,
//   1416011879276613674,
// ]

type RoleGroup struct {
  Id          int
  Heading     string
  Description string
  Multiple    bool
  Roles       []string
}

type GuildConfig struct {
  TargetChannelId string
  RoleGroups      []RoleGroup
}

type Config struct {
  token  string
  appId  string
  guilds map[string]GuildConfig
}

func readConfig(configDir string) (Config, error) {
  config := Config{
	guilds: make(map[string]GuildConfig),
  }

  token, err := os.ReadFile(configDir + "/token")
  if err != nil {
	return Config{}, err
  }
  config.token = strings.TrimSpace(string(token))

  appId, err := os.ReadFile(configDir + "/appid")
  if err != nil {
	return Config{}, err
  }
  config.appId = strings.TrimSpace(string(appId))

  guildConfigDirPath := configDir + "/guilds/"
  guildConfigDir, err := os.Open(guildConfigDirPath)
  if err != nil {
	log.Printf("No guild configs found in %v.\n", guildConfigDirPath)
  }

  files, err := guildConfigDir.ReadDir(-1)
  if err != nil {
	log.Printf("Error reading contents of directory %v\n", guildConfigDirPath)
  }

  for _, e := range files {
	e.Name()
	if e.IsDir() || !strings.HasSuffix(e.Name(), ".toml") {
	  continue
	}

	f, err := os.Open(guildConfigDirPath + e.Name())

	if err != nil {
	  log.Printf("Error opening guild config file %v: %v\n", e.Name(), err)
	  continue
	}

	var guildConfig GuildConfig
	// err = toml.NewDecoder(f).Strict(true).Decode(&guildConfig)
	err = toml.NewDecoder(f).DisallowUnknownFields().Decode(&guildConfig)
	if err != nil {
	  log.Printf("Error reading guild config file %v: %v\n", e.Name(), err)
	  var details *toml.StrictMissingError
	  if !errors.As(err, &details) {
		log.Panicf("err should have been a *toml.StrictMissingError, but got %s (%T)", err, err)
	  }

	  log.Println("\n" + details.String())
	}

	for i := range guildConfig.RoleGroups {
	  guildConfig.RoleGroups[i].Id = i
	}

	config.guilds[e.Name()[:len(e.Name())-5]] = guildConfig
  }

  return config, nil
}
