package ai

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/publishing"
)

func TestGenerateTimingRecommendations(t *testing.T) {
	ctx := context.Background()

	// Sample analytics data with timing information
	sampleAnalytics := []publishing.VideoAnalytics{
		{
			VideoID:             "video1",
			Title:               "Kubernetes Tutorial",
			Views:               50000,
			FirstWeekViews:      5000,
			FirstWeekCTR:        5.2,
			FirstWeekEngagement: 2.5,
			PublishedAt:         time.Date(2024, 1, 15, 16, 0, 0, 0, time.UTC),
		},
		{
			VideoID:             "video2",
			Title:               "Docker Guide",
			Views:               35000,
			FirstWeekViews:      3500,
			FirstWeekCTR:        4.8,
			FirstWeekEngagement: 2.1,
			PublishedAt:         time.Date(2024, 1, 22, 16, 0, 0, 0, time.UTC),
		},
		{
			VideoID:             "video3",
			Title:               "DevOps Best Practices",
			Views:               82000,
			FirstWeekViews:      8200,
			FirstWeekCTR:        6.1,
			FirstWeekEngagement: 3.2,
			PublishedAt:         time.Date(2024, 2, 5, 9, 0, 0, 0, time.UTC),
		},
	}

	validJSONResponse := `[
		{
			"day": "Monday",
			"time": "16:00",
			"reasoning": "Current baseline with strong performance. European end-of-workday timing shows consistent first-week metrics across multiple videos."
		},
		{
			"day": "Tuesday",
			"time": "09:00",
			"reasoning": "Testing morning slot to explore different audience availability patterns. Complements existing afternoon data."
		},
		{
			"day": "Wednesday",
			"time": "14:00",
			"reasoning": "Mid-week afternoon slot for experimental diversity. Provides temporal separation from morning and late afternoon slots."
		},
		{
			"day": "Thursday",
			"time": "21:00",
			"reasoning": "Evening slot to test after-work learning behavior. Explores whether evening publishing captures different audience segment."
		},
		{
			"day": "Friday",
			"time": "10:00",
			"reasoning": "End-of-week morning slot. Tests whether Friday differs from mid-week performance patterns."
		},
		{
			"day": "Saturday",
			"time": "11:00",
			"reasoning": "Weekend morning slot for comprehensive coverage. Evaluates whether weekend learning patterns exist for technical content."
		}
	]`

	tests := []struct {
		name              string
		analytics         []publishing.VideoAnalytics
		mockResponse      string
		mockError         error
		wantErr           bool
		expectedErrSubstr string
		validateResult    func(t *testing.T, recommendations []configuration.TimingRecommendation)
	}{
		{
			name:         "Successful recommendation generation",
			analytics:    sampleAnalytics,
			mockResponse: validJSONResponse,
			wantErr:      false,
			validateResult: func(t *testing.T, recommendations []configuration.TimingRecommendation) {
				if len(recommendations) != 6 {
					t.Errorf("Expected 6 recommendations, got %d", len(recommendations))
				}
				// Verify first recommendation
				if recommendations[0].Day != "Monday" {
					t.Errorf("Expected first recommendation day to be 'Monday', got '%s'", recommendations[0].Day)
				}
				if recommendations[0].Time != "16:00" {
					t.Errorf("Expected first recommendation time to be '16:00', got '%s'", recommendations[0].Time)
				}
				if len(recommendations[0].Reasoning) == 0 {
					t.Error("Expected non-empty reasoning")
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
			expectedErrSubstr: "AI returned empty timing recommendations",
		},
		{
			name:              "AI generation fails",
			analytics:         sampleAnalytics,
			mockError:         fmt.Errorf("mock AI generation error"),
			wantErr:           true,
			expectedErrSubstr: "AI timing analysis failed",
		},
		{
			name:      "JSON in markdown code block",
			analytics: sampleAnalytics,
			mockResponse: "```json\n" + validJSONResponse + "\n```",
			wantErr:   false,
			validateResult: func(t *testing.T, recommendations []configuration.TimingRecommendation) {
				if len(recommendations) != 6 {
					t.Errorf("Expected 6 recommendations, got %d", len(recommendations))
				}
			},
		},
		{
			name:      "JSON in plain markdown code block",
			analytics: sampleAnalytics,
			mockResponse: "```\n" + validJSONResponse + "\n```",
			wantErr:   false,
			validateResult: func(t *testing.T, recommendations []configuration.TimingRecommendation) {
				if len(recommendations) != 6 {
					t.Errorf("Expected 6 recommendations, got %d", len(recommendations))
				}
			},
		},
		{
			name:      "Too few recommendations (5)",
			analytics: sampleAnalytics,
			mockResponse: `[
				{"day": "Monday", "time": "16:00", "reasoning": "Test reasoning."},
				{"day": "Tuesday", "time": "09:00", "reasoning": "Test reasoning."},
				{"day": "Wednesday", "time": "14:00", "reasoning": "Test reasoning."},
				{"day": "Thursday", "time": "21:00", "reasoning": "Test reasoning."},
				{"day": "Friday", "time": "10:00", "reasoning": "Test reasoning."}
			]`,
			wantErr:           true,
			expectedErrSubstr: "expected 6-8 recommendations, got 5",
		},
		{
			name:      "Too many recommendations (9)",
			analytics: sampleAnalytics,
			mockResponse: `[
				{"day": "Monday", "time": "16:00", "reasoning": "Test reasoning."},
				{"day": "Tuesday", "time": "09:00", "reasoning": "Test reasoning."},
				{"day": "Wednesday", "time": "14:00", "reasoning": "Test reasoning."},
				{"day": "Thursday", "time": "21:00", "reasoning": "Test reasoning."},
				{"day": "Friday", "time": "10:00", "reasoning": "Test reasoning."},
				{"day": "Saturday", "time": "11:00", "reasoning": "Test reasoning."},
				{"day": "Sunday", "time": "12:00", "reasoning": "Test reasoning."},
				{"day": "Monday", "time": "08:00", "reasoning": "Test reasoning."},
				{"day": "Tuesday", "time": "15:00", "reasoning": "Test reasoning."}
			]`,
			wantErr:           true,
			expectedErrSubstr: "expected 6-8 recommendations, got 9",
		},
		{
			name:      "Invalid day name",
			analytics: sampleAnalytics,
			mockResponse: `[
				{"day": "Mon", "time": "16:00", "reasoning": "Test reasoning."},
				{"day": "Tuesday", "time": "09:00", "reasoning": "Test reasoning."},
				{"day": "Wednesday", "time": "14:00", "reasoning": "Test reasoning."},
				{"day": "Thursday", "time": "21:00", "reasoning": "Test reasoning."},
				{"day": "Friday", "time": "10:00", "reasoning": "Test reasoning."},
				{"day": "Saturday", "time": "11:00", "reasoning": "Test reasoning."}
			]`,
			wantErr:           true,
			expectedErrSubstr: "invalid day",
		},
		{
			name:      "Invalid time format",
			analytics: sampleAnalytics,
			mockResponse: `[
				{"day": "Monday", "time": "4pm", "reasoning": "Test reasoning."},
				{"day": "Tuesday", "time": "09:00", "reasoning": "Test reasoning."},
				{"day": "Wednesday", "time": "14:00", "reasoning": "Test reasoning."},
				{"day": "Thursday", "time": "21:00", "reasoning": "Test reasoning."},
				{"day": "Friday", "time": "10:00", "reasoning": "Test reasoning."},
				{"day": "Saturday", "time": "11:00", "reasoning": "Test reasoning."}
			]`,
			wantErr:           true,
			expectedErrSubstr: "invalid time format",
		},
		{
			name:      "Missing reasoning",
			analytics: sampleAnalytics,
			mockResponse: `[
				{"day": "Monday", "time": "16:00", "reasoning": ""},
				{"day": "Tuesday", "time": "09:00", "reasoning": "Test reasoning."},
				{"day": "Wednesday", "time": "14:00", "reasoning": "Test reasoning."},
				{"day": "Thursday", "time": "21:00", "reasoning": "Test reasoning."},
				{"day": "Friday", "time": "10:00", "reasoning": "Test reasoning."},
				{"day": "Saturday", "time": "11:00", "reasoning": "Test reasoning."}
			]`,
			wantErr:           true,
			expectedErrSubstr: "missing reasoning",
		},
		{
			name:              "Invalid JSON response",
			analytics:         sampleAnalytics,
			mockResponse:      "This is not JSON at all",
			wantErr:           true,
			expectedErrSubstr: "could not parse JSON",
		},
		{
			name:              "Malformed JSON",
			analytics:         sampleAnalytics,
			mockResponse:      `[{"day": "Monday", "time": "16:00"`,
			wantErr:           true,
			expectedErrSubstr: "could not parse JSON",
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

			gotRecommendations, _, _, err := GenerateTimingRecommendations(ctx, tt.analytics)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateTimingRecommendations() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("GenerateTimingRecommendations() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateTimingRecommendations() unexpected error = %v", err)
					return
				}
				if tt.validateResult != nil {
					tt.validateResult(t, gotRecommendations)
				}
			}
		})
	}
}

func TestIsValidTimeFormat(t *testing.T) {
	tests := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{"Valid time 00:00", "00:00", true},
		{"Valid time 09:00", "09:00", true},
		{"Valid time 16:00", "16:00", true},
		{"Valid time 23:59", "23:59", true},
		{"Invalid - too short", "9:00", false},
		{"Invalid - too long", "009:00", false},
		{"Invalid - no colon", "1600", false},
		{"Invalid - hour > 23", "24:00", false},
		{"Invalid - minute > 59", "16:60", false},
		{"Invalid - negative hour", "-1:00", false},
		{"Invalid - text", "invalid", false},
		{"Invalid - empty", "", false},
		{"Invalid - 12-hour format", "4pm", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidTimeFormat(tt.timeStr)
			if got != tt.expected {
				t.Errorf("isValidTimeFormat(%q) = %v, want %v", tt.timeStr, got, tt.expected)
			}
		})
	}
}

func TestExtractJSONFromMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "JSON in json code block",
			content:  "```json\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "JSON in plain code block",
			content:  "```\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "No code block",
			content:  "{\"key\": \"value\"}",
			expected: "",
		},
		{
			name:     "Incomplete code block",
			content:  "```json\n{\"key\": \"value\"}",
			expected: "",
		},
		{
			name:     "Multiple code blocks",
			content:  "```json\n{\"first\": \"block\"}\n```\nSome text\n```\n{\"second\": \"block\"}\n```",
			expected: "{\"first\": \"block\"}",
		},
		{
			name:     "Empty code block",
			content:  "```json\n\n```",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONFromMarkdown(tt.content)
			if got != tt.expected {
				t.Errorf("extractJSONFromMarkdown() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValidateRecommendations(t *testing.T) {
	tests := []struct {
		name              string
		recommendations   []configuration.TimingRecommendation
		wantErr           bool
		expectedErrSubstr string
	}{
		{
			name: "Valid 6 recommendations",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Monday", Time: "16:00", Reasoning: "Test reasoning 1"},
				{Day: "Tuesday", Time: "09:00", Reasoning: "Test reasoning 2"},
				{Day: "Wednesday", Time: "14:00", Reasoning: "Test reasoning 3"},
				{Day: "Thursday", Time: "21:00", Reasoning: "Test reasoning 4"},
				{Day: "Friday", Time: "10:00", Reasoning: "Test reasoning 5"},
				{Day: "Saturday", Time: "11:00", Reasoning: "Test reasoning 6"},
			},
			wantErr: false,
		},
		{
			name: "Valid 8 recommendations",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Monday", Time: "16:00", Reasoning: "Test reasoning 1"},
				{Day: "Tuesday", Time: "09:00", Reasoning: "Test reasoning 2"},
				{Day: "Wednesday", Time: "14:00", Reasoning: "Test reasoning 3"},
				{Day: "Thursday", Time: "21:00", Reasoning: "Test reasoning 4"},
				{Day: "Friday", Time: "10:00", Reasoning: "Test reasoning 5"},
				{Day: "Saturday", Time: "11:00", Reasoning: "Test reasoning 6"},
				{Day: "Sunday", Time: "12:00", Reasoning: "Test reasoning 7"},
				{Day: "Monday", Time: "08:00", Reasoning: "Test reasoning 8"},
			},
			wantErr: false,
		},
		{
			name: "Too few - 5 recommendations",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Monday", Time: "16:00", Reasoning: "Test reasoning 1"},
				{Day: "Tuesday", Time: "09:00", Reasoning: "Test reasoning 2"},
				{Day: "Wednesday", Time: "14:00", Reasoning: "Test reasoning 3"},
				{Day: "Thursday", Time: "21:00", Reasoning: "Test reasoning 4"},
				{Day: "Friday", Time: "10:00", Reasoning: "Test reasoning 5"},
			},
			wantErr:           true,
			expectedErrSubstr: "expected 6-8",
		},
		{
			name: "Invalid day name",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Mon", Time: "16:00", Reasoning: "Test reasoning 1"},
				{Day: "Tuesday", Time: "09:00", Reasoning: "Test reasoning 2"},
				{Day: "Wednesday", Time: "14:00", Reasoning: "Test reasoning 3"},
				{Day: "Thursday", Time: "21:00", Reasoning: "Test reasoning 4"},
				{Day: "Friday", Time: "10:00", Reasoning: "Test reasoning 5"},
				{Day: "Saturday", Time: "11:00", Reasoning: "Test reasoning 6"},
			},
			wantErr:           true,
			expectedErrSubstr: "invalid day",
		},
		{
			name: "Invalid time format",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Monday", Time: "16:00", Reasoning: "Test reasoning 1"},
				{Day: "Tuesday", Time: "9:00", Reasoning: "Test reasoning 2"},
				{Day: "Wednesday", Time: "14:00", Reasoning: "Test reasoning 3"},
				{Day: "Thursday", Time: "21:00", Reasoning: "Test reasoning 4"},
				{Day: "Friday", Time: "10:00", Reasoning: "Test reasoning 5"},
				{Day: "Saturday", Time: "11:00", Reasoning: "Test reasoning 6"},
			},
			wantErr:           true,
			expectedErrSubstr: "invalid time format",
		},
		{
			name: "Empty reasoning",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Monday", Time: "16:00", Reasoning: ""},
				{Day: "Tuesday", Time: "09:00", Reasoning: "Test reasoning 2"},
				{Day: "Wednesday", Time: "14:00", Reasoning: "Test reasoning 3"},
				{Day: "Thursday", Time: "21:00", Reasoning: "Test reasoning 4"},
				{Day: "Friday", Time: "10:00", Reasoning: "Test reasoning 5"},
				{Day: "Saturday", Time: "11:00", Reasoning: "Test reasoning 6"},
			},
			wantErr:           true,
			expectedErrSubstr: "missing reasoning",
		},
		{
			name: "Whitespace-only reasoning",
			recommendations: []configuration.TimingRecommendation{
				{Day: "Monday", Time: "16:00", Reasoning: "   "},
				{Day: "Tuesday", Time: "09:00", Reasoning: "Test reasoning 2"},
				{Day: "Wednesday", Time: "14:00", Reasoning: "Test reasoning 3"},
				{Day: "Thursday", Time: "21:00", Reasoning: "Test reasoning 4"},
				{Day: "Friday", Time: "10:00", Reasoning: "Test reasoning 5"},
				{Day: "Saturday", Time: "11:00", Reasoning: "Test reasoning 6"},
			},
			wantErr:           true,
			expectedErrSubstr: "missing reasoning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRecommendations(tt.recommendations)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateRecommendations() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("validateRecommendations() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("validateRecommendations() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestCalculateCurrentPattern(t *testing.T) {
	tests := []struct {
		name        string
		grouped     map[publishing.TimeSlot][]publishing.VideoAnalytics
		totalVideos int
		expected    int // expected number of summaries
	}{
		{
			name: "Single time slot",
			grouped: map[publishing.TimeSlot][]publishing.VideoAnalytics{
				{DayOfWeek: "Monday", TimeOfDay: "16:00"}: {
					{VideoID: "v1"},
					{VideoID: "v2"},
					{VideoID: "v3"},
				},
			},
			totalVideos: 3,
			expected:    1,
		},
		{
			name: "Multiple time slots",
			grouped: map[publishing.TimeSlot][]publishing.VideoAnalytics{
				{DayOfWeek: "Monday", TimeOfDay: "16:00"}: {
					{VideoID: "v1"},
					{VideoID: "v2"},
				},
				{DayOfWeek: "Tuesday", TimeOfDay: "09:00"}: {
					{VideoID: "v3"},
				},
				{DayOfWeek: "Wednesday", TimeOfDay: "14:00"}: {
					{VideoID: "v4"},
					{VideoID: "v5"},
					{VideoID: "v6"},
				},
			},
			totalVideos: 6,
			expected:    3,
		},
		{
			name:        "Empty grouped data",
			grouped:     map[publishing.TimeSlot][]publishing.VideoAnalytics{},
			totalVideos: 0,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCurrentPattern(tt.grouped, tt.totalVideos)
			if len(got) != tt.expected {
				t.Errorf("calculateCurrentPattern() returned %d summaries, want %d", len(got), tt.expected)
			}

			// Verify percentages add up to ~100% (allow for floating point rounding)
			totalPercentage := 0.0
			for _, summary := range got {
				totalPercentage += summary.Percentage
				// Verify count matches actual videos
				if videos, exists := tt.grouped[publishing.TimeSlot{DayOfWeek: summary.DayOfWeek, TimeOfDay: summary.TimeOfDay}]; exists {
					if summary.Count != len(videos) {
						t.Errorf("Summary count = %d, want %d for %s %s", summary.Count, len(videos), summary.DayOfWeek, summary.TimeOfDay)
					}
				}
			}

			if tt.totalVideos > 0 && len(got) > 0 {
				if totalPercentage < 99.9 || totalPercentage > 100.1 {
					t.Errorf("Total percentage = %.2f, want ~100.0", totalPercentage)
				}
			}
		})
	}
}

func TestParseTimingRecommendations(t *testing.T) {
	validJSON := `[
		{
			"day": "Monday",
			"time": "16:00",
			"reasoning": "Test reasoning"
		},
		{
			"day": "Tuesday",
			"time": "09:00",
			"reasoning": "Another test reasoning"
		}
	]`

	tests := []struct {
		name              string
		response          string
		wantErr           bool
		expectedCount     int
		expectedErrSubstr string
	}{
		{
			name:          "Valid direct JSON",
			response:      validJSON,
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name:          "JSON in markdown json block",
			response:      "```json\n" + validJSON + "\n```",
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name:          "JSON in markdown plain block",
			response:      "```\n" + validJSON + "\n```",
			wantErr:       false,
			expectedCount: 2,
		},
		{
			name:              "Invalid JSON",
			response:          "This is not JSON",
			wantErr:           true,
			expectedErrSubstr: "could not parse JSON",
		},
		{
			name:              "Malformed JSON",
			response:          `[{"day": "Monday"`,
			wantErr:           true,
			expectedErrSubstr: "could not parse JSON",
		},
		{
			name:              "Empty response",
			response:          "",
			wantErr:           true,
			expectedErrSubstr: "could not parse JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimingRecommendations(tt.response)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTimingRecommendations() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.expectedErrSubstr != "" && !strings.Contains(err.Error(), tt.expectedErrSubstr) {
					t.Errorf("parseTimingRecommendations() error = %q, want substring %q", err.Error(), tt.expectedErrSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("parseTimingRecommendations() unexpected error = %v", err)
					return
				}
				if len(got) != tt.expectedCount {
					t.Errorf("parseTimingRecommendations() returned %d recommendations, want %d", len(got), tt.expectedCount)
				}
			}
		})
	}
}
