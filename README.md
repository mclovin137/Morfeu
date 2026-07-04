# Morfeu

Projeto criado a partir do template **[Tllm](https://github.com/mclovin137/Tllm)** de governança multiagente (commit `2a86322`, 2026-07-04). O escopo, a stack e o roadmap do Morfeu nascem da **descoberta**: rode `/iniciar-projeto` na primeira sessão — até lá, este repositório carrega apenas a governança (agentes, skills, gates de qualidade e rastreabilidade do ciclo de desenvolvimento).

> **Fonte da verdade:** todas as regras vivem em [`roles.md`](roles.md). Este README é um mapa de entrada — na dúvida, o `roles.md` decide.

## Fluxo padrão

```
descoberta → roadmap → task → refinamento → PRD → implementação → auditoria → push → PR → CI/CD
```

## Como iniciar um projeto (passo a passo)

1. **Use este template** como base do repositório novo e abra o Claude Code na raiz.
2. **Rode a descoberta** — `/iniciar-projeto` (roles.md §6.15). É uma entrevista profunda em 9 blocos (identidade, escopo, requisitos não funcionais, volumetria, orçamento, stack, arquitetura/infra, integrações, observabilidade/resiliência), seguida de análise dos 5 agentes e de uma rodada de **confronto de trade-offs**. Gera:
   - `doc.md` — mini-UML da aplicação (atores, contexto, componentes, domínio, fluxos críticos, RNFs, observabilidade);
   - `lib.md` — stack e dependências planejadas (validadas no Context7);
   - `docs/roadmap.md` — entregas priorizadas, marcos e metodologia.
3. **Revise e aprove** os três artefatos — decisões arquiteturais relevantes viram candidatas a ADR (criadas só com sua autorização, `/criar-adr`).
4. **Quebre em task** — `/criar-task` sobre o item de maior prioridade do roadmap (branch própria, escopo ≤ 30 arquivos — §6.3).
5. **Refine a task** — `/refinar-task` (§6.14): os 5 agentes emitem parecer, debatem e convergem; a conclusão alimenta o PRD.
6. **Crie o PRD** — `/criar-prd` (§6.2): nenhuma implementação começa sem ele.
7. **Implemente** seguindo o PRD e os ADRs, mantendo `plan.md` atualizado. Mudança de schema? Só via `/criar-migration` (§6.10 — índices e anti-full-scan são critério de aprovação). Dependência nova? Registre no `lib.md` **antes** (§6.9).
8. **Audite** — `/auditoria` (§6.4): gate obrigatório de 13 itens. **Nenhum push sem veredito APROVADO.**
9. **Push e PR** — `/fluxo-git` (§6.5): branch `feature/NNNN-nome`, PR rastreável até task e PRD.
10. **Feche o ciclo** — CI verde, `state.md` atualizado e sessão registrada no vault Obsidian (`/obsidian-log` + `/obsidian-decide`, §6.13).

## Agentes (`.claude/agents/`)

Consultivos — analisam e recomendam; quem executa é a sessão principal (exceto o backend-dev, único executor).

| Agente | Para que serve |
|---|---|
| `arquiteto` | decisões estruturais, trade-offs, padrões, volumetria, escolha de tecnologia, anti-overengineering; consultado antes de qualquer ADR |
| `sre-devops` | infra local (docker-compose), observabilidade (métricas, logs, traces), gargalos, performance, resiliência, CI |
| `security` | riscos de segurança, authn/authz, secrets, CVEs de dependências; obrigatório na auditoria |
| `qa` | estratégia e cobertura de testes, idempotência da suíte, critérios de aceite; obrigatório na auditoria |
| `backend-dev` | implementação das tasks seguindo ADRs, PRD e diretrizes dos demais |

Todos os 5 participam da **descoberta** (§6.15) e do **refinamento de toda task** (§6.14).

## Skills do fluxo (`.claude/skills/`)

| Skill | Para que serve | Obrigatória quando |
|---|---|---|
| `iniciar-projeto` | descoberta/kickoff: entrevista + confronto multiagente → `doc.md`, `lib.md`, `roadmap.md` | todo início de projeto |
| `criar-task` | quebrar o roadmap em tasks pequenas e rastreáveis | todo trabalho novo |
| `refinar-task` | cerimônia de refinamento com os 5 agentes | entre a task e o PRD |
| `criar-prd` | detalhar a task antes de implementar | antes de toda implementação |
| `criar-adr` | registrar decisão arquitetural | só com autorização explícita do usuário |
| `criar-migration` | migration versionada com boas práticas de SQL (índices, anti-full-scan, zero-downtime) | toda alteração de schema |
| `auditoria` | gate de qualidade pré-push (13 itens) | antes de todo push |
| `fluxo-git` | branches, commits e PRs padronizados | todo trabalho com git |

## Skills de apoio

**Todas vendorizadas no repositório** (`.claude/skills/`) — clonou, tem tudo: `backend-patterns`, `postgres-patterns`, `docker-patterns`, `database-migrations`, `security-review`, `benchmark`, `browser-qa`, `design-system`, `frontend-patterns`, `jpa-patterns`, `tdd-workflow`, `e2e-testing`, `github-ops`, `git-workflow`, `architecture-decision-records`, `coding-standards`, `api-design`, `error-handling`, `deployment-patterns`, `security-scan`. Origem e créditos em `lib.md`; mapa skill×agente em `roles.md` §4.2. Única exceção: `ui-ux-pro-max` (UI/UX) é plugin, já declarado em `.claude/settings.json` — instala ao confiar no projeto.

## Hooks de sessão (Obsidian)

O repositório registra dois hooks (`.claude/settings.json` → `.claude/hooks/`):

- **SessionEnd** — ao fechar o chat, um agente em background salva o resumo da sessão no vault Obsidian (nota em `Dev Logs/` + propagação para nota do projeto, daily note e log de operações). Sessões triviais (< 2 prompts) não geram nota.
- **SessionStart** — ao abrir o chat, injeta no contexto a nota do projeto vinda do vault (resumo, atividade recente, decisões).

**Em outra máquina**: após clonar, defina `OBSIDIAN_VAULT_PATH` no `env` do seu `~/.claude/settings.json` apontando para o vault. Sem essa variável os hooks ficam **inertes** (sem erro). Para desligar o resumo automático: `OBSIDIAN_SESSION_SUMMARY_ENABLED=0`.

## Dependências

Registro obrigatório em [`lib.md`](lib.md) **antes** de qualquer dependência entrar (§6.9) — com justificativa, alternativas e CVEs. Hoje o projeto não tem dependências de runtime; a única ferramenta é o MCP **Context7** (`.mcp.json`), consultado sempre que houver dúvida de lib/framework/versão (§6.12) — nunca presumir.

## Arquivos de controle

| Arquivo | Conteúdo |
|---|---|
| `roles.md` | todas as regras — fonte da verdade |
| `doc.md` | mini-UML da aplicação (nasce na descoberta) |
| `lib.md` | registro de dependências e versões |
| `state.md` | estado atual + histórico do projeto |
| `plan.md` | plano vivo da task em andamento |
| `docs/` | `roadmap.md`, `adr/`, `prd/`, `tasks/` (numeração `NNNN`, índice em cada pasta) |

## Invariantes (nunca violar)

1. Nenhum projeto sem **descoberta**; nenhuma implementação sem **PRD**; nenhum PRD sem **refinamento**; nenhuma task sem **branch própria**.
2. Nenhum push sem **auditoria APROVADA**.
3. **ADR** só com autorização explícita do usuário.
4. Toda dependência registrada em **`lib.md`** antes de usar.
5. `plan.md`/`state.md` atualizados; fim de task registrado no **vault Obsidian** (§6.13).
