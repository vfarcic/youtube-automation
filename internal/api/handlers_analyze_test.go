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
	"devopstoolkit/youtube-automation/internal/publishing"
)

// mockAnalyzeService is a configurable mock for AnalyzeService.
type mockAnalyzeService struct {
	videos          []ai.VideoABData
	loadErr         error
	analytics       []publishing.VideoAnalytics
	analyticsErr    error
	firstWeekErr    error
	enriched        []ai.VideoABData
	analysisResult  ai.TitleAnalysisResult
	analysisRaw     string
	analysisErr     error
}

func (m *mockAnalyzeService) LoadVideosWithABData(indexPath, dataDir, manuscriptDir string) ([]ai.VideoABData, error) {
	return m.videos, m.loadErr
}

func (m *mockAnalyzeService) GetVideoAnalyticsForLastYear(ctx context.Context) ([]publishing.VideoAnalytics, error) {
	return m.analytics, m.analyticsErr
}

func (m *mockAnalyzeService) EnrichWithFirstWeekMetrics(ctx context.Context, analytics []publishing.VideoAnalytics) ([]publishing.VideoAnalytics, error) {
	if m.firstWeekErr != nil {
		return nil, m.firstWeekErr
	}
	return analytics, nil
}

func (m *mockAnalyzeService) EnrichWithAnalytics(videos []ai.VideoABData, analytics []publishing.VideoAnalytics) []ai.VideoABData {
	if m.enriched != nil {
		return m.enriched
	}
	return videos
}

func (m *mockAnalyzeService) AnalyzeTitles(ctx context.Context, videos []ai.VideoABData, baseDir string) (ai.TitleAnalysisResult, string, error) {
	return m.analysisResult, m.analysisRaw, m.analysisErr
}

// mockGitSync is a configurable mock for GitSyncService.
type mockGitSync struct {
	called  bool
	message string
	err     error
}

func (m *mockGitSync) CommitAndPush(message string) error {
	m.called = true
	m.message = message
	return m.err
}

func setupAnalyzeTestEnv(t *testing.T, analyzeSvc AnalyzeService) *testEnv {
	t.Helper()
	env := setupTestEnv(t)
	env.server.analyzeService = analyzeSvc
	env.server.dataDir = env.tmpDir
	return env
}

func TestHandleAnalyzeTitles(t *testing.T) {
	sampleVideos := []ai.VideoABData{
		{Category: "ai", VideoID: "vid1"},
		{Category: "kubernetes", VideoID: "vid2"},
	}
	sampleResult := ai.TitleAnalysisResult{
		HighPerformingPatterns: []ai.TitlePattern{
			{Pattern: "Provocative", Description: "Works well", Impact: "high", Examples: []string{"Why X is dead"}},
		},
		LowPerformingPatterns: []ai.TitlePattern{
			{Pattern: "Listicle", Description: "Underperforms", Impact: "low", Examples: []string{"Top 10 tools"}},
		},
		Recommendations: []ai.TitleRecommendation{
			{Recommendation: "Use provocative titles", Evidence: "55% share", Example: "Stop using X"},
		},
		TitlesMDContent: "# Title Patterns\n\n1. Provocative opinions work best",
	}

	tests := []struct {
		name       string
		mock       *mockAnalyzeService
		wantStatus int
		wantCount  int
	}{
		{
			name: "success",
			mock: &mockAnalyzeService{
				videos:         sampleVideos,
				analytics:      []publishing.VideoAnalytics{{VideoID: "vid1"}, {VideoID: "vid2"}},
				analysisResult: sampleResult,
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name: "no videos with AB data",
			mock: &mockAnalyzeService{
				videos: nil,
			},
			wantStatus: http.StatusOK,
			wantCount:  0,
		},
		{
			name: "load error",
			mock: &mockAnalyzeService{
				loadErr: fmt.Errorf("index not found"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "analytics error",
			mock: &mockAnalyzeService{
				videos:       sampleVideos,
				analyticsErr: fmt.Errorf("YouTube API failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "first week metrics error",
			mock: &mockAnalyzeService{
				videos:       sampleVideos,
				analytics:    []publishing.VideoAnalytics{},
				firstWeekErr: fmt.Errorf("rate limited"),
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "AI analysis error",
			mock: &mockAnalyzeService{
				videos:      sampleVideos,
				analytics:   []publishing.VideoAnalytics{},
				analysisErr: fmt.Errorf("AI provider failed"),
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAnalyzeTestEnv(t, tt.mock)

			req := httptest.NewRequest(http.MethodPost, "/api/analyze/titles", nil)
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}
			if tt.wantStatus == http.StatusOK {
				var resp AnalyzeTitlesResponse
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp.VideoCount != tt.wantCount {
					t.Errorf("videoCount = %d, want %d", resp.VideoCount, tt.wantCount)
				}
				if tt.wantCount > 0 {
					if len(resp.HighPerformingPatterns) == 0 {
						t.Error("expected high performing patterns")
					}
					if resp.TitlesMDContent == "" {
						t.Error("expected titlesMdContent")
					}
				}
			}
		})
	}
}

func TestHandleAnalyzeTitles_NotConfigured(t *testing.T) {
	env := setupTestEnv(t)
	// analyzeService is nil by default

	req := httptest.NewRequest(http.MethodPost, "/api/analyze/titles", nil)
	rr := httptest.NewRecorder()
	env.server.Router().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotImplemented)
	}
}

func TestHandleApplyTitlesTemplate(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		gitSync    *mockGitSync
		wantStatus int
		wantFile   bool
		wantSync   bool
	}{
		{
			name:       "success with git sync",
			body:       `{"content":"# New Title Patterns\n\n1. Use provocative titles"}`,
			gitSync:    &mockGitSync{},
			wantStatus: http.StatusOK,
			wantFile:   true,
			wantSync:   true,
		},
		{
			name:       "success without git sync",
			body:       `{"content":"# Patterns"}`,
			gitSync:    nil,
			wantStatus: http.StatusOK,
			wantFile:   true,
			wantSync:   false,
		},
		{
			name:       "git sync error returns warning",
			body:       `{"content":"# Patterns"}`,
			gitSync:    &mockGitSync{err: fmt.Errorf("push rejected")},
			wantStatus: http.StatusOK,
			wantFile:   true,
			wantSync:   true,
		},
		{
			name:       "empty content",
			body:       `{"content":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			body:       `{bad`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupAnalyzeTestEnv(t, &mockAnalyzeService{})
			if tt.gitSync != nil {
				env.server.gitSync = tt.gitSync
			}

			req := httptest.NewRequest(http.MethodPost, "/api/analyze/titles/apply", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()
			env.server.Router().ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d; body: %s", rr.Code, tt.wantStatus, rr.Body.String())
			}

			if tt.wantFile {
				content, err := os.ReadFile(filepath.Join(env.tmpDir, "titles.md"))
				if err != nil {
					t.Fatalf("titles.md not written: %v", err)
				}
				if len(content) == 0 {
					t.Error("titles.md is empty")
				}
			}

			if tt.wantSync && tt.gitSync != nil {
				if !tt.gitSync.called {
					t.Error("expected git sync to be called")
				}
				if tt.gitSync.err != nil {
					var resp ApplyTitlesResponse
					if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
						t.Fatalf("failed to decode response: %v", err)
					}
					if resp.SyncWarning == "" {
						t.Error("expected sync warning in response")
					}
				}
			}
		})
	}
}
