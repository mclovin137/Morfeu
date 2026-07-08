# plan.md — Plano da Task em Andamento

> Este arquivo é o plano vivo da task corrente **do projeto** — não confundir com o *plan mode* do Claude Code (que grava em `~/.claude/plans/`). Atualizado durante a implementação; reflete o estado real (regras em `roles.md` §6.11).

## Task atual

**Task 0001 — App Skeleton Go** (E0a)
- **Status:** refinada ✓ | PRD ✓ | **implementação completa (8 commits) + correções bloqueadores (6 commits)**
- **Branch:** `feature/0001-app-skeleton-go`
- **PRD:** [`docs/prd/0001-app-skeleton-go.md`](../prd/0001-app-skeleton-go.md) (ativo, 2026-07-08)
- **Roadmap:** E0a (Walking skeleton, subtask a) | Marco M1 (walking skeleton com TLS, CI/CD, observabilidade)
- **Implementação:** 2026-07-08 10:50–14:00 (8 commits iniciais + 6 commits de correção de bloqueadores)

### Objetivo

Estrutura base da aplicação Go (app skeleton) com Docker Compose (PostgreSQL + Redis), sistema de migrations (golang-migrate), e um endpoint de exemplo `GET /filmes` servindo dados do banco com cache read-through — pronta para adicionar lógica de negócio.

### Escopo

1. **Estrutura Go** (`main.go`, handlers, config, logger)
2. **Docker Compose** (PG + Redis)
3. **Migrations** (golang-migrate com seed)
4. **Handler `GET /filmes`** (cache read-through Redis, TTL 5 min)
5. **Integração mínima** (pgx, go-redis, sqlc, error logging)

### Arquivos (estimativa: ~15)

| Arquivo | Tipo | Nota |
|---|---|---|
| `cmd/app/main.go` | novo | entry point |
| `internal/config/config.go` | novo | env config |
| `internal/handler/film.go` | novo | handler GET /filmes |
| `internal/logger/logger.go` | novo | structured logger |
| `internal/cache/redis.go` | novo | cache layer |
| `internal/db/queries.sql` | novo | sqlc queries |
| `internal/db/queries.sql.go` | gerado | sqlc output |
| `internal/db/models.go` | gerado | sqlc output |
| `migrations/001_initial_schema.up.sql` | novo | schema + seed |
| `migrations/001_initial_schema.down.sql` | novo | rollback |
| `docker-compose.yml` | novo | PG + Redis |
| `.env.example` | novo | env template |
| `Dockerfile` | novo | multi-stage build |
| `go.mod` | update | add deps |
| `go.sum` | update | lock versions |
| `Makefile` | novo | compose/migrate/run targets |

### Dependências

- `github.com/labstack/echo/v4` (Echo web framework)
- `github.com/jackc/pgx/v5` (PG driver)
- `github.com/sqlc-dev/sqlc` (SQL compiler)
- `github.com/redis/go-redis/v9` (Redis client)
- `github.com/golang-migrate/migrate/v4` (migrations)
- `go.uber.org/zap` (logging)

### Critérios de aceite

- [x] CA01 — App compila sem erros (`go build ./cmd/app`)
- [x] CA02 — `docker-compose up` levanta PG + Redis em ~5s com healthchecks
- [x] CA03 — Migrations rodam: `make migrate` popula `films` com 10 registros, seed aplicada
- [x] CA04 — GET /filmes retorna JSON array com 10 filmes, status 200, Content-Type application/json
- [x] CA05 — Cache hit/miss validado: primeira requisição db, segunda cache com logs
- [x] CA06 — Redis graceful degradation: se Redis cair, fallback ao PG (200 com log warning)
- [x] CA07 — Logger estruturado: JSON (timestamps RFC3339, trace_id propagado, sem secrets)
- [x] CA08 — Código passa lint (`golangci-lint run` rules em .golangci.yml)
- [x] CA09 — Migrations idempotentes: `down; up` sem erro
- [x] CA10 — .env.example presente, .env gitignored
- [x] CA11 — GET /health endpoint: `{ "status": "ok", "db": "ok|error", "redis": "ok|error" }`
- [x] CA12 — README da task: setup, como rodar, diagrama, troubleshoot

## Implementação (Commits)

**Fase 1: Implementação Inicial (8 commits, ~1,660 LOC)**

| # | Commit | Descrição | LOC aprox |
|---|--------|-----------|----------|
| 1 | 4f491d3 | Init: go.mod, docker-compose, Dockerfile, main.go stub, Makefile | 150 |
| 2 | 268af8f | Config + Logger: env parsing, zap structured JSON logger | 200 |
| 3 | 28ea76c | Migrations: 001_initial_schema (CREATE TABLE films + 10 filmes seed) | 30 |
| 4 | d7aaf01 | sqlc: queries.sql, models.go, queries.sql.go, config yaml | 170 |
| 5 | 74b1471 | Handlers: film.go (GET /filmes), health.go (GET /health), testes | 220 |
| 6 | 7cc4a19 | Service + Cache: film.go service, redis.go cache layer, composição root | 400 |
| 7 | 69b69eb | Integration tests: testcontainers for PG, database tests | 160 |
| 8 | 7261c74 | Docs: .golangci.yml lint rules, README.md task, troubleshoot | 330 |

**Fase 2: Correção de Bloqueadores (6 commits, ~540 LOC)**

| # | Commit | Descrição | Bloqueadores |
|---|--------|-----------|--------------|
| 9 | 524a292 | Security: docker-compose sem credenciais, .env.docker-compose | #7 |
| 10 | 3e6f727 | Security: golang-jwt v5.2.2, lib.md atualizado | #8, #9 |
| 11 | a244f98 | QA: migrations sem ON CONFLICT (idempotência real) | #6 |
| 12 | 2957a13 | QA: 5 scenarios de integration tests (up/down/up, cache, graceful) | #1-3, #5-6 |
| 13 | 6f2fefc | QA: 3 scenarios de health endpoint tests (all ok, redis down, db down) | #4 |
| 14 | 007882e | QA: 3 scenarios de cache hit/miss tests (miss, hit, TTL) | #1 |

**Total** | | **14 commits, ~2,200 LOC (dentro do limite de 30 arquivos)** | |

### Riscos

| Risco | Probabilidade | Impacto | Mitigação |
|---|---|---|---|
| Cache miss/inconsistência | Médio | Médio | Testes unitários + TTL 5 min |
| Docker networking issues | Baixo | Alto | docker-compose test em CI |
| sqlc code generation issues | Baixo | Médio | Validar schema primeiro |
| Go modules conflict | Baixo | Médio | Context7 + lib.md check |
| Port conflicts | Médio | Baixo | compose service names |

### Impacto estimado

- **Código**: Médio (novo, foundation, sem lógica complexa)
- **Banco**: Baixo (schema trivial, seed fake)
- **Infra**: Médio (Docker Compose, networking)
- **Usuários**: N/A (dev-only)
- **Testes**: Médio (integração obrigatória para cache)

## Auditorias

- **2026-07-08 — APROVADA** (governança de economia de tokens, branch `chore/governanca-economia-tokens`, 4 commits): primeiro passe no **modelo híbrido novo** (§6.4 — revisor único de julgamento, `security` em Sonnet/medium). Mecânicos na sessão principal: 29 ≤ 30 arquivos, árvore limpa, grep de secrets no diff limpo (gitleaks ausente — instalar segue pendente), sem código Go/deps/migrations (itens 6, 10–12 N/A). Julgamento ✅ em todos: escopo = exatamente os 7 pontos autorizados pelo usuário (§8); sem conflito com ADRs 0001–0006; sem overengineering (mudança REDUZ cerimônia); hook alterado só com `--model haiku` (`--dangerously-skip-permissions` preexistente, pendência conhecida do state.md, sem regressão); settings sem secrets; consistência roles.md × CLAUDE.md × skills × agentes íntegra; numeração §4.4/§6.4.1–6/§6.14.1–8/§6.15.1–8 sem gaps. Sem correções obrigatórias. Push liberado.
- **2026-07-08 — Correção de Bloqueadores (reauditoria pendente)** (commits `524a292`+`3e6f727`+`a244f98`+`2957a13`+`6f2fefc`+`007882e`): Resolvidos 9 bloqueadores pré-push:
  - **Security** (3 itens): (7) docker-compose sem credenciais versionadas → .env.docker-compose; (8) golang-jwt v5.2.2+ atualizado em go.mod; (9) lib.md com zap, uuid, testcontainers registrados.
  - **QA** (6 itens): (1) cache hit/miss com Redis real (testcontainers) → `internal/integration_cache_test.go`; (2) migrations up/down/up idempotente → `TestMigrationsUpDownUpIdempotent`; (3) graceful degradation Redis → `TestGracefulDegradationRedisUnavailable`; (4) health endpoint tests → `internal/handler/health_test.go`; (5) handler→service e2e → `TestListFilmsE2E_FullStack`; (6) containers efêmeros + ON CONFLICT removido de migrations.
  - Todos os testes usam testcontainers reais (não mocks); containers por suite com `defer terminate()` para limpeza; 14 commits no total (8 iniciais + 6 correção).

- **2026-07-07 — APROVADA** (docs/lib + vendorização de skills Go, commits `019b664`+`46b7e61`+fix): 13 itens; diff de 49 arquivos justificado (§6.3 rege tasks de implementação; 39/49 são conteúdo de terceiros verbatim — superfície autoral ~13 arquivos; precedente do template com 28 skills). Security ✅ (item 8: zero secrets/PII em greps independentes, docs/lib com placeholders corretos; item 9: N/A + skills de terceiros sem conteúdo malicioso/telemetria/hooks embutidos, allowed-tools escopado; item 11: N/A — markdown puro, lib.md consistente arquivo a arquivo; extra: 4 achados pré-existentes do AgentShield no hook Obsidian avaliados como não bloqueantes → pendência registrada no state.md). QA: 1ª rodada REPROVADO por referência morta a `jpa-patterns` no README.md:58 → corrigida (lista duplicada substituída por referência a roles.md §4.2) + re-grep limpo → **APROVADO** (itens 6/7 N/A-justificados; consistência §4.2/lib.md/docs-lib/§6.9.4/§7/state.md verificada). security-scan (AgentShield) pós-vendorização: nenhum achado nas skills novas.
- **2026-07-07 — APROVADA** (registro da identidade visual, commit `93104b5`): 13 itens verificados; diff de 6 arquivos, 100% documentação (`docs/design/` novo + ponteiros em roles.md/doc.md/state.md). Security ✅ (item 8: zero secrets — 4 blobs base64 validados como woff2 genuínos por magic bytes, zero alta-entropia residual, PII só placeholders; item 9: N/A-justificado + JS do protótipo verificado sem rede/console/storage; item 11: nenhuma dependência nova — fontes embutidas são asset estático; pendência `@fontsource/*`→lib.md formalizada p/ quando a SPA nascer). QA ✅ (itens 6/7 N/A-justificados — sem código de produção no repo; consistência das 5 referências cruzadas confirmada; observação não bloqueante: notação "E0c" vs. E0 item (c) do roadmap). Itens 1–5, 10, 12, 13 na sessão principal: sem PRD/task por ser artefato de governança-design pré-task (mesmo precedente do bootstrap), escopo coerente, 6 ≤ 30 arquivos, lib.md intacto, sem schema, state.md/plan.md atualizados.
- **2026-07-07 — APROVADA** (bootstrap de governança, pré-push inicial): 13 itens verificados; security ✅ (árvore publicável + histórico sem secrets; CVEs do lib.md revalidadas no OSV), qa ✅ (itens 6/7 N/A-justificados — diff 100% documentação; consistência ADR 0006 ↔ doc.md ↔ roadmap confirmada; divergência lib.md "3–5→2–4 jornadas" corrigida no ato).
