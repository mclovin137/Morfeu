---
name: criar-migration
description: Cria uma migration versionada de banco de dados neste projeto, conforme roles.md §6.10. Obrigatória para TODA alteração estrutural de banco — nunca alterar schema manualmente.
---

# Criar Migration

Toda mudança estrutural de banco (tabela, coluna, índice, constraint, dado de referência) entra por migration versionada, vinculada à task e ao PRD.

## Pré-requisitos

1. A task e o PRD existem e preveem a mudança de banco.
2. A ferramenta de migration do projeto está definida (ver `lib.md` e ADRs). Se ainda não estiver, consulte o agente `arquiteto` e o usuário — a escolha da ferramenta é decisão relevante (possível ADR, roles.md §6.1).
3. **Consulte o Context7** para a sintaxe e convenções da ferramenta escolhida (Flyway, Liquibase, Prisma, Alembic, etc.) — não presuma (roles.md §6.12).
4. Leia `boas-praticas-sql.md` (nesta pasta) — índices, full scan e zero-downtime são critérios de aprovação, não sugestão. Apoio: skills `database-migrations` e `postgres-patterns`.

## Passo a passo

1. Nomeie no padrão da ferramenta; na ausência de convenção própria, use `VNNNN__descricao_curta` com numeração sequencial.
2. Escreva a migration com cabeçalho-comentário padrão:
   ```
   -- Task: NNNN — [nome]
   -- PRD: docs/prd/NNNN-titulo.md
   -- Descrição: [o que muda e por quê]
   -- Rollback: [como desfazer / arquivo de down / "não aplicável" com justificativa]
   ```
3. **Rollback**: defina a estratégia de reversão quando aplicável (script down, coluna mantida até task de limpeza, etc.).
4. **Idempotência**: use guardas quando a ferramenta permitir (`IF NOT EXISTS`, checagens) — migrations re-executáveis reduzem risco.
5. **Impacto em dados existentes**: avalie e documente (tabela grande? lock? backfill necessário?). Mudanças destrutivas (drop, rename) exigem confirmação do usuário e seguem expand-contract (`boas-praticas-sql.md` §3).
6. **Performance e índices** (`boas-praticas-sql.md`): toda FK nova indexada; índices para os padrões de acesso do PRD (ordem do composto verificada); predicados sargáveis; queries dos fluxos críticos validadas com `EXPLAIN (ANALYZE, BUFFERS)` sobre volume representativo — **full scan não justificado em tabela grande reprova a migration**. Em ambiente com tráfego: `CONCURRENTLY`, `NOT VALID`/`VALIDATE`, backfill em lotes.
7. **Valide localmente antes do push**: suba o banco do `docker-compose` (agente `sre-devops`) e aplique a migration do zero e sobre base já migrada.
8. Registre a migration no PRD (arquivos criados) e no `plan.md`.

## Checklist final

- [ ] Vinculada à task e ao PRD (cabeçalho)
- [ ] Rollback definido ou justificado
- [ ] Idempotente quando possível
- [ ] Impacto em dados existentes avaliado
- [ ] Toda FK nova indexada; índices justificados pelo padrão de acesso do PRD
- [ ] `EXPLAIN` validado nos fluxos críticos — sem full scan não justificado em tabela grande
- [ ] Operações compatíveis com tráfego quando houver produção (`CONCURRENTLY`, `NOT VALID`/`VALIDATE`, backfill em lotes, expand-contract)
- [ ] Validada localmente (aplicação limpa + incremental)
- [ ] Nenhuma alteração manual de schema fora dela
