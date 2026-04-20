package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/Tanveer-rajpurohit/p2/internal/utils"
)


type contextKey string
const ClaimsKey contextKey = "claims"

// RequireAuth — any logged in user
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.ResponseWithError(w, 401, "unauthorized")
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateToken(tokenStr, false)
		if err != nil {
			utils.ResponseWithError(w, 401, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(),ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAdmin — only admin users
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims ,ok := r.Context().Value(ClaimsKey).(*Claims)
		if !ok || claims.Role != "admin" {
			utils.ResponseWithError(w, 403, "forbidden")
			return 
		}

		next.ServeHTTP(w, r.WithContext(r.Context()))
	})
}