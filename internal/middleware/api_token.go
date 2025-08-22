package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/angel/go-api-sqlite/internal/handlers"
)

// APITokenMiddleware authenticates requests using API tokens
func APITokenMiddleware(h *handlers.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip token auth for token management endpoints
			if strings.HasPrefix(r.URL.Path, "/api/tokens") {
				next.ServeHTTP(w, r)
				return
			}

			// Check for API token in header
			token := r.Header.Get("X-API-Token")
			if token == "" {
				// No token found, let the regular auth middleware handle it
				next.ServeHTTP(w, r)
				return
			}

			// Validate API token
			if !h.AuthenticateToken(token) {
				http.Error(w, "Invalid API token", http.StatusUnauthorized)
				return
			}

			// Token is valid, mark the request as authenticated via API token
			ctx := context.WithValue(r.Context(), "auth_type", "api_token")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
