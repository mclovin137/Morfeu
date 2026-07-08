---
name: criar-prd
description: Cria o PRD (Product Requirements Document) de uma task deste projeto, conforme roles.md §6.2. Obrigatório antes de qualquer implementação — nenhuma task é implementada sem PRD.
---

# Criar PRD

Detalha uma task antes da implementação. O PRD orienta e delimita: nada é implementado fora dele sem atualizá-lo primeiro.

## Pré-requisitos

1. A task existe em `docs/tasks/` (senão, rode `criar-task` antes).
2. O **épico da task foi refinado** — existe `docs/refinamentos/ENN-*.md` com as exigências desta task (senão, rode `refinar-task` para o épico; roles.md §6.14). Task trivial dispensa refinamento (§6.14.6). Perguntas escaladas ao usuário no refinamento bloqueiam o PRD até resposta.
3. Você leu `roles.md`, `state.md` e os ADRs ativos que tocam o tema.

## Passo a passo

1. Use o mesmo número da task: `docs/prd/NNNN-titulo-kebab.md` (task e PRD compartilham o `NNNN`).
2. Copie o template `template-prd.md` (nesta pasta) e preencha todas as seções, **incorporando as exigências da task registradas no refinamento do épico** (obrigatório — roles.md §6.14.5).
3. Seja específico em **Arquivos que serão criados/modificados** — a auditoria compara o diff real contra esta lista.
4. **Sem nova rodada de agentes** (roles.md §6.2.5): o refinamento do épico já traz os pareceres. Consultar um agente pontualmente só para lacuna real que o refinamento não cobriu.
5. Dependência nova prevista? Justifique no PRD e planeje o registro em `lib.md` (roles.md §6.9).
6. Atualize `state.md` (PRD atual) e `plan.md` (PRD relacionado).

## Regra de manutenção

Se durante a implementação o plano mudar (arquivos diferentes, escopo ajustado, dependência trocada), **atualize o PRD imediatamente** — o PRD desatualizado reprova auditoria (item 1 do checklist da skill `auditoria`).

## Checklist final

- [ ] Todas as seções preenchidas (sem "N/A" por preguiça — justificar quando não aplicável)
- [ ] Critérios de aceite verificáveis objetivamente
- [ ] Plano de testes cobre fluxos principais, erro e borda
- [ ] Arquivos previstos listados
- [ ] state.md e plan.md atualizados
