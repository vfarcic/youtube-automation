package api

import (
	"net/http"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/app"
	"devopstoolkit/youtube-automation/internal/configuration"

	"github.com/go-chi/chi/v5"
)

// ApplyRandomTimingResponse is the JSON response for the apply-random-timing endpoint.
type ApplyRandomTimingResponse struct {
	NewDate      string `json:"newDate"`
	OriginalDate string `json:"originalDate"`
	Day          string `json:"day"`
	Time         string `json:"time"`
	Reasoning    string `json:"reasoning"`
	SyncWarning  string `json:"syncWarning,omitempty"`
}

// handleApplyRandomTiming handles POST /api/videos/{videoName}/apply-random-timing?category=X
func (s *Server) handleApplyRandomTiming(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "Query parameter 'category' is required")
		return
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	if video.Date == "" {
		respondError(w, http.StatusBadRequest, "No date set", "Video must have a date before applying random timing")
		return
	}

	dataDir := s.dataDir
	if dataDir == "" {
		dataDir = "."
	}
	settingsPath := filepath.Join(dataDir, "settings.yaml")

	recommendations, err := configuration.LoadTimingRecommendations(settingsPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to load timing recommendations", err.Error())
		return
	}
	if len(recommendations) == 0 {
		respondError(w, http.StatusBadRequest, "No timing recommendations", "No timing recommendations found in settings.yaml. Run timing analysis first.")
		return
	}

	originalDate := video.Date
	newDate, selectedRec, err := app.ApplyRandomTiming(video.Date, recommendations)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to apply timing", err.Error())
		return
	}

	video.Date = newDate
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := ApplyRandomTimingResponse{
		NewDate:      newDate,
		OriginalDate: originalDate,
		Day:          selectedRec.Day,
		Time:         selectedRec.Time,
		Reasoning:    selectedRec.Reasoning,
	}

	addSyncWarningStr(&resp.SyncWarning, s.videoService)

	respondJSON(w, http.StatusOK, resp)
}
