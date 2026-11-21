package server

import (
	"net/http"
	"net/http/httptest"
	"smart_budget/internal/auth"
	servicesv1 "smart_budget/internal/gen/smartbudget/services/v1"
	"smart_budget/internal/gen/smartbudget/services/v1/servicesv1connect"
	"smart_budget/internal/testutil"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
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

		body := response.Body.String()
		if !strings.Contains(body, `"success":false`) || !strings.Contains(body, `"error":"account_changed"`) {
			t.Fatalf("%s: unexpected error body %s", path, body)
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
	if response.Code != http.StatusOK {
		t.Fatalf("logout: %d", response.Code)
	}

	if body := response.Body.String(); !strings.Contains(body, `"success":true`) || !strings.Contains(body, `"data":null`) {
		t.Fatalf("logout: unexpected success body %s", body)
	}

	rpcServer := httptest.NewServer(handler)

	defer rpcServer.Close()

	client := servicesv1connect.NewAuthServiceClient(rpcServer.Client(), rpcServer.URL)
	check := connect.NewRequest(&servicesv1.CheckSessionRequest{SessionId: "session-b", UserId: "b"})
	check.Header().Set("Authorization", "Bearer internal-test")
	result, err := client.CheckSession(t.Context(), check)
	if err != nil || result.Msg.Active {
		t.Fatalf("revoked session: %v %v", result, err)
	}

	check.Header().Set("Authorization", "Bearer wrong-token")
	_, err = client.CheckSession(t.Context(), check)
	if connect.CodeOf(err) != connect.CodeUnauthenticated {
		t.Fatalf("invalid internal credential must not look like revoked session: %v", err)
	}
}
