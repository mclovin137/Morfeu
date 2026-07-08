# Task 0001 — App Skeleton Go

Implementação da estrutura base da aplicação Morfeu com Go + Echo, PostgreSQL + Redis, migrations, e endpoint de exemplo.

## Setup

### Pré-requisitos

- Docker e Docker Compose
- Go 1.21+
- migrate CLI (para rodar migrations manualmente)
- golangci-lint (para lint)

### Iniciar stack local

```bash
# Start PostgreSQL + Redis
make up

# Run migrations
make migrate

# Build the application
make build

# Run the application
make run
```

A aplicação estará disponível em `http://localhost:8080`.

## Endpoints

### GET /filmes

Retorna lista de filmes em cache ou do banco de dados.

```bash
curl http://localhost:8080/filmes
```

Resposta (200 OK):
```json
[
  {
    "id": 1,
    "title": "The Shawshank Redemption",
    "year": 1994,
    "runtime": 142,
    "synopsis": "...",
    "imdb_id": "tt0111161",
    "poster_url": "https://example.com/poster1.jpg"
  }
]
```

**Cache behavior:**
- Primeira requisição: busca no banco de dados (lento)
- Requisições subsequentes (< 5 minutos): servidas do Redis (rápido)
- Logs indicam `action: cache_hit` ou `action: cache_miss`

### GET /health

Verifica saúde da aplicação (PG + Redis).

```bash
curl http://localhost:8080/health
```

Resposta (200 OK):
```json
{
  "status": "ok",
  "db": "ok",
  "redis": "ok"
}
```

Se Redis estiver down, continua retornando 200 com `redis: error`. Fallback automático para banco.

## Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│ Client (curl/browser)                                           │
└──────────────────────────┬──────────────────────────────────────┘
                           │ HTTP
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│ Echo Router (main.go)                                           │
│ - GET /filmes → FilmHandler                                     │
│ - GET /health → HealthHandler                                   │
├─────────────────────────────────────────────────────────────────┤
│ Handlers (trace_id propagation, HTTP serialization)             │
├─────────────────────────────────────────────────────────────────┤
│ Services (business logic, cache orchestration)                  │
├─────────────────────────────────────────────────────────────────┤
│ Cache Layer (Redis read-through with 500ms timeout)             │
├─────────────────────────────────────────────────────────────────┤
│ Database Layer (sqlc-generated, pgx/v5)                         │
└─────────────────────────────────────────────────────────────────┘
           │                                          │
           ▼                                          ▼
    ┌─────────────┐                        ┌─────────────────┐
    │   Redis 7   │                        │  PostgreSQL 16  │
    │ (512MB max) │                        │  (films table)  │
    └─────────────┘                        └─────────────────┘
```

### Fluxo de requisição

1. **Handler** recebe requisição HTTP
   - Extrai/gera `X-Trace-ID` header
   - Injeta trace_id no contexto
   - Chama service

2. **Service** orquestra cache + db
   - Tenta `cache.Get("films:list")`
   - Se cache hit: retorna dados
   - Se cache miss/erro: query ao banco (`db.ListFilms`)
   - Persiste resultado no cache (best effort)

3. **Cache** com timeout (~500ms)
   - Timeout → retorna erro, service cai para db
   - Logs estruturados JSON (action, duration, key)

4. **Database** retorna dados
   - Queries geradas por sqlc (type-safe)
   - Pool pgx (min=5, max=25 conexões)

## Testes

### Unit tests

```bash
make test
```

Testa:
- Config parsing (env overrides)
- Logger initialization (JSON output)
- Cache layer (timeouts, logging)
- Handlers (HTTP response format, trace_id)

### Integration tests

```bash
# Build testcontainers environment
go test -tags=integration -race ./...
```

Testa:
- Database connectivity (real PostgreSQL container)
- Migrations idempotência (up → down → up)
- Service com cache hit/miss real
- Handler → Service → Cache → DB end-to-end

## Desenvolvimento

### Lint

```bash
make lint
```

Regras golangci-lint (ADR 0004):
- Complexidade cognitiva ≤ 15
- Funções ≤ 80 linhas
- Indentação ≤ 5 níveis
- Sem `else` (early return)
- Context sempre 1º parâmetro
- Sem secrets nos logs

### Code generation

Se modificar `migrations/` ou `internal/db/queries.sql`:

```bash
make gen
```

Regenera `internal/db/models.go` e `internal/db/queries.sql.go` via sqlc.

### Migrations

Para criar nova migration:

```bash
migrate create -ext sql -dir migrations -seq <name>
```

Exemplo:
```bash
migrate create -ext sql -dir migrations -seq add_sessions_table
```

Regras:
- Sempre `IF NOT EXISTS` / `IF EXISTS` (idempotente)
- Rollback definido no `.down.sql`
- Nomes: PascalCase, descritivos

## Troubleshooting

### Redis não conecta

```bash
docker-compose logs morfeu-redis
```

Se Redis estiver down, a aplicação continua funcionando com fallback ao PG.

### PostgreSQL não conecta

```bash
docker-compose logs morfeu-postgres
```

App não inicia sem PG (banco é obrigatório).

### Migrations falham

```bash
# Ver status
migrate -path ./migrations -database $DATABASE_URL version

# Forçar versão (último recurso)
migrate -path ./migrations -database $DATABASE_URL force <version>
```

### Cache não funciona

Verificar logs:
```bash
docker-compose logs morfeu-postgres
docker-compose logs morfeu-redis
```

Procurar por `cache_miss`, `cache_hit`, ou `cache_error` no stderr da app.

## Próximos passos (E0b–E0c)

- **E0b (Walking skeleton)**: Adicionar outbox pattern + RabbitMQ worker
- **E0c (Walking skeleton)**: TLS/HTTPS, CI/CD pipeline, observabilidade (traces OTel)
- **E1 (Identidade)**: Autenticação JWT, roles, auditoria de login

## Referências

- **ADR 0001** — Stack Go + Echo
- **ADR 0002** — sqlc + pgx strategy
- **ADR 0003** — Fronteiras de módulos (handler → service → db)
- **ADR 0004** — Padrões de código Go (logging, config, errors)
- **ADR 0006** — Estratégia de testes
- **PRD 0001** — Requisitos funcionais e critérios de aceite
- **roles.md** — Governança do projeto
