package configs

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var cfg Config

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DB       string
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		p.User, p.Password, p.Host, p.Port, p.DB)
}

type RedisConfig struct {
	Host string
	Port string
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type Config struct {
	Port     string
	Postgres PostgresConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("failed to load .env: %v", err))
	}

	cfg = Config{
		Port: mustEnv("PORT"),
		Postgres: PostgresConfig{
			Host:     mustEnv("POSTGRES_HOST"),
			Port:     mustEnv("POSTGRES_PORT"),
			User:     mustEnv("POSTGRES_USER"),
			Password: mustEnv("POSTGRES_PASSWORD"),
			DB:       mustEnv("POSTGRES_DB"),
		},
		Redis: RedisConfig{
			Host: mustEnv("REDIS_HOST"),
			Port: mustEnv("REDIS_PORT"),
		},
		JWT: JWTConfig{
			AccessSecret:  mustEnv("JWT_ACCESS_SECRET"),
			RefreshSecret: mustEnv("JWT_REFRESH_SECRET"),
			AccessTTL:     mustDuration("JWT_ACCESS_TTL"),
			RefreshTTL:    mustDuration("JWT_REFRESH_TTL"),
		},
	}

	return cfg
}

func Get() Config {
	return cfg
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("missing required env: %s", key))
	}
	return val
}

func mustDuration(key string) time.Duration {
	val := mustEnv(key)
	d, err := time.ParseDuration(val)
	if err != nil {
		panic(fmt.Sprintf("invalid duration for %s: %v", key, err))
	}
	return d
}
