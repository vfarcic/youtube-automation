package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServerRegistersRoutes(t *testing.T) {
	env := setupTestEnv(t)

	routes := []struct {
		method string
		path   string
		want   int // expected status (not 404/405 means route exists)
	}{
		{http.MethodGet, "/health", http.StatusOK},
		{http.MethodGet, "/api/videos/phases", http.StatusOK},
		{http.MethodGet, "/api/categories", http.StatusOK},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			req := httptest.NewRequest(rt.method, rt.path, nil)
			w := httptest.NewRecorder()
			env.server.Router().ServeHTTP(w, req)

			if w.Code == http.StatusNotFound || w.Code == http.StatusMethodNotAllowed {
				t.Errorf("route %s %s returned %d, expected route to be registered", rt.method, rt.path, w.Code)
			}
		})
	}
}

func TestNotFoundRoute(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
