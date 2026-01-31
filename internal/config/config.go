package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
// Single responsibility: centralize environment variable loading.
type Config struct {
	Server ServerConfig
	Auth   AuthConfig
	Rate   RateLimitConfig
	CORS   CORSConfig
}

type ServerConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type AuthConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
}

type RateLimitConfig struct {
	RequestsPerMinute int
	Enabled           bool
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// Load reads configuration from environment variables.
// Explicit defaults for every value.
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnv("PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		Auth: AuthConfig{
			AccessTokenSecret:  getEnv("ACCESS_TOKEN_SECRET", ""),
			RefreshTokenSecret: getEnv("REFRESH_TOKEN_SECRET", ""),
			AccessTokenTTL:     getDurationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
			RefreshTokenTTL:    getDurationEnv("REFRESH_TOKEN_TTL", 7*24*time.Hour),
		},
		Rate: RateLimitConfig{
			RequestsPerMinute: getIntEnv("RATE_LIMIT_RPM", 60),
			Enabled:           getBoolEnv("RATE_LIMIT_ENABLED", true),
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		},
	}

	// Validate critical configuration
	if cfg.Auth.AccessTokenSecret == "" {
		return nil, fmt.Errorf("ACCESS_TOKEN_SECRET is required")
	}
	if cfg.Auth.RefreshTokenSecret == "" {
		return nil, fmt.Errorf("REFRESH_TOKEN_SECRET is required")
	}

	return cfg, nil
}

// Helper functions for environment variable parsing
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
