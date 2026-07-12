package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear environment
	os.Clearenv()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.AppPort != "8080" {
		t.Errorf("Expected AppPort 8080, got %s", cfg.AppPort)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected LogLevel info, got %s", cfg.LogLevel)
	}

	if cfg.CacheTTL != 5*time.Minute {
		t.Errorf("Expected CacheTTL 5m, got %v", cfg.CacheTTL)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	// Clear and set environment
	os.Clearenv()
	if err := os.Setenv("APP_PORT", "9000"); err != nil {
		t.Fatalf("failed to set APP_PORT: %v", err)
	}
	if err := os.Setenv("LOG_LEVEL", "debug"); err != nil {
		t.Fatalf("failed to set LOG_LEVEL: %v", err)
	}
	if err := os.Setenv("CACHE_TTL_SECONDS", "600"); err != nil {
		t.Fatalf("failed to set CACHE_TTL_SECONDS: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.AppPort != "9000" {
		t.Errorf("Expected AppPort 9000, got %s", cfg.AppPort)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel debug, got %s", cfg.LogLevel)
	}

	if cfg.CacheTTL != 10*time.Minute {
		t.Errorf("Expected CacheTTL 10m, got %v", cfg.CacheTTL)
	}
}

func TestValidate_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://localhost",
		RedisURL:    "redis://localhost",
		AppPort:     "8080",
		LogLevel:    "invalid",
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid LogLevel")
	}
}

func TestValidate_InvalidPoolSizes(t *testing.T) {
	cfg := &Config{
		DatabaseURL: "postgres://localhost",
		RedisURL:    "redis://localhost",
		AppPort:     "8080",
		LogLevel:    "info",
		PoolMinSize: 10,
		PoolMaxSize: 5, // Invalid: max < min
	}

	err := cfg.Validate()
	if err == nil {
		t.Error("Expected validation error for PoolMaxSize < PoolMinSize")
	}
}
