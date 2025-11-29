package db

import (
	"context"
	"rest-api/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}
	return pool, nil
}
