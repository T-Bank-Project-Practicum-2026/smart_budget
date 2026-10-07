package notifications

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	mathrand "math/rand/v2"
	"net/http"
	"smart_budget/internal/db/notificationsdb"
	"smart_budget/internal/push"
	app "smart_budget/internal/runtime"
	"strconv"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

func RetryDelay(attempt int, header string, now time.Time) time.Duration {
	delay := min(time.Hour, 30*time.Second*time.Duration(1<<min(max(attempt-1, 0), 7)))
	delay += time.Duration(float64(delay) * mathrand.Float64() * 0.2)
	if seconds, err := strconv.Atoi(header); err == nil && seconds > 0 {
		delay = max(delay, time.Duration(min(seconds, 86400))*time.Second)
	} else if when, err := http.ParseTime(header); err == nil {
		delay = max(delay, min(24*time.Hour, when.Sub(now)))
	}

	return delay
}

func (s *Service) DeliverOne(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 25*time.Second)

	defer cancel()

	token := rand.Text()
	q := notificationsdb.New(s.DB)
	job, err := q.ClaimDelivery(ctx, token)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	delivery, err := q.GetDelivery(ctx, notificationsdb.GetDeliveryParams{EventID: job.EventID, SubscriptionID: job.SubscriptionID})
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}

	if err != nil {
		return true, err
	}

	finish := func(state, reason string, status int, retryAfter string) (bool, error) {
		err := q.CompleteDelivery(ctx, notificationsdb.CompleteDeliveryParams{State: state, LastStatus: int32(status), LastError: reason, NextAt: time.Now().Add(RetryDelay(int(job.Attempts), retryAfter, time.Now())), EventID: job.EventID, SubscriptionID: job.SubscriptionID, LeaseToken: token})
		if err == nil {
			slog.Info("push attempt completed", "event_id", job.EventID, "subscription_id", job.SubscriptionID, "attempt", job.Attempts, "state", state, "reason", reason, "provider_status", status)
		}

		return true, err
	}
	remove := func(reason string) (bool, error) {
		err := q.DeleteSessionSubscription(ctx, notificationsdb.DeleteSessionSubscriptionParams{ID: job.SubscriptionID, SessionID: delivery.SessionID})
		if err == nil {
			slog.Info("push subscription disabled", "subscription_id", job.SubscriptionID, "reason", reason)
		}

		return true, err
	}
	event, err := push.Decode(delivery.Payload)
	if err != nil {
		return finish("failed", "invalid_payload", 0, "")
	}

	if event.UserID != delivery.UserID {
		return finish("failed", "owner_mismatch", 0, "")
	}

	if time.Since(event.CreatedAt) >= 24*time.Hour {
		return finish("failed", "expired_event", 0, "")
	}

	if job.Attempts > 32 {
		return finish("failed", "attempt_limit", 0, "")
	}

	if !delivery.ExpiresAt.After(time.Now()) {
		return remove("expired_session")
	}

	sub := webpush.Subscription{Endpoint: delivery.Endpoint, Keys: webpush.Keys{P256dh: delivery.P256dh, Auth: delivery.Auth}}
	if ValidateSubscription(sub) != nil {
		return finish("failed", "invalid_subscription", 0, "")
	}

	active, err := s.CheckSession(ctx, delivery.UserID, delivery.SessionID)
	if err != nil {
		if job.Attempts >= 32 {
			return finish("failed", "attempt_limit", 0, "")
		}

		return finish("pending", "session_check_unavailable", 0, "")
	}

	if !active {
		return remove("revoked_session")
	}

	ttl := min(3600, int(time.Until(delivery.ExpiresAt).Seconds()))
	if ttl <= 0 {
		return remove("expired_session")
	}

	// A provider may deliver an already accepted message after logout. Keep its
	// payload neutral; account details are available only through authenticated API.
	body, _ := json.Marshal(map[string]string{"id": event.ID, "title": "Умный бюджет", "body": "У вас новое уведомление. Откройте приложение, чтобы прочитать его.", "url": "/"})
	sendCtx, stopSend := context.WithTimeout(ctx, 10*time.Second)
	response, sendErr := webpush.SendNotificationWithContext(sendCtx, body, &sub, &webpush.Options{HTTPClient: s.HTTPClient, Subscriber: s.Subject, VAPIDPublicKey: s.PublicKey, VAPIDPrivateKey: s.PrivateKey, TTL: ttl, Topic: event.ID})
	status, retryAfter := 0, ""
	if response != nil {
		status = response.StatusCode
		retryAfter = response.Header.Get("Retry-After")
		_ = response.Body.Close()
	}

	stopSend()

	switch {
	case status == 404 || status == 410:
		return remove("provider_gone")

	case sendErr == nil && status >= 200 && status < 300:
		return finish("delivered", "", status, "")

	case sendErr != nil || status == 429 || status >= 500:
		if job.Attempts < 32 {
			return finish("pending", "provider_temporary", status, retryAfter)
		}

		return finish("failed", "attempt_limit", status, "")

	default:
		return finish("failed", "provider_rejected", status, "")
	}
}

func (s *Service) Deliver(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)
	workers := s.DeliveryWorkers
	if workers <= 0 {
		workers = 4
	}

	workers = min(workers, 16)

	for range workers {
		group.Go(func() error { return s.deliverLoop(ctx) })
	}

	return group.Wait()
}

func (s *Service) deliverLoop(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		sent, err := s.DeliverOne(ctx)
		if err != nil && ctx.Err() == nil {
			app.LogError("notifications.delivery", err)
		}

		if sent && err == nil {
			continue
		}

		timer := time.NewTimer(time.Second)

		select {
		case <-ctx.Done():
			timer.Stop()

			return ctx.Err()

		case <-timer.C:
		}
	}
}
