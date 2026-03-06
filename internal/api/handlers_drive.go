package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// handleDriveUploadThumbnail handles multipart file upload to Google Drive
// and updates the video's ThumbnailVariant with the Drive file ID.
//
// POST /api/drive/upload/thumbnail/{videoName}?category=X&variantIndex=N
// Body: multipart/form-data with file field "thumbnail"
// Response: {"driveFileId": "...", "variantIndex": 0}
func (s *Server) handleDriveUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	if s.driveService == nil {
		respondError(w, http.StatusNotImplemented, "Google Drive not configured", "Set gdrive.credentialsFile in settings.yaml")
		return
	}

	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	variantIndexStr := r.URL.Query().Get("variantIndex")

	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "Query parameter 'category' is required")
		return
	}
	if variantIndexStr == "" {
		respondError(w, http.StatusBadRequest, "Missing variantIndex", "Query parameter 'variantIndex' is required")
		return
	}

	variantIndex, err := strconv.Atoi(variantIndexStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid variantIndex", "variantIndex must be an integer")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid multipart form", err.Error())
		return
	}

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Missing thumbnail file", "multipart field 'thumbnail' is required")
		return
	}
	defer file.Close()

	// Detect MIME type from the file header
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Get the video (name, category order)
	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	// Validate variant index
	if variantIndex < 0 || variantIndex >= len(video.ThumbnailVariants) {
		respondError(w, http.StatusBadRequest, "Invalid variantIndex",
			fmt.Sprintf("variantIndex %d out of range (video has %d variants)", variantIndex, len(video.ThumbnailVariants)))
		return
	}

	// Find or create a subfolder named after the video
	uploadFolderID := s.driveFolderID
	if s.driveFolderID != "" {
		subfolderID, err := s.driveService.FindOrCreateFolder(r.Context(), videoName, s.driveFolderID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Drive folder creation failed", err.Error())
			return
		}
		uploadFolderID = subfolderID
	}

	// Upload to Google Drive
	filename := fmt.Sprintf("thumbnail-%d%s", variantIndex, extensionFromMIME(mimeType))
	fileID, err := s.driveService.UploadFile(r.Context(), filename, file, mimeType, uploadFolderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Drive upload failed", err.Error())
		return
	}

	// Update the video's thumbnail variant
	video.ThumbnailVariants[variantIndex].DriveFileID = fileID
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := map[string]interface{}{
		"driveFileId":  fileID,
		"variantIndex": variantIndex,
	}
	if syncErr := s.videoService.LastSyncError(); syncErr != nil {
		resp["syncWarning"] = "git sync failed: " + syncErr.Error()
	} else if !s.videoService.IsSyncConfigured() {
		resp["syncWarning"] = "git sync not configured — changes saved locally only"
	}
	respondJSON(w, http.StatusOK, resp)
}

// extensionFromMIME returns a file extension for common image MIME types.
func extensionFromMIME(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ""
	}
}
