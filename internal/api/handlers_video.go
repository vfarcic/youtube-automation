package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/go-chi/chi/v5"
)

// ProgressInfo holds completed/total counts for a single phase.
type ProgressInfo struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

// VideoResponse wraps a storage.Video with computed fields.
type VideoResponse struct {
	storage.Video
	ID          string       `json:"id"`
	Phase       int          `json:"phase"`
	Init        ProgressInfo `json:"init"`
	Work        ProgressInfo `json:"work"`
	Define      ProgressInfo `json:"define"`
	Edit        ProgressInfo `json:"edit"`
	Publish     ProgressInfo `json:"publish"`
	PostPublish ProgressInfo `json:"postPublish"`
}

// VideoListItem is a lightweight representation of a video.
type VideoListItem struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Date     string `json:"date,omitempty"`
	Title    string `json:"title,omitempty"`
	Phase    int    `json:"phase"`
	Progress ProgressInfo `json:"progress"`
}

// createVideoRequest is the body for POST /api/videos.
type createVideoRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Date     string `json:"date,omitempty"`
}

// enrichVideo adds computed phase and progress to a video.
func (s *Server) enrichVideo(v storage.Video) VideoResponse {
	phase := video.CalculateVideoPhase(v)

	initC, initT := s.videoManager.CalculateInitialDetailsProgress(v)
	workC, workT := s.videoManager.CalculateWorkProgressProgress(v)
	defC, defT := s.videoManager.CalculateDefinePhaseCompletion(v)
	editC, editT := s.videoManager.CalculatePostProductionProgress(v)
	pubC, pubT := s.videoManager.CalculatePublishingProgress(v)
	postC, postT := s.videoManager.CalculatePostPublishProgress(v)

	return VideoResponse{
		Video: v,
		ID:    v.Category + "/" + v.Name,
		Phase: phase,
		Init:        ProgressInfo{Completed: initC, Total: initT},
		Work:        ProgressInfo{Completed: workC, Total: workT},
		Define:      ProgressInfo{Completed: defC, Total: defT},
		Edit:        ProgressInfo{Completed: editC, Total: editT},
		Publish:     ProgressInfo{Completed: pubC, Total: pubT},
		PostPublish: ProgressInfo{Completed: postC, Total: postT},
	}
}

// handleGetVideos returns all videos for a given phase.
func (s *Server) handleGetVideos(w http.ResponseWriter, r *http.Request) {
	phaseStr := r.URL.Query().Get("phase")
	if phaseStr == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: phase", "")
		return
	}

	phase, err := strconv.Atoi(phaseStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid phase parameter", "phase must be an integer")
		return
	}

	videos, err := s.videoService.GetVideosByPhase(phase)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get videos", err.Error())
		return
	}

	result := make([]VideoResponse, 0, len(videos))
	for _, v := range videos {
		result = append(result, s.enrichVideo(v))
	}
	respondJSON(w, http.StatusOK, result)
}

// handleGetVideosList returns a lightweight list of videos for a given phase.
func (s *Server) handleGetVideosList(w http.ResponseWriter, r *http.Request) {
	phaseStr := r.URL.Query().Get("phase")
	if phaseStr == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: phase", "")
		return
	}

	phase, err := strconv.Atoi(phaseStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid phase parameter", "phase must be an integer")
		return
	}

	videos, err := s.videoService.GetVideosByPhase(phase)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get videos", err.Error())
		return
	}

	items := make([]VideoListItem, 0, len(videos))
	for _, v := range videos {
		title := v.GetUploadTitle()
		overallC, overallT := s.videoManager.CalculateOverallProgress(v)
		items = append(items, VideoListItem{
			ID:       v.Category + "/" + v.Name,
			Name:     v.Name,
			Category: v.Category,
			Date:     v.Date,
			Title:    title,
			Phase:    video.CalculateVideoPhase(v),
			Progress: ProgressInfo{Completed: overallC, Total: overallT},
		})
	}
	respondJSON(w, http.StatusOK, items)
}

// handleGetVideo returns a single video by name and category.
func (s *Server) handleGetVideo(w http.ResponseWriter, r *http.Request) {
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

	respondJSON(w, http.StatusOK, s.enrichVideo(v))
}

// handleCreateVideo creates a new video.
func (s *Server) handleCreateVideo(w http.ResponseWriter, r *http.Request) {
	var req createVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.Name == "" || req.Category == "" {
		respondError(w, http.StatusBadRequest, "name and category are required", "")
		return
	}

	vi, err := s.videoService.CreateVideo(req.Name, req.Category, req.Date)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create video", err.Error())
		return
	}

	// Read back the created video to return enriched data
	v, err := s.videoService.GetVideo(vi.Name, vi.Category)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "video created but failed to read back", err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, s.enrichVideo(v))
}

// handleUpdateVideo updates an existing video.
func (s *Server) handleUpdateVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}

	// Get existing video first
	existing, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "video not found", err.Error())
		return
	}

	// Decode the update payload over the existing video
	if err := json.NewDecoder(r.Body).Decode(&existing); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if err := s.videoService.UpdateVideo(existing); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update video", err.Error())
		return
	}

	// Read back the updated video
	updated, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "video updated but failed to read back", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, s.enrichVideo(updated))
}

// handleDeleteVideo deletes a video by name and category.
func (s *Server) handleDeleteVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}

	if err := s.videoService.DeleteVideo(videoName, category); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete video", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
