# pgx v5.9+

Context7:

- Resolvido como `/jackc/pgx`
- Documentacao consultada para pool, queries, transacoes, batch, `CopyFrom` e
  cancelamento por contexto.

Registro local: [`lib.md`](../../lib.md) exige `jackc/pgx/v5 >= 5.9.2`.

## Papel no Morfeu

`pgx` e o driver PostgreSQL principal e tambem a base de integracao com `sqlc`.
O banco e a fonte da verdade do sistema; Redis e RabbitMQ nao devem guardar
estado transacional definitivo.

## Pool

Usar `pgxpool.Pool` como dependencia compartilhada do processo:

```go
func NewPool(ctx context.Context, dsn string, cfg PoolConfig) (*pgxpool.Pool, error) {
	pc, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pg dsn: %w", err)
	}

	pc.MaxConns = cfg.MaxConns
	pc.MinConns = cfg.MinConns
	pc.MaxConnLifetime = cfg.MaxConnLifetime
	pc.MaxConnIdleTime = cfg.MaxConnIdleTime
	pc.HealthCheckPeriod = cfg.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, pc)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return pool, nil
}
```

Definir `MaxConns` considerando:

- numero de processos;
- workers concorrentes;
- limites do PostgreSQL;
- custo das queries criticas;
- conexoes usadas por migrations, k6 e testes.

## Queries

Para queries geradas por `sqlc`, o handler ou service deve depender de uma
interface de repositorio, nao do pool diretamente. Dentro do repositorio, o
codigo gerado recebe o pool ou uma transacao.

Para chamadas manuais inevitaveis:

```go
row := pool.QueryRow(ctx, `
	SELECT id, status
	FROM reservations
	WHERE id = $1
`, id)

if err := row.Scan(&reservation.ID, &reservation.Status); err != nil {
	if errors.Is(err, pgx.ErrNoRows) {
		return Reservation{}, ErrReservationNotFound
	}
	return Reservation{}, fmt.Errorf("get reservation: %w", err)
}
```

Sempre passar `ctx` recebido da borda. Nao usar `context.Background()` em
repositorios.

## Transacoes

Context7 confirmou o padrao `BeginTx`, `defer Rollback` e `Commit` no fim. A
chamada a `Rollback` apos commit pode retornar erro e deve ser ignorada quando
for apenas limpeza.

```go
func (r *Repository) WithTx(ctx context.Context, fn func(q *db.Queries) error) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(r.queries.WithTx(tx)); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
```

Para reserva de assentos, avaliar isolamento e locks no PRD da feature. Caso
haja concorrencia forte no mesmo assento/sessao, usar SQL explicito com lock,
constraint unica parcial ou estrategia equivalente documentada em migration.

## Batch e CopyFrom

`Batch` reduz round trips quando varias operacoes independentes precisam ser
executadas juntas. Fechar sempre os resultados:

```go
batch := &pgx.Batch{}
batch.Queue("insert into audit_events(kind, payload) values($1, $2)", kind, payload)
batch.Queue("insert into outbox(topic, payload) values($1, $2)", topic, payload)

br := pool.SendBatch(ctx, batch)
defer br.Close()

if _, err := br.Exec(); err != nil {
	return err
}
if _, err := br.Exec(); err != nil {
	return err
}
```

`CopyFrom` e adequado para cargas internas ou importacao em lote, nao para fluxo
transacional de usuario final sem necessidade comprovada.

## Observabilidade

Instrumentar com `otelpgx` conforme registrado em `lib.md`. Regras:

- nao incluir SQL completo com dados sensiveis em atributos;
- medir latencia e erro por operacao;
- controlar cardinalidade de labels;
- incluir `request_id`/trace id nos logs correlacionados.

