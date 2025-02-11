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
}

type HTTPServer struct {
	Port string `yaml:"port" env:"HTTP_SERVER_PORT" env-default:"8080"`
}

type Postgres struct {
	DSN        string `yaml:"dsn" env:"POSTGRES_DSN" env-default:"postgres://postgres:postgres@localhost:5432/postgres"`
	Username   string `yaml:"username" env:"POSTGRES_USER" env-default:"postgres"`
	Password   string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Host       string `yaml:"host" env:"HOST" env-default:"localhost"`
	Port       int    `yaml:"port" env:"PORT" env-default:"5432"`
	DB         string `yaml:"db" env:"POSTGRES_DB" env-default:"postgres"`
	SSLMode    string `yaml:"sslMode" env:"SSL_MODE" env-default:"disable"`
	TimeoutSec int    `yaml:"timeoutSec" env:"TIMEOUT_SEC" env-default:"3"`
	MaxRetries int    `yaml:"maxRetries" env:"MAX_RETRIES" env-default:"3"`
}

type JWT struct {
	Secret           string `yaml:"secret"`
	AccessTTLMinutes int    `yaml:"jwtAccessTtl" env:"JWT_ACCESS_TTL_MINUTES" env-default:"15"`
	RefreshTTLDays   int    `yaml:"jwtRefreshTtl" env:"JWT_REFRESH_TTL_DAYS" env-default:"30"`
}

type Logger struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"debug"`
	File  string `yaml:"logFile" env:"LOG_FILE" env-default:"logs/app.log"`
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
