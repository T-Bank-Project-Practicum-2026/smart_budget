package notifications

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"smart_budget/internal/auth"
	"smart_budget/internal/db/notificationsdb"
	"smart_budget/internal/push"
	"smart_budget/internal/server"
	"smart_budget/internal/testutil"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/twmb/franz-go/pkg/kgo"
)

type pushClientFunc func(*http.Request) (*http.Response, error)

func (f pushClientFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

func sessionFixture(t *testing.T, db *pgxpool.Pool) *httptest.Server {
	t.Helper()
	_, err := db.Exec(t.Context(), `INSERT INTO app.users(id,email,password_hash) VALUES('u','u@example.test','unused');
 INSERT INTO app.auth_sessions(id,user_id,refresh_hash,created_at,expires_at) VALUES('session','u','hash',now()-interval '1 day',now()+interval '1 day');`)
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := auth.NewTokens(strings.Repeat("x", 32))
	if err != nil {
		t.Fatal(err)
	}

	service, err := auth.NewService(db, tokens)
	if err != nil {
		t.Fatal(err)
	}

	api := &server.Server{DB: db, Auth: service, Token: "internal-test-token"}
	httpServer := httptest.NewServer(api.Handler())
	t.Cleanup(httpServer.Close)

	return httpServer
}

func subscriptionFixture(t *testing.T, db *pgxpool.Pool, id string) {
	t.Helper()
	private, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	_, err = notificationsdb.New(db).UpsertSubscription(t.Context(), notificationsdb.UpsertSubscriptionParams{ID: id, UserID: "u", Endpoint: "https://fcm.googleapis.com/" + id, P256dh: base64.RawURLEncoding.EncodeToString(private.PublicKey().Bytes()), Auth: base64.RawURLEncoding.EncodeToString(make([]byte, 16)), SessionID: "session", ExpiresAt: time.Now().Add(time.Hour)})
	if err != nil {
		t.Fatal(err)
	}
}

func eventRecord(t *testing.T, id string, created time.Time) *kgo.Record {
	t.Helper()
	body, err := json.Marshal(push.Event{ID: id, UserID: "u", Title: "private-title", Body: "private-body", CreatedAt: created})
	if err != nil {
		t.Fatal(err)
	}

	return &kgo.Record{Topic: push.Topic, Offset: 1, Value: body}
}

func configuredService(t *testing.T, db *pgxpool.Pool, api *httptest.Server, calls *atomic.Int32) *Service {
	t.Helper()
	private, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	return &Service{DB: db, SessionURL: api.URL, Token: "internal-test-token", PublicKey: base64.RawURLEncoding.EncodeToString(private.PublicKey().Bytes()), PrivateKey: base64.RawURLEncoding.EncodeToString(private.Bytes()), Subject: "mailto:test@example.test", HTTPClient: pushClientFunc(func(r *http.Request) (*http.Response, error) {
		calls.Add(1)

		return &http.Response{StatusCode: 201, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
}

func scalar(t *testing.T, db *pgxpool.Pool, query string) int {
	t.Helper()

	var n int

	if err := db.QueryRow(t.Context(), query).Scan(&n); err != nil {
		t.Fatal(err)
	}

	return n
}

func TestSessionDeliveryLifecycle(t *testing.T) {
	for _, scenario := range []string{"closed_tab", "revoked", "expired", "auth_unavailable"} {
		t.Run(scenario, func(t *testing.T) {
			db := testutil.Database(t)
			api := sessionFixture(t, db)
			subscriptionFixture(t, db, "one")

			var calls atomic.Int32

			svc := configuredService(t, db, api, &calls)
			record := eventRecord(t, "event", time.Now())
			if err := svc.Accept(t.Context(), record); err != nil {
				t.Fatal(err)
			}

			if err := svc.Accept(t.Context(), record); err != nil {
				t.Fatal(err)
			}

			if got := scalar(t, db, "SELECT count(*) FROM notifications.jobs"); got != 1 {
				t.Fatalf("duplicate jobs: %d", got)
			}

			switch scenario {
			case "revoked":
				if _, err := db.Exec(t.Context(), "UPDATE app.auth_sessions SET revoked_at=now()"); err != nil {
					t.Fatal(err)
				}

			case "expired":
				if _, err := db.Exec(t.Context(), "UPDATE app.auth_sessions SET expires_at=now()-interval '1 second'"); err != nil {
					t.Fatal(err)
				}

			case "auth_unavailable":
				api.Close()
			}

			if _, err := svc.DeliverOne(t.Context()); err != nil {
				t.Fatal(err)
			}

			want := int32(0)
			if scenario == "closed_tab" {
				want = 1
			}

			if calls.Load() != want {
				t.Fatalf("push calls=%d, want %d", calls.Load(), want)
			}

			if scenario == "revoked" || scenario == "expired" {
				if got := scalar(t, db, "SELECT count(*) FROM notifications.subscriptions"); got != 0 {
					t.Fatal("revoked subscription retained")
				}
			}

			if scenario == "auth_unavailable" {
				if got := scalar(t, db, "SELECT count(*) FROM notifications.jobs WHERE state='pending' AND last_error='session_check_unavailable'"); got != 1 {
					t.Fatal("must retry without sending")
				}
			}
		})
	}
}

func TestMultipleBrowsersAndLeaseRecovery(t *testing.T) {
	db := testutil.Database(t)
	api := sessionFixture(t, db)
	subscriptionFixture(t, db, "one")
	subscriptionFixture(t, db, "two")

	var calls atomic.Int32

	svc := configuredService(t, db, api, &calls)
	if err := svc.Accept(t.Context(), eventRecord(t, "event", time.Now())); err != nil {
		t.Fatal(err)
	}

	if got := scalar(t, db, "SELECT count(*) FROM notifications.jobs"); got != 2 {
		t.Fatalf("jobs=%d", got)
	}

	q := notificationsdb.New(db)
	if err := q.DeleteUserSubscription(t.Context(), notificationsdb.DeleteUserSubscriptionParams{UserID: "other-user", Endpoint: "https://fcm.googleapis.com/one"}); err != nil {
		t.Fatal(err)
	}

	if got := scalar(t, db, "SELECT count(*) FROM notifications.subscriptions"); got != 2 {
		t.Fatal("foreign delete")
	}

	if err := q.DeleteUserSubscription(t.Context(), notificationsdb.DeleteUserSubscriptionParams{UserID: "u", Endpoint: "https://fcm.googleapis.com/one"}); err != nil {
		t.Fatal(err)
	}

	if _, err := q.ClaimDelivery(t.Context(), "crashed-worker"); err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(t.Context(), "UPDATE notifications.jobs SET lease_until=now()-interval '1 second'"); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.DeliverOne(t.Context()); err != nil {
		t.Fatal(err)
	}

	if calls.Load() != 1 {
		t.Fatal("lease not recovered")
	}

	if got := scalar(t, db, "SELECT count(*) FROM notifications.jobs WHERE state='delivered'"); got != 1 {
		t.Fatal("delivery not completed")
	}
}

func TestConcurrentDeliveryAndCleanup(t *testing.T) {
	db := testutil.Database(t)
	api := sessionFixture(t, db)

	var calls atomic.Int32

	svc := configuredService(t, db, api, &calls)

	for _, id := range []string{"one", "two", "three", "four"} {
		subscriptionFixture(t, db, id)
	}

	if err := svc.Accept(t.Context(), eventRecord(t, "event", time.Now())); err != nil {
		t.Fatal(err)
	}

	entered := make(chan struct{}, 4)
	release := make(chan struct{})
	svc.HTTPClient = pushClientFunc(func(r *http.Request) (*http.Response, error) {
		entered <- struct{}{}

		select {
		case <-release:
		case <-r.Context().Done():
			return nil, r.Context().Err()
		}

		calls.Add(1)

		return &http.Response{StatusCode: 201, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(""))}, nil
	})
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)

	defer cancel()

	done := make(chan error, 1)

	go func() { done <- svc.Deliver(ctx) }()

	for range 4 {
		select {
		case <-entered:
		case <-ctx.Done():
			t.Fatal("deliveries were not concurrent")
		}
	}

	close(release)
	deadline := time.Now().Add(5 * time.Second)

	for scalar(t, db, "SELECT count(*) FROM notifications.jobs WHERE state='delivered'") != 4 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	cancel()
	<-done
	if calls.Load() != 4 {
		t.Fatalf("calls=%d", calls.Load())
	}

	if _, err := db.Exec(t.Context(), "UPDATE notifications.inbox SET received_at=now()-interval '31 days'"); err != nil {
		t.Fatal(err)
	}

	if err := svc.Cleanup(t.Context()); err != nil {
		t.Fatal(err)
	}

	if got := scalar(t, db, "SELECT count(*) FROM notifications.inbox"); got != 0 {
		t.Fatal("inbox retained")
	}

	if err := svc.Accept(t.Context(), eventRecord(t, "event", time.Now().Add(-31*24*time.Hour))); err != nil {
		t.Fatal(err)
	}

	if got := scalar(t, db, "SELECT count(*) FROM notifications.jobs"); got != 0 {
		t.Fatal("old replay recreated jobs")
	}
}

func TestProviderStatuses(t *testing.T) {
	for _, status := range []int{404, 410, 429, 503, 400} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			db := testutil.Database(t)
			api := sessionFixture(t, db)
			subscriptionFixture(t, db, "one")

			var calls atomic.Int32

			svc := configuredService(t, db, api, &calls)
			svc.HTTPClient = pushClientFunc(func(r *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: status, Header: http.Header{"Retry-After": []string{"600"}}, Body: io.NopCloser(strings.NewReader(""))}, nil
			})
			if err := svc.Accept(t.Context(), eventRecord(t, "event", time.Now())); err != nil {
				t.Fatal(err)
			}

			if _, err := svc.DeliverOne(t.Context()); err != nil {
				t.Fatal(err)
			}

			switch status {
			case 404, 410:
				if got := scalar(t, db, "SELECT count(*) FROM notifications.subscriptions"); got != 0 {
					t.Fatal("gone subscription retained")
				}

			case 429, 503:
				if got := scalar(t, db, "SELECT count(*) FROM notifications.jobs WHERE state='pending' AND last_error='provider_temporary' AND next_at>now()+interval '590 seconds'"); got != 1 {
					t.Fatal("retry or Retry-After lost")
				}

			default:
				if got := scalar(t, db, "SELECT count(*) FROM notifications.jobs WHERE state='failed' AND last_error='provider_rejected'"); got != 1 {
					t.Fatal("permanent failure must stop")
				}
			}
		})
	}
}
