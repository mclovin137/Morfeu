# Task 0002 — Outbox + RabbitMQ + worker idempotente (E0b)

- **Data:** 2026-07-12
- **Status:** em andamento
- **Branch:** `feature/0002-outbox-rabbitmq`
- **PRD:** docs/prd/0002-outbox-rabbitmq.md — criar antes de implementar (just-in-time, §6.2.5, consumindo o refinamento do E0)
- **Item do roadmap:** E0 (walking skeleton), entrega **E0b** — ordem aprovada pelo usuário em 2026-07-11 (conformidade ✅ → E0c-CI ✅ → **E0b** → E0c-CD → E0d); exigências em `docs/refinamentos/E0-walking-skeleton.md` §"Task E0b"; **ADR 0007 (Mensageria: RabbitMQ)** aceito em 2026-07-11

> Numeração: o slot 0002 estava reservado para a E0b desde o refinamento (a branch `feature/0002-outbox-rabbitmq` foi criada na época do rascunho original, depois reescopado).

## Objetivo

Provar o caminho assíncrono do walking skeleton: um evento de domínio real (`catalogo.filme_criado`) atravessa **outbox transacional → relay com publisher confirms → RabbitMQ → consumidor idempotente com dedup na mesma TX**, com DLQ e graceful shutdown — a espinha dorsal de mensageria que a saga do checkout (E6) vai reutilizar.

## Escopo

- **Outbox (plataforma):** migration da tabela `outbox_events`; `Enqueue(tx, evt)`; relay com polling 500ms–1s, `SELECT ... FOR UPDATE SKIP LOCKED`, publisher confirms **síncronos** (`PublishWithDeferredConfirm` + `WaitContext`, timeout ~5s) — `published_at` só após confirm; reconexão manual (`NotifyClose` + backoff exponencial com teto). Relay roda **só em `-mode=worker`/`all`** (nunca no `api`).
- **UM evento, UMA queue:** `catalogo.filme_criado`, emitido por **subcomando CLI** do binário (stdlib `flag`, sem cobra, sem endpoint HTTP) que insere filme + evento **na mesma TX** via service real (`WithTx`).
- **Envelope:** `message_id` (uuid), `event_type`, `aggregate_id`, `occurred_at`, `payload`, `traceparent` em header AMQP (propagator W3C via pacote leve `go.opentelemetry.io/otel/propagation`; SDK completo só no E0d).
- **Consumidor (runtime em plataforma, handler no domínio):** dedup por `processed_messages` **na mesma TX do efeito** (wrapper de plataforma); payload malformado → DLQ sem derrubar o processo; graceful shutdown por ctx (SIGTERM, timeout 30s, sem goroutine órfã).
- **Topologia (ADR 0007):** exchange topic `morfeu.events`, quorum queue, `x-delivery-limit=3`, DLQ declarada; declaração idempotente no startup de todo processo que toca o broker, parâmetros idênticos (evita 406 PRECONDITION_FAILED).
- **Compose:** RabbitMQ `3.13-management-alpine` com `vm_memory_high_watermark` absoluto (~768MB) + `disk_free_limit` (~512MB); 15672 só em `127.0.0.1`; healthcheck `rabbitmq-diagnostics -q ping` + `condition: service_healthy` no worker; credenciais via `.env` não versionado (nunca `guest/guest`).
- **Fronteiras (ADR 0003 + depguard):** nenhum domínio importa `amqp091-go`; ownership excepcional das tabelas `outbox_events`/`processed_messages` pela plataforma documentado no PRD.
- **Segurança:** logs do relay/consumer sem payload completo em `info` (só `event_type`, `aggregate_id`, `message_id`); outbox nunca carrega segredo/PII no payload; `amqp091-go` registrado no `lib.md` com versão pinada + CVE checada.
- **Observabilidade (fonte do dado, sem dashboard):** contador/query de `outbox_pendentes` exposto internamente para o E0d consumir.
- **Health:** readiness ganha RabbitMQ sem quebrar o contrato da E0a.

## Fora de escopo

- Segundo evento (`FilmUpdated`) e segunda queue — cortados pelo refinamento (anti-overengineering).
- Tooling de replay da DLQ (nasce com a saga, E6).
- LISTEN/NOTIFY no lugar do polling (otimização futura registrada, não implementada).
- Limpeza/arquivamento da tabela outbox (débito consciente registrado).
- Métricas custom além do contador `outbox_pendentes` (E0d).
- Dashboard/SDK OTel completo (E0d).

## Arquivos esperados (~15–17, ≤ 30)

- Migrations: `000002_outbox_events.{up,down}.sql`, `000003_processed_messages.{up,down}.sql` — 4
- Plataforma outbox: `internal/outbox/` (enqueue, relay, queries sqlc + gerado) — ~4
- Plataforma broker/worker: `internal/broker/` (cliente AMQP, topologia, consumer runtime, dedup wrapper) — ~3
- Domínio: handler do evento no consumidor + emissão via service do catálogo (`WithTx`) — ~2
- `cmd/morfeu/main.go` (modos api/worker/all + subcomando CLI de escrita) — 1
- `docker-compose.yml` + `.env.example` — 2
- Testes de integração (testcontainers RabbitMQ real) — ~3
- Controle: `lib.md`, `state.md`, `plan.md`, `docs/tasks/`, `docs/prd/` — (não contam código)

## Dependências esperadas

- **`github.com/rabbitmq/amqp091-go`** (nova — registrar em `lib.md` com versão pinada + checagem de CVE antes do primeiro import).
- **`go.opentelemetry.io/otel` (propagation)** (nova, pacote leve — registrar em `lib.md`; SDK completo só no E0d).
- Testcontainers: módulo RabbitMQ (já coberto pelo testcontainers-go existente ou submódulo a registrar).

## Critérios de aceite

- [ ] CA01 — Outbox transacional: rollback da TX do efeito → evento **não** persiste na outbox (teste de integração).
- [ ] CA02 — Relay só marca `published_at` **após** publisher confirm; broker fora do ar → evento permanece pendente e é reentregue depois (retry sem crash).
- [ ] CA03 — Consumidor idempotente: mesma mensagem entregue 2× → efeito aplicado 1× + ack na 2ª (dedup por `processed_messages` na mesma TX).
- [ ] CA04 — Mensagem envenenada (payload malformado) vai à DLQ após `x-delivery-limit=3` sem travar a fila principal nem derrubar o processo.
- [ ] CA05 — Graceful shutdown (SIGTERM) sem perda nem duplicação; `goleak.VerifyTestMain` no pacote do worker sem vazamento.
- [ ] CA06 — Migrations 002/003 up→down→up idempotentes + `sqlc vet` limpo (gates do CI).
- [ ] CA07 — Regressão: contrato da E0a intacto (`GET /filmes`, health/readiness com RabbitMQ adicionado sem quebrar formato); asserts existentes inalterados.
- [ ] CA08 — Fronteiras: depguard prova que nenhum domínio importa `amqp091-go`; CLI de escrita usa o service real na mesma TX.
- [ ] CA09 — Suíte completa verde com `-race` no CI (testcontainers RabbitMQ: container por pacote, filas nomeadas por teste, wait por log `Server startup complete`, timeout 90–120s, sem sleep fixo).

## Riscos

- Flakiness testcontainers+RabbitMQ no runner ARM64 → wait por log, timeouts 90–120s, container por pacote, filas por teste (exigência do refinamento).
- Complexidade da reconexão manual (NotifyClose + backoff) → comportamento durante desconexão especificado no PRD; teste de broker fora do ar cobre o caminho.
- Declaração de topologia divergente entre processos → 406 PRECONDITION_FAILED → função única de declaração idempotente compartilhada por api/worker/CLI.
- Dedup fora da TX do efeito (bug clássico) → wrapper de plataforma força `processed_messages` na mesma TX; teste CA03 é o gate.

## Estimativa de impacto

Alto em arquitetura (espinha dorsal de mensageria que a saga E6 reutiliza; 2 tabelas novas de plataforma), médio em infra local (RabbitMQ entra no compose), baixo em usuários (sem endpoint novo). Primeira dependência Go nova desde a E0a (`amqp091-go`).
