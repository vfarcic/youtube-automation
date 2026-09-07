package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Publish, action and social routes act on video state, so they must refresh
// the data repo first. Skipping the pull is how Hugo publishing kept failing
// with "Video has no title" against a clone that had drifted behind a fix
// already pushed upstream.
func TestSyncBeforeAction_PullsOnStateChangingRoutes(t *testing.T) {
	routes := []struct {
		name   string
		method string
		path   string
	}{
		{name: "publish hugo", method: http.MethodPost, path: "/api/publish/hugo/test-video?category=Test"},
		{name: "publish youtube", method: http.MethodPost, path: "/api/publish/youtube/test-video?category=Test"},
		{name: "request edit", method: http.MethodPost, path: "/api/actions/request-edit/test-video?category=Test"},
		{name: "notify sponsors", method: http.MethodPost, path: "/api/actions/notify-sponsors/test-video?category=Test"},
		{name: "social post", method: http.MethodPost, path: "/api/social/bluesky/test-video?category=Test"},
	}

	for _, rt := range routes {
		t.Run(rt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			gs := &mockGitSync{}
			env.server.gitSync = gs

			req := httptest.NewRequest(rt.method, rt.path, nil)
			env.server.Router().ServeHTTP(httptest.NewRecorder(), req)

			// The handler's own outcome is irrelevant here — it may well fail on
			// unconfigured services. What matters is that the sync ran first.
			if !gs.pullCalled {
				t.Error("expected PullIfStale to be called before the handler read video state")
			}
			if gs.pullMaxAge != pullOnReadThrottle {
				t.Errorf("PullIfStale maxAge = %v, want %v", gs.pullMaxAge, pullOnReadThrottle)
			}
		})
	}
}

func TestSyncBeforeAction_ProceedsWhenPullFails(t *testing.T) {
	env := setupTestEnv(t)
	gs := &mockGitSync{pullErr: errors.New("network unreachable")}
	env.server.gitSync = gs

	req := httptest.NewRequest(http.MethodPost, "/api/publish/hugo/test-video?category=Test", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	// A transient pull failure must not block publishing outright; the request
	// reaches the handler and gets a normal (non-502) answer.
	if w.Code == http.StatusBadGateway {
		t.Errorf("pull failure should not short-circuit the request, got %d", w.Code)
	}
	if !gs.pullCalled {
		t.Error("expected PullIfStale to be attempted")
	}
}

func TestSyncBeforeAction_NoopWhenGitSyncUnconfigured(t *testing.T) {
	env := setupTestEnv(t)
	env.server.gitSync = nil

	req := httptest.NewRequest(http.MethodPost, "/api/publish/hugo/test-video?category=Test", nil)
	w := httptest.NewRecorder()

	// Must not panic on a nil gitSync (the local-CLI configuration).
	env.server.Router().ServeHTTP(w, req)
}
