package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestSPAHandler(t *testing.T) {
	fakeFS := fstest.MapFS{
		"index.html":       {Data: []byte("<html>SPA</html>")},
		"assets/main.js":   {Data: []byte("console.log('app')")},
		"assets/style.css": {Data: []byte("body{}")},
	}

	env := setupTestEnv(t)
	srv := NewServer(env.server.videoService, env.server.videoManager, env.server.aspectService, env.server.filesystem, "", fakeFS)

	tests := []struct {
		name            string
		path            string
		wantStatus      int
		wantContentType string
		wantBody        string
	}{
		{
			name:       "serves index.html at root",
			path:       "/",
			wantStatus: http.StatusOK,
			wantBody:   "<html>SPA</html>",
		},
		{
			name:       "serves static JS file",
			path:       "/assets/main.js",
			wantStatus: http.StatusOK,
			wantBody:   "console.log('app')",
		},
		{
			name:       "serves static CSS file",
			path:       "/assets/style.css",
			wantStatus: http.StatusOK,
			wantBody:   "body{}",
		},
		{
			name:       "client-side route falls back to index.html",
			path:       "/phases/1",
			wantStatus: http.StatusOK,
			wantBody:   "<html>SPA</html>",
		},
		{
			name:       "deep client-side route falls back to index.html",
			path:       "/videos/devops/my-video",
			wantStatus: http.StatusOK,
			wantBody:   "<html>SPA</html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()
			srv.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rr.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rr.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestSPAHandlerAPIRoutesStillWork(t *testing.T) {
	fakeFS := fstest.MapFS{
		"index.html": {Data: []byte("<html>SPA</html>")},
	}

	env := setupTestEnv(t)
	srv := NewServer(env.server.videoService, env.server.videoManager, env.server.aspectService, env.server.filesystem, "", fakeFS)

	// API routes should still return JSON, not index.html
	req := httptest.NewRequest(http.MethodGet, "/api/videos/phases", nil)
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("/api/videos/phases status = %d, want %d", rr.Code, http.StatusOK)
	}

	ct := rr.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestSPAHandlerHealthStillWorks(t *testing.T) {
	fakeFS := fstest.MapFS{
		"index.html": {Data: []byte("<html>SPA</html>")},
	}

	env := setupTestEnv(t)
	srv := NewServer(env.server.videoService, env.server.videoManager, env.server.aspectService, env.server.filesystem, "", fakeFS)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("/health status = %d, want %d", rr.Code, http.StatusOK)
	}
}

func TestNoFrontendFS(t *testing.T) {
	env := setupTestEnv(t)
	// nil frontendFS should not register SPA handler
	srv := NewServer(env.server.videoService, env.server.videoManager, env.server.aspectService, env.server.filesystem, "", nil)

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rr := httptest.NewRecorder()
	srv.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
}
