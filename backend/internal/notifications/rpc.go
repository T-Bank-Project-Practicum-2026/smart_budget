package notifications

import (
	"context"
	"crypto/rand"
	"errors"
	"smart_budget/internal/db/notificationsdb"
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"smart_budget/internal/rpc"
	app "smart_budget/internal/runtime"
	"time"

	"connectrpc.com/connect"
	webpush "github.com/SherClockHolmes/webpush-go"
)

func (s *Service) UpsertSubscription(ctx context.Context, req *connect.Request[servicesv1.UpsertSubscriptionRequest]) (*connect.Response[servicesv1.UpsertSubscriptionResponse], error) {
	m := req.Msg
	sub := webpush.Subscription{Endpoint: m.Endpoint, Keys: webpush.Keys{P256dh: m.P256Dh, Auth: m.Auth}}
	if !rpc.ValidID(m.UserId) || !rpc.ValidID(m.SessionId) || m.SessionExpiresAt.CheckValid() != nil || !m.SessionExpiresAt.AsTime().After(time.Now()) || ValidateSubscription(sub) != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid subscription"))
	}

	affected, err := notificationsdb.New(s.DB).UpsertSubscription(ctx, notificationsdb.UpsertSubscriptionParams{ID: rand.Text(), UserID: m.UserId, Endpoint: m.Endpoint, P256dh: m.P256Dh, Auth: m.Auth, SessionID: m.SessionId, ExpiresAt: m.SessionExpiresAt.AsTime()})
	if err != nil {
		app.LogError("subscriptions.upsert", err)

		return nil, connect.NewError(connect.CodeUnavailable, errors.New("subscription unavailable"))
	}

	if affected == 0 {
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("subscription conflict"))
	}

	return connect.NewResponse(&servicesv1.UpsertSubscriptionResponse{}), nil
}

func (s *Service) DeleteSubscription(ctx context.Context, req *connect.Request[servicesv1.DeleteSubscriptionRequest]) (*connect.Response[servicesv1.DeleteSubscriptionResponse], error) {
	m := req.Msg
	if !rpc.ValidID(m.UserId) || m.Endpoint == "" || len(m.Endpoint) > 4096 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid subscription"))
	}

	if err := notificationsdb.New(s.DB).DeleteUserSubscription(ctx, notificationsdb.DeleteUserSubscriptionParams{UserID: m.UserId, Endpoint: m.Endpoint}); err != nil {
		app.LogError("subscriptions.delete", err)

		return nil, connect.NewError(connect.CodeUnavailable, errors.New("subscription unavailable"))
	}

	return connect.NewResponse(&servicesv1.DeleteSubscriptionResponse{}), nil
}
