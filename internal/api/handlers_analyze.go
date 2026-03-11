package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/ai"
)

// --- Response types ---

type AnalyzeTitlesResponse struct {
	VideoCount            int                        `json:"videoCount"`
	HighPerformingPatterns []ai.TitlePattern          `json:"highPerformingPatterns"`
	LowPerformingPatterns  []ai.TitlePattern          `json:"lowPerformingPatterns"`
	Recommendations        []ai.TitleRecommendation   `json:"recommendations"`
	TitlesMDContent        string                     `json:"titlesMdContent"`
}

type ApplyTitlesRequest struct {
	Content string `json:"content"`
}

type ApplyTitlesResponse struct {
	Applied     bool   `json:"applied"`
	SyncWarning string `json:"syncWarning,omitempty"`
}

// handleAnalyzeTitles runs the full title analysis pipeline:
// load A/B data → fetch YouTube analytics → enrich → AI analysis.
func (s *Server) handleAnalyzeTitles(w http.ResponseWriter, r *http.Request) {
	if s.analyzeService == nil {
		respondError(w, http.StatusNotImplemented, "analyze service not configured", "")
		return
	}

	dataDir := s.dataDir
	if dataDir == "" {
		dataDir = "."
	}

	// Load videos with A/B data
	indexPath := filepath.Join(dataDir, "index.yaml")
	manuscriptDir := filepath.Join(dataDir, "manuscript")
	videos, err := s.analyzeService.LoadVideosWithABData(indexPath, dataDir, manuscriptDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load A/B data", err.Error())
		return
	}

	if len(videos) == 0 {
		respondJSON(w, http.StatusOK, AnalyzeTitlesResponse{VideoCount: 0})
		return
	}

	// Fetch YouTube analytics
	analytics, err := s.analyzeService.GetVideoAnalyticsForLastYear(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch YouTube analytics", err.Error())
		return
	}

	// Enrich with first-week metrics
	analyticsWithFirstWeek, err := s.analyzeService.EnrichWithFirstWeekMetrics(r.Context(), analytics)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to fetch first-week metrics", err.Error())
		return
	}

	// Join A/B data with analytics
	enrichedVideos := s.analyzeService.EnrichWithAnalytics(videos, analyticsWithFirstWeek)

	// Run AI analysis (baseDir=dataDir so audit trail goes to dataDir/tmp/, which is gitignored)
	result, _, err := s.analyzeService.AnalyzeTitles(r.Context(), enrichedVideos, dataDir)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI analysis failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, AnalyzeTitlesResponse{
		VideoCount:            len(enrichedVideos),
		HighPerformingPatterns: result.HighPerformingPatterns,
		LowPerformingPatterns:  result.LowPerformingPatterns,
		Recommendations:        result.Recommendations,
		TitlesMDContent:        result.TitlesMDContent,
	})
}

// handleApplyTitlesTemplate writes the titles.md content to the data directory
// and commits+pushes if git sync is configured.
func (s *Server) handleApplyTitlesTemplate(w http.ResponseWriter, r *http.Request) {
	var req ApplyTitlesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "content is required", "")
		return
	}

	dataDir := s.dataDir
	if dataDir == "" {
		dataDir = "."
	}

	titlesPath := filepath.Join(dataDir, "titles.md")
	if err := os.WriteFile(titlesPath, []byte(req.Content), 0644); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to write titles.md", err.Error())
		return
	}

	resp := ApplyTitlesResponse{Applied: true}

	if s.gitSync != nil {
		if err := s.gitSync.CommitAndPush("Update titles.md from title analysis"); err != nil {
			resp.SyncWarning = err.Error()
		}
	}

	respondJSON(w, http.StatusOK, resp)
}
