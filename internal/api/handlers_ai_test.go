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
	err             error
}

func (m *mockAIService) SuggestTitles(ctx context.Context, manuscript string) ([]string, error) {
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
	env := setupAITestEnv(t, &mockAIService{tags: "go,kubernetes,devops"})
	seedVideoWithManuscript(t, env, "test-video", "devops", "# Manuscript")

	req := httptest.NewRequest(http.MethodPost, "/api/ai/tags/devops/test-video", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp AITagsResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Tags != "go,kubernetes,devops" {
		t.Errorf("tags = %q, want %q", resp.Tags, "go,kubernetes,devops")
	}
}

func TestHandleAITweets(t *testing.T) {
	env := setupAITestEnv(t, &mockAIService{tweets: []string{"tweet1", "tweet2"}})
	seedVideoWithManuscript(t, env, "test-video", "devops", "# Manuscript")

	req := httptest.NewRequest(http.MethodPost, "/api/ai/tweets/devops/test-video", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp AITweetsResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Tweets) != 2 {
		t.Errorf("tweets count = %d, want 2", len(resp.Tweets))
	}
}

func TestHandleAIDescriptionTags(t *testing.T) {
	env := setupAITestEnv(t, &mockAIService{descriptionTags: "#go #k8s #devops"})
	seedVideoWithManuscript(t, env, "test-video", "devops", "# Manuscript")

	req := httptest.NewRequest(http.MethodPost, "/api/ai/description-tags/devops/test-video", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp AIDescriptionTagsResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.DescriptionTags != "#go #k8s #devops" {
		t.Errorf("descriptionTags = %q, want %q", resp.DescriptionTags, "#go #k8s #devops")
	}
}

func TestHandleAIShorts(t *testing.T) {
	candidates := []ai.ShortCandidate{
		{ID: "short1", Title: "Short One", Text: "text", Rationale: "reason"},
	}
	env := setupAITestEnv(t, &mockAIService{shorts: candidates})
	seedVideoWithManuscript(t, env, "test-video", "devops", "# Manuscript")

	req := httptest.NewRequest(http.MethodPost, "/api/ai/shorts/devops/test-video", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp AIShortsResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if len(resp.Candidates) != 1 || resp.Candidates[0].ID != "short1" {
		t.Errorf("unexpected shorts response: %+v", resp)
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
	env := setupAITestEnv(t, &mockAIService{amaTitle: "My AMA Title"})
	seedVideoWithManuscript(t, env, "test-video", "devops", "# Transcript")

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ama/title", strings.NewReader(`{"category":"devops","name":"test-video"}`))
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rr.Code, http.StatusOK, rr.Body.String())
	}
	var resp AIAMATitleResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Title != "My AMA Title" {
		t.Errorf("title = %q, want %q", resp.Title, "My AMA Title")
	}
}

func TestHandleAIAMADescription(t *testing.T) {
	env := setupAITestEnv(t, &mockAIService{amaDescription: "AMA desc"})
	seedVideoWithManuscript(t, env, "test-video", "devops", "# Transcript")

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ama/description", strings.NewReader(`{"category":"devops","name":"test-video"}`))
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp AIAMADescriptionResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Description != "AMA desc" {
		t.Errorf("description = %q, want %q", resp.Description, "AMA desc")
	}
}

func TestHandleAIAMATimecodes(t *testing.T) {
	env := setupAITestEnv(t, &mockAIService{amaTimecodes: "00:00 Intro\n01:00 Q1"})
	seedVideoWithManuscript(t, env, "test-video", "devops", "# Transcript")

	req := httptest.NewRequest(http.MethodPost, "/api/ai/ama/timecodes", strings.NewReader(`{"category":"devops","name":"test-video"}`))
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	var resp AIAMATimecodesResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Timecodes != "00:00 Intro\n01:00 Q1" {
		t.Errorf("timecodes = %q, want %q", resp.Timecodes, "00:00 Intro\n01:00 Q1")
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
