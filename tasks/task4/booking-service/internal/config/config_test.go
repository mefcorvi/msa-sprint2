package config

import (
	"slices"
	"testing"
	"time"

	"github.com/caarlos0/env/v11"
)

func TestDefaults(t *testing.T) {
	cfg, err := env.ParseAsWithOptions[Config](env.Options{Environment: map[string]string{}})

	if err != nil {
		t.Fatal(err)
	}
	if cfg.GRPCPort != 9090 || cfg.MonolithTimeout != 2*time.Second || cfg.OutboxInterval != 500*time.Millisecond {
		t.Fatalf("unexpected defaults %+v", cfg)
	}
}

func TestOverridesFromEnvironment(t *testing.T) {
	cfg, err := env.ParseAsWithOptions[Config](env.Options{Environment: map[string]string{
		"KAFKA_BOOTSTRAP_SERVERS": "kafka-1:9092,kafka-2:9092",
		"MONOLITH_TIMEOUT":        "750ms",
		"LOG_LEVEL":               "debug",
	}})

	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cfg.KafkaBrokers, []string{"kafka-1:9092", "kafka-2:9092"}) || cfg.MonolithTimeout != 750*time.Millisecond || cfg.LogLevel != "debug" {
		t.Fatalf("unexpected config %+v", cfg)
	}
}

func TestInvalidDurationIsRejected(t *testing.T) {
	_, err := env.ParseAsWithOptions[Config](env.Options{Environment: map[string]string{"MONOLITH_TIMEOUT": "soon"}})

	if err == nil {
		t.Fatal("expected an error for an invalid duration")
	}
}
