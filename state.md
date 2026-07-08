# state.md — Estado Atual do Projeto

Atualizado ao final de cada task e antes de cada PR (regras em `roles.md` §6.11).

## Estado atual

**Descoberta concluída (2026-07-06).** Morfeu definido: **venda de ingressos de cinema online** (portfólio/aprendizado, dev solo, 10–12h/sem, horizonte 7–9 meses). Stack: Go + Echo, sqlc + pgx/v5, PostgreSQL (fonte da verdade, inclusive trava de assento), Redis (cache), RabbitMQ (saga do checkout), React + Vite, VM Oracle Always Free (PAYG), observabilidade self-hosted, k6. Artefatos gerados: `doc.md` (mini-UML), `lib.md` (stack `planejada`, validada Context7/OSV), `docs/roadmap.md` (E0–E12, marcos M1–M5).

- **Última task concluída:** nenhuma (descoberta não é task)
- **Task atual:** 0001 (App Skeleton Go — E0a) — status: não iniciada → próximo: `/refinar-task`
- **Branch atual:** `feature/0001-app-skeleton-go`
- **PRD atual:** nenhum (será criado após refinamento)
- **ADRs ativos:** 0001 (Go+Echo) · 0002 (sqlc+pgx) · 0003 (fronteiras/camadas) · 0004 (padrões de código Go) · 0005 (DDD tático + patterns) · 0006 (estratégia de testes) — todos aceitos em 2026-07-07 com autorização explícita. Candidatas restantes (trava de assento, RabbitMQ, saga) nascem nos refinamentos E4/E6.

## Últimas decisões relevantes

- 2026-07-07 — **Skills Go vendorizadas + docs/lib/ formalizada** (decisão do usuário após levantamento de skills/agentes p/ a stack): 8 skills de Go idiomático de samber/cc-skills-golang (`concurrency`, `context`, `error-handling`, `testing`, `database`, `safety`, `lint`, `project-layout`) + `golang-observability-opentelemetry` de bobmatnyc/claude-mpm-skills (gatilho do §4.3 disparado pela definição da stack); `jpa-patterns` removida (órfã pós-Go); **ADRs 0003–0005 prevalecem sobre skills em conflito**; nenhum agente novo (golang-pro externo colidiria com backend-dev); RabbitMQ/Stripe: sem skill — docs/lib + Context7 agora, MCP oficial Stripe como candidata p/ E6. Pasta `docs/lib/` (docs técnicos das dependências, base Context7) formalizada em roles.md §6.9.4/§7.
- 2026-07-07 — **Identidade visual "A Sala Escura" aprovada** (exploração de design, fora do fluxo de task — sem código de produção). Paleta de 7 tokens (noite azul-violeta + tungstênio + papoula), Fraunces/Schibsted Grotesk/Spline Sans Mono, 5 regras de movimento, ClassInd com cores oficiais. Registrada em `docs/design/` (spec + protótipo autocontido). Variação clara de login ("saguão") explorada e **descartada pelo usuário**. Orienta os PRDs de E0c/E5/E8; fontes self-hosted a registrar no `lib.md` quando a SPA nascer.
- 2026-07-07 — **6 ADRs criados** após confronto multiagente sobre a proposta do usuário (Object Calisthenics + Strategy/Factory/State + Controller→Actor→Resolver→Service→DAO + DDD crítico + Playwright/base própria). Usuário aceitou as 4 recomendações de consenso: camadas mapeadas idiomaticamente (handler→service→sqlc; Actor=orquestrador da saga; Resolver=composition root; DAO=sqlc+ports), calisthenics adaptado via `.golangci.yml`, **PT no domínio + EN técnico**, Playwright como topo da pirâmide com bases descartáveis. Consulta CAP do usuário respondida: PostgreSQL confirmado (domínio CP; cartaz em cache é o lado eventual deliberado).
- 2026-07-06 — **Descoberta completa** (entrevista A–I + 5 pareceres + confronto). Decisões-chave e trade-offs aceitos registrados em `doc.md` §10; destaques: trava de assento no **PG** (usuário reverteu Redis após consenso 4/4), **backup diário revertido** (era "sem backup"), auditoria enxuta (era "completa"), fluxo único convidado+conta, e-mail pós-pivô na saga, circuit breaker só no gateway, **Stripe** sandbox, repo **público**, TTL da trava 10 min + 1 extensão, cancelamento até 2h antes, check-in QR fora do MVP, Tempo entra com a saga, alertas via **Discord**, repo será movido p/ **WSL ext4**, conta Oracle nova em **PAYG**.
- 2026-07-04 — Config sensível: `.env` (CONTEXT7_API_KEY + Obsidian) gitignored; `.env.example` versionado; `.mcp.json` com `${CONTEXT7_API_KEY:-}`; chave também em `.claude/settings.local.json`.
- 2026-07-04 — Projeto criado a partir do template Tllm (governança multiagente completa herdada).

## Pendências técnicas

- **Hardening do hook `obsidian-session-end.sh`** (achado da auditoria de 2026-07-07, severidade média, não bloqueante): trocar `--dangerously-skip-permissions` por allowlist escopada (`--allowedTools "Read Write Edit Glob Grep"`) e mover o log de `/tmp` para diretório do usuário; junto, hardening da config `.claude/` (permissions block no settings.json) — candidata a task junto com o enforcement do gate §6.4.5.
- **Item 0 do roadmap (pré-bootstrap):** mover repo p/ WSL ext4; criar conta Oracle PAYG + VM A1 (semana 1 — capacidade é loteria); remover `pom.xml` placeholder na task de bootstrap. ~~Criar repo GitHub público e primeiro push~~ → feito em 2026-07-07 (auditoria aprovada; secret scanning + push protection habilitados na criação).
- **Rotacionar a CONTEXT7_API_KEY** (exposta no chat da sessão de descoberta — severidade baixa; dashboard context7.com) e atualizar `.env` + `.claude/settings.local.json`.
- Domínio para e-mail transacional (Resend/Brevo exigem domínio verificado) — pendente; demo nasce em subdomínio gratuito.
- Preencher steps reais do `.github/workflows/ci.yml` no E0.
- Definições que ficaram para o refinamento: intervalo de limpeza entre sessões; alvo do SLI de checkout fim-a-fim (pós-baseline); detalhes do template JSON de sala.

## Riscos conhecidos

- Governança convencional (sem enforcement técnico por hooks).
- Top 5 riscos do roadmap em `docs/roadmap.md` (Java-em-Go, saga inflada, mapa de assentos, backoffice sumidouro, capacidade A1).
- Repo ainda em `/mnt/c` até o item 0 (I/O lento em builds Go/Vite).

## Próximos passos

1. Item 0 do roadmap — pré-bootstrap de ambiente (WSL ext4, Oracle, GitHub público).
2. `/criar-task` sobre a fatia (a) do E0 (walking skeleton) → `/refinar-task` → `/criar-prd` → implementação.
3. No refinamento do E0: propor criação dos ADRs candidatos 1–2 (stack Go+Echo; sqlc+pgx) com autorização do usuário.

## Histórico resumido

| Data | Evento |
|------|--------|
| 2026-07-07 | **Identidade visual "A Sala Escura" aprovada e registrada** em `docs/design/` (paleta, tipografia, movimento, voz + protótipo navegável de login/cadastro e home). |
| 2026-07-07 | **ADRs 0001–0006 criados** (stack, dados, camadas, código, DDD/patterns, testes) após confronto multiagente da proposta de padrões do usuário. |
| 2026-07-06 | **Descoberta concluída**: entrevista A–I, 5 pareceres, confronto (14 decisões), `doc.md` + `lib.md` + `docs/roadmap.md` gerados. |
| 2026-07-04 | Config: `.env`/`.env.example`/`.mcp.json` (Context7 key + Obsidian); chave protegida do git. |
| 2026-07-04 | Projeto Morfeu criado a partir do template Tllm `2a86322` (28 skills, 5 agentes, hooks de sessão Obsidian, arquivos de controle). |
