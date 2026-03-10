package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestAnalyzeTitles(t *testing.T) {
	ctx := context.Background()

	// Sample A/B data for testing
	sampleVideos := []VideoABData{
		{
			Category:            "devops",
			Date:                "2024-01-15T10:00",
			DayOfWeek:           "Monday",
			VideoID:             "video1",
			Titles:              []storage.TitleVariant{{Index: 1, Text: "How to Deploy Kubernetes", Share: 65.0}, {Index: 2, Text: "Kubernetes Deployment Guide", Share: 35.0}},
			FirstWeekViews:      50000,

			FirstWeekLikes:      1200,
			FirstWeekComments:   150,
			FirstWeekEngagement: 2.7,
			HasAnalytics:        true,
		},
		{
			Category:            "devops",
			Date:                "2024-03-20T14:30",
			DayOfWeek:           "Wednesday",
			VideoID:             "video2",
			Titles:              []storage.TitleVariant{{Index: 1, Text: "Docker vs Podman - Complete Comparison", Share: 55.0}, {Index: 2, Text: "Podman or Docker?", Share: 45.0}},
			FirstWeekViews:      35000,

			FirstWeekLikes:      890,
			FirstWeekComments:   95,
			FirstWeekEngagement: 2.8,
			HasAnalytics:        true,
		},
		{
			Category:            "devops",
			Date:                "2024-02-10T09:15",
			DayOfWeek:           "Saturday",
			VideoID:             "video3",
			Titles:              []storage.TitleVariant{{Index: 1, Text: "Top 5 DevOps Tools in 2024", Share: 70.0}, {Index: 2, Text: "DevOps Tools You Need", Share: 30.0}},
			FirstWeekViews:      82000,

			FirstWeekLikes:      2100,
			FirstWeekComments:   280,
			FirstWeekEngagement: 2.9,
			HasAnalytics:        true,
		},
	}

	validJSONResponse := `{
		"highPerformingPatterns": [
			{
				"pattern": "Titles with numbers",
				"description": "Titles containing numbers perform significantly better",
				"impact": "Numbers averaged 67% share vs 33% without",
				"examples": ["Top 5 DevOps Tools (share: 70%)", "How to Deploy Kubernetes (share: 65%)"]
			}
		],
		"lowPerformingPatterns": [],
		"recommendations": [
			{
				"recommendation": "Include numbers in 30-40% of titles",
				"evidence": "Titles with numbers averaged 67% A/B share",
				"example": "Transform 'Kubernetes Guide' to 'Top 5 Kubernetes Best Practices'"
			}
		],
		"titlesMdContent": "# Title Generation Guidelines\n\nBased on A/B test data:\n- Use numbers in titles\n- Keep titles specific\n\n{{.ManuscriptContent}}"
	}`

	tests := []struct {
		name              string
		videos            []VideoABData
		mockResponse      string
		mockError         error
		wantErr           bool
		expectedErrSubstr string
		validateResponse  func(t *testing.T, result TitleAnalysisResult)
	}{
		{
			name:         "Successful analysis with valid data",
			videos:       sampleVideos,
			mockResponse: validJSONResponse,
			wantErr:      false,
			validateResponse: func(t *testing.T, result TitleAnalysisResult) {
				if len(result.HighPerformingPatterns) == 0 {
					t.Error("Expected at least one high-performing pattern")
				}
				if len(result.Recommendations) == 0 {
					t.Error("Expected at least one recommendation")
				}
				if result.TitlesMDContent == "" {
					t.Error("Expected non-empty TitlesMDContent")
				}
			},
		},
		{
			name:              "Empty video data",
			videos:            []VideoABData{},
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "no video data provided",
		},
		{
			name:              "AI returns empty response",
			videos:            sampleVideos,
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "AI returned empty analysis",
		},
		{
			name:              "AI generation fails",
			videos:            sampleVideos,
			mockError:         fmt.Errorf("mock AI generation error"),
			wantErr:           true,
			expectedErrSubstr: "AI analysis generation failed",
		},
		{
			name: "Single video analysis",
			videos: []VideoABData{
				{
					Category:  "devops",
					Date:      "2024-01-01",
					DayOfWeek: "Monday",
					VideoID:   "video1",
					Titles:    []storage.TitleVariant{{Index: 1, Text: "Test Video", Share: 60.0}, {Index: 2, Text: "Alt Title", Share: 40.0}},
				},
			},
			mockResponse: `{
				"highPerformingPatterns": [],
				"lowPerformingPatterns": [],
				"recommendations": [],
				"titlesMdContent": "# Titles\n\nInsufficient data for meaningful analysis.\n\n{{.ManuscriptContent}}"
			}`,
			wantErr: false,
			validateResponse: func(t *testing.T, result TitleAnalysisResult) {
				if result.TitlesMDContent == "" {
					t.Error("Expected non-empty TitlesMDContent for single video analysis")
				}
			},
		},
		{
			name:   "Large dataset",
			videos: generateLargeABDataset(100),
			mockResponse: `{
				"highPerformingPatterns": [
					{
						"pattern": "Pattern from 100 videos",
						"description": "Large dataset analysis reveals trends",
						"impact": "Significant sample size",
						"examples": ["Video 1 (share: 65%)", "Video 2 (share: 70%)"]
					}
				],
				"lowPerformingPatterns": [],
				"recommendations": [
					{
						"recommendation": "Apply patterns from large dataset",
						"evidence": "100 videos analyzed with A/B data",
						"example": "Example based on data"
					}
				],
				"titlesMdContent": "# Title Generation\n\nBased on 100 videos:\n- Pattern A\n- Pattern B\n\n{{.ManuscriptContent}}"
			}`,
			wantErr: false,
			validateResponse: func(t *testing.T, result TitleAnalysisResult) {
				if len(result.HighPerformingPatterns) == 0 {
					t.Error("Expected at least one pattern from large dataset")
				}
				if len(result.Recommendations) == 0 {
					t.Error("Expected recommendations from large dataset")
				}
				if result.TitlesMDContent == "" {
					t.Error("Expected non-empty TitlesMDContent from large dataset")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockProvider{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Store original GetAIProvider function
			originalGetAIProvider := GetAIProvider
			defer func() { GetAIProvider = originalGetAIProvider }()

			// Mock the GetAIProvider function
			GetAIProvider = func() (AIProvider, error) {
				return mock, nil
			}

			gotAnalysis, _, err := AnalyzeTitles(ctx, tt.videos)

			if tt.wantErr {
				if err == nil {
					t.Errorf("AnalyzeTitles() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("AnalyzeTitles() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("AnalyzeTitles() unexpected error = %v", err)
					return
				}
				if tt.validateResponse != nil {
					tt.validateResponse(t, gotAnalysis)
				}
			}
		})
	}
}

func TestAnalyzeTitles_TemplateExecution(t *testing.T) {
	ctx := context.Background()

	videos := []VideoABData{
		{
			Category:  "devops",
			Date:      "2024-01-01",
			DayOfWeek: "Monday",
			VideoID:   "test1",
			Titles:    []storage.TitleVariant{{Index: 1, Text: "Test Title with Special Characters: <>&", Share: 55.0}, {Index: 2, Text: "Alt Title", Share: 45.0}},
		},
	}

	// Use a mock that returns valid JSON response
	validJSON := `{
		"highPerformingPatterns": [],
		"lowPerformingPatterns": [],
		"recommendations": [],
		"titlesMdContent": "# Titles\n\nSpecial characters handled.\n\n{{.ManuscriptContent}}"
	}`

	mockProvider := &MockProvider{
		response: validJSON,
		err:      nil,
	}

	// Store original GetAIProvider function
	originalGetAIProvider := GetAIProvider
	defer func() { GetAIProvider = originalGetAIProvider }()

	GetAIProvider = func() (AIProvider, error) {
		return mockProvider, nil
	}

	result, rawResponse, err := AnalyzeTitles(ctx, videos)
	if err != nil {
		t.Fatalf("AnalyzeTitles() unexpected error = %v", err)
	}

	// Verify we got valid results (template was successfully executed and AI returned data)
	if rawResponse == "" {
		t.Errorf("Expected non-empty rawResponse from AnalyzeTitles")
	}

	// Verify the result was properly parsed
	if result.TitlesMDContent == "" {
		t.Error("Expected non-empty TitlesMDContent in parsed result")
	}

	// Verify prompt was saved to the mock provider (contains the A/B data)
	if !strings.Contains(mockProvider.lastPrompt, "Test Title with Special Characters") {
		t.Error("Expected prompt to contain the title with special characters")
	}

	// Verify prompt contains A/B test format markers
	if !strings.Contains(mockProvider.lastPrompt, "A/B Test") {
		t.Error("Expected prompt to contain A/B Test section header")
	}
}

// Helper function to generate large A/B dataset for testing
func generateLargeABDataset(count int) []VideoABData {
	videos := make([]VideoABData, count)
	for i := 0; i < count; i++ {
		videos[i] = VideoABData{
			Category:            "devops",
			Date:                fmt.Sprintf("2024-01-%02dT10:00", (i%28)+1),
			DayOfWeek:           "Monday",
			VideoID:             fmt.Sprintf("video%d", i),
			Titles:              []storage.TitleVariant{{Index: 1, Text: fmt.Sprintf("Test Video %d", i), Share: 55.0 + float64(i%20)}, {Index: 2, Text: fmt.Sprintf("Alt Title %d", i), Share: 45.0 - float64(i%20)}},
			FirstWeekViews:      int64(1000 + i*100),

			FirstWeekLikes:      int64(50 + i*2),
			FirstWeekComments:   int64(10 + i),
			FirstWeekEngagement: float64(2+i%3) + 0.5,
			HasAnalytics:        true,
		}
	}
	return videos
}
