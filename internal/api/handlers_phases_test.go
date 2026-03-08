package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestHandleGetPhases(t *testing.T) {
	t.Run("returns all phases", func(t *testing.T) {
		env := setupTestEnv(t)

		req := httptest.NewRequest(http.MethodGet, "/api/videos/phases", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body: %s", w.Code, http.StatusOK, w.Body.String())
		}

		var phases []PhaseInfo
		json.NewDecoder(w.Body).Decode(&phases)
		if len(phases) != 8 {
			t.Errorf("got %d phases, want 8", len(phases))
		}

		// Verify sorted by ID
		for i := 1; i < len(phases); i++ {
			if phases[i].ID < phases[i-1].ID {
				t.Errorf("phases not sorted: id %d before %d", phases[i-1].ID, phases[i].ID)
			}
		}
	})

	t.Run("counts videos correctly", func(t *testing.T) {
		env := setupTestEnv(t)
		// Seed a video in Ideas phase (phase 7 = no date, not blocked, not delayed)
		seedVideo(t, env, storage.Video{
			Name:     "idea-1",
			Category: "devops",
		})
		seedVideo(t, env, storage.Video{
			Name:     "idea-2",
			Category: "devops",
		})

		req := httptest.NewRequest(http.MethodGet, "/api/videos/phases", nil)
		w := httptest.NewRecorder()
		env.server.Router().ServeHTTP(w, req)

		var phases []PhaseInfo
		json.NewDecoder(w.Body).Decode(&phases)

		// Find ideas phase (7)
		for _, p := range phases {
			if p.ID == 7 {
				if p.Count != 2 {
					t.Errorf("Ideas phase count = %d, want 2", p.Count)
				}
				if p.Name != "Ideas" {
					t.Errorf("Ideas phase name = %q, want %q", p.Name, "Ideas")
				}
				return
			}
		}
		t.Error("Ideas phase not found")
	})
}
