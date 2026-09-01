package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration.
type Config struct {
	Env      string
	Port     string
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	OTP      OTPConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type OTPConfig struct {
	ExpiryMinutes int
	DevMode       bool // When true, OTP "111111" always works (dev only)
}

// Load reads config from environment variables.
// A .env file in the working directory is loaded automatically if present.
func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error — .env is optional in production

	return &Config{
		Env:  env("APP_ENV", "development"),
		Port: env("APP_PORT", "8080"),
		Database: DatabaseConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5432"),
			User:     env("DB_USER", "cp_user"),
			Password: env("DB_PASSWORD", "cp_password"),
			Name:     env("DB_NAME", "community_platform"),
			SSLMode:  env("DB_SSL_MODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     env("REDIS_HOST", "localhost"),
			Port:     env("REDIS_PORT", "6379"),
			Password: env("REDIS_PASSWORD", ""),
		},
		JWT: JWTConfig{
			AccessSecret:  env("JWT_ACCESS_SECRET", "dev-access-secret-change-in-prod"),
			RefreshSecret: env("JWT_REFRESH_SECRET", "dev-refresh-secret-change-in-prod"),
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
		OTP: OTPConfig{
			ExpiryMinutes: 5,
			DevMode:       env("OTP_DEV_MODE", "true") == "true",
		},
	}, nil
}

// IsDevelopment returns true when running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// env returns the value of an environment variable or a fallback default.
func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
