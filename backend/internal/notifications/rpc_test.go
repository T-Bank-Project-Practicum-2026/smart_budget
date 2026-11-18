package notifications

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"smart_budget/internal/auth"
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"smart_budget/internal/gen/smartbudget/services/v1/servicesv1connect"
	"smart_budget/internal/push"
	"smart_budget/internal/server"
	"smart_budget/internal/testutil"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSubscriptionRPCAndPublicAdapter(t *testing.T) {
	db := testutil.Database(t)
	api := sessionFixture(t, db)
	svc := &Service{DB: db, Token: "internal-test-token", SessionURL: api.URL}

	var contentType atomic.Value

	handler := svc.Handler()
	endpoint := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType.Store(r.Header.Get("Content-Type"))
		handler.ServeHTTP(w, r)
	}))

	defer endpoint.Close()

	client := servicesv1connect.NewSubscriptionsServiceClient(endpoint.Client(), endpoint.URL)
	private, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	req := connect.NewRequest(&servicesv1.UpsertSubscriptionRequest{UserId: "u", SessionId: "session", SessionExpiresAt: timestamppb.New(time.Now().Add(time.Hour)), Endpoint: "https://fcm.googleapis.com/rpc", P256Dh: base64.RawURLEncoding.EncodeToString(private.PublicKey().Bytes()), Auth: base64.RawURLEncoding.EncodeToString(make([]byte, 16))})
	if _, err := client.UpsertSubscription(t.Context(), req); connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("missing credential: %v", err)
	}

	req.Header().Set("Authorization", "Bearer internal-test-token")
	if _, err := client.UpsertSubscription(t.Context(), req); err != nil {
		t.Fatal(err)
	}

	if contentType.Load() != "application/proto" {
		t.Fatalf("wire content type: %s", contentType.Load())
	}

	req.Msg.UserId = "foreign"
	if _, err := client.UpsertSubscription(t.Context(), req); connect.CodeOf(err) != connect.CodeAlreadyExists {
		t.Fatalf("foreign overwrite: %v", err)
	}

	remove := connect.NewRequest(&servicesv1.DeleteSubscriptionRequest{UserId: "foreign", Endpoint: req.Msg.Endpoint})
	remove.Header().Set("Authorization", "Bearer internal-test-token")
	if _, err := client.DeleteSubscription(t.Context(), remove); err != nil {
		t.Fatal(err)
	}

	if scalar(t, db, "SELECT count(*) FROM notifications.subscriptions") != 1 {
		t.Fatal("foreign delete")
	}

	req.Msg.UserId = "u"
	req.Msg.SessionExpiresAt = nil
	if _, err := client.UpsertSubscription(t.Context(), req); connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("missing expiry: %v", err)
	}

	tokens, err := auth.NewTokens(strings.Repeat("x", 32))
	if err != nil {
		t.Fatal(err)
	}

	authService, err := auth.NewService(db, tokens)
	if err != nil {
		t.Fatal(err)
	}

	access, err := tokens.Issue("u", "session", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	public := (&server.Server{DB: db, Auth: authService, Token: svc.Token, NotificationsURL: endpoint.URL, Client: endpoint.Client(), Origins: []string{"http://localhost:3000"}}).Handler()

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		body, err := json.Marshal(map[string]any{"endpoint": req.Msg.Endpoint, "keys": map[string]string{"p256dh": req.Msg.P256Dh, "auth": req.Msg.Auth}})
		if err != nil {
			t.Fatal(err)
		}

		request := httptest.NewRequestWithContext(t.Context(), method, "/api/v1/push/subscription", strings.NewReader(string(body)))
		request.Header.Set("Origin", "http://localhost:3000")
		request.Header.Set("X-CSRF-Protection", "1")
		request.AddCookie(&http.Cookie{Name: "budget_access", Value: access})
		result := httptest.NewRecorder()
		public.ServeHTTP(result, request)
		if result.Code != http.StatusOK || strings.TrimSpace(result.Body.String()) != `{"success":true,"data":null}` {
			t.Fatalf("public %s: %d %s", method, result.Code, result.Body.String())
		}
	}

	if scalar(t, db, "SELECT count(*) FROM notifications.subscriptions") != 0 {
		t.Fatal("subscription retained")
	}

	// The public JSON endpoint must write the generated contract atomically and remain idempotent.
	for range 2 {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/notifications/test", nil)
		request.Header.Set("Origin", "http://localhost:3000")
		request.Header.Set("X-CSRF-Protection", "1")
		request.Header.Set("Idempotency-Key", "protobuf-test-key-0001")
		request.AddCookie(&http.Cookie{Name: "budget_access", Value: access})
		result := httptest.NewRecorder()
		public.ServeHTTP(result, request)
		if result.Code != http.StatusAccepted {
			t.Fatalf("enqueue: %d %s", result.Code, result.Body.String())
		}
	}

	if scalar(t, db, "SELECT count(*) FROM app.outbox") != 1 {
		t.Fatal("duplicate outbox events")
	}

	var payload []byte

	if err := db.QueryRow(t.Context(), "SELECT payload FROM app.outbox").Scan(&payload); err != nil {
		t.Fatal(err)
	}

	message := &servicesv1.NotificationRequested{}
	if err := protojson.Unmarshal(payload, message); err != nil || message.UserId != "u" || message.CreatedAt.CheckValid() != nil {
		t.Fatalf("outbox contract: %v %v", message, err)
	}

	active, err := svc.CheckSession(t.Context(), "foreign", "session")
	if err != nil || active {
		t.Fatalf("foreign session: %v %v", active, err)
	}

	svc.Token = "wrong-token"
	if _, err := svc.CheckSession(t.Context(), "u", "session"); err == nil {
		t.Fatal("invalid service credential interpreted as inactive session")
	}
}

func TestProtobufReplayAndMalformedPayload(t *testing.T) {
	db := testutil.Database(t)
	subscriptionFixture(t, db, "one")
	svc := &Service{DB: db}
	stored, err := push.Encode(push.Event{ID: "replay", UserID: "u", Title: "title", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}

	record := eventRecord(t, "replay", time.Now())

	for range 2 {
		if err := svc.Accept(t.Context(), record); err != nil {
			t.Fatal(err)
		}
	}

	if scalar(t, db, "SELECT count(*) FROM notifications.jobs") != 1 {
		t.Fatal("replay duplicated delivery jobs")
	}

	if err := svc.Accept(t.Context(), &kgo.Record{Topic: push.Topic, Offset: 2, Value: stored}); err != nil {
		t.Fatal(err)
	}

	if err := svc.Accept(t.Context(), &kgo.Record{Topic: push.Topic, Offset: 3, Value: []byte{0xff}}); err != nil {
		t.Fatal(err)
	}

	if scalar(t, db, "SELECT count(*) FROM notifications.rejected_events") != 2 {
		t.Fatal("invalid payloads not quarantined")
	}
}
