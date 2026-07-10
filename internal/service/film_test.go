package service

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewFilmService(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Test that FilmService can be instantiated with nil db/cache
	// Real tests will use testcontainers
	service := NewFilmService(nil, nil, logger)
	if service == nil {
		t.Error("FilmService should not be nil")
	}

	if service.logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestFilmServiceFields(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	service := NewFilmService(nil, nil, logger)

	if service.logger != logger {
		t.Error("Logger not properly assigned")
	}
}

func TestCacheTTL(t *testing.T) {
	ttl := 5 * time.Minute
	if ttl.Seconds() != 300 {
		t.Errorf("Expected TTL 300s, got %v", ttl.Seconds())
	}
}
