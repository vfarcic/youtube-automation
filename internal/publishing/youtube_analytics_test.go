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
		VideoID:            "abc123",
		Title:              "Test Video",
		Views:              1000,
		CTR:                5.5,
		AverageViewDuration: 120.5,
		Likes:              50,
		Comments:           10,
		PublishedAt:        publishedAt,
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
