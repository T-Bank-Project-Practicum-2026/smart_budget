package notifications

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"smart_budget/internal/db/notificationsdb"
	app "smart_budget/internal/runtime"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	DB                                    *pgxpool.Pool
	Token, PublicKey, PrivateKey, Subject string
	HTTPClient                            webpush.HTTPClient
	SessionURL                            string
	SessionClient                         *http.Client
	DeliveryWorkers                       int
}

func ValidateSubscription(s webpush.Subscription) error {
	u, err := url.Parse(s.Endpoint)
	if err != nil || len(s.Endpoint) > 4096 || u.Scheme != "https" || u.User != nil || u.Fragment != "" || (u.Port() != "" && u.Port() != "443") {
		return errors.New("invalid push endpoint")
	}

	host := strings.ToLower(u.Hostname())
	if host != "fcm.googleapis.com" && host != "web.push.apple.com" && host != "updates.push.services.mozilla.com" && !strings.HasSuffix(host, ".push.services.mozilla.com") {
		return errors.New("unsupported push provider")
	}

	public, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s.Keys.P256dh, "="))
	if err != nil {
		return errors.New("invalid p256dh key")
	}

	if _, err := ecdh.P256().NewPublicKey(public); err != nil {
		return errors.New("invalid p256dh key")
	}

	auth, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(s.Keys.Auth, "="))
	if err != nil || len(auth) != 16 {
		return errors.New("invalid auth key")
	}

	return nil
}

func PushHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }, Transport: &http.Transport{
		TLSHandshakeTimeout: 5 * time.Second,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}

			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}

			for _, ip := range ips {
				if !ip.IP.IsGlobalUnicast() || ip.IP.IsPrivate() || ip.IP.IsLoopback() || ip.IP.IsLinkLocalUnicast() {
					return nil, errors.New("push endpoint must be public")
				}
			}

			for _, ip := range ips {
				conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
				if err == nil {
					return conn, nil
				}
			}

			return nil, errors.New("push endpoint unavailable")
		},
	}}
}

func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", app.Health("notifications"))
	mux.HandleFunc("GET /api/ready", app.Ready(s.DB.Ping))
	mux.HandleFunc("/internal/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+s.Token)) != 1 || s.Token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)

			return
		}

		user := r.Header.Get("X-User-ID")
		if user == "" || len(user) > 100 {
			http.Error(w, "invalid user", 400)

			return
		}

		if r.Method != "PUT" && r.Method != "DELETE" {
			w.WriteHeader(405)

			return
		}

		var sub webpush.Subscription

		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192)).Decode(&sub) != nil {
			http.Error(w, "invalid subscription", 400)

			return
		}

		if r.Method == "DELETE" {
			if sub.Endpoint == "" {
				w.WriteHeader(400)

				return
			}

			if err := notificationsdb.New(s.DB).DeleteUserSubscription(r.Context(), notificationsdb.DeleteUserSubscriptionParams{UserID: user, Endpoint: sub.Endpoint}); err != nil {
				app.LogError("subscriptions.delete", err)
				w.WriteHeader(503)

				return
			}

			w.WriteHeader(204)

			return
		}

		if ValidateSubscription(sub) != nil {
			http.Error(w, "invalid subscription", 400)

			return
		}

		sessionID := r.Header.Get("X-Session-ID")
		expires, err := time.Parse(time.RFC3339Nano, r.Header.Get("X-Session-Expires"))
		if err != nil || sessionID == "" || len(sessionID) > 100 || !expires.After(time.Now()) {
			http.Error(w, "invalid session", 400)

			return
		}

		affected, err := notificationsdb.New(s.DB).UpsertSubscription(r.Context(), notificationsdb.UpsertSubscriptionParams{ID: rand.Text(), UserID: user, Endpoint: sub.Endpoint, P256dh: sub.Keys.P256dh, Auth: sub.Keys.Auth, SessionID: sessionID, ExpiresAt: expires})
		if err != nil {
			app.LogError("subscriptions.upsert", err)
			w.WriteHeader(503)

			return
		}

		if affected == 0 {
			w.WriteHeader(409)

			return
		}

		w.WriteHeader(204)
	})

	return mux
}
