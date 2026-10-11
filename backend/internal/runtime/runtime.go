package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/sync/errgroup"
)

func Required(name string) (string, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return "", errors.New(name + " is required")
	}

	return value, nil
}

// LogError records error classes, never SQL text, endpoint URLs or credentials.
func LogError(operation string, err error) {
	var pgErr *pgconn.PgError

	code := ""
	if errors.As(err, &pgErr) {
		code = pgErr.Code
	}

	slog.Error("operation failed", "operation", operation, "error_type", fmt.Sprintf("%T", err), "sqlstate", code)
}

// @Summary Готовность API
// @Tags Состояние сервиса
// @Success 204 "Выполнено, без тела ответа"
// @Failure 503 "PostgreSQL недоступен"
// @Router /api/ready [get]
func Ready(check func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)

		defer cancel()

		if err := check(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func Serve(ctx context.Context, handler http.Handler, workers ...func(context.Context) error) error {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	})
	group.Go(func() error {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)

		defer cancel()

		if err := srv.Shutdown(shutdown); err != nil {
			_ = srv.Close()

			return err
		}

		return nil
	})

	for _, worker := range workers {
		group.Go(func() error {
			err := worker(ctx)
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		})
	}

	return group.Wait()
}

type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"smart_budget"`
}

// @Summary Проверка API
// @Tags Состояние сервиса
// @Produce json
// @Success 200 {object} HealthResponse "Сервис работает"
// @Router /api/health [get]
func Health(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{Status: "ok", Service: service})
	}
}
