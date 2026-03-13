package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/publishing"
	"devopstoolkit/youtube-automation/internal/storage"
)

// mockPublishingService implements PublishingService for testing.
type mockPublishingService struct {
	uploadVideoID     string
	uploadVideoErr    error
	uploadThumbnailErr error
	uploadShortID     string
	uploadShortErr    error
	hugoPath          string
	hugoErr           error
	transcript        string
	transcriptErr     error
	metadata          *publishing.VideoMetadata
	metadataErr       error
	blueSkyErr        error
	slackErr          error
	updateAMAErr      error
}

func (m *mockPublishingService) UploadVideo(_ context.Context, _ *storage.Video) (string, error) {
	return m.uploadVideoID, m.uploadVideoErr
}
func (m *mockPublishingService) UploadThumbnail(_ context.Context, _, _ string) error {
	return m.uploadThumbnailErr
}
func (m *mockPublishingService) UploadShort(_ context.Context, _ string, _ storage.Short, _ string) (string, error) {
	return m.uploadShortID, m.uploadShortErr
}
func (m *mockPublishingService) CreateHugoPost(_ context.Context, _ *storage.Video, _ *publishing.HugoPostOptions) (string, error) {
	return m.hugoPath, m.hugoErr
}
func (m *mockPublishingService) GetTranscript(_ context.Context, _ string) (string, error) {
	return m.transcript, m.transcriptErr
}
func (m *mockPublishingService) GetVideoMetadata(_ context.Context, _ string) (*publishing.VideoMetadata, error) {
	return m.metadata, m.metadataErr
}
func (m *mockPublishingService) PostBlueSky(_ context.Context, _, _, _ string) error {
	return m.blueSkyErr
}
func (m *mockPublishingService) PostSlack(_ context.Context, _ *storage.Video, _ string) error {
	return m.slackErr
}
func (m *mockPublishingService) UpdateAMAVideo(_ context.Context, _, _, _, _, _ string) error {
	return m.updateAMAErr
}

func setupPublishTestEnv(t *testing.T, mock *mockPublishingService) *testEnv {
	t.Helper()
	env := setupTestEnv(t)
	env.server.publishingService = mock
	return env
}

func seedPublishVideo(t *testing.T, env *testEnv) {
	t.Helper()
	v := storage.Video{
		Name:     "test-video",
		Category: "devops",
		Titles:   []storage.TitleVariant{{Index: 1, Text: "Test Video Title"}},
		VideoId:  "yt-abc123",
		Gist:     "manuscript/devops/test-video.md",
		Tweet:    "Check out [YOUTUBE]",
		Description: "A test video",
		UploadVideo: "/tmp/video.mp4",
		Thumbnail:   "/tmp/thumb.png",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One", FilePath: "/tmp/short1.mp4", ScheduledDate: "2026-01-01T10:00"},
		},
	}
	seedVideo(t, env, v)
}

// --- YouTube Publish Tests ---

func TestHandlePublishYouTube(t *testing.T) {
	tests := []struct {
		name       string
		videoName  string
		category   string
		mock       *mockPublishingService
		seedVideo  bool
		wantStatus int
	}{
		{
			name:       "not configured",
			videoName:  "test-video",
			category:   "devops",
			mock:       nil,
			seedVideo:  false,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "missing category",
			videoName:  "test-video",
			category:   "",
			mock:       &mockPublishingService{uploadVideoID: "yt-new"},
			seedVideo:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "video not found",
			videoName:  "nonexistent",
			category:   "devops",
			mock:       &mockPublishingService{uploadVideoID: "yt-new"},
			seedVideo:  false,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "success",
			videoName:  "test-video",
			category:   "devops",
			mock:       &mockPublishingService{uploadVideoID: "yt-new-id"},
			seedVideo:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "upload error",
			videoName:  "test-video",
			category:   "devops",
			mock:       &mockPublishingService{uploadVideoErr: fmt.Errorf("upload failed")},
			seedVideo:  true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *testEnv
			if tt.mock != nil {
				env = setupPublishTestEnv(t, tt.mock)
			} else {
				env = setupTestEnv(t)
			}
			if tt.seedVideo {
				seedPublishVideo(t, env)
			}

			url := fmt.Sprintf("/api/publish/youtube/%s", tt.videoName)
			if tt.category != "" {
				url += "?category=" + tt.category
			}
			req := httptest.NewRequest(http.MethodPost, url, nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp PublishYouTubeResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.VideoID != "yt-new-id" {
					t.Errorf("videoId = %q, want %q", resp.VideoID, "yt-new-id")
				}
			}
		})
	}
}

// --- Thumbnail Publish Tests ---

func TestHandlePublishThumbnail(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		mock       *mockPublishingService
		seedVideo  bool
		wantStatus int
	}{
		{
			name:       "not configured",
			url:        "/api/publish/youtube/test-video/thumbnail?category=devops",
			mock:       nil,
			seedVideo:  false,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "success",
			url:        "/api/publish/youtube/test-video/thumbnail?category=devops",
			mock:       &mockPublishingService{},
			seedVideo:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing category",
			url:        "/api/publish/youtube/test-video/thumbnail",
			mock:       &mockPublishingService{},
			seedVideo:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "video not found",
			url:        "/api/publish/youtube/nonexistent/thumbnail?category=devops",
			mock:       &mockPublishingService{},
			seedVideo:  false,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "video has no YouTube ID",
			url:        "/api/publish/youtube/no-yt-video/thumbnail?category=devops",
			mock:       &mockPublishingService{},
			seedVideo:  false,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *testEnv
			if tt.mock != nil {
				env = setupPublishTestEnv(t, tt.mock)
			} else {
				env = setupTestEnv(t)
			}
			if tt.seedVideo {
				seedPublishVideo(t, env)
			}

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
		})
	}
}

// --- Short Publish Tests ---

func TestHandlePublishShort(t *testing.T) {
	tests := []struct {
		name       string
		shortID    string
		url        string
		mock       *mockPublishingService
		seedVideo  bool
		wantStatus int
	}{
		{
			name:       "not configured",
			shortID:    "short1",
			url:        "/api/publish/youtube/test-video/shorts/short1?category=devops",
			mock:       nil,
			seedVideo:  false,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "short not found",
			shortID:    "nonexistent",
			url:        "/api/publish/youtube/test-video/shorts/nonexistent?category=devops",
			mock:       &mockPublishingService{uploadShortID: "yt-short"},
			seedVideo:  true,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "success",
			shortID:    "short1",
			url:        "/api/publish/youtube/test-video/shorts/short1?category=devops",
			mock:       &mockPublishingService{uploadShortID: "yt-short-id"},
			seedVideo:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing category",
			shortID:    "short1",
			url:        "/api/publish/youtube/test-video/shorts/short1",
			mock:       &mockPublishingService{},
			seedVideo:  false,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "upload error",
			shortID:    "short1",
			url:        "/api/publish/youtube/test-video/shorts/short1?category=devops",
			mock:       &mockPublishingService{uploadShortErr: fmt.Errorf("upload failed")},
			seedVideo:  true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *testEnv
			if tt.mock != nil {
				env = setupPublishTestEnv(t, tt.mock)
			} else {
				env = setupTestEnv(t)
			}
			if tt.seedVideo {
				seedPublishVideo(t, env)
			}

			req := httptest.NewRequest(http.MethodPost, tt.url, nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp PublishShortResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.YouTubeID != "yt-short-id" {
					t.Errorf("youtubeId = %q, want %q", resp.YouTubeID, "yt-short-id")
				}
			}
		})
	}
}

// --- Hugo Publish Tests ---

func TestHandlePublishHugo(t *testing.T) {
	tests := []struct {
		name       string
		mock       *mockPublishingService
		seedVideo  bool
		wantStatus int
	}{
		{
			name:       "not configured",
			mock:       nil,
			seedVideo:  false,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "success",
			mock:       &mockPublishingService{hugoPath: "/content/devops/test-video/_index.md"},
			seedVideo:  true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "hugo error",
			mock:       &mockPublishingService{hugoErr: fmt.Errorf("hugo failed")},
			seedVideo:  true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *testEnv
			if tt.mock != nil {
				env = setupPublishTestEnv(t, tt.mock)
			} else {
				env = setupTestEnv(t)
			}
			if tt.seedVideo {
				seedPublishVideo(t, env)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/publish/hugo/test-video?category=devops", nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp PublishHugoResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.HugoPath != "/content/devops/test-video/_index.md" {
					t.Errorf("hugoPath = %q, want %q", resp.HugoPath, "/content/devops/test-video/_index.md")
				}
			}
		})
	}
}

// --- Transcript Tests ---

func TestHandleGetTranscript(t *testing.T) {
	tests := []struct {
		name       string
		videoID    string
		mock       *mockPublishingService
		wantStatus int
	}{
		{
			name:       "not configured",
			videoID:    "abc123",
			mock:       nil,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "missing videoId",
			videoID:    "",
			mock:       &mockPublishingService{transcript: "text"},
			wantStatus: http.StatusNotFound, // chi won't match empty param
		},
		{
			name:       "success",
			videoID:    "abc123",
			mock:       &mockPublishingService{transcript: "1\n00:00 --> 00:01\nHello\n"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "error",
			videoID:    "abc123",
			mock:       &mockPublishingService{transcriptErr: fmt.Errorf("no captions")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *testEnv
			if tt.mock != nil {
				env = setupPublishTestEnv(t, tt.mock)
			} else {
				env = setupTestEnv(t)
			}

			url := fmt.Sprintf("/api/publish/transcript/%s", tt.videoID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp TranscriptResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Transcript == "" {
					t.Error("expected non-empty transcript")
				}
			}
		})
	}
}

// --- Metadata Tests ---

func TestHandleGetMetadata(t *testing.T) {
	tests := []struct {
		name       string
		videoID    string
		mock       *mockPublishingService
		wantStatus int
	}{
		{
			name:       "not configured",
			videoID:    "abc123",
			mock:       nil,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:    "success",
			videoID: "abc123",
			mock: &mockPublishingService{metadata: &publishing.VideoMetadata{
				Title:       "Test",
				Description: "Desc",
				Tags:        []string{"go"},
				PublishedAt: "2026-01-01T00:00:00Z",
			}},
			wantStatus: http.StatusOK,
		},
		{
			name:       "error",
			videoID:    "abc123",
			mock:       &mockPublishingService{metadataErr: fmt.Errorf("not found")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *testEnv
			if tt.mock != nil {
				env = setupPublishTestEnv(t, tt.mock)
			} else {
				env = setupTestEnv(t)
			}

			url := fmt.Sprintf("/api/publish/metadata/%s", tt.videoID)
			req := httptest.NewRequest(http.MethodGet, url, nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp MetadataResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Title != "Test" {
					t.Errorf("title = %q, want %q", resp.Title, "Test")
				}
			}
		})
	}
}

// --- Social Post Tests ---

func TestHandleSocialPost(t *testing.T) {
	tests := []struct {
		name       string
		platform   string
		mock       *mockPublishingService
		seedVideo  bool
		wantStatus int
		wantPosted bool
	}{
		{
			name:       "not configured",
			platform:   "bluesky",
			mock:       nil,
			seedVideo:  false,
			wantStatus: http.StatusNotImplemented,
		},
		{
			name:       "unknown platform",
			platform:   "twitter",
			mock:       &mockPublishingService{},
			seedVideo:  true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "bluesky success",
			platform:   "bluesky",
			mock:       &mockPublishingService{},
			seedVideo:  true,
			wantStatus: http.StatusOK,
			wantPosted: true,
		},
		{
			name:       "bluesky error",
			platform:   "bluesky",
			mock:       &mockPublishingService{blueSkyErr: fmt.Errorf("auth failed")},
			seedVideo:  true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "linkedin returns message",
			platform:   "linkedin",
			mock:       &mockPublishingService{},
			seedVideo:  true,
			wantStatus: http.StatusOK,
			wantPosted: false,
		},
		{
			name:       "hackernews returns message",
			platform:   "hackernews",
			mock:       &mockPublishingService{},
			seedVideo:  true,
			wantStatus: http.StatusOK,
			wantPosted: false,
		},
		{
			name:       "dot returns message",
			platform:   "dot",
			mock:       &mockPublishingService{},
			seedVideo:  true,
			wantStatus: http.StatusOK,
			wantPosted: false,
		},
		{
			name:       "slack success",
			platform:   "slack",
			mock:       &mockPublishingService{},
			seedVideo:  true,
			wantStatus: http.StatusOK,
			wantPosted: true,
		},
		{
			name:       "slack error",
			platform:   "slack",
			mock:       &mockPublishingService{slackErr: fmt.Errorf("slack failed")},
			seedVideo:  true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var env *testEnv
			if tt.mock != nil {
				env = setupPublishTestEnv(t, tt.mock)
			} else {
				env = setupTestEnv(t)
			}
			if tt.seedVideo {
				seedPublishVideo(t, env)
			}

			url := fmt.Sprintf("/api/social/%s/test-video?category=devops", tt.platform)
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader("{}"))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp SocialPostResponse
				json.NewDecoder(rr.Body).Decode(&resp)
				if resp.Posted != tt.wantPosted {
					t.Errorf("posted = %v, want %v", resp.Posted, tt.wantPosted)
				}
				if !tt.wantPosted && resp.Message == "" {
					t.Error("expected non-empty message for copy-paste platforms")
				}
			}
		})
	}
}

// --- Social Message Formatter Tests ---

func TestFormatLinkedInMessage(t *testing.T) {
	video := &storage.Video{
		Titles:  []storage.TitleVariant{{Index: 1, Text: "My Video"}},
		VideoId: "abc123",
		Tweet:   "Check out [YOUTUBE]",
	}
	msg := formatLinkedInMessage(video)
	if !strings.Contains(msg, "My Video") {
		t.Error("expected title in LinkedIn message")
	}
	if !strings.Contains(msg, "https://youtu.be/abc123") {
		t.Error("expected YouTube URL in LinkedIn message")
	}
}

func TestFormatHNMessage(t *testing.T) {
	video := &storage.Video{
		Titles:  []storage.TitleVariant{{Index: 1, Text: "My Video"}},
		VideoId: "abc123",
	}
	msg := formatHNMessage(video)
	if !strings.Contains(msg, "Title: My Video") {
		t.Error("expected title in HN message")
	}
}

func TestFormatDOTMessage(t *testing.T) {
	video := &storage.Video{
		Titles:      []storage.TitleVariant{{Index: 1, Text: "My Video"}},
		VideoId:     "abc123",
		Description: "A great video",
	}
	msg := formatDOTMessage(video)
	if !strings.Contains(msg, "My Video") {
		t.Error("expected title in DOT message")
	}
	if !strings.Contains(msg, "A great video") {
		t.Error("expected description in DOT message")
	}
}

// --- createTempFromReader Tests ---

// errReader is an io.Reader that always returns an error.
type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

func TestCreateTempFromReader(t *testing.T) {
	tests := []struct {
		name      string
		reader    io.Reader
		filename  string
		wantExt   string
		wantErr   bool
		wantBody  string
	}{
		{
			name:     "success with content",
			reader:   strings.NewReader("video data here"),
			filename: "video.mp4",
			wantExt:  ".mp4",
			wantBody: "video data here",
		},
		{
			name:     "custom extension",
			reader:   strings.NewReader("webm data"),
			filename: "clip.webm",
			wantExt:  ".webm",
			wantBody: "webm data",
		},
		{
			name:     "default extension for empty filename",
			reader:   strings.NewReader("data"),
			filename: "",
			wantExt:  ".mp4",
			wantBody: "data",
		},
		{
			name:    "read error",
			reader:  &errReader{},
			filename: "video.mp4",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := createTempFromReader(tt.reader, tt.filename)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			defer os.Remove(path)

			ext := filepath.Ext(path)
			if ext != tt.wantExt {
				t.Errorf("ext = %q, want %q", ext, tt.wantExt)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read temp file: %v", err)
			}
			if string(content) != tt.wantBody {
				t.Errorf("content = %q, want %q", string(content), tt.wantBody)
			}
		})
	}
}

// --- AMA Apply Tests ---

func TestHandleAMAApply(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		mock       *mockPublishingService
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"videoId":"abc123","title":"AMA Title","description":"AMA Desc","tags":"ama,q&a","timecodes":"00:00 Intro"}`,
			mock:       &mockPublishingService{},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing videoId",
			body:       `{"title":"AMA Title"}`,
			mock:       &mockPublishingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "update error",
			body:       `{"videoId":"abc123","title":"AMA Title"}`,
			mock:       &mockPublishingService{updateAMAErr: errors.New("YouTube API error")},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid JSON",
			body:       `{invalid}`,
			mock:       &mockPublishingService{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupPublishTestEnv(t, tt.mock)
			req := httptest.NewRequest(http.MethodPost, "/api/ama/apply", strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer test-token")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AMAApplyResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if !resp.Success {
					t.Error("expected success = true")
				}
			}
		})
	}
}

func TestHandleAMAApplyNotConfigured(t *testing.T) {
	env := setupTestEnv(t)
	// Don't set publishingService
	req := httptest.NewRequest(http.MethodPost, "/api/ama/apply", strings.NewReader(`{"videoId":"abc"}`))
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
	}
}

// --- AMA Generate Tests ---

func TestHandleAMAGenerate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		pubMock    *mockPublishingService
		wantStatus int
	}{
		{
			name:       "success",
			body:       `{"videoId":"abc123"}`,
			pubMock:    &mockPublishingService{transcript: "Hello, welcome to the AMA"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing videoId",
			body:       `{}`,
			pubMock:    &mockPublishingService{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "transcript error",
			body:       `{"videoId":"abc123"}`,
			pubMock:    &mockPublishingService{transcriptErr: errors.New("no captions")},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			env.server.publishingService = tt.pubMock
			env.server.aiService = &mockAIService{
				amaContent: &ai.AMAContent{
					Title:       "Generated AMA Title",
					Description: "Generated AMA Description",
					Tags:        "ama,generated",
					Timecodes:   "00:00 Intro\n01:00 Q1",
				},
			}
			req := httptest.NewRequest(http.MethodPost, "/api/ama/generate", strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer test-token")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body = %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AMAGenerateResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.Title != "Generated AMA Title" {
					t.Errorf("title = %q, want %q", resp.Title, "Generated AMA Title")
				}
				if resp.Transcript != "Hello, welcome to the AMA" {
					t.Errorf("transcript = %q, want %q", resp.Transcript, "Hello, welcome to the AMA")
				}
			}
		})
	}
}

