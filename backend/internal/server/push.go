package server

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"smart_budget/internal/db/apidb"
	"smart_budget/internal/push"
	app "smart_budget/internal/runtime"
	"time"

	"github.com/jackc/pgx/v5"
)

// @Summary Подписка на push
// @Tags Push-уведомления
// @Accept json
// @Produce json
// @Security accessCookie
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Param body body SubscriptionInput true "Push-подписка"
// @Success 204 "Выполнено, без тела ответа"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/push/subscription [put]
func (s *Server) subscribe(w http.ResponseWriter, r *http.Request) {
	s.subscription(w, r)
}

// @Summary Отписка от push
// @Tags Push-уведомления
// @Accept json
// @Produce json
// @Security accessCookie
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Param body body UnsubscribeInput true "Push-подписка"
// @Success 204 "Выполнено, без тела ответа"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/push/subscription [delete]
func (s *Server) unsubscribe(w http.ResponseWriter, r *http.Request) {
	s.subscription(w, r)
}

// @Summary Публичный VAPID-ключ
// @Tags Push-уведомления
// @Produce json
// @Success 200 {object} PublicKeyResponse "Публичный ключ"
// @Router /api/v1/push/key [get]
func (s *Server) publicKey(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, PublicKeyResponse{PublicKey: s.PublicKey})
}

func (s *Server) subscription(w http.ResponseWriter, r *http.Request) {
	identity, err := s.identity(r)
	if err != nil {
		respondError(w, err)

		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 8192))
	if err != nil {
		writeError(w, 400, "invalid_subscription")

		return
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, s.NotificationsURL+"/internal/subscriptions", bytes.NewReader(body))
	if err != nil {
		writeError(w, 503, "service_unavailable")

		return
	}

	req.Header.Set("Authorization", "Bearer "+s.Token)
	req.Header.Set("X-User-ID", identity.UserID)
	req.Header.Set("X-Session-ID", identity.ID)
	req.Header.Set("X-Session-Expires", identity.ExpiresAt.Format(time.RFC3339Nano))
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.Client.Do(req)
	if err != nil {
		app.LogError("subscriptions.proxy", err)
		writeError(w, 503, "service_unavailable")

		return
	}

	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNoContent:
		w.WriteHeader(http.StatusNoContent)

	case http.StatusBadRequest:
		writeError(w, 400, "invalid_subscription")

	case http.StatusConflict:
		writeError(w, 409, "subscription_conflict")

	default:
		writeError(w, 503, "service_unavailable")
	}
}

// @Summary Тестовое уведомление
// @Description Добавляет уведомление в очередь, не чаще раза в 10 секунд. Повторный запрос с тем же Idempotency-Key возвращает тот же ID.
// @Tags Уведомления
// @Produce json
// @Security accessCookie
// @Param X-CSRF-Protection header string true "Защита от CSRF" Enums(1)
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Param Idempotency-Key header string true "Ключ идемпотентности" minLength(16) maxLength(100)
// @Success 202 {object} QueuedResponse "Уведомление в очереди"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/notifications/test [post]
func (s *Server) send(w http.ResponseWriter, r *http.Request) {
	user, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	key := r.Header.Get("Idempotency-Key")
	if len(key) < 16 || len(key) > 100 {
		writeError(w, 400, "invalid_idempotency_key")

		return
	}

	event := push.Event{ID: rand.Text(), UserID: user, Title: "Умный бюджет", Body: "Тестовое уведомление. Всё работает!", CreatedAt: time.Now().UTC()}
	tx, err := s.DB.Begin(r.Context())
	if err != nil {
		respondError(w, err)

		return
	}

	defer func() { _ = tx.Rollback(r.Context()) }()

	queries := apidb.New(tx)
	last, err := queries.LockPushUser(r.Context(), user)
	if err != nil {
		respondError(w, err)

		return
	}

	previous, err := queries.FindNotification(r.Context(), apidb.FindNotificationParams{UserID: user, RequestKey: key})
	if err == nil {
		writeJSON(w, 202, QueuedResponse{ID: previous, Status: "queued"})

		return
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		writeError(w, 503, "service_unavailable")

		return
	}

	if last != nil && time.Since(*last) < 10*time.Second {
		w.Header().Set("Retry-After", "10")
		writeError(w, 429, "push_rate_limited")

		return
	}

	payload, err := json.Marshal(event)
	if err != nil {
		writeError(w, 500, "notification_failed")

		return
	}

	if err := queries.CreateNotification(r.Context(), apidb.CreateNotificationParams{ID: event.ID, UserID: user, RequestKey: key, Title: event.Title, Body: event.Body}); err != nil {
		respondError(w, err)

		return
	}

	if err := queries.EnqueueNotification(r.Context(), apidb.EnqueueNotificationParams{ID: event.ID, Payload: payload}); err != nil {
		respondError(w, err)

		return
	}

	if err := queries.RecordPushTime(r.Context(), user); err != nil {
		respondError(w, err)

		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		respondError(w, err)

		return
	}

	writeJSON(w, 202, QueuedResponse{ID: event.ID, Status: "queued"})
}

// @Summary Последние 20 уведомлений
// @Description Сначала новые.
// @Tags Уведомления
// @Produce json
// @Security accessCookie
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Success 200 {array} push.Event "Список уведомлений"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/notifications [get]
func (s *Server) list(w http.ResponseWriter, r *http.Request) {
	user, err := s.user(r)
	if err != nil {
		respondError(w, err)

		return
	}

	rows, err := apidb.New(s.DB).ListNotifications(r.Context(), user)
	if err != nil {
		respondError(w, err)

		return
	}

	events := make([]push.Event, 0, len(rows))

	for _, row := range rows {
		events = append(events, push.Event{ID: row.ID, Title: row.Title, Body: row.Body, CreatedAt: row.CreatedAt})
	}

	writeJSON(w, 200, events)
}
