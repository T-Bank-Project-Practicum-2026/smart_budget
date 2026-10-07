package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"smart_budget/internal/postgres"
	app "smart_budget/internal/runtime"
	"smart_budget/internal/server"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	defer stop()

	dbctx, cancel := context.WithTimeout(ctx, 10*time.Second)

	defer cancel()

	db, err := postgres.NewPool(dbctx, "")
	if err != nil {
		return err
	}
	defer db.Close()

	srv, err := server.New(db)
	if err != nil {
		return err
	}

	return app.Serve(ctx, srv.Handler())
}
