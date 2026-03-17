package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"devopstoolkit/youtube-automation/internal/publishing"

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
		respondError(w, http.StatusNotImplemented, "Google Drive not configured", "Set gdrive.credentialsFile in settings.yaml or GDRIVE_CREDENTIALS_FILE env var")
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
	addSyncWarningMap(resp, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// handleDriveUploadVideo handles multipart file upload of a video to Google Drive
// and updates the video's VideoFile and VideoDriveFileID fields.
//
// POST /api/drive/upload/video/{videoName}?category=X
// Body: multipart/form-data with file field "video"
// Response: {"driveFileId": "...", "videoFile": "drive://..."}
func (s *Server) handleDriveUploadVideo(w http.ResponseWriter, r *http.Request) {
	if s.driveService == nil {
		respondError(w, http.StatusNotImplemented, "Google Drive not configured", "Set gdrive.credentialsFile in settings.yaml or GDRIVE_CREDENTIALS_FILE env var")
		return
	}

	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")

	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "Query parameter 'category' is required")
		return
	}

	// Parse multipart form (32MB memory buffer, spills to disk for large files)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid multipart form", err.Error())
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Missing video file", "multipart field 'video' is required")
		return
	}
	defer file.Close()

	// Detect MIME type from the file header
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	// Get the video
	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
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

	// Upload to Google Drive with title-based filename
	filename := "video" + videoExtensionFromMIME(mimeType)
	if len(video.Titles) > 0 && video.Titles[0].Text != "" {
		filename = publishing.SanitizeTitle(video.Titles[0].Text) + videoExtensionFromMIME(mimeType)
	}
	fileID, err := s.driveService.UploadFile(r.Context(), filename, file, mimeType, uploadFolderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Drive upload failed", err.Error())
		return
	}

	// Update video fields
	video.VideoDriveFileID = fileID
	video.VideoFile = "drive://" + fileID
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := map[string]interface{}{
		"driveFileId": fileID,
		"videoFile":   video.VideoFile,
	}
	addSyncWarningMap(resp, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// handleDriveDownloadVideo streams a video file from Google Drive.
//
// GET /api/drive/download/video/{videoName}?category=X
func (s *Server) handleDriveDownloadVideo(w http.ResponseWriter, r *http.Request) {
	if s.driveService == nil {
		respondError(w, http.StatusNotImplemented, "Google Drive not configured", "Set gdrive.credentialsFile in settings.yaml or GDRIVE_CREDENTIALS_FILE env var")
		return
	}

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

	if video.VideoDriveFileID == "" {
		respondError(w, http.StatusNotFound, "No video file uploaded", "VideoDriveFileID is empty")
		return
	}

	content, mimeType, filename, err := s.driveService.GetFile(r.Context(), video.VideoDriveFileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to download from Drive", err.Error())
		return
	}
	defer content.Close()

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := io.Copy(w, content); err != nil {
		// Response already started, can't send error JSON
		return
	}
}

// handleDriveUploadShort handles multipart file upload of a short video to Google Drive
// and updates the short's DriveFileID and FilePath fields.
//
// POST /api/drive/upload/short/{videoName}/{shortId}?category=X
// Body: multipart/form-data with file field "short"
// Response: {"driveFileId": "...", "filePath": "drive://..."}
func (s *Server) handleDriveUploadShort(w http.ResponseWriter, r *http.Request) {
	if s.driveService == nil {
		respondError(w, http.StatusNotImplemented, "Google Drive not configured", "Set gdrive.credentialsFile in settings.yaml or GDRIVE_CREDENTIALS_FILE env var")
		return
	}

	videoName := chi.URLParam(r, "videoName")
	shortID := chi.URLParam(r, "shortId")
	category := r.URL.Query().Get("category")

	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "Query parameter 'category' is required")
		return
	}

	// Parse multipart form (32MB memory buffer)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid multipart form", err.Error())
		return
	}

	file, header, err := r.FormFile("short")
	if err != nil {
		respondError(w, http.StatusBadRequest, "Missing short file", "multipart field 'short' is required")
		return
	}
	defer file.Close()

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	// Find the short by ID
	shortIdx := -1
	for i, sh := range video.Shorts {
		if sh.ID == shortID {
			shortIdx = i
			break
		}
	}
	if shortIdx < 0 {
		respondError(w, http.StatusNotFound, "Short not found", fmt.Sprintf("short ID %q not found", shortID))
		return
	}

	// Create nested folder: video subfolder → shorts subfolder
	uploadFolderID := s.driveFolderID
	if s.driveFolderID != "" {
		videoFolderID, err := s.driveService.FindOrCreateFolder(r.Context(), videoName, s.driveFolderID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Drive folder creation failed", err.Error())
			return
		}
		shortsFolderID, err := s.driveService.FindOrCreateFolder(r.Context(), "shorts", videoFolderID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Drive folder creation failed", err.Error())
			return
		}
		uploadFolderID = shortsFolderID
	}

	// Upload with filename based on short ID
	filename := shortID + videoExtensionFromMIME(mimeType)
	fileID, err := s.driveService.UploadFile(r.Context(), filename, file, mimeType, uploadFolderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Drive upload failed", err.Error())
		return
	}

	// Update the short's Drive fields; clear stale YouTubeID since the file changed
	video.Shorts[shortIdx].DriveFileID = fileID
	video.Shorts[shortIdx].FilePath = "drive://" + fileID
	video.Shorts[shortIdx].YouTubeID = ""
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := map[string]interface{}{
		"driveFileId": fileID,
		"filePath":    "drive://" + fileID,
	}
	addSyncWarningMap(resp, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// handleDriveDownloadShort streams a short video file from Google Drive.
//
// GET /api/drive/download/short/{videoName}/{shortId}?category=X
func (s *Server) handleDriveDownloadShort(w http.ResponseWriter, r *http.Request) {
	if s.driveService == nil {
		respondError(w, http.StatusNotImplemented, "Google Drive not configured", "Set gdrive.credentialsFile in settings.yaml or GDRIVE_CREDENTIALS_FILE env var")
		return
	}

	videoName := chi.URLParam(r, "videoName")
	shortID := chi.URLParam(r, "shortId")
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

	// Find the short by ID
	var shortIdx int = -1
	for i, sh := range video.Shorts {
		if sh.ID == shortID {
			shortIdx = i
			break
		}
	}
	if shortIdx < 0 {
		respondError(w, http.StatusNotFound, "Short not found", fmt.Sprintf("short ID %q not found", shortID))
		return
	}

	if video.Shorts[shortIdx].DriveFileID == "" {
		respondError(w, http.StatusNotFound, "No short file uploaded", "DriveFileID is empty")
		return
	}

	content, mimeType, filename, err := s.driveService.GetFile(r.Context(), video.Shorts[shortIdx].DriveFileID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to download from Drive", err.Error())
		return
	}
	defer content.Close()

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	if _, err := io.Copy(w, content); err != nil {
		return
	}
}

// videoExtensionFromMIME returns a file extension for common video MIME types.
func videoExtensionFromMIME(mimeType string) string {
	switch mimeType {
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	default:
		return ""
	}
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
