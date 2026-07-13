package outbox

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/mclovin137/morfeu/internal/broker"
	"github.com/mclovin137/morfeu/internal/outbox/db"
)

const (
	// relayPollInterval está dentro da janela 500ms-1s pedida pelo RF03.
	relayPollInterval = 750 * time.Millisecond
	relayBatchSize    = int32(50)
)

// Publisher é a porta que o relay usa para publicar — implementada por
// *broker.Client em produção; pequena o bastante para um fake nos testes que
// não precisam do broker real (a suíte de integração usa o broker real, ADR
// 0006, mas a interface permite testar o loop do relay isoladamente também).
type Publisher interface {
	PublicarComConfirm(ctx context.Context, msg broker.Message) error
}

// Relay é o loop de polling que publica eventos pendentes no broker com
// confirms síncronos, marcando published_at só após confirm positivo (RN02).
// Roda só em -mode=worker|all (RF04) — o wiring fica no composition root
// (cmd/morfeu/main.go), este tipo não sabe de flags.
type Relay struct {
	pool      Pool
	publisher Publisher
	logger    *zap.Logger
}

// NewRelay cria um Relay pronto para Run.
func NewRelay(pool Pool, publisher Publisher, logger *zap.Logger) *Relay {
	return &Relay{pool: pool, publisher: publisher, logger: logger}
}

// Run bloqueia processando lotes em intervalos de relayPollInterval até ctx
// ser cancelado (RF04: shutdown por ctx — para o polling e retorna assim que
// o lote em andamento termina; sem goroutine órfã porque é síncrona: quem
// chama Run em sua própria goroutine decide quando parar de esperar por ela).
func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(relayPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.processarLote(ctx); err != nil {
				r.logger.Warn("relay: erro processando lote de eventos pendentes", zap.Error(err))
			}
		}
	}
}

// processarLote lê até relayBatchSize eventos pendentes com
// SELECT ... FOR UPDATE SKIP LOCKED (RF03) e publica cada um dentro da mesma
// transação da leitura. Falha em publicar/confirmar qualquer evento faz a TX
// inteira sofrer rollback (RN02): eventos já marcados nesta mesma iteração
// voltam a ficar pendentes e serão republicados no próximo ciclo — duplicação
// aceita e tratada pelo consumidor (0005), nunca perda.
func (r *Relay) processarLote(ctx context.Context) error {
	return WithTx(ctx, r.pool, func(tx Tx) error {
		q := db.New(tx)

		eventos, err := q.SelectPendentesParaPublicar(ctx, relayBatchSize)
		if err != nil {
			return fmt.Errorf("relay: selecionar pendentes: %w", err)
		}

		for _, evt := range eventos {
			if err := r.publicarUm(ctx, q, evt); err != nil {
				return err
			}
		}
		return nil
	})
}

// publicarUm publica um único evento e só então marca published_at (RF03).
func (r *Relay) publicarUm(ctx context.Context, q *db.Queries, evt db.OutboxEvent) error {
	id, err := uuid.FromBytes(evt.ID.Bytes[:])
	if err != nil {
		return fmt.Errorf("relay: id do evento inválido: %w", err)
	}

	msg := broker.Message{
		RoutingKey:  evt.EventType,
		Body:        evt.Payload,
		MessageID:   id.String(),
		Type:        evt.EventType,
		AggregateID: evt.AggregateID,
		OccurredAt:  evt.OccurredAt.Time,
		Traceparent: novoTraceparent(),
	}

	// RNF01: log em info só com event_type/aggregate_id/message_id — nunca o
	// payload completo (isso ficaria, no máximo, em debug).
	logFields := []zap.Field{
		zap.String("event_type", evt.EventType),
		zap.String("aggregate_id", evt.AggregateID),
		zap.String("message_id", id.String()),
	}

	if err := r.publisher.PublicarComConfirm(ctx, msg); err != nil {
		r.logger.Info("relay: publish não confirmado, evento segue pendente",
			append(logFields, zap.Error(err))...,
		)
		return err
	}

	published := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	if err := q.MarcarPublicado(ctx, db.MarcarPublicadoParams{ID: evt.ID, PublishedAt: published}); err != nil {
		return fmt.Errorf("relay: marcar publicado: %w", err)
	}

	r.logger.Info("relay: evento publicado", logFields...)
	return nil
}

// novoTraceparent gera um traceparent W3C válido via
// go.opentelemetry.io/otel/propagation, sem SDK: como o relay é um processo
// desacoplado do request original (a outbox quebra a cadeia síncrona), não
// há span de origem para propagar — um novo trace/span aleatório é gerado
// a cada publish e injetado no formato W3C (RF05).
func novoTraceparent() string {
	var traceID trace.TraceID
	var spanID trace.SpanID
	_, _ = rand.Read(traceID[:])
	_, _ = rand.Read(spanID[:])

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceID,
		SpanID:     spanID,
		TraceFlags: trace.FlagsSampled,
	})
	ctx := trace.ContextWithSpanContext(context.Background(), spanContext)

	carrier := propagation.MapCarrier{}
	propagation.TraceContext{}.Inject(ctx, carrier)
	return carrier.Get("traceparent")
}
