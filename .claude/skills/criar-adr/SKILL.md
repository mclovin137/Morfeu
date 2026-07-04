---
name: criar-adr
description: Cria um ADR (Architecture Decision Record) deste projeto, conforme roles.md §6.1. Usar apenas para decisões arquiteturais relevantes e SOMENTE com autorização explícita do usuário.
---

# Criar ADR

Registra uma decisão arquitetural relevante em `docs/adr/`.

## Pré-requisitos (bloqueantes)

1. **Autorização explícita do usuário nesta conversa.** Sem ela, pare e pergunte.
2. **Relevância real** (roles.md §6.1): a decisão afeta estrutura, tecnologia, modelo de dados ou contratos entre módulos/sistemas. Decisão trivial não gera ADR.

## Passo a passo

1. Consulte o agente `arquiteto` para estruturar a decisão: alternativas, trade-offs, impacto.
2. Determine o próximo número sequencial: liste `docs/adr/` e use `NNNN` + 1 (4 dígitos).
3. Copie o template `template-adr.md` (nesta pasta) e preencha **todas** as seções — em especial:
   - **Alternativas consideradas e descartadas**: cada uma com o motivo do descarte.
   - **Trade-offs** e **Estratégias para minimizar os trade-offs**: nunca deixar vazios.
   - **ADRs relacionados**: verificar conflito ou substituição de ADRs existentes.
4. Salve em `docs/adr/NNNN-titulo-kebab.md`.
5. Atualize o índice em `docs/adr/README.md`.
6. Atualize `state.md` (seção "ADRs ativos").
7. Se o ADR substituir outro, marque o antigo com status `substituído` apontando para o novo.

## Checklist final

- [ ] Usuário autorizou explicitamente
- [ ] Motivo da decisão está claro para quem chegar depois
- [ ] O que foi considerado e descartado está documentado
- [ ] Trade-offs e mitigações preenchidos
- [ ] Índice do README e state.md atualizados
