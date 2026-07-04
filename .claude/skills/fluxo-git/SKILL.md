---
name: fluxo-git
description: Fluxo de versionamento deste projeto, conforme roles.md §6.5 — nomenclatura de branch, regras de push, abertura de PR e rastreabilidade task↔PRD↔branch. Usar em todo trabalho com git neste repositório.
---

# Fluxo Git

Garante rastreabilidade e organização no versionamento: cada task tem branch e PRD próprios, e todo push passou por auditoria.

## Nomenclatura de branch

```
feature/NNNN-nome-da-task      nova funcionalidade
fix/NNNN-nome-do-ajuste        correção
chore/NNNN-nome-da-manutencao  manutenção/tooling
refactor/NNNN-nome-do-refactor refatoração sem mudança de comportamento
```

`NNNN` = número da task (mesmo do PRD). O nome reflete o objetivo da task.

## Regras (roles.md §6.5)

1. **Nunca** misturar múltiplas tasks na mesma branch.
2. **Nenhum push sem auditoria aprovada** (skill `auditoria` — roles.md §6.4).
3. Nenhum PR sem PRD e sem testes mínimos.
4. Arquivo fora do escopo do PRD só entra no diff se o PRD for atualizado antes.
5. Commits pequenos e descritivos, em PT-BR, no imperativo ("adiciona", "corrige").

## Antes do push — sequência obrigatória

1. `plan.md` reflete o estado final da task; `state.md` atualizado.
2. Auditoria executada e **APROVADA**.
3. Push da branch da task (nunca direto na `main`).

## Abertura de PR

Use o template `.github/pull_request_template.md`. O corpo do PR deve conter:

- **Resumo** da alteração
- **Task** relacionada (`docs/tasks/NNNN-*.md`)
- **PRD** relacionado (`docs/prd/NNNN-*.md`)
- **ADRs** relacionados (quando existirem)
- **Arquivos principais** alterados
- **Testes executados** (e resultado)
- **Riscos** conhecidos
- **Evidências de validação** (saída de testes, screenshots, logs)
- **Observações para revisão**

Após abrir o PR, o CI/CD executa todos os testes definidos (roles.md §2, etapa 8). PR só merge com pipeline verde.

## Ao concluir a task (pós-merge)

1. Atualizar `state.md` (última task concluída) e `docs/roadmap.md` (status do item).
2. Registrar a sessão no vault Obsidian: `/obsidian-log` + `/obsidian-decide` (roles.md §6.13).

## Skills de apoio

`git-workflow` (estratégias de branch, convenções de commit, merge vs rebase), `github-ops` (operações de PR, issues, CI e releases via gh CLI).
