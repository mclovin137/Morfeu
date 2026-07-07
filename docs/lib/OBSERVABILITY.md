# Observabilidade

Dependencias registradas:

- `go.opentelemetry.io/otel >= 1.41.0`
- `otelecho`
- `otelpgx`

Servicos planejados:

- Prometheus
- Grafana
- Loki
- Tempo
- Grafana Alloy

## Objetivo

Observabilidade deve existir desde o primeiro dia, com amostragem planejada de
10% e 100% de erros conforme `lib.md`. O foco e debugar fluxos criticos:

- criacao de sessao/login;
- listagem de filmes/sessoes;
- reserva de assento;
- pagamento/webhook;
- outbox relay e workers.

## Traces

Instrumentar bordas:

- HTTP Echo com `otelecho`;
- PostgreSQL com `otelpgx`;
- RabbitMQ propagando contexto via headers;
- chamadas externas Stripe/TMDB/Resend ou Brevo.

Regra: cada trace deve permitir seguir uma reserva do request ate banco, outbox,
mensageria e worker.

## Metricas

Metricas recomendadas:

- latencia HTTP por rota normalizada, metodo e status;
- total de requests por rota normalizada;
- erros por codigo de dominio;
- pool pgx: conexoes em uso, idle, tempo de acquire;
- cache hit/miss Redis por operacao;
- tamanho e idade da outbox;
- filas RabbitMQ: mensagens prontas, unacked, consumers;
- duracao de jobs/workers;
- webhooks Stripe recebidos, rejeitados e processados.

Evitar cardinalidade alta:

- nao usar user ID como label;
- nao usar reservation ID como label;
- nao usar URL completa se contiver parametros variaveis;
- nao usar payloads como atributos.

## Logs

Logs estruturados em stdout:

```json
{
  "level": "info",
  "msg": "reservation created",
  "request_id": "...",
  "trace_id": "...",
  "reservation_id": "...",
  "user_id": "..."
}
```

Nao logar:

- senha ou hash;
- token JWT;
- segredo Stripe;
- assinatura completa de webhook;
- dados completos de pagamento.

## Loki, Tempo e Alloy

Alloy substitui Promtail conforme `lib.md`. Papel esperado:

- coletar logs da aplicacao e containers;
- encaminhar traces para Tempo;
- encaminhar metricas para Prometheus quando aplicavel;
- manter configuracao versionada com limites de retencao.

## Alertas

Alertas minimos:

- API fora do ar;
- taxa de erro HTTP alta;
- p95 acima do SLO definido no PRD;
- outbox com eventos antigos;
- fila RabbitMQ sem consumer;
- Redis indisponivel;
- PostgreSQL indisponivel ou pool saturado;
- uso de disco alto;
- VM sem heartbeat externo.

Canal planejado: Discord via Grafana Alerting webhook.

