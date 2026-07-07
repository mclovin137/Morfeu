# ADR 0006 — Estratégia de testes

- **Status:** aceito
- **Data:** 2026-07-07
- **Task/PRD relacionados:** confronto de padrões de código (2026-07-07); descoberta (parecer QA); materializa-se do E0 ao E12

## Contexto

O usuário propôs "Playwright com estratégia de testes idempotentes com base própria". O parecer do QA concluiu: a intenção (idempotência + isolamento) converge integralmente com a estratégia da descoberta (roles.md §6.7.2) — o ajuste é de escopo: Playwright não pode ser a ferramenta *central* (E2E-cêntrico = feedback de minutos, flakiness estrutural, e os fluxos críticos do domínio não se provam via browser). O usuário aceitou a estratégia consolidada no confronto.

## Escopo

Cobre: pirâmide e papel de cada camada, matriz de banco por camada, regras de idempotência, política de retries/flaky, gates de CI. Não cobre: metodologia de baseline/regressão de performance (skill `benchmark`) nem cenários específicos por task (vivem nos PRDs).

## Decisão

### Pirâmide e papel de cada camada

| Camada | Ferramenta | O que valida | Entra em |
|---|---|---|---|
| Unit (~70%) | `go test`, table-driven, tempo injetado | domínio puro: matriz da saga (passo × resultado), preço, não-conflito de sessão, validações de aggregate | E0 |
| Integração | testcontainers-go (PG/Redis/RabbitMQ), `-race` sempre | SQL do sqlc, **corrida de assento** (N goroutines + barreira → exatamente 1 sucesso + query de invariante), outbox/relay, dedup, DLQ, migrations | E0 |
| Contrato | oapi-codegen + httptest | shape da API vs OpenAPI; códigos de erro estáveis | E1–E2 |
| E2E | Playwright, **2–4 jornadas**, POM, seletores role/test-id | jornada real na SPA (compra completa; disputa em 2 navegadores como *demo* M3; cancelamento; programação backoffice) | E5/E8 |
| Carga | k6 (gerador fora da VM), nightly/manual — **nunca gate de PR** | SLO p95<300ms @ 300 req/s; invariante pós-carga | E12 |

Fluxos críticos se provam na camada certa: disputa de assento = integração (browser não exercita corrida determinística); matriz da saga = unit; TTL/expiração = integração com relógio controlado (proibido `sleep` espalhado).

### Matriz de banco por camada ("base própria" = por construção, não por disciplina)

| Camada | Banco | Isolamento | Reset |
|---|---|---|---|
| Unit | nenhum | fakes nas bordas | n/a |
| Integração | PG efêmero (testcontainers, container por suíte) | dados por UUID por teste; `t.Parallel()` seguro | container morre com a suíte |
| E2E | **stack descartável**: `docker-compose.test.yml` (PG+Redis+RabbitMQ+app dedicados) | dados únicos criados **via API** (`e2e-<runId>-<uuid>`) | `up` → `migrate` → seed mínimo → suíte → `down -v` |
| k6 | VM em **modo load-test** (fakes de gateway/e-mail) | dataset sintético | limpeza pós-run + query de invariante |

Seed determinístico mínimo do E2E: conta do operador + o necessário para autenticar; todo o resto cada teste cria via API. **Proibido banco de teste persistente com limpeza pós-run** (acumula estado, esconde dependência de ordem).

### Regras de idempotência e confiabilidade

1. Nenhum teste depende de outro nem de ordem; `fullyParallel: true`; specs autossuficientes.
2. Arrange **via API, nunca via UI de outro teste**; `storageState` por worker.
3. Asserções web-first; **proibidos** `waitForTimeout` e `waitForLoadState('networkidle')` (este ADR sobrepõe os exemplos da skill `e2e-testing`).
4. **Retries: 0 local; 1 no CI exclusivamente para colher `trace: 'on-first-retry'`; teste que passou no retry = flaky = defeito imediato** (bloqueia). `forbidOnly: true` no CI.
5. Stripe: jornadas de gate rodam contra o **adapter fake**; **uma** jornada nightly/manual contra sandbox real (Stripe CLI). Webhook testado em integração com **fixtures assinadas** (válida/inválida/expirada, replay de event-id, amount divergente, fora de ordem).
6. Teardown global do E2E roda a **query de invariante de assento** (mesma do k6).
7. Fakes **apenas nas bordas** (gateway, e-mail, TMDB, relógio, bus); PG sempre real via testcontainers (o SQL é a lógica — ADR 0002); **proibido mock camada a camada** (testa-se o caso de uso pela API pública do módulo).

### Gates e complementos de CI

- Cobertura: **80% por pacote de domínio** (gate a partir do E2; código gerado excluído), global informativa.
- Job de migrations em PR que toque `migrations/`: `up → down → up` (down falhou = migration sem rollback, §6.10.2) + `sqlc vet`.
- `@axe-core/playwright` nas 3 páginas-chave (cartaz, mapa, checkout), falha em serious/critical; teclado no mapa como asserção de jornada.
- Lint de paridade de chaves i18n (front vs catálogo pt-BR; códigos de erro da API).
- Golden file do JSON do mapa de assentos (unit Go, não snapshot de DOM).
- Smoke E2E read-only pós-deploy contra a URL pública (distinto da suíte de PR).
- Restore mensal do backup em container efêmero + invariantes (E11, fora do CI de PR) — backup não testado não é backup.
- Mutation testing: **fora do CI** (custo > retorno para dev solo); exercício manual pontual em `reserva`/`pedido` pós-E6, se desejado.

## Tecnologias ou padrões envolvidos

go test (table-driven, -race), testcontainers-go, oapi-codegen/httptest, Playwright (+axe), k6, golden files, fixtures de webhook assinadas.

## Benefícios

- Feedback em segundos onde se itera (unit/integração); browser só onde é insubstituível.
- Idempotência **por construção** (bases descartáveis) elimina a classe inteira de bugs de estado acumulado.
- Flaky = defeito mantém a suíte confiável no longo prazo (a alternativa corrói a confiança em ~3 meses).
- Invariante do fluxo crítico verificada em três pontos (integração, teardown E2E, pós-carga).

## Trade-offs

- +30–60s de bootstrap por run de E2E (stack descartável).
- Builds vermelhos mais frequentes no curto prazo (flaky bloqueante).
- Toolchain JS no repo Go (Playwright): package.json, cache de browsers no CI (~2–4h/mês de manutenção com 2–4 jornadas).

## Riscos

- Suíte E2E crescer além das jornadas → limite explícito (2–4) revisado na auditoria.
- Fixtures de webhook desatualizarem com a API do Stripe → smoke nightly contra sandbox real detecta.

## Estratégias para minimizar os trade-offs

- Bootstrap E2E → meio-termo permitido em dev: manter stack de pé e recriar só o volume do PG (`down -v` seletivo); no CI, sempre do zero.
- Builds vermelhos → trace on-first-retry dá diagnóstico imediato; flaky corrigido na hora custa minutos, adiado custa a confiança da suíte.

## Impacto esperado

`docker-compose.test.yml` (E5), config do Playwright e do k6 versionadas; template de critérios de aceite da skill `criar-prd` ganha a seção fixa de plano de testes (cenários G/W/T, invariantes SQL, mapa cenário→camada, idempotência, SLI); CI do E0 nasce com `-race`, cobertura, migrations e `sqlc vet`.

## Alternativas consideradas e descartadas

- **Playwright como estratégia central (proposta original)** — descartada: feedback de minutos mata TDD a 10–12h/semana; corrida de assento e matriz da saga não se provam deterministicamente via browser; flakiness estrutural. Intenção do usuário preservada na camada E2E. Confrontada e aceita a alternativa pelo usuário.
- **Base de teste dedicada persistente com limpeza** — descartada: acumula estado, esconde dependência de ordem, a limpeza vira código às avessas; pior das três interpretações de "base própria".
- **Retries: 2 e "passou, passou"** — descartada: band-aid que corrói a suíte; retry é instrumento de diagnóstico.
- **Pact/contract testing formal** — descartada: consumidor único (o próprio SPA); oapi-codegen + httptest basta.
- **Chaos/toxiproxy e mutation testing no CI** — descartadas: desproporcionais para dev solo; timeout/retry/breaker testam-se com fake lento + relógio injetado.

## ADRs relacionados

- ADR 0002 (sqlc) — PG real nos testes de integração; `sqlc vet`.
- ADR 0003 (camadas) — teste no nível do caso de uso; fakes nas bordas definidas lá.
- ADR 0004 (código) — `-race`, tempo injetado, table-driven.
- ADR 0005 (DDD) — dupla verificação aggregate (unit) + constraint (integração).
