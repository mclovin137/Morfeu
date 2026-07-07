# plan.md — Plano da Task em Andamento

> Este arquivo é o plano vivo da task corrente **do projeto** — não confundir com o *plan mode* do Claude Code (que grava em `~/.claude/plans/`). Atualizado durante a implementação; reflete o estado real (regras em `roles.md` §6.11).

## Task atual

**Nenhuma task em andamento.** Descoberta e ADRs 0001–0006 concluídos (bootstrap de governança, pré-task). A primeira task, criada via `criar-task` sobre o item 0/E0 do roadmap, preenche este arquivo com: objetivo, escopo, fora de escopo, branch, PRD, tabela de arquivos, dependências, plano de testes, pendências e riscos.

## Auditorias

- **2026-07-07 — APROVADA** (registro da identidade visual, commit `93104b5`): 13 itens verificados; diff de 6 arquivos, 100% documentação (`docs/design/` novo + ponteiros em roles.md/doc.md/state.md). Security ✅ (item 8: zero secrets — 4 blobs base64 validados como woff2 genuínos por magic bytes, zero alta-entropia residual, PII só placeholders; item 9: N/A-justificado + JS do protótipo verificado sem rede/console/storage; item 11: nenhuma dependência nova — fontes embutidas são asset estático; pendência `@fontsource/*`→lib.md formalizada p/ quando a SPA nascer). QA ✅ (itens 6/7 N/A-justificados — sem código de produção no repo; consistência das 5 referências cruzadas confirmada; observação não bloqueante: notação "E0c" vs. E0 item (c) do roadmap). Itens 1–5, 10, 12, 13 na sessão principal: sem PRD/task por ser artefato de governança-design pré-task (mesmo precedente do bootstrap), escopo coerente, 6 ≤ 30 arquivos, lib.md intacto, sem schema, state.md/plan.md atualizados.
- **2026-07-07 — APROVADA** (bootstrap de governança, pré-push inicial): 13 itens verificados; security ✅ (árvore publicável + histórico sem secrets; CVEs do lib.md revalidadas no OSV), qa ✅ (itens 6/7 N/A-justificados — diff 100% documentação; consistência ADR 0006 ↔ doc.md ↔ roadmap confirmada; divergência lib.md "3–5→2–4 jornadas" corrigida no ato).
