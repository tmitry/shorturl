package configs

import (
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

const (
	hashSalt        = "_X@kQePA8dmBiZVBHax*zUUi"
	hashMinLength   = 5
	fileStoragePath = ""
	jwtSignatureKey = "sRhs-tWB!Kq7RLCHYek6QFks"
)

/*
AppConfig uses the following precedence order. Each item takes precedence over the item below it:
- Flag
- Env
- YAML
- Default.
*/
type AppConfig struct {
	HashSalt        string `env:"APP_HASH_SALT" yaml:"hash_salt"`
	HashMinLength   int    `env:"APP_HASH_MIN_LENGTH" yaml:"hash_min_length"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" yaml:"file_storage_path"`
	JWTSignatureKey string `env:"JWT_SIGNATURE_KEY" yaml:"jwt_signature_key"`
}

var AppCfg = NewAppConfig(hashSalt, hashMinLength, fileStoragePath, jwtSignatureKey)

func NewAppConfig(hashSalt string, hashMinLength int, fileStoragePath, jwtSignatureKey string) *AppConfig {
	return &AppConfig{
		HashSalt:        hashSalt,
		HashMinLength:   hashMinLength,
		FileStoragePath: fileStoragePath,
		JWTSignatureKey: jwtSignatureKey,
	}
}

func GetAppConfig(flagConfig *FlagConfig) *AppConfig {
	appCfg := NewAppConfig("", 0, "", "")

	defaultAppCfg := AppCfg

	envAppCfg := NewAppConfig("", 0, "", "")
	if err := env.Parse(envAppCfg); err != nil {
		log.Panic(err)
	}

	flagAppCfg := NewAppConfig("", 0, flagConfig.FileStoragePath, flagConfig.JWTSignatureKey)

	yamlAppCfg := NewAppConfig("", 0, "", "")

	if flagConfig.AppConfigPath != "" {
		file, err := os.Open(flagConfig.AppConfigPath)
		if err != nil {
			log.Panic(err)
		}

		decoder := yaml.NewDecoder(file)

		if err = decoder.Decode(yamlAppCfg); err != nil {
			log.Panic(err)
		}
	}

	priorityConfigs := []*AppConfig{flagAppCfg, envAppCfg, yamlAppCfg, defaultAppCfg}

	buildConfig(appCfg, priorityConfigs)

	return appCfg
}
