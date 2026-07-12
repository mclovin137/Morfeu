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

-- name: ProvaQueryQuebrada :many
SELEC * FORM films WHERE;
