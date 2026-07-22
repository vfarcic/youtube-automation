package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/service"
)

func TestHandleGetCategories(t *testing.T) {
	t.Run("returns categories from manuscript dir", func(t *testing.T) {
		env := setupTestEnv(t)

		// Create additional category dirs
		for _, cat := range []string{"cloud-native", "kubernetes"} {
			dir := filepath.Join(env.tmpDir, "manuscript", cat)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
		}

		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var categories []service.Category
		json.NewDecoder(w.Body).Decode(&categories)
		// We created devops in setupTestEnv, plus cloud-native and kubernetes
		if len(categories) != 3 {
			t.Errorf("got %d categories, want 3", len(categories))
		}

		// Verify sorted alphabetically by name
		for i := 1; i < len(categories); i++ {
			if categories[i].Name < categories[i-1].Name {
				t.Errorf("categories not sorted: %q before %q", categories[i-1].Name, categories[i].Name)
			}
		}
	})

	t.Run("triggers throttled pull when git sync configured", func(t *testing.T) {
		env := setupTestEnv(t)

		gs := &mockGitSync{}
		env.server.gitSync = gs

		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		if !gs.pullCalled {
			t.Error("expected PullIfStale to be called so newly-pushed categories appear")
		}
		if gs.pullMaxAge != pullOnReadThrottle {
			t.Errorf("PullIfStale maxAge = %v, want %v", gs.pullMaxAge, pullOnReadThrottle)
		}
	})

	t.Run("still serves categories when pull fails", func(t *testing.T) {
		env := setupTestEnv(t)

		gs := &mockGitSync{pullErr: errors.New("pull failed")}
		env.server.gitSync = gs

		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		// Pull failure must not block the response — local copy is served.
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}
		var categories []service.Category
		if err := json.NewDecoder(w.Body).Decode(&categories); err != nil {
			t.Fatalf("decode: %v", err)
		}
	})

	t.Run("returns empty when no categories", func(t *testing.T) {
		tmpDir := t.TempDir()
		origDir, _ := os.Getwd()
		os.Chdir(tmpDir)
		t.Cleanup(func() { os.Chdir(origDir) })

		// Create manuscript dir but no subdirs
		os.MkdirAll(filepath.Join(tmpDir, "manuscript"), 0755)
		indexPath := filepath.Join(tmpDir, "index.yaml")
		os.WriteFile(indexPath, []byte("[]"), 0644)

		env := setupTestEnv(t)
		req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}
