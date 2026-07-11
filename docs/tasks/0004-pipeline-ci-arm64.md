# Task 0004 — Pipeline de CI + build ARM64 (E0c-CI)

- **Data:** 2026-07-11
- **Status:** pendente
- **Branch:** `chore/0004-pipeline-ci-arm64`
- **PRD:** docs/prd/0004-pipeline-ci-arm64.md — criar antes de implementar
- **Item do roadmap:** E0 (walking skeleton), entrega **E0c-CI** — nova ordem aprovada pelo usuário em 2026-07-11; exigências em `docs/refinamentos/E0-walking-skeleton.md` §"Task E0c-CI"

> Numeração: **0002 permanece reservada para a E0b** (outbox + RabbitMQ). Esta task recebe 0004.

## Objetivo

Substituir o CI placeholder pelo pipeline real de gates — a **peça mecânica do gate híbrido de auditoria** (roles.md §6.4). A partir do merge desta task, os itens mecânicos da auditoria (lint, vet, testes com `-race`, govulncheck, sqlc vet, gitleaks, migrations, build) passam a rodar no CI, e a skill `auditoria` reduz ao passe único de julgamento.

## Escopo

- Workflow `build-test` (gate de PR, roda sempre): `golangci-lint`, `go vet`, `go test -race ./...` (suíte completa, testcontainers reais), `govulncheck`, `sqlc vet`, gitleaks (falha o job), job de migrations **condicional por path** (`migrations/` → up→down→up + sqlc vet), build `linux/arm64` → GHCR com tag por SHA + **smoke do binário dentro do job** (roda, não só builda). Runner ARM64 nativo (`ubuntu-24.04-arm`), sem QEMU; cache GHA no build Docker.
- Supply chain: actions pinadas por **SHA completo** + Dependabot para atualizar os pins; imagem base do Dockerfile pinada por digest; `GITHUB_TOKEN` com `permissions:` explícito (`packages: write`), sem PAT.
- **Quitação do débito de lint da E0a** (pré-condição do gate: o lint não pode nascer vermelho na main): 43 issues registradas na auditoria da 0003 (errcheck 26, revive 6, staticcheck 4, errorlint 4, gosec 2, funlen 1) — corrigir no código; supressão `nolint` justificada só onde correção real não se aplicar. Inclui trocar `time.Sleep(1s)` por wait strategies nos 3 arquivos de teste (achado da mesma auditoria).
- Testcontainers no CI: timeouts 90–120s, wait por log. Suíte completa da E0a verde no runner real — falha por diferença de ambiente é bug desta task, não flaky.
- Validação: **PRs de prova descartáveis** demonstrando bloqueio real de cada gate (lint quebrado, data race injetada, CVE conhecida, SQL inválido, secret de teste, migration sem down; job de migrations não roda quando não toca `migrations/`).
- Transição de governança declarada no PRD: atualização do texto da skill `auditoria`/`roles.md` §6.4 sai do modo transição.
- Auditoria pontual do histórico completo (gitleaks full-history) registrada — já executada nas auditorias de 2026-07-10/11 (1 falso positivo conhecido); repetir e registrar formalmente antes do primeiro deploy real (E0c-CD).

## Fora de escopo

- **Shell da SPA** — recomendação do refinamento era "nasce aqui, confirmar na abertura do PRD"; decisão na abertura desta task: **fica para task própria** (0005 candidata), porque a quitação do débito de lint (~11 arquivos de código) consome a folga da estimativa (limite de 30 arquivos, roles.md §6.3). Registrado aqui como desvio consciente da recomendação não-bloqueante.
- Deploy, Caddy/TLS, smoke em URL pública (E0c-CD, bloqueada pela VM Oracle).
- Observabilidade (E0d).

## Arquivos esperados (~20–24)

- `.github/workflows/ci.yml` (reescrito) e possivelmente 1 workflow auxiliar (migrations condicional) — 1–2
- `.github/dependabot.yml` (novo) — 1
- `Dockerfile` (digest pin) — 1
- Lint: `cmd/morfeu/main.go`, `internal/logger/logger.go`, `internal/cache/redis.go`, `internal/catalogo/{handler,service}.go` e arquivos de teste (`internal/catalogo/*_test.go`, `internal/health/health_test.go`, `internal/config/config_test.go`, `internal/cache/redis_test.go`, `internal/integration*.go`) — ~11–13
- `.golangci.yml` (ajustes finos, se necessários) — 0–1
- Governança/controle: `roles.md` §6.4 (fim da transição), skill `auditoria` (modo-alvo), `docs/tasks/`, `docs/prd/`, `state.md`, `plan.md`, `docs/roadmap.md` — ~6

## Dependências esperadas

Nenhuma dependência Go nova (govulncheck/gitleaks/sqlc rodam como binários/actions no CI). Actions novas registradas no PRD com SHA pinado (não vão ao `lib.md` — não são dependências de runtime; convenção a confirmar no PRD).

## Critérios de aceite

- [ ] CA01 — PR aberto dispara o workflow e **todos os gates passam** na main atual (lint zero issues, suíte `-race` completa verde no runner ARM64, govulncheck, sqlc vet, gitleaks).
- [ ] CA02 — Cada gate **bloqueia de verdade**: PRs de prova descartáveis com evidência de falha (lint, race, CVE, SQL inválido, secret, migration sem down) registradas no PRD/PR final.
- [ ] CA03 — Job de migrations roda **somente** quando o diff toca `migrations/` (prova positiva e negativa).
- [ ] CA04 — Imagem `linux/arm64` publicada no GHCR com tag por SHA; smoke do binário dentro do job (processo sobe e responde) — sem QEMU.
- [ ] CA05 — Supply chain: 100% das actions pinadas por SHA completo; Dependabot ativo; base image por digest; `permissions:` explícito no workflow.
- [ ] CA06 — Débito de lint quitado: `golangci-lint run ./...` → 0 issues (nolint apenas com justificativa inline).
- [ ] CA07 — Governança atualizada: roles.md §6.4 sai do modo transição; skill `auditoria` passa ao modo-alvo (julgamento único; mecânicos no CI).

## Riscos

- Flakiness testcontainers no runner ARM64 → wait por log, timeouts 90–120s, container por pacote (exigência do refinamento).
- Minutos de CI (repo público tem minutos ilimitados em runners padrão; `ubuntu-24.04-arm` é gratuito p/ repos públicos — validar no PRD com fonte atual).
- Correções de lint alterarem comportamento sem querer → regra: só correções mecânicas (checagem de erro, conversões seguras); suíte completa como rede; asserts intactos.
- PRs de prova poluírem o histórico → branches descartáveis, fechados sem merge, evidência por screenshot/log no PRD.

## Estimativa de impacto

Médio em infra de desenvolvimento (CI vira gate obrigatório de merge), baixo em código de produção (correções mecânicas de lint), nenhum em banco/usuários. Alto em governança: encerra o modo transição da auditoria (§6.4).
