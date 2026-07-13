-- name: ListFilms :many
SELECT
    id,
    title,
    year,
    runtime,
    synopsis,
    imdb_id,
    poster_url,
    created_at
FROM films
ORDER BY created_at DESC
LIMIT 100;

-- name: GetFilm :one
SELECT
    id,
    title,
    year,
    runtime,
    synopsis,
    imdb_id,
    poster_url,
    created_at
FROM films
WHERE id = $1;

-- name: InsertFilm :one
-- Usada por CreateFilm (RF02/task 0002): o subcomando CLI criar-filme insere o
-- filme e enfileira catalogo.filme_criado na mesma TX via outbox.Enqueue.
-- films.id é BIGINT sem identity/sequence (schema da migration 001); o próximo
-- id é calculado por MAX(id)+1 dentro do próprio statement — simplificação
-- aceitável para uma ferramenta de operador único (CLI, sem concorrência real);
-- colisão eventual é reportada como erro de constraint (exit code != 0 na CLI).
INSERT INTO films (id, title, year, runtime, synopsis)
SELECT COALESCE(MAX(id), 0) + 1, $1, $2, $3, $4
FROM films
RETURNING id, title, year, runtime, synopsis, imdb_id, poster_url, created_at;
