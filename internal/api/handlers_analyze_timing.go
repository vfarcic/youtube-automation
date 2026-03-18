package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/configuration"
)

// --- Response/Request types ---

type GetTimingResponse struct {
	Recommendations []configuration.TimingRecommendation `json:"recommendations"`
}

type PutTimingRequest struct {
	Recommendations []configuration.TimingRecommendation `json:"recommendations"`
}

type PutTimingResponse struct {
	Saved       bool   `json:"saved"`
	SyncWarning string `json:"syncWarning,omitempty"`
}

type GenerateTimingResponse struct {
	Recommendations []configuration.TimingRecommendation `json:"recommendations"`
	VideoCount      int                                  `json:"videoCount"`
}

// handleGetTimingRecommendations returns the current timing recommendations from settings.yaml.
// GET /api/analyze/timing
func (s *Server) handleGetTimingRecommendations(w http.ResponseWriter, r *http.Request) {
	dataDir := s.dataDir
	if dataDir == "" {
		dataDir = "."
	}
	settingsPath := filepath.Join(dataDir, "settings.yaml")

	recs, err := configuration.LoadTimingRecommendations(settingsPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load timing recommendations", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GetTimingResponse{Recommendations: recs})
}

// handlePutTimingRecommendations saves timing recommendations to settings.yaml.
// PUT /api/analyze/timing
func (s *Server) handlePutTimingRecommendations(w http.ResponseWriter, r *http.Request) {
	var req PutTimingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	dataDir := s.dataDir
	if dataDir == "" {
		dataDir = "."
	}
	settingsPath := filepath.Join(dataDir, "settings.yaml")

	if err := ensureSettingsFile(settingsPath); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to initialize settings", err.Error())
		return
	}

	if err := configuration.SaveTimingRecommendations(settingsPath, req.Recommendations); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save timing recommendations", err.Error())
		return
	}

	resp := PutTimingResponse{Saved: true}

	if s.gitSync != nil {
		if err := s.gitSync.CommitAndPush("Update timing recommendations"); err != nil {
			resp.SyncWarning = err.Error()
		}
	}

	respondJSON(w, http.StatusOK, resp)
}

// handleGenerateTimingRecommendations runs the AI timing analysis pipeline.
// POST /api/analyze/timing/generate
func (s *Server) handleGenerateTimingRecommendations(w http.ResponseWriter, r *http.Request) {
	if s.analyzeService == nil {
		respondError(w, http.StatusNotImplemented, "analyze service not configured", "")
		return
	}

	analytics, err := s.analyzeService.GetVideoAnalyticsForLastYear(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch YouTube analytics", err.Error())
		return
	}

	recs, _, err := s.analyzeService.GenerateTimingRecommendations(r.Context(), analytics)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI timing analysis failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, GenerateTimingResponse{
		Recommendations: recs,
		VideoCount:      len(analytics),
	})
}

// ensureSettingsFile creates settings.yaml with minimal valid YAML if it doesn't exist.
func ensureSettingsFile(settingsPath string) error {
	if _, err := os.Stat(settingsPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(settingsPath, []byte("{}\n"), 0644)
}
