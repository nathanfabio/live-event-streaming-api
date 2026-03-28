package middleware

import (
	"context"
	"net/http"
	"strings"
)


func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Missing Authorization", http.StatusUnauthorized)
				return
			}
			//Mock validation - In prod, verify JWT signature here
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 {
				http.Error(w, "Invalid format", http.StatusUnauthorized)
				return 
			}

			userID := parts[1] //In prod, extract from JWT claims

			ctx := context.WithValue(r.Context(), "usedID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}