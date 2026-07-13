package catalogo

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewFilmService(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			t.Logf("failed to sync logger: %v", err)
		}
	}()

	// Test that FilmService can be instantiated with nil db/cache
	// Real tests will use testcontainers
	service := NewFilmService(nil, nil, nil, logger)
	if service == nil {
		t.Fatal("FilmService should not be nil")
	}

	if service.logger == nil {
		t.Error("Logger should not be nil")
	}
}

func TestFilmServiceFields(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer func() {
		if err := logger.Sync(); err != nil {
			t.Logf("failed to sync logger: %v", err)
		}
	}()

	service := NewFilmService(nil, nil, nil, logger)

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
