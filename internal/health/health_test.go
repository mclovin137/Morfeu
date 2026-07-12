package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// setupTestDBForHealth creates an ephemeral PostgreSQL test container
func setupTestDBForHealth(ctx context.Context, t *testing.T) (*pgxpool.Pool, testcontainers.Container) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "morfeu_test",
		},
		// O log aparece duas vezes (initdb + processo final) — esperar a 2ª ocorrência
		// evita conectar durante o restart interno do initdb.
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(120 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start DB container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate db container: %v", termErr)
		}
		t.Fatalf("Failed to get DB host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate db container: %v", termErr)
		}
		t.Fatalf("Failed to get DB port: %v", err)
	}

	dsn := "postgres://postgres:postgres@" + host + ":" + port.Port() + "/morfeu_test"
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate db container: %v", termErr)
		}
		t.Fatalf("Failed to create pool: %v", err)
	}

	return pool, container
}

// setupTestRedisForHealth creates an ephemeral Redis test container
func setupTestRedisForHealth(ctx context.Context, t *testing.T) (*redis.Client, testcontainers.Container) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(90 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate redis container: %v", termErr)
		}
		t.Fatalf("Failed to get Redis host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("failed to terminate redis container: %v", termErr)
		}
		t.Fatalf("Failed to get Redis port: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port.Port(),
	})

	return client, container
}

// TestHealthEndpoint_AllOK validates health endpoint when all services are healthy
func TestHealthEndpoint_AllOK(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup DB
	pool, dbContainer := setupTestDBForHealth(ctx, t)
	defer func() {
		if err := dbContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate db container: %v", err)
		}
	}()
	defer pool.Close()

	// Setup Redis
	redisClient, redisContainer := setupTestRedisForHealth(ctx, t)
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}()

	// Create health handler
	handler := NewHealthHandler(pool, redisClient)

	// Create echo context with test request/response
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.Check(c)
	if err != nil {
		t.Fatalf("Health handler failed: %v", err)
	}

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Status string `json:"status"`
		DB     string `json:"db"`
		Redis  string `json:"redis"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", resp.Status)
	}
	if resp.DB != "ok" {
		t.Errorf("Expected db 'ok', got '%s'", resp.DB)
	}
	if resp.Redis != "ok" {
		t.Errorf("Expected redis 'ok', got '%s'", resp.Redis)
	}
}

// TestHealthEndpoint_RedisDown validates health endpoint when Redis is unavailable
func TestHealthEndpoint_RedisDown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup DB only
	pool, dbContainer := setupTestDBForHealth(ctx, t)
	defer func() {
		if err := dbContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate db container: %v", err)
		}
	}()
	defer pool.Close()

	// Use unreachable Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999",
	})

	// Create health handler
	handler := NewHealthHandler(pool, redisClient)

	// Create echo context with test request/response
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.Check(c)
	if err != nil {
		t.Fatalf("Health handler failed: %v", err)
	}

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Status string `json:"status"`
		DB     string `json:"db"`
		Redis  string `json:"redis"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.DB != "ok" {
		t.Errorf("Expected db 'ok', got '%s'", resp.DB)
	}
	if resp.Redis == "ok" {
		t.Error("Expected redis 'error', but got 'ok'")
	}
}

// TestHealthEndpoint_DBDown validates health endpoint when database is unavailable
func TestHealthEndpoint_DBDown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Setup Redis only
	redisClient, redisContainer := setupTestRedisForHealth(ctx, t)
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}()

	// Use unreachable DB pool
	pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:9999/db")
	if err != nil {
		t.Fatalf("Failed to create pool config: %v", err)
	}
	defer pool.Close()

	// Create health handler
	handler := NewHealthHandler(pool, redisClient)

	// Create echo context with test request/response
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call handler
	err = handler.Check(c)
	if err != nil {
		t.Fatalf("Health handler failed: %v", err)
	}

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Status string `json:"status"`
		DB     string `json:"db"`
		Redis  string `json:"redis"`
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if resp.DB == "ok" {
		t.Error("Expected db 'error', but got 'ok'")
	}
	if resp.Redis != "ok" {
		t.Errorf("Expected redis 'ok', got '%s'", resp.Redis)
	}
}
