---
name: auditoria
description: Auditoria pré-push deste projeto, conforme roles.md §6.4. Gate OBRIGATÓRIO — nenhum push acontece sem auditoria APROVADA. Valida escopo, PRD, ADRs, testes, segurança, dependências, migrations e arquivos de controle.
---

# Auditoria (gate pré-push)

Garante que a implementação respeita o escopo da task, os padrões definidos, os critérios de aceite, os testes, a segurança e a qualidade esperada. **Nenhum push antes da auditoria. Veredito REPROVADO bloqueia o push até correção.**

## Passo a passo

1. Levante o contexto: task atual (`state.md`), PRD (`docs/prd/NNNN-*.md`), ADRs ativos, `plan.md`.
2. Gere o diff completo da branch: `git diff main...HEAD` + `git status` (inclui não commitados).
3. Percorra o checklist abaixo, item a item, com evidência.
4. Acione o agente **`security`** para os itens 8, 9 e 11, e o agente **`qa`** para os itens 6 e 7 — inclua os vereditos deles.
5. Emita o relatório final (formato abaixo).
6. Se APROVADO: registre no `plan.md` (auditoria aprovada em AAAA-MM-DD) — o push está liberado (skill `fluxo-git`).
7. Se REPROVADO: liste as correções necessárias; re-auditar após corrigir.

## Checklist (13 itens — todos obrigatórios)

| # | Item | Como validar |
|---|---|---|
| 1 | Implementação respeita o **PRD** | diff vs. escopo/requisitos/critérios de aceite do PRD |
| 2 | Implementação respeita os **ADRs** existentes | diff vs. decisões em `docs/adr/` |
| 3 | Sem **overengineering** | abstrações/generalizações além do requisito real? |
| 4 | Sem alteração **fora do escopo** | todo arquivo do diff está previsto no PRD (ou o PRD foi atualizado) |
| 5 | Arquivos modificados **fazem sentido** para a task | ≤ 30 arquivos (roles.md §6.3) e coerentes |
| 6 | **Testes** criados/atualizados | fluxos críticos alterados têm teste; suíte passa |
| 7 | Testes são **idempotentes** | sem dependência de ordem/estado compartilhado (parecer do `qa`) |
| 8 | Sem **secrets expostos** | grep no diff por credenciais, tokens, chaves (parecer do `security`) |
| 9 | Sem **dados sensíveis em logs** | parecer do `security` |
| 10 | Dependências novas documentadas no **`lib.md`** | roles.md §6.9 |
| 11 | Sem **CVEs relevantes** nas dependências | parecer do `security` (fontes atuais) |
| 12 | **Migrations** criadas quando necessário | mudança de schema no diff → migration correspondente (roles.md §6.10) |
| 13 | **`state.md` atualizado** e **`plan.md` reflete a task** | roles.md §6.11; qualidade geral do código (legibilidade, manutenção, separação de responsabilidades) |

## Formato do relatório

```
# Auditoria — Task NNNN — AAAA-MM-DD

Veredito: APROVADO | REPROVADO

| # | Item | Resultado | Evidência/Observação |
|---|------|-----------|----------------------|
| 1 | PRD  | ✅/❌     | ...                  |
...

## Correções necessárias (se REPROVADO)
1. ...
```
