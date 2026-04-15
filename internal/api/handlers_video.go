package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
	"devopstoolkit/youtube-automation/pkg/utils"

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
	SyncWarning string       `json:"syncWarning,omitempty"`
	PullWarning string       `json:"pullWarning,omitempty"`
}

// VideoListItem is a lightweight representation of a video.
type VideoListItem struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Category    string       `json:"category"`
	Date        string       `json:"date,omitempty"`
	Title       string       `json:"title,omitempty"`
	Phase       int          `json:"phase"`
	Progress    ProgressInfo `json:"progress"`
	Sponsored   bool         `json:"sponsored"`
	IsFarFuture bool         `json:"isFarFuture"`
}

// isSponsored returns true if the video has a non-trivial sponsorship amount and is not blocked.
func isSponsored(v storage.Video) bool {
	amount := v.Sponsorship.Amount
	return amount != "" && amount != "-" && amount != "N/A" && v.Sponsorship.Blocked == ""
}

// isFarFuture returns true if the video's date is more than 3 months in the future
// and the video is in the Started phase.
func isFarFuture(v storage.Video, phase int) bool {
	if phase != workflow.PhaseStarted || v.Date == "" {
		return false
	}
	farFuture, err := utils.IsFarFutureDate(v.Date, "2006-01-02T15:04", time.Now())
	if err != nil {
		return false
	}
	return farFuture
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
		p := video.CalculateVideoPhase(v)
		items = append(items, VideoListItem{
			ID:          v.Category + "/" + v.Name,
			Name:        v.Name,
			Category:    v.Category,
			Date:        v.Date,
			Title:       title,
			Phase:       p,
			Progress:    ProgressInfo{Completed: overallC, Total: overallT},
			Sponsored:   isSponsored(v),
			IsFarFuture: isFarFuture(v, p),
		})
	}
	respondJSON(w, http.StatusOK, items)
}

// handleSearchVideos searches across all videos by a query string.
func (s *Server) handleSearchVideos(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		respondJSON(w, http.StatusOK, []VideoListItem{})
		return
	}

	videos, err := s.videoService.SearchVideos(q)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "search failed", err.Error())
		return
	}

	items := make([]VideoListItem, 0, len(videos))
	for _, v := range videos {
		title := v.GetUploadTitle()
		overallC, overallT := s.videoManager.CalculateOverallProgress(v)
		p := video.CalculateVideoPhase(v)
		items = append(items, VideoListItem{
			ID:          v.Category + "/" + v.Name,
			Name:        v.Name,
			Category:    v.Category,
			Date:        v.Date,
			Title:       title,
			Phase:       p,
			Progress:    ProgressInfo{Completed: overallC, Total: overallT},
			Sponsored:   isSponsored(v),
			IsFarFuture: isFarFuture(v, p),
		})
	}
	respondJSON(w, http.StatusOK, items)
}

// pullOnReadThrottle is the minimum interval between pull-on-read attempts
// when loading a video detail page. Multiple detail loads within this window
// reuse the local working copy without hitting the remote.
const pullOnReadThrottle = 10 * time.Second

// handleGetVideo returns a single video by name and category.
//
// Before reading from disk, the handler attempts a throttled `git pull --rebase`
// so the user sees changes pushed externally to the data repo. Pull failures
// are surfaced via the response's pullWarning field — the local copy is still
// returned so the page remains usable.
func (s *Server) handleGetVideo(w http.ResponseWriter, r *http.Request) {
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "missing required query parameter: category", "")
		return
	}

	var pullWarning string
	if s.gitSync != nil {
		if err := s.gitSync.PullIfStale(pullOnReadThrottle); err != nil {
			pullWarning = "git pull failed: " + err.Error()
		}
	}

	v, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "video not found", err.Error())
		return
	}

	resp := s.enrichVideo(v)
	resp.PullWarning = pullWarning
	respondJSON(w, http.StatusOK, resp)
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
