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

// LogError опускает текст ошибки, потому что в нём могут быть SQL или секреты.
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
// @Produce json
// @Success 200 {object} ReadyResponse "Сервис готов"
// @Failure 503 {object} ErrorResponse "PostgreSQL недоступен"
// @Router /api/ready [get]
func Ready(check func(context.Context) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)

		defer cancel()

		if err := check(ctx); err != nil {
			WriteError(w, http.StatusServiceUnavailable, "service_unavailable")

			return
		}

		WriteSuccess(w, http.StatusOK, ReadyData{Status: "ok"})
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

type SuccessResponse struct {
	Success bool `json:"success" example:"true"`
	Data    any  `json:"data"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"service_unavailable"`
}

type HealthData struct {
	Status  string `json:"status" example:"ok"`
	Service string `json:"service" example:"smart_budget"`
}

type HealthResponse struct {
	Success bool       `json:"success" example:"true"`
	Data    HealthData `json:"data"`
}

type ReadyData struct {
	Status string `json:"status" example:"ok"`
}

type ReadyResponse struct {
	Success bool      `json:"success" example:"true"`
	Data    ReadyData `json:"data"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteSuccess(w http.ResponseWriter, status int, data any) {
	writeJSON(w, status, SuccessResponse{Success: true, Data: data})
}

func WriteError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, ErrorResponse{Success: false, Error: code})
}

// @Summary Проверка API
// @Tags Состояние сервиса
// @Produce json
// @Success 200 {object} HealthResponse "Сервис работает"
// @Router /api/health [get]
func Health(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		WriteSuccess(w, http.StatusOK, HealthData{Status: "ok", Service: service})
	}
}
