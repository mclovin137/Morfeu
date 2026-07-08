# Task 0001 — App Skeleton Go

- **Data:** 2026-07-08
- **Status:** não iniciada
- **Branch:** `feature/0001-app-skeleton-go`
- **PRD:** `docs/prd/0001-app-skeleton-go.md` (criar antes de implementar)
- **Item do roadmap:** E0a (Walking skeleton, subtask a)

## Objetivo

Estrutura base da aplicação Go (app skeleton) com Docker Compose (PostgreSQL + Redis), sistema de migrations (golang-migrate), e um endpoint de exemplo `GET /filmes` servindo dados do banco com cache read-through — pronta para adicionar lógica de negócio.

## Escopo

1. **Estrutura Go** (`main.go`, handlers, config, logger):
   - Entry point (`cmd/app/main.go`)
   - Router Echo com middleware básico (logger, recovery, CORS)
   - Config via env (DSN PG, Redis URL, porta)
   - Simple structured logger (stdout JSON-ready para OTel depois)

2. **Docker Compose**:
   - PostgreSQL 16 (volume local)
   - Redis 7 (volume local, sem persistência por enquanto)
   - Network padrão bridge

3. **Migrations** (`migrations/` via golang-migrate):
   - `001_initial_schema.up.sql`: tabela `films` (id, title, year, runtime, synopsys, imdb_id, poster_url, created_at)
   - Seed com ~10 filmes TMDB reais (stub de dados)

4. **Handler `GET /filmes`**:
   - Busca no banco com sqlc (query `ListFilms`)
   - Cache read-through Redis (TTL 5 min)
   - Retorna JSON array com status 200
   - Headers apropriados (Content-Type, Cache-Control)

5. **Integração mínima**:
   - PG via `pgx` driver
   - Redis via `go-redis/v9`
   - sqlc configurado (queries pré-compiladas)
   - Error logging sem panic (graceful degradation no Redis se cair)

## Fora de escopo

- Autenticação / JWT (E2)
- Validações de entrada (input sanitization — será refatorado com schemas na E3)
- CI/CD pipeline (E0c)
- RabbitMQ / outbox (E0b)
- Observabilidade (traces/métricas — E0d e E10)
- Formatação OpenAPI / Swagger (E0c)
- Multi-tenancy, RBAC (E2)
- Pagination, filtering (será E3+)

## Arquivos esperados

| Arquivo/Diretório | Tipo | Nota |
|---|---|---|
| `cmd/app/main.go` | novo | entry point, setup composer |
| `internal/config/config.go` | novo | config from env |
| `internal/handler/film.go` | novo | handler GET /filmes |
| `internal/logger/logger.go` | novo | structured logger |
| `internal/cache/redis.go` | novo | cache layer (read-through) |
| `internal/db/queries.sql` | novo | sqlc query definitions |
| `internal/db/queries.sql.go` | gerado | sqlc output |
| `internal/db/models.go` | gerado | sqlc output |
| `migrations/001_initial_schema.up.sql` | novo | schema + seed |
| `migrations/001_initial_schema.down.sql` | novo | rollback |
| `docker-compose.yml` | novo | PG + Redis |
| `.env.example` | novo | env template |
| `Dockerfile` | novo | multi-stage build (scratch) |
| `go.mod` | update | add dependencies |
| `go.sum` | update | lock versions |
| `Makefile` | novo | compose up/down, migrate, run targets |

**Estimativa: ~15 arquivos (dentro do limite de 30).**

## Dependências esperadas

### Go modules
- `github.com/labstack/echo/v4` (web framework)
- `github.com/jackc/pgx/v5` (PG driver)
- `github.com/sqlc-dev/sqlc` (SQL compiler)
- `github.com/redis/go-redis/v9` (Redis client)
- `github.com/golang-migrate/migrate/v4` (migrations)
- `go.uber.org/zap` (logging — já definido nos ADRs)

### Infra (Docker)
- PostgreSQL 16 official image
- Redis 7 official image

### Registrado em `lib.md`
- Todas as dependências devem estar listadas antes de usar

## Critérios de aceite

- [ ] App compila sem erros (`go build ./cmd/app`)
- [ ] `docker-compose up` levanta PG + Redis + app em ~5s (banco inicializado)
- [ ] Migrations rodam: `./app migrate up` popula tabela `films` e executa seed
- [ ] `GET http://localhost:8080/filmes` retorna JSON array com ~10 filmes, status 200
- [ ] Segunda requisição a `/filmes` vem do Redis (verificado via logs ou TTL no header)
- [ ] `docker-compose down` para tudo sem erro
- [ ] README.md da task documenta: setup, como rodar, como testar, diagrama da tela E0a
- [ ] Nenhum secret (API key, senha) hardcoded; tudo via `.env` (gitignored)
- [ ] Código passa lint `golangci-lint run` (sem warnings)
- [ ] Logger estruturado (JSON) já ativo (não apenas `log.Println`)

## Riscos

| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| Cache miss/inconsistência (stale data, expiry logic) | Médio | Médio | Testes unitários do cache layer + integração com Redis; iniciar com TTL alto (5 min) |
| Docker networking issues (PG não resolve, conexão recusada) | Baixo | Alto | docker-compose test em CI; ensure default bridge network |
| sqlc code generation issues (wrong types, schema mismatch) | Baixo | Médio | Validar schema primeiro (up), depois rodar `sqlc generate`; checklist de tipos |
| Go modules conflict com ADRs (versão errada de lib) | Baixo | Médio | Consultar Context7 antes de `go get`; validar contra lib.md |
| Port conflicts (8080, 5432, 6379 já em uso) | Médio | Baixo | Use compose service names; localhost:8080 é dev-only |

## Estimativa de impacto

- **Código**: Médio (novo, foundation, mas sem lógica complexa)
- **Banco**: Baixo (schema trivial, seed fake)
- **Infra**: Médio (Docker Compose, networking)
- **Usuários**: N/A (dev-only, não chega a produção)
- **Testes**: Médio (testes de integração obrigatórios para cache)

## Próximos passos (após esta task)

1. **Refinamento multiagente** (`refinar-task` — 5 agentes debatem PRD)
2. **PRD** (`criar-prd` com plano de testes aprovado)
3. **Implementação** (backend-dev no branch)
4. **Auditoria pré-push** (security + qa)
5. **Push** + PR + CI/CD (E0c será responsável por finalizá-lo)
