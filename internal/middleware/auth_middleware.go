package middleware

import (
	"context"
	"net/http"
)

type ContextKey string

const (
	UserIDKey ContextKey = "userID"
	RoleKey   ContextKey = "role"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 1. Extract from Kong headers
		userID := r.Header.Get("X-User-ID")
		role := r.Header.Get("X-User-Role")

		// 2. Validate presence (critical)
		if userID == "" {
			http.Error(w, "missing user identity", http.StatusUnauthorized)
			return
		}

		if role == "" {
			http.Error(w, "missing user role", http.StatusUnauthorized)
			return
		}

		// 3. Inject into context
		ctx := r.Context()
		ctx = context.WithValue(ctx, UserIDKey, userID)
		ctx = context.WithValue(ctx, RoleKey, role)

		// 4. Continue request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) string {
	val, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		return ""
	}
	return val
}

func GetRole(r *http.Request) string {
	val, ok := r.Context().Value(RoleKey).(string)
	if !ok {
		return ""
	}
	return val
}

func RequireRole(r *http.Request, allowed string) bool {
	return GetRole(r) == allowed
}
