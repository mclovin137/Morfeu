# PRD 0003 — Conformidade package-by-domain (pré-E0b)

- **Task:** docs/tasks/0003-conformidade-package-by-domain.md
- **Branch:** `refactor/0003-conformidade-package-by-domain`
- **Data:** 2026-07-10
- **Status:** ativo

> Consome as exigências de `docs/refinamentos/E0-walking-skeleton.md` §"Task de conformidade" (épico E0 refinado em 2026-07-09 — sem nova rodada de agentes, roles.md §6.2.5) + exigência da auditoria de 2026-07-10 (teste HTTP real de `GET /filmes`).

## Objetivo

Eliminar o drift package-by-layer da E0a (pacotes `internal/handler`, `internal/service`, `internal/db`) realinhando ao layout package-by-domain do ADR 0003, antes que a E0b (que nasce package-by-domain) o amplifique. **Zero mudança de comportamento** — nenhum contrato HTTP, schema ou regra de negócio muda.

## Escopo

1. **Módulo `catalogo`** em `internal/catalogo/` no layout do ADR 0003: `handler.go`, `service.go`, `queries.sql`, `db/` (gerado pelo sqlc), testes junto ao código. `errors.go` **não** nasce agora — não existe erro sentinela de domínio na E0a (arquivo vazio seria overengineering); nasce com o primeiro erro real (E0b+).
2. **Composition root único** em `cmd/morfeu/` (era `cmd/app/`) — wiring manual preservado (ADR 0003: Resolver = `main.go`).
3. **Plataforma explícita**: `internal/config`, `internal/logger`, `internal/cache` permanecem como pacotes de plataforma no nível `internal/` (já explícitos); health/readiness — endpoint de plataforma, não de domínio — move de `internal/handler/` para `internal/health/`.
4. **`depguard` ativo** no `.golangci.yml`: (a) módulo não importa módulo (hoje só `catalogo` — regra nasce pronta p/ E0b); (b) domínio não importa driver de infra direto (`pgx`, `go-redis` só em plataforma/`db` gerado); rodado via container (sem golangci-lint local; enforcement contínuo chega no E0c-CI).
5. **`sqlc.yaml`** com geração por módulo (`internal/catalogo/queries.sql` → `internal/catalogo/db/`); regeneração sem edição manual do gerado.
6. **Makefile e Dockerfile** apontando para `./cmd/morfeu`.
7. **Exigência da auditoria 2026-07-10**: teste de integração HTTP real de `GET /filmes` no padrão de `health_test.go` (testcontainers reais + `httptest` + echo real), cobrindo RF03/CA04 do PRD 0001 (status 200, `Content-Type: application/json`, corpo JSON válido); remoção do comentário em `film_test.go` que afirma cobertura inexistente.

## Fora de escopo

- Feature nova, endpoint novo, mudança de contrato ou de schema (nenhuma migration).
- Outbox/RabbitMQ (E0b/task 0002), CI (E0c-CI), deploy (E0c-CD), observabilidade (E0d).
- `internal/plataforma/` como pacote-pai — reorganização adicional sem exigência no refinamento.
- Repository manual, aggregates, orquestrador (ADR 0005 — critérios não atendidos no catálogo).
- Instalação de golangci-lint no ambiente (roda via container nesta task).

## Requisitos funcionais

- RF01 — O binário compila e serve exatamente os mesmos endpoints da E0a (`GET /filmes`, health/readiness) com as mesmas respostas.
- RF02 — `GET /filmes` tem teste de integração no nível HTTP real (handler echo de verdade, não chamada direta ao service).

## Requisitos não funcionais

- RNF01 — Layout conforme ADR 0003 (tabela de camadas + layout padrão de módulo).
- RNF02 — `depguard` verde com as duas regras de fronteira do ADR 0003 (§Decisão, regra 4).
- RNF03 — Código gerado pelo sqlc reproduzível: `sqlc generate` não produz diff além do movido.
- RNF04 — Suíte completa verde com `-race` (via container `golang:1.25`, padrão do ambiente registrado no state.md).

## Regras de negócio

- RN01 — Nenhuma (task estrutural; regras de negócio existentes movem-se intactas).

## Critérios de aceite

- [ ] CA01 — **Suíte da E0a passa sem alteração de asserts** (critério único do refinamento): nenhum assert existente modificado; apenas paths/imports de teste ajustados e testes novos adicionados.
- [ ] CA02 — `internal/handler`, `internal/service`, `internal/db` não existem mais; `internal/catalogo/{handler.go,service.go,queries.sql,db/}` e `internal/health/` existem; `cmd/morfeu` compila (`go build ./...`).
- [ ] CA03 — `golangci-lint run` (container) verde com `depguard` ativo; violação proposital de import (teste manual durante o dev) é bloqueada.
- [ ] CA04 — `sqlc generate` após o ajuste do `sqlc.yaml` regenera `internal/catalogo/db/` idêntico ao commitado.
- [ ] CA05 — Novo teste HTTP real de `GET /filmes` verde: 200 + `Content-Type: application/json` + corpo decodifica na lista de filmes; comentário enganoso de `film_test.go` removido.
- [ ] CA06 — `docker build` funciona com o novo path (`cmd/morfeu`).

## Plano de testes

| Cenário | Tipo | Cobre |
|---|---|---|
| Suíte E0a completa pós-movimentação (unit + integração testcontainers + `-race`) | regressão | CA01, RF01 |
| `GET /filmes` via echo real + PG/Redis testcontainers: 200, content-type, corpo JSON (feliz) | integração | CA05, RF02 |
| `GET /filmes` HTTP com Redis indisponível (graceful degradation preservada no nível HTTP) | integração | RF01 (borda — reusa cenário existente `TestGracefulDegradationRedisUnavailable`, sem alterar asserts) |
| Import proibido (domínio→pgx driver) injetado localmente → depguard falha | manual dev | CA03 (erro) |
| `sqlc generate` diff vazio | mecânico | CA04 |

## Plano de implementação

1. `git mv` em blocos, validando `go build ./...` + suíte a cada bloco: (a) `cmd/app`→`cmd/morfeu`; (b) `internal/handler/health*`→`internal/health/`; (c) `internal/handler/film*` + `internal/service/*`→`internal/catalogo/` (renomear `film.go`→`handler.go`/`service.go`, ajustar packages); (d) `internal/db/queries.sql`→`internal/catalogo/queries.sql` + `sqlc.yaml` + regenerar em `internal/catalogo/db/`.
2. Ajustar imports nos testes de integração da raiz `internal/`.
3. Makefile + Dockerfile (`./cmd/morfeu`).
4. `depguard` no `.golangci.yml`; rodar via container; testar bloqueio com violação proposital (não commitada).
5. Teste HTTP real de `GET /filmes` em `internal/catalogo/handler_test.go` (padrão `health_test.go`); remover comentário enganoso.
6. Suíte completa com `-race`; atualizar `plan.md`/`state.md`; auditoria; PR.

## Arquivos que serão criados

- `internal/catalogo/handler.go` — ex-`internal/handler/film.go` (package `catalogo`)
- `internal/catalogo/handler_test.go` — ex-`internal/handler/film_test.go` + novo teste HTTP real de `GET /filmes` (CA05)
- `internal/catalogo/service.go` — ex-`internal/service/film.go`
- `internal/catalogo/service_test.go` — ex-`internal/service/film_test.go`
- `internal/catalogo/queries.sql` — ex-`internal/db/queries.sql`
- `internal/catalogo/db/models.go`, `internal/catalogo/db/queries.sql.go` — gerados pelo sqlc no path novo
- `internal/health/health.go`, `internal/health/health_test.go` — ex-`internal/handler/health*.go`
- `cmd/morfeu/main.go` — ex-`cmd/app/main.go` (imports atualizados)
- `docs/prd/0003-conformidade-package-by-domain.md` — este PRD

## Arquivos que serão modificados

- `sqlc.yaml` — queries/out por módulo (`internal/catalogo/`)
- `.golangci.yml` — depguard (2 regras de fronteira do ADR 0003)
- `Makefile` — build `./cmd/morfeu`
- `Dockerfile` — build `./cmd/morfeu`
- `internal/integration_test.go`, `internal/integration_cache_test.go` — imports novos (asserts intactos)
- `plan.md`, `state.md`, `docs/tasks/0003-*.md`, `docs/tasks/README.md` — controle

(Removidos por movimentação: `cmd/app/`, `internal/handler/`, `internal/service/`, `internal/db/`.)

**Total estimado: ~21 arquivos tocados (9 criados/movidos-destino + 12 modificados/controle) ≤ 30.** Acima da estimativa de 12–16 da task porque o diff conta origem+destino das movimentações e os arquivos de controle; a superfície autoral real é pequena (imports, packages, 1 teste novo, 2 configs).

## Skills de apoio (roles.md §4.4 — carga seletiva, máx. 3)

- `golang-project-layout` — convenções cmd/internal e reorganização de pacotes
- `golang-lint` — configuração do depguard no `.golangci.yml`
- `golang-testing` — teste de integração HTTP (httptest + testcontainers)

## Dependências utilizadas

Nenhuma nova (`lib.md` inalterado). `depguard` é linter embutido no golangci-lint (ferramenta, não dependência do módulo); golangci-lint roda via container.

## Nota de implementação (2026-07-10 — desvios descobertos, regra de manutenção do PRD)

Pré-condições descobertas durante a implementação — nenhuma é feature nova nem muda comportamento observável; todas eram necessárias para os próprios critérios de aceite (CA03/CA04/CA06). Mesmo padrão do fix emergencial: configs que nunca tinham sido executadas de verdade.

1. **`sqlc.yaml` estava em formato v1 inválido** (nunca rodou; o "gerado" anterior era manual) → migrado para `version: "2"`. Sem isso CA04 era inalcançável.
2. **`.golangci.yml` idem** — rejeitado pelo golangci-lint 2.12.2 → migrado para config v2; depguard com regra única `catalogo-domain` (allowlist `strict`) cobrindo as duas fronteiras do ADR 0003; `_test.go` excluídos (ADR 0006 usa testcontainers reais).
3. **Dockerfile** com `golang:1.21-alpine`, incompatível com `go.mod go 1.25.0` (`docker build` já falhava antes da task) → `golang:1.25-alpine` (necessário p/ CA06).
4. **`NewFilmService(*pgxpool.Pool)` → `NewFilmService(*db.Queries)`** — sem isso o depguard baniria a construção do próprio service (domínio recebendo driver cru); wiring ajustado em `cmd/morfeu/main.go` e testes de integração (`db.New(pool)`), zero assert alterado.

Arquivos adicionais tocados por esta nota: nenhum além dos já listados (sqlc.yaml, .golangci.yml, Dockerfile, service.go, main.go, testes de integração já previstos).

## Impactos técnicos

- Import paths internos mudam (`internal/handler|service|db` → `internal/catalogo`, `internal/health`); nenhum contrato externo muda.
- Banco: zero (sem migration; queries movem de arquivo, não de conteúdo).
- Infra: Dockerfile/Makefile apontam para `cmd/morfeu`; compose inalterado.
- Prepara a E0b: fronteiras de módulo + depguard prontos antes do outbox.

## Riscos

- Renomeação em massa quebra import silenciosamente → blocos pequenos com `go build ./...` + suíte por bloco.
- sqlc gerar código divergente no path novo → regenerar e comparar com o gerado atual antes de apagar o antigo.
- `.golangci.yml` com depguard mal configurado passa em falso → teste de bloqueio com violação proposital durante o dev (CA03).
- Ambiente sem gcc/lint local → container `golang:1.25` (padrão já registrado) p/ `-race` e golangci-lint.

## Estratégia de rollback

Sem migration e sem mudança de comportamento: rollback = revert do merge (nenhum estado persistente afetado). Durante o dev, cada bloco de movimentação é um commit — `git revert`/`reset` granular.
