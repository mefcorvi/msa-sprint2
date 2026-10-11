// Package consumer feeds Kafka records to the history handler. Offsets are committed only after a batch
// is processed, so delivery is at least once and the handler must be idempotent.
package consumer

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"

	"hotelio/booking-history-service/internal/history"
)

const retryInterval = 2 * time.Second

func Run(ctx context.Context, client *kgo.Client, handler *history.Handler, log zerolog.Logger) {
	for {
		fetches := client.PollFetches(ctx)
		if ctx.Err() != nil || fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			log.Warn().Str("topic", topic).Int32("partition", partition).Err(err).Msg("fetch failed")
		})

		processed := true
		fetches.EachRecord(func(r *kgo.Record) {
			if processed {
				processed = handleWithRetry(ctx, handler, r, log)
			}
		})
		if !processed {
			return
		}
		if err := client.CommitUncommittedOffsets(ctx); err != nil {
			log.Warn().Err(err).Msg("offset commit failed")
		}
	}
}

// handleWithRetry blocks the partition until the record is stored: skipping it would lose history.
// It returns false only when ctx is cancelled.
func handleWithRetry(ctx context.Context, handler *history.Handler, r *kgo.Record, log zerolog.Logger) bool {
	for {
		err := handler.Handle(ctx, r.Value)
		switch {
		case err == nil:
			return true
		case errors.Is(err, history.ErrInvalidEvent):
			log.Error().Int32("partition", r.Partition).Int64("offset", r.Offset).Err(err).Msg("skipping invalid event")
			return true
		}
		log.Warn().Int32("partition", r.Partition).Int64("offset", r.Offset).Err(err).Msg("event processing failed, will retry")
		select {
		case <-ctx.Done():
			return false
		case <-time.After(retryInterval):
		}
	}
}
