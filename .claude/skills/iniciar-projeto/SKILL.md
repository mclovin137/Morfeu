---
name: iniciar-projeto
description: Descoberta de projeto (kickoff) conforme roles.md §6.15. OBRIGATÓRIA em todo início de projeto, ANTES do roadmap — entrevista profunda com o usuário + análise multiagente para levantar escopo, requisitos funcionais e não funcionais, volumetria, orçamento, atores, público-alvo, dependências, arquitetura, stack e infra. Gera doc.md (mini-UML), lib.md e docs/roadmap.md iniciais.
---

# Iniciar Projeto (descoberta / kickoff)

Cerimônia de descoberta que abre todo projeto: uma entrevista estruturada e profunda com o usuário, seguida de análise dos 5 agentes e de uma rodada de **confronto** — as escolhas do usuário são desafiadas com trade-offs explícitos antes de virarem documento. O resultado alimenta tudo que vem depois: `doc.md` (mini-UML da aplicação), `lib.md` (stack e dependências) e `docs/roadmap.md` (entregas e metodologia).

## Pré-requisitos

1. Você leu `roles.md` (a fonte da verdade).
2. Se `doc.md` já existe com conteúdo real, confirme com o usuário se é uma **re-descoberta** (atualização consciente) — nunca sobrescrever cegamente.

## Postura do condutor (regras da entrevista)

- **Máximo de perguntas.** Use `AskUserQuestion` em blocos de até 4 perguntas por chamada, seguindo o banco `perguntas.md`. Não pule bloco por conta própria: se parecer irrelevante, pergunte ao usuário se é irrelevante.
- **Você não é um coletor passivo de requisitos.** Toda escolha do usuário (stack, arquitetura, infra, escopo) deve ser confrontada quando existir alternativa com melhor trade-off. Apresente benefícios, riscos, custo operacional, complexidade adicionada e manutenção futura (roles.md §1). A palavra final é sempre do usuário — mas ele decide **informado**.
- **Anti-overengineering é critério de confronto nos dois sentidos**: desafie tanto a escolha frágil demais quanto a grandiosa demais para a volumetria declarada.
- **Resposta vaga em volumetria ou orçamento não encerra o assunto.** Ofereça faixas concretas como opções (ex.: "< 100 req/s", "100–1k", "1k–10k") e registre a estimativa assumida.
- **Dúvida sobre lib/framework/versão → Context7** (roles.md §6.12). Nunca presumir versão, recurso ou compatibilidade.
- **Adapte o fluxo**: respostas anteriores moldam as próximas perguntas (sem frontend → pular perguntas de UI; sem dados pessoais → encurtar compliance). Registre o que foi pulado e por quê.

## Passo a passo

### Fase 1 — Entrevista (blocos A–H)

Conduza os blocos de `perguntas.md` **na ordem**, via `AskUserQuestion` (multiSelect quando as opções não forem excludentes; o usuário sempre pode responder livre via "Other"). Mantenha um rascunho do dossiê de respostas no scratchpad conforme avança.

| Bloco | Tema |
|---|---|
| A | Identidade e visão — nome, problema, objetivo, público-alvo, atores |
| B | Escopo funcional — funcionalidades, prioridades, fora de escopo, fluxos críticos |
| C | Requisitos não funcionais — performance, disponibilidade, segurança, compliance |
| D | Volumetria — usuários, requisições, dados, crescimento, picos |
| E | Orçamento e restrições — budget, prazo, equipe, licenças |
| F | Stack — linguagem, framework, banco, cache, mensageria |
| G | Arquitetura e infra — estilo arquitetural, cloud, containers, CI/CD, ambientes |
| H | Dependências e integrações — APIs de terceiros, auth, pagamentos, notificações |
| I | Observabilidade e resiliência (etapa SRE) — ferramentas, métricas/SLOs, monitoramento de logs, alertas, padrões (saga/compensação, retry, circuit breaker, outbox/DLQ, rollback de deploy) |

### Fase 2 — Análise multiagente (paralelo)

Envie o dossiê consolidado aos 4 agentes consultivos **em paralelo**:

| Agente | Análise esperada |
|---|---|
| `arquiteto` | arquitetura proposta vs alternativas, trade-offs, aderência stack↔volumetria, risco de overengineering, candidatas a ADR |
| `sre-devops` | infra mínima viável dentro do orçamento, custo mensal estimado, **plano de observabilidade** (ferramentas, métricas/SLOs, logs, alertas — valida o bloco I), padrões de resiliência (saga/compensação, retry, DLQ) proporcionais à volumetria, CI/CD, ambientes |
| `security` | riscos, dados sensíveis, compliance (LGPD etc.), estratégia de authn/authz, CVEs das libs candidatas |
| `qa` | testabilidade dos requisitos, critérios de aceite mensuráveis, estratégia de testes por camada, riscos de qualidade |

Depois, envie os 4 pareceres ao `backend-dev` para o parecer de **viabilidade**: esforço macro, ordem de construção sugerida, riscos de implementação.

Cada parecer deve terminar com: **recomendações**, **pontos a confrontar com o usuário** e **perguntas adicionais**.

### Fase 3 — Confronto e convergência

1. Consolide os desafios dos agentes e volte ao usuário com a **rodada de confronto** (`AskUserQuestion`): cada divergência vira uma pergunta com as opções e seus trade-offs explicitados — a recomendação dos agentes vem primeiro, marcada "(Recomendada)".
2. Perguntas adicionais dos agentes → rodada extra de entrevista, se necessário.
3. A decisão final é do usuário. Quando ele contrariar a recomendação, registre a escolha **com o trade-off aceito** — isso vai para o `doc.md`, não se perde.
4. Decisões arquiteturais relevantes identificadas → listar como **candidatas a ADR**. NÃO criar ADR nesta cerimônia: recomendação apenas, criação exige autorização explícita (roles.md §6.1).

### Fase 4 — Geração dos artefatos

1. **`doc.md`** (raiz) a partir de `template-doc.md` — o mini-UML da aplicação: visão, atores, contexto, componentes, domínio, fluxos críticos, RNFs, volumetria, stack e trade-offs aceitos (diagramas em Mermaid).
2. **`lib.md`** — versão inicial: stack escolhida e dependências planejadas, cada uma validada no **Context7** (versão estável atual, manutenção, CVEs) e marcada como `planejada` até efetivamente entrar no build (roles.md §6.9).
3. **`docs/roadmap.md`** — versão inicial: épicos/entregas priorizados a partir do escopo, riscos, marcos ligados ao prazo declarado e a metodologia (o fluxo de roles.md §2).
4. **`state.md`** atualizado (descoberta concluída) e registro no vault: `/obsidian-log` + `/obsidian-decide` (roles.md §6.13).
5. Apresente o resumo final ao usuário: decisões-chave, trade-offs aceitos, candidatas a ADR e o próximo passo — `criar-task` sobre o primeiro item do roadmap.

## Checklist final

- [ ] Blocos A–I cobertos (ou irrelevância confirmada pelo usuário, registrada no dossiê)
- [ ] Plano de observabilidade e padrões de resiliência definidos e proporcionais à volumetria (bloco I — anti-overengineering vale aqui também)
- [ ] Os 5 agentes analisaram o dossiê e emitiram parecer
- [ ] Toda divergência confrontada com o usuário; trade-offs aceitos registrados
- [ ] `doc.md` criado com o mini-UML completo (diagramas Mermaid renderizáveis)
- [ ] `lib.md` inicial com stack validada no Context7
- [ ] `docs/roadmap.md` inicial priorizado, com marcos e metodologia
- [ ] Candidatas a ADR listadas (nenhum ADR criado sem autorização — §6.1)
- [ ] `state.md` atualizado e sessão registrada no vault (§6.13)
