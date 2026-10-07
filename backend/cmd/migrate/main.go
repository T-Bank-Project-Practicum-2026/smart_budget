package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/signal"
	"smart_budget/internal/postgres"
	"smart_budget/migrations"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) != 2 || (os.Args[1] != "up" && os.Args[1] != "status") {
		return errors.New("usage: migrate up|status")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	ctx, cancel := context.WithTimeout(ctx, time.Minute)

	defer cancel()

	pool, err := postgres.NewPool(ctx, "")
	if err != nil {
		return err
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)

	defer func() { _ = db.Close() }()

	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.Files, goose.WithSessionLocker(locker))
	if err != nil {
		return err
	}

	if os.Args[1] == "up" {
		applied, err := provider.Up(ctx)
		if err != nil {
			return err
		}

		log.Printf("applied app migrations: %d", len(applied))
	}

	notificationFiles, err := fs.Sub(migrations.Files, "notifications")
	if err != nil {
		return err
	}

	notificationProvider, err := goose.NewProvider(goose.DialectPostgres, db, notificationFiles, goose.WithTableName("notifications.goose_db_version"), goose.WithSessionLocker(locker))
	if err != nil {
		return err
	}

	if os.Args[1] == "up" {
		applied, err := notificationProvider.Up(ctx)
		if err != nil {
			return err
		}

		log.Printf("applied notifications migrations: %d", len(applied))

		for role, variable := range map[string]string{"budget_api": "API_DB_PASSWORD", "budget_worker": "WORKER_DB_PASSWORD", "budget_notifications": "NOTIFICATIONS_DB_PASSWORD"} {
			password := os.Getenv(variable)
			if password == "" {
				return fmt.Errorf("%s is required", variable)
			}

			var statement string

			if err := pool.QueryRow(ctx, "SELECT format('ALTER ROLE %I LOGIN PASSWORD %L', $1::text, $2::text)", role, password).Scan(&statement); err != nil {
				return errors.New("cannot configure database role")
			}

			if _, err := pool.Exec(ctx, statement); err != nil {
				return errors.New("cannot configure database role")
			}
		}

		return nil
	}

	for _, p := range []*goose.Provider{provider, notificationProvider} {
		status, err := p.Status(ctx)
		if err != nil {
			return err
		}

		for _, migration := range status {
			log.Printf("%s: %s", migration.Source.Path, migration.State)
		}
	}

	return nil
}
