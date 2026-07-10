package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/mclovin137/morfeu/internal/cache"
	"github.com/mclovin137/morfeu/internal/catalogo"
	"github.com/mclovin137/morfeu/internal/config"
	"github.com/mclovin137/morfeu/internal/health"
	"github.com/mclovin137/morfeu/internal/logger"
)

func main() {
	// Load configuration from environment
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log := logger.NewLogger(cfg.LogLevel)
	defer log.Sync()

	log.Info("Starting Morfeu application",
		zap.String("app_port", cfg.AppPort),
		zap.String("log_level", cfg.LogLevel),
	)

	// Create database connection pool
	dbPool, err := createDBPool(cfg, log)
	if err != nil {
		log.ErrorMsg("Failed to create database pool", zap.Error(err))
		os.Exit(1)
	}
	defer dbPool.Close()

	log.Info("Database pool created", zap.Int("min_size", cfg.PoolMinSize), zap.Int("max_size", cfg.PoolMaxSize))

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.ErrorMsg("Failed to connect to Redis", zap.Error(err))
		// Don't exit - Redis is optional for read-through cache
	} else {
		log.Info("Connected to Redis")
	}

	// Run migrations
	if err := runMigrations(cfg.DatabaseURL, log); err != nil {
		log.ErrorMsg("Migration failed", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Migrations completed")

	// Create cache layer
	cacheLayer := cache.NewRedisCache(redisClient, log.Logger)

	// Create service layer
	filmService := catalogo.NewFilmService(dbPool, cacheLayer, log.Logger)

	// Create handlers
	filmHandler := catalogo.NewFilmHandler(filmService)
	healthHandler := health.NewHealthHandler(dbPool, redisClient)

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339} | ${method} ${uri} | ${status} | ${latency_human}` + "\n",
	}))

	// Routes
	e.GET("/health", healthHandler.Check)
	e.GET("/filmes", filmHandler.ListFilms)

	// Start server in a goroutine
	go func() {
		if err := e.Start(":" + cfg.AppPort); err != nil && err != http.ErrServerClosed {
			log.ErrorMsg("Server error", zap.Error(err))
		}
	}()

	log.Info("Server started", zap.String("port", cfg.AppPort))

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.ErrorMsg("Server shutdown error", zap.Error(err))
		os.Exit(1)
	}

	log.Info("Server stopped")
}

// createDBPool creates a PostgreSQL connection pool
func createDBPool(cfg *config.Config, log *logger.Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	poolConfig.MinConns = int32(cfg.PoolMinSize)
	poolConfig.MaxConns = int32(cfg.PoolMaxSize)
	poolConfig.MaxConnLifetime = time.Minute * 15
	poolConfig.MaxConnIdleTime = time.Minute * 5
	poolConfig.ConnConfig.ConnectTimeout = cfg.PoolTimeout

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

// runMigrations runs pending database migrations
func runMigrations(databaseURL string, log *logger.Logger) error {
	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, dirty, err := m.Version()
	if err == migrate.ErrNilVersion {
		log.Info("No migrations applied")
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	log.Info("Migrations applied", zap.Uint("version", version), zap.Bool("dirty", dirty))
	return nil
}
