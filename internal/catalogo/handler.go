package catalogo

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// FilmHandler handles film-related HTTP requests
type FilmHandler struct {
	service *FilmService
}

// NewFilmHandler creates a new film handler
func NewFilmHandler(service *FilmService) *FilmHandler {
	return &FilmHandler{
		service: service,
	}
}

// FilmResponse represents a film in the HTTP response
type FilmResponse struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Year      *int32  `json:"year,omitempty"`
	Runtime   *int32  `json:"runtime,omitempty"`
	Synopsis  *string `json:"synopsis,omitempty"`
	ImdbID    *string `json:"imdb_id,omitempty"`
	PosterUrl *string `json:"poster_url,omitempty"`
}

// ListFilms handles GET /filmes request
func (h *FilmHandler) ListFilms(c echo.Context) error {
	// Extract or generate trace ID
	traceID := c.Request().Header.Get("X-Trace-ID")
	if traceID == "" {
		traceID = uuid.New().String()
	}

	// Create context with trace ID
	ctx := context.WithValue(c.Request().Context(), "trace_id", traceID)

	// Call service
	films, err := h.service.ListFilms(ctx)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to list films",
		})
	}

	// Convert to response format
	response := make([]FilmResponse, len(films))
	for i, film := range films {
		response[i] = FilmResponse{
			ID:        film.ID,
			Title:     film.Title,
			Year:      film.Year,
			Runtime:   film.Runtime,
			Synopsis:  film.Synopsis,
			ImdbID:    film.ImdbID,
			PosterUrl: film.PosterUrl,
		}
	}

	c.Response().Header().Set("Cache-Control", "public, max-age=300")
	return c.JSON(http.StatusOK, response)
}
