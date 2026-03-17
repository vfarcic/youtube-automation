package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/go-chi/chi/v5"
)

// emailNotConfiguredMessage returns a user-friendly message explaining why email was not sent.
func emailNotConfiguredMessage(svc EmailService, settings *configuration.SettingsEmail, recipientField string) string {
	if svc == nil {
		return "Email not configured: EMAIL_PASSWORD is not set"
	}
	if settings == nil {
		return "Email not configured: email settings are missing"
	}
	if settings.From == "" {
		return "Email not configured: EMAIL_FROM (or email.from in settings.yaml) is not set"
	}
	return fmt.Sprintf("Email not configured: %s address is not set (use EMAIL_%s env var or email settings in settings.yaml)",
		recipientField, recipientField)
}

// EmailService abstracts email sending for action endpoints.
type EmailService interface {
	SendThumbnail(from, to string, video storage.Video) error
	SendEdit(from, to string, video storage.Video) error
	SendSponsors(from, to string, videoID, sponsorshipPrice, videoTitle string) error
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
	} else {
		resp.EmailError = emailNotConfiguredMessage(s.emailService, s.emailSettings, "ThumbnailTo")
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

	video.RequestEdit = true
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := ActionResponse{
		Video: s.enrichVideo(video),
	}

	// Resolve the gist path for the email attachment only — don't persist the resolved path
	emailVideo := video
	if emailVideo.Gist != "" && s.filesystem != nil {
		emailVideo.Gist = s.filesystem.ResolvePath(emailVideo.Gist)
	}

	// Resolve and read sponsor ad file content for the email
	if emailVideo.Sponsorship.AdFile != "" && s.filesystem != nil {
		adPath := s.filesystem.ResolvePath(filepath.Join("manuscript", "ads", emailVideo.Sponsorship.AdFile))
		adBytes, err := os.ReadFile(adPath)
		if err != nil {
			emailVideo.AdContent = fmt.Sprintf("[Warning: Ad file '%s' could not be read: %s]", emailVideo.Sponsorship.AdFile, err.Error())
		} else {
			emailVideo.AdContent = string(adBytes)
		}
	}

	if s.emailService != nil && s.emailSettings != nil && s.emailSettings.From != "" && s.emailSettings.EditTo != "" {
		if err := s.emailService.SendEdit(s.emailSettings.From, s.emailSettings.EditTo, emailVideo); err != nil {
			resp.EmailError = err.Error()
		} else {
			resp.EmailSent = true
		}
	} else {
		resp.EmailError = emailNotConfiguredMessage(s.emailService, s.emailSettings, "EditTo")
	}

	if syncErr := s.videoService.LastSyncError(); syncErr != nil {
		resp.SyncWarning = "git sync failed: " + syncErr.Error()
	} else if !s.videoService.IsSyncConfigured() {
		resp.SyncWarning = "git sync not configured — changes saved locally only"
	}

	respondJSON(w, http.StatusOK, resp)
}

// handleNotifySponsors handles POST /api/actions/notify-sponsors/{videoName}?category=X
func (s *Server) handleNotifySponsors(w http.ResponseWriter, r *http.Request) {
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

	if video.NotifiedSponsors {
		respondJSON(w, http.StatusOK, ActionResponse{
			AlreadyRequested: true,
			Video:            s.enrichVideo(video),
		})
		return
	}

	// Validate sponsorship exists
	amount := video.Sponsorship.Amount
	if amount == "" || amount == "N/A" || amount == "-" {
		respondError(w, http.StatusBadRequest, "No sponsorship", "Video has no valid sponsorship amount")
		return
	}
	if video.Sponsorship.Emails == "" {
		respondError(w, http.StatusBadRequest, "No sponsor emails", "Video has no sponsor email addresses configured")
		return
	}

	resp := ActionResponse{
		Video: s.enrichVideo(video),
	}

	if s.emailService == nil || s.emailSettings == nil || s.emailSettings.From == "" {
		resp.EmailError = emailNotConfiguredMessage(s.emailService, s.emailSettings, "From")
		respondJSON(w, http.StatusOK, resp)
		return
	}

	if err := s.emailService.SendSponsors(s.emailSettings.From, video.Sponsorship.Emails, video.VideoId, video.Sponsorship.Amount, video.GetUploadTitle()); err != nil {
		resp.EmailError = err.Error()
		respondJSON(w, http.StatusOK, resp)
		return
	}

	resp.EmailSent = true
	video.NotifiedSponsors = true
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}
	resp.Video = s.enrichVideo(video)

	if syncErr := s.videoService.LastSyncError(); syncErr != nil {
		resp.SyncWarning = "git sync failed: " + syncErr.Error()
	} else if !s.videoService.IsSyncConfigured() {
		resp.SyncWarning = "git sync not configured — changes saved locally only"
	}

	respondJSON(w, http.StatusOK, resp)
}
