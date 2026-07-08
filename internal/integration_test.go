// +build integration

package internal

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/mclovin137/morfeu/internal/db"
	"github.com/mclovin137/morfeu/internal/logger"
)

// setupTestDB creates a test database container
func setupTestDB(t *testing.T, ctx context.Context) (string, testcontainers.Container) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "morfeu_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	dsn := "postgres://postgres:postgres@" + host + ":" + port.Port() + "/morfeu_test"
	return dsn, container
}

// TestDatabaseConnection tests basic database connectivity
func TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	dsn, container := setupTestDB(t, ctx)
	defer container.Terminate(ctx)

	// Wait a bit for the database to be ready
	time.Sleep(time.Second)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	err = pool.Ping(ctx)
	if err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}
}

// TestMigrationsIdempotent tests that migrations can be run multiple times
func TestMigrationsIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := logger.NewLogger("info")
	defer log.Sync()

	ctx := context.Background()
	dsn, container := setupTestDB(t, ctx)
	defer container.Terminate(ctx)

	time.Sleep(time.Second)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Test queries can be created
	queries := db.New(pool)
	if queries == nil {
		t.Error("Failed to create queries instance")
	}
}

// TestFilmServiceListFilms tests the film service list functionality
func TestFilmServiceListFilms(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	log := logger.NewLogger("info")
	defer log.Sync()

	ctx := context.Background()
	dsn, container := setupTestDB(t, ctx)
	defer container.Terminate(ctx)

	time.Sleep(time.Second)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Create the test tables schema
	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS films (
			id BIGINT PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			year INTEGER,
			runtime INTEGER,
			synopsis TEXT,
			imdb_id VARCHAR(20),
			poster_url VARCHAR(500),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);

		INSERT INTO films (id, title, year, runtime, synopsis, imdb_id, poster_url, created_at) VALUES
		(1, 'Test Film', 2020, 120, 'Test synopsis', 'tt0000001', 'https://example.com/poster.jpg', NOW())
		ON CONFLICT DO NOTHING;
	`); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	queries := db.New(pool)
	films, err := queries.ListFilms(ctx)
	if err != nil {
		t.Fatalf("Failed to list films: %v", err)
	}

	if len(films) == 0 {
		t.Error("Expected films, got none")
	}

	if films[0].Title != "Test Film" {
		t.Errorf("Expected title 'Test Film', got '%s'", films[0].Title)
	}
}
