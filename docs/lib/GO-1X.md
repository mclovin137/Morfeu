# Go 1.x estavel

Registro local: [`lib.md`](../../lib.md) define Go como linguagem planejada do
backend, com versao "1.x estavel". A escolha foi feita por aprendizado,
cross-compile simples para ARM e binario estatico com `CGO_ENABLED=0` quando
possivel.

## Papel no Morfeu

Go sera a base de todo o backend: API HTTP, workers, outbox relay, acesso a
PostgreSQL, cache Redis, publicacao/consumo RabbitMQ e instrumentacao
OpenTelemetry.

O projeto deve ser simples de operar em uma VM pequena. Isso favorece:

- um binario por processo;
- configuracao por variaveis de ambiente;
- shutdown gracioso por `context.Context`;
- logs estruturados em stdout/stderr;
- poucos frameworks e dependencias explicitas.

## Regras de projeto

Usar `context.Context` em todas as chamadas I/O:

```go
func (s *Service) CreateOrder(ctx context.Context, input CreateOrderInput) (Order, error) {
	return s.repo.Create(ctx, input)
}
```

Nao criar contextos soltos dentro de repositorios ou clients. O contexto deve
vir da borda da requisicao HTTP, worker ou comando.

Separar responsabilidades:

- `cmd/api`: bootstrap do servidor HTTP.
- `cmd/worker`: consumers e jobs.
- `internal/platform`: banco, cache, mensageria, observabilidade, config.
- `internal/<modulo>`: dominio, casos de uso, handlers e repositorios do modulo.

Evitar pacote `pkg/` ate existir necessidade real de API publica compartilhada.

## Erros

Erros internos devem preservar causa com `%w`:

```go
if err != nil {
	return fmt.Errorf("create reservation: %w", err)
}
```

Na borda HTTP, converter erros de dominio para contrato de API:

- validacao: `400`;
- autenticacao ausente/invalida: `401`;
- permissao insuficiente: `403`;
- recurso inexistente: `404`;
- conflito de regra de negocio: `409`;
- falha inesperada: `500` sem vazar detalhe interno.

## Concorrencia

Usar goroutines apenas quando houver beneficio claro. Toda goroutine longa deve
ter:

- contexto cancelavel;
- log de inicio e fim;
- estrategia de retry/backoff quando falar com rede;
- caminho de shutdown.

Para workers RabbitMQ, limitar concorrencia por configuracao. Nao deixar o
prefetch e o numero de goroutines crescerem sem relacao com CPU, memoria e pool
do banco.

## Build e runtime

Padrao esperado:

```bash
go test ./...
go test -race ./...
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/morfeu-api ./cmd/api
```

O build para ARM deve ser validado cedo porque a producao-demo planejada usa
Oracle Cloud A1.

## Seguranca

Nao logar secrets, tokens JWT, senhas, dados completos de pagamento ou payloads
de webhook com assinatura. Logs devem carregar IDs correlacionaveis, nao dados
sensiveis.

Usar `govulncheck` no CI como gate continuo de vulnerabilidades Go.

