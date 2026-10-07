package notifications

import (
	"context"
	"log/slog"
	"smart_budget/internal/db/notificationsdb"
	"smart_budget/internal/push"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

func (s *Service) Accept(ctx context.Context, record *kgo.Record) error {
	event, err := push.Decode(record.Value)
	reason := ""
	if err != nil {
		reason = "invalid_payload"
	} else if time.Since(event.CreatedAt) >= 24*time.Hour {
		reason = "expired"
	} else if event.CreatedAt.After(time.Now().Add(5 * time.Minute)) {
		reason = "future_event"
	}

	if reason != "" {
		err = notificationsdb.New(s.DB).RejectEvent(ctx, notificationsdb.RejectEventParams{Topic: record.Topic, Partition: record.Partition, RecordOffset: record.Offset, Reason: reason})
		if err == nil {
			slog.Warn("notification event rejected", "topic", record.Topic, "partition", record.Partition, "offset", record.Offset, "reason", reason)
		}

		return err
	}

	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback(ctx) }()

	queries := notificationsdb.New(tx)
	inserted, err := queries.AcceptEvent(ctx, notificationsdb.AcceptEventParams{ID: event.ID, Payload: record.Value})
	if err != nil {
		return err
	}

	if inserted > 0 {
		err = queries.CreateDeliveryJobs(ctx, notificationsdb.CreateDeliveryJobsParams{EventID: event.ID, UserID: event.UserID})
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Service) Consume(client *kgo.Client) func(context.Context) error {
	return func(ctx context.Context) error {
		defer client.AllowRebalance()

		for {
			fetches := client.PollRecords(ctx, 100)
			if ctx.Err() != nil {
				return ctx.Err()
			}

			if errs := fetches.Errors(); len(errs) > 0 {
				client.AllowRebalance()

				return errs[0].Err
			}

			for _, record := range fetches.Records() {
				if err := s.Accept(ctx, record); err != nil {
					client.AllowRebalance()

					return err
				}

				if err := client.CommitRecords(ctx, record); err != nil {
					client.AllowRebalance()

					return err
				}
			}

			client.AllowRebalance()
		}
	}
}
