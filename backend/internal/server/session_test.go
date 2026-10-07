package server

import (
	"net/http"
	"net/http/httptest"
	"smart_budget/internal/auth"
	"smart_budget/internal/testutil"
	"strings"
	"testing"
	"time"
)

func TestAccountGuardAndServerLogout(t *testing.T) {
	db := testutil.Database(t)
	_, err := db.Exec(t.Context(), `INSERT INTO app.users(id,email,password_hash) VALUES('a','a@example.test','unused'),('b','b@example.test','unused');
 INSERT INTO app.auth_sessions(id,user_id,refresh_hash,expires_at) VALUES('session-b','b','unused',now()+interval '1 day');`)
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := auth.NewTokens(strings.Repeat("x", 32))
	if err != nil {
		t.Fatal(err)
	}

	service, err := auth.NewService(db, tokens)
	if err != nil {
		t.Fatal(err)
	}

	access, err := tokens.Issue("b", "session-b", time.Now())
	if err != nil {
		t.Fatal(err)
	}

	srv := &Server{DB: db, Auth: service, Token: "internal-test", Origins: []string{"http://localhost:3000"}}
	handler := srv.Handler()

	for _, path := range []string{"/api/v1/notifications/test", "/api/v1/auth/logout"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("X-CSRF-Protection", "1")
		req.Header.Set("X-Account-ID", "a")
		req.AddCookie(&http.Cookie{Name: accessCookie, Value: access})
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != http.StatusConflict {
			t.Fatalf("%s: status %d", path, response.Code)
		}
	}

	if _, err = service.ActiveSession(t.Context(), "session-b", "b"); err != nil {
		t.Fatal("stale tab revoked another account")
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("X-CSRF-Protection", "1")
	req.Header.Set("X-Account-ID", "b")
	req.AddCookie(&http.Cookie{Name: accessCookie, Value: access})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != 204 {
		t.Fatalf("logout: %d", response.Code)
	}

	req = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/internal/sessions/session-b?userId=b", nil)
	req.Header.Set("Authorization", "Bearer internal-test")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != 401 {
		t.Fatalf("revoked session: %d", response.Code)
	}

	req.Header.Set("Authorization", "Bearer wrong-token")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	if response.Code != 403 {
		t.Fatalf("invalid internal credential must not look like revoked session: %d", response.Code)
	}
}
