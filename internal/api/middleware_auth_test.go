package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBearerTokenAuth(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name       string
		token      string
		authHeader string
		queryToken string
		wantStatus int
	}{
		{
			name:       "valid token",
			token:      "secret",
			authHeader: "Bearer secret",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid token",
			token:      "secret",
			authHeader: "Bearer wrong",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing authorization header",
			token:      "secret",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "malformed header no Bearer prefix",
			token:      "secret",
			authHeader: "Basic secret",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Bearer with no token value",
			token:      "secret",
			authHeader: "Bearer ",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "auth disabled with empty token",
			token:      "",
			authHeader: "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "auth disabled passes any header",
			token:      "",
			authHeader: "Bearer anything",
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid query token without header",
			token:      "secret",
			queryToken: "secret",
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid query token without header",
			token:      "secret",
			queryToken: "wrong",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "header takes precedence over query token",
			token:      "secret",
			authHeader: "Bearer secret",
			queryToken: "wrong",
			wantStatus: http.StatusOK,
		},
		{
			name:       "malformed header falls back to valid query token",
			token:      "secret",
			authHeader: "Basic secret",
			queryToken: "secret",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mw := bearerTokenAuth(tt.token)
			handler := mw(okHandler)

			url := "/api/videos"
			if tt.queryToken != "" {
				url += "?token=" + tt.queryToken
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
		})
	}
}

func TestHealthPublicWhileAPIRequiresAuth(t *testing.T) {
	env := setupTestEnv(t)

	// Override: create a server with auth enabled
	srv := NewServer(env.server.videoService, env.server.videoManager, env.server.aspectService, env.server.filesystem, nil, "test-token", nil)

	// Health endpoint should be accessible without auth
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("/health status = %d, want %d", rr.Code, http.StatusOK)
	}

	// API endpoint should require auth
	req = httptest.NewRequest(http.MethodGet, "/api/videos", nil)
	rr = httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("/api/videos without auth status = %d, want %d", rr.Code, http.StatusUnauthorized)
	}

	// API endpoint with valid auth should pass auth (not 401)
	req = httptest.NewRequest(http.MethodGet, "/api/videos", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rr = httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)
	if rr.Code == http.StatusUnauthorized {
		t.Errorf("/api/videos with valid auth should not return 401")
	}
}

func TestCORSPreflightPassesWithoutAuth(t *testing.T) {
	env := setupTestEnv(t)

	srv := NewServer(env.server.videoService, env.server.videoManager, env.server.aspectService, env.server.filesystem, nil, "test-token", nil)

	req := httptest.NewRequest(http.MethodOptions, "/api/videos", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)

	// OPTIONS should be handled by CORS middleware before auth
	if rr.Code != http.StatusNoContent {
		t.Errorf("OPTIONS preflight status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}
