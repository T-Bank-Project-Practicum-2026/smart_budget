package outbox

import (
	"context"
	"errors"
	"log/slog"
	"smart_budget/internal/db/outboxdb"
	"smart_budget/internal/push"
	app "smart_budget/internal/runtime"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Publisher struct {
	Topic string
	DB    *pgxpool.Pool
	Kafka *kgo.Client
}

func (p *Publisher) PublishOne(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	tx, err := p.DB.Begin(ctx)
	if err != nil {
		return false, err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	queries := outboxdb.New(tx)
	pending, err := queries.LockNextEvent(ctx)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	event, err := push.Decode(pending.Payload)
	if err != nil {
		if err := queries.QuarantineEvent(ctx, pending.ID); err != nil {
			return false, err
		}

		if err := tx.Commit(ctx); err != nil {
			return false, err
		}

		slog.Error("outbox event quarantined", "event_id", pending.ID, "reason", "invalid_payload")

		return true, nil
	}

	produceCtx, stopProduce := context.WithTimeout(ctx, 5*time.Second)
	produceErr := p.Kafka.ProduceSync(produceCtx, &kgo.Record{Topic: p.topic(), Key: []byte(event.UserID), Value: pending.Payload}).FirstErr()

	stopProduce()
	if produceErr != nil {
		delay := time.Duration(1<<min(pending.Attempts+1, 8)) * time.Second
		if err := queries.RetryEvent(ctx, outboxdb.RetryEventParams{ID: pending.ID, NextAt: time.Now().Add(delay)}); err != nil {
			return false, err
		}

		if err := tx.Commit(ctx); err != nil {
			return false, err
		}

		slog.Warn("outbox publish deferred", "event_id", pending.ID, "attempt", pending.Attempts+1, "reason", "broker_unavailable")

		return false, nil
	}

	if err := queries.MarkPublished(ctx, pending.ID); err != nil {
		return false, err
	}

	return true, tx.Commit(ctx)
}

func (p *Publisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			for i := 0; i < 100; i++ {
				sent, err := p.PublishOne(ctx)
				if err != nil {
					app.LogError("outbox.publish", err)
					break
				}

				if !sent {
					break
				}
			}
		}
	}
}

func (p *Publisher) topic() string {
	if p.Topic != "" {
		return p.Topic
	}

	return push.Topic
}
