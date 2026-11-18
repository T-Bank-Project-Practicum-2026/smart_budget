package server

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestConfiguredOrigins(t *testing.T) {
	t.Setenv("NOTIFICATIONS_URL", "http://notifications:8080")
	t.Setenv("NOTIFICATIONS_TOKEN", "test-token")
	t.Setenv("VAPID_PUBLIC_KEY", "test-key")
	t.Setenv("JWT_SECRET", strings.Repeat("x", 32))
	t.Setenv("AUTH_COOKIE_SECURE", "false")
	t.Setenv("APP_ORIGINS", "http://localhost:3000, http://192.168.1.10:3000 ")

	srv, err := New(nil)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{"http://localhost:3000", "http://192.168.1.10:3000"}; !reflect.DeepEqual(srv.Origins, want) {
		t.Fatalf("origins = %q, want %q", srv.Origins, want)
	}

	handler := srv.Handler()

	for _, tc := range []struct {
		name, origin, csrf string
		status             int
	}{
		{"localhost", "http://localhost:3000", "1", http.StatusMethodNotAllowed},
		{"LAN", "http://192.168.1.10:3000", "1", http.StatusMethodNotAllowed},
		{"wrong port", "http://192.168.1.10:3001", "1", http.StatusForbidden},
		{"wrong scheme", "https://192.168.1.10:3000", "1", http.StatusForbidden},
		{"untrusted", "http://evil.test", "1", http.StatusForbidden},
		{"missing origin", "", "1", http.StatusForbidden},
		{"null origin", "null", "1", http.StatusForbidden},
		{"missing CSRF", "http://localhost:3000", "", http.StatusForbidden},
		{"wrong CSRF", "http://localhost:3000", "0", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// POST to a GET-only route proves the CSRF gate passed without changing data.
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/health", nil)
			req.Header.Set("Origin", tc.origin)
			req.Header.Set("X-CSRF-Protection", tc.csrf)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d: %s", response.Code, tc.status, response.Body.String())
			}
		})
	}

	t.Setenv("APP_ORIGINS", "http://localhost:3000, https://budget.test")
	if _, err := New(nil); err == nil || err.Error() != "HTTPS origins require AUTH_COOKIE_SECURE=true" {
		t.Fatalf("insecure HTTPS config: %v", err)
	}

	t.Setenv("AUTH_COOKIE_SECURE", "true")
	if _, err := New(nil); err != nil {
		t.Fatalf("secure HTTPS config: %v", err)
	}

	t.Setenv("APP_ORIGINS", "http://localhost:3000, ")
	if _, err := New(nil); err == nil {
		t.Fatal("empty origin must be rejected")
	}
}
