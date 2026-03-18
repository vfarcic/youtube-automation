package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/publishing"
)

// --- GET /api/analyze/timing ---

func TestHandleGetTimingRecommendations(t *testing.T) {
	tests := []struct {
		name       string
		settings   string
		wantStatus int
		wantLen    int
	}{
		{
			name: "success with recommendations",
			settings: `timing:
  recommendations:
    - day: Wednesday
      time: "14:00"
      reasoning: "Mid-week peak"
    - day: Monday
      time: "09:00"
      reasoning: "Week start"
`,
			wantStatus: http.StatusOK,
			wantLen:    2,
		},
		{
			name: "empty recommendations array",
			settings: `timing:
  recommendations: []
`,
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
		{
			name:       "no settings file returns empty array",
			settings:   "", // won't write file
			wantStatus: http.StatusOK,
			wantLen:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			dataDir := t.TempDir()
			env.server.SetDataDir(dataDir)

			if tt.settings != "" {
				if err := os.WriteFile(filepath.Join(dataDir, "settings.yaml"), []byte(tt.settings), 0644); err != nil {
					t.Fatal(err)
				}
			}

			req := httptest.NewRequest(http.MethodGet, "/api/analyze/timing", nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			var resp GetTimingResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if len(resp.Recommendations) != tt.wantLen {
				t.Errorf("recommendations length = %d, want %d", len(resp.Recommendations), tt.wantLen)
			}
		})
	}
}

// --- PUT /api/analyze/timing ---

func TestHandlePutTimingRecommendations(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		gitSync    *mockGitSync
		noSettings bool
		wantStatus int
		wantSaved  bool
		wantSync   string
	}{
		{
			name:       "success with git sync",
			body:       `{"recommendations":[{"day":"Wednesday","time":"14:00","reasoning":"Mid-week peak"}]}`,
			gitSync:    &mockGitSync{},
			wantStatus: http.StatusOK,
			wantSaved:  true,
		},
		{
			name:       "success without git sync",
			body:       `{"recommendations":[{"day":"Monday","time":"09:00","reasoning":"Week start"}]}`,
			gitSync:    nil,
			wantStatus: http.StatusOK,
			wantSaved:  true,
		},
		{
			name:       "git sync failure returns warning",
			body:       `{"recommendations":[{"day":"Friday","time":"16:00","reasoning":"Weekend prep"}]}`,
			gitSync:    &mockGitSync{err: fmt.Errorf("push rejected")},
			wantStatus: http.StatusOK,
			wantSaved:  true,
			wantSync:   "push rejected",
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "settings file missing bootstraps and succeeds",
			body:       `{"recommendations":[{"day":"Monday","time":"09:00","reasoning":"test"}]}`,
			noSettings: true,
			wantStatus: http.StatusOK,
			wantSaved:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			dataDir := t.TempDir()
			env.server.SetDataDir(dataDir)

			if !tt.noSettings {
				// Write an initial settings.yaml so SaveTimingRecommendations can read/update it
				if err := os.WriteFile(filepath.Join(dataDir, "settings.yaml"), []byte("timing:\n  recommendations: []\n"), 0644); err != nil {
					t.Fatal(err)
				}
			}

			if tt.gitSync != nil {
				env.server.gitSync = tt.gitSync
			}

			req := httptest.NewRequest(http.MethodPut, "/api/analyze/timing", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			if tt.wantSaved {
				var resp PutTimingResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if !resp.Saved {
					t.Error("expected saved = true")
				}
				if tt.wantSync != "" && !strings.Contains(resp.SyncWarning, tt.wantSync) {
					t.Errorf("syncWarning = %q, want it to contain %q", resp.SyncWarning, tt.wantSync)
				}
			}
		})
	}
}

// --- POST /api/analyze/timing/generate ---

func TestHandleGenerateTimingRecommendations_NotConfigured(t *testing.T) {
	env := setupTestEnv(t)
	// analyzeService is nil by default

	req := httptest.NewRequest(http.MethodPost, "/api/analyze/timing/generate", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotImplemented)
	}
}

func TestHandleGenerateTimingRecommendations(t *testing.T) {
	sampleRecs := []configuration.TimingRecommendation{
		{Day: "Wednesday", Time: "14:00", Reasoning: "Mid-week peak"},
		{Day: "Monday", Time: "09:00", Reasoning: "Week start"},
		{Day: "Friday", Time: "16:00", Reasoning: "Weekend prep"},
		{Day: "Tuesday", Time: "10:00", Reasoning: "Early week"},
		{Day: "Thursday", Time: "15:00", Reasoning: "Late week"},
		{Day: "Saturday", Time: "11:00", Reasoning: "Weekend morning"},
	}

	tests := []struct {
		name         string
		mock         *mockAnalyzeService
		wantStatus   int
		wantRecCount int
		wantVideo    int
	}{
		{
			name: "success",
			mock: &mockAnalyzeService{
				analytics:  []publishing.VideoAnalytics{{VideoID: "vid1"}, {VideoID: "vid2"}},
				timingRecs: sampleRecs,
			},
			wantStatus:   http.StatusOK,
			wantRecCount: 6,
			wantVideo:    2,
		},
		{
			name: "analytics fetch error",
			mock: &mockAnalyzeService{
				analyticsErr: fmt.Errorf("YouTube API failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "AI generation error",
			mock: &mockAnalyzeService{
				analytics: []publishing.VideoAnalytics{{VideoID: "vid1"}},
				timingErr: fmt.Errorf("AI provider failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAnalyzeTestEnv(t, tt.mock)
			dataDir := t.TempDir()
			env.server.SetDataDir(dataDir)

			// Write settings.yaml so auto-save works
			if err := os.WriteFile(filepath.Join(dataDir, "settings.yaml"), []byte("timing:\n  recommendations: []\n"), 0644); err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/analyze/timing/generate", nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			if tt.wantStatus == http.StatusOK {
				var resp GenerateTimingResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(resp.Recommendations) != tt.wantRecCount {
					t.Errorf("recommendations length = %d, want %d", len(resp.Recommendations), tt.wantRecCount)
				}
				if resp.VideoCount != tt.wantVideo {
					t.Errorf("videoCount = %d, want %d", resp.VideoCount, tt.wantVideo)
				}
			}
		})
	}
}
