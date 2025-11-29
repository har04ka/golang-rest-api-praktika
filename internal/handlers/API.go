package handlers

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

type API struct {
	Pool *pgxpool.Pool
}

func NewAPI(pool *pgxpool.Pool) *API {
	return &API{
		Pool: pool,
	}
}
