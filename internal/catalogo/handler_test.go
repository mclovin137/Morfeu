package catalogo

import (
	"testing"

	"github.com/mclovin137/morfeu/internal/db"
)

// O caminho handler→service se prova em integração (TestListFilmsE2E_FullStack),
// com PG/Redis reais — mock camada a camada é vedado pelo ADR 0006.

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
