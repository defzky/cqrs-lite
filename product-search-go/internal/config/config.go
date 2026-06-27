package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	NatsURL     string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DB_URL", "postgres://postgres:postgres123@localhost:5432/product?sslmode=disable"),
		Port:        getEnv("SERVER_PORT", "8081"),
		NatsURL:     getEnv("NATS_URL", "nats://localhost:4222"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
