package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	HTTPPort          string
	DatabaseURL       string
	MigrationsPath    string
	JWTPrivateKeyPath string
	JWTPublicKeyPath  string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	CORSOrigins       string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://optistock:optistock@localhost:5432/optistock?sslmode=disable"),
		MigrationsPath:    getEnv("MIGRATIONS_PATH", "migrations"),
		JWTPrivateKeyPath: getEnv("JWT_PRIVATE_KEY_PATH", "keys/private.pem"),
		JWTPublicKeyPath:  getEnv("JWT_PUBLIC_KEY_PATH", "keys/public.pem"),
		CORSOrigins:       getEnv("CORS_ORIGINS", "http://localhost:5173"),
	}

	accessMinutes, err := strconv.Atoi(getEnv("ACCESS_TOKEN_TTL_MINUTES", "15"))
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL_MINUTES: %w", err)
	}
	refreshDays, err := strconv.Atoi(getEnv("REFRESH_TOKEN_TTL_DAYS", "7"))
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL_DAYS: %w", err)
	}

	cfg.AccessTokenTTL = time.Duration(accessMinutes) * time.Minute
	cfg.RefreshTokenTTL = time.Duration(refreshDays) * 24 * time.Hour

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
