package catalogo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/mclovin137/morfeu/internal/cache"
	"github.com/mclovin137/morfeu/internal/catalogo/db"
)

func TestFilmResponseConversion(t *testing.T) {
	film := db.Film{
		ID:    1,
		Title: "Test Film",
	}

	response := FilmResponse{
		ID:    film.ID,
		Title: film.Title,
	}

	if response.ID != 1 || response.Title != "Test Film" {
		t.Error("Film response conversion failed")
	}
}

// setupTestDBForCatalogo creates an ephemeral PostgreSQL test container
func setupTestDBForCatalogo(ctx context.Context, t *testing.T) (*pgxpool.Pool, testcontainers.Container) {
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

// setupTestRedisForCatalogo creates an ephemeral Redis test container
func setupTestRedisForCatalogo(ctx context.Context, t *testing.T) (*redis.Client, testcontainers.Container) {
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

// seedTestFilms creates the films table and inserts fixture rows
func seedTestFilms(ctx context.Context, pool *pgxpool.Pool) error {
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
	`); err != nil {
		return err
	}

	_, err := pool.Exec(ctx, `
		INSERT INTO films (id, title, year, runtime, synopsis, imdb_id, poster_url, created_at) VALUES
		(1, 'The Shawshank Redemption', 1994, 142, 'Two imprisoned men bond over redemption.', 'tt0111161', 'https://example.com/poster1.jpg', NOW()),
		(2, 'The Dark Knight', 2008, 152, 'Batman fights the Joker.', 'tt0468569', 'https://example.com/poster2.jpg', NOW());
	`)
	return err
}

// TestListFilmsHTTP_Integration valida GET /filmes via echo real (não chamada
// direta ao service), com PG/Redis reais (testcontainers) — RF02/CA05 do PRD
// 0003, cobrindo o achado da auditoria de 2026-07-10.
func TestListFilmsHTTP_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	pool, dbContainer := setupTestDBForCatalogo(ctx, t)
	defer func() {
		if err := dbContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate db container: %v", err)
		}
	}()
	defer pool.Close()

	redisClient, redisContainer := setupTestRedisForCatalogo(ctx, t)
	defer func() {
		if err := redisContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate redis container: %v", err)
		}
	}()
	defer func() {
		if err := redisClient.Close(); err != nil {
			t.Logf("failed to close redis client: %v", err)
		}
	}()

	time.Sleep(time.Second)

	if err := seedTestFilms(ctx, pool); err != nil {
		t.Fatalf("Failed to seed films: %v", err)
	}

	cacheLayer := cache.NewRedisCache(redisClient, zap.NewNop())
	svc := NewFilmService(db.New(pool), cacheLayer, zap.NewNop())
	h := NewFilmHandler(svc)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/filmes", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.ListFilms(c); err != nil {
		t.Fatalf("ListFilms handler failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get(echo.HeaderContentType)
	if !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	var films []FilmResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &films); err != nil {
		t.Fatalf("Failed to decode response body into film list: %v", err)
	}

	if len(films) != 2 {
		t.Errorf("Expected 2 films, got %d", len(films))
	}
}
