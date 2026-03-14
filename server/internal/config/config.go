package config

import "os"

type Config struct {
	DatabaseURL string
	Port        string
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://bullshit:bullshit@localhost:5432/bullshit?sslmode=disable"),
		Port:        getEnv("PORT", "8080"),
	}
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
