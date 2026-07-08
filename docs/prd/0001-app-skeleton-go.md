# PRD 0001 — App Skeleton Go

- **Task:** `docs/tasks/0001-app-skeleton-go.md`
- **Branch:** `feature/0001-app-skeleton-go`
- **Data:** 2026-07-08
- **Status:** ativo (refinamento ✅ aprovado; implementação não iniciada)
- **Refinamento:** [Seção da task consolidada em 2026-07-08 · Pareceres: arquiteto ✓ sre-devops ✓ security ✓ qa ✓ backend-dev ✓]

---

## Objetivo

Implementar a estrutura base da aplicação Go (app skeleton) com Docker Compose (PostgreSQL + Redis), sistema de migrations (golang-migrate), e um endpoint de exemplo `GET /filmes` servindo dados do banco com cache read-through — pronta para adicionar lógica de negócio nos épicos E1–E6.

**Valor entregue:** Foundation reproducível localmente, camadas handler→service→sqlc consolidadas, padrões de logger/config/cache estabelecidos conforme ADRs 0001–0006. Bloqueia Marco M1 (URL pública viva em TLS — entra em E0c).

---

## Escopo

### Estrutura Go (cmd + internal)

1. **`cmd/app/main.go`** — Entry point, composition root (ADR 0001):
   - Lê config via env (DATABASE_URL, REDIS_URL, LOG_LEVEL, APP_PORT)
   - Inicia logger (zap estruturado JSON)
   - Cria pools: PG (pgx) + Redis (go-redis)
   - Roota golang-migrate e executa migrations (síncronas, antes de listener HTTP)
   - Monta router Echo com middlewares (logger, recovery)
   - Inicia HTTP server com graceful shutdown (timeout 30s)

2. **`internal/config/config.go`** — Config from env (ADR 0004):
   - Struct com campos: DatabaseURL, RedisURL, LogLevel, AppPort, CacheTTL, etc.
   - Defaults sensatos (porta 8080, cache TTL 5 min, log info)
   - Validação mínima (não vazia, formatos)

3. **`internal/logger/logger.go`** — Structured logger (ADR 0004):
   - Wrapper zap com campos: timestamp (RFC3339), level, message, trace_id, error
   - Não expõe secrets (mascara DATABASE_URL, REDIS_URL)
   - Exporta *zap.Logger para injetar em handlers/services

4. **`internal/handler/film.go`** — Handler GET /filmes:
   - Recebe echo.Context, extrai trace_id (ou gera novo)
   - Chama service.ListFilms(ctx)
   - Retorna JSON + headers (Content-Type, Cache-Control para cache hit)
   - Trata erro (500 com log estruturado, não panic)

5. **`internal/handler/health.go`** — Health endpoint:
   - GET /health → 200 com `{ "status": "ok", "db": "ok|error", "redis": "ok|error" }`
   - Valida PG connection (SELECT 1), Redis PING
   - Não é critério de aceite, mas essencial para M1

6. **`internal/service/film.go`** — Service layer (ADR 0003):
   - `ListFilms(ctx context.Context) ([]Film, error)` — orquestra cache + db
   - Recebe cache + db como dependências (injeção no construtor)
   - Fallback: Redis timeout (~500ms) → query direto ao PG (sem retry, log warning)

7. **`internal/cache/redis.go`** — Cache layer (read-through):
   - `Get(ctx, key string) ([]byte, error)` — busca Redis
   - `Set(ctx, key string, value []byte, ttl time.Duration) error`
   - Timeout ~500ms; timeout → fallback, sem panic
   - Evento loggado: `{ "action": "cache_miss", "duration_ms": ... }` ou `cache_hit`

8. **`internal/db/`** — SQL + sqlc (ADR 0002):
   - `queries.sql` — Query `ListFilms`: `SELECT id, title, year, runtime, synopsis, imdb_id, poster_url, created_at FROM films ORDER BY created_at DESC LIMIT 100`
   - `sqlc.yaml` — Config de geração (versão 1.x, target Go)
   - Gerados: `models.go`, `queries.sql.go` (via `make gen`)

9. **`migrations/`** — golang-migrate (ADR 0006):
   - `001_initial_schema.up.sql` — CREATE TABLE films + seed (10 filmes fake)
   - `001_initial_schema.down.sql` — DROP TABLE IF EXISTS films

### Infraestrutura local

10. **`docker-compose.yml`**:
    - PostgreSQL 16 (volume `morfeu-postgres`, env POSTGRES_PASSWORD, healthcheck pg_isready)
    - Redis 7 (volume `morfeu-redis`, healthcheck redis-cli PING)
    - Network bridge padrão
    - Services: `postgres`, `redis` (nomes para DNS interno)

11. **`.env.example`**:
    ```
    DATABASE_URL=postgres://postgres:postgres@localhost:5432/morfeu
    REDIS_URL=redis://localhost:6379/0
    LOG_LEVEL=info
    APP_PORT=8080
    CACHE_TTL_SECONDS=300
    ```

12. **`Dockerfile`** (multi-stage):
    - Builder: `go build -o app ./cmd/app`
    - Runtime: scratch (sem base; app estático)
    - EXPOSE 8080

13. **`Makefile`**:
    - `make up` — docker-compose up -d
    - `make down` — docker-compose down
    - `make gen` — sqlc generate
    - `make migrate` — migrate -path ./migrations -database $DATABASE_URL up
    - `make lint` — golangci-lint run
    - `make test` — go test -race -cover ./...
    - `make build` — go build -o app ./cmd/app
    - `make run` — ./app

14. **`go.mod` / `go.sum`**:
    - Dependências: github.com/labstack/echo/v4 ≥4.15.0, jackc/pgx/v5 ≥5.9.2, redis/go-redis/v9 ≥9.7.3, golang-migrate/migrate/v4 latest, go.uber.org/zap latest
    - Versionamento per lib.md

### Configuração

15. **`.gitignore`** (atualizar):
    - Adicionar: `.env` (não `.env.example`), `app` (binário), `*.log`

16. **`README.md` da task** (novo ou seção em `README.md` raiz):
    - Setup (clone, `docker-compose up`, `make migrate`)
    - Como rodar: `make run` (ou `./app` direto)
    - Testar: `curl http://localhost:8080/filmes`
    - Diagnosticar: `make lint`, `make test`, `docker-compose logs postgres`
    - Diagrama: ASCII ou Mermaid com estrutura (handler → service → cache → db)

---

## Fora de escopo

- **Autenticação / JWT** (E2 — Identidade)
- **Input validation / sanitização** (E3 — Sessões)
- **CI/CD pipeline** (E0c — Walking skeleton)
- **RabbitMQ / outbox** (E0b — Walking skeleton)
- **Observabilidade profunda** — métricas, traces OTel (E0d — Observabilidade)
- **TLS / HTTPS** (E0c ou pré-deploy hardening)
- **Rate limiting** (E3+)
- **Pagination / filtering** (E3+)
- **HTTPErrorHandler centralizado** (E1 — com autenticação)

---

## Requisitos Funcionais

- **RF01** — App Go inicializa com config via env, sem secrets hardcoded
- **RF02** — Migrations rodam síncronas no startup (antes de listener HTTP)
- **RF03** — GET /filmes retorna JSON array com ~10 filmes, status 200
- **RF04** — Dados são lidos do PG na primeira requisição (cache miss)
- **RF05** — Dados são servidos do Redis na segunda requisição dentro de TTL (cache hit)
- **RF06** — Se Redis cair/timeout, fallback síncrono ao PG (graceful degradation)
- **RF07** — GET /health retorna `{ "status": "ok", "db": "ok", "redis": "ok" }` ou status parcial se serviço cair

---

## Requisitos Não Funcionais

- **RNF01** — Logger estruturado JSON, sem hardcoding de secrets, com trace_id propagado
- **RNF02** — PG pool: min 5, max 25 conexões, timeout 30s; configurável via env
- **RNF03** — Redis timeout ~500ms; fallback não bloqueia
- **RNF04** — Migrations idempotentes (up, down, up sem erro)
- **RNF05** — App compila sem warnings (`golangci-lint run` limpo)
- **RNF06** — Dockerfile multi-stage, binário estático, imagem < 50MB
- **RNF07** — Docker Compose levanta stack em ~5s com healthchecks validados
- **RNF08** — Código segue ADRs 0001–0006 (Echo router, sqlc queries, handler→service→db, padrões Go, log estruturado)

---

## Regras de Negócio

- **RN01** — Catálogo é somente leitura em E0a (GET /filmes); updates entram em E2 (CRUD backoffice)
- **RN02** — Seed é determinístico (10 filmes reais TMDB com dados fake) — testes reproduzíveis
- **RN03** — Cache TTL = 5 min; não há invalidação manual (refresh = expiração natural)

---

## Critérios de Aceite

Incorporando refinamento multiagente (2026-07-08):

- [ ] **CA01** — App compila sem erros (`go build ./cmd/app`)
- [ ] **CA02** — `docker-compose up` levanta PG + Redis + app em ~5s; healthchecks passam
- [ ] **CA03** — Migrations rodam: `make migrate` popula `films` com 10 registros, seed aplicada
- [ ] **CA04** — GET /filmes retorna JSON array com 10 filmes, status 200, Content-Type application/json
- [ ] **CA05** — Cache hit/miss validado: primeira requisição é lenta (db + log `"action":"cache_miss"`), segunda é rápida (cache + log `"action":"cache_hit"` com TTL header)
- [ ] **CA06** — Redis graceful degradation: `docker stop morfeu-redis` → GET /filmes retorna 200 com dados do banco (log `"cache_unavailable":true`)
- [ ] **CA07** — Logger estruturado: todos os eventos saem em JSON (timestamps RFC3339, trace_id propagado, nenhum secret exposto)
- [ ] **CA08** — Código passa lint (`golangci-lint run` sem warnings)
- [ ] **CA09** — Migrations idempotentes: `make migrate down; make migrate up` sem erro (rollback limpo)
- [ ] **CA10** — .env.example presente, .env gitignored
- [ ] **CA11** — GET /health endpoint retorna `{ "status": "ok", "db": "ok|error", "redis": "ok|error" }`
- [ ] **CA12** — README da task documenta: setup, como rodar, como testar, diagrama, troubleshoot

---

## Plano de Testes

Incorporando refinamento QA (ADR 0006 + roles.md §6.7):

### Testes Unitários

| Cenário | Cobertura | Validação |
|---|---|---|
| **Config parsing** | `internal/config` | Defaults, env override, validação (não vazio, formatos) |
| **Logger initialization** | `internal/logger` | zap config, JSON output, trace_id field |
| **Cache Get/Set** | `internal/cache` (mock clock) | TTL expiração, timeout handling, mask secrets em logs |
| **Handler /filmes** | `internal/handler` (mock service) | HTTP 200, JSON válido, trace_id propagado |

### Testes de Integração

| Cenário | Ferramenta | Validação |
|---|---|---|
| **Migrations up/down/up** | testcontainers-go + postgres | Idempotência, schema correto, seed após up |
| **Cache hit/miss** | testcontainers-go + postgres + redis | Primeira req = db, segunda = cache, TTL respeitado |
| **Graceful degradation Redis** | testcontainers (stop redis mid-test) | Fallback ao PG, log warning, status 200 |
| **Handler → Service → Cache → DB** | testcontainers stack real | End-to-end com dados reais, latência logging |
| **Health endpoint** | testcontainers stack real | Valida conexões PG + Redis, status parcial se um cai |

### CI Gate

- [ ] `go build ./cmd/app` sem erro
- [ ] `go test -race ./...` sem race conditions
- [ ] `golangci-lint run` sem warnings
- [ ] `sqlc vet` sem erros
- [ ] `migrations/` job: up → down → up (se PR tocar migrations)

**Cobertura alvo:** Informativa (60%) para E0a. Gate 80% ativa em E2+.

---

## Plano de Implementação

**Sequência (7–8 commits esperados):**

1. **Init + Docker Compose** (`cmd/main.go` stub, `docker-compose.yml`, `.env.example`)
2. **Config + Logger** (`internal/config`, `internal/logger`)
3. **Migrations + schema** (`migrations/001_*`, seed data)
4. **sqlc queries + generated** (`internal/db/queries.sql`, `make gen`, `go.mod` atualizado)
5. **Handlers** (`internal/handler/film.go`, `internal/handler/health.go`)
6. **Service + Cache** (`internal/service`, `internal/cache`, composição em `main.go`)
7. **Testes** (`*_test.go` unit + integration, testcontainers)
8. **Docs + Lint** (`README.md` da task, `golangci-lint run`, commits cleanup)

**Ordem:** Cada fase é revisável, sem dependencies bloquantes. Fase 1-3 = infra; 4-5 = API surface; 6-7 = lógica; 8 = qualidade.

---

## Arquivos que serão criados

| Arquivo | Linhas aprox. | Propósito |
|---|---|---|
| `cmd/app/main.go` | 80–100 | Entry point, composition root |
| `internal/config/config.go` | 30–50 | Config from env |
| `internal/logger/logger.go` | 40–60 | zap wrapper |
| `internal/handler/film.go` | 30–50 | GET /filmes handler |
| `internal/handler/health.go` | 30–40 | GET /health handler |
| `internal/service/film.go` | 50–80 | ListFilms service |
| `internal/cache/redis.go` | 50–80 | Cache layer (Get/Set) |
| `internal/db/queries.sql` | 20–30 | sqlc query definition |
| `internal/db/models.go` | 50–80 | sqlc generated (types) |
| `internal/db/queries.sql.go` | 100–150 | sqlc generated (queries) |
| `migrations/001_initial_schema.up.sql` | 30–50 | Schema + seed (10 filmes) |
| `migrations/001_initial_schema.down.sql` | 5–10 | Rollback |
| `docker-compose.yml` | 40–60 | PG + Redis + healthchecks |
| `.env.example` | 5–10 | Env template |
| `Dockerfile` | 15–25 | Multi-stage build |
| `Makefile` | 30–50 | Dev targets |
| `README.md` (task) | 80–120 | Setup + diagrams |
| `*_test.go` (unit + integration) | 200–300 | Testes (testcontainers) |

**Total:** ~18–20 arquivos novos, ~1200–1500 LOC (dentro do limite 30 arquivos).

---

## Arquivos que serão modificados

| Arquivo | Mudança |
|---|---|
| `go.mod` | Adicionar deps: echo, pgx, sqlc, redis, golang-migrate, zap (versões per lib.md) |
| `go.sum` | Lock versions |
| `.gitignore` | Adicionar: `.env`, `app`, `*.log` |
| `roles.md` | Nenhuma (não precisa) |
| `lib.md` | Nenhuma (deps já registradas em 2026-07-07) |

---

## Dependências Utilizadas

**Todas validadas em `lib.md` (2026-07-07):**

| Dependência | Versão | CVE | Justificativa |
|---|---|---|---|
| `github.com/labstack/echo/v4` | ≥4.15.0 | Nenhuma aberta | HTTP framework (ADR 0001) |
| `github.com/jackc/pgx/v5` | ≥5.9.2 | CVE-2026-41889 ✅ fixed | PG driver (ADR 0002) |
| `github.com/sqlc-dev/sqlc` | latest | Nenhuma | SQL compiler, dev tool |
| `github.com/redis/go-redis/v9` | ≥9.7.3 | CVE-2025-29923 ✅ fixed | Redis client |
| `github.com/golang-migrate/migrate/v4` | latest | Nenhuma | Migration tool |
| `go.uber.org/zap` | latest | Nenhuma | Structured logger (ADR 0004) |

**Nenhuma dependência nova.** Todas em `lib.md` § Dependências da Stack.

---

## Impactos Técnicos

### Módulos / Contratos

- **Handler interface** (echo.Context → error): padrão estabelecido para E1–E6 (todos handlers seguem)
- **Service pattern** (ctx, dependências → resultado): base para E2–E9 (serviços com lógica)
- **Cache layer** (Get/Set com fallback): modelo reutilizável em E6 (saga + eventos)

### Banco de dados

- **Schema `films` versão 1** (001): próxima migration será 002 em E2 (CRUD) ou E4 (salas + sessões)
- **Idempotência esperada:** migrations sempre `IF NOT EXISTS` / `IF EXISTS` para segurança em CI

### Infraestrutura local

- **Docker Compose padrão:** será template em E0c (CI/CD) e E11 (hardening); staging/prod descartarão compose

### Observabilidade

- **Logger JSON**: base para Loki (E0d); trace_id propagação prepare ELK stack
- **Health endpoint**: monitora PG + Redis; base para UptimeRobot / healthchecks.io (E0d)

### CI/CD (posterioridade)

- **Migrations em CI:** job criado em E0c (pré-req: `DATABASE_URL` injectable)
- **Lint + test gate:** ativa em E0c; baseline cobertura (60%) informativa agora, gate 80% em E2

---

## Riscos

| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| **Cache miss/inconsistência** | Média | Médio | Testes unitários de hit/miss via testcontainers; TTL simples (5 min) sem lógica complexa |
| **Docker networking (PG/Redis não resolvem)** | Baixa | Alto | Healthchecks no compose validam conectividade; dev testa `docker-compose up` antes do PR |
| **sqlc code generation mismatch** | Baixa | Médio | Validar schema antes de `sqlc generate`; skill postgres-patterns valida em review |
| **Go modules conflito** | Baixa | Médio | Context7 consulted antes de `go get`; lib.md versões mínimas respeitadas |
| **Port/volume conflicts em localhost** | Média | Baixo | Docker service names (postgres, redis) não dependem de localhost; 8080 é dev-only |
| **Migrations não rollback** | Baixa | Alto | Teste obrigatório up → down → up em CI; skill golang-database valida sintaxe |
| **Secrets expostos em logs** | Baixa | Alto | Logger mascara env vars; revisão de código em auditoria (skill security-review) |

---

## Estratégia de Rollback

**Task 0001 é infraestrutura — rollback é simples:**

1. **Code:** Revert branch (git reset) — zero schema lock
2. **DB:** Migrations são idempotentes; `migrate down` limpa tabela sem perda (é seed)
3. **Infra:** `docker-compose down -v` remove volumes; `docker-compose up` recria do zero
4. **Se produção:** Caddy (E0c) roteia tráfego; rollback = deploy versão anterior (imagem GHCR anterior)

**Sem downtime esperado em E0a** (é local dev). Pré-deploy (E0c) tem blue-green later.

---

## Definições & Esclarecimentos

### Context Propagation Pattern

```go
// Em handler
traceID := c.Request().Header.Get("X-Trace-ID")
if traceID == "" {
    traceID = uuid.NewString()
}
ctx := context.WithValue(c.Request().Context(), "trace_id", traceID)

// Service recebe ctx, propaga para logger
svc.ListFilms(ctx) // logger usa ctx.Value("trace_id")
```

### Cache Fallback Behavior

```go
// Em service
films, err := cache.Get(ctx, "films:list")
if err != nil || films == nil {
    films, err = db.ListFilms(ctx) // fallback síncrono
    if err == nil {
        cache.Set(ctx, "films:list", films, 5*time.Minute) // best effort
    }
}
return films, err
```

### Seedagem (10 Filmes Fake)

Dados públicos TMDB com synopsis/poster fake:

```sql
INSERT INTO films (id, title, year, runtime, synopsis, imdb_id, poster_url, created_at) VALUES
(1, 'The Shawshank Redemption', 1994, 142, 'Two imprisoned men bond...', 'tt0111161', 'https://example.com/poster1.jpg', NOW()),
(2, 'The Dark Knight', 2008, 152, 'Batman fights a new threat...', 'tt0468569', 'https://example.com/poster2.jpg', NOW()),
...
(10, 'Inception', 2010, 148, 'A thief steals secrets...', 'tt1375666', 'https://example.com/poster10.jpg', NOW());
```

---

## Referências

- **ADR 0001** — Stack Go + Echo
- **ADR 0002** — sqlc + pgx strategy
- **ADR 0003** — Handler → Service → Data layers
- **ADR 0004** — Go code patterns (logger, config, error handling)
- **ADR 0006** — Test strategy (testcontainers, idempotency)
- **roles.md §6.14** — Refinement requirements
- **docs/roadmap.md** — M1 (walking skeleton) + E0a scope

---

**PRD Status: ATIVO** · Pronto para implementação. Backend-dev segue planejamento acima + ADRs. Auditoria (security + qa) valida contra este PRD antes do push.
