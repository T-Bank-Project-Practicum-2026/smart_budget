package notifications

import (
	"context"
	"errors"
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"smart_budget/internal/gen/smartbudget/services/v1/servicesv1connect"
	"smart_budget/internal/rpc"
	"time"

	"connectrpc.com/connect"
)

// CheckSession asks the auth owner; unavailability never permits delivery.
func (s *Service) CheckSession(ctx context.Context, userID, sessionID string) (bool, error) {
	if s.SessionURL == "" || s.Token == "" || sessionID == "" {
		return false, errors.New("session check is not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)

	defer cancel()

	client := servicesv1connect.NewAuthServiceClient(rpc.Client(s.SessionClient), s.SessionURL)
	req := connect.NewRequest(&servicesv1.CheckSessionRequest{UserId: userID, SessionId: sessionID})
	req.Header().Set("Authorization", "Bearer "+s.Token)
	response, err := client.CheckSession(ctx, req)
	if err != nil {
		return false, err
	}

	return response.Msg.Active, nil
}
