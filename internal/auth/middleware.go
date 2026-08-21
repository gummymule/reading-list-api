package auth

import (
	"context"
	"net/http"

	"reading-list-api/internal/response"
)

type contextKey string

const userIDKey contextKey = "userID"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "04", "Not authenticated")
			return
		}

		claims, err := ParseToken(cookie.Value)
		if err != nil {
			response.Error(w, http.StatusUnauthorized, "04", "Invalid or expired session")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}
