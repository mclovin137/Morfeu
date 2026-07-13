package catalogo

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/mclovin137/morfeu/internal/cache"
	"github.com/mclovin137/morfeu/internal/catalogo/db"
	"github.com/mclovin137/morfeu/internal/outbox"
)

// filmsCacheTTL: PRD 0001 (CA05) — cache read-through com TTL de 5 minutos
const filmsCacheTTL = 5 * time.Minute

// eventoFilmeCriado é o event_type do evento de domínio emitido por
// CreateFilm na outbox (RF01/RF02, PRD 0002).
const eventoFilmeCriado = "catalogo.filme_criado"

// FilmService handles film-related business logic
type FilmService struct {
	queries *db.Queries
	pool    outbox.Pool
	cache   cache.Cache
	logger  *zap.Logger
}

// NewFilmService creates a new film service. queries é o DAO gerado pelo sqlc
// (internal/catalogo/db) — o domínio depende dele, não do pool pgx cru (ADR 0003).
// pool é usado só para abrir a transação de CreateFilm via outbox.WithTx
// (RF01/RF02) — nunca importado diretamente aqui (alias outbox.Pool).
func NewFilmService(queries *db.Queries, pool outbox.Pool, cache cache.Cache, logger *zap.Logger) *FilmService {
	return &FilmService{
		queries: queries,
		pool:    pool,
		cache:   cache,
		logger:  logger,
	}
}

// CreateFilmParams são os dados de entrada de CreateFilm (RF02): usados pelo
// subcomando CLI criar-filme, validados antes de chegar aqui (forma) — a
// única regra de negócio é a atomicidade filme+evento (RN01).
type CreateFilmParams struct {
	Title    string
	Year     *int32
	Runtime  *int32
	Synopsis *string
}

// filmeCriadoPayload é o shape JSON do payload do evento catalogo.filme_criado
// (RF05) — o consumidor (task 0005) faz o parse deste mesmo shape.
type filmeCriadoPayload struct {
	ID       int64   `json:"id"`
	Title    string  `json:"title"`
	Year     *int32  `json:"year,omitempty"`
	Runtime  *int32  `json:"runtime,omitempty"`
	Synopsis *string `json:"synopsis,omitempty"`
}

// CreateFilm insere o filme e enfileira catalogo.filme_criado na outbox
// dentro da mesma transação (RF01/RF02): rollback de um é rollback do outro
// (RN01, CA01). É o único caminho de escrita do módulo catalogo até agora.
func (fs *FilmService) CreateFilm(ctx context.Context, params CreateFilmParams) (db.Film, error) {
	var created db.Film

	err := outbox.WithTx(ctx, fs.pool, func(tx outbox.Tx) error {
		q := fs.queries.WithTx(tx)

		film, err := q.InsertFilm(ctx, db.InsertFilmParams{
			Title:    params.Title,
			Year:     params.Year,
			Runtime:  params.Runtime,
			Synopsis: params.Synopsis,
		})
		if err != nil {
			return fmt.Errorf("inserir filme: %w", err)
		}

		payload, err := json.Marshal(filmeCriadoPayload{
			ID:       film.ID,
			Title:    film.Title,
			Year:     film.Year,
			Runtime:  film.Runtime,
			Synopsis: film.Synopsis,
		})
		if err != nil {
			return fmt.Errorf("serializar payload do evento %s: %w", eventoFilmeCriado, err)
		}

		if _, err := outbox.Enqueue(ctx, tx, outbox.Evento{
			EventType:   eventoFilmeCriado,
			AggregateID: strconv.FormatInt(film.ID, 10),
			OccurredAt:  time.Now(),
			Payload:     payload,
		}); err != nil {
			return err
		}

		created = film
		return nil
	})
	if err != nil {
		fs.logger.Error("falha ao criar filme", zap.Error(err))
		return db.Film{}, err
	}

	fs.logger.Info("filme criado",
		zap.Int64("film_id", created.ID),
		zap.String("event_type", eventoFilmeCriado),
	)
	return created, nil
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
		unmarshalErr := json.Unmarshal(cachedData, &films)
		if unmarshalErr == nil {
			return films, nil
		}
		fs.logger.Warn("failed to unmarshal cached data",
			zap.Error(unmarshalErr),
		)
		// Continue to fetch from database
	}

	// Cache miss or error - fetch from database
	fs.logger.Info("fetching from database",
		zap.String("action", "database_query"),
	)

	films, err := fs.queries.ListFilms(ctx)
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
