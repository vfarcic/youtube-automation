package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	Category string `json:"category"`
	Name     string `json:"name"`
}

// ThumbnailConfigRequest is the JSON body for POST /api/videos/{videoName}/thumbnail-config.
type ThumbnailConfigRequest struct {
	Category     string `json:"category"`
	Tagline      string `json:"tagline"`
	Illustration string `json:"illustration"`
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

	// Validate path params to prevent traversal
	if strings.Contains(req.Category, "..") || strings.Contains(req.Category, "/") || strings.Contains(req.Category, "\\") ||
		strings.Contains(req.Name, "..") || strings.Contains(req.Name, "/") || strings.Contains(req.Name, "\\") {
		respondError(w, http.StatusBadRequest, "invalid category or name", "")
		return
	}

	// Load stored tagline and illustration from video
	video, err := s.videoService.GetVideo(req.Name, req.Category)
	if err != nil {
		log.Printf("video not found for %s/%s: %v", req.Category, req.Name, err)
		respondError(w, http.StatusNotFound, "video not found", "")
		return
	}
	if video.Tagline == "" {
		respondError(w, http.StatusBadRequest, "tagline must be set before generating thumbnails", "")
		return
	}

	// Load creator photos: prefer Drive screenshots, fall back to local photoDir
	photos, err := s.loadScreenshotsFromDrive(r.Context(), req.Name)
	if err != nil {
		log.Printf("failed to load screenshots from Drive for %s: %v", req.Name, err)
		respondError(w, http.StatusInternalServerError, "Failed to load creator photos from Drive", "")
		return
	}
	if len(photos) == 0 {
		// Fall back to local photo directory
		photos, err = loadPhotos(s.photoDir)
		if err != nil {
			log.Printf("failed to load photos from %s: %v", s.photoDir, err)
			respondError(w, http.StatusInternalServerError, "Failed to load creator photos", "")
			return
		}
	}
	if len(photos) == 0 {
		respondError(w, http.StatusBadRequest, "No creator photos found — upload screenshots to the video's Drive folder", "")
		return
	}

	// Build prompts
	cfgWith := thumbnail.BuildPromptConfig(video.Tagline, video.Illustration, nil)
	cfgWithout := thumbnail.BuildPromptConfig(video.Tagline, "", nil)

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

	// Log internal error details server-side only; return user-friendly messages to clients
	var sanitizedErrors []string
	for _, e := range genErrs {
		log.Printf("thumbnail generation error: %v", e)
		sanitizedErrors = append(sanitizedErrors, classifyGenerationError(e))
	}

	if len(metas) == 0 && len(sanitizedErrors) > 0 {
		respondError(w, http.StatusInternalServerError, classifyGenerationError(genErrs[0]), "")
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

// handleSaveThumbnailConfig saves the selected tagline and illustration to the video.
//
// POST /api/videos/{videoName}/thumbnail-config
func (s *Server) handleSaveThumbnailConfig(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	if _, ok := validatePathParam(w, videoName, "videoName"); !ok {
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req ThumbnailConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", "")
		return
	}

	if req.Category == "" {
		respondError(w, http.StatusBadRequest, "category is required", "")
		return
	}
	if req.Tagline == "" {
		respondError(w, http.StatusBadRequest, "tagline is required", "")
		return
	}

	video, err := s.videoService.GetVideo(videoName, req.Category)
	if err != nil {
		log.Printf("video not found for %s/%s: %v", req.Category, videoName, err)
		respondError(w, http.StatusNotFound, "video not found", "")
		return
	}

	video.Tagline = req.Tagline
	video.Illustration = req.Illustration

	if err := s.videoService.UpdateVideo(video); err != nil {
		log.Printf("failed to save video %s/%s: %v", req.Category, videoName, err)
		respondError(w, http.StatusInternalServerError, "Failed to save video", "")
		return
	}

	resp := map[string]interface{}{
		"tagline":      video.Tagline,
		"illustration": video.Illustration,
	}
	addSyncWarningMap(resp, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// --- Helpers ---

// loadScreenshotsFromDrive downloads screenshot-* files from the video's Drive folder.
// Returns nil, nil if Drive is not configured (allowing fallback to local photos).
// Downloaded bytes are held in memory only for the duration of the caller's request.
func (s *Server) loadScreenshotsFromDrive(ctx context.Context, videoName string) ([][]byte, error) {
	if s.driveService == nil || s.driveFolderID == "" {
		return nil, nil
	}

	folderID, err := s.driveService.FindOrCreateFolder(ctx, videoName, s.driveFolderID)
	if err != nil {
		return nil, fmt.Errorf("finding Drive folder for %s: %w", videoName, err)
	}

	files, err := s.driveService.ListFilesInFolder(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("listing files in Drive folder: %w", err)
	}

	var photos [][]byte
	for _, f := range files {
		if !strings.HasPrefix(strings.ToLower(f.Name), "screenshot") {
			continue
		}
		reader, _, _, err := s.driveService.GetFile(ctx, f.ID)
		if err != nil {
			log.Printf("failed to download %s from Drive: %v", f.Name, err)
			continue
		}
		data, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			log.Printf("failed to read %s from Drive: %v", f.Name, err)
			continue
		}
		photos = append(photos, data)
	}
	return photos, nil
}

// classifyGenerationError returns a user-friendly message for a provider error
// without leaking internal details like API keys or paths.
func classifyGenerationError(err error) string {
	if errors.Is(err, thumbnail.ErrGeminiNoPhotos) || errors.Is(err, thumbnail.ErrGPTImageNoPhotos) {
		return "No creator photos found — upload screenshots to the video's Drive folder"
	}
	if errors.Is(err, thumbnail.ErrGeminiImageFiltered) || errors.Is(err, thumbnail.ErrGPTImageContentFiltered) {
		return "Image was blocked by content safety filters — try a different tagline or illustration"
	}
	if errors.Is(err, thumbnail.ErrGeminiAPIError) || errors.Is(err, thumbnail.ErrGPTImageAPIError) {
		return "Image generation service error — please try again later"
	}
	if errors.Is(err, thumbnail.ErrGeminiEmptyPrompt) || errors.Is(err, thumbnail.ErrGPTImageEmptyPrompt) {
		return "Thumbnail prompt is empty — set a tagline before generating"
	}
	return "Image generation failed — please try again later"
}

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
