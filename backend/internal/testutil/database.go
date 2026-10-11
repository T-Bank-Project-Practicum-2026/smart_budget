// Package testutil provides isolated databases for integration tests.
package testutil

import (
	"context"
	"crypto/rand"
	"io/fs"
	"os"
	"smart_budget/migrations"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Database(t *testing.T) *pgxpool.Pool {
	t.Helper()
	if os.Getenv("PGHOST") == "" {
		t.Skip("run make test for PostgreSQL integration tests")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)

	defer cancel()

	config, err := pgxpool.ParseConfig("")
	if err != nil {
		t.Fatal(err)
	}

	admin, err := pgxpool.NewWithConfig(ctx, config.Copy())
	if err != nil {
		t.Fatal(err)
	}

	name := "budget_test_" + strings.ToLower(rand.Text())
	if _, err = admin.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{name}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}

	var pool *pgxpool.Pool

	t.Cleanup(func() {
		if pool != nil {
			pool.Close()
		}

		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)

		defer stop()

		if _, err := admin.Exec(cleanup, "DROP DATABASE "+pgx.Identifier{name}.Sanitize()+" WITH (FORCE)"); err != nil {
			t.Error(err)
		}

		admin.Close()
	})
	config.ConnConfig.Database = name
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}

	db := stdlib.OpenDBFromPool(pool)

	defer func() { _ = db.Close() }()

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.Files)
	if err != nil {
		t.Fatal(err)
	}

	if _, err = provider.Up(ctx); err != nil {
		t.Fatal(err)
	}

	files, err := fs.Sub(migrations.Files, "notifications")
	if err != nil {
		t.Fatal(err)
	}

	provider, err = goose.NewProvider(goose.DialectPostgres, db, files, goose.WithTableName("notifications.goose_db_version"))
	if err != nil {
		t.Fatal(err)
	}

	if _, err = provider.Up(ctx); err != nil {
		t.Fatal(err)
	}

	return pool
}
