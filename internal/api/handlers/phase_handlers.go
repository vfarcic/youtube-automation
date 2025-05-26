package handlers

import (
	"encoding/json"
	"net/http"

	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/go-chi/chi/v5"
)

// PhaseHandlers contains all handlers for phase-specific video updates
type PhaseHandlers struct {
	videoService *service.VideoService
}

// NewPhaseHandlers creates a new PhaseHandlers instance
func NewPhaseHandlers(videoService *service.VideoService) *PhaseHandlers {
	return &PhaseHandlers{
		videoService: videoService,
	}
}

// UpdateInitialPhase updates the initial phase details of a video
type InitialPhaseRequest struct {
	ProjectName          string `json:"projectName"`
	ProjectURL           string `json:"projectURL"`
	SponsorshipAmount    string `json:"sponsorshipAmount"`
	SponsorshipEmails    string `json:"sponsorshipEmails"`
	SponsorshipBlocked   string `json:"sponsorshipBlockedReason"`
	PublishDate          string `json:"publishDate"`
	Delayed              bool   `json:"delayed"`
	GistPath             string `json:"gistPath"`
}

// UpdateInitialPhase updates the initial phase details of a video
func (h *PhaseHandlers) UpdateInitialPhase(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	// Get the current video
	video, err := h.videoService.GetVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode the request
	var req InitialPhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update the video with the new data
	video.ProjectName = req.ProjectName
	video.ProjectURL = req.ProjectURL
	video.Sponsorship.Amount = req.SponsorshipAmount
	video.Sponsorship.Emails = req.SponsorshipEmails
	video.Sponsorship.BlockedReason = req.SponsorshipBlocked
	video.Date = req.PublishDate
	video.Delayed = req.Delayed
	video.Gist = req.GistPath
	
	// Save the updated video
	updatedVideo, err := h.videoService.UpdateVideo(videoID, video)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	respondJSON(w, updatedVideo)
}

type WorkPhaseRequest struct {
	CodeDone          bool   `json:"codeDone"`
	TalkingHeadDone   bool   `json:"talkingHeadDone"`
	ScreenRecordingDone bool `json:"screenRecordingDone"`
	RelatedVideos     string `json:"relatedVideos"`
	ThumbnailsDone    bool   `json:"thumbnailsDone"`
	DiagramsDone      bool   `json:"diagramsDone"`
	ScreenshotsDone   bool   `json:"screenshotsDone"`
	FilesLocation     string `json:"filesLocation"`
	Tagline           string `json:"tagline"`
	TaglineIdeas      string `json:"taglineIdeas"`
	OtherLogosAssets  string `json:"otherLogosAssets"`
}

// UpdateWorkPhase updates the work phase details of a video
func (h *PhaseHandlers) UpdateWorkPhase(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	// Get the current video
	video, err := h.videoService.GetVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode the request
	var req WorkPhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update the video with the new data
	video.Code = req.CodeDone
	video.Head = req.TalkingHeadDone
	video.Screen = req.ScreenRecordingDone
	video.RelatedVideos = req.RelatedVideos
	video.Thumbnails = req.ThumbnailsDone
	video.Diagrams = req.DiagramsDone
	video.Screenshots = req.ScreenshotsDone
	video.Location = req.FilesLocation
	video.Tagline = req.Tagline
	video.TaglineIdeas = req.TaglineIdeas
	video.OtherLogos = req.OtherLogosAssets
	
	// Save the updated video
	updatedVideo, err := h.videoService.UpdateVideo(videoID, video)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	respondJSON(w, updatedVideo)
}

type DefinitionPhaseRequest struct {
	Title                     string `json:"title"`
	Description               string `json:"description"`
	Highlight                 string `json:"highlight"`
	Tags                      string `json:"tags"`
	DescriptionTags           string `json:"descriptionTags"`
	TweetText                 string `json:"tweetText"`
	AnimationsScript          string `json:"animationsScript"`
	RequestThumbnailGeneration bool   `json:"requestThumbnailGeneration"`
}

// UpdateDefinitionPhase updates the definition phase details of a video
func (h *PhaseHandlers) UpdateDefinitionPhase(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	// Get the current video
	video, err := h.videoService.GetVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode the request
	var req DefinitionPhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update the video with the new data
	video.Title = req.Title
	video.Description = req.Description
	video.Highlight = req.Highlight
	video.Tags = req.Tags
	video.DescriptionTags = req.DescriptionTags
	video.Tweet = req.TweetText
	video.Animations = req.AnimationsScript
	video.RequestThumbnail = req.RequestThumbnailGeneration
	
	// Save the updated video
	updatedVideo, err := h.videoService.UpdateVideo(videoID, video)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	respondJSON(w, updatedVideo)
}

type PostProductionPhaseRequest struct {
	ThumbnailPath string `json:"thumbnailPath"`
	Members       string `json:"members"`
	RequestEdit   bool   `json:"requestEdit"`
	Timecodes     string `json:"timecodes"`
	MovieDone     bool   `json:"movieDone"`
	SlidesDone    bool   `json:"slidesDone"`
}

// UpdatePostProductionPhase updates the post-production phase details of a video
func (h *PhaseHandlers) UpdatePostProductionPhase(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	// Get the current video
	video, err := h.videoService.GetVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode the request
	var req PostProductionPhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update the video with the new data
	video.Thumbnail = req.ThumbnailPath
	video.Members = req.Members
	video.RequestEdit = req.RequestEdit
	video.Timecodes = req.Timecodes
	video.Movie = req.MovieDone
	video.Slides = req.SlidesDone
	
	// Save the updated video
	updatedVideo, err := h.videoService.UpdateVideo(videoID, video)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	respondJSON(w, updatedVideo)
}

type PublishingPhaseRequest struct {
	VideoFilePath   string `json:"videoFilePath"`
	UploadToYouTube bool   `json:"uploadToYouTube"`
	CreateHugoPost  bool   `json:"createHugoPost"`
}

// UpdatePublishingPhase updates the publishing phase details of a video
func (h *PhaseHandlers) UpdatePublishingPhase(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	// Get the current video
	video, err := h.videoService.GetVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode the request
	var req PublishingPhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update the video with the new data
	video.UploadVideo = req.VideoFilePath
	
	// Handle YouTube upload if requested
	if req.UploadToYouTube {
		// This would be handled by a service call in a real implementation
		// For now, just update the video state
		video.VideoId = "youtube-id-placeholder"
	}
	
	// Handle Hugo post creation if requested
	if req.CreateHugoPost {
		// This would be handled by a service call in a real implementation
		// For now, just update the video state
		video.HugoPath = "hugo-path-placeholder"
	}
	
	// Save the updated video
	updatedVideo, err := h.videoService.UpdateVideo(videoID, video)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	respondJSON(w, updatedVideo)
}

type PostPublishPhaseRequest struct {
	BlueSkyPostSent          bool   `json:"blueSkyPostSent"`
	LinkedInPostSent         bool   `json:"linkedInPostSent"`
	SlackPostSent            bool   `json:"slackPostSent"`
	YouTubeHighlightCreated  bool   `json:"youTubeHighlightCreated"`
	YouTubePinnedCommentAdded bool  `json:"youTubePinnedCommentAdded"`
	RepliedToYouTubeComments bool   `json:"repliedToYouTubeComments"`
	GdeAdvocuPostSent        bool   `json:"gdeAdvocuPostSent"`
	CodeRepositoryURL        string `json:"codeRepositoryURL"`
	NotifiedSponsors         bool   `json:"notifiedSponsors"`
}

// UpdatePostPublishPhase updates the post-publish phase details of a video
func (h *PhaseHandlers) UpdatePostPublishPhase(w http.ResponseWriter, r *http.Request) {
	videoID := chi.URLParam(r, "video_id")
	
	// Get the current video
	video, err := h.videoService.GetVideo(videoID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			http.Error(w, "video not found", http.StatusNotFound)
		} else if err == service.ErrInvalidRequest {
			http.Error(w, "invalid video ID format", http.StatusBadRequest)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	
	// Decode the request
	var req PostPublishPhaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update the video with the new data
	video.BlueSkyPosted = req.BlueSkyPostSent
	video.LinkedInPosted = req.LinkedInPostSent
	video.SlackPosted = req.SlackPostSent
	video.YouTubeHighlight = req.YouTubeHighlightCreated
	video.YouTubeComment = req.YouTubePinnedCommentAdded
	video.YouTubeCommentReply = req.RepliedToYouTubeComments
	video.GDE = req.GdeAdvocuPostSent
	video.Repo = req.CodeRepositoryURL
	video.NotifiedSponsors = req.NotifiedSponsors
	
	// Save the updated video
	updatedVideo, err := h.videoService.UpdateVideo(videoID, video)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	respondJSON(w, updatedVideo)
}