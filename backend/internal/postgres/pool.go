package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, errors.New("postgres: invalid connection configuration")
	}

	config.MaxConns = 10
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, errors.New("postgres: cannot create connection pool")
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()

		return nil, errors.New("postgres: connection check failed")
	}

	return pool, nil
}
