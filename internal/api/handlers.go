package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/storage"
	video2 "devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
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
	Video VideoWithID `json:"video"`
}

// VideoWithID extends storage.Video with the string-based ID field
type VideoWithID struct {
	ID string `json:"id"`
	storage.Video
}

// VideoListItem represents a lightweight video object optimized for list views
// Reduces payload size from ~8.8KB to ~200 bytes per video (97% reduction)
type VideoListItem struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Title     string        `json:"title"`
	Date      string        `json:"date"`
	Thumbnail string        `json:"thumbnail"`
	Category  string        `json:"category"`
	Status    string        `json:"status"`
	Phase     int           `json:"phase"`
	Progress  VideoProgress `json:"progress"`
}

// VideoProgress represents the completion status for a video
type VideoProgress struct {
	Completed int `json:"completed"`
	Total     int `json:"total"`
}

// VideoListResponse contains the optimized video list for frontend consumption
type VideoListResponse struct {
	Videos []VideoListItem `json:"videos"`
}

// VideoPhaseListHandler handles GET /api/videos/list requests for video list data
// This endpoint provides a lightweight API for frontend video lists with calculated progress
func (s *Server) VideoPhaseListHandler(w http.ResponseWriter, r *http.Request) {
	// Implementation of VideoPhaseListHandler
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

// AI Request/Response types
type AIRequest struct {
	Manuscript string `json:"manuscript"`
}

type AITitlesResponse struct {
	Titles []string `json:"titles"`
}

type AIDescriptionResponse struct {
	Description string `json:"description"`
}

type AITagsResponse struct {
	Tags []string `json:"tags"`
}

type AITweetsResponse struct {
	Tweets []string `json:"tweets"`
}


type AIDescriptionTagsResponse struct {
	DescriptionTags []string `json:"description_tags"`
}

type AnimationsResponse struct {
	Animations []string `json:"animations"`
}

// createVideo handles POST /api/videos
func (s *Server) createVideo(w http.ResponseWriter, r *http.Request) {
	var req CreateVideoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Name == "" || req.Category == "" {
		writeError(w, http.StatusBadRequest, "name and category are required")
		return
	}

	// API creates videos with empty date by default
	video, err := s.videoService.CreateVideo(req.Name, req.Category, "")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create video")
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

	var videos []storage.Video
	var err error

	if phaseParam == "" {
		// No phase parameter provided - return all videos from all phases
		videos, err = s.videoService.GetAllVideos()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get videos", err.Error())
			return
		}
	} else {
		// Phase parameter provided - validate and return videos for specific phase
		phase, parseErr := strconv.Atoi(phaseParam)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "Invalid phase parameter", parseErr.Error())
			return
		}

		videos, err = s.videoService.GetVideosByPhase(phase)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get videos", err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, GetVideosResponse{Videos: videos})
}

// getVideosList handles GET /api/videos/list?phase={phase_id}
// Returns optimized lightweight video data for frontend list views
// Reduces payload from ~8.8KB per video to ~200 bytes (97% reduction)
func (s *Server) getVideosList(w http.ResponseWriter, r *http.Request) {
	phaseParam := r.URL.Query().Get("phase")

	var videos []storage.Video
	var err error

	if phaseParam == "" {
		// No phase parameter provided - return all videos from all phases
		videos, err = s.videoService.GetAllVideos()
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get videos", err.Error())
			return
		}
	} else {
		// Phase parameter provided - validate and return videos for specific phase
		phase, parseErr := strconv.Atoi(phaseParam)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "Invalid phase parameter", parseErr.Error())
			return
		}

		videos, err = s.videoService.GetVideosByPhase(phase)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to get videos", err.Error())
			return
		}
	}

	// Transform to lightweight format for optimal performance
	lightweightVideos := transformToVideoListItems(videos)

	writeJSON(w, http.StatusOK, VideoListResponse{Videos: lightweightVideos})
}

// generateVideoID creates a string-based ID for a video using category and name
func generateVideoID(video storage.Video) string {
	return video.Category + "/" + video.Name
}

// transformToVideoListItems converts full Video objects to lightweight VideoListItem format
// This reduces payload size from ~8.8KB to ~200 bytes per video (97% reduction)
func transformToVideoListItems(videos []storage.Video) []VideoListItem {
	result := make([]VideoListItem, 0, len(videos))

	for _, video := range videos {
		// Use shared video manager for consistent progress calculation
		videoManager := video2.NewManager(nil) // We don't need filePathFunc for calculations
		overallCompleted, overallTotal := videoManager.CalculateOverallProgress(video)

		// Determine status based on publishing completion using video manager
		status := "draft"
		publishCompleted, publishTotal := videoManager.CalculatePublishingProgress(video)
		if publishTotal > 0 && publishCompleted == publishTotal {
			status = "published"
		}

		// Use the shared phase calculation logic
		phase := video2.CalculateVideoPhase(video)

		// Handle edge cases for missing fields
		title := video.GetUploadTitle()
		if title == "" {
			title = video.Name // Fallback to name if title is empty
		}

		date := video.Date
		if date == "" {
			date = "TBD" // Indicate date to be determined
		}

		thumbnail := video.Thumbnail
		if thumbnail == "" {
			thumbnail = "default.jpg" // Default thumbnail placeholder
		}

		// Generate string-based ID from category and name (already sanitized)
		videoID := video.Category + "/" + video.Name

		item := VideoListItem{
			ID:        videoID,
			Name:      video.Name, // Name is now already sanitized
			Title:     title,
			Date:      date,
			Thumbnail: thumbnail,
			Category:  video.Category,
			Status:    status,
			Phase:     phase,
			Progress: VideoProgress{
				Completed: overallCompleted,
				Total:     overallTotal,
			},
		}

		result = append(result, item)
	}

	return result
}

// getVideo handles GET /api/videos/{videoName}?category={category}
func (s *Server) getVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")

	if videoName == "" || category == "" {
		writeError(w, http.StatusBadRequest, "video name and category query parameter are required", "")
		return
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		writeError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	// No need for videoForAPI since video.Name is now already sanitized
	videoWithID := VideoWithID{
		ID:    generateVideoID(video),
		Video: video,
	}
	writeJSON(w, http.StatusOK, GetVideoResponse{Video: videoWithID})
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

	// No need for videoForAPI since req.Video.Name is already correct
	videoWithID := VideoWithID{
		ID:    generateVideoID(req.Video),
		Video: req.Video,
	}
	writeJSON(w, http.StatusOK, GetVideoResponse{Video: videoWithID})
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
		writeError(w, http.StatusBadRequest, "video name and category query parameter are required", "")
		return
	}

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON for update data", err.Error())
		return
	}

	// First, get the video
	videoToUpdate, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		writeError(w, http.StatusNotFound, "Video not found or error fetching video", err.Error())
		return
	}

	// Now, update the phase with a pointer to the fetched video
	updatedVideoPtr, err := s.videoService.UpdateVideoPhase(&videoToUpdate, phase, updateData)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update video phase", err.Error())
		return
	}

	// Ensure the pointer is not nil before dereferencing for the response
	if updatedVideoPtr == nil {
		writeError(w, http.StatusInternalServerError, "UpdateVideoPhase returned nil video without error", "")
		return
	}

	// No need for videoForAPI since updatedVideoPtr.Name is already correct
	videoWithID := VideoWithID{
		ID:    generateVideoID(*updatedVideoPtr),
		Video: *updatedVideoPtr,
	}
	writeJSON(w, http.StatusOK, GetVideoResponse{Video: videoWithID})
}

// getEditingAspects handles GET /api/editing/aspects
// Returns lightweight overview of all aspects without fields
// Optional query params: videoName and category for completion tracking
func (s *Server) getEditingAspects(w http.ResponseWriter, r *http.Request) {
	aspectOverview := s.aspectService.GetAspectsOverview()

	// Check for optional video context for completion tracking
	videoName := r.URL.Query().Get("videoName")
	category := r.URL.Query().Get("category")

	// If video context is provided, calculate completion counts
	if videoName != "" && category != "" {
		video, err := s.videoService.GetVideo(videoName, category)
		if err != nil {
			// If video is not found, continue with default 0 completion counts
			// Don't fail the entire request for invalid video context
		} else {
			// Calculate completion counts for each aspect
			s.calculateCompletionCounts(&aspectOverview, video)
		}
	}

	writeJSON(w, http.StatusOK, aspectOverview)
}

// getAspectFields handles GET /api/editing/aspects/{aspectKey}/fields
// Returns detailed field information for a specific aspect
func (s *Server) getAspectFields(w http.ResponseWriter, r *http.Request) {
	aspectKey := chi.URLParam(r, "aspectKey")
	if aspectKey == "" {
		writeError(w, http.StatusBadRequest, "aspect key is required")
		return
	}

	aspectFields, err := s.aspectService.GetAspectFields(aspectKey)
	if err != nil {
		if err.Error() == "aspect not found" {
			writeError(w, http.StatusNotFound, "aspect not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to get aspect fields")
		return
	}

	writeJSON(w, http.StatusOK, aspectFields)
}

// calculateCompletionCounts updates the aspectOverview with completion counts based on video data
// Uses the shared video manager calculation functions for consistency with CLI
func (s *Server) calculateCompletionCounts(aspectOverview *aspect.AspectOverview, video storage.Video) {
	// Use shared video manager for consistent progress calculations
	videoManager := video2.NewManager(nil) // We don't need filePathFunc for calculations

	// Map aspect keys to their corresponding calculation functions
	aspectCalculations := map[string]func(storage.Video) (int, int){
		"initial-details": videoManager.CalculateInitialDetailsProgress,
		"work-progress":   videoManager.CalculateWorkProgressProgress,
		"definition":      videoManager.CalculateDefinePhaseCompletion,
		"post-production": videoManager.CalculatePostProductionProgress,
		"publishing":      videoManager.CalculatePublishingProgress,
		"post-publish":    videoManager.CalculatePostPublishProgress,
	}

	// Update completion counts for each aspect
	for i, aspectSummary := range aspectOverview.Aspects {
		if calcFunc, exists := aspectCalculations[aspectSummary.Key]; exists {
			completed, _ := calcFunc(video)
			aspectOverview.Aspects[i].CompletedFieldCount = completed
		}
		// If aspect not found in mapping, keep default value of 0
	}
}

// Utility functions
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorData := map[string]string{"error": message}
	if len(details) > 0 && details[0] != "" {
		errorData["details"] = details[0]
	}

	json.NewEncoder(w).Encode(errorData)
}

// AI Handlers - these call the existing AI functions directly

// aiTitles handles POST /api/ai/titles
func (s *Server) aiTitles(w http.ResponseWriter, r *http.Request) {
	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Manuscript == "" {
		writeError(w, http.StatusBadRequest, "manuscript is required")
		return
	}


	titles, err := ai.SuggestTitles(r.Context(), req.Manuscript)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate titles", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AITitlesResponse{Titles: titles})
}

// aiTitlesWithVideo handles POST /api/ai/titles/{videoName}?category={category}
// This new endpoint uses URL parameters instead of JSON payload and reads manuscript from video files
func (s *Server) aiTitlesWithVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video name from URL path
	videoName := chi.URLParam(r, "videoName")
	if videoName == "" {
		writeError(w, http.StatusBadRequest, "videoName is required")
		return
	}

	// Extract category from query parameter
	category := r.URL.Query().Get("category")
	if category == "" {
		writeError(w, http.StatusBadRequest, "category query parameter is required")
		return
	}

	// Get manuscript content using our new VideoService method
	manuscriptContent, err := s.videoService.GetVideoManuscript(videoName, category)
	if err != nil {
		// Check if it's a "video not found" error vs other errors
		if strings.Contains(err.Error(), "failed to get video") {
			writeError(w, http.StatusNotFound, "Video not found", err.Error())
			return
		}
		// Check if it's a "gist field empty" error (user error)
		if strings.Contains(err.Error(), "gist field is empty") {
			writeError(w, http.StatusBadRequest, "Video manuscript not configured", err.Error())
			return
		}
		// Other errors (file read errors, etc.) are server errors
		writeError(w, http.StatusInternalServerError, "Failed to read manuscript", err.Error())
		return
	}

	// Get AI configuration

	// Generate titles using AI
	titles, err := ai.SuggestTitles(r.Context(), manuscriptContent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate titles", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AITitlesResponse{Titles: titles})
}

// aiDescription handles POST /api/ai/description
func (s *Server) aiDescription(w http.ResponseWriter, r *http.Request) {
	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Manuscript == "" {
		writeError(w, http.StatusBadRequest, "manuscript is required")
		return
	}


	description, err := ai.SuggestDescription(r.Context(), req.Manuscript)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate description", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AIDescriptionResponse{Description: description})
}

// aiTags handles POST /api/ai/tags
func (s *Server) aiTags(w http.ResponseWriter, r *http.Request) {
	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Manuscript == "" {
		writeError(w, http.StatusBadRequest, "manuscript is required")
		return
	}


	tags, err := ai.SuggestTags(r.Context(), req.Manuscript)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate tags", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AITagsResponse{Tags: []string{tags}})
}

// aiTweets handles POST /api/ai/tweets
func (s *Server) aiTweets(w http.ResponseWriter, r *http.Request) {
	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Manuscript == "" {
		writeError(w, http.StatusBadRequest, "manuscript is required")
		return
	}


	tweets, err := ai.SuggestTweets(r.Context(), req.Manuscript)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate tweets", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AITweetsResponse{Tweets: tweets})
}


// aiDescriptionTags handles POST /api/ai/description-tags
func (s *Server) aiDescriptionTags(w http.ResponseWriter, r *http.Request) {
	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Manuscript == "" {
		writeError(w, http.StatusBadRequest, "manuscript is required")
		return
	}


	descriptionTags, err := ai.SuggestDescriptionTags(r.Context(), req.Manuscript)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate description tags", err.Error())
		return
	}

	// Split the space-separated tags into a slice
	var descriptionTagsSlice []string
	if descriptionTags != "" {
		descriptionTagsSlice = strings.Fields(descriptionTags)
	}
	writeJSON(w, http.StatusOK, AIDescriptionTagsResponse{DescriptionTags: descriptionTagsSlice})
}

// aiDescriptionWithVideo handles POST /api/ai/description/{videoName}?category={category}
// This new endpoint uses URL parameters instead of JSON payload and reads manuscript from video files
func (s *Server) aiDescriptionWithVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video name from URL path
	videoName := chi.URLParam(r, "videoName")
	if videoName == "" {
		writeError(w, http.StatusBadRequest, "videoName is required")
		return
	}

	// Extract category from query parameter
	category := r.URL.Query().Get("category")
	if category == "" {
		writeError(w, http.StatusBadRequest, "category query parameter is required")
		return
	}

	// Get manuscript content using our VideoService method
	manuscriptContent, err := s.videoService.GetVideoManuscript(videoName, category)
	if err != nil {
		// Check if it's a "video not found" error vs other errors
		if strings.Contains(err.Error(), "failed to get video") {
			writeError(w, http.StatusNotFound, "Video not found", err.Error())
			return
		}
		// Check if it's a "gist field empty" error (user error)
		if strings.Contains(err.Error(), "gist field is empty") {
			writeError(w, http.StatusBadRequest, "Video manuscript not configured", err.Error())
			return
		}
		// Other errors (file read errors, etc.) are server errors
		writeError(w, http.StatusInternalServerError, "Failed to read manuscript", err.Error())
		return
	}

	// Get AI configuration

	// Generate description using AI
	description, err := ai.SuggestDescription(r.Context(), manuscriptContent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate description", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AIDescriptionResponse{Description: description})
}

// aiTagsWithVideo handles POST /api/ai/tags/{videoName}?category={category}
// This new endpoint uses URL parameters instead of JSON payload and reads manuscript from video files
func (s *Server) aiTagsWithVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video name from URL path
	videoName := chi.URLParam(r, "videoName")
	if videoName == "" {
		writeError(w, http.StatusBadRequest, "videoName is required")
		return
	}

	// Extract category from query parameter
	category := r.URL.Query().Get("category")
	if category == "" {
		writeError(w, http.StatusBadRequest, "category query parameter is required")
		return
	}

	// Get manuscript content using our VideoService method
	manuscriptContent, err := s.videoService.GetVideoManuscript(videoName, category)
	if err != nil {
		// Check if it's a "video not found" error vs other errors
		if strings.Contains(err.Error(), "failed to get video") {
			writeError(w, http.StatusNotFound, "Video not found", err.Error())
			return
		}
		// Check if it's a "gist field empty" error (user error)
		if strings.Contains(err.Error(), "gist field is empty") {
			writeError(w, http.StatusBadRequest, "Video manuscript not configured", err.Error())
			return
		}
		// Other errors (file read errors, etc.) are server errors
		writeError(w, http.StatusInternalServerError, "Failed to read manuscript", err.Error())
		return
	}

	// Get AI configuration

	// Generate tags using AI
	tags, err := ai.SuggestTags(r.Context(), manuscriptContent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate tags", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AITagsResponse{Tags: []string{tags}})
}

// aiTweetsWithVideo handles POST /api/ai/tweets/{videoName}?category={category}
// This new endpoint uses URL parameters instead of JSON payload and reads manuscript from video files
func (s *Server) aiTweetsWithVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video name from URL path
	videoName := chi.URLParam(r, "videoName")
	if videoName == "" {
		writeError(w, http.StatusBadRequest, "videoName is required")
		return
	}

	// Extract category from query parameter
	category := r.URL.Query().Get("category")
	if category == "" {
		writeError(w, http.StatusBadRequest, "category query parameter is required")
		return
	}

	// Get manuscript content using our VideoService method
	manuscriptContent, err := s.videoService.GetVideoManuscript(videoName, category)
	if err != nil {
		// Check if it's a "video not found" error vs other errors
		if strings.Contains(err.Error(), "failed to get video") {
			writeError(w, http.StatusNotFound, "Video not found", err.Error())
			return
		}
		// Check if it's a "gist field empty" error (user error)
		if strings.Contains(err.Error(), "gist field is empty") {
			writeError(w, http.StatusBadRequest, "Video manuscript not configured", err.Error())
			return
		}
		// Other errors (file read errors, etc.) are server errors
		writeError(w, http.StatusInternalServerError, "Failed to read manuscript", err.Error())
		return
	}

	// Get AI configuration

	// Generate tweets using AI
	tweets, err := ai.SuggestTweets(r.Context(), manuscriptContent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate tweets", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, AITweetsResponse{Tweets: tweets})
}


// aiDescriptionTagsWithVideo handles POST /api/ai/description-tags/{videoName}?category={category}
// This new endpoint uses URL parameters instead of JSON payload and reads manuscript from video files
func (s *Server) aiDescriptionTagsWithVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video name from URL path
	videoName := chi.URLParam(r, "videoName")
	if videoName == "" {
		writeError(w, http.StatusBadRequest, "videoName is required")
		return
	}

	// Extract category from query parameter
	category := r.URL.Query().Get("category")
	if category == "" {
		writeError(w, http.StatusBadRequest, "category query parameter is required")
		return
	}

	// Get manuscript content using our VideoService method
	manuscriptContent, err := s.videoService.GetVideoManuscript(videoName, category)
	if err != nil {
		// Check if it's a "video not found" error vs other errors
		if strings.Contains(err.Error(), "failed to get video") {
			writeError(w, http.StatusNotFound, "Video not found", err.Error())
			return
		}
		// Check if it's a "gist field empty" error (user error)
		if strings.Contains(err.Error(), "gist field is empty") {
			writeError(w, http.StatusBadRequest, "Video manuscript not configured", err.Error())
			return
		}
		// Other errors (file read errors, etc.) are server errors
		writeError(w, http.StatusInternalServerError, "Failed to read manuscript", err.Error())
		return
	}

	// Get AI configuration

	// Generate description tags using AI
	descriptionTags, err := ai.SuggestDescriptionTags(r.Context(), manuscriptContent)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate description tags", err.Error())
		return
	}

	// Split the space-separated tags into a slice
	var descriptionTagsSlice []string
	if descriptionTags != "" {
		descriptionTagsSlice = strings.Fields(descriptionTags)
	}
	writeJSON(w, http.StatusOK, AIDescriptionTagsResponse{DescriptionTags: descriptionTagsSlice})
}

// animations handles GET /api/animations/{videoName}?category={category}
// This endpoint parses manuscript files to extract animations (no AI involved)
func (s *Server) animations(w http.ResponseWriter, r *http.Request) {
	// Extract video name from URL path
	videoName := chi.URLParam(r, "videoName")
	if videoName == "" {
		writeError(w, http.StatusBadRequest, "videoName is required")
		return
	}

	// Extract category from query parameter
	category := r.URL.Query().Get("category")
	if category == "" {
		writeError(w, http.StatusBadRequest, "category query parameter is required")
		return
	}

	// Get manuscript content using our VideoService method
	manuscriptContent, err := s.videoService.GetVideoManuscript(videoName, category)
	if err != nil {
		// Check if it's a "video not found" error vs other errors
		if strings.Contains(err.Error(), "failed to get video") {
			writeError(w, http.StatusNotFound, "Video not found", err.Error())
			return
		}
		// Check if it's a "gist field empty" error (user error)
		if strings.Contains(err.Error(), "gist field is empty") {
			writeError(w, http.StatusBadRequest, "Video manuscript not configured", err.Error())
			return
		}
		// Other errors (file read errors, etc.) are server errors
		writeError(w, http.StatusInternalServerError, "Failed to read manuscript", err.Error())
		return
	}

	// Create temporary file for manuscript content
	tempFile, err := os.CreateTemp("", "manuscript-*.md")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create temporary file", err.Error())
		return
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(manuscriptContent); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to write manuscript content", err.Error())
		return
	}
	tempFile.Close()

	// Use existing filesystem.Operations.GetAnimations() function
	fsOps := &filesystem.Operations{}
	animations, _, err := fsOps.GetAnimations(tempFile.Name())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to process animations", err.Error())
		return
	}

	// Convert the result to the expected format
	// The GetAnimations function returns []string directly, so we can use it as-is
	if animations == nil {
		animations = []string{}
	}

	writeJSON(w, http.StatusOK, AnimationsResponse{Animations: animations})
}
