package db

import (
	"context"
	"rest-api/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitDatabase(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
