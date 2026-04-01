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

			var provided string
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				provided = strings.TrimPrefix(authHeader, "Bearer ")
			} else if qToken := r.URL.Query().Get("token"); qToken != "" {
				provided = qToken
			} else {
				respondError(w, http.StatusUnauthorized, "Unauthorized", "missing Authorization header")
				return
			}
			if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
				respondError(w, http.StatusUnauthorized, "Unauthorized", "invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
