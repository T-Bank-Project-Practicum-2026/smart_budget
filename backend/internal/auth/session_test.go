package auth

import (
	"errors"
	"smart_budget/internal/testutil"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRefreshRotationAndLogoutOwner(t *testing.T) {
	db := testutil.Database(t)
	_, err := db.Exec(t.Context(), "INSERT INTO app.users(id,email,password_hash) VALUES('a','a@example.test','unused')")
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec(t.Context(), "INSERT INTO app.auth_sessions(id,user_id,refresh_hash,expires_at) VALUES('session','a',$1,now()+interval '1 day')", Hash("original"))
	if err != nil {
		t.Fatal(err)
	}

	tokens, err := NewTokens(strings.Repeat("x", 32))
	if err != nil {
		t.Fatal(err)
	}

	service, err := NewService(db, tokens)
	if err != nil {
		t.Fatal(err)
	}

	var group sync.WaitGroup

	results := make(chan error, 2)

	for range 2 {
		group.Go(func() { _, err := service.Refresh(t.Context(), "original"); results <- err })
	}

	group.Wait()
	close(results)
	successes, rejected := 0, 0

	for err := range results {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrUnauthorized) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}

	if successes != 1 || rejected != 1 {
		t.Fatalf("rotation successes=%d rejected=%d", successes, rejected)
	}

	expiredAccess, err := tokens.Issue("a", "session", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	if _, err = service.ActiveSession(t.Context(), "session", "a"); err != nil {
		t.Fatal("access expiry must not expire server session")
	}

	if _, err = db.Exec(t.Context(), "UPDATE app.auth_sessions SET refresh_hash=$1", Hash("current")); err != nil {
		t.Fatal(err)
	}

	if err = service.Logout(t.Context(), expiredAccess, "current", "b"); !errors.Is(err, ErrAccountChanged) {
		t.Fatalf("wrong owner: %v", err)
	}

	if err = service.Logout(t.Context(), expiredAccess, "current", "a"); err != nil {
		t.Fatal(err)
	}

	if _, err = service.ActiveSession(t.Context(), "session", "a"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("logout not effective: %v", err)
	}
}
