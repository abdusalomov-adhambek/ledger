package config

import (
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func Load() (*Config, error) {
	var cfg Config

	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
