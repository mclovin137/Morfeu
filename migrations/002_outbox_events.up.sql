-- Task: 0002 — Outbox + RabbitMQ: lado produtor (E0b, parte 1/2)
-- PRD: docs/prd/0002-outbox-rabbitmq.md
-- Descrição: cria a tabela outbox_events (ownership excepcional da plataforma,
--   RN03 do PRD 0002 / ADR 0003) para o padrão transactional outbox: o evento
--   de domínio é gravado na mesma TX do efeito (RF01) e o relay (internal/outbox)
--   publica no RabbitMQ com publisher confirms síncronos, marcando published_at
--   só após confirm positivo (RN02). Índice parcial em published_at IS NULL
--   sustenta o polling do relay (SELECT ... FOR UPDATE SKIP LOCKED, RF03) sem
--   full scan à medida que a tabela cresce com eventos já publicados.
-- Rollback: 002_outbox_events.down.sql (DROP TABLE cascata do índice).

CREATE TABLE IF NOT EXISTS outbox_events (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type   TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    occurred_at  TIMESTAMPTZ NOT NULL,
    payload      JSONB NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

-- Sustenta o polling do relay: só pendentes (published_at IS NULL) são varridos;
-- parcial mantém o índice pequeno independente do volume histórico já publicado.
CREATE INDEX IF NOT EXISTS idx_outbox_events_pending
    ON outbox_events (created_at)
    WHERE published_at IS NULL;
