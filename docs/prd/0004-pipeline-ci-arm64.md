# PRD 0004 — Pipeline de CI + build ARM64 (E0c-CI)

- **Task:** docs/tasks/0004-pipeline-ci-arm64.md
- **Branch:** `chore/0004-pipeline-ci-arm64`
- **Data:** 2026-07-11
- **Status:** ativo

> Consome as exigências do refinamento do épico E0 (`docs/refinamentos/E0-walking-skeleton.md` §"Task E0c-CI") — sem nova rodada de agentes (roles.md §6.2.5). Skills de apoio à implementação (roles.md §4.4): `golang-lint`, `github-ops`.

## Objetivo

Substituir o CI placeholder (`.github/workflows/ci.yml`, 100% `echo TODO`) pelo pipeline real de gates. É a **peça mecânica do gate híbrido de auditoria** (roles.md §6.4): a partir do merge, lint/vet/testes-race/govulncheck/sqlc-vet/gitleaks/migrations/build rodam no CI em todo PR, e a skill `auditoria` reduz ao passe único de julgamento (fim do modo transição). O fix emergencial da E0a (PR #3) provou o custo da ausência deste gate: uma E0a que nunca compilou chegou à main.

## Escopo

1. **Workflow `ci.yml` real** (gate de PR + push na main), runner **`ubuntu-24.04-arm`** (ARM64 nativo, sem QEMU; gratuito para repositórios públicos — GA desde 2025-01; confirmar na doc oficial do GitHub durante a implementação e registrar a fonte no PR).
2. **Gates bloqueantes**, nesta ordem de custo: `golangci-lint` (config v2 do repo) → `go vet` → `gitleaks` (working tree + histórico do PR) → `govulncheck` → `sqlc vet` + `sqlc generate --check` (diff vazio) → `go test -race ./...` **suíte completa** com testcontainers reais (timeout de job 20 min; wait por log; 90–120s por container) → build Docker `linux/arm64` → push GHCR `ghcr.io/mclovin137/morfeu` com tag por SHA → **smoke da imagem no job** (services PG/Redis do Actions; `docker run` da imagem; `curl /health` → 200).
3. **Job de migrations condicional por path** (`migrations/**` no diff → PG efêmero, `migrate up → down → up` + `sqlc vet`); não roda em PRs que não tocam migrations.
4. **Supply chain**: todas as actions pinadas por **SHA completo** (comentário com a versão); `.github/dependabot.yml` (ecosystems: `github-actions`, `gomod`, `docker`); base image do `Dockerfile` pinada por **digest**; `permissions:` explícito no workflow (default `contents: read`; `packages: write` só no job de build/push); sem PAT.
5. **Quitação do débito de lint da E0a** (pré-condição: gate não pode nascer vermelho) — 43 issues registradas na auditoria da 0003:
   - `errcheck` (26): checagem de erro em `defer Terminate/Close` dos testes (helper `t.Cleanup` com log do erro) e retornos ignorados em produção;
   - `revive` (6): inclui `context-as-argument` (ctx como 1º parâmetro nos helpers de teste);
   - `staticcheck` (4): `middleware.LoggerWithConfig` deprecado → `RequestLoggerWithConfig`; SA5011 em `service_test.go`; QF1008 em `logger.go`;
   - `errorlint` (4), `gosec` (2: conversões int→int32 com validação de range em `main.go`), `funlen` (1: extrair função);
   - `nolint` apenas com justificativa inline onde correção real não couber. Regra dura: **só correções mecânicas, zero mudança de comportamento, asserts intactos** (mesmo critério da 0003).
   - Junto: trocar `time.Sleep(time.Second)` por wait strategies (`wait.ForLog`/`ForListeningPort`) nos 3 arquivos de teste (achado da auditoria da 0003).
6. **PRs de prova descartáveis** (validação de que cada gate bloqueia): lint quebrado; data race injetada; CVE conhecida (dep de teste); SQL inválido; secret de teste (dummy, formato AWS key); migration sem down; prova negativa/positiva do path-filter de migrations. Branches fechadas sem merge; evidência (link do run vermelho) registrada na seção "Evidências" do PR final.
7. **Transição de governança**: `roles.md` §6.4 sai do modo transição (pré-push = scan de secrets local; mecânicos no CI; julgamento em passe único no PR); skill `.claude/skills/auditoria/SKILL.md` atualizada para o modo-alvo; nota em `CLAUDE.md` (invariante 2) removendo a cláusula de transição.

## Fora de escopo

- **Shell da SPA** (decisão de abertura da task: task própria — limite de 30 arquivos após absorver o débito de lint; desvio consciente da recomendação não-bloqueante do refinamento, registrado no task doc).
- Deploy, Caddy/TLS, smoke em URL pública (E0c-CD, bloqueada pela VM Oracle); CD de qualquer natureza.
- Observabilidade (E0d). Cobertura mínima como gate (fica para quando houver baseline).
- Auditoria formal full-history pré-deploy (registrada como passo do E0c-CD; gitleaks full-history já roda aqui como gate).

## Requisitos funcionais

- RF01 — Todo PR dispara o workflow; **qualquer gate vermelho bloqueia o merge** (branch protection: required check `ci`).
- RF02 — Job de migrations roda **se e somente se** o diff toca `migrations/**`.
- RF03 — Push na `main` publica imagem `linux/arm64` no GHCR com tags `sha-<SHA>` e `latest`.
- RF04 — Smoke: a imagem publicada sobe com PG/Redis reais e responde `200` em `GET /health` dentro do job.
- RF05 — `golangci-lint run ./...` retorna **0 issues** na branch (débito quitado).

## Requisitos não funcionais

- RNF01 — Segurança de supply chain: actions por SHA completo; base image por digest; `permissions:` mínimo; Dependabot ativo nos 3 ecosystems; `GITHUB_TOKEN` (sem PAT).
- RNF02 — Tempo total do workflow ≤ 20 min no caminho feliz (cache GHA de Go modules/build e camadas Docker).
- RNF03 — Reprodutibilidade: versões das ferramentas fixadas (golangci-lint, govulncheck, sqlc, gitleaks, golang-migrate) — mesmo major/minor usado localmente.
- RNF04 — Zero mudança de comportamento observável no código de produção (correções de lint são mecânicas; suíte completa é a rede).

## Regras de negócio

- RN01 — N/A no sentido de domínio (task de infraestrutura de desenvolvimento). Regra de governança equivalente: a partir do merge, **nenhum merge sem CI verde + passe de julgamento** (roles.md §6.4.1) — o modo transição da skill `auditoria` encerra.

## Critérios de aceite

- [ ] CA01 — Workflow real substitui o placeholder; todos os gates verdes na branch (incluindo suíte `-race` completa com testcontainers no runner ARM64).
- [ ] CA02 — Cada gate bloqueia de verdade: 6 PRs de prova vermelhos (lint, race, CVE, SQL inválido, secret, migration sem down) + evidência dos runs registrada no PR final.
- [ ] CA03 — Path-filter de migrations comprovado nos dois sentidos (roda quando toca `migrations/**`; não roda quando não toca).
- [ ] CA04 — Imagem `linux/arm64` no GHCR com tag por SHA; smoke no job com `curl /health` → 200; sem QEMU no build.
- [ ] CA05 — Supply chain: 100% das actions por SHA completo; Dependabot (`github-actions` + `gomod` + `docker`); Dockerfile por digest; `permissions:` explícito.
- [ ] CA06 — `golangci-lint run ./...` → 0 issues; `git diff` das correções de lint não altera nenhum assert; suíte completa verde com `-race`.
- [ ] CA07 — Governança atualizada: roles.md §6.4 (modo-alvo), skill `auditoria` (modo-alvo), CLAUDE.md (invariante 2 sem cláusula de transição).
- [ ] CA08 — Branch protection da `main` exige o check `ci` (configurado via `gh api`; evidência no PR).

## Plano de testes

| Cenário | Tipo | Cobre |
|---|---|---|
| Suíte completa da E0a no runner ARM64 (`go test -race ./...`, testcontainers PG/Redis) | integração (CI) | regressão zero pós-correções de lint; ambiente ARM64 real |
| PR de prova: lint quebrado / race injetada / CVE / SQL inválido / secret dummy / migration sem down | e2e do pipeline | cada gate bloqueia (CA02) |
| PR tocando `migrations/**` vs. PR sem tocar | e2e do pipeline | path-filter (CA03) |
| Smoke da imagem no job (PG/Redis services + `curl /health`) | e2e do pipeline | binário ARM64 roda de verdade (CA04) |
| `sqlc generate --check` | CI | código gerado sincronizado com queries (regressão da 0003) |
| Wait strategies nos testes (substituindo `time.Sleep`) | integração local + CI | determinismo da suíte (achado da auditoria 0003) |

## Plano de implementação

1. **Bloco lint** (commits pequenos por linter): errcheck → revive/context → staticcheck (logger deprecado) → gosec (conversões validadas) → errorlint → funlen; depois wait strategies no lugar de `time.Sleep`. Validação por bloco: `golangci-lint run` + suíte `-short`; ao final, suíte completa `-race` via container.
2. **Workflow `ci.yml`**: jobs `gates` (lint→vet→gitleaks→govulncheck→sqlc) e `test` (suíte -race), `migrations` (condicional por path), `build` (docker buildx arm64 → GHCR + smoke). Actions pinadas por SHA; cache GHA.
3. **`dependabot.yml`** + digest no Dockerfile + `permissions:` no workflow.
4. Abrir PR da task → validar CA01 no run real; ajustar flakiness (timeouts/wait) se houver.
5. **PRs de prova** (um por gate) a partir da branch da task; coletar links dos runs vermelhos; fechar sem merge.
6. Branch protection via `gh api` (required check `ci`).
7. **Governança**: roles.md §6.4, skill `auditoria`, CLAUDE.md — modo-alvo.
8. Atualizar plan.md/state.md/task 0004; auditoria pré-push (última no modo transição); PR final.

## Arquivos que serão criados

- `.github/dependabot.yml` — pins de actions + gomod + docker

## Arquivos que serão modificados

- `.github/workflows/ci.yml` — reescrito (placeholder → pipeline real)
- `Dockerfile` — base image por digest
- `cmd/morfeu/main.go` — staticcheck (logger middleware), gosec (int→int32 com validação)
- `internal/logger/logger.go` — QF1008
- `internal/cache/redis.go`, `internal/catalogo/handler.go`, `internal/catalogo/service.go` — errcheck/errorlint pontuais
- `internal/catalogo/handler_test.go`, `internal/catalogo/service_test.go`, `internal/health/health_test.go`, `internal/config/config_test.go`, `internal/cache/redis_test.go`, `internal/integration_test.go`, `internal/integration_cache_test.go` — errcheck (cleanup), revive (ctx 1º param), wait strategies, funlen
- `.golangci.yml` — ajustes finos se necessários (sem afrouxar regra para esconder débito)
- `roles.md` (§6.4), `.claude/skills/auditoria/SKILL.md`, `CLAUDE.md` — fim do modo transição
- `docs/tasks/0004-pipeline-ci-arm64.md`, `docs/tasks/README.md`, `state.md`, `plan.md`, `docs/roadmap.md` — controle

**Total estimado: ~22 arquivos** (≤ 30 ✓).

## Dependências utilizadas

Nenhuma dependência Go nova (`go.mod` intacto — correções de lint usam stdlib/APIs existentes). Ferramentas de CI (golangci-lint, govulncheck, sqlc, gitleaks, golang-migrate) e actions do workflow **não são dependências de runtime** — ficam registradas neste PRD com versão/SHA pinados (convenção: actions não entram no `lib.md`; se o usuário preferir registrá-las, adicionar seção "tooling" ao lib.md em task futura).

## Impactos técnicos

- Governança: encerra o modo transição do §6.4 — todo o fluxo de auditoria muda de forma permanente (mecânicos no CI, julgamento único no PR).
- Infra de dev: GHCR passa a receber imagens ARM64 por SHA (base para o deploy do E0c-CD).
- Código: zero impacto de comportamento (correções mecânicas); testes ficam mais determinísticos (wait strategies).
- Banco/contratos: nenhum.

## Riscos

- Flakiness testcontainers no runner ARM64 → wait por log, timeout 90–120s por container, container por pacote; 1 re-run diagnóstico antes de qualquer ajuste (falha por ambiente = bug da task, não flaky — refinamento).
- Runner `ubuntu-24.04-arm` indisponível/pago para o repo → fallback documentado: `ubuntu-latest` + QEMU só para o job de build (mais lento), gates nativos em x64; registrar desvio no PRD.
- Correção de lint alterar comportamento → só correções mecânicas; suíte completa `-race` como rede; asserts intactos (CA06).
- PRs de prova dispararem alertas do GitHub (secret scanning) → usar secrets dummy de formato conhecido mas inválido (ex.: `AKIA` + zeros); nunca credencial real.
- `sqlc vet` exigir banco → usar PG service no job (mesma imagem 16-alpine do compose).

## Estratégia de rollback

Sem migrations. Rollback = reverter o merge do PR (workflow volta ao anterior; correções de lint são revertíveis por commit). O gate de branch protection pode ser desativado via `gh api` em emergência (registrar no state.md se ocorrer).
