package configs

import (
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

const (
	address           = "localhost:8080"
	baseURL           = "http://localhost:8080"
	readHeaderTimeout = 2
	compressionLevel  = 5
	jwtSignatureKey   = "sRhs-tWB!Kq7RLCHYek6QFks"
)

/*
ServerConfig uses the following precedence order. Each item takes precedence over the item below it:
- Flag
- Env
- YAML
- Default.
*/
type ServerConfig struct {
	Address           string `env:"SERVER_ADDRESS" yaml:"address"`
	BaseURL           string `env:"BASE_URL" yaml:"base_url"`
	ReadHeaderTimeout int    `env:"SERVER_READ_HEADER_TIMEOUT" yaml:"read_header_timeout"`
	CompressionLevel  int    `env:"SERVER_COMPRESSION_LEVEL" yaml:"compression_level"`
	JWTSignatureKey   string `env:"JWT_SIGNATURE_KEY" yaml:"jwt_signature_key"`
}

func NewServerConfig(address, baseURL, jwtSignatureKey string, readHeaderTimeout, compressionLevel int) *ServerConfig {
	return &ServerConfig{
		Address:           address,
		BaseURL:           baseURL,
		ReadHeaderTimeout: readHeaderTimeout,
		CompressionLevel:  compressionLevel,
		JWTSignatureKey:   jwtSignatureKey,
	}
}

func NewDefaultServerConfig() *ServerConfig {
	return NewServerConfig(address, baseURL, jwtSignatureKey, readHeaderTimeout, compressionLevel)
}

func GetServerConfig(flagConfig *FlagConfig) *ServerConfig {
	serverCfg := NewServerConfig("", "", "", 0, 0)

	defaultServerCfg := NewDefaultServerConfig()

	envServerCfg := NewServerConfig("", "", "", 0, 0)
	if err := env.Parse(envServerCfg); err != nil {
		log.Panic(err)
	}

	yamlServerCfg := NewServerConfig("", "", "", 0, 0)

	if flagConfig.ServerConfigPath != "" {
		file, err := os.Open(flagConfig.ServerConfigPath)
		if err != nil {
			log.Panic(err)
		}

		decoder := yaml.NewDecoder(file)

		if err = decoder.Decode(yamlServerCfg); err != nil {
			log.Panic(err)
		}
	}

	flagServerCfg := NewServerConfig(flagConfig.Address, flagConfig.BaseURL, flagConfig.JWTSignatureKey, 0, 0)

	priorityConfigs := []*ServerConfig{flagServerCfg, envServerCfg, yamlServerCfg, defaultServerCfg}

	buildConfig(serverCfg, priorityConfigs)

	return serverCfg
}
