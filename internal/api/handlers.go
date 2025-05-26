package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/workflow"

	"github.com/go-chi/chi/v5"
)

// Request/Response types
type CreateVideoRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

type CreateVideoResponse struct {
	Video storage.VideoIndex `json:"video"`
}

type VideoPhasesResponse struct {
	Phases []PhaseInfo `json:"phases"`
}

type PhaseInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type GetVideosResponse struct {
	Videos []storage.Video `json:"videos"`
}

type GetVideoResponse struct {
	Video storage.Video `json:"video"`
}

type UpdateVideoRequest struct {
	Video storage.Video `json:"video"`
}

type MoveVideoRequest struct {
	TargetDirectoryPath string `json:"target_directory_path"`
}

type CategoriesResponse struct {
	Categories []CategoryInfo `json:"categories"`
}

type CategoryInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// createVideo handles POST /api/videos
func (s *Server) createVideo(w http.ResponseWriter, r *http.Request) {
	var req CreateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	if req.Name == "" || req.Category == "" {
		writeError(w, http.StatusBadRequest, "name and category are required", "")
		return
	}

	video, err := s.videoService.CreateVideo(req.Name, req.Category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create video", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, CreateVideoResponse{Video: video})
}

// getVideoPhases handles GET /api/videos/phases
func (s *Server) getVideoPhases(w http.ResponseWriter, r *http.Request) {
	phases, err := s.videoService.GetVideoPhases()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get video phases", err.Error())
		return
	}

	var phaseInfos []PhaseInfo
	for id, count := range phases {
		if count > 0 {
			phaseInfos = append(phaseInfos, PhaseInfo{
				ID:    id,
				Name:  workflow.PhaseNames[id],
				Count: count,
			})
		}
	}

	writeJSON(w, http.StatusOK, VideoPhasesResponse{Phases: phaseInfos})
}

// getVideos handles GET /api/videos?phase={phase_id}
func (s *Server) getVideos(w http.ResponseWriter, r *http.Request) {
	phaseParam := r.URL.Query().Get("phase")
	if phaseParam == "" {
		writeError(w, http.StatusBadRequest, "phase parameter is required", "")
		return
	}

	phase, err := strconv.Atoi(phaseParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid phase parameter", err.Error())
		return
	}

	videos, err := s.videoService.GetVideosByPhase(phase)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get videos", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, GetVideosResponse{Videos: videos})
}

// getVideo handles GET /api/videos/{videoName}?category={category}
func (s *Server) getVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")

	if videoName == "" || category == "" {
		writeError(w, http.StatusBadRequest, "video name and category are required", "")
		return
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		writeError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, GetVideoResponse{Video: video})
}

// updateVideo handles PUT /api/videos/{videoName}
func (s *Server) updateVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	
	var req UpdateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Ensure the video name matches the URL parameter
	req.Video.Name = videoName

	if err := s.videoService.UpdateVideo(req.Video); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update video", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, GetVideoResponse{Video: req.Video})
}

// deleteVideo handles DELETE /api/videos/{videoName}?category={category}
func (s *Server) deleteVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")

	if videoName == "" || category == "" {
		writeError(w, http.StatusBadRequest, "video name and category are required", "")
		return
	}

	if err := s.videoService.DeleteVideo(videoName, category); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete video", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// moveVideo handles POST /api/videos/{videoName}/move
func (s *Server) moveVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")

	var req MoveVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	if videoName == "" || category == "" || req.TargetDirectoryPath == "" {
		writeError(w, http.StatusBadRequest, "video name, category, and target directory path are required", "")
		return
	}

	if err := s.videoService.MoveVideo(videoName, category, req.TargetDirectoryPath); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to move video", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Video moved successfully"})
}

// getCategories handles GET /api/categories
func (s *Server) getCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := s.videoService.GetCategories()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get categories", err.Error())
		return
	}

	var categoryInfos []CategoryInfo
	for _, cat := range categories {
		categoryInfos = append(categoryInfos, CategoryInfo{
			Name: cat.Name,
			Path: cat.Path,
		})
	}

	writeJSON(w, http.StatusOK, CategoriesResponse{Categories: categoryInfos})
}

// Phase-specific update handlers
func (s *Server) updateVideoInitialDetails(w http.ResponseWriter, r *http.Request) {
	s.updateVideoPhase(w, r, "initial-details")
}

func (s *Server) updateVideoWorkProgress(w http.ResponseWriter, r *http.Request) {
	s.updateVideoPhase(w, r, "work-progress")
}

func (s *Server) updateVideoDefinition(w http.ResponseWriter, r *http.Request) {
	s.updateVideoPhase(w, r, "definition")
}

func (s *Server) updateVideoPostProduction(w http.ResponseWriter, r *http.Request) {
	s.updateVideoPhase(w, r, "post-production")
}

func (s *Server) updateVideoPublishing(w http.ResponseWriter, r *http.Request) {
	s.updateVideoPhase(w, r, "publishing")
}

func (s *Server) updateVideoPostPublish(w http.ResponseWriter, r *http.Request) {
	s.updateVideoPhase(w, r, "post-publish")
}

// updateVideoPhase is a generic handler for phase-specific updates
func (s *Server) updateVideoPhase(w http.ResponseWriter, r *http.Request, phase string) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")

	if videoName == "" || category == "" {
		writeError(w, http.StatusBadRequest, "video name and category are required", "")
		return
	}

	// Get the current video
	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		writeError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	// Decode the update request
	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON", err.Error())
		return
	}

	// Apply the updates based on the phase
	if err := s.applyPhaseUpdates(&video, phase, updateData); err != nil {
		writeError(w, http.StatusBadRequest, "Failed to apply updates", err.Error())
		return
	}

	// Save the updated video
	if err := s.videoService.UpdateVideo(video); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, GetVideoResponse{Video: video})
}

// applyPhaseUpdates applies updates to specific video phase fields
func (s *Server) applyPhaseUpdates(video *storage.Video, phase string, updateData map[string]interface{}) error {
	switch phase {
	case "initial-details":
		return s.applyInitialDetailsUpdates(video, updateData)
	case "work-progress":
		return s.applyWorkProgressUpdates(video, updateData)
	case "definition":
		return s.applyDefinitionUpdates(video, updateData)
	case "post-production":
		return s.applyPostProductionUpdates(video, updateData)
	case "publishing":
		return s.applyPublishingUpdates(video, updateData)
	case "post-publish":
		return s.applyPostPublishUpdates(video, updateData)
	default:
		return fmt.Errorf("unknown phase: %s", phase)
	}
}

// Utility functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   message,
		Message: details,
	})
}