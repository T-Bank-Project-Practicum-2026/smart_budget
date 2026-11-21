package notifications

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"errors"
	"net"
	"net/http"
	"net/url"
	"smart_budget/internal/gen/smartbudget/services/v1/servicesv1connect"
	"smart_budget/internal/rpc"
	app "smart_budget/internal/runtime"
	"strings"
	"time"

	"connectrpc.com/connect"

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
	mux.Handle(servicesv1connect.NewSubscriptionsServiceHandler(s, connect.WithInterceptors(rpc.Authenticate(s.Token)), connect.WithReadMaxBytes(8192)))

	return mux
}
