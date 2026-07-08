---
name: iniciar-projeto
description: Descoberta de projeto (kickoff) conforme roles.md §6.15. OBRIGATÓRIA em todo início de projeto, ANTES do roadmap — entrevista profunda com o usuário + análise multiagente para levantar escopo, requisitos funcionais e não funcionais, volumetria, orçamento, atores, público-alvo, dependências, arquitetura, stack, infra e design. Confronta o usuário com o quadro completo de trade-offs para escolher caminhos e prioridades. Gera doc.md (mini-UML), lib.md, docs/roadmap.md, direção de design (docs/design/), skills de apoio da stack e ADRs iniciais (com autorização).
---

# Iniciar Projeto (descoberta / kickoff)

Cerimônia de descoberta que abre todo projeto: uma entrevista estruturada e profunda com o usuário, seguida de análise dos 5 agentes e de uma rodada de **confronto** — as escolhas do usuário são desafiadas com o quadro completo de trade-offs antes de virarem documento, e é o usuário quem escolhe os caminhos e define as prioridades. O resultado alimenta tudo que vem depois: `doc.md` (mini-UML da aplicação), `lib.md` (stack e dependências), `docs/roadmap.md` (entregas e metodologia), a direção de design em `docs/design/`, as skills de apoio da stack vendorizadas e os ADRs iniciais (criados com autorização explícita — §6.1).

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

### Fase 1 — Entrevista (blocos A–J)

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
| J | Design e identidade visual — personalidade, referências, tom/voz, tema, acessibilidade, densidade, design system (só com interface; alimenta a direção de design da Fase 4) |

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

### Fase 3 — Confronto, trade-offs e prioridades

1. Consolide os desafios dos agentes num **quadro único com TODOS os trade-offs relevantes** (decisão × opções × o que se ganha/perde em cada uma × recomendação) e apresente ao usuário antes de qualquer escolha.
2. Volte ao usuário com a **rodada de confronto** (`AskUserQuestion`): cada divergência vira uma pergunta com as opções e seus trade-offs explicitados — a recomendação dos agentes vem primeiro, marcada "(Recomendada)". O usuário escolhe **o caminho de cada decisão**.
3. Feche a rodada com a **definição de prioridades**: o usuário ordena as entregas/épicos (o que é inegociável no MVP, o que pode esperar) — isso alimenta diretamente o roadmap.
4. Perguntas adicionais dos agentes → rodada extra de entrevista, se necessário.
5. A decisão final é do usuário. Quando ele contrariar a recomendação, registre a escolha **com o trade-off aceito** — isso vai para o `doc.md`, não se perde.
6. Decisões arquiteturais relevantes identificadas → listar como **propostas de ADR inicial** (viram ADRs na Fase 5, mediante autorização).

### Fase 4 — Skills e design

1. **Levantamento de skills** (roles.md §6.15.7): com a stack definida na entrevista, busque skills/agentes de apoio para ela (repositórios conhecidos, marketplace, pools já avaliados no §4.3) — avalie aderência real, proponha ao usuário e **vendorize as aprovadas** em `.claude/skills/`, registrando origem/commit no `lib.md`.
2. **Direção de design** (bloco J): se o projeto tem interface, gere a direção de design inicial em `docs/design/` — paleta, tipografia, tom, tema, referências aceitas — validando com o usuário (iterar com protótipo leve quando fizer sentido). Sem interface: registrar N/A justificado.

### Fase 5 — Geração dos artefatos

1. **`doc.md`** (raiz) a partir de `template-doc.md` — o mini-UML da aplicação: visão, atores, contexto, componentes, domínio, fluxos críticos, RNFs, volumetria, stack e trade-offs aceitos (diagramas em Mermaid).
2. **`lib.md`** — versão inicial: stack escolhida e dependências planejadas, cada uma validada no **Context7** (versão estável atual, manutenção, CVEs) e marcada como `planejada` até efetivamente entrar no build (roles.md §6.9).
3. **`docs/roadmap.md`** — versão inicial: épicos/entregas priorizados **conforme as prioridades definidas pelo usuário na Fase 3**, riscos, marcos ligados ao prazo declarado e a metodologia (o fluxo de roles.md §2).
4. **ADRs iniciais** (roles.md §6.15.6): apresente as propostas de ADR das decisões estruturais do confronto e peça **autorização explícita** para criá-las em bloco (`criar-adr`). Sem autorização, ficam como candidatas listadas no `doc.md` — a descoberta não trava por isso.
5. **`state.md`** atualizado (descoberta concluída); registro no vault: o hook `SessionEnd` cobre a sessão, rode `/obsidian-decide` para as decisões da descoberta (roles.md §6.13).
6. Apresente o resumo final ao usuário: decisões-chave, trade-offs aceitos, prioridades, ADRs criados/candidatas, direção de design, skills vendorizadas e o próximo passo — `refinar-task` sobre o primeiro épico do roadmap.

## Checklist final

- [ ] Blocos A–J cobertos (ou irrelevância confirmada pelo usuário, registrada no dossiê)
- [ ] Plano de observabilidade e padrões de resiliência definidos e proporcionais à volumetria (bloco I — anti-overengineering vale aqui também)
- [ ] Os 5 agentes analisaram o dossiê e emitiram parecer
- [ ] Quadro consolidado de **todos** os trade-offs apresentado; toda divergência confrontada; escolhas e trade-offs aceitos registrados
- [ ] **Prioridades definidas pelo usuário** e refletidas no roadmap
- [ ] Skills de apoio da stack levantadas e (as aprovadas) vendorizadas com origem no `lib.md`
- [ ] Direção de design inicial em `docs/design/` (ou N/A justificado — sem interface)
- [ ] `doc.md` criado com o mini-UML completo (diagramas Mermaid renderizáveis)
- [ ] `lib.md` inicial com stack validada no Context7
- [ ] `docs/roadmap.md` inicial priorizado, com marcos e metodologia
- [ ] **ADRs iniciais criados com autorização explícita** (ou candidatas listadas no `doc.md` — §6.1)
- [ ] `state.md` atualizado e decisões registradas no vault (§6.13)
