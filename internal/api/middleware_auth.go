package api

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// bearerTokenAuth returns middleware that validates Bearer token authentication.
// If token is empty, authentication is disabled and all requests pass through.
func bearerTokenAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "Unauthorized", "missing Authorization header")
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				respondError(w, http.StatusUnauthorized, "Unauthorized", "invalid Authorization header format")
				return
			}

			provided := strings.TrimPrefix(authHeader, "Bearer ")
			if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
				respondError(w, http.StatusUnauthorized, "Unauthorized", "invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
