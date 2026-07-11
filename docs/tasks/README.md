# Tasks

Tasks do projeto, criadas via skill `criar-task` a partir do roadmap (regras em `roles.md` §6.3). Template em `.claude/skills/criar-task/template-task.md`.

Convenção: `NNNN-titulo-kebab.md`, numeração sequencial a partir de `0001`. Cada task = 1 branch + 1 PRD + máx. 30 arquivos alterados.

## Índice

| # | Título | Roadmap | Status | Branch | PRD |
|---|--------|---------|--------|--------|-----|
| 0001 | [App Skeleton Go](./0001-app-skeleton-go.md) | E0a | concluída (merge `6916ab6`; fix de build PR #3 `fa2a136`) | `feature/0001-app-skeleton-go` | [0001](../prd/0001-app-skeleton-go.md) |
| 0002 | *(reservada)* Outbox + RabbitMQ — E0b | E0b | desbloqueada (refinamento respondido; ADR 0007 criado) — na fila após E0c-CI | — | — |
| 0003 | [Conformidade package-by-domain](./0003-conformidade-package-by-domain.md) | E0 (pré-E0b) | concluída (auditoria APROVADA; merge `a391cdb`, PR #4) | `refactor/0003-conformidade-package-by-domain` | [0003](../prd/0003-conformidade-package-by-domain.md) |
| 0004 | [Pipeline de CI + build ARM64](./0004-pipeline-ci-arm64.md) | E0c-CI | em andamento | `chore/0004-pipeline-ci-arm64` | [0004](../prd/0004-pipeline-ci-arm64.md) |
