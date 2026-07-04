---
name: refinar-task
description: Refinamento multiagente de uma task deste projeto, conforme roles.md §6.14. OBRIGATÓRIO entre a criação da task e o PRD — os 5 agentes analisam a task, debatem entre si e convergem numa conclusão registrada que alimenta o PRD.
---

# Refinar Task (cerimônia multiagente)

Simula uma cerimônia de refinamento: **toda task passa por todos os agentes**, cada um emite parecer do seu domínio, os pareceres são confrontados em debate e a conclusão consolidada é registrada na task — antes de qualquer PRD.

## Pré-requisitos

1. A task existe em `docs/tasks/NNNN-*.md` (senão, rode `criar-task` antes).
2. Você leu `roles.md`, `state.md` e os ADRs ativos.

## Passo a passo

### Rodada 1 — Pareceres independentes (paralelo)

Invoque os 4 agentes consultivos **em paralelo**, cada um recebendo a task completa e a instrução de responder do seu domínio:

| Agente | Parecer esperado |
|---|---|
| `arquiteto` | encaixe na arquitetura/ADRs, padrões aplicáveis, risco de overengineering, volumetria |
| `sre-devops` | impacto em infra/ambiente, observabilidade necessária, gargalos previsíveis |
| `security` | riscos de segurança, dados sensíveis, dependências/CVEs, controles necessários |
| `qa` | estratégia e cenários de teste, critérios de aceite testáveis, riscos de regressão |

Depois, invoque o `backend-dev` com os 4 pareceres para o parecer de **viabilidade de implementação**: esforço, quebra em passos, estimativa de arquivos (≤ 30, roles.md §6.3), dúvidas técnicas.

Cada parecer deve terminar com: **posição** (seguir / seguir com ressalvas / reformular a task), **exigências para o PRD** e **perguntas abertas**.

### Rodada 2 — Debate

1. Confronte os pareceres e liste as **divergências** (ex.: arquiteto pede simplicidade × security pede controle adicional; qa pede e2e × backend-dev aponta custo).
2. Para cada divergência, debata citando as regras: anti-overengineering (roles.md §1), segurança (§6.6), testes (§6.7), observabilidade (§6.8). Quando útil, reenvie o ponto conflitante aos agentes envolvidos para tréplica (máx. 1 rodada extra).
3. Classifique cada divergência como: **consenso alcançado** (com a regra que decidiu) ou **sem consenso → escalar ao usuário** (nunca decidir sozinho questão arquiteturalmente relevante — §6.1).

### Rodada 3 — Conclusão

1. Redija a síntese: escopo confirmado/ajustado, exigências consolidadas para o PRD, riscos priorizados, perguntas escaladas ao usuário.
2. Registre na task (`docs/tasks/NNNN-*.md`) a seção abaixo.
3. Atualize `plan.md` (status: task refinada) e, se o escopo mudou, a própria task e o `docs/tasks/README.md`.
4. Próximo passo do fluxo: `criar-prd` **consumindo as exigências do refinamento**.

## Formato do registro (apêndice na task)

```markdown
## Refinamento — AAAA-MM-DD

### Pareceres
| Agente | Posição | Pontos-chave |
|---|---|---|
| arquiteto | seguir com ressalvas | ... |
| sre-devops | ... | ... |
| security | ... | ... |
| qa | ... | ... |
| backend-dev | ... | ... |

### Debate (divergências e resolução)
1. [divergência] → [consenso: decisão + regra] | [escalado ao usuário]

### Conclusão
- Escopo final: ...
- Exigências para o PRD: [lista que o criar-prd DEVE incorporar]
- Riscos priorizados: ...
- Perguntas ao usuário: [se houver — bloqueiam o PRD até resposta]
```

## Checklist final

- [ ] Os 5 agentes emitiram parecer
- [ ] Divergências explicitadas e cada uma resolvida por regra ou escalada
- [ ] Seção "Refinamento" registrada na task
- [ ] Exigências prontas para o PRD (ou perguntas escaladas ao usuário)
- [ ] plan.md atualizado
