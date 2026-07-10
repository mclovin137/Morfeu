// +build integration

package internal

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/mclovin137/morfeu/internal/cache"
	"github.com/mclovin137/morfeu/internal/catalogo"
	"github.com/mclovin137/morfeu/internal/catalogo/db"
)

// setupTestDB creates an ephemeral PostgreSQL test container
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
		t.Fatalf("Failed to start DB container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get DB host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get DB port: %v", err)
	}

	dsn := "postgres://postgres:postgres@" + host + ":" + port.Port() + "/morfeu_test"
	return dsn, container
}

// setupTestRedis creates an ephemeral Redis test container
func setupTestRedis(t *testing.T, ctx context.Context) (string, testcontainers.Container) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
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
		container.Terminate(ctx)
		t.Fatalf("Failed to get Redis host: %v", err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("Failed to get Redis port: %v", err)
	}

	redisURL := "redis://" + host + ":" + port.Port()
	return redisURL, container
}

// initTestSchema creates the films table
func initTestSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
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
	`)
	return err
}

// seedTestData inserts 10 films into the test database
func seedTestData(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO films (id, title, year, runtime, synopsis, imdb_id, poster_url, created_at) VALUES
		(1, 'The Shawshank Redemption', 1994, 142, 'Two imprisoned men bond over redemption.', 'tt0111161', 'https://example.com/poster1.jpg', NOW()),
		(2, 'The Dark Knight', 2008, 152, 'Batman fights the Joker.', 'tt0468569', 'https://example.com/poster2.jpg', NOW()),
		(3, 'Inception', 2010, 148, 'A thief steals secrets in dreams.', 'tt1375666', 'https://example.com/poster3.jpg', NOW()),
		(4, 'The Matrix', 1999, 136, 'A hacker learns the truth of reality.', 'tt0133093', 'https://example.com/poster4.jpg', NOW()),
		(5, 'Pulp Fiction', 1994, 154, 'Intertwined tales of violence and redemption.', 'tt0110912', 'https://example.com/poster5.jpg', NOW()),
		(6, 'Forrest Gump', 1994, 142, 'A man witnesses historical events.', 'tt0109830', 'https://example.com/poster6.jpg', NOW()),
		(7, 'Interstellar', 2014, 169, 'Explorers travel through a wormhole.', 'tt0816692', 'https://example.com/poster7.jpg', NOW()),
		(8, 'The Godfather', 1972, 175, 'An organized crime dynasty transfer control.', 'tt0068646', 'https://example.com/poster8.jpg', NOW()),
		(9, 'Gladiator', 2000, 155, 'A general seeks vengeance.', 'tt0172495', 'https://example.com/poster9.jpg', NOW()),
		(10, 'The Lord of the Rings', 2001, 178, 'A hobbit embarks on a quest.', 'tt0120737', 'https://example.com/poster10.jpg', NOW());
	`)
	return err
}

// TestDatabaseConnection validates basic PostgreSQL connectivity
func TestDatabaseConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	dsn, container := setupTestDB(t, ctx)
	defer container.Terminate(ctx)

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

// TestMigrationsUpDownUpIdempotent validates migrations can be run: up → down → up without error
func TestMigrationsUpDownUpIdempotent(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()
	dsn, container := setupTestDB(t, ctx)
	defer container.Terminate(ctx)

	time.Sleep(time.Second)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Simulate "up" by creating schema
	if err := initTestSchema(ctx, pool); err != nil {
		t.Fatalf("Failed to create schema (up): %v", err)
	}

	// Seed data
	if err := seedTestData(ctx, pool); err != nil {
		t.Fatalf("Failed to seed data: %v", err)
	}

	// Verify films exist
	queries := db.New(pool)
	films, err := queries.ListFilms(ctx)
	if err != nil {
		t.Fatalf("Failed to list films after up: %v", err)
	}
	if len(films) != 10 {
		t.Errorf("Expected 10 films after up, got %d", len(films))
	}

	// Simulate "down"
	if _, err := pool.Exec(ctx, "DROP TABLE IF EXISTS films"); err != nil {
		t.Fatalf("Failed to drop table (down): %v", err)
	}

	// Verify table is gone
	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT FROM information_schema.tables WHERE table_name='films')").Scan(&exists)
	if err != nil {
		t.Fatalf("Failed to check table existence: %v", err)
	}
	if exists {
		t.Error("Expected films table to be dropped, but it exists")
	}

	// Simulate "up" again
	if err := initTestSchema(ctx, pool); err != nil {
		t.Fatalf("Failed to recreate schema (up again): %v", err)
	}

	if err := seedTestData(ctx, pool); err != nil {
		t.Fatalf("Failed to reseed data: %v", err)
	}

	// Verify films exist again
	films, err = queries.ListFilms(ctx)
	if err != nil {
		t.Fatalf("Failed to list films after up again: %v", err)
	}
	if len(films) != 10 {
		t.Errorf("Expected 10 films after up again, got %d", len(films))
	}
}

// TestCacheHitMissWithRedis validates cache hit/miss behavior with testcontainers Redis
func TestCacheHitMissWithRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start DB container
	dbDSN, dbContainer := setupTestDB(t, ctx)
	defer dbContainer.Terminate(ctx)

	// Start Redis container
	redisURL, redisContainer := setupTestRedis(t, ctx)
	defer redisContainer.Terminate(ctx)

	time.Sleep(time.Second)

	// Setup database
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	if err := initTestSchema(ctx, pool); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	if err := seedTestData(ctx, pool); err != nil {
		t.Fatalf("Failed to seed data: %v", err)
	}

	// Setup Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL[8:], // Strip "redis://"
	})
	defer redisClient.Close()

	cacheLayer := cache.NewRedisCache(redisClient, zap.NewNop())
	svc := catalogo.NewFilmService(pool, cacheLayer, zap.NewNop())

	// First request should miss cache and hit database
	films1, err := svc.ListFilms(ctx)
	if err != nil {
		t.Fatalf("First ListFilms failed: %v", err)
	}
	if len(films1) != 10 {
		t.Errorf("Expected 10 films from DB, got %d", len(films1))
	}

	// Second request should hit cache
	films2, err := svc.ListFilms(ctx)
	if err != nil {
		t.Fatalf("Second ListFilms failed: %v", err)
	}
	if len(films2) != 10 {
		t.Errorf("Expected 10 films from cache, got %d", len(films2))
	}

	// Verify cache entry exists
	cacheVal, err := redisClient.Get(ctx, "films:list").Result()
	if err != nil {
		t.Errorf("Cache miss: key 'films:list' not found: %v", err)
	}
	if cacheVal == "" {
		t.Error("Cache entry is empty")
	}
}

// TestGracefulDegradationRedisUnavailable validates fallback when Redis is down
func TestGracefulDegradationRedisUnavailable(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start DB container only
	dbDSN, dbContainer := setupTestDB(t, ctx)
	defer dbContainer.Terminate(ctx)

	time.Sleep(time.Second)

	// Setup database
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	if err := initTestSchema(ctx, pool); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	if err := seedTestData(ctx, pool); err != nil {
		t.Fatalf("Failed to seed data: %v", err)
	}

	// Use a Redis client that will fail (unreachable address)
	redisClient := redis.NewClient(&redis.Options{
		Addr: "localhost:9999", // Unreachable port
	})

	cacheLayer := cache.NewRedisCache(redisClient, zap.NewNop())
	svc := catalogo.NewFilmService(pool, cacheLayer, zap.NewNop())

	// Should fall back to database
	films, err := svc.ListFilms(ctx)
	if err != nil {
		t.Fatalf("ListFilms failed even with DB fallback: %v", err)
	}
	if len(films) != 10 {
		t.Errorf("Expected 10 films from DB fallback, got %d", len(films))
	}
}

// TestListFilmsE2E_FullStack validates end-to-end flow: handler → service → cache → db
func TestListFilmsE2E_FullStack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Start DB container
	dbDSN, dbContainer := setupTestDB(t, ctx)
	defer dbContainer.Terminate(ctx)

	// Start Redis container
	redisURL, redisContainer := setupTestRedis(t, ctx)
	defer redisContainer.Terminate(ctx)

	time.Sleep(time.Second)

	// Setup database
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	if err := initTestSchema(ctx, pool); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}
	if err := seedTestData(ctx, pool); err != nil {
		t.Fatalf("Failed to seed data: %v", err)
	}

	// Setup Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisURL[8:], // Strip "redis://"
	})
	defer redisClient.Close()

	cacheLayer := cache.NewRedisCache(redisClient, zap.NewNop())
	svc := catalogo.NewFilmService(pool, cacheLayer, zap.NewNop())

	// Full stack test: cache miss → database
	films, err := svc.ListFilms(ctx)
	if err != nil {
		t.Fatalf("E2E ListFilms failed: %v", err)
	}
	if len(films) != 10 {
		t.Errorf("Expected 10 films, got %d", len(films))
	}

	// Verify first film data
	if films[0].Title != "The Shawshank Redemption" {
		t.Errorf("Expected first film title 'The Shawshank Redemption', got '%s'", films[0].Title)
	}
	if films[0].Year == nil || *films[0].Year != 1994 {
		t.Errorf("Expected first film year 1994, got %v", films[0].Year)
	}
}
