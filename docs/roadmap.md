# Roadmap

Visão geral das entregas do projeto. Cada item vira uma ou mais tasks pequenas via skill `criar-task` (fluxo completo em `roles.md` §2).

## Visão

Projeto Morfeu — visão, escopo e entregas serão definidos pela **descoberta** (`iniciar-projeto`, roles.md §6.15), que substitui este placeholder pelo roadmap real.

## Itens

| # | Item | Prioridade | Status | Riscos/Dependências | Tasks |
|---|------|-----------|--------|---------------------|-------|
| 1 | Descoberta do projeto via `iniciar-projeto` (escopo, requisitos, volumetria, stack, observabilidade — gera `doc.md`, `lib.md` e substitui este roadmap) | alta | pendente | entrevista com o usuário; possíveis ADRs | — |

## Como usar

1. Rode a descoberta (item 1) — ela gera o roadmap real.
2. Escolha o item de maior prioridade com dependências resolvidas e rode `criar-task`.
3. Siga o fluxo: task → refinamento → PRD → implementação → auditoria → push → PR → CI/CD.
