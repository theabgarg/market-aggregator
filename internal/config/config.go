package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port string
	TargetSymbol string
}

func Load() (*AppConfig, error) {
	_ = godotenv.Load();

	cfg := &AppConfig{
		Port: os.Getenv("PORT"),
		TargetSymbol: os.Getenv("TARGET_SYMBOL"),
	}

	if cfg.Port == ""{
		cfg.Port = "8080"
	}
	if cfg.TargetSymbol == "" {
		return nil, fmt.Errorf("TARGET_SYMBOL is required")
	}

	return cfg, nil
}