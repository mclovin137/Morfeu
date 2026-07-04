---
name: criar-task
description: Quebra o roadmap deste projeto em tasks pequenas e rastreáveis, conforme roles.md §6.3. Usar ao iniciar qualquer trabalho novo — cada task tem branch própria, PRD próprio e no máximo 30 arquivos alterados.
---

# Criar Task

Transforma um item do roadmap em uma task pequena, revisável e rastreável.

## Pré-requisitos

1. O item existe (ou faz sentido) no `docs/roadmap.md` — se o roadmap estiver vazio ou desatualizado, alinhe com o usuário primeiro.
2. Você leu `roles.md` e `state.md`.

## Passo a passo

1. Determine o próximo número: liste `docs/tasks/` e use `NNNN` + 1 (4 dígitos, começa em `0001`).
2. Copie o template `template-task.md` (nesta pasta) e preencha todas as seções.
3. **Valide o tamanho**: estime os arquivos alterados. Se a estimativa passar de **30 arquivos**, quebre em tasks menores ANTES de criar (roles.md §6.3).
4. Salve em `docs/tasks/NNNN-titulo-kebab.md` e atualize o índice em `docs/tasks/README.md`.
5. Crie a branch da task: `feature/NNNN-titulo` (ou `fix/`, `chore/`, `refactor/` — ver skill `fluxo-git`).
6. Atualize `state.md` (task atual, branch atual).
7. Regenere `plan.md` para a nova task (objetivo, escopo, arquivos previstos, status "não iniciada").
8. Próximo passo obrigatório do fluxo: **refinamento multiagente** (`refinar-task`, roles.md §6.14) e só então o PRD (`criar-prd`).

## Checklist final

- [ ] Escopo pequeno e revisável (≤ 30 arquivos estimados)
- [ ] Critérios de aceite claros e verificáveis
- [ ] Branch criada com nomenclatura correta
- [ ] docs/tasks/README.md, state.md e plan.md atualizados
- [ ] Roadmap referencia a task (rastreabilidade)
