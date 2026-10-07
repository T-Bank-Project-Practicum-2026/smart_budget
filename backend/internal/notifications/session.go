package notifications

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

// CheckSession asks the auth owner; no browser tab or access cookie is needed.
// Unavailability is an error, never permission to send.
func (s *Service) CheckSession(ctx context.Context, userID, sessionID string) (bool, error) {
	if s.SessionURL == "" || s.Token == "" || sessionID == "" {
		return false, errors.New("session check is not configured")
	}

	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)

	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.SessionURL+"/internal/sessions/"+url.PathEscape(sessionID)+"?userId="+url.QueryEscape(userID), nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+s.Token)
	client := s.SessionClient
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}
	}

	response, err := client.Do(req)
	if err != nil {
		return false, err
	}

	defer func() { _ = response.Body.Close() }()

	switch response.StatusCode {
	case http.StatusNoContent:
		return true, nil

	case http.StatusUnauthorized:
		return false, nil

	default:
		return false, errors.New("session check unavailable")
	}
}
