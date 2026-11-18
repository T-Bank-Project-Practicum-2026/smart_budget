package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"smart_budget/internal/db/apidb"
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"smart_budget/internal/gen/smartbudget/services/v1/servicesv1connect"
	"smart_budget/internal/push"
	"smart_budget/internal/rpc"
	app "smart_budget/internal/runtime"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

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
// @Success 200 {object} SuccessResponse "Подписка сохранена"
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
// @Success 200 {object} SuccessResponse "Подписка удалена"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/push/subscription [delete]
func (s *Server) unsubscribe(w http.ResponseWriter, r *http.Request) {
	s.subscription(w, r)
}

// @Summary Публичный VAPID-ключ
// @Tags Push-уведомления
// @Produce json
// @Success 200 {object} PublicKeySuccessResponse "Публичный ключ"
// @Router /api/v1/push/key [get]
func (s *Server) publicKey(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, http.StatusOK, PublicKeyResponse{PublicKey: s.PublicKey})
}

func (s *Server) subscription(w http.ResponseWriter, r *http.Request) {
	identity, err := s.identity(r)
	if err != nil {
		respondError(w, err)

		return
	}

	var input SubscriptionInput

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8192))
	if decoder.Decode(&input) != nil || !errors.Is(decoder.Decode(&struct{}{}), io.EOF) {
		writeError(w, 400, "invalid_subscription")

		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)

	defer cancel()

	client := servicesv1connect.NewSubscriptionsServiceClient(rpc.Client(s.Client), s.NotificationsURL)
	if r.Method == http.MethodDelete {
		req := connect.NewRequest(&servicesv1.DeleteSubscriptionRequest{UserId: identity.UserID, Endpoint: input.Endpoint})
		req.Header().Set("Authorization", "Bearer "+s.Token)
		_, err = client.DeleteSubscription(ctx, req)
	} else {
		req := connect.NewRequest(&servicesv1.UpsertSubscriptionRequest{UserId: identity.UserID, SessionId: identity.ID, SessionExpiresAt: timestamppb.New(identity.ExpiresAt), Endpoint: input.Endpoint, P256Dh: input.Keys.P256dh, Auth: input.Keys.Auth})
		req.Header().Set("Authorization", "Bearer "+s.Token)
		_, err = client.UpsertSubscription(ctx, req)
	}

	if err == nil {
		writeSuccess(w, http.StatusOK, nil)

		return
	}

	switch connect.CodeOf(err) {
	case connect.CodeInvalidArgument:
		writeError(w, 400, "invalid_subscription")

	case connect.CodeAlreadyExists:
		writeError(w, 409, "subscription_conflict")

	default:
		app.LogError("subscriptions.rpc", err)
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
// @Success 202 {object} QueuedSuccessResponse "Уведомление в очереди"
// @Failure default {object} ErrorResponse "Ошибка запроса"
// @Router /api/v1/notifications/test [post]
func (s *Server) send(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		respondError(w, err)

		return
	}

	defer func() { _ = tx.Rollback(ctx) }()

	queries := apidb.New(tx)
	last, err := queries.LockPushUser(ctx, user)
	if err != nil {
		respondError(w, err)

		return
	}

	previous, err := queries.FindNotification(ctx, apidb.FindNotificationParams{UserID: user, RequestKey: key})
	if err == nil {
		writeSuccess(w, 202, QueuedResponse{ID: previous, Status: "queued"})

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

	payload, err := push.Encode(event)
	if err != nil {
		writeError(w, 500, "notification_failed")

		return
	}

	if err := queries.CreateNotification(ctx, apidb.CreateNotificationParams{ID: event.ID, UserID: user, RequestKey: key, Title: event.Title, Body: event.Body}); err != nil {
		respondError(w, err)

		return
	}

	if err := queries.EnqueueNotification(ctx, apidb.EnqueueNotificationParams{ID: event.ID, Payload: payload}); err != nil {
		respondError(w, err)

		return
	}

	if err := queries.RecordPushTime(ctx, user); err != nil {
		respondError(w, err)

		return
	}

	if err := tx.Commit(ctx); err != nil {
		respondError(w, err)

		return
	}

	writeSuccess(w, 202, QueuedResponse{ID: event.ID, Status: "queued"})
}

// @Summary Последние 20 уведомлений
// @Description Сначала новые.
// @Tags Уведомления
// @Produce json
// @Security accessCookie
// @Param X-Account-ID header string false "ID пользователя. Если не совпадает с сессией — ошибка 409"
// @Success 200 {object} NotificationsResponse "Список уведомлений"
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

	writeSuccess(w, 200, events)
}
