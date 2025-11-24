package handlers

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type API struct {
	Ctx  context.Context
	Pool *pgxpool.Pool
}

func NewAPI(ctx context.Context, pool *pgxpool.Pool) *API {
	return &API{
		Ctx:  ctx,
		Pool: pool,
	}
}
