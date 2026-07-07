# sqlc

Context7:

- Resolvido como `/websites/sqlc_dev_en`
- Documentacao consultada para `sqlc.yaml`, PostgreSQL, `pgx/v5`,
  anotacoes `:one`, `:many`, `:exec`, tipos gerados e uso com transacoes.

Registro local: [`lib.md`](../../lib.md) define sqlc como ferramenta planejada
de desenvolvimento.

## Papel no Morfeu

`sqlc` gera codigo Go type-safe a partir de SQL explicito. A decisao combina
com os ADRs existentes: SQL visivel, migrations versionadas e menos magia que
ORMs por reflection.

## Layout recomendado

```text
internal/
  platform/
    db/
      migrations/
        000001_create_users.up.sql
        000001_create_users.down.sql
      query/
        users.sql
        reservations.sql
      sqlc/
        db.go
        models.go
        users.sql.go
sqlc.yaml
```

O diretorio gerado nao deve receber edicoes manuais.

## Configuracao

Context7 confirmou o uso de `version: "2"`, `engine: "postgresql"` e
`sql_package: "pgx/v5"`.

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "internal/platform/db/migrations"
    queries: "internal/platform/db/query"
    gen:
      go:
        package: "db"
        out: "internal/platform/db/sqlc"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
```

Usar a mesma fonte de schema das migrations. Se uma migration ainda nao foi
aplicada, o SQL gerado tambem nao deve assumir sua existencia.

## Anotacoes SQL

```sql
-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at
FROM users
WHERE email = $1
LIMIT 1;

-- name: ListSessionsByMovie :many
SELECT id, movie_id, starts_at, room_id
FROM sessions
WHERE movie_id = $1
ORDER BY starts_at;

-- name: CreateOutboxEvent :one
INSERT INTO outbox_events (topic, aggregate_id, payload)
VALUES ($1, $2, $3)
RETURNING id, created_at;

-- name: MarkOutboxEventPublished :exec
UPDATE outbox_events
SET published_at = now()
WHERE id = $1;
```

Usar `:one` quando a query deve retornar exatamente um registro. Tratar
`pgx.ErrNoRows` na camada de repositorio e converter para erro de dominio.

## Uso com pgx

```go
type UserRepository struct {
	q *db.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{q: db.New(pool)}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrUserNotFound
		}
		return User{}, fmt.Errorf("get user by email: %w", err)
	}
	return mapUser(row), nil
}
```

## Transacoes com `WithTx`

Context7 confirmou que o codigo gerado pode receber uma transacao por `WithTx`.

```go
func (r *ReservationRepository) CreateReservationTx(
	ctx context.Context,
	fn func(q *db.Queries) error,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := r.q.WithTx(tx)
	if err := fn(qtx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
```

Reservas, pagamento e outbox devem ser consistentes: gravar estado transacional
e evento de outbox na mesma transacao.

## Tipos nulos

Com `pgx/v5`, campos nullable podem aparecer como tipos `pgtype`. Mapear para
tipos de dominio na borda do repositorio para nao vazar detalhe de persistencia:

```go
func textValue(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}
```

