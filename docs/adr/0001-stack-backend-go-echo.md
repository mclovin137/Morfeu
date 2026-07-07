# ADR 0001 — Stack backend: Go + Echo (monolito modular, binário único)

- **Status:** aceito
- **Data:** 2026-07-07
- **Task/PRD relacionados:** descoberta (`doc.md`), roadmap E0

## Contexto

O template Tllm é stack-agnóstico e o repositório nasceu com um scaffold JVM placeholder (`pom.xml` vazio). A descoberta (2026-07-06) definiu o Morfeu como sistema de venda de ingressos de cinema — domínio transacional com concorrência de inventário, saga de pagamento e alvo de carga simulada de 100–1k req/s — desenvolvido por dev solo com objetivo declarado de **aprender Go**. Era preciso formalizar a stack antes do walking skeleton (E0).

## Escopo

Cobre: linguagem do backend, framework HTTP, topologia de processo. Não cobre: camada de dados (ADR 0002), fronteiras internas (ADR 0003), frontend (React + Vite, decisão registrada em `doc.md` §9 sem ADR — sem disputa).

## Decisão

**Backend em Go com o framework Echo v4, como monolito modular num binário único com flag `-mode=api|worker|all`.**

Um único processo no MVP (`-mode=all`); o worker (consumidor RabbitMQ, relay do outbox, sweeper) é separável por flag sem refatoração quando o teste de carga justificar.

## Tecnologias ou padrões envolvidos

Go (versão estável corrente), Echo v4 (≥ 4.15.0 — CVE-2022-40083 corrigida em 4.9.0), build `CGO_ENABLED=0` multi-arch (amd64 dev / arm64 VM Oracle), monolito modular package-by-domain.

## Benefícios

- Objetivo de aprendizado atendido diretamente → o usuário escolheu Go para dominar a linguagem.
- Cross-compile trivial para a VM ARM (Oracle A1) sem QEMU → pipeline CI simples e barato.
- Footprint pequeno (binário estático, ~100–300 MB RAM sob carga) → cabe com folga no free tier ao lado de PG/Redis/RabbitMQ/observabilidade.
- Echo: middleware maduro (JWT, rate limit), performance adequada com folga para 1k req/s, validado ativo no Context7 (v4.15.0).
- Binário único `-mode` → um deploy, uma config, observabilidade unificada; separação futura a custo ~zero.

## Trade-offs

- Rampa de aprendizado: o dev vem de JVM — as primeiras 4–6 semanas são infladas (estimativa já embutida no roadmap de 7–9 meses).
- Abandona o ambiente JVM já preparado na máquina (JDK 25, IntelliJ/Maven).
- Echo é menos "padrão de mercado brasileiro" que Spring para portfólio Java — mas o portfólio alvo é Go.
- Risco nº 1 do roadmap: "escrever Java em Go" (mitigado pelos ADRs 0003–0005).

## Riscos

- Dialeto não idiomático por hábito JVM → probabilidade alta, impacto médio (mitigado por ADR 0004 + lint).
- Framework HTTP trocado no futuro → probabilidade baixa; handlers finos (ADR 0003) limitam o raio da troca.

## Estratégias para minimizar os trade-offs

- Rampa → walking skeleton cedo (E0) consolida o padrão handler→service→sqlc num módulo simples antes dos críticos.
- Java-em-Go → wiring manual no `main`, interfaces só com segundo implementador real, `.golangci.yml` (ADR 0004), Context7 antes de qualquer API de lib (roles.md §6.12).
- Acoplamento ao Echo → regra do ADR 0003: nenhuma regra de negócio em handler; `echo.Context` não atravessa a fronteira do service.

## Impacto esperado

Remoção do `pom.xml` placeholder (item 0 do roadmap); `go.mod` no E0; CI com runner ARM64 (`ubuntu-24.04-arm`); toda a estrutura de módulos do ADR 0003 pressupõe esta decisão.

## Alternativas consideradas e descartadas

- **Java + Spring Boot** — descartada porque o objetivo declarado é aprender Go; footprint JVM pesa no free tier; o ambiente JVM local (builds só via IntelliJ Windows) era atrito adicional.
- **Kotlin** — descartada: mesma JVM, curva extra sem atender o objetivo Go.
- **Node.js/TypeScript** — descartada: unificaria com o frontend, mas abandona o objetivo de aprendizado e perde o perfil de concorrência/performance desejado para o exercício de carga.
- **Gin / chi / stdlib puro** — descartadas em favor do Echo pela preferência do usuário na descoberta (validada sem CVEs abertas); handlers finos tornam a diferença entre eles marginal.
- **Microserviços** — descartada: custo operacional absurdo para dev solo numa VM; nada na volumetria justifica (roles.md §1).

## ADRs relacionados

- ADR 0002 (camada de dados) — depende deste.
- ADR 0003 (fronteiras de módulos) — materializa a topologia interna deste.
- ADR 0004 (padrões de código) — mitiga o risco Java-em-Go deste.
