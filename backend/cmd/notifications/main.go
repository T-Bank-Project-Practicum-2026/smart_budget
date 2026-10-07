package main

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"log"
	"os"
	"os/signal"
	"smart_budget/internal/kafka"
	"smart_budget/internal/notifications"
	"smart_budget/internal/postgres"
	"smart_budget/internal/push"
	app "smart_budget/internal/runtime"
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

	settings := map[string]string{}

	for _, key := range []string{"KAFKA_BROKERS", "NOTIFICATIONS_TOKEN", "VAPID_PUBLIC_KEY", "VAPID_PRIVATE_KEY", "VAPID_SUBJECT", "AUTH_API_URL"} {
		value, err := app.Required(key)
		if err != nil {
			return err
		}

		settings[key] = value
	}

	privateBytes, err := base64.RawURLEncoding.DecodeString(settings["VAPID_PRIVATE_KEY"])
	if err != nil {
		return errors.New("invalid VAPID private key")
	}

	private, err := ecdh.P256().NewPrivateKey(privateBytes)
	if err != nil {
		return errors.New("invalid VAPID private key")
	}

	if base64.RawURLEncoding.EncodeToString(private.PublicKey().Bytes()) != settings["VAPID_PUBLIC_KEY"] {
		return errors.New("VAPID keys do not match")
	}

	dbctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	db, err := postgres.NewPool(dbctx, "")
	if err != nil {
		return err
	}
	defer db.Close()

	client, err := kafka.NewClient(strings.Split(settings["KAFKA_BROKERS"], ","), kgo.ClientID("notifications"), kgo.ConsumerGroup("notifications-v1"), kgo.ConsumeTopics(push.Topic), kgo.DisableAutoCommit(), kgo.BlockRebalanceOnPoll())
	if err != nil {
		return err
	}
	defer client.Close()

	svc := &notifications.Service{DB: db, Token: settings["NOTIFICATIONS_TOKEN"], PublicKey: settings["VAPID_PUBLIC_KEY"], PrivateKey: settings["VAPID_PRIVATE_KEY"], Subject: settings["VAPID_SUBJECT"], HTTPClient: notifications.PushHTTPClient(), SessionURL: settings["AUTH_API_URL"]}

	return app.Serve(ctx, svc.Handler(), svc.Consume(client), svc.Deliver, svc.Maintain)
}
