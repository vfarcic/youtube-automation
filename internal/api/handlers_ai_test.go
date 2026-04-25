package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/storage"
)

// mockAIService is a configurable mock for AIService.
type mockAIService struct {
	titles          []string
	description     string
	tags            string
	tweets          []string
	descriptionTags string
	shorts          []ai.ShortCandidate
	thumbnails      ai.VariationPrompts
	translate       *ai.VideoMetadataOutput
	amaContent      *ai.AMAContent
	amaTitle        string
	amaDescription  string
	amaTimecodes    string
	taglineAndIllustrations *ai.TaglineAndIllustrationsResult
	err                    error
}

func (m *mockAIService) SuggestTitles(ctx context.Context, manuscript string, dataDir string) ([]string, error) {
	return m.titles, m.err
}
func (m *mockAIService) SuggestDescription(ctx context.Context, manuscript string) (string, error) {
	return m.description, m.err
}
func (m *mockAIService) SuggestTags(ctx context.Context, manuscript string) (string, error) {
	return m.tags, m.err
}
func (m *mockAIService) SuggestTweets(ctx context.Context, manuscript string) ([]string, error) {
	return m.tweets, m.err
}
func (m *mockAIService) SuggestDescriptionTags(ctx context.Context, manuscript string) (string, error) {
	return m.descriptionTags, m.err
}
func (m *mockAIService) AnalyzeShorts(ctx context.Context, manuscript string) ([]ai.ShortCandidate, error) {
	return m.shorts, m.err
}
func (m *mockAIService) GenerateThumbnailVariations(ctx context.Context, imagePath string) (ai.VariationPrompts, error) {
	return m.thumbnails, m.err
}
func (m *mockAIService) TranslateVideoMetadata(ctx context.Context, input ai.VideoMetadataInput, targetLanguage string) (*ai.VideoMetadataOutput, error) {
	return m.translate, m.err
}
func (m *mockAIService) GenerateAMAContent(ctx context.Context, transcript string) (*ai.AMAContent, error) {
	return m.amaContent, m.err
}
func (m *mockAIService) GenerateAMATitle(ctx context.Context, transcript string) (string, error) {
	return m.amaTitle, m.err
}
func (m *mockAIService) GenerateAMADescription(ctx context.Context, transcript string) (string, error) {
	return m.amaDescription, m.err
}
func (m *mockAIService) GenerateAMATimecodes(ctx context.Context, transcript string) (string, error) {
	return m.amaTimecodes, m.err
}
func (m *mockAIService) SuggestTaglineAndIllustrations(ctx context.Context, manuscript string) (*ai.TaglineAndIllustrationsResult, error) {
	return m.taglineAndIllustrations, m.err
}

// setupAITestEnv creates a test environment with a mock AI service.
func setupAITestEnv(t *testing.T, mock *mockAIService) *testEnv {
	t.Helper()
	env := setupTestEnv(t)
	env.server.aiService = mock
	return env
}

// seedVideoWithManuscript seeds a video and writes a manuscript file for it.
func seedVideoWithManuscript(t *testing.T, env *testEnv, name, category, manuscriptContent string) {
	t.Helper()
	v := storage.Video{
		Name:     name,
		Category: category,
		Gist:     filepath.Join("manuscript", category, name+".md"),
	}
	seedVideo(t, env, v)
	mdPath := filepath.Join(env.tmpDir, "manuscript", category, name+".md")
	if err := os.WriteFile(mdPath, []byte(manuscriptContent), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestHandleAITitles(t *testing.T) {
	tests := []struct {
		name       string
		category   string
		videoName  string
		mock       *mockAIService
		hasManus   bool
		wantStatus int
		wantTitles []string
	}{
		{
			name:       "success",
			category:   "devops",
			videoName:  "test-video",
			mock:       &mockAIService{titles: []string{"Title A", "Title B"}},
			hasManus:   true,
			wantStatus: http.StatusOK,
			wantTitles: []string{"Title A", "Title B"},
		},
		{
			name:       "manuscript not found",
			category:   "devops",
			videoName:  "nonexistent",
			mock:       &mockAIService{},
			hasManus:   false,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "AI error",
			category:   "devops",
			videoName:  "test-video",
			mock:       &mockAIService{err: fmt.Errorf("AI provider failed")},
			hasManus:   true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, tt.videoName, tt.category, "# Test Manuscript")
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/ai/titles/%s/%s", tt.category, tt.videoName), nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AITitlesResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if len(resp.Titles) != len(tt.wantTitles) {
					t.Errorf("titles count = %d, want %d", len(resp.Titles), len(tt.wantTitles))
				}
			}
		})
	}
}

func TestHandleAIDescription(t *testing.T) {
	tests := []struct {
		name       string
		hasManus   bool
		mock       *mockAIService
		wantStatus int
	}{
		{
			name:       "success",
			hasManus:   true,
			mock:       &mockAIService{description: "A great video description"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "manuscript not found",
			hasManus:   false,
			mock:       &mockAIService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "AI error",
			hasManus:   true,
			mock:       &mockAIService{err: fmt.Errorf("fail")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			videoName := "test-video"
			if tt.hasManus {
				seedVideoWithManuscript(t, env, videoName, "devops", "# Manuscript")
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/ai/description/devops/%s", videoName), nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusOK {
				var resp AIDescriptionResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Description != "A great video description" {
					t.Errorf("description = %q, want %q", resp.Description, "A great video description")
				}
			}
		})
	}
}

func TestHandleAITags(t *testing.T) {
	tests := []struct {
		name       string
		videoName  string
		mock       *mockAIService
		hasManus   bool
		wantStatus int
	}{
		{
			name:       "success",
			videoName:  "test-video",
			mock:       &mockAIService{tags: "go,kubernetes,devops"},
			hasManus:   true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "AI error",
			videoName:  "test-video",
			mock:       &mockAIService{err: fmt.Errorf("AI failed")},
			hasManus:   true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "manuscript not found",
			videoName:  "nonexistent",
			mock:       &mockAIService{},
			hasManus:   false,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, tt.videoName, "devops", "# Manuscript")
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/ai/tags/devops/%s", tt.videoName), nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AITagsResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Tags != "go,kubernetes,devops" {
					t.Errorf("tags = %q, want %q", resp.Tags, "go,kubernetes,devops")
				}
			}
		})
	}
}

func TestHandleAITweets(t *testing.T) {
	tests := []struct {
		name       string
		videoName  string
		mock       *mockAIService
		hasManus   bool
		wantStatus int
	}{
		{
			name:       "success",
			videoName:  "test-video",
			mock:       &mockAIService{tweets: []string{"tweet1", "tweet2"}},
			hasManus:   true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "AI error",
			videoName:  "test-video",
			mock:       &mockAIService{err: fmt.Errorf("AI failed")},
			hasManus:   true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "manuscript not found",
			videoName:  "nonexistent",
			mock:       &mockAIService{},
			hasManus:   false,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, tt.videoName, "devops", "# Manuscript")
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/ai/tweets/devops/%s", tt.videoName), nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AITweetsResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if len(resp.Tweets) != 2 {
					t.Errorf("tweets count = %d, want 2", len(resp.Tweets))
				}
			}
		})
	}
}

func TestHandleAIDescriptionTags(t *testing.T) {
	tests := []struct {
		name       string
		videoName  string
		mock       *mockAIService
		hasManus   bool
		wantStatus int
	}{
		{
			name:       "success",
			videoName:  "test-video",
			mock:       &mockAIService{descriptionTags: "#go #k8s #devops"},
			hasManus:   true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "AI error",
			videoName:  "test-video",
			mock:       &mockAIService{err: fmt.Errorf("AI failed")},
			hasManus:   true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "manuscript not found",
			videoName:  "nonexistent",
			mock:       &mockAIService{},
			hasManus:   false,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, tt.videoName, "devops", "# Manuscript")
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/ai/description-tags/devops/%s", tt.videoName), nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AIDescriptionTagsResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.DescriptionTags != "#go #k8s #devops" {
					t.Errorf("descriptionTags = %q, want %q", resp.DescriptionTags, "#go #k8s #devops")
				}
			}
		})
	}
}

func TestHandleAIShorts(t *testing.T) {
	tests := []struct {
		name            string
		videoName       string
		mock            *mockAIService
		hasManus        bool
		manuscriptBody  string
		wantStatus      int
		wantWarning     string   // substring expected in markersWarning ("" = no warning)
		wantMarkerIDs   []string // short IDs whose markers should appear in the file
		wantNoMarkerIDs []string // short IDs whose markers should NOT appear
	}{
		{
			name:      "markers inserted",
			videoName: "test-video",
			mock: &mockAIService{shorts: []ai.ShortCandidate{
				{ID: "short1", Title: "Short One", Text: "Target text", Rationale: "reason"},
			}},
			hasManus:       true,
			manuscriptBody: "Intro.\n\nTarget text\n\nOutro.",
			wantStatus:     http.StatusOK,
			wantMarkerIDs:  []string{"short1"},
		},
		{
			name:      "partial failure",
			videoName: "test-video",
			mock: &mockAIService{shorts: []ai.ShortCandidate{
				{ID: "short1", Title: "Found", Text: "Found text", Rationale: "r"},
				{ID: "short2", Title: "Missing", Text: "Missing text", Rationale: "r"},
			}},
			hasManus:        true,
			manuscriptBody:  "Intro.\n\nFound text\n\nOutro.",
			wantStatus:      http.StatusOK,
			wantWarning:     "markers inserted",
			wantMarkerIDs:   []string{"short1"},
			wantNoMarkerIDs: []string{"short2"},
		},
		{
			name:      "no segments found",
			videoName: "test-video",
			mock: &mockAIService{shorts: []ai.ShortCandidate{
				{ID: "short1", Title: "Nope", Text: "Nonexistent text", Rationale: "r"},
			}},
			hasManus:        true,
			manuscriptBody:  "Completely different content",
			wantStatus:      http.StatusOK,
			wantWarning:     "no short segments found",
			wantNoMarkerIDs: []string{"short1"},
		},
		{
			name:       "AI error",
			videoName:  "test-video",
			mock:       &mockAIService{err: fmt.Errorf("AI failed")},
			hasManus:   true,
			manuscriptBody: "# Manuscript",
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "manuscript not found",
			videoName:  "nonexistent",
			mock:       &mockAIService{},
			hasManus:   false,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			manuscriptContent := tt.manuscriptBody
			if manuscriptContent == "" {
				manuscriptContent = "# Manuscript"
			}
			if tt.hasManus {
				seedVideoWithManuscript(t, env, tt.videoName, "devops", manuscriptContent)
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/ai/shorts/devops/%s", tt.videoName), nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus != http.StatusOK {
				return
			}

			var resp AIShortsResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if len(resp.Candidates) != len(tt.mock.shorts) {
				t.Errorf("candidates count = %d, want %d", len(resp.Candidates), len(tt.mock.shorts))
			}

			// Check warning
			if tt.wantWarning == "" && resp.MarkersWarning != "" {
				t.Errorf("unexpected markersWarning = %q", resp.MarkersWarning)
			}
			if tt.wantWarning != "" && !strings.Contains(resp.MarkersWarning, tt.wantWarning) {
				t.Errorf("markersWarning = %q, want substring %q", resp.MarkersWarning, tt.wantWarning)
			}

			// Read manuscript file to verify markers on disk
			mdPath := filepath.Join(env.tmpDir, "manuscript", "devops", tt.videoName+".md")
			fileBytes, err := os.ReadFile(mdPath)
			if err != nil {
				t.Fatalf("failed to read manuscript file: %v", err)
			}
			fileContent := string(fileBytes)

			for _, id := range tt.wantMarkerIDs {
				startMarker := fmt.Sprintf("TODO: Short (id: %s) (start)", id)
				endMarker := fmt.Sprintf("TODO: Short (id: %s) (end)", id)
				if !strings.Contains(fileContent, startMarker) {
					t.Errorf("manuscript missing start marker for %s", id)
				}
				if !strings.Contains(fileContent, endMarker) {
					t.Errorf("manuscript missing end marker for %s", id)
				}
			}
			for _, id := range tt.wantNoMarkerIDs {
				startMarker := fmt.Sprintf("TODO: Short (id: %s) (start)", id)
				if strings.Contains(fileContent, startMarker) {
					t.Errorf("manuscript should NOT contain marker for %s but does", id)
				}
			}
		})
	}
}

func TestHandleAIThumbnails(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       *mockAIService
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"imagePath":"/tmp/thumb.png"}`,
			mock:       &mockAIService{thumbnails: ai.VariationPrompts{Subtle: "subtle prompt", Bold: "bold prompt"}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing imagePath and driveFileId",
			body:       `{"imagePath":"","driveFileId":""}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "AI error",
			body:       `{"imagePath":"/tmp/thumb.png"}`,
			mock:       &mockAIService{err: fmt.Errorf("vision failed")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "driveFileId without drive service returns error",
			body:       `{"driveFileId":"drive-abc123"}`,
			mock:       &mockAIService{thumbnails: ai.VariationPrompts{Subtle: "s", Bold: "b"}},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			req := httptest.NewRequest(http.MethodPost, "/api/ai/thumbnails", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AIThumbnailsResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Subtle != "subtle prompt" || resp.Bold != "bold prompt" {
					t.Errorf("unexpected response: %+v", resp)
				}
			}
		})
	}
}

func TestHandleAIThumbnails_DriveFileID(t *testing.T) {
	mock := &mockAIService{thumbnails: ai.VariationPrompts{Subtle: "drive subtle", Bold: "drive bold"}}
	env := setupAITestEnv(t, mock)
	env.server.driveService = &mockDriveService{
		getFileContent: "fake-image-data",
		getFileMIME:    "image/png",
		getFileName:    "thumb.png",
	}

	body := `{"driveFileId":"drive-thumb-123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/ai/thumbnails", strings.NewReader(body))
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var resp AIThumbnailsResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Subtle != "drive subtle" || resp.Bold != "drive bold" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestHandleAITranslate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		seedVideo  bool
		mock       *mockAIService
		wantStatus int
	}{
		{
			name:      "success",
			body:      `{"category":"devops","name":"test-video","targetLanguage":"es"}`,
			seedVideo: true,
			mock: &mockAIService{translate: &ai.VideoMetadataOutput{
				Title:       "Titulo",
				Description: "Descripcion",
				Tags:        "etiquetas",
				Timecodes:   "00:00 Intro",
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing fields",
			body:       `{"category":"devops","name":"test-video"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "video not found",
			body:       `{"category":"devops","name":"nonexistent","targetLanguage":"es"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "AI error",
			body:      `{"category":"devops","name":"test-video","targetLanguage":"es"}`,
			seedVideo: true,
			mock:      &mockAIService{err: fmt.Errorf("translation failed")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.seedVideo {
				seedVideoWithManuscript(t, env, "test-video", "devops", "# Manuscript")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/ai/translate", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AITranslateResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Title != "Titulo" {
					t.Errorf("title = %q, want %q", resp.Title, "Titulo")
				}
			}
		})
	}
}

func TestHandleAIAMAContent(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		hasManus   bool
		mock       *mockAIService
		wantStatus int
	}{
		{
			name:     "success",
			body:     `{"category":"devops","name":"test-video"}`,
			hasManus: true,
			mock: &mockAIService{amaContent: &ai.AMAContent{
				Title: "AMA Title", Timecodes: "00:00 Intro", Description: "AMA Desc", Tags: "ama,devops",
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing fields",
			body:       `{"category":"devops"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "manuscript not found",
			body:       `{"category":"devops","name":"nonexistent"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusNotFound,
		},
		{
			name:     "AI error",
			body:     `{"category":"devops","name":"test-video"}`,
			hasManus: true,
			mock:     &mockAIService{err: fmt.Errorf("fail")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, "test-video", "devops", "# AMA Transcript")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/ai/ama/content", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AIAMAContentResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Title != "AMA Title" {
					t.Errorf("title = %q, want %q", resp.Title, "AMA Title")
				}
			}
		})
	}
}

func TestHandleAIAMATitle(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		hasManus   bool
		mock       *mockAIService
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"category":"devops","name":"test-video"}`,
			hasManus:   true,
			mock:       &mockAIService{amaTitle: "My AMA Title"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "AI error",
			body:       `{"category":"devops","name":"test-video"}`,
			hasManus:   true,
			mock:       &mockAIService{err: fmt.Errorf("AI failed")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "missing fields",
			body:       `{"category":"devops"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "manuscript not found",
			body:       `{"category":"devops","name":"nonexistent"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, "test-video", "devops", "# Transcript")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/ai/ama/title", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AIAMATitleResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Title != "My AMA Title" {
					t.Errorf("title = %q, want %q", resp.Title, "My AMA Title")
				}
			}
		})
	}
}

func TestHandleAIAMADescription(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		hasManus   bool
		mock       *mockAIService
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"category":"devops","name":"test-video"}`,
			hasManus:   true,
			mock:       &mockAIService{amaDescription: "AMA desc"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "AI error",
			body:       `{"category":"devops","name":"test-video"}`,
			hasManus:   true,
			mock:       &mockAIService{err: fmt.Errorf("AI failed")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "missing fields",
			body:       `{"category":"devops"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "manuscript not found",
			body:       `{"category":"devops","name":"nonexistent"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, "test-video", "devops", "# Transcript")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/ai/ama/description", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AIAMADescriptionResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Description != "AMA desc" {
					t.Errorf("description = %q, want %q", resp.Description, "AMA desc")
				}
			}
		})
	}
}

func TestHandleAIAMATimecodes(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		hasManus   bool
		mock       *mockAIService
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"category":"devops","name":"test-video"}`,
			hasManus:   true,
			mock:       &mockAIService{amaTimecodes: "00:00 Intro\n01:00 Q1"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "AI error",
			body:       `{"category":"devops","name":"test-video"}`,
			hasManus:   true,
			mock:       &mockAIService{err: fmt.Errorf("AI failed")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "missing fields",
			body:       `{"category":"devops"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "manuscript not found",
			body:       `{"category":"devops","name":"nonexistent"}`,
			mock:       &mockAIService{},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, "test-video", "devops", "# Transcript")
			}

			req := httptest.NewRequest(http.MethodPost, "/api/ai/ama/timecodes", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AIAMATimecodesResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Timecodes != "00:00 Intro\n01:00 Q1" {
					t.Errorf("timecodes = %q, want %q", resp.Timecodes, "00:00 Intro\n01:00 Q1")
				}
			}
		})
	}
}

func TestManuscriptPathParamsMissing(t *testing.T) {
	env := setupAITestEnv(t, &mockAIService{})

	// Empty category — will match a different route or 404
	req := httptest.NewRequest(http.MethodPost, "/api/ai/titles//", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	// chi treats empty path params as 404 (no route match)
	if rr.Code == http.StatusOK {
		t.Errorf("expected non-200 for empty params, got %d", rr.Code)
	}
}


func TestHandleAITaglineAndIllustrations(t *testing.T) {
	tests := []struct {
		name              string
		category          string
		videoName         string
		mock              *mockAIService
		hasManus          bool
		wantStatus        int
		wantTaglines      int
		wantIllustrations int
	}{
		{
			name:     "success",
			category: "devops",
			videoName: "test-video",
			mock: &mockAIService{taglineAndIllustrations: &ai.TaglineAndIllustrationsResult{
				Taglines:      []string{"Secure Everything", "Lock It Down", "Zero Trust"},
				Illustrations: []string{"Fortress protecting servers", "Shield around clusters", "Cracked lock being fixed"},
			}},
			hasManus:          true,
			wantStatus:        http.StatusOK,
			wantTaglines:      3,
			wantIllustrations: 3,
		},
		{
			name:       "manuscript not found",
			category:   "devops",
			videoName:  "nonexistent",
			mock:       &mockAIService{},
			hasManus:   false,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "AI error",
			category:   "devops",
			videoName:  "test-video",
			mock:       &mockAIService{err: fmt.Errorf("AI provider failed")},
			hasManus:   true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, tt.mock)
			if tt.hasManus {
				seedVideoWithManuscript(t, env, tt.videoName, tt.category, "# Test Manuscript\nSome content here.")
			}

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/ai/tagline-and-illustrations/%s/%s", tt.category, tt.videoName), nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AITaglineAndIllustrationsResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if len(resp.Taglines) != tt.wantTaglines {
					t.Errorf("taglines count = %d, want %d", len(resp.Taglines), tt.wantTaglines)
				}
				if len(resp.Illustrations) != tt.wantIllustrations {
					t.Errorf("illustrations count = %d, want %d", len(resp.Illustrations), tt.wantIllustrations)
				}
			}
		})
	}
}

func TestHandleAITaglineAndIllustrations_PathTraversal(t *testing.T) {
	// Path traversal tests use URL-encoded slashes (%2F) because chi's router
	// treats literal slashes as path separators (no route match → 404).
	// URL-encoded slashes are decoded within the matched segment, reaching the handler.
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "path traversal in category with encoded slashes",
			url:        "/api/ai/tagline-and-illustrations/..%2F..%2Fetc/passwd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "path traversal in name with encoded slashes",
			url:        "/api/ai/tagline-and-illustrations/devops/..%2F..%2Fetc%2Fpasswd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "backslash in category",
			url:        "/api/ai/tagline-and-illustrations/dev%5Cops/test-video",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "backslash in name",
			url:        "/api/ai/tagline-and-illustrations/devops/test%5Cvideo",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "dot-dot without slash in category",
			url:        "/api/ai/tagline-and-illustrations/..devops/test-video",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, &mockAIService{})

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

func TestHandleAIIllustrations_NoPathLeakage(t *testing.T) {
	env := setupAITestEnv(t, &mockAIService{})

	// Request a nonexistent video — the error should NOT contain file system paths
	req := httptest.NewRequest(http.MethodPost, "/api/ai/tagline-and-illustrations/devops/nonexistent", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}

	body := rr.Body.String()
	// Error body must not contain internal file paths
	if strings.Contains(body, env.tmpDir) {
		t.Errorf("error response leaks internal path: %s", body)
	}
	if strings.Contains(body, "/manuscript/") {
		t.Errorf("error response leaks manuscript path: %s", body)
	}
	if strings.Contains(body, ".yaml") {
		t.Errorf("error response leaks yaml path: %s", body)
	}

	// Verify the response has a generic error message
	var resp ErrorResponse
	if err := json.NewDecoder(strings.NewReader(body)).Decode(&resp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if resp.Error != "manuscript not found" {
		t.Errorf("error = %q, want %q", resp.Error, "manuscript not found")
	}
	if resp.Detail != "" {
		t.Errorf("detail should be empty, got %q", resp.Detail)
	}
}

func TestGetManuscriptFromPath_PathTraversal(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "dot-dot in category (URL-encoded slash)",
			url:        "/api/ai/titles/..%2F..%2Fetc/passwd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "dot-dot in name (URL-encoded slash)",
			url:        "/api/ai/titles/devops/..%2F..%2Fetc%2Fpasswd",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "backslash in name",
			url:        "/api/ai/titles/devops/test%5Cvideo",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAITestEnv(t, &mockAIService{})

			// Test via titles endpoint which uses getManuscriptFromPath
			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}
