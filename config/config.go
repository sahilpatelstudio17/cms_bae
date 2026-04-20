package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port            string
	Environment     string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiresHours int
}

func LoadConfig() AppConfig {
	_ = godotenv.Load()

	expiresHours := 24
	if raw := os.Getenv("JWT_EXPIRES_HOURS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err == nil && parsed > 0 {
			expiresHours = parsed
		}
	}

	cfg := AppConfig{
		Port:            getEnv("PORT", getEnv("APP_PORT", "8080")),
		Environment:     getEnv("APP_ENV", "development"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       getEnv("JWT_SECRET", "change-me"),
		JWTExpiresHours: expiresHours,
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
