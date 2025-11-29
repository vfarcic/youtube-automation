package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/publishing"
)

func TestAnalyzeTitles(t *testing.T) {
	ctx := context.Background()

	// Sample analytics data for testing
	sampleAnalytics := []publishing.VideoAnalytics{
		{
			VideoID:            "video1",
			Title:              "How to Deploy Kubernetes",
			Views:              50000,
			CTR:                5.2,
			AverageViewDuration: 420.5,
			Likes:              1200,
			Comments:           150,
			PublishedAt:        time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		},
		{
			VideoID:            "video2",
			Title:              "Docker vs Podman - Complete Comparison",
			Views:              35000,
			CTR:                4.8,
			AverageViewDuration: 380.0,
			Likes:              890,
			Comments:           95,
			PublishedAt:        time.Date(2024, 3, 20, 14, 30, 0, 0, time.UTC),
		},
		{
			VideoID:            "video3",
			Title:              "Top 5 DevOps Tools in 2024",
			Views:              82000,
			CTR:                6.1,
			AverageViewDuration: 510.2,
			Likes:              2100,
			Comments:           280,
			PublishedAt:        time.Date(2024, 2, 10, 9, 15, 0, 0, time.UTC),
		},
	}

	validJSONResponse := `{
		"highPerformingPatterns": [
			{
				"pattern": "Titles with numbers",
				"description": "Titles containing numbers perform significantly better",
				"impact": "40% more views on average",
				"examples": ["Top 5 DevOps Tools", "3 Ways to Deploy Kubernetes"]
			}
		],
		"lowPerformingPatterns": [],
		"titleLengthAnalysis": {
			"optimalRange": "50-65 characters",
			"finding": "Mid-length titles perform best",
			"data": "Average views: 50-65 chars = 45K, <50 chars = 32K, >65 chars = 38K"
		},
		"contentTypeAnalysis": {
			"finding": "Tutorial content outperforms news",
			"topPerformers": ["Tutorials", "Comparisons"],
			"data": "Tutorials avg 50K views, News avg 25K views"
		},
		"engagementPatterns": {
			"finding": "Question titles drive more comments",
			"likesPattern": "Specific outcomes get more likes",
			"commentsPattern": "Questions generate discussions",
			"watchTimePattern": "Comprehensive titles have higher retention"
		},
		"recommendations": [
			{
				"recommendation": "Include numbers in 30-40% of titles",
				"evidence": "Titles with numbers average 45% more views",
				"example": "Transform 'Kubernetes Guide' to 'Top 5 Kubernetes Best Practices'"
			}
		],
		"promptSuggestions": [
			"Include numbers in 30-40% of titles",
			"Keep titles between 50-65 characters"
		]
	}`

	tests := []struct {
		name              string
		analytics         []publishing.VideoAnalytics
		mockResponse      string
		mockError         error
		wantErr           bool
		expectedErrSubstr string
		validateResponse  func(t *testing.T, result TitleAnalysisResult)
	}{
		{
			name:         "Successful analysis with valid data",
			analytics:    sampleAnalytics,
			mockResponse: validJSONResponse,
			wantErr:      false,
			validateResponse: func(t *testing.T, result TitleAnalysisResult) {
				if len(result.HighPerformingPatterns) == 0 {
					t.Error("Expected at least one high-performing pattern")
				}
				if len(result.Recommendations) == 0 {
					t.Error("Expected at least one recommendation")
				}
			},
		},
		{
			name:              "Empty analytics data",
			analytics:         []publishing.VideoAnalytics{},
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "no analytics data provided",
		},
		{
			name:              "AI returns empty response",
			analytics:         sampleAnalytics,
			mockResponse:      "",
			wantErr:           true,
			expectedErrSubstr: "AI returned empty analysis",
		},
		{
			name:              "AI generation fails",
			analytics:         sampleAnalytics,
			mockError:         fmt.Errorf("mock AI generation error"),
			wantErr:           true,
			expectedErrSubstr: "AI analysis generation failed",
		},
		{
			name: "Single video analysis",
			analytics: []publishing.VideoAnalytics{
				{
					VideoID:            "video1",
					Title:              "Test Video",
					Views:              1000,
					AverageViewDuration: 200.0,
					Likes:              50,
					Comments:           10,
					PublishedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			mockResponse: `{
				"highPerformingPatterns": [],
				"lowPerformingPatterns": [],
				"titleLengthAnalysis": {
					"optimalRange": "N/A",
					"finding": "Limited data - single video only",
					"data": "Insufficient data for length analysis"
				},
				"contentTypeAnalysis": {
					"finding": "Single video - no comparison available",
					"topPerformers": [],
					"data": "N/A"
				},
				"engagementPatterns": {
					"finding": "Limited data available",
					"likesPattern": "N/A",
					"commentsPattern": "N/A",
					"watchTimePattern": "N/A"
				},
				"recommendations": [],
				"promptSuggestions": ["Need more videos for meaningful analysis"]
			}`,
			wantErr:      false,
			validateResponse: func(t *testing.T, result TitleAnalysisResult) {
				if result.TitleLengthAnalysis.Finding == "" {
					t.Error("Expected non-empty finding for single video analysis")
				}
				if len(result.PromptSuggestions) == 0 {
					t.Error("Expected at least one prompt suggestion")
				}
			},
		},
		{
			name:      "Large dataset",
			analytics: generateLargeAnalyticsDataset(100),
			mockResponse: `{
				"highPerformingPatterns": [
					{
						"pattern": "Pattern from 100 videos",
						"description": "Large dataset analysis reveals trends",
						"impact": "Significant sample size",
						"examples": ["Video 1", "Video 2"]
					}
				],
				"lowPerformingPatterns": [],
				"titleLengthAnalysis": {
					"optimalRange": "50-60 characters",
					"finding": "Clear pattern from large dataset",
					"data": "Analysis of 100 videos"
				},
				"contentTypeAnalysis": {
					"finding": "Strong patterns with 100 videos",
					"topPerformers": ["Type A", "Type B"],
					"data": "Large dataset provides confidence"
				},
				"engagementPatterns": {
					"finding": "Clear engagement trends",
					"likesPattern": "Pattern identified",
					"commentsPattern": "Pattern identified",
					"watchTimePattern": "Pattern identified"
				},
				"recommendations": [
					{
						"recommendation": "Apply patterns from large dataset",
						"evidence": "100 videos analyzed",
						"example": "Example based on data"
					}
				],
				"promptSuggestions": ["Pattern 1", "Pattern 2"]
			}`,
			wantErr: false,
			validateResponse: func(t *testing.T, result TitleAnalysisResult) {
				if len(result.HighPerformingPatterns) == 0 {
					t.Error("Expected at least one pattern from large dataset")
				}
				if len(result.Recommendations) == 0 {
					t.Error("Expected recommendations from large dataset")
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

			gotAnalysis, _, _, err := AnalyzeTitles(ctx, tt.analytics)

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

	analytics := []publishing.VideoAnalytics{
		{
			VideoID:            "test1",
			Title:              "Test Title with Special Characters: <>&",
			Views:              1000,
			AverageViewDuration: 100.0,
			Likes:              50,
			Comments:           10,
			PublishedAt:        time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	// Use a mock that returns valid JSON response
	validJSON := `{
		"highPerformingPatterns": [],
		"lowPerformingPatterns": [],
		"titleLengthAnalysis": {
			"optimalRange": "N/A",
			"finding": "Special characters test",
			"data": "Test data"
		},
		"contentTypeAnalysis": {
			"finding": "Template execution successful",
			"topPerformers": [],
			"data": "Test"
		},
		"engagementPatterns": {
			"finding": "Test",
			"likesPattern": "Test",
			"commentsPattern": "Test",
			"watchTimePattern": "Test"
		},
		"recommendations": [],
		"promptSuggestions": ["Test suggestion"]
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

	result, prompt, rawResponse, err := AnalyzeTitles(ctx, analytics)
	if err != nil {
		t.Fatalf("AnalyzeTitles() unexpected error = %v", err)
	}

	// Verify we got valid results (template was successfully executed and AI returned data)
	if prompt == "" {
		t.Errorf("Expected non-empty prompt from AnalyzeTitles")
	}
	if rawResponse == "" {
		t.Errorf("Expected non-empty rawResponse from AnalyzeTitles")
	}

	// Verify the result was properly parsed
	if result.TitleLengthAnalysis.Finding == "" {
		t.Error("Expected non-empty finding in parsed result")
	}

	// Verify special characters in the title were handled correctly in template
	if !strings.Contains(prompt, "Test Title with Special Characters") {
		t.Error("Expected prompt to contain the title with special characters")
	}
}

// Helper function to generate large analytics dataset for testing
func generateLargeAnalyticsDataset(count int) []publishing.VideoAnalytics {
	analytics := make([]publishing.VideoAnalytics, count)
	for i := 0; i < count; i++ {
		analytics[i] = publishing.VideoAnalytics{
			VideoID:            fmt.Sprintf("video%d", i),
			Title:              fmt.Sprintf("Test Video %d", i),
			Views:              int64(1000 + i*100),
			AverageViewDuration: float64(200 + i*5),
			Likes:              int64(50 + i*2),
			Comments:           int64(10 + i),
			PublishedAt:        time.Date(2024, 1, 1+i%28, 0, 0, 0, 0, time.UTC),
		}
	}
	return analytics
}
