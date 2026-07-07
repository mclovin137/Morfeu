# Cache e mensageria

Bibliotecas registradas:

- `redis/go-redis/v9 >= 9.7.3`
- `rabbitmq/amqp091-go`

Servicos planejados:

- Redis oficial
- RabbitMQ oficial multi-arch

## Redis com go-redis

Redis e cache, nao fonte da verdade. O `doc.md` explicita que estado
transacional nao deve morar no Redis.

Usos adequados:

- cache de cartaz;
- cache de mapa de assentos derivado;
- rate limit;
- locks curtos apenas se houver desenho e timeout bem definidos.

Usos proibidos ou de alto risco:

- reserva definitiva de assento;
- status final de pagamento;
- controle unico de outbox;
- fila transacional principal.

Exemplo:

```go
rdb := redis.NewClient(&redis.Options{
	Addr:         cfg.RedisAddr,
	Password:     cfg.RedisPassword,
	DB:           0,
	ReadTimeout:  500 * time.Millisecond,
	WriteTimeout: 500 * time.Millisecond,
})

func (c *MovieCache) GetPoster(ctx context.Context, movieID string) ([]byte, error) {
	key := "movie:poster:" + movieID
	b, err := c.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, fmt.Errorf("redis get poster: %w", err)
	}
	return b, nil
}
```

Definir TTL em todo cache derivado. Chaves devem ter prefixo por dominio.

## RabbitMQ com amqp091-go

RabbitMQ sera usado pelo outbox relay e workers. O cliente `amqp091-go` e o
sucessor mantido do antigo `streadway/amqp`.

Padroes:

- declarar exchanges/queues no bootstrap do worker;
- usar mensagens persistentes para eventos relevantes;
- ack somente apos processamento bem-sucedido;
- nack/requeue com limite ou DLQ;
- configurar prefetch;
- propagar trace context em headers.

Exemplo conceitual:

```go
if err := ch.Qos(prefetch, 0, false); err != nil {
	return err
}

msgs, err := ch.Consume(queue, consumerTag, false, false, false, false, nil)
if err != nil {
	return err
}

for msg := range msgs {
	if err := handle(ctx, msg.Body, msg.Headers); err != nil {
		_ = msg.Nack(false, shouldRequeue(err))
		continue
	}
	_ = msg.Ack(false)
}
```

## Outbox

Para consistencia:

1. Caso de uso grava entidade de dominio no PostgreSQL.
2. Na mesma transacao, grava evento em `outbox_events`.
3. Relay le eventos pendentes.
4. Relay publica no RabbitMQ.
5. Relay marca evento como publicado.

Nao publicar no RabbitMQ dentro da mesma transacao antes do commit do banco.

## Operacao

RabbitMQ na VM deve respeitar limites definidos no `lib.md`:

- `vm_memory_high_watermark` absoluto por volta de 768 MB;
- `disk_free_limit` obrigatorio;
- alertas por fila crescendo, consumers ausentes e taxa de nack.

