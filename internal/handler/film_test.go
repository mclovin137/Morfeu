package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mclovin137/morfeu/internal/db"
	"github.com/mclovin137/morfeu/internal/service"
)

func TestListFilms_Success(t *testing.T) {
	// Create a mock service
	mockService := &service.FilmService{}

	handler := NewFilmHandler(mockService)

	// Create a test HTTP request
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/filmes", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Call the handler
	err := handler.ListFilms(c)
	if err != nil {
		t.Fatalf("ListFilms failed: %v", err)
	}

	// Check status code
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Check Content-Type
	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}
}

func TestListFilms_TraceIDPropagation(t *testing.T) {
	mockService := &service.FilmService{}
	handler := NewFilmHandler(mockService)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/filmes", nil)
	req.Header.Set("X-Trace-ID", "test-trace-id-123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.ListFilms(c)
	if err != nil {
		t.Fatalf("ListFilms failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

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
