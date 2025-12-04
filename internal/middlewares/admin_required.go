package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"rest-api/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

const IsAdminKey contextKey = "isAdmin"

func UserStatusCheck(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(int64)
			if !ok || userID == 0 {
				utils.WriteJSONError(w, http.StatusUnauthorized, "not_authorized", "you are not authorized")
				return
			}

			var isAdmin bool
			err := pool.QueryRow(
				r.Context(),
				"select is_admin from users where id = $1",
				userID,
			).Scan(&isAdmin)

			if err != nil {
				fmt.Println("database : ", err)
				utils.WriteJSONError(w, http.StatusInternalServerError, "db_error", "failed to fetch user info")
				return
			}

			if !isAdmin {
				utils.WriteJSONError(w, http.StatusForbidden, "forbidden", "you must be an admin to access this resource")
				return
			}

			ctx := context.WithValue(r.Context(), IsAdminKey, isAdmin)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AddUserStatus(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value(UserIDKey).(int64)
			if ok && userID != 0 {
				var isAdmin bool
				err := pool.QueryRow(
					r.Context(),
					"select is_admin from users where id = $1",
					userID,
				).Scan(&isAdmin)

				if err == nil {
					ctx := context.WithValue(r.Context(), IsAdminKey, isAdmin)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
