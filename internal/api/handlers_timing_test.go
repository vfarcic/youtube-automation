package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestHandleApplyRandomTiming_MissingCategory(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/videos/test-video/apply-random-timing", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleApplyRandomTiming_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/videos/nonexistent/apply-random-timing?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleApplyRandomTiming_NoDate(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/videos/test-video/apply-random-timing?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleApplyRandomTiming_NoRecommendations(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Date:     "2025-12-02T16:00",
	})

	// Write an empty settings.yaml (no timing recommendations)
	settingsContent := "timing:\n  recommendations: []\n"
	if err := os.WriteFile(filepath.Join(env.tmpDir, "settings.yaml"), []byte(settingsContent), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/videos/test-video/apply-random-timing?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleApplyRandomTiming_Success(t *testing.T) {
	env := setupTestEnv(t)
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Date:     "2025-12-02T16:00", // Tuesday
	})

	// Write settings.yaml with timing recommendations
	settingsContent := `timing:
  recommendations:
    - day: Wednesday
      time: "14:00"
      reasoning: "Mid-week engagement peak"
    - day: Wednesday
      time: "14:00"
      reasoning: "Mid-week engagement peak"
`
	if err := os.WriteFile(filepath.Join(env.tmpDir, "settings.yaml"), []byte(settingsContent), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/videos/test-video/apply-random-timing?category=devops", nil)
	w := httptest.NewRecorder()
	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ApplyRandomTimingResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.OriginalDate != "2025-12-02T16:00" {
		t.Errorf("expected originalDate '2025-12-02T16:00', got '%s'", resp.OriginalDate)
	}
	if resp.NewDate == "" {
		t.Error("expected newDate to be set")
	}
	// With all recommendations being Wednesday 14:00, the new date should be Wed of that week
	if resp.NewDate != "2025-12-03T14:00" {
		t.Errorf("expected newDate '2025-12-03T14:00', got '%s'", resp.NewDate)
	}
	if resp.Day != "Wednesday" {
		t.Errorf("expected day 'Wednesday', got '%s'", resp.Day)
	}
	if resp.Time != "14:00" {
		t.Errorf("expected time '14:00', got '%s'", resp.Time)
	}
	if resp.Reasoning == "" {
		t.Error("expected reasoning to be set")
	}
	if resp.SyncWarning == "" {
		t.Error("expected syncWarning since git sync is not configured")
	}
}
