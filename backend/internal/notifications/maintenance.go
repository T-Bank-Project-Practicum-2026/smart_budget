package notifications

import (
	"context"
	"log/slog"
	"smart_budget/internal/db/notificationsdb"
	app "smart_budget/internal/runtime"
	"time"
)

func (s *Service) Maintain(ctx context.Context) error {
	ticker := time.NewTicker(time.Minute)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:
			if err := s.Cleanup(ctx); err != nil {
				app.LogError("notifications.maintenance", err)
			}
		}
	}
}

func (s *Service) Cleanup(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	q := notificationsdb.New(s.DB)

	var total int64

	for _, prune := range []func(context.Context) (int64, error){q.PruneSubscriptions, q.PruneJobs, q.PruneInbox, q.PruneRejectedEvents} {
		n, err := prune(ctx)
		if err != nil {
			return err
		}

		total += n
	}

	status, err := q.DeliveryStatus(ctx)
	if err != nil {
		return err
	}

	slog.Info("notifications maintenance", "removed", total, "pending", status.Pending, "failed", status.Failed)

	return nil
}
