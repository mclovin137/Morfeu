-- Task: 0002 — Outbox + RabbitMQ: lado produtor (E0b, parte 1/2)
-- PRD: docs/prd/0002-outbox-rabbitmq.md
-- Rollback de 002_outbox_events.up.sql: remove a tabela outbox_events e,
-- por cascata, o índice parcial idx_outbox_events_pending.

DROP TABLE IF EXISTS outbox_events;
