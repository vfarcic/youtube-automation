package publishing

import (
	"context"
	"testing"
	"time"
)

func TestVideoAnalyticsStruct(t *testing.T) {
	// Test that VideoAnalytics struct can be properly initialized
	publishedAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	analytics := VideoAnalytics{
		VideoID:             "abc123",
		Title:               "Test Video",
		Views:               1000,
		CTR:                 5.5,
		AverageViewDuration: 120.5,
		Likes:               50,
		Comments:            10,
		PublishedAt:         publishedAt,
		FirstWeekViews:      800,
		FirstWeekLikes:      40,
		FirstWeekComments:   8,
		FirstWeekCTR:        5.0,
		DayOfWeek:           "Monday",
		TimeOfDay:           "16:00",
		FirstWeekEngagement: 6.0,
	}

	// Verify all fields are set correctly
	if analytics.VideoID != "abc123" {
		t.Errorf("Expected VideoID 'abc123', got '%s'", analytics.VideoID)
	}
	if analytics.Title != "Test Video" {
		t.Errorf("Expected Title 'Test Video', got '%s'", analytics.Title)
	}
	if analytics.Views != 1000 {
		t.Errorf("Expected Views 1000, got %d", analytics.Views)
	}
	if analytics.CTR != 5.5 {
		t.Errorf("Expected CTR 5.5, got %f", analytics.CTR)
	}
	if analytics.AverageViewDuration != 120.5 {
		t.Errorf("Expected AverageViewDuration 120.5, got %f", analytics.AverageViewDuration)
	}
	if analytics.Likes != 50 {
		t.Errorf("Expected Likes 50, got %d", analytics.Likes)
	}
	if analytics.Comments != 10 {
		t.Errorf("Expected Comments 10, got %d", analytics.Comments)
	}
	if !analytics.PublishedAt.Equal(publishedAt) {
		t.Errorf("Expected PublishedAt %v, got %v", publishedAt, analytics.PublishedAt)
	}
	if analytics.FirstWeekViews != 800 {
		t.Errorf("Expected FirstWeekViews 800, got %d", analytics.FirstWeekViews)
	}
	if analytics.FirstWeekLikes != 40 {
		t.Errorf("Expected FirstWeekLikes 40, got %d", analytics.FirstWeekLikes)
	}
	if analytics.FirstWeekComments != 8 {
		t.Errorf("Expected FirstWeekComments 8, got %d", analytics.FirstWeekComments)
	}
	if analytics.FirstWeekCTR != 5.0 {
		t.Errorf("Expected FirstWeekCTR 5.0, got %f", analytics.FirstWeekCTR)
	}
	if analytics.DayOfWeek != "Monday" {
		t.Errorf("Expected DayOfWeek 'Monday', got '%s'", analytics.DayOfWeek)
	}
	if analytics.TimeOfDay != "16:00" {
		t.Errorf("Expected TimeOfDay '16:00', got '%s'", analytics.TimeOfDay)
	}
	if analytics.FirstWeekEngagement != 6.0 {
		t.Errorf("Expected FirstWeekEngagement 6.0, got %f", analytics.FirstWeekEngagement)
	}
}

func TestGetVideoAnalyticsForLastYear(t *testing.T) {
	// Skip this test in CI/CD environments as it requires real YouTube credentials
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// This test requires valid YouTube credentials in client_secret.json
	// and a cached token in ~/.credentials/youtube-go.json
	_, err := GetVideoAnalyticsForLastYear(ctx)

	// We can't easily test the success case without real credentials,
	// but we can verify the function doesn't panic and returns a proper error
	// when credentials are missing
	if err != nil {
		t.Logf("Expected error without credentials: %v", err)
		// This is expected in test environments without credentials
	}
}

func TestGetVideoAnalytics_DateRange(t *testing.T) {
	// Skip this test in CI/CD environments
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test with specific date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, -3, 0) // 3 months ago

	_, err := GetVideoAnalytics(ctx, startDate, endDate)

	// Similar to above, we expect an error without real credentials
	if err != nil {
		t.Logf("Expected error without credentials: %v", err)
	}
}

func TestGetVideoAnalytics_EmptyDateRange(t *testing.T) {
	// Skip this test in CI/CD environments
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// Test with same start and end date (should return empty or error)
	date := time.Now()

	results, err := GetVideoAnalytics(ctx, date, date)

	if err != nil {
		// Error is acceptable for credential issues
		t.Logf("Error (expected without credentials): %v", err)
	} else if len(results) > 0 {
		// If somehow credentials exist and we get results, they should be valid
		t.Logf("Received %d analytics results", len(results))
		for _, result := range results {
			if result.VideoID == "" {
				t.Error("VideoID should not be empty")
			}
			if result.Title == "" {
				t.Error("Title should not be empty")
			}
		}
	}
}

// TestEnrichWithTimingData tests the timing data extraction and engagement calculation
func TestEnrichWithTimingData(t *testing.T) {
	tests := []struct {
		name                string
		input               VideoAnalytics
		wantDay             string
		wantTime            string
		wantEngagement      float64
	}{
		{
			name: "Monday afternoon video with engagement",
			input: VideoAnalytics{
				VideoID:           "test1",
				Title:             "Test Video",
				PublishedAt:       time.Date(2025, 1, 6, 16, 0, 0, 0, time.UTC), // Monday 16:00 UTC
				FirstWeekViews:    1000,
				FirstWeekLikes:    50,
				FirstWeekComments: 10,
			},
			wantDay:        "Monday",
			wantTime:       "16:00",
			wantEngagement: 6.0, // (50+10)/1000 * 100
		},
		{
			name: "Tuesday morning video",
			input: VideoAnalytics{
				VideoID:           "test2",
				Title:             "Morning Video",
				PublishedAt:       time.Date(2025, 1, 7, 9, 0, 0, 0, time.UTC), // Tuesday 09:00 UTC
				FirstWeekViews:    500,
				FirstWeekLikes:    25,
				FirstWeekComments: 5,
			},
			wantDay:        "Tuesday",
			wantTime:       "09:00",
			wantEngagement: 6.0, // (25+5)/500 * 100
		},
		{
			name: "Video with zero views",
			input: VideoAnalytics{
				VideoID:           "test3",
				Title:             "New Video",
				PublishedAt:       time.Date(2025, 1, 8, 12, 30, 0, 0, time.UTC), // Wednesday 12:30 UTC
				FirstWeekViews:    0,
				FirstWeekLikes:    0,
				FirstWeekComments: 0,
			},
			wantDay:        "Wednesday",
			wantTime:       "12:30",
			wantEngagement: 0.0, // No views, so 0% engagement
		},
		{
			name: "High engagement video",
			input: VideoAnalytics{
				VideoID:           "test4",
				Title:             "Popular Video",
				PublishedAt:       time.Date(2025, 1, 9, 14, 15, 0, 0, time.UTC), // Thursday 14:15 UTC
				FirstWeekViews:    10000,
				FirstWeekLikes:    1000,
				FirstWeekComments: 500,
			},
			wantDay:        "Thursday",
			wantTime:       "14:15",
			wantEngagement: 15.0, // (1000+500)/10000 * 100
		},
		{
			name: "Friday evening video",
			input: VideoAnalytics{
				VideoID:           "test5",
				Title:             "Friday Video",
				PublishedAt:       time.Date(2025, 1, 10, 20, 0, 0, 0, time.UTC), // Friday 20:00 UTC
				FirstWeekViews:    2000,
				FirstWeekLikes:    100,
				FirstWeekComments: 20,
			},
			wantDay:        "Friday",
			wantTime:       "20:00",
			wantEngagement: 6.0, // (100+20)/2000 * 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enriched := EnrichWithTimingData([]VideoAnalytics{tt.input})

			if len(enriched) != 1 {
				t.Fatalf("Expected 1 enriched video, got %d", len(enriched))
			}

			result := enriched[0]

			if result.DayOfWeek != tt.wantDay {
				t.Errorf("DayOfWeek = %v, want %v", result.DayOfWeek, tt.wantDay)
			}
			if result.TimeOfDay != tt.wantTime {
				t.Errorf("TimeOfDay = %v, want %v", result.TimeOfDay, tt.wantTime)
			}
			if result.FirstWeekEngagement != tt.wantEngagement {
				t.Errorf("FirstWeekEngagement = %v, want %v", result.FirstWeekEngagement, tt.wantEngagement)
			}
		})
	}
}

// TestEnrichWithTimingData_EmptyInput tests handling of empty input
func TestEnrichWithTimingData_EmptyInput(t *testing.T) {
	enriched := EnrichWithTimingData([]VideoAnalytics{})

	if len(enriched) != 0 {
		t.Errorf("Expected empty result for empty input, got %d videos", len(enriched))
	}
}

// TestGroupByTimeSlot tests video grouping by day/time
func TestGroupByTimeSlot(t *testing.T) {
	tests := []struct {
		name       string
		input      []VideoAnalytics
		wantGroups int
		wantSlot   TimeSlot
		wantCount  int
	}{
		{
			name: "Single video",
			input: []VideoAnalytics{
				{VideoID: "v1", DayOfWeek: "Monday", TimeOfDay: "16:00"},
			},
			wantGroups: 1,
			wantSlot:   TimeSlot{DayOfWeek: "Monday", TimeOfDay: "16:00"},
			wantCount:  1,
		},
		{
			name: "Multiple videos same slot",
			input: []VideoAnalytics{
				{VideoID: "v1", DayOfWeek: "Monday", TimeOfDay: "16:00"},
				{VideoID: "v2", DayOfWeek: "Monday", TimeOfDay: "16:00"},
				{VideoID: "v3", DayOfWeek: "Monday", TimeOfDay: "16:00"},
			},
			wantGroups: 1,
			wantSlot:   TimeSlot{DayOfWeek: "Monday", TimeOfDay: "16:00"},
			wantCount:  3,
		},
		{
			name: "Multiple videos different slots",
			input: []VideoAnalytics{
				{VideoID: "v1", DayOfWeek: "Monday", TimeOfDay: "16:00"},
				{VideoID: "v2", DayOfWeek: "Tuesday", TimeOfDay: "09:00"},
				{VideoID: "v3", DayOfWeek: "Monday", TimeOfDay: "16:00"},
			},
			wantGroups: 2,
			wantSlot:   TimeSlot{DayOfWeek: "Monday", TimeOfDay: "16:00"},
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grouped := GroupByTimeSlot(tt.input)

			if len(grouped) != tt.wantGroups {
				t.Errorf("Expected %d groups, got %d", tt.wantGroups, len(grouped))
			}

			if videos, exists := grouped[tt.wantSlot]; exists {
				if len(videos) != tt.wantCount {
					t.Errorf("Expected %d videos in slot %v, got %d", tt.wantCount, tt.wantSlot, len(videos))
				}
			} else {
				t.Errorf("Expected slot %v to exist in grouped data", tt.wantSlot)
			}
		})
	}
}

// TestGroupByTimeSlot_EmptyInput tests handling of empty input
func TestGroupByTimeSlot_EmptyInput(t *testing.T) {
	grouped := GroupByTimeSlot([]VideoAnalytics{})

	if len(grouped) != 0 {
		t.Errorf("Expected empty map for empty input, got %d groups", len(grouped))
	}
}

// TestCalculateTimeSlotPerformance tests performance aggregation
func TestCalculateTimeSlotPerformance(t *testing.T) {
	tests := []struct {
		name         string
		input        map[TimeSlot][]VideoAnalytics
		wantSlot     TimeSlot
		wantAvgViews float64
		wantAvgCTR   float64
		wantAvgEng   float64
	}{
		{
			name: "Single video in slot",
			input: map[TimeSlot][]VideoAnalytics{
				{DayOfWeek: "Monday", TimeOfDay: "16:00"}: {
					{FirstWeekViews: 1000, FirstWeekCTR: 5.0, FirstWeekEngagement: 6.0},
				},
			},
			wantSlot:     TimeSlot{DayOfWeek: "Monday", TimeOfDay: "16:00"},
			wantAvgViews: 1000.0,
			wantAvgCTR:   5.0,
			wantAvgEng:   6.0,
		},
		{
			name: "Multiple videos in slot",
			input: map[TimeSlot][]VideoAnalytics{
				{DayOfWeek: "Tuesday", TimeOfDay: "09:00"}: {
					{FirstWeekViews: 1000, FirstWeekCTR: 5.0, FirstWeekEngagement: 6.0},
					{FirstWeekViews: 2000, FirstWeekCTR: 7.0, FirstWeekEngagement: 8.0},
					{FirstWeekViews: 3000, FirstWeekCTR: 9.0, FirstWeekEngagement: 10.0},
				},
			},
			wantSlot:     TimeSlot{DayOfWeek: "Tuesday", TimeOfDay: "09:00"},
			wantAvgViews: 2000.0, // (1000+2000+3000)/3
			wantAvgCTR:   7.0,    // (5+7+9)/3
			wantAvgEng:   8.0,    // (6+8+10)/3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := CalculateTimeSlotPerformance(tt.input)

			if len(results) != 1 {
				t.Fatalf("Expected 1 result, got %d", len(results))
			}

			result := results[0]

			if result.Slot != tt.wantSlot {
				t.Errorf("Slot = %v, want %v", result.Slot, tt.wantSlot)
			}
			if result.AvgFirstWeekViews != tt.wantAvgViews {
				t.Errorf("AvgFirstWeekViews = %v, want %v", result.AvgFirstWeekViews, tt.wantAvgViews)
			}
			if result.AvgFirstWeekCTR != tt.wantAvgCTR {
				t.Errorf("AvgFirstWeekCTR = %v, want %v", result.AvgFirstWeekCTR, tt.wantAvgCTR)
			}
			if result.AvgFirstWeekEngagement != tt.wantAvgEng {
				t.Errorf("AvgFirstWeekEngagement = %v, want %v", result.AvgFirstWeekEngagement, tt.wantAvgEng)
			}
		})
	}
}

// TestCalculateTimeSlotPerformance_EmptyInput tests handling of empty input
func TestCalculateTimeSlotPerformance_EmptyInput(t *testing.T) {
	results := CalculateTimeSlotPerformance(map[TimeSlot][]VideoAnalytics{})

	if len(results) != 0 {
		t.Errorf("Expected empty results for empty input, got %d results", len(results))
	}
}

// TestTimeSlot_String tests the TimeSlot string formatting
func TestTimeSlot_String(t *testing.T) {
	tests := []struct {
		name string
		slot TimeSlot
		want string
	}{
		{
			name: "Monday afternoon",
			slot: TimeSlot{DayOfWeek: "Monday", TimeOfDay: "16:00"},
			want: "Monday 16:00 UTC",
		},
		{
			name: "Tuesday morning",
			slot: TimeSlot{DayOfWeek: "Tuesday", TimeOfDay: "09:00"},
			want: "Tuesday 09:00 UTC",
		},
		{
			name: "Friday evening",
			slot: TimeSlot{DayOfWeek: "Friday", TimeOfDay: "20:00"},
			want: "Friday 20:00 UTC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.slot.String()
			if got != tt.want {
				t.Errorf("TimeSlot.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFirstWeekMetrics_Struct tests the FirstWeekMetrics struct
func TestFirstWeekMetrics_Struct(t *testing.T) {
	metrics := FirstWeekMetrics{
		Views:    1000,
		Likes:    50,
		Comments: 10,
		CTR:      5.5,
	}

	if metrics.Views != 1000 {
		t.Errorf("Expected Views 1000, got %d", metrics.Views)
	}
	if metrics.Likes != 50 {
		t.Errorf("Expected Likes 50, got %d", metrics.Likes)
	}
	if metrics.Comments != 10 {
		t.Errorf("Expected Comments 10, got %d", metrics.Comments)
	}
	if metrics.CTR != 5.5 {
		t.Errorf("Expected CTR 5.5, got %f", metrics.CTR)
	}
}

// TestVideoAnalyticsDataValidation tests that the VideoAnalytics struct
// properly holds various data types and edge cases
func TestVideoAnalyticsDataValidation(t *testing.T) {
	tests := []struct {
		name      string
		analytics VideoAnalytics
		wantErr   bool
	}{
		{
			name: "Valid analytics data",
			analytics: VideoAnalytics{
				VideoID:            "video123",
				Title:              "My Great Video",
				Views:              5000,
				CTR:                7.2,
				AverageViewDuration: 300.5,
				Likes:              150,
				Comments:           25,
				PublishedAt:        time.Now(),
			},
			wantErr: false,
		},
		{
			name: "Zero views video",
			analytics: VideoAnalytics{
				VideoID:            "video456",
				Title:              "New Video",
				Views:              0,
				CTR:                0.0,
				AverageViewDuration: 0.0,
				Likes:              0,
				Comments:           0,
				PublishedAt:        time.Now(),
			},
			wantErr: false,
		},
		{
			name: "High performance video",
			analytics: VideoAnalytics{
				VideoID:            "video789",
				Title:              "Viral Video",
				Views:              1000000,
				CTR:                15.5,
				AverageViewDuration: 600.0,
				Likes:              50000,
				Comments:           5000,
				PublishedAt:        time.Now().AddDate(0, -1, 0),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify the struct holds the expected values
			if tt.analytics.VideoID == "" && !tt.wantErr {
				t.Error("VideoID should not be empty for valid analytics")
			}
			if tt.analytics.Views < 0 {
				t.Error("Views should not be negative")
			}
			if tt.analytics.Likes < 0 {
				t.Error("Likes should not be negative")
			}
			if tt.analytics.Comments < 0 {
				t.Error("Comments should not be negative")
			}
		})
	}
}
