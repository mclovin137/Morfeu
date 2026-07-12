# PRD 0002 — Outbox + RabbitMQ: lado produtor (E0b, parte 1/2)

- **Task:** docs/tasks/0002-outbox-rabbitmq.md
- **Branch:** `feature/0002-outbox-rabbitmq`
- **Data:** 2026-07-12
- **Status:** ativo

> Just-in-time (roles.md §6.2.5), sem nova rodada de agentes: consome `docs/refinamentos/E0-walking-skeleton.md` §"Task E0b" + **ADR 0007** (Mensageria: RabbitMQ). **Divisão da E0b registrada na abertura** (task doc, roles.md §6.3): a task inteira estimou ~32–33 arquivos reais (> 30); 0002 = produtor, 0005 = consumidor. **Skills de apoio para a implementação (§4.4): `golang-concurrency`, `golang-database`** (relay concorrente com shutdown + TX/SKIP LOCKED/pgx).

## Objetivo

Provar a metade produtora do caminho assíncrono do walking skeleton: `catalogo.filme_criado` nasce **na mesma TX** do efeito de domínio (outbox transacional), e o **relay** o publica no RabbitMQ com **publisher confirms síncronos**, topologia completa (exchange topic, quorum queue, DLQ) declarada idempotentemente, reconexão manual e tolerância a broker fora do ar. É a infraestrutura que a saga do checkout (E6) reutiliza sem redesenho.

## Escopo

Conforme task doc (fonte: refinamento §E0b + ADR 0007): migration `outbox_events`; `internal/outbox/` (Enqueue + relay); `internal/broker/` (cliente AMQP, reconexão, topologia, publish+confirm); `CreateFilm` no service do catálogo + subcomando CLI emitindo na mesma TX; envelope com `traceparent`; RabbitMQ no compose com limites de memória/disco; contador `outbox_pendentes`; depguard domínio ↛ `amqp091-go`.

## Fora de escopo

Lado consumidor completo → **task 0005** (consumer runtime, dedup `processed_messages` na mesma TX, DLQ exercitada, projeção, readiness com RabbitMQ, migration 003). `FilmUpdated`/segunda queue (cortados no refinamento); replay de DLQ (E6); LISTEN/NOTIFY (futura); limpeza da outbox (débito registrado); métricas custom além do contador (E0d); SDK OTel completo (E0d).

## Requisitos funcionais

- RF01 — `outbox.Enqueue(ctx, tx, evt)` grava o evento na tabela `outbox_events` **usando a TX recebida** (nunca abre TX própria); chamável por qualquer service de domínio dentro de `WithTx`.
- RF02 — Subcomando CLI `criar-filme` (stdlib `flag`; ex.: `morfeu criar-filme -titulo=... -sinopse=... -duracao=...`): valida entrada, chama `CreateFilm` do service real do catálogo, que insere o filme **e** enfileira `catalogo.filme_criado` na mesma TX; imprime o id criado; exit code ≠ 0 em erro.
- RF03 — Relay: loop com polling 500ms–1s; lê pendentes (`published_at IS NULL`) com `SELECT ... FOR UPDATE SKIP LOCKED` (lote pequeno, ex. 50); publica na exchange `morfeu.events` com routing key = `event_type`; aguarda confirm síncrono (`PublishWithDeferredConfirm` + `WaitContext`, timeout 5s); **só então** marca `published_at` (na mesma TX da leitura). Confirm negado/timeout → TX faz rollback, evento continua pendente.
- RF04 — Relay ativo **somente** em `-mode=worker|all` (flag existente do binário; `api` nunca publica). Shutdown por ctx (SIGTERM): para o polling, espera publicação em curso, fecha conexão — sem goroutine órfã.
- RF05 — Envelope AMQP: body = payload JSON do evento; propriedades/headers: `message_id` (uuid v4), `type` (= `event_type`), headers `aggregate_id`, `occurred_at` (RFC3339) e `traceparent` (formato W3C, gerado/propagado via `go.opentelemetry.io/otel/propagation` — sem SDK, contexto vazio gera novo traceparent); `content_type: application/json`; `delivery_mode: persistent`.
- RF06 — `internal/broker`: conexão única por processo com reconexão manual — `NotifyClose` dispara re-dial com backoff exponencial (base 1s, teto 30s, jitter); durante desconexão o relay segue rodando e falhando o publish (eventos ficam pendentes — comportamento especificado: **nenhum crash, nenhuma perda**); reconectado, o ciclo normal reassume.
- RF07 — Declaração idempotente da topologia no startup de todo processo que toca o broker (worker/all e o próprio teste): exchange topic durável `morfeu.events`; queue quorum `catalogo.filme_criado` com `x-delivery-limit=3` e DLX/DLQ (`morfeu.events.dlx` + queue `catalogo.filme_criado.dlq`); binding por routing key. Parâmetros imutáveis em código — dois startups consecutivos não geram 406.
- RF08 — `outbox.Pendentes(ctx) (int64, error)`: contagem de eventos não publicados, exposta como função de plataforma (fonte do dado para o E0d; sem endpoint/métrica nesta task).

## Requisitos não funcionais

- RNF01 — Segurança: credenciais do RabbitMQ via env (`RABBITMQ_URL` ou host/user/pass), nunca `guest/guest` em compose, nunca hardcode; logs do relay em `info` só com `event_type`, `aggregate_id`, `message_id` (payload completo no máximo em `debug`); payload de outbox nunca contém segredo/PII (princípio registrado).
- RNF02 — Compose: `rabbitmq:3.13-management-alpine`, `vm_memory_high_watermark.absolute` ~768MB, `disk_free_limit.absolute` ~512MB, porta 15672 mapeada só em `127.0.0.1`, healthcheck `rabbitmq-diagnostics -q ping`.
- RNF03 — Fronteiras (ADR 0003): `amqp091-go` importável só por `internal/broker` (regra depguard); domínio enxerga a porta `outbox.Enqueue`/interface de publicação, nunca AMQP cru.
- RNF04 — Consistência sqlc (ADR 0002): todo SQL novo via `queries.sql` + `sqlc generate` sem diff manual; `sqlc vet` limpo.
- RNF05 — Testes (ADR 0006): integração com testcontainers **reais** (PG + RabbitMQ), `-race`, container por pacote (TestMain), wait por log `Server startup complete` (RabbitMQ), timeout 90–120s, sem sleep fixo; `goleak.VerifyTestMain` no pacote do relay.
- RNF06 — Graceful degradation: broker indisponível não derruba worker nem api; log de erro com backoff, sem loop quente.

## Regras de negócio

- RN01 — Evento `catalogo.filme_criado` só existe se o filme foi criado (mesma TX — atomicidade absoluta; rollback de um é rollback do outro).
- RN02 — `published_at` é a única marca de publicação e só é gravada após confirm positivo do broker (at-least-once garantido do lado produtor; duplicação é aceita e tratada pelo consumidor na 0005).
- RN03 — Ownership excepcional: `outbox_events` pertence à **plataforma** (`internal/outbox`), não a um domínio — exceção documentada ao package-by-domain (ADR 0003), autorizada pelo refinamento §E0b.

## Critérios de aceite

- [ ] CA01 — Rollback da TX do efeito → evento não persiste na outbox (integração).
- [ ] CA02 — Relay marca `published_at` só após confirm; a mensagem publicada é observável na queue real (consumo AMQP cru dentro do teste) com envelope completo (RF05).
- [ ] CA03 — Broker parado (container stop): eventos ficam pendentes, relay não crasha; broker religado: eventos são publicados (retry/reconexão fim-a-fim).
- [ ] CA04 — Topologia idempotente: declaração 2× consecutivas sem 406; queue é quorum com `x-delivery-limit=3` e DLQ existente (inspeção via API/declaração passiva no teste).
- [ ] CA05 — CLI `criar-filme` cria filme + evento na mesma TX via service real; relay só roda em `-mode=worker|all`.
- [ ] CA06 — Migration 002 up→down→up idempotente; `sqlc vet` + `sqlc diff` limpos (gates do CI).
- [ ] CA07 — Regressão: `GET /filmes` e health intactos; asserts pré-existentes inalterados.
- [ ] CA08 — depguard: import de `amqp091-go` fora de `internal/broker` é bloqueado (violação proposital testada e revertida, como na 0003).
- [ ] CA09 — Suíte completa verde com `-race` no CI ARM64; goleak sem vazamento no pacote do relay.

## Plano de testes

| Cenário | Tipo | Cobre |
|---|---|---|
| `CreateFilm` via `WithTx` commit → filme + evento presentes; rollback forçado → nenhum dos dois persiste | integração (PG real) | CA01, RN01 |
| Relay publica pendente; confirm ok → `published_at` preenchido; mensagem consumida crua da queue com `message_id`/`type`/headers/`traceparent` corretos | integração (PG+RabbitMQ reais) | CA02, RF03, RF05 |
| Broker parado → publish falha, evento segue pendente, sem crash (assert por estado + log); broker religado → evento publicado | integração (stop/start do container) | CA03, RF06, RNF06 |
| Declaração de topologia 2× → sem erro; queue quorum + `x-delivery-limit` + DLQ verificáveis | integração | CA04, RF07 |
| `Pendentes()` reflete a contagem antes/depois da publicação | integração | RF08 |
| CLI com flags inválidas → exit ≠ 0, nada persiste; flags válidas → id impresso | integração leve (execução da função do subcomando, não do binário) | RF02, CA05 |
| Suíte E0a completa (filmes, cache, health, migrations) inalterada | regressão | CA07 |
| goleak no TestMain do pacote do relay | unit/infra | CA09 |

## Plano de implementação

1. Registrar `amqp091-go`, `otel` (propagation) e `goleak` no `lib.md` (versões pinadas + CVE via OSV/govulncheck) **antes** do primeiro import; `go get` das três.
2. Migration 002 (`criar-migration`): `outbox_events` (id uuid PK, event_type, aggregate_id, occurred_at, payload jsonb, created_at, published_at nullable; índice parcial em `published_at IS NULL`).
3. `sqlc.yaml`: novo bloco p/ `internal/outbox/queries.sql` → `internal/outbox/db` (sem interface); queries: insert, select-pendentes FOR UPDATE SKIP LOCKED com limit, mark-published, count-pendentes; `sqlc generate`.
4. `internal/outbox/outbox.go`: tipo `Evento`, `Enqueue(ctx, tx, evt)`, `Pendentes(ctx)`.
5. `internal/broker/client.go`: dial + `NotifyClose`/backoff, `DeclararTopologia()`, `PublicarComConfirm(ctx, msg)`.
6. `internal/outbox/relay.go`: loop de polling + lote + publish + mark-published na mesma TX; shutdown por ctx.
7. `internal/catalogo`: `CreateFilm` no service (`WithTx` + `Enqueue`); query `InsertFilm`; regen.
8. `cmd/morfeu/main.go`: subcomando `criar-filme`; wiring do relay em `-mode=worker|all`.
9. `docker-compose.yml` + `.env.example`: serviço RabbitMQ com limites/healthcheck; variáveis de conexão.
10. `.golangci.yml`: depguard — `amqp091-go` só em `internal/broker`; violação proposital p/ validar, revertida.
11. Testes de integração (`relay_integration_test.go`, TestMain com goleak + containers PG/RabbitMQ por pacote, filas nomeadas por teste); suíte completa `-race`.
12. Atualizar `plan.md`/`state.md`; gate: gitleaks pré-push → push → PR → CI verde → passe único de julgamento → merge.

## Arquivos que serão criados

- `migrations/002_outbox_events.up.sql` / `.down.sql` — tabela outbox
- `internal/outbox/outbox.go` — Evento + Enqueue + Pendentes
- `internal/outbox/relay.go` — relay (polling, confirms, shutdown)
- `internal/outbox/queries.sql` — SQL da outbox
- `internal/outbox/db/db.go`, `db/models.go`, `db/queries.sql.go` — gerados (sqlc)
- `internal/outbox/relay_integration_test.go` — integração + goleak
- `internal/broker/client.go` — cliente AMQP (conexão/reconexão/topologia/publish+confirm)
- `docs/prd/0002-outbox-rabbitmq.md` — este PRD

## Arquivos que serão modificados

- `internal/catalogo/service.go` — `CreateFilm` com `WithTx` + `Enqueue`
- `internal/catalogo/queries.sql` — `InsertFilm`
- `internal/catalogo/db/queries.sql.go`, `db/querier.go` — regen sqlc
- `cmd/morfeu/main.go` — subcomando `criar-filme` + wiring do relay por modo
- `docker-compose.yml` — serviço RabbitMQ
- `.env.example` — variáveis RabbitMQ
- `sqlc.yaml` — bloco outbox
- `.golangci.yml` — depguard amqp091-go
- `go.mod`, `go.sum` — deps novas
- `docs/tasks/0002-outbox-rabbitmq.md`, `docs/tasks/README.md`, `plan.md`, `state.md`, `lib.md` — controle

*(Total estimado no diff: ~27 ≤ 30.)*

## Dependências utilizadas

- **Novas** (registrar no `lib.md` no passo 1, antes do import — roles.md §6.9): `github.com/rabbitmq/amqp091-go` (cliente oficial mantido pelo core team do RabbitMQ; ADR 0007), `go.opentelemetry.io/otel` (só `propagation`, leve; SDK no E0d), `go.uber.org/goleak` (test-only).
- Existentes: pgx/v5 + sqlc (ADR 0002), Echo (intacto), testcontainers-go (módulo RabbitMQ), golang-migrate (CI).

## Impactos técnicos

Banco: +1 tabela de plataforma (`outbox_events`), ownership excepcional documentado (RN03). Módulos: novos `internal/outbox` e `internal/broker` (plataforma); catálogo ganha via de escrita. Contratos HTTP: inalterados. Infra local: +1 serviço stateful no compose (RabbitMQ com limites). CI: suíte passa a subir RabbitMQ via testcontainers (timeout 90–120s já previsto no pipeline).

## Riscos

- Flakiness testcontainers+RabbitMQ no ARM64 → wait por log, 90–120s, container por pacote, filas por teste.
- Reconexão manual com corrida entre re-dial e publish → canal/conexão atrás de mutex + teste CA03 com stop/start real.
- Topologia divergente (406) → função única `DeclararTopologia` compartilhada por worker/CLI/testes (CA04).
- Divisão produtor/consumidor mascarar incompatibilidade de envelope → CA02 valida o envelope consumindo cru; a 0005 usa o mesmo contrato (RF05 é a especificação).
- Estouro do limite de 30 arquivos → lista fechada acima; qualquer arquivo extra exige atualizar este PRD antes (regra de manutenção).

## Estratégia de rollback

Migration 002 tem `down` completo (drop da tabela + índice) validado pelo gate up→down→up do CI. Código: revert do merge do PR restaura a E0a+CI intactos (nenhum contrato HTTP alterado; RabbitMQ no compose é aditivo). Se o RabbitMQ se mostrar inviável na A1 (condição de reversão do ADR 0007), a troca fica localizada em `internal/broker` — o contrato `outbox.Enqueue`/relay permanece.
