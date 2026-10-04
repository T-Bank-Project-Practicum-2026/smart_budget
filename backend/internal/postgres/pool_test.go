package postgres_test

import (
	"context"
	"os"
	"smart_budget/internal/postgres"
	"testing"
	"time"
)

func TestConnection(t *testing.T) {
	if os.Getenv("PGHOST") == "" {
		t.Skip("run make test for PostgreSQL integration check")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)

	defer cancel()

	pool, err := postgres.NewPool(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	var value int

	if err := pool.QueryRow(ctx, "SELECT 1").Scan(&value); err != nil {
		t.Fatal(err)
	}

	if value != 1 {
		t.Fatalf("SELECT 1 returned %d", value)
	}
}
