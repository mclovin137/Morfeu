# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Governança

Este projeto segue governança multiagente. **`roles.md` é a fonte da verdade** de todas as regras — leia-o antes de qualquer trabalho. Agentes em `.claude/agents/`, skills do fluxo em `.claude/skills/`.

Fluxo padrão: `descoberta → roadmap → refinamento do épico → task → PRD → implementação → gate pré-push → push → PR → CI + revisão → merge`.

## Invariantes (nunca violar, mesmo sem ler roles.md)

1. **Nenhum projeto sem descoberta** (`iniciar-projeto` — entrevista + confronto multiagente com quadro de trade-offs e prioridades; gera `doc.md`, `lib.md`, `roadmap.md`, direção de design e ADRs iniciais autorizados); **nenhuma implementação sem PRD** (`criar-prd`, just-in-time ao abrir a task); nenhum PRD sem **refinamento multiagente do épico** (`refinar-task` — os 5 agentes debatem o épico uma única vez; task trivial dispensa, roles.md §6.14.6); nenhuma task sem branch própria e escopo ≤ 30 arquivos (`criar-task`).
2. **Gate híbrido de auditoria** (roles.md §6.4): pré-push exige scan de secrets limpo (gitleaks local); **nenhum merge sem CI verde + revisão de julgamento aprovada** (skill `auditoria`, passe único sobre o diff do PR).
3. **ADR só com autorização explícita do usuário** (`criar-adr`).
4. Toda dependência nova registrada em `lib.md` antes de usar; dúvida sobre lib/framework/versão → consultar **Context7**, nunca presumir.
5. Manter `plan.md` atualizado nos 3 checkpoints da task (abertura, meio/desvio, fechamento — roles.md §6.11) e `state.md` ao concluir; o hook `SessionEnd` registra a sessão no vault automaticamente — `/obsidian-decide` apenas para decisões relevantes (roles.md §6.13).

> Nota: `plan.md` na raiz é o plano da task corrente **do projeto** — não confundir com o plan mode do Claude Code.

## Stack e ambiente

Stack Go 1.25 + Echo + sqlc/pgx + Redis (definida na descoberta; ADRs 0001–0007). **Antes de rodar build/test/lint, leia `docs/ambiente-dev.md`** — comandos canônicos do WSL2 (Go fora do PATH, `-race` e lint só via container, `GOTOOLCHAIN=auto` p/ sqlc). Não redescobrir o ambiente por tentativa e erro.
