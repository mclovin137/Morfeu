---
name: refinar-task
description: Refinamento multiagente por ÉPICO deste projeto, conforme roles.md §6.14. OBRIGATÓRIO antes do primeiro PRD de cada épico — os 5 agentes analisam o épico inteiro uma única vez, debatem e produzem exigências por task. Tasks triviais dispensam; tasks padrão consomem o refinamento do épico já registrado.
---

# Refinar Épico (cerimônia multiagente, 1× por épico)

Simula uma cerimônia de refinamento: **cada épico passa uma única vez por todos os agentes**, que emitem parecer do seu domínio sobre o épico inteiro, debatem as divergências e produzem **exigências por task planejada** — antes de qualquer PRD do épico. Tasks subsequentes do mesmo épico **não** repetem a cerimônia (roles.md §6.14.6).

## Antes de invocar — a cerimônia é necessária?

1. Já existe `docs/refinamentos/ENN-*.md` para o épico? → **Não refinar de novo**; siga ao `criar-prd` consumindo as exigências da task.
2. A task é **trivial** (docs, chore, fix pequeno, sem decisão de projeto)? → dispensa refinamento (roles.md §6.14.6).
3. Épico **XL** (ex.: E6 — saga) pode ganhar rodada complementar focada numa task crítica — **só com aval do usuário**.

## Pré-requisitos

1. O épico existe no `docs/roadmap.md` com suas tasks previstas.
2. Você leu `roles.md`, `state.md` e identificou os ADRs que tocam o tema.

## Passo a passo

### Rodada 0 — Brief pré-digerido (dieta de contexto, §6.14.3)

Monte **um único brief** na sessão principal e envie o MESMO texto a todos os agentes — eles **não devem reler o repositório**:

- O épico (descrição do roadmap) e as tasks previstas com escopo estimado.
- Decisões relevantes do `doc.md` (domínio, fluxos críticos, RNFs que tocam o épico).
- Trechos dos ADRs pertinentes (não o arquivo inteiro).
- Estado atual resumido (o que já existe implementado que o épico toca).

**Modelo:** pareceres consultivos podem rodar em modelo menor (parâmetro `model` do Agent tool — ex.: `sonnet`); a sessão principal e a implementação permanecem no modelo principal.

### Rodada 1 — Pareceres independentes (paralelo)

Invoque os 4 agentes consultivos **em paralelo**, cada um recebendo o brief e a instrução de responder do seu domínio **por task do épico**:

| Agente | Parecer esperado |
|---|---|
| `arquiteto` | encaixe na arquitetura/ADRs, padrões aplicáveis, drift arquitetural, risco de overengineering, volumetria |
| `sre-devops` | impacto em infra/ambiente, observabilidade necessária, gargalos previsíveis |
| `security` | riscos de segurança, dados sensíveis, dependências/CVEs, controles necessários |
| `qa` | estratégia e cenários de teste, critérios de aceite testáveis, riscos de regressão |

Depois, invoque o `backend-dev` com os 4 pareceres para o parecer de **viabilidade de implementação**: esforço, quebra em tasks/passos, estimativa de arquivos por task (≤ 30, roles.md §6.3), dúvidas técnicas.

Cada parecer deve terminar com: **posição** (seguir / seguir com ressalvas / reformular), **exigências para os PRDs (por task)** e **perguntas abertas**.

### Rodada 2 — Debate

1. Confronte os pareceres e liste as **divergências** (ex.: arquiteto pede simplicidade × security pede controle adicional; qa pede e2e × backend-dev aponta custo).
2. Para cada divergência, debata citando as regras: anti-overengineering (roles.md §1), segurança (§6.6), testes (§6.7), observabilidade (§6.8). Quando útil, reenvie o ponto conflitante aos agentes envolvidos para tréplica (máx. 1 rodada extra).
3. Classifique cada divergência como: **consenso alcançado** (com a regra que decidiu) ou **sem consenso → escalar ao usuário** (nunca decidir sozinho questão arquiteturalmente relevante — §6.1).

### Rodada 3 — Conclusão

1. Redija a síntese: escopo do épico confirmado/ajustado, **exigências consolidadas por task**, riscos priorizados, perguntas escaladas ao usuário.
2. Registre em `docs/refinamentos/ENN-nome-kebab.md` (formato abaixo) e atualize o índice `docs/refinamentos/README.md`.
3. Atualize `plan.md` (épico refinado) e, se o escopo mudou, o `docs/roadmap.md` (com aval do usuário).
4. Próximo passo do fluxo: `criar-task` (se a task não existe) → `criar-prd` **consumindo as exigências da task no refinamento** — sem nova rodada de agentes (§6.2.5).

## Formato do registro (`docs/refinamentos/ENN-nome.md`)

```markdown
# Refinamento — Épico ENN (nome) — AAAA-MM-DD

## Pareceres
| Agente | Posição | Pontos-chave |
|---|---|---|
| arquiteto | seguir com ressalvas | ... |
| sre-devops | ... | ... |
| security | ... | ... |
| qa | ... | ... |
| backend-dev | ... | ... |

## Debate (divergências e resolução)
1. [divergência] → [consenso: decisão + regra] | [escalado ao usuário]

## Conclusão
- Escopo final do épico: ...
- Riscos priorizados: ...
- Perguntas ao usuário: [se houver — bloqueiam os PRDs afetados até resposta]

## Exigências por task
### Task (a) — nome
- [lista que o criar-prd DEVE incorporar]
### Task (b) — nome
- ...
```

## Checklist final

- [ ] Brief pré-digerido enviado (agentes não releram o repo)
- [ ] Os 5 agentes emitiram parecer cobrindo todas as tasks do épico
- [ ] Divergências explicitadas e cada uma resolvida por regra ou escalada
- [ ] `docs/refinamentos/ENN-*.md` registrado + índice atualizado
- [ ] Exigências por task prontas para os PRDs (ou perguntas escaladas ao usuário)
- [ ] plan.md atualizado
