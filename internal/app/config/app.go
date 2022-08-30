package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

const (
	hashSalt      = "_X@kQePA8dmBiZVBHax*zUUi"
	hashMinLength = 5
)

/*
AppConfig uses the following precedence order. Each item takes precedence over the item below it:
- Flag todo
- Env
- Config todo
- Default.
*/
type AppConfig struct {
	HashSalt      string `env:"APP_HASH_SALT"`
	HashMinLength int    `env:"APP_HASH_MIN_LENGTH"`
}

var AppCfg = GetAppConfig()

func NewAppConfig(hashSalt string, hashMinLength int) *AppConfig {
	return &AppConfig{
		HashSalt:      hashSalt,
		HashMinLength: hashMinLength,
	}
}

func GetAppConfig() *AppConfig {
	appCfg := NewAppConfig("", 0)

	defaultAppCfg := NewAppConfig(hashSalt, hashMinLength)

	envAppCfg := NewAppConfig("", 0)
	if err := env.Parse(envAppCfg); err != nil {
		log.Fatal(err)
	}

	priorityConfigs := []*AppConfig{envAppCfg, defaultAppCfg}

	initConfig(appCfg, priorityConfigs)

	return appCfg
}
