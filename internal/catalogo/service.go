package catalogo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/mclovin137/morfeu/internal/cache"
	"github.com/mclovin137/morfeu/internal/db"
)

// filmsCacheTTL: PRD 0001 (CA05) — cache read-through com TTL de 5 minutos
const filmsCacheTTL = 5 * time.Minute

// FilmService handles film-related business logic
type FilmService struct {
	db    *pgxpool.Pool
	cache cache.Cache
	logger *zap.Logger
}

// NewFilmService creates a new film service
func NewFilmService(db *pgxpool.Pool, cache cache.Cache, logger *zap.Logger) *FilmService {
	return &FilmService{
		db:    db,
		cache: cache,
		logger: logger,
	}
}

// ListFilms lists all films with cache read-through strategy
func (fs *FilmService) ListFilms(ctx context.Context) ([]db.Film, error) {
	cacheKey := "films:list"

	// Try to get from cache first
	cachedData, err := fs.cache.Get(ctx, cacheKey)
	if err != nil {
		fs.logger.Warn("cache error, falling back to database",
			zap.String("action", "cache_error"),
			zap.Error(err),
		)
	}

	if cachedData != nil {
		fs.logger.Info("serving from cache",
			zap.String("action", "cache_hit"),
		)
		var films []db.Film
		if err := json.Unmarshal(cachedData, &films); err != nil {
			fs.logger.Warn("failed to unmarshal cached data",
				zap.Error(err),
			)
			// Continue to fetch from database
		} else {
			return films, nil
		}
	}

	// Cache miss or error - fetch from database
	fs.logger.Info("fetching from database",
		zap.String("action", "database_query"),
	)

	queries := db.New(fs.db)
	films, err := queries.ListFilms(ctx)
	if err != nil {
		fs.logger.Error("failed to fetch films from database",
			zap.Error(err),
		)
		return nil, err
	}

	// Store in cache for future requests (best effort)
	if data, err := json.Marshal(films); err == nil {
		if err := fs.cache.Set(ctx, cacheKey, data, filmsCacheTTL); err != nil {
			fs.logger.Warn("failed to set cache",
				zap.Error(err),
			)
			// Not critical - we already have the data from database
		}
	}

	return films, nil
}
