package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL     string        `env:"BOOKING_DB_URL" envDefault:"postgres://booking:booking@localhost:5433/booking"`
	KafkaBrokers    []string      `env:"KAFKA_BOOTSTRAP_SERVERS" envDefault:"localhost:9092" envSeparator:","`
	EventsTopic     string        `env:"BOOKING_EVENTS_TOPIC" envDefault:"booking.events"`
	MonolithURL     string        `env:"MONOLITH_URL" envDefault:"http://localhost:8084"`
	MonolithTimeout time.Duration `env:"MONOLITH_TIMEOUT" envDefault:"2s"`
	OutboxInterval  time.Duration `env:"OUTBOX_POLL_INTERVAL" envDefault:"500ms"`
	GRPCPort        int           `env:"GRPC_PORT" envDefault:"9090"`
	HTTPPort        int           `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	EnableFeatureX  bool          `env:"ENABLE_FEATURE_X" envDefault:"false"`
	AppVersion      string        `env:"APP_VERSION" envDefault:"v1"`
}

func Load() (Config, error) {
	return env.ParseAs[Config]()
}
