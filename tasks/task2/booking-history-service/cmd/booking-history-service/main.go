package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"

	"hotelio/booking-history-service/internal/config"
	"hotelio/booking-history-service/internal/consumer"
	"hotelio/booking-history-service/internal/history"
	"hotelio/booking-history-service/internal/postgres"
)

const consumerGroup = "booking-history-service"

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log := zerolog.New(os.Stdout).With().Timestamp().Str("service", "booking-history-service").Logger()
	if err := run(log); err != nil {
		log.Error().Err(err).Msg("booking-history-service stopped")
		os.Exit(1)
	}
}

func run(log zerolog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("LOG_LEVEL: %w", err)
	}
	log = log.Level(level)

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	kafka, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.ConsumerGroup(consumerGroup),
		kgo.ConsumeTopics(cfg.EventsTopic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		kgo.DisableAutoCommit(),
		// booking-service owns and creates the topic; pick it up (and new partitions) soon after it appears.
		kgo.MetadataMaxAge(30*time.Second),
	)
	if err != nil {
		return err
	}
	defer kafka.Close()

	handler := history.NewHandler(postgres.NewStore(pool), log)
	go func() {
		log.Info().Str("topic", cfg.EventsTopic).Str("group", consumerGroup).Msg("consuming")
		consumer.Run(ctx, kafka, handler, log)
	}()

	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", cfg.HTTPPort), Handler: healthHandler(pool), ReadHeaderTimeout: 5 * time.Second}
	errs := make(chan error, 1)
	go func() {
		if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("shutting down")
	case err := <-errs:
		return err
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}

func healthHandler(pool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})
	return mux
}
