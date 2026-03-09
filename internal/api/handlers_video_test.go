package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestHandleCreateVideo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTestEnv(t)

		body := `{"name":"my-test-video","category":"devops"}`
		req := httptest.NewRequest(http.MethodPost, "/api/videos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var resp VideoResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if resp.Name != "my-test-video" {
			t.Errorf("name = %q, want %q", resp.Name, "my-test-video")
		}
		if resp.Category != "devops" {
			t.Errorf("category = %q, want %q", resp.Category, "devops")
		}
		if resp.ID == "" {
			t.Error("id should not be empty")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		env := setupTestEnv(t)

		body := `{"category":"devops"}`
		req := httptest.NewRequest(http.MethodPost, "/api/videos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodPost, "/api/videos", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("with date", func(t *testing.T) {
		env := setupTestEnv(t)

		body := `{"name":"dated-video","category":"devops","date":"2025-01-15T10:00"}`
		req := httptest.NewRequest(http.MethodPost, "/api/videos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusCreated, w.Body.String())
		}

		var resp VideoResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if resp.Date != "2025-01-15T10:00" {
			t.Errorf("date = %q, want %q", resp.Date, "2025-01-15T10:00")
		}
	})
}

func TestHandleGetVideo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:     "test-video",
			Category: "devops",
			Date:     "2025-01-01T10:00",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/videos/test-video?category=devops", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var resp VideoResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if resp.Name != "test-video" {
			t.Errorf("name = %q, want %q", resp.Name, "test-video")
		}
	})

	t.Run("missing category", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos/test-video", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos/nonexistent?category=devops", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestHandleGetVideos(t *testing.T) {
	t.Run("returns videos by phase", func(t *testing.T) {
		env := setupTestEnv(t)
		// Phase 7 = Ideas (no date, no blocked, no delayed)
		seedVideo(t, env, storage.Video{
			Name:     "idea-video",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/videos?phase=7", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var resp []VideoResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if len(resp) != 1 {
			t.Fatalf("got %d videos, want 1", len(resp))
		}
		if resp[0].Name != "idea-video" {
			t.Errorf("name = %q, want %q", resp[0].Name, "idea-video")
		}
	})

	t.Run("missing phase", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid phase", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos?phase=abc", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos?phase=0", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var resp []VideoResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if len(resp) != 0 {
			t.Errorf("got %d videos, want 0", len(resp))
		}
	})
}

func TestHandleGetVideosList(t *testing.T) {
	t.Run("returns lightweight list", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:     "list-video",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/videos/list?phase=7", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var items []VideoListItem
		json.NewDecoder(w.Body).Decode(&items)
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].Name != "list-video" {
			t.Errorf("name = %q, want %q", items[0].Name, "list-video")
		}
	})

	t.Run("missing phase", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos/list", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleUpdateVideo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:     "update-me",
			Category: "devops",
		})

		body := `{"date":"2025-06-01T12:00"}`
		req := httptest.NewRequest(http.MethodPut, "/api/videos/update-me?category=devops", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var resp VideoResponse
		json.NewDecoder(w.Body).Decode(&resp)
		if resp.Date != "2025-06-01T12:00" {
			t.Errorf("date = %q, want %q", resp.Date, "2025-06-01T12:00")
		}
	})

	t.Run("missing category", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodPut, "/api/videos/test", bytes.NewBufferString(`{}`))
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodPut, "/api/videos/nonexistent?category=devops", bytes.NewBufferString(`{}`))
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:     "bad-json",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodPut, "/api/videos/bad-json?category=devops", bytes.NewBufferString("{invalid"))
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleDeleteVideo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:     "delete-me",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodDelete, "/api/videos/delete-me?category=devops", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusNoContent, w.Body.String())
		}

		// Verify it's gone
		req2 := httptest.NewRequest(http.MethodGet, "/api/videos/delete-me?category=devops", nil)
		w2 := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w2, req2)
		if w2.Code != http.StatusNotFound {
			t.Errorf("after delete: status = %d, want %d", w2.Code, http.StatusNotFound)
		}
	})

	t.Run("missing category", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/videos/test", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleSearchVideos(t *testing.T) {
	t.Run("returns matching videos", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:        "kubernetes-deploy",
			Category:    "devops",
			Description: "Deploy apps with Kubernetes",
		})
		seedVideo(t, env, storage.Video{
			Name:     "docker-basics",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/videos/search?q=kubernetes", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var items []VideoListItem
		json.NewDecoder(w.Body).Decode(&items)
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].Name != "kubernetes-deploy" {
			t.Errorf("name = %q, want %q", items[0].Name, "kubernetes-deploy")
		}
	})

	t.Run("empty query returns empty list", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos/search?q=", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var items []VideoListItem
		json.NewDecoder(w.Body).Decode(&items)
		if len(items) != 0 {
			t.Errorf("got %d items, want 0", len(items))
		}
	})

	t.Run("no match returns empty list", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:     "some-video",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/videos/search?q=nonexistent", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var items []VideoListItem
		json.NewDecoder(w.Body).Decode(&items)
		if len(items) != 0 {
			t.Errorf("got %d items, want 0", len(items))
		}
	})

	t.Run("case insensitive search", func(t *testing.T) {
		env := setupTestEnv(t)
		seedVideo(t, env, storage.Video{
			Name:     "terraform-guide",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/videos/search?q=TERRAFORM", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
		}

		var items []VideoListItem
		json.NewDecoder(w.Body).Decode(&items)
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
	})
}

func TestEnrichVideo(t *testing.T) {
	env := setupTestEnv(t)

	v := storage.Video{
		Name:     "enrich-test",
		Category: "devops",
		Date:     "2025-01-01T10:00",
	}

	enriched := env.server.enrichVideo(v)

	if enriched.ID != "devops/enrich-test" {
		t.Errorf("id = %q, want %q", enriched.ID, "devops/enrich-test")
	}
	// Video with only a date should be in "Started" phase (4)
	if enriched.Phase != 4 {
		t.Errorf("phase = %d, want 4 (Started)", enriched.Phase)
	}
	if enriched.Init.Total == 0 {
		t.Error("init total should be > 0")
	}
}
