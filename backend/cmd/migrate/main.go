package main

import (
	"context"
	"errors"
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

		log.Printf("applied migrations: %d", len(applied))

		return nil
	}

	status, err := provider.Status(ctx)
	if err != nil {
		return err
	}

	for _, migration := range status {
		log.Printf("%s: %s", migration.Source.Path, migration.State)
	}

	return nil
}
