package configs

import (
	"reflect"

	flag "github.com/spf13/pflag"
)

type Config interface {
	ServerConfig | AppConfig
}

func InitConfigs() {
	flagConfig := NewFlagConfig()

	AppCfg = GetAppConfig(flagConfig)
	ServerCfg = GetServerConfig(flagConfig)
}

func buildConfig[C Config](config *C, precedenceConfigs []*C) {
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
	Address          string
	BaseURL          string
	FileStoragePath  string
	ServerConfigPath string
	AppConfigPath    string
	JWTSignatureKey  string
}

func NewFlagConfig() *FlagConfig {
	flagConfig := &FlagConfig{
		Address:          "",
		BaseURL:          "",
		FileStoragePath:  "",
		ServerConfigPath: "",
		AppConfigPath:    "",
		JWTSignatureKey:  "",
	}

	flag.StringVarP(&flagConfig.Address, "server_address", "a", "", "Server address")
	flag.StringVarP(&flagConfig.BaseURL, "base_url", "b", "", "Base URL")
	flag.StringVarP(&flagConfig.FileStoragePath, "file_path", "f", "", "File storage path")
	flag.StringVar(&flagConfig.ServerConfigPath, "server_config_path", "", "Server config path")
	flag.StringVar(&flagConfig.AppConfigPath, "app_config_path", "", "App config path")
	flag.StringVar(&flagConfig.JWTSignatureKey, "jwt_signature_key", "", "JWT Signature key")
	flag.Parse()

	return flagConfig
}
