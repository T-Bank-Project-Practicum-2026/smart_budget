package server

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"smart_budget/internal/auth"
	app "smart_budget/internal/runtime"
)

var errInvalidRequest = errors.New("invalid_request")

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return errInvalidRequest
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errInvalidRequest
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, ErrorResponse{Error: code})
}

func respondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errInvalidRequest):
		writeError(w, 400, "invalid_request")

	case errors.Is(err, auth.ErrInvalidEmail):
		writeError(w, 400, "invalid_email")

	case errors.Is(err, auth.ErrInvalidPassword):
		writeError(w, 400, "invalid_password")

	case errors.Is(err, auth.ErrInvalidTimezone):
		writeError(w, 400, "invalid_timezone")

	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, 401, "invalid_credentials")

	case errors.Is(err, auth.ErrUnauthorized):
		writeError(w, 401, "authentication_required")

	case errors.Is(err, auth.ErrAccountExists):
		writeError(w, 409, "account_exists")

	case errors.Is(err, auth.ErrAccountChanged):
		writeError(w, 409, "account_changed")

	case errors.Is(err, auth.ErrTooManyAttempts):
		w.Header().Set("Retry-After", "900")
		writeError(w, 429, "too_many_attempts")

	default:
		app.LogError("api.request", err)
		writeError(w, 503, "service_unavailable")
	}
}
