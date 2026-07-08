# plan.md — Plano da Task em Andamento

> Este arquivo é o plano vivo da task corrente **do projeto** — não confundir com o *plan mode* do Claude Code (que grava em `~/.claude/plans/`). Atualizado durante a implementação; reflete o estado real (regras em `roles.md` §6.11).

## Task atual

**Task 0001 — App Skeleton Go** (E0a)
- **Status:** não iniciada → próximo: `/refinar-task`
- **Branch:** `feature/0001-app-skeleton-go`
- **PRD:** pendente (será criado após refinamento + aprovação multiagente)
- **Roadmap:** E0a (Walking skeleton, subtask a)

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

- [ ] App compila sem erros (`go build ./cmd/app`)
- [ ] `docker-compose up` levanta PG + Redis + app em ~5s
- [ ] Migrations rodam e populam tabela `films`
- [ ] `GET http://localhost:8080/filmes` retorna JSON com ~10 filmes, status 200
- [ ] Segunda requisição vem do Redis (verificado via logs)
- [ ] `docker-compose down` para tudo sem erro
- [ ] README.md da task documenta setup, como rodar, diagrama
- [ ] Nenhum secret hardcoded; tudo via `.env` (gitignored)
- [ ] Código passa `golangci-lint run` sem warnings
- [ ] Logger estruturado (JSON) ativo

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

- **2026-07-07 — APROVADA** (docs/lib + vendorização de skills Go, commits `019b664`+`46b7e61`+fix): 13 itens; diff de 49 arquivos justificado (§6.3 rege tasks de implementação; 39/49 são conteúdo de terceiros verbatim — superfície autoral ~13 arquivos; precedente do template com 28 skills). Security ✅ (item 8: zero secrets/PII em greps independentes, docs/lib com placeholders corretos; item 9: N/A + skills de terceiros sem conteúdo malicioso/telemetria/hooks embutidos, allowed-tools escopado; item 11: N/A — markdown puro, lib.md consistente arquivo a arquivo; extra: 4 achados pré-existentes do AgentShield no hook Obsidian avaliados como não bloqueantes → pendência registrada no state.md). QA: 1ª rodada REPROVADO por referência morta a `jpa-patterns` no README.md:58 → corrigida (lista duplicada substituída por referência a roles.md §4.2) + re-grep limpo → **APROVADO** (itens 6/7 N/A-justificados; consistência §4.2/lib.md/docs-lib/§6.9.4/§7/state.md verificada). security-scan (AgentShield) pós-vendorização: nenhum achado nas skills novas.
- **2026-07-07 — APROVADA** (registro da identidade visual, commit `93104b5`): 13 itens verificados; diff de 6 arquivos, 100% documentação (`docs/design/` novo + ponteiros em roles.md/doc.md/state.md). Security ✅ (item 8: zero secrets — 4 blobs base64 validados como woff2 genuínos por magic bytes, zero alta-entropia residual, PII só placeholders; item 9: N/A-justificado + JS do protótipo verificado sem rede/console/storage; item 11: nenhuma dependência nova — fontes embutidas são asset estático; pendência `@fontsource/*`→lib.md formalizada p/ quando a SPA nascer). QA ✅ (itens 6/7 N/A-justificados — sem código de produção no repo; consistência das 5 referências cruzadas confirmada; observação não bloqueante: notação "E0c" vs. E0 item (c) do roadmap). Itens 1–5, 10, 12, 13 na sessão principal: sem PRD/task por ser artefato de governança-design pré-task (mesmo precedente do bootstrap), escopo coerente, 6 ≤ 30 arquivos, lib.md intacto, sem schema, state.md/plan.md atualizados.
- **2026-07-07 — APROVADA** (bootstrap de governança, pré-push inicial): 13 itens verificados; security ✅ (árvore publicável + histórico sem secrets; CVEs do lib.md revalidadas no OSV), qa ✅ (itens 6/7 N/A-justificados — diff 100% documentação; consistência ADR 0006 ↔ doc.md ↔ roadmap confirmada; divergência lib.md "3–5→2–4 jornadas" corrigida no ato).
