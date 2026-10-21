package server

import (
	"errors"
	"net/http"
	"os"
	"smart_budget/internal/auth"
	"smart_budget/internal/importer"
	"smart_budget/internal/swagger"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	app "smart_budget/internal/runtime"
)

type Server struct {
	Auth                               *auth.Service
	Imports                            *importer.Service
	SecureCookies                      bool
	DB                                 *pgxpool.Pool
	NotificationsURL, Token, PublicKey string
	Origins                            []string
	Client                             *http.Client
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", app.Health("smart_budget"))
	mux.HandleFunc("GET /api/ready", app.Ready(s.DB.Ping))
	mux.HandleFunc("GET /swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	mux.Handle("GET /swagger/", swagger.Handler())
	mux.HandleFunc("GET /internal/sessions/{id}", s.sessionStatus)
	mux.HandleFunc("POST /api/v1/auth/register", s.register)
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/refresh", s.refresh)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	mux.HandleFunc("GET /api/v1/auth/me", s.me)
	mux.HandleFunc("POST /api/v1/imports/tbank", s.createTBankImport)
	mux.HandleFunc("GET /api/v1/imports", s.listImports)
	mux.HandleFunc("GET /api/v1/imports/{id}", s.getImport)
	mux.HandleFunc("GET /api/v1/operations", s.listOperations)
	mux.HandleFunc("GET /api/v1/push/key", s.publicKey)
	mux.HandleFunc("PUT /api/v1/push/subscription", s.subscribe)
	mux.HandleFunc("DELETE /api/v1/push/subscription", s.unsubscribe)
	mux.HandleFunc("POST /api/v1/notifications/test", s.send)
	mux.HandleFunc("GET /api/v1/notifications", s.list)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != "GET" && r.Method != "HEAD" {
			allowed := false

			for _, origin := range s.Origins {
				if r.Header.Get("Origin") == origin {
					allowed = true
				}
			}

			if !allowed || r.Header.Get("X-CSRF-Protection") != "1" {
				writeError(w, http.StatusForbidden, "forbidden_origin")

				return
			}
		}

		mux.ServeHTTP(w, r)
	})
}

func New(db *pgxpool.Pool) (*Server, error) {
	url, err := app.Required("NOTIFICATIONS_URL")
	if err != nil {
		return nil, err
	}

	token, err := app.Required("NOTIFICATIONS_TOKEN")
	if err != nil {
		return nil, err
	}

	key, err := app.Required("VAPID_PUBLIC_KEY")
	if err != nil {
		return nil, err
	}

	origins, err := app.Required("APP_ORIGINS")
	if err != nil {
		return nil, err
	}

	secret, err := app.Required("JWT_SECRET")
	if err != nil {
		return nil, err
	}

	tokens, err := auth.NewTokens(secret)
	if err != nil {
		return nil, err
	}

	service, err := auth.NewService(db, tokens)
	if err != nil {
		return nil, err
	}

	imports := importer.New(db)

	secure := false
	if value := os.Getenv("AUTH_COOKIE_SECURE"); value != "" {
		secure, err = strconv.ParseBool(value)
		if err != nil {
			return nil, errors.New("invalid AUTH_COOKIE_SECURE")
		}
	}

	for _, origin := range strings.Split(origins, ",") {
		if strings.HasPrefix(origin, "https://") && !secure {
			return nil, errors.New("HTTPS origins require AUTH_COOKIE_SECURE=true")
		}
	}

	return &Server{Auth: service, Imports: imports, SecureCookies: secure, DB: db, NotificationsURL: url, Token: token, PublicKey: key, Origins: strings.Split(origins, ","), Client: &http.Client{Timeout: 10 * time.Second}}, nil
}
