# Ferramentas de desenvolvimento e infraestrutura

Dependencias e servicos registrados:

- `golang-migrate/migrate v4`
- `testcontainers-go`
- `k6`
- `golangci-lint`
- `govulncheck`
- `@upstash/context7-mcp`
- Stripe CLI
- PostgreSQL, Redis, RabbitMQ
- Caddy
- Oracle Cloud Always Free
- Stripe, TMDB, Resend ou Brevo, UptimeRobot/healthchecks.io, Discord

## golang-migrate/migrate

Migrations devem ser SQL puro, versionadas e reversiveis quando possivel:

```text
internal/platform/db/migrations/
  000001_create_users.up.sql
  000001_create_users.down.sql
```

Regras:

- toda alteracao de schema passa por migration;
- indices acompanham queries previstas;
- evitar full scan em caminhos criticos;
- mudancas destrutivas exigem plano de compatibilidade;
- migration deve ser idempotente no processo de deploy, nao no SQL individual.

Comandos esperados:

```bash
migrate -path internal/platform/db/migrations -database "$DATABASE_URL" up
migrate -path internal/platform/db/migrations -database "$DATABASE_URL" down 1
```

## testcontainers-go

Usar para testes de integracao com servicos reais:

- PostgreSQL para repositorios/sqlc/migrations;
- Redis para cache/rate limit;
- RabbitMQ para outbox relay/workers.

Padrao:

- container por suite quando o servico for caro;
- dados isolados por UUID;
- limpar estado entre testes;
- `go test -race` ao menos na suite principal;
- nao depender de ordem entre testes.

## k6

k6 e ferramenta de carga, nao gate de PR. Usar para validar SLO antes de marco
ou mudanca de performance relevante.

Thresholds devem refletir o PRD:

```js
export const options = {
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
}
```

Gerador de carga roda fora da VM para nao medir a propria competicao por CPU.

## golangci-lint e govulncheck

CI minimo:

```bash
go test ./...
go test -race ./...
golangci-lint run ./...
govulncheck ./...
```

`govulncheck` e obrigatorio pelo parecer de security registrado no `lib.md`.

## @upstash/context7-mcp

Context7 e a fonte operacional para documentacao atualizada de bibliotecas
durante sessoes de desenvolvimento.

Configuracao registrada no projeto:

- `.mcp.json` referencia `${CONTEXT7_API_KEY:-}`;
- a chave real fica fora do git;
- `.claude/settings.local.json` injeta a variavel localmente quando necessario;
- sem chave, o servidor pode operar em modo anonimo com limite menor.

Uso esperado:

1. Antes de adicionar dependencia, resolver o ID da biblioteca no Context7.
2. Consultar documentacao da versao planejada ou mais proxima.
3. Registrar decisao, versao, riscos e alternativas no `lib.md`.
4. Criar ou atualizar o arquivo correspondente em `docs/lib/`.

Nao usar memoria como fonte final quando houver duvida de API, versao, CVE,
configuracao ou comportamento de framework.

## Stripe CLI

Usar `stripe listen` no desenvolvimento local para webhooks sem tunel externo.
Regras:

- validar assinatura do webhook no backend;
- nao confiar apenas no frontend para confirmar pagamento;
- idempotencia por event ID;
- registrar eventos recebidos/processados.

## Caddy

Caddy sera reverse proxy e TLS automatico:

- servir SPA estatica;
- encaminhar `/api/*` para backend;
- configurar headers de seguranca;
- manter API e SPA em same-origin quando possivel.

## Servicos externos

Stripe:

- usar test mode no desenvolvimento;
- criar modo de load-test fake para nao bater limites de sandbox.

TMDB:

- respeitar limite free e atribuicao obrigatoria;
- cachear respostas quando permitido.

Resend/Brevo:

- dominio verificado;
- templates transacionais versionados;
- fila/retry para envio.

UptimeRobot/healthchecks.io:

- watchdog externo para detectar VM morta;
- alertas independentes da propria infraestrutura Grafana.

Discord:

- canal de alertas operacional;
- mensagens objetivas com ambiente, servico, severidade e link do painel.
