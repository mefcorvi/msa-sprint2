package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"hotelio/booking-service/internal/booking"
	"hotelio/booking-service/internal/config"
	"hotelio/booking-service/internal/grpcapi"
	"hotelio/booking-service/internal/httpapi"
	"hotelio/booking-service/internal/monolith"
	"hotelio/booking-service/internal/outbox"
	"hotelio/booking-service/internal/pb"
	"hotelio/booking-service/internal/postgres"
)

func main() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
	log := zerolog.New(os.Stdout).With().Timestamp().Str("service", "booking-service").Logger()
	if err := run(log); err != nil {
		log.Error().Err(err).Msg("booking-service stopped")
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

	// pgxpool connects lazily, so the service starts even when the database is not up yet.
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	kafka, err := kgo.NewClient(kgo.SeedBrokers(cfg.KafkaBrokers...))
	if err != nil {
		return err
	}
	defer kafka.Close()

	var migrated atomic.Bool
	go func() {
		if migrateWithRetry(ctx, pool, log) {
			migrated.Store(true)
			outbox.NewRelay(pool, kafka, cfg.EventsTopic, cfg.OutboxInterval, log).Run(ctx)
		}
	}()
	ready := func(ctx context.Context) error {
		if !migrated.Load() {
			return errors.New("database migrations are not applied yet")
		}
		return pool.Ping(ctx)
	}

	service := booking.NewService(monolith.NewClient(cfg.MonolithURL, cfg.MonolithTimeout), postgres.NewStore(pool), time.Now, log)

	grpcServer := grpc.NewServer()
	pb.RegisterBookingServiceServer(grpcServer, grpcapi.NewServer(service, log))
	reflection.Register(grpcServer)

	grpcAddr := fmt.Sprintf(":%d", cfg.GRPCPort)
	httpServer := &http.Server{Addr: fmt.Sprintf(":%d", cfg.HTTPPort), Handler: httpapi.NewHandler(ready, cfg.EnableFeatureX, cfg.AppVersion), ReadHeaderTimeout: 5 * time.Second}

	errs := make(chan error, 2)
	go func() {
		lis, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			errs <- err
			return
		}
		log.Info().Str("addr", grpcAddr).Msg("gRPC server started")
		errs <- grpcServer.Serve(lis)
	}()
	go func() {
		log.Info().Str("addr", httpServer.Addr).Bool("featureX", cfg.EnableFeatureX).Str("version", cfg.AppVersion).Msg("HTTP server started")
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
	grpcServer.GracefulStop()
	return httpServer.Shutdown(shutdownCtx)
}

// migrateWithRetry waits for the database: in Kubernetes it may start after the service.
// It returns false only when ctx is cancelled.
func migrateWithRetry(ctx context.Context, pool *pgxpool.Pool, log zerolog.Logger) bool {
	for {
		err := postgres.Migrate(ctx, pool)
		if err == nil {
			log.Info().Msg("database migrated")
			return true
		}
		log.Warn().Err(err).Msg("database is not ready, will retry")
		select {
		case <-ctx.Done():
			return false
		case <-time.After(2 * time.Second):
		}
	}
}
