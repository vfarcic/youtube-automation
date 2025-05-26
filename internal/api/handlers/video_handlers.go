package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/go-chi/chi/v5"
)

// VideoHandlers contains all handlers for video-related endpoints
type VideoHandlers struct {
	videoService *service.VideoService
}

// NewVideoHandlers creates a new VideoHandlers instance
func NewVideoHandlers(videoService *service.VideoService) *VideoHandlers {
	return &VideoHandlers{
		videoService: videoService,
	}
}

// GetVideoPhases returns all video phases with counts
func (h *VideoHandlers) GetVideoPhases(w http.ResponseWriter, r *http.Request) {
	phases, err := h.videoService.GetVideoPhases()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, phases)
}

// GetVideosByPhase returns all videos in a specific phase
func (h *VideoHandlers) GetVideosByPhase(w http.ResponseWriter, r *http.Request) {
	phaseIDStr := r.URL.Query().Get("phase")
	if phaseIDStr == "" {
		http.Error(w, "phase parameter is required", http.StatusBadRequest)
		return
	}

	phaseID, err := strconv.Atoi(phaseIDStr)
	if err != nil {
		http.Error(w, "invalid phase ID", http.StatusBadRequest)
		return
	}

	videos, err := h.videoService.GetVideosByPhase(phaseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, videos)
}

// GetVideo returns a specific video
func (h *VideoHandlers) GetVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	video, err := h.videoService.GetVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, video)
}

// CreateVideo creates a new video
func (h *VideoHandlers) CreateVideo(w http.ResponseWriter, r *http.Request) {
	var req service.VideoCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	video, err := h.videoService.CreateVideo(req)
	if err != nil {
		if err == service.ErrInvalidRequest {
			http.Error(w, "invalid request: name and category are required", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	respondJSON(w, video)
}

// UpdateVideo updates a video
func (h *VideoHandlers) UpdateVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	var updatedVideo storage.Video
	if err := json.NewDecoder(r.Body).Decode(&updatedVideo); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	video, err := h.videoService.UpdateVideo(videoID, updatedVideo)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	respondJSON(w, video)
}

// DeleteVideo deletes a video
func (h *VideoHandlers) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	err := h.videoService.DeleteVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetCategories returns all available video categories
func (h *VideoHandlers) GetCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.videoService.GetCategories()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to a response structure
	response := struct {
		Categories []string `json:"categories"`
	}{
		Categories: categories,
	}

	respondJSON(w, response)
}

// Helper function to send JSON responses
func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}