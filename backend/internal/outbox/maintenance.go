package outbox

import (
	"context"
	"log/slog"
	"smart_budget/internal/db/outboxdb"
	app "smart_budget/internal/runtime"
	"time"
)

func (p *Publisher) Maintain(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := p.Cleanup(ctx); err != nil {
				app.LogError("outbox.maintenance", err)
			}
		}
	}
}

func (p *Publisher) Cleanup(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	q := outboxdb.New(p.DB)

	var total int64

	for _, prune := range []func(context.Context) (int64, error){q.PruneSessions, q.PruneAuthAttempts, q.PrunePublishedEvents} {
		n, err := prune(ctx)
		if err != nil {
			return err
		}

		total += n
	}

	status, err := q.OutboxStatus(ctx)
	if err != nil {
		return err
	}

	slog.Info("outbox maintenance", "removed", total, "pending", status.Pending, "quarantined", status.Quarantined)

	return nil
}
