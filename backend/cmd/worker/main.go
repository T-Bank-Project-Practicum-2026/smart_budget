package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"smart_budget/internal/kafka"
	"smart_budget/internal/outbox"
	"smart_budget/internal/postgres"
	app "smart_budget/internal/runtime"
	"strings"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	dbctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	db, err := postgres.NewPool(dbctx, "")
	if err != nil {
		return err
	}
	defer db.Close()

	brokers, err := app.Required("KAFKA_BROKERS")
	if err != nil {
		return err
	}

	client, err := kafka.NewClient(strings.Split(brokers, ","))
	if err != nil {
		return err
	}
	defer client.Close()

	publisher := &outbox.Publisher{DB: db, Kafka: client}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", app.Health("worker"))
	mux.HandleFunc("GET /api/ready", app.Ready(func(ctx context.Context) error {
		if err := db.Ping(ctx); err != nil {
			return err
		}

		return client.Ping(ctx)
	}))

	return app.Serve(ctx, mux, publisher.Run, publisher.Maintain)
}
