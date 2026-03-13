package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/thumbnail"

	"github.com/go-chi/chi/v5"
)

// --- Response types ---

type PublishYouTubeResponse struct {
	VideoID     string `json:"videoId"`
	SyncWarning string `json:"syncWarning,omitempty"`
}

type PublishThumbnailResponse struct {
	Success     bool   `json:"success"`
	SyncWarning string `json:"syncWarning,omitempty"`
}

type PublishShortResponse struct {
	YouTubeID   string `json:"youtubeId"`
	SyncWarning string `json:"syncWarning,omitempty"`
}

type PublishHugoResponse struct {
	HugoPath    string `json:"hugoPath"`
	SyncWarning string `json:"syncWarning,omitempty"`
}

type TranscriptResponse struct {
	Transcript string `json:"transcript"`
}

type MetadataResponse struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	PublishedAt string   `json:"publishedAt"`
}

type SocialPostResponse struct {
	Posted      bool   `json:"posted"`
	Message     string `json:"message,omitempty"`
	PostURL     string `json:"postUrl,omitempty"`
	SyncWarning string `json:"syncWarning,omitempty"`
}

// --- Handlers ---

// handlePublishYouTube uploads video + thumbnail to YouTube.
// POST /api/publish/youtube/{videoName}?category=X
func (s *Server) handlePublishYouTube(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
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

	// Resolve video file: if Drive-hosted, download to temp file
	uploadPath := video.UploadVideo
	if video.VideoDriveFileID != "" && s.driveService != nil {
		content, _, filename, err := s.driveService.GetFile(r.Context(), video.VideoDriveFileID)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to download video from Drive", err.Error())
			return
		}
		defer content.Close()

		tmpFile, err := createTempFromReader(content, filename)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to create temp video file", err.Error())
			return
		}
		defer os.Remove(tmpFile)
		uploadPath = tmpFile
		video.UploadVideo = uploadPath
	}

	if uploadPath == "" {
		respondError(w, http.StatusBadRequest, "No video file available", "Upload a video file first")
		return
	}

	// Resolve thumbnail for the upload check
	ref, refErr := thumbnail.ResolveThumbnail(&video)
	if refErr != nil {
		video.Thumbnail = "pending"
	} else if ref.Path != "" {
		video.Thumbnail = ref.Path
	} else if ref.DriveFileID != "" {
		video.Thumbnail = "drive://" + ref.DriveFileID
	}

	videoID, err := s.publishingService.UploadVideo(r.Context(), &video)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "YouTube upload failed", err.Error())
		return
	}

	video.VideoId = videoID

	// Upload thumbnail if available
	if refErr == nil {
		tnErr := thumbnail.WithThumbnailFile(r.Context(), ref, s.driveService, func(path string) error {
			return s.publishingService.UploadThumbnail(r.Context(), videoID, path)
		})
		if tnErr != nil {
			fmt.Printf("Warning: thumbnail upload failed: %v\n", tnErr)
		}
	}

	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := PublishYouTubeResponse{VideoID: videoID}
	addSyncWarningStr(&resp.SyncWarning, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// handlePublishThumbnail re-uploads a thumbnail for an existing YouTube video.
// POST /api/publish/youtube/{videoName}/thumbnail?category=X
func (s *Server) handlePublishThumbnail(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "")
		return
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	if video.VideoId == "" {
		respondError(w, http.StatusBadRequest, "Video has no YouTube ID", "Publish to YouTube first")
		return
	}

	ref, err := thumbnail.ResolveThumbnail(&video)
	if err != nil {
		respondError(w, http.StatusBadRequest, "No thumbnail available", err.Error())
		return
	}

	tnErr := thumbnail.WithThumbnailFile(r.Context(), ref, s.driveService, func(path string) error {
		return s.publishingService.UploadThumbnail(r.Context(), video.VideoId, path)
	})
	if tnErr != nil {
		respondError(w, http.StatusInternalServerError, "Thumbnail upload failed", tnErr.Error())
		return
	}

	respondJSON(w, http.StatusOK, PublishThumbnailResponse{Success: true})
}

// handlePublishShort uploads a short to YouTube.
// POST /api/publish/youtube/{videoName}/shorts/{shortId}?category=X
func (s *Server) handlePublishShort(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	videoName := chi.URLParam(r, "videoName")
	shortID := chi.URLParam(r, "shortId")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "")
		return
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

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

	short := video.Shorts[shortIdx]
	if short.FilePath == "" {
		respondError(w, http.StatusBadRequest, "Short has no file path", "")
		return
	}

	ytID, err := s.publishingService.UploadShort(r.Context(), short.FilePath, short, video.VideoId)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Short upload failed", err.Error())
		return
	}

	video.Shorts[shortIdx].YouTubeID = ytID
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := PublishShortResponse{YouTubeID: ytID}
	addSyncWarningStr(&resp.SyncWarning, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// handlePublishHugo creates a Hugo blog post.
// POST /api/publish/hugo/{videoName}?category=X
func (s *Server) handlePublishHugo(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "")
		return
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	title := video.GetUploadTitle()
	if title == "" {
		respondError(w, http.StatusBadRequest, "Video has no title", "")
		return
	}

	hugoOpts := &publishing.HugoPostOptions{
		DriveService:  s.driveService,
		DriveFolderID: s.driveFolderID,
	}
	hugoPath, err := s.publishingService.CreateHugoPost(r.Context(), &video, hugoOpts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Hugo post creation failed", err.Error())
		return
	}

	video.HugoPath = hugoPath
	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	resp := PublishHugoResponse{HugoPath: hugoPath}
	addSyncWarningStr(&resp.SyncWarning, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// handleGetTranscript fetches captions for a YouTube video.
// GET /api/publish/transcript/{videoId}
func (s *Server) handleGetTranscript(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	videoID := chi.URLParam(r, "videoId")
	if videoID == "" {
		respondError(w, http.StatusBadRequest, "Missing videoId", "")
		return
	}

	transcript, err := s.publishingService.GetTranscript(r.Context(), videoID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch transcript", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, TranscriptResponse{Transcript: transcript})
}

// handleGetMetadata fetches YouTube video metadata.
// GET /api/publish/metadata/{videoId}
func (s *Server) handleGetMetadata(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	videoID := chi.URLParam(r, "videoId")
	if videoID == "" {
		respondError(w, http.StatusBadRequest, "Missing videoId", "")
		return
	}

	meta, err := s.publishingService.GetVideoMetadata(r.Context(), videoID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch metadata", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, MetadataResponse{
		Title:       meta.Title,
		Description: meta.Description,
		Tags:        meta.Tags,
		PublishedAt: meta.PublishedAt,
	})
}

// handleSocialPost posts to a social media platform or returns text for copy-paste.
// POST /api/social/{platform}/{videoName}?category=X
func (s *Server) handleSocialPost(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	platform := chi.URLParam(r, "platform")
	videoName := chi.URLParam(r, "videoName")
	category := r.URL.Query().Get("category")
	if category == "" {
		respondError(w, http.StatusBadRequest, "Missing category", "")
		return
	}

	video, err := s.videoService.GetVideo(videoName, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "Video not found", err.Error())
		return
	}

	resp := SocialPostResponse{}

	switch platform {
	case "bluesky":
		if video.Tweet == "" {
			respondError(w, http.StatusBadRequest, "No tweet text configured", "Set tweet field first")
			return
		}
		if video.VideoId == "" {
			respondError(w, http.StatusBadRequest, "No YouTube video ID", "Publish to YouTube first")
			return
		}

		thumbnailPath := ""
		ref, refErr := thumbnail.ResolveThumbnail(&video)
		if refErr == nil {
			_ = thumbnail.WithThumbnailFile(r.Context(), ref, s.driveService, func(path string) error {
				thumbnailPath = path
				return nil
			})
		}

		err = s.publishingService.PostBlueSky(r.Context(), video.Tweet, video.VideoId, thumbnailPath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "BlueSky post failed", err.Error())
			return
		}
		video.BlueSkyPosted = true
		resp.Posted = true

	case "slack":
		if video.VideoId == "" {
			respondError(w, http.StatusBadRequest, "No YouTube video ID", "Publish to YouTube first")
			return
		}

		videoPath := ""
		if s.filesystem != nil {
			videoPath = s.filesystem.GetFilePath(video.Category, video.Name, "yaml")
		}

		err = s.publishingService.PostSlack(r.Context(), &video, videoPath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "Slack post failed", err.Error())
			return
		}
		video.SlackPosted = true
		resp.Posted = true

	case "linkedin":
		resp.Posted = false
		resp.Message = formatLinkedInMessage(&video)
		video.LinkedInPosted = true

	case "hackernews":
		resp.Posted = false
		resp.Message = formatHNMessage(&video)
		video.HNPosted = true

	case "dot":
		resp.Posted = false
		resp.Message = formatDOTMessage(&video)
		video.DOTPosted = true

	default:
		respondError(w, http.StatusBadRequest, "Unknown platform", fmt.Sprintf("platform %q not supported", platform))
		return
	}

	if err := s.videoService.UpdateVideo(video); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save video", err.Error())
		return
	}

	addSyncWarningStr(&resp.SyncWarning, s.videoService)
	respondJSON(w, http.StatusOK, resp)
}

// --- Social message formatters ---

func formatLinkedInMessage(video *storage.Video) string {
	title := video.GetUploadTitle()
	youtubeURL := ""
	if video.VideoId != "" {
		youtubeURL = publishing.GetYouTubeURL(video.VideoId)
	}
	hugoURL := ""
	if title != "" && video.Gist != "" {
		cat := publishing.GetCategoryFromFilePath(video.Gist)
		hugoURL = publishing.ConstructHugoURL(title, cat)
	}

	var b strings.Builder
	b.WriteString(title + "\n\n")
	if video.Tweet != "" {
		text := strings.ReplaceAll(video.Tweet, "[YOUTUBE]", youtubeURL)
		b.WriteString(text + "\n\n")
	}
	if youtubeURL != "" {
		b.WriteString(youtubeURL + "\n")
	}
	if hugoURL != "" {
		b.WriteString(hugoURL + "\n")
	}
	if video.DescriptionTags != "" {
		b.WriteString("\n" + video.DescriptionTags)
	}
	return b.String()
}

func formatHNMessage(video *storage.Video) string {
	title := video.GetUploadTitle()
	hugoURL := ""
	if title != "" && video.Gist != "" {
		cat := publishing.GetCategoryFromFilePath(video.Gist)
		hugoURL = publishing.ConstructHugoURL(title, cat)
	}
	url := hugoURL
	if url == "" && video.VideoId != "" {
		url = publishing.GetYouTubeURL(video.VideoId)
	}
	return fmt.Sprintf("Title: %s\nURL: %s", title, url)
}

func formatDOTMessage(video *storage.Video) string {
	title := video.GetUploadTitle()
	youtubeURL := ""
	if video.VideoId != "" {
		youtubeURL = publishing.GetYouTubeURL(video.VideoId)
	}
	hugoURL := ""
	if title != "" && video.Gist != "" {
		cat := publishing.GetCategoryFromFilePath(video.Gist)
		hugoURL = publishing.ConstructHugoURL(title, cat)
	}

	var b strings.Builder
	b.WriteString(title + "\n\n")
	if video.Description != "" {
		b.WriteString(video.Description + "\n\n")
	}
	if youtubeURL != "" {
		b.WriteString("YouTube: " + youtubeURL + "\n")
	}
	if hugoURL != "" {
		b.WriteString("Blog: " + hugoURL + "\n")
	}
	return b.String()
}

// --- AMA Handlers ---

// AMAApplyRequest is the request body for applying AMA content to YouTube.
type AMAApplyRequest struct {
	VideoID     string `json:"videoId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Timecodes   string `json:"timecodes"`
}

// AMAApplyResponse is the response for the AMA apply endpoint.
type AMAApplyResponse struct {
	Success bool `json:"success"`
}

// AMAGenerateRequest is the request body for generating AMA content from a YouTube video ID.
type AMAGenerateRequest struct {
	VideoID string `json:"videoId"`
}

// AMAGenerateResponse is the response for the AMA generate endpoint.
type AMAGenerateResponse struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	Timecodes   string `json:"timecodes"`
	Transcript  string `json:"transcript"`
}

// handleAMAGenerate fetches a YouTube video's transcript and generates AMA content.
// POST /api/ama/generate
func (s *Server) handleAMAGenerate(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	var req AMAGenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if req.VideoID == "" {
		respondError(w, http.StatusBadRequest, "videoId is required", "")
		return
	}

	transcript, err := s.publishingService.GetTranscript(r.Context(), req.VideoID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch transcript", err.Error())
		return
	}

	content, err := s.aiService.GenerateAMAContent(r.Context(), transcript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, AMAGenerateResponse{
		Title:       content.Title,
		Description: content.Description,
		Tags:        content.Tags,
		Timecodes:   content.Timecodes,
		Transcript:  transcript,
	})
}

// handleAMAApply applies AMA content to a YouTube video.
// POST /api/ama/apply
func (s *Server) handleAMAApply(w http.ResponseWriter, r *http.Request) {
	if s.publishingService == nil {
		respondError(w, http.StatusNotImplemented, "Publishing not configured", "")
		return
	}

	var req AMAApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	if req.VideoID == "" {
		respondError(w, http.StatusBadRequest, "videoId is required", "")
		return
	}

	err := s.publishingService.UpdateAMAVideo(r.Context(), req.VideoID, req.Title, req.Description, req.Tags, req.Timecodes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update YouTube video", err.Error())
		return
	}

	respondJSON(w, http.StatusOK, AMAApplyResponse{Success: true})
}

// --- Helpers ---

// addSyncWarningStr populates a sync warning string pointer from service state.
func addSyncWarningStr(target *string, vs videoServiceSyncer) {
	if syncErr := vs.LastSyncError(); syncErr != nil {
		*target = "git sync failed: " + syncErr.Error()
	} else if !vs.IsSyncConfigured() {
		*target = "git sync not configured — changes saved locally only"
	}
}

// videoServiceSyncer is an interface for sync status checking.
type videoServiceSyncer interface {
	LastSyncError() error
	IsSyncConfigured() bool
}

func createTempFromReader(r io.Reader, filename string) (string, error) {
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".mp4"
	}

	tmpFile, err := os.CreateTemp("", "video-*"+ext)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(tmpFile, r); err != nil {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
		return "", err
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}
