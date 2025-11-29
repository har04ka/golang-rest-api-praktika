package middlewares

import (
	"context"
	"net/http"
	"rest-api/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthCheck(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			var user uint64
			token := cookie.Value

			err = pool.QueryRow(
				r.Context(),
				"select user_id from sessions where token_hash = $1",
				utils.HashTokenHMAC(token),
			).Scan(&user)

			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
