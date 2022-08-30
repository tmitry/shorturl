package config

import (
	"log"

	"github.com/caarlos0/env/v6"
)

const (
	address           = "localhost:8080"
	baseURL           = "http://localhost:8080"
	readHeaderTimeout = 2
)

/*
ServerConfig uses the following precedence order. Each item takes precedence over the item below it:
- Flag todo
- Env
- Config todo
- Default.
*/
type ServerConfig struct {
	Address           string `env:"SERVER_ADDRESS"`
	BaseURL           string `env:"BASE_URL"`
	ReadHeaderTimeout int    `env:"SERVER_READ_HEADER_TIMEOUT"`
}

var ServerCfg = GetServerConfig()

func NewServerConfig(address, baseURL string, readHeaderTimeout int) *ServerConfig {
	return &ServerConfig{
		Address:           address,
		BaseURL:           baseURL,
		ReadHeaderTimeout: readHeaderTimeout,
	}
}

func GetServerConfig() *ServerConfig {
	serverCfg := NewServerConfig("", "", 0)

	defaultServerCfg := NewServerConfig(address, baseURL, readHeaderTimeout)

	envServerCfg := NewServerConfig("", "", 0)
	if err := env.Parse(envServerCfg); err != nil {
		log.Fatal(err)
	}

	priorityConfigs := []*ServerConfig{envServerCfg, defaultServerCfg}

	initConfig(serverCfg, priorityConfigs)

	return serverCfg
}
