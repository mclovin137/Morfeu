# Task 0002 — Outbox + RabbitMQ: lado produtor (E0b, parte 1/2)

- **Data:** 2026-07-12
- **Status:** em andamento
- **Branch:** `feature/0002-outbox-rabbitmq`
- **PRD:** docs/prd/0002-outbox-rabbitmq.md (just-in-time, §6.2.5, consumindo o refinamento do E0)
- **Item do roadmap:** E0 (walking skeleton), entrega **E0b — parte 1/2 (produtor)**; exigências em `docs/refinamentos/E0-walking-skeleton.md` §"Task E0b"; **ADR 0007 (Mensageria: RabbitMQ)** aceito em 2026-07-11

> **Divisão da E0b (decisão de abertura do PRD, roles.md §6.3):** ao listar os arquivos reais para o PRD (incluindo gerados do sqlc, go.mod/go.sum, compose, testes e docs de controle — todos contam no diff auditado), a E0b inteira estimou **~32–33 arquivos > limite de 30**. A estimativa de 13–17 do refinamento não contava gerados/controle. Divisão aplicada: **0002 = produtor** (outbox + relay + confirms + topologia + emissão CLI), **0005 = consumidor** (runtime + dedup na mesma TX + DLQ + projeção + readiness RabbitMQ). As exigências do refinamento seguem válidas, distribuídas entre as duas; o walking skeleton assíncrono fecha na 0005.
>
> Numeração: o slot 0002 estava reservado para a E0b desde o refinamento; a branch `feature/0002-outbox-rabbitmq` foi recriada da main pós-0004 (rascunho antigo `84a4b50` superseded pelo reescopo).

## Objetivo

Construir o **lado produtor** da mensageria do walking skeleton: um evento de domínio real (`catalogo.filme_criado`) é gravado na **outbox na mesma TX** do efeito de domínio e publicado no RabbitMQ pelo **relay com publisher confirms síncronos** — com topologia completa declarada (exchange, quorum queue, DLQ), reconexão manual e sobrevivência a broker fora do ar. Base que a saga do checkout (E6) reutiliza.

## Escopo

- **Migration 002:** tabela `outbox_events` (ownership excepcional da plataforma, documentado no PRD).
- **Outbox (plataforma, `internal/outbox/`):** `Enqueue(tx, evt)`; relay com polling 500ms–1s, `SELECT ... FOR UPDATE SKIP LOCKED`, publisher confirms **síncronos** (`PublishWithDeferredConfirm` + `WaitContext`, timeout ~5s) — `published_at` só após confirm; broker fora do ar → evento permanece pendente e é reentregue (retry sem crash). Relay roda **só em `-mode=worker`/`all`**.
- **Cliente broker (plataforma, `internal/broker/`):** conexão amqp091-go, reconexão manual (`NotifyClose` + backoff exponencial com teto), publish com confirm, **declaração idempotente da topologia completa** no startup (exchange topic `morfeu.events`, quorum queue com `x-delivery-limit=3`, DLQ — parâmetros idênticos em todo processo, evita 406).
- **Emissão real:** subcomando CLI do binário (stdlib `flag`, sem cobra, sem endpoint HTTP) que cria filme + evento **na mesma TX** via service real (`WithTx`); service do catálogo ganha `CreateFilm`.
- **Envelope:** `message_id` (uuid), `event_type`, `aggregate_id`, `occurred_at`, `payload`, `traceparent` em header AMQP (propagator W3C via pacote leve `go.opentelemetry.io/otel/propagation`).
- **Compose:** RabbitMQ `3.13-management-alpine`, `vm_memory_high_watermark` absoluto (~768MB) + `disk_free_limit` (~512MB), 15672 só em `127.0.0.1`, healthcheck `rabbitmq-diagnostics -q ping`; credenciais via `.env` (nunca `guest/guest`).
- **Fronteiras (ADR 0003 + depguard):** nenhum domínio importa `amqp091-go`.
- **Segurança:** logs do relay sem payload completo em `info` (só `event_type`, `aggregate_id`, `message_id`); outbox nunca carrega segredo/PII; `amqp091-go` registrado no `lib.md` com versão pinada + CVE checada antes do primeiro import.
- **Observabilidade (fonte do dado):** query/contador de `outbox_pendentes` exposto internamente para o E0d consumir.

## Fora de escopo

- **Lado consumidor → task 0005 (E0b parte 2/2):** consumer runtime, dedup por `processed_messages` na mesma TX do efeito, DLQ exercitada por mensagem envenenada, projeção de domínio, graceful shutdown do consumer, readiness com RabbitMQ, migration 003.
- Segundo evento (`FilmUpdated`) e segunda queue — cortados pelo refinamento.
- Tooling de replay da DLQ (E6); LISTEN/NOTIFY (otimização futura registrada); limpeza da tabela outbox (débito consciente); métricas custom além do contador (E0d).

## Arquivos esperados (~27 no diff do PR, ≤ 30)

- Migrations: `002_outbox_events.{up,down}.sql` — 2
- `internal/outbox/`: `outbox.go`, `relay.go`, `queries.sql`, `db/` gerado (`db.go`, `models.go`, `queries.sql.go`) — 6
- `internal/broker/client.go` — 1
- `internal/catalogo/`: `service.go`, `queries.sql`, `db/` regen (`queries.sql.go`, `querier.go`) — 4
- `cmd/morfeu/main.go` (modos api|worker|all + subcomando criar-filme) — 1
- `docker-compose.yml`, `.env.example` — 2
- `sqlc.yaml`, `.golangci.yml` — 2
- `go.mod`, `go.sum` — 2
- `internal/outbox/relay_integration_test.go` (testcontainers PG+RabbitMQ, goleak) — 1
- Controle: `docs/prd/0002-*.md`, `docs/tasks/0002-*.md`, `docs/tasks/README.md`, `plan.md`, `state.md`, `lib.md` — 6

## Dependências esperadas

- **`github.com/rabbitmq/amqp091-go`** (nova — registrar em `lib.md` com versão pinada + checagem de CVE antes do primeiro import).
- **`go.opentelemetry.io/otel`** (nova, só o pacote leve `propagation`; SDK completo no E0d — registrar em `lib.md`).
- **`go.uber.org/goleak`** (nova, test-only — registrar em `lib.md`).
- Testcontainers: módulo RabbitMQ do testcontainers-go já presente (verificar se exige submódulo).

## Critérios de aceite

- [ ] CA01 — Outbox transacional: rollback da TX do efeito → evento **não** persiste na outbox (teste de integração).
- [ ] CA02 — Relay só marca `published_at` **após** publisher confirm; mensagem chega ao broker (consumo de verificação no próprio teste, via canal AMQP cru do pacote de teste).
- [ ] CA03 — Broker fora do ar → evento permanece pendente, relay não crasha e reentrega quando o broker volta.
- [ ] CA04 — Topologia declarada idempotentemente: dois startups consecutivos com parâmetros idênticos, sem 406; quorum queue + `x-delivery-limit=3` + DLQ existem no broker.
- [ ] CA05 — Subcomando CLI cria filme + evento na mesma TX via service real; relay ativo só em `-mode=worker|all` (nunca `api`).
- [ ] CA06 — Migration 002 up→down→up idempotente + `sqlc vet` limpo (gates do CI).
- [ ] CA07 — Regressão: contrato da E0a intacto (`GET /filmes`, health); asserts existentes inalterados.
- [ ] CA08 — Fronteiras: depguard prova que nenhum domínio importa `amqp091-go`.
- [ ] CA09 — Suíte completa verde com `-race` no CI; `goleak` no pacote do relay; testcontainers RabbitMQ com wait por log (`Server startup complete`), timeout 90–120s, sem sleep fixo.

## Riscos

- Flakiness testcontainers+RabbitMQ no runner ARM64 → wait por log, timeouts 90–120s, container por pacote (exigência do refinamento).
- Complexidade da reconexão manual (NotifyClose + backoff) → comportamento durante desconexão especificado no PRD; CA03 cobre o caminho.
- Declaração de topologia divergente entre processos → função única de declaração idempotente compartilhada (CA04).
- Divisão da E0b esconder integração produtor↔consumidor → mitigado: CA02 já verifica a mensagem no broker via consumo cru no teste; a 0005 fecha o ciclo com o consumer real.

## Estimativa de impacto

Alto em arquitetura (espinha dorsal de mensageria; tabela nova de plataforma), médio em infra local (RabbitMQ no compose), baixo em usuários (sem endpoint novo). Primeira dependência Go nova desde a E0a (`amqp091-go`).
