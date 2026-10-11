package config

import (
	"slices"
	"testing"

	"github.com/caarlos0/env/v11"
)

func TestDefaults(t *testing.T) {
	cfg, err := env.ParseAsWithOptions[Config](env.Options{Environment: map[string]string{}})

	if err != nil {
		t.Fatal(err)
	}
	if cfg.EventsTopic != "booking.events" || cfg.HTTPPort != 8080 || cfg.LogLevel != "info" {
		t.Fatalf("unexpected defaults %+v", cfg)
	}
}

func TestOverridesFromEnvironment(t *testing.T) {
	cfg, err := env.ParseAsWithOptions[Config](env.Options{Environment: map[string]string{
		"KAFKA_BOOTSTRAP_SERVERS": "kafka-1:9092,kafka-2:9092",
		"HTTP_PORT":               "8081",
	}})

	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cfg.KafkaBrokers, []string{"kafka-1:9092", "kafka-2:9092"}) || cfg.HTTPPort != 8081 {
		t.Fatalf("unexpected config %+v", cfg)
	}
}
