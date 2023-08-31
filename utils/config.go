package utils

import (
	"fmt"

	"github.com/caarlos0/env/v7"
	"github.com/joho/godotenv"
)

type EnvConfig struct {
	Host       string `env:"RABBITMQ_DB_HOST,required"`
	User       string `env:"RABBITMQ_DB_USER,required"`
	Port       int    `env:"RABBITMQ_DB_PORT,required"`
	ServerPort int    `env:"PORT,required"`
	Password   string `env:"RABBITMQ_DB_PASSWORD,required"`
}

func LoadConnectionConfig() (*EnvConfig, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("unable to load .env file: %e", err)
	}

	config := EnvConfig{}

	err = env.Parse(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse environment variables: %e", err)
	}

	return &config, nil
}
