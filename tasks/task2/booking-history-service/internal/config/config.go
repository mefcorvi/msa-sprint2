package config

import (
	"github.com/caarlos0/env/v11"
)

type Config struct {
	DatabaseURL  string   `env:"HISTORY_DB_URL" envDefault:"postgres://history:history@localhost:5434/history"`
	KafkaBrokers []string `env:"KAFKA_BOOTSTRAP_SERVERS" envDefault:"localhost:9092" envSeparator:","`
	EventsTopic  string   `env:"BOOKING_EVENTS_TOPIC" envDefault:"booking.events"`
	HTTPPort     int      `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel     string   `env:"LOG_LEVEL" envDefault:"info"`
}

func Load() (Config, error) {
	return env.ParseAs[Config]()
}
