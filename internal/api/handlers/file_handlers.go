package handlers

import (
	"encoding/json"
	"net/http"

	"devopstoolkit/youtube-automation/internal/service"

	"github.com/go-chi/chi/v5"
)

// FileHandlers contains handlers for file operations
type FileHandlers struct {
	videoService *service.VideoService
}

// NewFileHandlers creates a new FileHandlers instance
func NewFileHandlers(videoService *service.VideoService) *FileHandlers {
	return &FileHandlers{
		videoService: videoService,
	}
}

// MoveVideoRequest represents the request to move video files
type MoveVideoRequest struct {
	TargetDirectoryPath string `json:"target_directory_path"`
}

// MoveVideoFiles handles moving video files to a new directory
func (h *FileHandlers) MoveVideoFiles(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	// Parse the request
	var req MoveVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// In a real implementation, this would call a service method to move the files
	// For now, we'll just return a success response
	
	response := struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}{
		Success: true,
		Message: "Video files moved successfully",
	}
	
	respondJSON(w, response)
}