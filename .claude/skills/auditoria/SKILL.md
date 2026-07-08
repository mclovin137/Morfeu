---
name: auditoria
description: Gate de auditoria deste projeto, conforme roles.md §6.4 (modelo híbrido). Itens mecânicos rodam no CI; itens de julgamento num único passe de revisão. TRANSIÇÃO — enquanto o CI (E0c) não existir, esta skill roda pré-push completa e nenhum push acontece sem veredito APROVADO.
---

# Auditoria (gate híbrido — roles.md §6.4)

Garante que a implementação respeita o escopo da task, os padrões definidos, os critérios de aceite, os testes, a segurança e a qualidade esperada.

**Modelo-alvo (com CI do E0c no ar):** itens mecânicos (6–8, 10–12 e parte do 13) são validados pelo pipeline — não re-executar manualmente; esta skill cobre só os **itens de julgamento** (1–5, 9, qualidade do 13) num **único passe** sobre o diff do PR. Nenhum merge sem CI verde + este passe aprovado.

**Transição (sem CI):** esta skill roda **pré-push** cobrindo os 13 itens. Nenhum push sem veredito APROVADO.

## Passo a passo

1. Levante o contexto: task atual (`state.md`), PRD (`docs/prd/NNNN-*.md`), ADRs ativos, `plan.md`.
2. Gere o diff completo da branch: `git diff main...HEAD` + `git status` (inclui não commitados).
3. Itens mecânicos: se o CI existir, confira o resultado do pipeline; senão, rode os comandos localmente (testes, lint, govulncheck, gitleaks) — sem agente.
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
| 5 | Arquivos modificados **fazem sentido** para a task | julgamento (contagem ≤ 30: CI) | coerentes com a task (roles.md §6.3) |
| 6 | **Testes** criados/atualizados | CI | fluxos críticos alterados têm teste; suíte passa (`go test -race` + cobertura) |
| 7 | Testes são **idempotentes** | julgamento | sem dependência de ordem/estado compartilhado |
| 8 | Sem **secrets expostos** | CI + pré-push | gitleaks no diff e no histórico; push protection do GitHub |
| 9 | Sem **dados sensíveis em logs** | julgamento | grep + leitura dos pontos de log do diff |
| 10 | Dependências novas documentadas no **`lib.md`** | CI | diff `go.mod` × `lib.md` (roles.md §6.9) |
| 11 | Sem **CVEs relevantes** nas dependências | CI (+ julgamento se dep. nova) | `govulncheck`; dep. nova → fontes atuais |
| 12 | **Migrations** criadas quando necessário | CI | mudança de schema no diff → migration correspondente validada em PG efêmero (roles.md §6.10) |
| 13 | **`state.md` atualizado** e **`plan.md` reflete a task** | CI (presença) + julgamento (qualidade) | roles.md §6.11; qualidade geral do código (legibilidade, manutenção, separação de responsabilidades) |

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
