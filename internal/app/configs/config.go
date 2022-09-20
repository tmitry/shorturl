package configs

import (
	"reflect"

	flag "github.com/spf13/pflag"
)

type ConfigInterface interface {
	ServerConfig | AppConfig | DatabaseConfig
}

type Config struct {
	App      *AppConfig
	Server   *ServerConfig
	Database *DatabaseConfig
}

func NewConfig() *Config {
	flagConfig := NewFlagConfig()

	return &Config{
		App:      GetAppConfig(flagConfig),
		Server:   GetServerConfig(flagConfig),
		Database: GetDatabaseConfig(flagConfig),
	}
}

func NewDefaultConfig() *Config {
	return &Config{
		App:      NewDefaultAppConfig(),
		Server:   NewDefaultServerConfig(),
		Database: NewDefaultDatabaseConfig(),
	}
}

func buildConfig[C ConfigInterface](config *C, precedenceConfigs []*C) {
	configValue := reflect.ValueOf(config).Elem()

	for _, priorityConfig := range precedenceConfigs {
		priorityConfigValue := reflect.ValueOf(priorityConfig).Elem()

		for index := 0; index < configValue.NumField(); index++ {
			if reflect.Zero(configValue.Field(index).Type()).Interface() != configValue.Field(index).Interface() {
				continue
			}

			if priorityConfigValue.Field(index).Interface() != reflect.Zero(configValue.Field(index).Type()).Interface() {
				configValue.Field(index).Set(priorityConfigValue.Field(index))
			}
		}
	}
}

type FlagConfig struct {
	Address            string
	BaseURL            string
	FileStoragePath    string
	ServerConfigPath   string
	AppConfigPath      string
	DatabaseConfigPath string
	JWTSignatureKey    string
	DatabaseDSN        string
}

func NewFlagConfig() *FlagConfig {
	flagConfig := &FlagConfig{
		Address:            "",
		BaseURL:            "",
		FileStoragePath:    "",
		ServerConfigPath:   "",
		AppConfigPath:      "",
		DatabaseConfigPath: "",
		JWTSignatureKey:    "",
		DatabaseDSN:        "",
	}

	flag.StringVarP(&flagConfig.Address, "server_address", "a", "", "Server address")
	flag.StringVarP(&flagConfig.BaseURL, "base_url", "b", "", "Base URL")
	flag.StringVarP(&flagConfig.FileStoragePath, "file_path", "f", "", "File storage path")
	flag.StringVar(&flagConfig.ServerConfigPath, "server_config_path", "", "Server config path")
	flag.StringVar(&flagConfig.AppConfigPath, "app_config_path", "", "App config path")
	flag.StringVar(&flagConfig.DatabaseConfigPath, "database_config_path", "", "Database config path")
	flag.StringVar(&flagConfig.JWTSignatureKey, "jwt_signature_key", "", "JWT Signature key")
	flag.StringVarP(&flagConfig.DatabaseDSN, "database_dsn", "d", "", "Database DSN")
	flag.Parse()

	return flagConfig
}
