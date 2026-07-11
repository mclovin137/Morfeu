# ADR 0007 — Mensageria: RabbitMQ

- **Status:** aceito
- **Data:** 2026-07-11
- **Task/PRD relacionados:** refinamento E0 (`docs/refinamentos/E0-walking-skeleton.md`, task E0b), `doc.md` §10/§11

## Contexto

O projeto exige comunicação assíncrona por evento (e-mail pós-pivô com retry+DLQ, projeções, saga do checkout no E6) via padrão outbox. Para a volumetria real (~zero, portfólio), uma fila em PostgreSQL bastaria — e foi a recomendação da arquitetura na descoberta. O usuário decidiu por **RabbitMQ**, e a divergência fica registrada com o racional: **o domínio de mensageria é o fim do projeto, não o meio** — aprender broker real (topologia AMQP, confirms, redelivery, DLQ, operação sob memória limitada) é objetivo declarado de aprendizado/portfólio (`doc.md` §10). Este ADR consolida a decisão e o contrato de entrega definidos no refinamento do E0.

## Escopo

Cobre: escolha do broker, topologia, contrato de entrega (produção e consumo), biblioteca cliente, limites operacionais na VM Oracle A1. Não cobre: replay de DLQ (E6), métricas/alertas de fila (E0d), desenho da saga do checkout (E6), schema/payload de cada evento (PRDs dos épicos).

## Decisão

**RabbitMQ 3.13 (imagem `management-alpine`) como broker único, com entrega at-least-once + efeito idempotente.** Topologia: exchange topic única `morfeu.events`, routing key por domínio, **quorum queues** com `x-delivery-limit=3` e DLQ declarada; declaração idempotente no startup (args imutáveis em código, evitando 406 PRECONDITION_FAILED). Produção: **outbox na mesma TX do efeito de domínio** (ADR 0002, `WithTx`) + relay com polling 500ms–1s e `SELECT FOR UPDATE SKIP LOCKED`, publicando com **publisher confirms síncronos** (`PublishWithDeferredConfirm` + `WaitContext`); relay ativo só em `-mode=worker|all` (ADR 0001). Consumo: **dedup via `processed_messages` na mesma TX do efeito**. Envelope: `message_id`, `event_type`, `aggregate_id`, `occurred_at`, `payload`, `traceparent` (W3C em header AMQP). Cliente: `rabbitmq/amqp091-go` com **reconexão manual** (`NotifyClose` + backoff) encapsulada em `plataforma` (ADR 0003). Credenciais via `.env`, nunca `guest/guest`; UI de management (15672) só localhost/túnel.

## Tecnologias ou padrões envolvidos

RabbitMQ 3.13 (`management-alpine`), `rabbitmq/amqp091-go`, padrões: transactional outbox, publisher confirms, idempotent consumer (dedup table), DLQ com delivery limit, topic exchange, trace context W3C.

## Benefícios

- Aprendizado real do domínio-alvo do projeto: topologia AMQP, confirms, redelivery, DLQ, operação de broker sob restrição de memória — exatamente o que fila em PG não ensinaria.
- DLQ e limite de redelivery **nativos** (quorum + `x-delivery-limit`) — sem código próprio de poison message.
- Outbox no PG mantém a fonte da verdade transacional: broker pode cair/perder estado que o relay reemite.
- Topologia única e simples (1 exchange, routing key por domínio) — sem fan-out especulativo.
- Base pronta para a saga do E6 (compensações orientadas a evento) sem redesenho.

## Trade-offs

- **+1 serviço stateful na A1** (Erlang VM, ~150–250MB) competindo com PG/Redis/observabilidade.
- **amqp091-go não tem auto-reconnect** — reconexão é código nosso.
- **Quorum queue em nó único**: Raft de 1 réplica não dá alta disponibilidade — paga-se o overhead sem o benefício de replicação.
- **At-least-once**: todo consumidor é obrigado a deduplicar; disciplina de código permanente.
- Latência de ponta a ponta ≥ polling do relay (500ms–1s) — aceitável: nenhum fluxo assíncrono exige <1s.
- Complexidade operacional (upgrade, políticas, usuários) para dev solo.

## Riscos

- **Pressão de memória/OOM na A1** — prob. média, impacto alto: alarme de memória do Rabbit bloqueia publishers e pode derrubar o nó junto com PG/Redis.
- **Perda de mensagem por publish sem confirm** — prob. baixa, impacto alto (e-mail/projeção some silenciosamente).
- **Mensagem envenenada em loop de redelivery** — prob. média, impacto médio (consome CPU/fila).
- **Drift de topologia (406 no startup)** — prob. média no início do aprendizado, impacto baixo (falha ruidosa no boot).
- **Conexão instável (WSL local / VM)** — prob. média, impacto médio (relay/consumidor param).

## Estratégias para minimizar os trade-offs

- Memória: `vm_memory_high_watermark` absoluto ~768MB + `disk_free_limit` ~512MB; imagem alpine sem plugins extras; alarme bloqueando publisher é comportamento desejado (backpressure) — o outbox segura a fila no PG.
- Perda de mensagem: confirms síncronos obrigatórios no relay; outbox só marca `published` após confirm — reemissão automática em falha.
- Poison message: `x-delivery-limit=3` → DLQ; replay manual fica explicitamente fora até o E6.
- Reconexão: wrapper único em `plataforma` (`NotifyClose` + backoff exponencial), testado com testcontainers (ADR 0006); handlers de domínio nunca veem AMQP cru (ADR 0003).
- Quorum em nó único: aceito conscientemente — paridade com padrão de produção e `x-delivery-limit` nativo valem o overhead marginal em volumetria zero.
- Dedup: verificação em revisão de PR — nenhum consumidor entra sem `processed_messages` na mesma TX.

## Condições de reversão

Revisitar esta decisão se: (a) OOM/instabilidade recorrente na A1 mesmo com watermarks — degradar para fila em PG (River/SKIP LOCKED; o outbox e as portas em `plataforma` tornam a troca localizada); (b) surgir necessidade real de replay histórico/streaming (E6+) — reavaliar log distribuído; (c) `amqp091-go` deixar de ser mantida pelo core team do RabbitMQ.

## Impacto esperado

`plataforma` ganha cliente AMQP (conexão, reconexão, declaração de topologia, publish com confirm); módulos de domínio ganham handlers de consumo e escrita no outbox dentro das TXs existentes. Docker Compose e a A1 ganham o serviço RabbitMQ com limites configurados. Testes de integração sobem RabbitMQ real via testcontainers (ADR 0006).

## Alternativas consideradas e descartadas

- **Fila em PostgreSQL (SKIP LOCKED puro ou River)** — tecnicamente **suficiente e mais simples** para a volumetria real: zero infra nova, transacional nativo (dispensaria até o outbox), operação trivial. Era a recomendação anti-overengineering da arquitetura. Descartada por decisão do usuário: não ensina broker, topologia, confirms nem DLQ — o objetivo do projeto. Permanece como rota de reversão.
- **Redis Streams** — Redis já estará na VM, mas consumer groups + `XAUTOCLAIM` têm semântica menos padronizada, sem DLQ nativa, persistência (AOF/RDB) mais frágil que quorum queue, e competiria por RAM com o cache. Aprendizado pouco transferível.
- **NATS JetStream** — footprint excelente para a A1 (seria a vencedora se o critério fosse só recursos); descartada porque o modelo streams/consumers se afasta do AMQP clássico, DLQ não é nativa (max-deliver + advisories), e RabbitMQ tem mais aderência ao objetivo de aprendizado e à saga com DLQ madura.
- **Kafka/Redpanda** — log distribuído é overkill absoluto para a volumetria; Kafka (JVM) inviável na A1; Redpanda mais leve, mas o modelo de log/offsets não casa com fila de trabalho + DLQ por mensagem; complexidade de retenção/particionamento sem nenhum ganho.
- **SQS ou serviço gerenciado** — descartada: budget R$0 com lock-in de nuvem, credenciais externas em dev, testes de integração infiéis (LocalStack aproxima, não iguala) e zero aprendizado de operação de broker.

## ADRs relacionados

- ADR 0001 (binário único) — relay e consumidores vivem em `-mode=worker|all`.
- ADR 0002 (sqlc+pgx) — outbox e `processed_messages` usam `WithTx` na mesma TX do efeito.
- ADR 0003 (fronteiras) — cliente AMQP/reconexão/topologia são `plataforma`; handlers de consumo são domínio.
- ADR 0005 (DDD tático) — eventos de domínio e a saga do E6 consumirão esta infraestrutura.
- ADR 0006 (testes) — RabbitMQ real via testcontainers na integração.
