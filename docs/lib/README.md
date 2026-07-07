# Docs de bibliotecas do Morfeu

Esta pasta consolida a documentacao tecnica das dependencias registradas em
[`lib.md`](../../lib.md). O objetivo e reduzir decisoes por memoria ou
suposicao durante implementacao.

## Fonte de verdade

- Registro de dependencias: [`lib.md`](../../lib.md)
- Arquitetura e restricoes: [`doc.md`](../../doc.md)
- Decisoes ja aprovadas: [`docs/adr/`](../adr/)

## Consultas Context7 realizadas

O MCP Context7 foi usado nesta atualizacao para buscar documentacao relevante
das bibliotecas centrais do backend:

| Biblioteca | Context7 ID | Conteudo usado |
|---|---|---|
| Echo | `/labstack/echo/v4.15.0` | roteamento, grupos, middleware, binder, validator, erros HTTP, shutdown |
| pgx | `/jackc/pgx` | pool, transacoes, batch, CopyFrom, contexto |
| sqlc | `/websites/sqlc_dev_en` | `sqlc.yaml`, `pgx/v5`, anotacoes SQL, codigo gerado |

As demais dependencias foram documentadas a partir do registro local do projeto,
padroes esperados de uso e restricoes arquiteturais ja aprovadas. Antes de
introduzir qualquer dependencia no build, revalidar a versao exata no Context7
e atualizar o `lib.md`.

## Arquivos

| Arquivo | Cobre |
|---|---|
| [`GO-1X.md`](GO-1X.md) | Go 1.x estavel, layout, contexto, erros, concorrencia, build |
| [`ECHO-V4.15.md`](ECHO-V4.15.md) | `labstack/echo/v4`, APIs HTTP, middleware, binding, erros |
| [`PGX-V5.9.md`](PGX-V5.9.md) | `jackc/pgx/v5`, pool, queries, transacoes, batch, COPY |
| [`SQLC.md`](SQLC.md) | sqlc com PostgreSQL e `pgx/v5` |
| [`AUTH-CRYPTO.md`](AUTH-CRYPTO.md) | JWT, `echo-jwt`, Argon2id, `x/crypto` |
| [`CACHE-MESSAGING.md`](CACHE-MESSAGING.md) | Redis/go-redis, RabbitMQ/amqp091-go, outbox |
| [`OBSERVABILITY.md`](OBSERVABILITY.md) | OpenTelemetry, Prometheus, Grafana, Loki, Tempo, Alloy |
| [`FRONTEND.md`](FRONTEND.md) | React, Vite, Playwright |
| [`DEVTOOLS-INFRA.md`](DEVTOOLS-INFRA.md) | migrate, testcontainers-go, k6, linters, Stripe CLI, Caddy |

