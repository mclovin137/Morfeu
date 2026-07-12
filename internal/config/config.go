package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration loaded from environment variables
type Config struct {
	DatabaseURL string
	RedisURL    string
	LogLevel    string
	AppPort     string
	CacheTTL    time.Duration
	PoolMinSize int
	PoolMaxSize int
	PoolTimeout time.Duration
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/morfeu"),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		AppPort:     getEnv("APP_PORT", "8080"),
		PoolMinSize: getEnvInt("POOL_MIN_SIZE", 5),
		PoolMaxSize: getEnvInt("POOL_MAX_SIZE", 25),
	}

	// Parse cache TTL
	ttlSeconds := getEnvInt("CACHE_TTL_SECONDS", 300)
	cfg.CacheTTL = time.Duration(ttlSeconds) * time.Second

	// Parse pool timeout
	poolTimeoutSeconds := getEnvInt("POOL_TIMEOUT_SECONDS", 30)
	cfg.PoolTimeout = time.Duration(poolTimeoutSeconds) * time.Second

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.RedisURL == "" {
		return fmt.Errorf("REDIS_URL is required")
	}
	if c.AppPort == "" {
		return fmt.Errorf("APP_PORT is required")
	}
	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error")
	}
	if c.PoolMinSize < 1 {
		return fmt.Errorf("POOL_MIN_SIZE must be >= 1")
	}
	if c.PoolMaxSize < c.PoolMinSize {
		return fmt.Errorf("POOL_MAX_SIZE must be >= POOL_MIN_SIZE")
	}
	if c.CacheTTL < 1*time.Second {
		return fmt.Errorf("CACHE_TTL_SECONDS must be >= 1")
	}

	return nil
}

// getEnv returns environment variable value or default
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// getEnvInt returns environment variable as int or default
func getEnvInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultVal
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultVal
	}

	return value
}
