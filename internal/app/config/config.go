package config

import "reflect"

type Config interface {
	ServerConfig | AppConfig
}

func initConfig[C Config](config *C, priorityConfigs []*C) {
	configValue := reflect.ValueOf(config).Elem()

	for _, priorityConfig := range priorityConfigs {
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
