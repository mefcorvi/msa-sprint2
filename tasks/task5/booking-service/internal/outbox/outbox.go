// Package outbox implements the transactional outbox: events are stored in the same transaction as the
// business data and published to Kafka at least once by Relay, so consumers must be idempotent.
package outbox

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	batchSize       = 100
	topicPartitions = 3
	topicReplicas   = 1
	retryInterval   = 2 * time.Second
)

// Add must be called inside the transaction that writes the business data.
func Add(ctx context.Context, tx pgx.Tx, eventID, eventType, key string, payload []byte) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO outbox_event (event_id, event_type, aggregate_id, payload)
		VALUES ($1, $2, $3, $4)`,
		eventID, eventType, key, string(payload))
	return err
}

type Relay struct {
	pool     *pgxpool.Pool
	kafka    *kgo.Client
	topic    string
	interval time.Duration
	log      zerolog.Logger
}

func NewRelay(pool *pgxpool.Pool, kafka *kgo.Client, topic string, interval time.Duration, log zerolog.Logger) *Relay {
	return &Relay{pool: pool, kafka: kafka, topic: topic, interval: interval, log: log}
}

// Run creates the topic (booking-service owns it) and then publishes pending events until ctx is done.
func (r *Relay) Run(ctx context.Context) {
	for ctx.Err() == nil {
		if err := r.ensureTopic(ctx); err == nil {
			break
		} else {
			r.log.Warn().Str("topic", r.topic).Err(err).Msg("cannot create topic, will retry")
			sleep(ctx, retryInterval)
		}
	}
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.publishPending(ctx); err != nil && ctx.Err() == nil {
				r.log.Warn().Err(err).Msg("outbox publishing failed, will retry")
			}
		}
	}
}

func (r *Relay) ensureTopic(ctx context.Context) error {
	_, err := kadm.NewClient(r.kafka).CreateTopic(ctx, topicPartitions, topicReplicas, nil, r.topic)
	if errors.Is(err, kerr.TopicAlreadyExists) {
		return nil
	}
	return err
}

type pendingEvent struct {
	id      int64
	key     string
	payload string
}

func (r *Relay) publishPending(ctx context.Context) error {
	return pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, aggregate_id, payload FROM outbox_event
			WHERE published_at IS NULL
			ORDER BY id
			LIMIT $1
			FOR UPDATE SKIP LOCKED`, batchSize)
		if err != nil {
			return err
		}
		events, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (pendingEvent, error) {
			var e pendingEvent
			return e, row.Scan(&e.id, &e.key, &e.payload)
		})
		if err != nil || len(events) == 0 {
			return err
		}

		records := make([]*kgo.Record, len(events))
		ids := make([]int64, len(events))
		for i, e := range events {
			records[i] = &kgo.Record{Topic: r.topic, Key: []byte(e.key), Value: []byte(e.payload)}
			ids[i] = e.id
		}
		if err := r.kafka.ProduceSync(ctx, records...).FirstErr(); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE outbox_event SET published_at = now() WHERE id = ANY($1)`, ids); err != nil {
			return err
		}
		r.log.Info().Int("count", len(events)).Str("topic", r.topic).Msg("published outbox events")
		return nil
	})
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.After(d):
	}
}
