package kafka_test

import (
	"context"
	"os"
	"smart_budget/internal/kafka"
	"strings"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("KAFKA_BROKERS is not set; run make test for the Kafka integration check")
	}

	client, err := kafka.NewClient(strings.Split(brokers, ","))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)

	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("ping Kafka: %v", err)
	}
}
