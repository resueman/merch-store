package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPServer `yaml:"httpServer"`
	Postgres   `yaml:"postgres"`
	JWT        `yaml:"jwt"`
	Logger     `yaml:"logger"`
	TxManager  `yaml:"txManager"`
}

type HTTPServer struct {
	Port string `yaml:"port" env:"HTTP_SERVER_PORT" env-default:":8080"`
}

type Postgres struct {
	DSN string `yaml:"dsn" env:"DSN" env-required:"true"`
}

type JWT struct {
	Secret string `yaml:"secret" env:"JWT_SECRET" env-required:"true"`
	TTLMin int    `yaml:"ttlMin" env:"JWT_TTL_MINUTES" env-required:"true"`
}

type Logger struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"debug"`
	File  string `yaml:"logFile" env:"LOG_FILE" env-default:"logs/app.log"`
}

type TxManager struct {
	TimeoutMs  int `yaml:"timeoutMs" env:"TX_TIMEOUT_MS" env-required:"true"`
	MaxRetries int `yaml:"maxRetries" env:"TX_MAX_RETRIES" env-required:"true"`
}

//nolint:exhaustruct
func NewConfig(configPath string) (*Config, error) {
	config := &Config{}

	if err := cleanenv.ReadConfig(configPath, config); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	if err := cleanenv.UpdateEnv(config); err != nil {
		return nil, fmt.Errorf("error updating env: %w", err)
	}

	return config, nil
}
