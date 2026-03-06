package api

import (
	"encoding/json"
	"net/http"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/thumbnail"

	"github.com/go-chi/chi/v5"
)

// --- Response types ---

type AITitlesResponse struct {
	Titles []string `json:"titles"`
}

type AIDescriptionResponse struct {
	Description string `json:"description"`
}

type AITagsResponse struct {
	Tags string `json:"tags"`
}

type AITweetsResponse struct {
	Tweets []string `json:"tweets"`
}

type AIDescriptionTagsResponse struct {
	DescriptionTags string `json:"descriptionTags"`
}

type AIShortsResponse struct {
	Candidates []ai.ShortCandidate `json:"candidates"`
}

type AIThumbnailsResponse struct {
	Subtle string `json:"subtle"`
	Bold   string `json:"bold"`
}

type AITranslateResponse struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        string   `json:"tags"`
	Timecodes   string   `json:"timecodes"`
	ShortTitles []string `json:"shortTitles,omitempty"`
}

type AIAMAContentResponse struct {
	Title       string `json:"title"`
	Timecodes   string `json:"timecodes"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
}

type AIAMATitleResponse struct {
	Title string `json:"title"`
}

type AIAMADescriptionResponse struct {
	Description string `json:"description"`
}

type AIAMATimecodesResponse struct {
	Timecodes string `json:"timecodes"`
}

// --- Request types (body-based endpoints) ---

type AIThumbnailsRequest struct {
	Category    string `json:"category"`
	Name        string `json:"name"`
	ImagePath   string `json:"imagePath"`
	DriveFileID string `json:"driveFileId"`
}

type AITranslateRequest struct {
	Category       string `json:"category"`
	Name           string `json:"name"`
	TargetLanguage string `json:"targetLanguage"`
}

type AIAMARequest struct {
	Category string `json:"category"`
	Name     string `json:"name"`
}

// --- Manuscript-based handlers (6 endpoints) ---

// handleAITitles generates title suggestions from the video manuscript.
func (s *Server) handleAITitles(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getManuscriptFromPath(w, r)
	if !ok {
		return
	}
	titles, err := s.aiService.SuggestTitles(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AITitlesResponse{Titles: titles})
}

// handleAIDescription generates a description from the video manuscript.
func (s *Server) handleAIDescription(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getManuscriptFromPath(w, r)
	if !ok {
		return
	}
	desc, err := s.aiService.SuggestDescription(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIDescriptionResponse{Description: desc})
}

// handleAITags generates tags from the video manuscript.
func (s *Server) handleAITags(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getManuscriptFromPath(w, r)
	if !ok {
		return
	}
	tags, err := s.aiService.SuggestTags(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AITagsResponse{Tags: tags})
}

// handleAITweets generates tweet suggestions from the video manuscript.
func (s *Server) handleAITweets(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getManuscriptFromPath(w, r)
	if !ok {
		return
	}
	tweets, err := s.aiService.SuggestTweets(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AITweetsResponse{Tweets: tweets})
}

// handleAIDescriptionTags generates hashtag suggestions from the video manuscript.
func (s *Server) handleAIDescriptionTags(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getManuscriptFromPath(w, r)
	if !ok {
		return
	}
	tags, err := s.aiService.SuggestDescriptionTags(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIDescriptionTagsResponse{DescriptionTags: tags})
}

// handleAIShorts analyzes the manuscript for potential YouTube Shorts.
func (s *Server) handleAIShorts(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getManuscriptFromPath(w, r)
	if !ok {
		return
	}
	candidates, err := s.aiService.AnalyzeShorts(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIShortsResponse{Candidates: candidates})
}

// --- Body-based handlers (6 endpoints) ---

// handleAIThumbnails generates thumbnail variation prompts for an image.
func (s *Server) handleAIThumbnails(w http.ResponseWriter, r *http.Request) {
	var req AIThumbnailsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	if req.ImagePath == "" && req.DriveFileID == "" {
		respondError(w, http.StatusBadRequest, "imagePath or driveFileId is required", "")
		return
	}

	ref := thumbnail.ThumbnailRef{Path: req.ImagePath, DriveFileID: req.DriveFileID}
	var prompts ai.VariationPrompts
	err := thumbnail.WithThumbnailFile(r.Context(), ref, s.driveService, func(localPath string) error {
		var genErr error
		prompts, genErr = s.aiService.GenerateThumbnailVariations(r.Context(), localPath)
		return genErr
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIThumbnailsResponse{Subtle: prompts.Subtle, Bold: prompts.Bold})
}

// handleAITranslate translates video metadata to a target language.
func (s *Server) handleAITranslate(w http.ResponseWriter, r *http.Request) {
	var req AITranslateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}
	if req.Category == "" || req.Name == "" || req.TargetLanguage == "" {
		respondError(w, http.StatusBadRequest, "category, name, and targetLanguage are required", "")
		return
	}
	v, err := s.videoService.GetVideo(req.Name, req.Category)
	if err != nil {
		respondError(w, http.StatusNotFound, "video not found", err.Error())
		return
	}
	// Use the first title variant as the primary title
	var title string
	if len(v.Titles) > 0 {
		title = v.Titles[0].Text
	}
	// Build short titles from the video's shorts
	var shortTitles []string
	for _, short := range v.Shorts {
		if short.Title != "" {
			shortTitles = append(shortTitles, short.Title)
		}
	}
	input := ai.VideoMetadataInput{
		Title:       title,
		Description: v.Description,
		Tags:        v.Tags,
		Timecodes:   v.Timecodes,
		ShortTitles: shortTitles,
	}
	output, err := s.aiService.TranslateVideoMetadata(r.Context(), input, req.TargetLanguage)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AITranslateResponse{
		Title:       output.Title,
		Description: output.Description,
		Tags:        output.Tags,
		Timecodes:   output.Timecodes,
		ShortTitles: output.ShortTitles,
	})
}

// handleAIAMAContent generates all AMA content from the manuscript.
func (s *Server) handleAIAMAContent(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getAMAManuscript(w, r)
	if !ok {
		return
	}
	content, err := s.aiService.GenerateAMAContent(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIAMAContentResponse{
		Title:       content.Title,
		Timecodes:   content.Timecodes,
		Description: content.Description,
		Tags:        content.Tags,
	})
}

// handleAIAMATitle generates a title for an AMA video.
func (s *Server) handleAIAMATitle(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getAMAManuscript(w, r)
	if !ok {
		return
	}
	title, err := s.aiService.GenerateAMATitle(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIAMATitleResponse{Title: title})
}

// handleAIAMADescription generates a description for an AMA video.
func (s *Server) handleAIAMADescription(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getAMAManuscript(w, r)
	if !ok {
		return
	}
	desc, err := s.aiService.GenerateAMADescription(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIAMADescriptionResponse{Description: desc})
}

// handleAIAMATimecodes generates timecodes for an AMA video.
func (s *Server) handleAIAMATimecodes(w http.ResponseWriter, r *http.Request) {
	manuscript, ok := s.getAMAManuscript(w, r)
	if !ok {
		return
	}
	timecodes, err := s.aiService.GenerateAMATimecodes(r.Context(), manuscript)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "AI generation failed", err.Error())
		return
	}
	respondJSON(w, http.StatusOK, AIAMATimecodesResponse{Timecodes: timecodes})
}

// --- Helpers ---

// getManuscriptFromPath extracts category and name from URL path params,
// fetches the manuscript, and returns it. Returns false if an error response was sent.
func (s *Server) getManuscriptFromPath(w http.ResponseWriter, r *http.Request) (string, bool) {
	category := chi.URLParam(r, "category")
	name := chi.URLParam(r, "name")
	if category == "" || name == "" {
		respondError(w, http.StatusBadRequest, "category and name are required", "")
		return "", false
	}
	manuscript, err := s.videoService.GetVideoManuscript(name, category)
	if err != nil {
		respondError(w, http.StatusNotFound, "manuscript not found", err.Error())
		return "", false
	}
	return manuscript, true
}

// getAMAManuscript extracts category and name from the request body and fetches the manuscript.
func (s *Server) getAMAManuscript(w http.ResponseWriter, r *http.Request) (string, bool) {
	var req AIAMARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return "", false
	}
	if req.Category == "" || req.Name == "" {
		respondError(w, http.StatusBadRequest, "category and name are required", "")
		return "", false
	}
	manuscript, err := s.videoService.GetVideoManuscript(req.Name, req.Category)
	if err != nil {
		respondError(w, http.StatusNotFound, "manuscript not found", err.Error())
		return "", false
	}
	return manuscript, true
}
