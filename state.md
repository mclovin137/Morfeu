# state.md — Estado Atual do Projeto

Atualizado ao final de cada task e antes de cada PR (regras em `roles.md` §6.11).

## Estado atual

Projeto **Morfeu** recém-criado a partir do template Tllm (commit `2a86322`, 2026-07-04). **Aguardando descoberta** (`iniciar-projeto`, roles.md §6.15) — escopo, stack, roadmap real, `doc.md` e `lib.md` nascem dela.

- **Última task concluída:** nenhuma
- **Task atual:** nenhuma
- **Branch atual:** `main`
- **PRD atual:** nenhum
- **ADRs ativos:** nenhum

## Últimas decisões relevantes

- 2026-07-04 — Projeto criado a partir do template Tllm (governança multiagente completa herdada; histórico do template fica no repo de origem).

## Pendências técnicas

- Rodar a descoberta (`/iniciar-projeto`) — define nome/escopo/stack e gera `doc.md`, `lib.md` e `docs/roadmap.md` reais.
- Criar remote no GitHub e fazer o primeiro push (após a descoberta, com auditoria quando houver código).
- Preencher os steps reais do `.github/workflows/ci.yml` quando a stack existir.

## Riscos conhecidos

- Regras de governança dependem de disciplina convencional (sem enforcement técnico por hooks ainda).
- Repo em `/mnt/c` (WSL): git mais lento; `core.filemode=false` configurado.

## Próximos passos

1. `/iniciar-projeto` — descoberta completa (entrevista + confronto multiagente).
2. Revisar e aprovar `doc.md`, `lib.md` e `docs/roadmap.md` gerados.
3. `/criar-task` sobre o item 1 do roadmap.

## Histórico resumido

| Data | Evento |
|------|--------|
| 2026-07-04 | Projeto Morfeu criado a partir do template Tllm `2a86322` (28 skills, 5 agentes, hooks de sessão Obsidian, arquivos de controle). |
