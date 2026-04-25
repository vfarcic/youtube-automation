package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/thumbnail"

	"github.com/go-chi/chi/v5"
)

// --- Request / Response types ---

// ThumbnailGenerateRequest is the JSON body for POST /api/thumbnails/generate.
type ThumbnailGenerateRequest struct {
	Category     string `json:"category"`
	Name         string `json:"name"`
	Illustration string `json:"illustration,omitempty"`
	Tagline      string `json:"tagline"`
}

// ThumbnailGenerateMeta describes one generated thumbnail in the response.
type ThumbnailGenerateMeta struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Style    string `json:"style"`
}

// ThumbnailGenerateResponse is the JSON body returned by POST /api/thumbnails/generate.
type ThumbnailGenerateResponse struct {
	Thumbnails []ThumbnailGenerateMeta `json:"thumbnails"`
	Errors     []string                `json:"errors,omitempty"`
}

// ThumbnailSelectRequest is the JSON body for POST /api/thumbnails/generated/{id}/select.
type ThumbnailSelectRequest struct {
	Category     string `json:"category"`
	Name         string `json:"name"`
	VariantIndex int    `json:"variantIndex"`
}

// --- Handlers ---

// handleGenerateThumbnails generates thumbnails across all configured providers.
//
// POST /api/thumbnails/generate
func (s *Server) handleGenerateThumbnails(w http.ResponseWriter, r *http.Request) {
	if len(s.imageGenerators) == 0 {
		respondError(w, http.StatusNotImplemented, "Thumbnail generation not configured", "No image generation providers available")
		return
	}
	if s.imageStore == nil {
		respondError(w, http.StatusNotImplemented, "Thumbnail generation not configured", "Image store not available")
		return
	}

	// Limit request body to 1MB to prevent large-payload DoS
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ThumbnailGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	if req.Category == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "category and name are required", "")
		return
	}
	if req.Tagline == "" {
		respondError(w, http.StatusBadRequest, "tagline is required", "")
		return
	}

	// Validate path params to prevent traversal
	if strings.Contains(req.Category, "..") || strings.Contains(req.Category, "/") || strings.Contains(req.Category, "\\") ||
		strings.Contains(req.Name, "..") || strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") {
		respondError(w, http.StatusBadRequest, "invalid category or name", "")
		return
	}

	// Load creator photos
	photos, err := loadPhotos(s.photoDir)
	if err != nil {
		log.Printf("failed to load photos from %s: %v", s.photoDir, err)
		respondError(w, http.StatusInternalServerError, "Failed to load creator photos", "")
		return
	}

	// Build prompts
	cfgWith := thumbnail.BuildPromptConfig(req.Tagline, req.Illustration, nil)
	cfgWithout := thumbnail.BuildPromptConfig(req.Tagline, "", nil)

	promptWith := thumbnail.BuildPrompt(cfgWith)
	promptWithout := thumbnail.BuildPrompt(cfgWithout)

	genReq := thumbnail.GenerateRequest{
		PromptWithIllustration:    promptWith,
		PromptWithoutIllustration: promptWithout,
		Photos:                    photos,
	}

	// Generate thumbnails across all providers
	images, genErrs := thumbnail.GenerateThumbnails(r.Context(), s.imageGenerators, genReq)

	// Store results
	var metas []ThumbnailGenerateMeta
	for _, img := range images {
		id, storeErr := s.imageStore.Add(img)
		if storeErr != nil {
			log.Printf("failed to store generated image: %v", storeErr)
			genErrs = append(genErrs, storeErr)
			continue
		}
		metas = append(metas, ThumbnailGenerateMeta{
			ID:       id,
			Provider: img.Provider,
			Style:    img.Style,
		})
	}

	// Log internal error details server-side only; return generic messages to clients
	var sanitizedErrors []string
	for _, e := range genErrs {
		log.Printf("thumbnail generation error: %v", e)
		sanitizedErrors = append(sanitizedErrors, "a provider failed to generate an image")
	}

	if len(metas) == 0 && len(sanitizedErrors) > 0 {
		respondError(w, http.StatusInternalServerError, "All providers failed", "")
		return
	}

	respondJSON(w, http.StatusOK, ThumbnailGenerateResponse{
		Thumbnails: metas,
		Errors:     sanitizedErrors,
	})
}

// handleGetGeneratedThumbnail returns the raw image bytes for a generated thumbnail.
//
// GET /api/thumbnails/generated/{id}
func (s *Server) handleGetGeneratedThumbnail(w http.ResponseWriter, r *http.Request) {
	if s.imageStore == nil {
		respondError(w, http.StatusNotImplemented, "Thumbnail generation not configured", "")
		return
	}

	id := chi.URLParam(r, "id")
	if _, ok := validatePathParam(w, id, "id"); !ok {
		return
	}

	img, ok := s.imageStore.Get(id)
	if !ok {
		respondError(w, http.StatusNotFound, "Image not found or expired", "")
		return
	}

	contentType := http.DetectContentType(img.Data)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(img.Data)))
	w.WriteHeader(http.StatusOK)
	w.Write(img.Data)
}

// handleSelectGeneratedThumbnail uploads a generated thumbnail to Google Drive
// and saves it as a ThumbnailVariant on the video.
//
// POST /api/thumbnails/generated/{id}/select
func (s *Server) handleSelectGeneratedThumbnail(w http.ResponseWriter, r *http.Request) {
	if s.imageStore == nil {
		respondError(w, http.StatusNotImplemented, "Thumbnail generation not configured", "")
		return
	}
	if s.driveService == nil {
		respondError(w, http.StatusNotImplemented, "Google Drive not configured", "")
		return
	}

	id := chi.URLParam(r, "id")
	if _, ok := validatePathParam(w, id, "id"); !ok {
		return
	}

	// Limit request body to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ThumbnailSelectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	if req.Category == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "category and name are required", "")
		return
	}
	if strings.Contains(req.Category, "..") || strings.Contains(req.Category, "/") || strings.Contains(req.Category, "\\") ||
		strings.Contains(req.Name, "..") || strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") {
		respondError(w, http.StatusBadRequest, "invalid category or name", "")
		return
	}

	// Atomically claim the image from store (get + remove in one locked operation)
	// to prevent TOCTOU race where two concurrent requests could both upload the same image.
	img, ok := s.imageStore.Claim(id)
	if !ok {
		respondError(w, http.StatusNotFound, "Image not found or expired", "")
		return
	}

	// Get the video
	video, err := s.videoService.GetVideo(req.Name, req.Category)
	if err != nil {
		log.Printf("video lookup failed for %s/%s: %v", req.Category, req.Name, err)
		respondError(w, http.StatusNotFound, "Video not found", "")
		return
	}

	// Validate variant index; auto-create if it's the next slot
	if req.VariantIndex < 0 || req.VariantIndex > len(video.ThumbnailVariants) {
		respondError(w, http.StatusBadRequest, "Invalid variantIndex", "")
		return
	}
	if req.VariantIndex == len(video.ThumbnailVariants) {
		video.ThumbnailVariants = append(video.ThumbnailVariants, storage.ThumbnailVariant{
			Index: req.VariantIndex + 1,
		})
	}

	// Find or create subfolder for the video
	uploadFolderID := s.driveFolderID
	if s.driveFolderID != "" {
		subfolderID, err := s.driveService.FindOrCreateFolder(r.Context(), req.Name, s.driveFolderID)
		if err != nil {
			log.Printf("Drive folder creation failed for %s: %v", req.Name, err)
			respondError(w, http.StatusInternalServerError, "Drive folder creation failed", "")
			return
		}
		uploadFolderID = subfolderID
	}

	// Detect MIME type from image data
	mimeType := http.DetectContentType(img.Data)
	ext := extensionFromMIME(mimeType)
	if ext == "" {
		ext = ".png" // default
	}

	filename := fmt.Sprintf("thumbnail-%d-generated%s", req.VariantIndex, ext)
	fileID, err := s.driveService.UploadFile(r.Context(), filename, bytes.NewReader(img.Data), mimeType, uploadFolderID)
	if err != nil {
		log.Printf("Drive upload failed for %s: %v", filename, err)
		respondError(w, http.StatusInternalServerError, "Drive upload failed", "")
		return
	}

	// Update the video's thumbnail variant
	video.ThumbnailVariants[req.VariantIndex].DriveFileID = fileID
	if err := s.videoService.UpdateVideo(video); err != nil {
		log.Printf("failed to save video %s/%s: %v", req.Category, req.Name, err)
		respondError(w, http.StatusInternalServerError, "Failed to save video", "")
		return
	}

	resp := map[string]interface{}{
		"driveFileId":  fileID,
		"variantIndex": req.VariantIndex,
	}
	addSyncWarningMap(resp, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// --- Helpers ---

// loadPhotos reads all image files from the given directory.
// Returns an empty slice if dir is empty or doesn't exist.
func loadPhotos(dir string) ([][]byte, error) {
	if dir == "" {
		return nil, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading photo directory: %w", err)
	}

	var photos [][]byte
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if !strings.HasSuffix(name, ".jpg") && !strings.HasSuffix(name, ".jpeg") &&
			!strings.HasSuffix(name, ".png") && !strings.HasSuffix(name, ".webp") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("reading photo %s: %w", entry.Name(), err)
		}
		photos = append(photos, data)
	}
	return photos, nil
}
