package publishing

import (
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

func TestIsShort(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		want     bool
	}{
		// Shorts (≤ 60 seconds)
		{name: "30 seconds", duration: "PT30S", want: true},
		{name: "60 seconds exactly", duration: "PT60S", want: true},
		{name: "45 seconds", duration: "PT45S", want: true},
		{name: "1 minute (60s)", duration: "PT1M", want: true},
		{name: "59 seconds", duration: "PT59S", want: true},

		// Not Shorts (> 60 seconds)
		{name: "61 seconds", duration: "PT61S", want: false},
		{name: "1 minute 1 second", duration: "PT1M1S", want: false},
		{name: "2 minutes", duration: "PT2M", want: false},
		{name: "5 minutes 30 seconds", duration: "PT5M30S", want: false},
		{name: "10 minutes", duration: "PT10M", want: false},
		{name: "15 minutes 42 seconds", duration: "PT15M42S", want: false},
		{name: "1 hour", duration: "PT1H", want: false},
		{name: "1 hour 2 minutes 3 seconds", duration: "PT1H2M3S", want: false},

		// Edge cases
		{name: "empty string", duration: "", want: false},
		{name: "0 seconds", duration: "PT0S", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isShort(tt.duration)
			if got != tt.want {
				t.Errorf("isShort(%q) = %v, want %v", tt.duration, got, tt.want)
			}
		})
	}
}

// TestChannelDemographicsStruct tests that ChannelDemographics struct can be properly initialized
func TestChannelDemographicsStruct(t *testing.T) {
	demographics := ChannelDemographics{
		AgeGroups: []AgeGroupData{
			{AgeGroup: "age18-24", Percentage: 15.5},
			{AgeGroup: "age25-34", Percentage: 35.2},
			{AgeGroup: "age35-44", Percentage: 25.0},
			{AgeGroup: "age45-54", Percentage: 15.0},
			{AgeGroup: "age55-64", Percentage: 7.3},
			{AgeGroup: "age65-", Percentage: 2.0},
		},
		Gender: []GenderData{
			{Gender: "male", Percentage: 85.0},
			{Gender: "female", Percentage: 14.0},
			{Gender: "user_specified", Percentage: 1.0},
		},
	}

	// Verify age groups
	if len(demographics.AgeGroups) != 6 {
		t.Errorf("Expected 6 age groups, got %d", len(demographics.AgeGroups))
	}

	// Verify first age group
	if demographics.AgeGroups[0].AgeGroup != "age18-24" {
		t.Errorf("Expected first age group 'age18-24', got '%s'", demographics.AgeGroups[0].AgeGroup)
	}
	if demographics.AgeGroups[0].Percentage != 15.5 {
		t.Errorf("Expected first age group percentage 15.5, got %f", demographics.AgeGroups[0].Percentage)
	}

	// Verify gender data
	if len(demographics.Gender) != 3 {
		t.Errorf("Expected 3 gender entries, got %d", len(demographics.Gender))
	}

	// Verify first gender entry
	if demographics.Gender[0].Gender != "male" {
		t.Errorf("Expected first gender 'male', got '%s'", demographics.Gender[0].Gender)
	}
	if demographics.Gender[0].Percentage != 85.0 {
		t.Errorf("Expected male percentage 85.0, got %f", demographics.Gender[0].Percentage)
	}
}

// TestAgeGroupDataStruct tests the AgeGroupData struct
func TestAgeGroupDataStruct(t *testing.T) {
	tests := []struct {
		name       string
		ageGroup   string
		percentage float64
	}{
		{name: "age13-17", ageGroup: "age13-17", percentage: 5.0},
		{name: "age18-24", ageGroup: "age18-24", percentage: 20.0},
		{name: "age25-34", ageGroup: "age25-34", percentage: 35.0},
		{name: "age35-44", ageGroup: "age35-44", percentage: 25.0},
		{name: "age45-54", ageGroup: "age45-54", percentage: 10.0},
		{name: "age55-64", ageGroup: "age55-64", percentage: 4.0},
		{name: "age65-", ageGroup: "age65-", percentage: 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := AgeGroupData{
				AgeGroup:   tt.ageGroup,
				Percentage: tt.percentage,
			}

			if data.AgeGroup != tt.ageGroup {
				t.Errorf("AgeGroup = %v, want %v", data.AgeGroup, tt.ageGroup)
			}
			if data.Percentage != tt.percentage {
				t.Errorf("Percentage = %v, want %v", data.Percentage, tt.percentage)
			}
		})
	}
}

// TestGenderDataStruct tests the GenderData struct
func TestGenderDataStruct(t *testing.T) {
	tests := []struct {
		name       string
		gender     string
		percentage float64
	}{
		{name: "male", gender: "male", percentage: 85.0},
		{name: "female", gender: "female", percentage: 14.0},
		{name: "user_specified", gender: "user_specified", percentage: 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := GenderData{
				Gender:     tt.gender,
				Percentage: tt.percentage,
			}

			if data.Gender != tt.gender {
				t.Errorf("Gender = %v, want %v", data.Gender, tt.gender)
			}
			if data.Percentage != tt.percentage {
				t.Errorf("Percentage = %v, want %v", data.Percentage, tt.percentage)
			}
		})
	}
}

// TestGeographicDistributionStruct tests that GeographicDistribution struct can be properly initialized
func TestGeographicDistributionStruct(t *testing.T) {
	distribution := GeographicDistribution{
		Countries: []CountryData{
			{CountryCode: "US", Views: 500000, Percentage: 35.0},
			{CountryCode: "IN", Views: 300000, Percentage: 21.0},
			{CountryCode: "GB", Views: 150000, Percentage: 10.5},
			{CountryCode: "DE", Views: 100000, Percentage: 7.0},
			{CountryCode: "CA", Views: 80000, Percentage: 5.6},
		},
	}

	if len(distribution.Countries) != 5 {
		t.Errorf("Expected 5 countries, got %d", len(distribution.Countries))
	}

	// Verify first country (US)
	if distribution.Countries[0].CountryCode != "US" {
		t.Errorf("Expected first country 'US', got '%s'", distribution.Countries[0].CountryCode)
	}
	if distribution.Countries[0].Views != 500000 {
		t.Errorf("Expected US views 500000, got %d", distribution.Countries[0].Views)
	}
	if distribution.Countries[0].Percentage != 35.0 {
		t.Errorf("Expected US percentage 35.0, got %f", distribution.Countries[0].Percentage)
	}
}

// TestCountryDataStruct tests the CountryData struct
func TestCountryDataStruct(t *testing.T) {
	tests := []struct {
		name        string
		countryCode string
		views       int64
		percentage  float64
	}{
		{name: "United States", countryCode: "US", views: 500000, percentage: 35.0},
		{name: "India", countryCode: "IN", views: 300000, percentage: 21.0},
		{name: "United Kingdom", countryCode: "GB", views: 150000, percentage: 10.5},
		{name: "Germany", countryCode: "DE", views: 100000, percentage: 7.0},
		{name: "Canada", countryCode: "CA", views: 80000, percentage: 5.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := CountryData{
				CountryCode: tt.countryCode,
				Views:       tt.views,
				Percentage:  tt.percentage,
			}

			if data.CountryCode != tt.countryCode {
				t.Errorf("CountryCode = %v, want %v", data.CountryCode, tt.countryCode)
			}
			if data.Views != tt.views {
				t.Errorf("Views = %v, want %v", data.Views, tt.views)
			}
			if data.Percentage != tt.percentage {
				t.Errorf("Percentage = %v, want %v", data.Percentage, tt.percentage)
			}
		})
	}
}

// TestChannelStatisticsStruct tests that ChannelStatistics struct can be properly initialized
func TestChannelStatisticsStruct(t *testing.T) {
	tests := []struct {
		name              string
		subscriberCount   int64
		totalViews        int64
		videoCount        int64
		hiddenSubscribers bool
	}{
		{
			name:              "Normal channel",
			subscriberCount:   250000,
			totalViews:        50000000,
			videoCount:        500,
			hiddenSubscribers: false,
		},
		{
			name:              "Large channel",
			subscriberCount:   1000000,
			totalViews:        500000000,
			videoCount:        1500,
			hiddenSubscribers: false,
		},
		{
			name:              "Hidden subscribers",
			subscriberCount:   0,
			totalViews:        10000000,
			videoCount:        200,
			hiddenSubscribers: true,
		},
		{
			name:              "New channel",
			subscriberCount:   100,
			totalViews:        5000,
			videoCount:        10,
			hiddenSubscribers: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := ChannelStatistics{
				SubscriberCount:   tt.subscriberCount,
				TotalViews:        tt.totalViews,
				VideoCount:        tt.videoCount,
				HiddenSubscribers: tt.hiddenSubscribers,
			}

			if stats.SubscriberCount != tt.subscriberCount {
				t.Errorf("SubscriberCount = %v, want %v", stats.SubscriberCount, tt.subscriberCount)
			}
			if stats.TotalViews != tt.totalViews {
				t.Errorf("TotalViews = %v, want %v", stats.TotalViews, tt.totalViews)
			}
			if stats.VideoCount != tt.videoCount {
				t.Errorf("VideoCount = %v, want %v", stats.VideoCount, tt.videoCount)
			}
			if stats.HiddenSubscribers != tt.hiddenSubscribers {
				t.Errorf("HiddenSubscribers = %v, want %v", stats.HiddenSubscribers, tt.hiddenSubscribers)
			}
		})
	}
}

// TestChannelDemographicsEmptyData tests handling of empty demographics data
func TestChannelDemographicsEmptyData(t *testing.T) {
	demographics := ChannelDemographics{
		AgeGroups: []AgeGroupData{},
		Gender:    []GenderData{},
	}

	if len(demographics.AgeGroups) != 0 {
		t.Errorf("Expected 0 age groups, got %d", len(demographics.AgeGroups))
	}
	if len(demographics.Gender) != 0 {
		t.Errorf("Expected 0 gender entries, got %d", len(demographics.Gender))
	}
}

// TestGeographicDistributionEmptyData tests handling of empty geographic data
func TestGeographicDistributionEmptyData(t *testing.T) {
	distribution := GeographicDistribution{
		Countries: []CountryData{},
	}

	if len(distribution.Countries) != 0 {
		t.Errorf("Expected 0 countries, got %d", len(distribution.Countries))
	}
}

// TestCountryDataPercentageCalculation tests that percentage values are valid
func TestCountryDataPercentageCalculation(t *testing.T) {
	// Simulate a realistic distribution where percentages should sum to ~100%
	countries := []CountryData{
		{CountryCode: "US", Views: 350000, Percentage: 35.0},
		{CountryCode: "IN", Views: 200000, Percentage: 20.0},
		{CountryCode: "GB", Views: 150000, Percentage: 15.0},
		{CountryCode: "DE", Views: 100000, Percentage: 10.0},
		{CountryCode: "CA", Views: 80000, Percentage: 8.0},
		{CountryCode: "AU", Views: 50000, Percentage: 5.0},
		{CountryCode: "FR", Views: 40000, Percentage: 4.0},
		{CountryCode: "NL", Views: 30000, Percentage: 3.0},
	}

	var totalPercentage float64
	for _, country := range countries {
		if country.Percentage < 0 || country.Percentage > 100 {
			t.Errorf("Invalid percentage for %s: %f (should be 0-100)", country.CountryCode, country.Percentage)
		}
		totalPercentage += country.Percentage
	}

	// Total percentage should be close to 100% (allowing for rounding)
	if totalPercentage < 99.0 || totalPercentage > 101.0 {
		t.Errorf("Total percentage = %f, expected close to 100%%", totalPercentage)
	}
}

// TestEngagementMetricsStruct tests that EngagementMetrics struct can be properly initialized
func TestEngagementMetricsStruct(t *testing.T) {
	tests := []struct {
		name                string
		averageViewDuration float64
		likes               int64
		comments            int64
		shares              int64
		views               int64
		videoCount          int64
	}{
		{
			name:                "Normal engagement",
			averageViewDuration: 330.5,
			likes:               50000,
			comments:            5000,
			shares:              2000,
			views:               1000000,
			videoCount:          100,
		},
		{
			name:                "High engagement",
			averageViewDuration: 600.0,
			likes:               200000,
			comments:            50000,
			shares:              25000,
			views:               5000000,
			videoCount:          200,
		},
		{
			name:                "New channel low engagement",
			averageViewDuration: 120.0,
			likes:               100,
			comments:            10,
			shares:              5,
			views:               1000,
			videoCount:          5,
		},
		{
			name:                "Zero values",
			averageViewDuration: 0,
			likes:               0,
			comments:            0,
			shares:              0,
			views:               0,
			videoCount:          0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := EngagementMetrics{
				AverageViewDuration: tt.averageViewDuration,
				Likes:               tt.likes,
				Comments:            tt.comments,
				Shares:              tt.shares,
				Views:               tt.views,
				VideoCount:          tt.videoCount,
			}

			if metrics.AverageViewDuration != tt.averageViewDuration {
				t.Errorf("AverageViewDuration = %v, want %v", metrics.AverageViewDuration, tt.averageViewDuration)
			}
			if metrics.Likes != tt.likes {
				t.Errorf("Likes = %v, want %v", metrics.Likes, tt.likes)
			}
			if metrics.Comments != tt.comments {
				t.Errorf("Comments = %v, want %v", metrics.Comments, tt.comments)
			}
			if metrics.Shares != tt.shares {
				t.Errorf("Shares = %v, want %v", metrics.Shares, tt.shares)
			}
			if metrics.Views != tt.views {
				t.Errorf("Views = %v, want %v", metrics.Views, tt.views)
			}
			if metrics.VideoCount != tt.videoCount {
				t.Errorf("VideoCount = %v, want %v", metrics.VideoCount, tt.videoCount)
			}
		})
	}
}

// TestEngagementMetricsEngagementRateCalculation tests engagement rate calculation logic
func TestEngagementMetricsEngagementRateCalculation(t *testing.T) {
	tests := []struct {
		name           string
		likes          int64
		comments       int64
		views          int64
		wantRate       float64
		wantCalculable bool
	}{
		{
			name:           "Normal engagement rate",
			likes:          50000,
			comments:       5000,
			views:          1000000,
			wantRate:       5.5,
			wantCalculable: true,
		},
		{
			name:           "High engagement rate",
			likes:          10000,
			comments:       5000,
			views:          100000,
			wantRate:       15.0,
			wantCalculable: true,
		},
		{
			name:           "Low engagement rate",
			likes:          100,
			comments:       10,
			views:          100000,
			wantRate:       0.11,
			wantCalculable: true,
		},
		{
			name:           "Zero views - cannot calculate",
			likes:          100,
			comments:       10,
			views:          0,
			wantRate:       0,
			wantCalculable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := EngagementMetrics{
				Likes:    tt.likes,
				Comments: tt.comments,
				Views:    tt.views,
			}

			// Test engagement rate calculation logic
			if tt.wantCalculable {
				calculatedRate := float64(metrics.Likes+metrics.Comments) / float64(metrics.Views) * 100
				if calculatedRate != tt.wantRate {
					t.Errorf("Engagement rate = %v, want %v", calculatedRate, tt.wantRate)
				}
			} else {
				// Verify we can detect when calculation is not possible
				if metrics.Views != 0 {
					t.Error("Expected Views to be 0 for non-calculable case")
				}
			}
		})
	}
}
