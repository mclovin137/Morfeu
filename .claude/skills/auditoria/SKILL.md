---
name: auditoria
description: Gate de auditoria deste projeto, conforme roles.md §6.4 (modelo híbrido). Itens mecânicos rodam no CI (.github/workflows/ci.yml); esta skill cobre o scan local de secrets pré-push e os itens de julgamento num único passe de revisão sobre o diff do PR. Nenhum merge sem CI verde + passe aprovado.
---

# Auditoria (gate híbrido — roles.md §6.4)

Garante que a implementação respeita o escopo da task, os padrões definidos, os critérios de aceite, os testes, a segurança e a qualidade esperada.

**Modo vigente (CI da task 0004 no ar):** os itens mecânicos são validados pelo pipeline (`.github/workflows/ci.yml`) — não re-executar manualmente. Esta skill cobre: (a) **pré-push**, o scan local de secrets (gitleaks) limpo — único requisito para o push (§6.4.1); (b) **no PR**, os **itens de julgamento** num **único passe** sobre o diff, incluindo os mecânicos ainda sem job próprio (§6.4.5: cobertura, contagem ≤ 30 arquivos, diff `go.mod` × `lib.md`, presença de `plan.md`/`state.md`). Nenhum merge sem CI verde + este passe aprovado.

## Passo a passo

1. Levante o contexto: task atual (`state.md`), PRD (`docs/prd/NNNN-*.md`), ADRs ativos, `plan.md`.
2. Gere o diff completo da branch: `git diff main...HEAD` + `git status` (inclui não commitados).
3. Itens mecânicos: confira o resultado do pipeline no PR (`gh pr checks`); pré-push, rode só o gitleaks local (árvore + histórico). Os mecânicos sem job próprio (§6.4.5) entram no passe de julgamento.
4. Itens de julgamento: **um único agente revisor** (`qa` ou `security`, conforme o tema dominante da task — não os dois) percorre os itens 1–5, 7, 9 e a qualidade do 13 sobre o diff, com evidência por item. Dependência nova no diff → o passe inclui o item 11 com fontes atuais.
5. Emita o relatório final (formato abaixo).
6. Se APROVADO: registre no `plan.md` (auditoria aprovada em AAAA-MM-DD) — o push está liberado (skill `fluxo-git`).
7. Se REPROVADO: liste as correções; após corrigir, **revalide apenas os itens reprovados** (máx. 1 rodada extra; persistindo, escale ao usuário). Nunca re-auditar o diff inteiro por correção pontual (§6.4.4).

## Checklist (13 itens — todos obrigatórios; coluna "Gate" indica onde o item vive no modelo-alvo)

| # | Item | Gate | Como validar |
|---|---|---|---|
| 1 | Implementação respeita o **PRD** | julgamento | diff vs. escopo/requisitos/critérios de aceite do PRD |
| 2 | Implementação respeita os **ADRs** existentes | julgamento | diff vs. decisões em `docs/adr/` |
| 3 | Sem **overengineering** | julgamento | abstrações/generalizações além do requisito real? |
| 4 | Sem alteração **fora do escopo** | julgamento | todo arquivo do diff está previsto no PRD (ou o PRD foi atualizado) |
| 5 | Arquivos modificados **fazem sentido** para a task | julgamento (inclui contagem ≤ 30 — §6.4.5) | coerentes com a task (roles.md §6.3) |
| 6 | **Testes** criados/atualizados | CI | fluxos críticos alterados têm teste; suíte passa (`go test -race`; gate de cobertura futuro — §6.4.5) |
| 7 | Testes são **idempotentes** | julgamento | sem dependência de ordem/estado compartilhado |
| 8 | Sem **secrets expostos** | CI + pré-push | gitleaks no diff e no histórico; push protection do GitHub |
| 9 | Sem **dados sensíveis em logs** | julgamento | grep + leitura dos pontos de log do diff |
| 10 | Dependências novas documentadas no **`lib.md`** | julgamento (§6.4.5) | diff `go.mod` × `lib.md` (roles.md §6.9) |
| 11 | Sem **CVEs relevantes** nas dependências | CI (+ julgamento se dep. nova) | `govulncheck`; dep. nova → fontes atuais |
| 12 | **Migrations** criadas quando necessário | CI | mudança de schema no diff → migration correspondente validada em PG efêmero (roles.md §6.10) |
| 13 | **`state.md` atualizado** e **`plan.md` reflete a task** | julgamento (§6.4.5) | roles.md §6.11; qualidade geral do código (legibilidade, manutenção, separação de responsabilidades) |

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
