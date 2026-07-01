package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	AppVersion        string
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
	// Root .env when running from backend/; local .env overrides.
	_ = godotenv.Load("../.env")
	_ = godotenv.Load()

	databaseURL, err := resolveDatabaseURL()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppVersion:        getEnv("APP_VERSION", "1.0.0"),
		HTTPPort:          getEnv("HTTP_PORT", "8080"),
		DatabaseURL:       databaseURL,
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

func resolveDatabaseURL() (string, error) {
	if explicit := os.Getenv("DATABASE_URL"); explicit != "" {
		return explicit, nil
	}

	user, err := requireEnv("POSTGRES_USER")
	if err != nil {
		return "", err
	}
	password, err := requireEnv("POSTGRES_PASSWORD")
	if err != nil {
		return "", err
	}
	dbName, err := requireEnv("POSTGRES_DB")
	if err != nil {
		return "", err
	}

	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")

	if password == "change-me-before-starting" {
		return "", fmt.Errorf("set POSTGRES_PASSWORD in .env before starting (copy from .env.example)")
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(user),
		url.QueryEscape(password),
		host,
		port,
		url.PathEscape(dbName),
	), nil
}

func requireEnv(key string) (string, error) {
	value := os.Getenv(key)
	if value == "" {
		return "", fmt.Errorf("missing required environment variable: %s", key)
	}
	return value, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
