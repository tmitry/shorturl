package configs

import (
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

const (
	hashSalt                   = "_X@kQePA8dmBiZVBHax*zUUi"
	hashMinLength              = 5
	fileStoragePath            = ""
	deletionBufferMaxSize      = 500
	deletionBufferClearTimeout = 5 // Idle time (in seconds) after which buffer will be cleared.
)

/*
AppConfig uses the following precedence order. Each item takes precedence over the item below it:
- Flag
- Env
- YAML
- Default.
*/
type AppConfig struct {
	HashSalt                   string `env:"APP_HASH_SALT" yaml:"hash_salt"`
	HashMinLength              int    `env:"APP_HASH_MIN_LENGTH" yaml:"hash_min_length"`
	FileStoragePath            string `env:"FILE_STORAGE_PATH" yaml:"file_storage_path"`
	DeletionBufferMaxSize      int    `env:"APP_DELETION_BUFFER_MAX_SIZE" yaml:"deletion_buffer_max_size"`
	DeletionBufferClearTimeout int    `env:"APP_DELETION_CLEAR_TIMEOUT" yaml:"deletion_buffer_clear_timeout"`
}

func NewAppConfig(
	hashSalt string,
	hashMinLength int,
	fileStoragePath string,
	deletionBufferMaxSize int,
	deletionBufferClearTimeout int,
) *AppConfig {
	return &AppConfig{
		HashSalt:                   hashSalt,
		HashMinLength:              hashMinLength,
		FileStoragePath:            fileStoragePath,
		DeletionBufferMaxSize:      deletionBufferMaxSize,
		DeletionBufferClearTimeout: deletionBufferClearTimeout,
	}
}

func NewDefaultAppConfig() *AppConfig {
	return NewAppConfig(hashSalt, hashMinLength, fileStoragePath, deletionBufferMaxSize, deletionBufferClearTimeout)
}

func GetAppConfig(flagConfig *FlagConfig) *AppConfig {
	appCfg := NewAppConfig("", 0, "", 0, 0)

	defaultAppCfg := NewDefaultAppConfig()

	envAppCfg := NewAppConfig("", 0, "", 0, 0)
	if err := env.Parse(envAppCfg); err != nil {
		log.Panic(err)
	}

	flagAppCfg := NewAppConfig("", 0, flagConfig.FileStoragePath, 0, 0)

	yamlAppCfg := NewAppConfig("", 0, "", 0, 0)

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
