# Task 0003 — Conformidade package-by-domain (pré-E0b)

- **Data:** 2026-07-10
- **Status:** implementada — aguardando auditoria
- **Branch:** `refactor/0003-conformidade-package-by-domain`
- **PRD:** [docs/prd/0003-conformidade-package-by-domain.md](../prd/0003-conformidade-package-by-domain.md)
- **Item do roadmap:** E0 (walking skeleton) — correção de drift da E0a contra o ADR 0003; exigências em `docs/refinamentos/E0-walking-skeleton.md` §"Task de conformidade"

> Numeração: **0002 permanece reservada para a E0b** (outbox + RabbitMQ), já referenciada com esse número no refinamento do E0, em `state.md` e `plan.md`. Esta task recebe 0003 para preservar a rastreabilidade dos registros existentes.

## Objetivo

Realinhar a estrutura da E0a ao layout package-by-domain do ADR 0003, eliminando o drift package-by-layer antes que a E0b (que nasce package-by-domain) o amplifique. Zero mudança de comportamento.

## Escopo

- Mover o catálogo para `internal/catalogo/` (layout ADR 0003: `handler.go`, `service.go`, `errors.go`, `queries.sql`, `db/` gerado).
- `cmd/app` → `cmd/morfeu` (composition root único).
- Pacotes de plataforma explícitos (config, logger, cache, db).
- Ativar `depguard` no `.golangci.yml`: módulo não importa módulo; domínio não importa driver de infra direto.
- Ajustar `sqlc.yaml` (paths de geração por módulo) e Makefile.
- **Exigência da auditoria de 2026-07-10** (achado não-bloqueante do fix E0a): teste de integração HTTP real para `GET /filmes` no padrão de `health_test.go` (testcontainers + `httptest` + echo real, cobrindo RF03/CA04 do PRD 0001 — status 200 e `Content-Type: application/json`) e correção do comentário em `internal/handler/film_test.go` que afirma cobertura inexistente.

## Fora de escopo

- Qualquer feature nova, endpoint novo ou mudança de contrato (E0b em diante).
- Outbox/RabbitMQ (task 0002/E0b), CI (E0c-CI), deploy (E0c-CD), observabilidade (E0d).
- Alteração de asserts da suíte existente (critério de aceite exige que passem intactos; exceção: o novo teste HTTP é adição, e o comentário corrigido não é assert).

## Arquivos esperados

~12–16 (≤ 30): movimentação de `internal/handler|service|db/*` → `internal/catalogo/*` (+ ajuste de imports nos testes de integração), `cmd/app/main.go` → `cmd/morfeu/main.go`, `.golangci.yml` (depguard), `sqlc.yaml`, `Makefile`, novo teste HTTP de `GET /filmes`, arquivos de controle (`plan.md`, `state.md`).

## Dependências esperadas

Nenhuma dependência nova (`lib.md` inalterado). `depguard` é linter do golangci-lint, não dependência do módulo.

## Critérios de aceite

- [x] **Suíte da E0a passa sem alteração de asserts** (critério único do refinamento) — unit + integração com testcontainers reais + `-race` (verificado via container `golang:1.25`).
- [x] Layout conforme ADR 0003: `internal/catalogo/` + plataforma explícita (`internal/health`) + `cmd/morfeu`.
- [x] `depguard` ativo e verde (módulo↛módulo; domínio↛driver de infra) — `.golangci.yml` migrado para config v2 (v1 nunca rodou de verdade); testado com violação proposital (bloqueada e revertida).
- [x] `sqlc generate` reproduz o código gerado nos paths novos sem diff manual.
- [x] Novo teste de integração HTTP real de `GET /filmes` verde (status 200 + `Content-Type: application/json`); comentário enganoso removido.
- [x] `docker build` funcional com `cmd/morfeu` (Dockerfile também corrigido: base `golang:1.21-alpine` incompatível com `go.mod go 1.25.0`, pré-existente).

Desvios do escopo original registrados no PRD/plan.md: `sqlc.yaml` estava em formato v1 inválido (nunca rodou), corrigido para v2; `.golangci.yml` idem; `FilmService` passou a depender de `db.Queries` (gerado) em vez do `pgxpool.Pool` cru — pré-requisito da regra depguard de fronteira, sem mudança de comportamento; Dockerfile com base Go desatualizada corrigida. Nenhum assert de teste pré-existente alterado.

## Riscos

- Renomeações em massa quebrarem imports silenciosamente → mitigação: `go build ./...` + suíte completa a cada bloco de movimentação.
- Código gerado pelo sqlc divergir após mudança de paths → mitigação: regenerar e diffar contra o gerado atual antes de mover testes.
- Drift entre `.golangci.yml` novo e ambiente sem golangci-lint local → mitigação: rodar via container (mesmo padrão do `-race`); enforcement definitivo chega no E0c-CI.

## Estimativa de impacto

Baixo — refatoração estrutural sem mudança de comportamento, sem banco (schema intacto), sem infra, sem usuários. Prepara o terreno da E0b.
