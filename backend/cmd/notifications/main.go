package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"smart_budget/internal/kafka"
	"smart_budget/internal/notifications"
	"strings"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	brokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")

	for i, broker := range brokers {
		brokers[i] = strings.TrimSpace(broker)
		if brokers[i] == "" {
			return errors.New("KAFKA_BROKERS must contain non-empty broker addresses")
		}
	}

	client, err := kafka.NewClient(brokers, kgo.ClientID("notifications"))
	if err != nil {
		return fmt.Errorf("create Kafka client: %w", err)
	}
	defer client.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	err = client.Ping(pingCtx)

	cancel()
	if err != nil {
		return fmt.Errorf("connect to Kafka: %w", err)
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           notifications.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)

	go func() {
		log.Printf("notifications HTTP server listening on %s", addr)

		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		return srv.Shutdown(shutdownCtx)
	}
}
