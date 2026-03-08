package api

import (
	"net/http"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/go-chi/chi/v5"
)

// EmailService abstracts email sending for action endpoints.
type EmailService interface {
	SendThumbnail(from, to string, video storage.Video) error
	SendEdit(from, to string, video storage.Video) error
}

// ActionResponse is the JSON response for action endpoints.
type ActionResponse struct {
	AlreadyRequested bool          `json:"alreadyRequested"`
	EmailSent        bool          `json:"emailSent"`
	EmailError       string        `json:"emailError,omitempty"`
	Video            VideoResponse `json:"video"`
	SyncWarning      string        `json:"syncWarning,omitempty"`
}

// SetEmailService configures email sending for action endpoints.
func (s *Server) SetEmailService(es EmailService, settings *configuration.SettingsEmail) {
	s.emailService = es
	s.emailSettings = settings
}

// handleRequestThumbnail handles POST /api/actions/request-thumbnail/{videoName}?category=X
func (s *Server) handleRequestThumbnail(w http.ResponseWriter, r *http.Request) {
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

	if video.RequestThumbnail {
		respondJSON(w, http.StatusOK, ActionResponse{
			AlreadyRequested: true,
			Video:            s.enrichVideo(video),
		})
		return
	}

	video.RequestThumbnail = true
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := ActionResponse{
		Video: s.enrichVideo(video),
	}

	if s.emailService != nil && s.emailSettings != nil && s.emailSettings.From != "" && s.emailSettings.ThumbnailTo != "" {
		if err := s.emailService.SendThumbnail(s.emailSettings.From, s.emailSettings.ThumbnailTo, video); err != nil {
			resp.EmailError = err.Error()
		} else {
			resp.EmailSent = true
		}
	}

	if syncErr := s.videoService.LastSyncError(); syncErr != nil {
		resp.SyncWarning = "git sync failed: " + syncErr.Error()
	} else if !s.videoService.IsSyncConfigured() {
		resp.SyncWarning = "git sync not configured — changes saved locally only"
	}

	respondJSON(w, http.StatusOK, resp)
}

// handleRequestEdit handles POST /api/actions/request-edit/{videoName}?category=X
func (s *Server) handleRequestEdit(w http.ResponseWriter, r *http.Request) {
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

	if video.RequestEdit {
		respondJSON(w, http.StatusOK, ActionResponse{
			AlreadyRequested: true,
			Video:            s.enrichVideo(video),
		})
		return
	}

	// Resolve the gist path so the email attachment works
	if video.Gist != "" && s.filesystem != nil {
		video.Gist = s.filesystem.ResolvePath(video.Gist)
	}

	video.RequestEdit = true
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := ActionResponse{
		Video: s.enrichVideo(video),
	}

	if s.emailService != nil && s.emailSettings != nil && s.emailSettings.From != "" && s.emailSettings.EditTo != "" {
		if err := s.emailService.SendEdit(s.emailSettings.From, s.emailSettings.EditTo, video); err != nil {
			resp.EmailError = err.Error()
		} else {
			resp.EmailSent = true
		}
	}

	if syncErr := s.videoService.LastSyncError(); syncErr != nil {
		resp.SyncWarning = "git sync failed: " + syncErr.Error()
	} else if !s.videoService.IsSyncConfigured() {
		resp.SyncWarning = "git sync not configured — changes saved locally only"
	}

	respondJSON(w, http.StatusOK, resp)
}
