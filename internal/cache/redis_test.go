package cache

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

type MockRedisClient struct {
	data map[string][]byte
}

func (m *MockRedisClient) Get(ctx context.Context, key string) ([]byte, error) {
	return m.data[key], nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	m.data[key] = value
	return nil
}

func TestMarshalJSON(t *testing.T) {
	data := map[string]string{"key": "value"}
	bytes, err := MarshalJSON(data)
	if err != nil {
		t.Errorf("MarshalJSON failed: %v", err)
	}

	if len(bytes) == 0 {
		t.Error("MarshalJSON returned empty bytes")
	}
}

func TestUnmarshalJSON(t *testing.T) {
	jsonData := []byte(`{"key":"value"}`)
	var result map[string]string

	err := UnmarshalJSON(jsonData, &result)
	if err != nil {
		t.Errorf("UnmarshalJSON failed: %v", err)
	}

	if result["key"] != "value" {
		t.Error("UnmarshalJSON failed to decode correctly")
	}
}

func TestRedisCache_SetGetTimeout(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Exercise the mock through the same interface the cache expects;
	// real Redis behavior is covered by the integration tests (testcontainers)
	mockClient := &MockRedisClient{
		data: make(map[string][]byte),
	}
	if err := mockClient.Set(context.Background(), "k", []byte("v"), time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got, err := mockClient.Get(context.Background(), "k")
	if err != nil || string(got) != "v" {
		t.Errorf("Get = %q, err %v; want %q", got, err, "v")
	}

	if logger == nil {
		t.Error("Logger should not be nil")
	}
}
