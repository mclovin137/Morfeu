# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Governança

Este projeto segue governança multiagente. **`roles.md` é a fonte da verdade** de todas as regras — leia-o antes de qualquer trabalho. Agentes em `.claude/agents/`, skills do fluxo em `.claude/skills/`.

Fluxo padrão: `descoberta → roadmap → refinamento do épico → task → PRD → implementação → gate pré-push → push → PR → CI + revisão → merge`.

## Invariantes (nunca violar, mesmo sem ler roles.md)

1. **Nenhum projeto sem descoberta** (`iniciar-projeto` — entrevista + confronto multiagente com quadro de trade-offs e prioridades; gera `doc.md`, `lib.md`, `roadmap.md`, direção de design e ADRs iniciais autorizados); **nenhuma implementação sem PRD** (`criar-prd`, just-in-time ao abrir a task); nenhum PRD sem **refinamento multiagente do épico** (`refinar-task` — os 5 agentes debatem o épico uma única vez; task trivial dispensa, roles.md §6.14.6); nenhuma task sem branch própria e escopo ≤ 30 arquivos (`criar-task`).
2. **Gate híbrido de auditoria** (roles.md §6.4): pré-push exige scan de secrets limpo; **nenhum merge sem CI verde + revisão de julgamento aprovada**. Transição: enquanto o CI (E0c) não existir, nenhum push sem a skill `auditoria` APROVADA.
3. **ADR só com autorização explícita do usuário** (`criar-adr`).
4. Toda dependência nova registrada em `lib.md` antes de usar; dúvida sobre lib/framework/versão → consultar **Context7**, nunca presumir.
5. Manter `plan.md` atualizado durante a task e `state.md` ao concluir; o hook `SessionEnd` registra a sessão no vault automaticamente — `/obsidian-decide` apenas para decisões relevantes (roles.md §6.13).

> Nota: `plan.md` na raiz é o plano da task corrente **do projeto** — não confundir com o plan mode do Claude Code.

## Scaffold atual (placeholder)

O template é **stack-agnóstico**. O scaffold Java foi removido em 2026-07-03 — resta apenas o `pom.xml` placeholder (sem dependências nem plugins) e `src/` vazio. A stack real nasce da descoberta (`iniciar-projeto`, roadmap item 2).

- No WSL há JDK 25 (Corretto/SDKMAN), mas **não há `mvn` no PATH nem wrapper** — builds Maven são feitos pelo IntelliJ no Windows.
- Testes exigem adicionar framework ao `pom.xml` primeiro (registrar em `lib.md`).
