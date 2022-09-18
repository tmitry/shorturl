package configs

import (
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

const (
	dsn = ""
)

/*
DatabaseConfig uses the following precedence order. Each item takes precedence over the item below it:
- Flag
- Env
- YAML
- Default.
*/
type DatabaseConfig struct {
	DSN string `env:"DATABASE_DSN" yaml:"dsn"`
}

func NewDatabaseConfig(dsn string) *DatabaseConfig {
	return &DatabaseConfig{
		DSN: dsn,
	}
}

func NewDefaultDatabaseConfig() *DatabaseConfig {
	return NewDatabaseConfig(dsn)
}

func GetDatabaseConfig(flagConfig *FlagConfig) *DatabaseConfig {
	databaseCfg := NewDatabaseConfig("")

	defaultDatabaseCfg := NewDefaultDatabaseConfig()

	envDatabaseCfg := NewDatabaseConfig("")
	if err := env.Parse(envDatabaseCfg); err != nil {
		log.Panic(err)
	}

	flagDatabaseCfg := NewDatabaseConfig(flagConfig.DatabaseDSN)

	yamlDatabaseCfg := NewDatabaseConfig("")

	if flagConfig.DatabaseConfigPath != "" {
		file, err := os.Open(flagConfig.DatabaseConfigPath)
		if err != nil {
			log.Panic(err)
		}

		decoder := yaml.NewDecoder(file)

		if err = decoder.Decode(yamlDatabaseCfg); err != nil {
			log.Panic(err)
		}
	}

	priorityConfigs := []*DatabaseConfig{flagDatabaseCfg, envDatabaseCfg, yamlDatabaseCfg, defaultDatabaseCfg}

	buildConfig(databaseCfg, priorityConfigs)

	return databaseCfg
}
