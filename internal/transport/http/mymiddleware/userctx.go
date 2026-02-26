package middleware

import (
	"context"
	"net/http"
	"strconv"

	"my_mdb/internal/service"
	"my_mdb/internal/transport/http/respond"

	"github.com/go-chi/chi/v5"
)

type ctxKey string

const userIDKey ctxKey = "user_id"

func UserIDFromCtx(ctx context.Context) (int, bool) {
	v := ctx.Value(userIDKey)
	id, ok := v.(int)
	return id, ok
}

func UserCtx(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			raw := chi.URLParam(r, "userID")
			if raw == "" {
				next.ServeHTTP(w, r)
				return
			}

			id, err := strconv.Atoi(raw)
			if err != nil {
				respond.Error(w, r, respond.ErrBadRequest("invalid userID"))
				return
			}

			if err := auth.ValidateUserID(r.Context(), id); err != nil {
				respond.FromServiceError(w, r, err)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
