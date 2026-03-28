package configs

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppEnv            string
	AppPort           string
	PostgresDSN       string
	RedisAddr         string
	RedisPassword     string
	RedisDB           int
	JWTAccessSecret   string
	JWTRefreshSecret  string
	JWTAccessTTLMin   int
	JWTRefreshTTLHour int
}

func Load() (Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid REDIS_DB: %w", err)
	}
	accessTTL, err := strconv.Atoi(getEnv("JWT_ACCESS_TTL_MINUTES", "15"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_ACCESS_TTL_MINUTES: %w", err)
	}
	refreshTTL, err := strconv.Atoi(getEnv("JWT_REFRESH_TTL_HOURS", "168"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid JWT_REFRESH_TTL_HOURS: %w", err)
	}

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		getEnv("POSTGRES_USER", "app_user"),
		getEnv("POSTGRES_PASSWORD", "app_pass"),
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_DB", "app_db"),
		getEnv("POSTGRES_SSLMODE", "disable"),
	)

	cfg := Config{
		AppEnv:            getEnv("APP_ENV", "local"),
		AppPort:           getEnv("APP_PORT", "8080"),
		PostgresDSN:       dsn,
		RedisAddr:         getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:     getEnv("REDIS_PASSWORD", ""),
		RedisDB:           redisDB,
		JWTAccessSecret:   getEnv("JWT_ACCESS_SECRET", "change_me_access_secret"),
		JWTRefreshSecret:  getEnv("JWT_REFRESH_SECRET", "change_me_refresh_secret"),
		JWTAccessTTLMin:   accessTTL,
		JWTRefreshTTLHour: refreshTTL,
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
