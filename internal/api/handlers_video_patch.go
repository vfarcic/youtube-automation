package api

import (
	"encoding/json"
	"net/http"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/go-chi/chi/v5"
)

// validAspectKeys defines the set of recognised aspect keys for PATCH validation.
var validAspectKeys = map[string]bool{
	aspect.AspectKeyInitialDetails: true,
	aspect.AspectKeyWorkProgress:   true,
	aspect.AspectKeyDefinition:     true,
	aspect.AspectKeyPostProduction: true,
	aspect.AspectKeyPublishing:     true,
	aspect.AspectKeyPostPublish:    true,
	aspect.AspectKeyAnalysis:       true,
}

// --- Response types ---------------------------------------------------------

// AspectProgressInfo holds progress counts for a single aspect.
type AspectProgressInfo struct {
	AspectKey string `json:"aspectKey"`
	Title     string `json:"title"`
	Completed int    `json:"completed"`
	Total     int    `json:"total"`
}

// OverallProgressResponse combines overall progress with per-aspect breakdown.
type OverallProgressResponse struct {
	Overall ProgressInfo         `json:"overall"`
	Aspects []AspectProgressInfo `json:"aspects"`
}

// ManuscriptResponse wraps the manuscript content.
type ManuscriptResponse struct {
	Content string `json:"content"`
}

// AnimationsResponse wraps animation cues and section headers.
type AnimationsResponse struct {
	Animations []string `json:"animations"`
	Sections   []string `json:"sections"`
}

// --- Handlers ---------------------------------------------------------------

// handlePatchVideoAspect applies a partial update to a single aspect of a video.
//
// Query params: category (required), aspect (required).
// Body: JSON object with field names (using JSON tags) as keys.
func (s *Server) handlePatchVideoAspect(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}
	aspectKey := r.URL.Query().Get("aspect")
	if aspectKey == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: aspect", "")
		return
	}
	if !validAspectKeys[aspectKey] {
		respondError(w, http.StatusBadRequest, "invalid aspect key", aspectKey)
		return
	}

	// Load current video
	v, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "video not found", err.Error())
		return
	}

	// Decode body
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	// Build set of valid field names for this aspect
	aspectFields, err := s.aspectService.GetAspectFields(aspectKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get aspect fields", err.Error())
		return
	}
	validFields := make(map[string]bool, len(aspectFields.Fields))
	for _, f := range aspectFields.Fields {
		validFields[f.FieldName] = true
	}

	// Apply each field
	for fieldName, value := range body {
		if !validFields[fieldName] {
			respondError(w, http.StatusBadRequest, "field not valid for aspect", fieldName)
			return
		}
		if err := aspect.SetFieldValueByJSONPath(&v, fieldName, value); err != nil {
			respondError(w, http.StatusBadRequest, "failed to set field", err.Error())
			return
		}
	}

	// Persist
	if err := s.videoService.UpdateVideo(v); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update video", err.Error())
		return
	}

	// Read back and return enriched response
	updated, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "video updated but failed to read back", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s.enrichVideo(updated))
}

// handleGetVideoProgress returns overall and per-aspect progress for a video.
func (s *Server) handleGetVideoProgress(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}

	v, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "video not found", err.Error())
		return
	}

	overallC, overallT := s.videoManager.CalculateOverallProgress(v)

	aspects := []AspectProgressInfo{
		s.calculateAspectProgress(aspect.AspectKeyInitialDetails, v),
		s.calculateAspectProgress(aspect.AspectKeyWorkProgress, v),
		s.calculateAspectProgress(aspect.AspectKeyDefinition, v),
		s.calculateAspectProgress(aspect.AspectKeyPostProduction, v),
		s.calculateAspectProgress(aspect.AspectKeyPublishing, v),
		s.calculateAspectProgress(aspect.AspectKeyPostPublish, v),
		s.calculateAspectProgress(aspect.AspectKeyAnalysis, v),
	}

	respondJSON(w, http.StatusOK, OverallProgressResponse{
		Overall: ProgressInfo{Completed: overallC, Total: overallT},
		Aspects: aspects,
	})
}

// handleGetVideoAspectProgress returns progress for a single aspect of a video.
func (s *Server) handleGetVideoAspectProgress(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	aspectKey := chi.URLParam(r, "aspect")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}
	if !validAspectKeys[aspectKey] {
		respondError(w, http.StatusBadRequest, "invalid aspect key", aspectKey)
		return
	}

	v, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "video not found", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, s.calculateAspectProgress(aspectKey, v))
}

// handleGetVideoManuscript returns the manuscript content for a video.
func (s *Server) handleGetVideoManuscript(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}

	content, err := s.videoService.GetVideoManuscript(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "manuscript not found", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, ManuscriptResponse{Content: content})
}

// handleGetVideoAnimations returns animation cues and section headers for a video.
func (s *Server) handleGetVideoAnimations(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}

	v, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "video not found", err.Error())
		return
	}

	if v.Gist == "" {
		respondError(w, http.StatusNotFound, "no gist path set for video", "")
		return
	}

	animations, sections, err := s.filesystem.GetAnimations(s.filesystem.ResolvePath(v.Gist))
	if err != nil {
		respondError(w, http.StatusNotFound, "failed to read animations", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, AnimationsResponse{
		Animations: animations,
		Sections:   sections,
	})
}

// --- Helpers ----------------------------------------------------------------

// aspectTitles maps aspect keys to their display titles for progress responses.
var aspectTitles = map[string]string{
	aspect.AspectKeyInitialDetails: "Initial Details",
	aspect.AspectKeyWorkProgress:   "Work Progress",
	aspect.AspectKeyDefinition:     "Definition",
	aspect.AspectKeyPostProduction: "Post Production",
	aspect.AspectKeyPublishing:     "Publishing",
	aspect.AspectKeyPostPublish:    "Post Publish",
	aspect.AspectKeyAnalysis:       "Analysis",
}

// calculateAspectProgress delegates to the correct Manager method for the given aspect.
func (s *Server) calculateAspectProgress(aspectKey string, v storage.Video) AspectProgressInfo {
	var completed, total int
	switch aspectKey {
	case aspect.AspectKeyInitialDetails:
		completed, total = s.videoManager.CalculateInitialDetailsProgress(v)
	case aspect.AspectKeyWorkProgress:
		completed, total = s.videoManager.CalculateWorkProgressProgress(v)
	case aspect.AspectKeyDefinition:
		completed, total = s.videoManager.CalculateDefinePhaseCompletion(v)
	case aspect.AspectKeyPostProduction:
		completed, total = s.videoManager.CalculatePostProductionProgress(v)
	case aspect.AspectKeyPublishing:
		completed, total = s.videoManager.CalculatePublishingProgress(v)
	case aspect.AspectKeyPostPublish:
		completed, total = s.videoManager.CalculatePostPublishProgress(v)
	case aspect.AspectKeyAnalysis:
		completed, total = s.videoManager.CalculateAnalysisProgress(v)
	default:
		completed, total = video.CalculateVideoPhase(v), 0 // fallback (shouldn't happen)
	}

	return AspectProgressInfo{
		AspectKey: aspectKey,
		Title:     aspectTitles[aspectKey],
		Completed: completed,
		Total:     total,
	}
}
