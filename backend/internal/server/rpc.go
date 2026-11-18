package server

import (
	"context"
	"errors"
	"smart_budget/internal/auth"
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"smart_budget/internal/rpc"
	app "smart_budget/internal/runtime"

	"connectrpc.com/connect"
)

func (s *Server) CheckSession(ctx context.Context, req *connect.Request[servicesv1.CheckSessionRequest]) (*connect.Response[servicesv1.CheckSessionResponse], error) {
	if !rpc.ValidID(req.Msg.UserId) || !rpc.ValidID(req.Msg.SessionId) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("invalid session identity"))
	}

	_, err := s.Auth.ActiveSession(ctx, req.Msg.SessionId, req.Msg.UserId)
	if errors.Is(err, auth.ErrUnauthorized) {
		return connect.NewResponse(&servicesv1.CheckSessionResponse{Active: false}), nil
	}

	if err != nil {
		app.LogError("sessions.check", err)

		return nil, connect.NewError(connect.CodeUnavailable, errors.New("session check unavailable"))
	}

	return connect.NewResponse(&servicesv1.CheckSessionResponse{Active: true}), nil
}
