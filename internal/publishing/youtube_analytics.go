package publishing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/youtubeanalytics/v2"
)

// VideoAnalytics holds performance metrics for a single video
type VideoAnalytics struct {
	VideoID            string
	Title              string
	Views              int64   // Total cumulative views
	CTR                float64 // Click-through rate (percentage)
	AverageViewDuration float64 // In seconds
	Likes              int64   // Total cumulative likes
	Comments           int64   // Total cumulative comments
	PublishedAt        time.Time

	// First-week performance metrics (populated by GetFirstWeekMetrics)
	// These provide apples-to-apples comparison across videos regardless of age
	FirstWeekViews    int64   // Views in days 0-7 after publish
	FirstWeekLikes    int64   // Likes in first week
	FirstWeekComments int64   // Comments in first week
	FirstWeekCTR      float64 // CTR in first week

	// Computed timing fields (populated by EnrichWithTimingData)
	DayOfWeek           string  // "Monday", "Tuesday", etc.
	TimeOfDay           string  // "16:00", "09:00" (UTC, 24-hour format HH:MM)
	FirstWeekEngagement float64 // (FirstWeekLikes + FirstWeekComments) / FirstWeekViews * 100 (as percentage)
}

// GetVideoAnalytics fetches video performance data from YouTube Analytics API
// for videos published between startDate and endDate.
//
// Returns:
//   - []VideoAnalytics: Array of video analytics data
//   - error: Any error encountered during the API calls
func GetVideoAnalytics(ctx context.Context, startDate, endDate time.Time) ([]VideoAnalytics, error) {
	client := getClient(ctx)

	// Initialize YouTube Data API service (for video titles)
	youtubeService, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Initialize YouTube Analytics API service
	analyticsService, err := youtubeanalytics.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube Analytics service: %w", err)
	}

	// Format dates for API request (YYYY-MM-DD)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Get channel ID from settings
	channelID := configuration.GlobalSettings.YouTube.ChannelId
	if channelID == "" {
		return nil, fmt.Errorf("YouTube channel ID not configured in settings.yaml")
	}

	// Fetch analytics data
	// Metrics: views, averageViewDuration, likes, comments, cardClickRate
	// Dimensions: video (group by video ID)
	// Note: cardClickRate represents CTR (click-through rate from impressions)
	// MaxResults: 200 top-performing videos provides sufficient sample for timing analysis
	analyticsCall := analyticsService.Reports.Query().
		Ids("channel=="+channelID).
		StartDate(startDateStr).
		EndDate(endDateStr).
		Metrics("views,estimatedMinutesWatched,likes,comments,averageViewDuration,cardClickRate").
		Dimensions("video").
		Sort("-views").
		MaxResults(200)

	analyticsResponse, err := analyticsCall.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analytics data: %w", err)
	}

	// Check if we have any data
	if analyticsResponse.Rows == nil || len(analyticsResponse.Rows) == 0 {
		return []VideoAnalytics{}, nil
	}

	// Extract video IDs to fetch titles and metadata
	videoIDs := make([]string, 0, len(analyticsResponse.Rows))
	for _, row := range analyticsResponse.Rows {
		if len(row) > 0 {
			videoID, ok := row[0].(string)
			if ok {
				videoIDs = append(videoIDs, videoID)
			}
		}
	}

	// Fetch video metadata (titles, publish dates, broadcast status, and duration) from YouTube Data API
	videoMetadata := make(map[string]struct {
		Title               string
		PublishedAt         time.Time
		LiveBroadcastContent string
		Duration            string // ISO 8601 duration format (e.g., "PT15M30S")
	})

	// YouTube Data API allows up to 50 video IDs per request
	// Split into chunks if needed
	for i := 0; i < len(videoIDs); i += 50 {
		end := i + 50
		if end > len(videoIDs) {
			end = len(videoIDs)
		}
		chunk := videoIDs[i:end]

		videosCall := youtubeService.Videos.List([]string{"snippet", "contentDetails"}).Id(chunk...)
		videosResponse, err := videosCall.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
		}

		for _, video := range videosResponse.Items {
			publishedAt, _ := time.Parse(time.RFC3339, video.Snippet.PublishedAt)
			videoMetadata[video.Id] = struct {
				Title               string
				PublishedAt         time.Time
				LiveBroadcastContent string
				Duration            string
			}{
				Title:               video.Snippet.Title,
				PublishedAt:         publishedAt,
				LiveBroadcastContent: video.Snippet.LiveBroadcastContent,
				Duration:            video.ContentDetails.Duration,
			}
		}
	}

	// Parse analytics response and combine with video metadata
	results := make([]VideoAnalytics, 0, len(analyticsResponse.Rows))

	for _, row := range analyticsResponse.Rows {
		if len(row) < 7 {
			continue // Skip malformed rows (need 7 fields now)
		}

		videoID, _ := row[0].(string)
		views := int64(row[1].(float64))
		// estimatedMinutesWatched := row[2].(float64) // Not currently used
		likes := int64(row[3].(float64))
		comments := int64(row[4].(float64))
		avgViewDuration := row[5].(float64)
		ctr := row[6].(float64) // cardClickRate (CTR percentage)

		// Get video metadata
		metadata, exists := videoMetadata[videoID]
		if !exists {
			// Skip videos without metadata (shouldn't happen)
			continue
		}

		// Skip live streams and premieres - they have different performance characteristics
		// LiveBroadcastContent values: "none" (regular video), "live", "upcoming", "completed"
		if metadata.LiveBroadcastContent != "none" {
			continue
		}

		// Skip videos published before the start date
		// Note: YouTube Analytics API filters by view dates, not publish dates
		if metadata.PublishedAt.Before(startDate) {
			continue
		}

		// Skip Shorts (videos ≤ 60 seconds) - they have different performance characteristics
		if isShort(metadata.Duration) {
			continue
		}

		results = append(results, VideoAnalytics{
			VideoID:            videoID,
			Title:              metadata.Title,
			Views:              views,
			CTR:                ctr,
			AverageViewDuration: avgViewDuration,
			Likes:              likes,
			Comments:           comments,
			PublishedAt:        metadata.PublishedAt,
		})
	}

	return results, nil
}

// GetVideoAnalyticsForLastYear is a convenience function that fetches
// video analytics for the last 365 days.
//
// Returns:
//   - []VideoAnalytics: Array of video analytics data
//   - error: Any error encountered during the API call
func GetVideoAnalyticsForLastYear(ctx context.Context) ([]VideoAnalytics, error) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -365) // 365 days ago
	return GetVideoAnalytics(ctx, startDate, endDate)
}

// FirstWeekMetrics holds performance metrics for the first week after video publish
type FirstWeekMetrics struct {
	Views    int64
	Likes    int64
	Comments int64
	CTR      float64
}

// GetFirstWeekMetrics fetches performance metrics for a specific video
// during its first 7 days after publication.
//
// This provides apples-to-apples comparison across videos regardless of age,
// since the YouTube algorithm prioritizes early performance.
//
// Parameters:
//   - ctx: Context for the API call
//   - videoID: The YouTube video ID
//   - publishDate: When the video was published
//
// Returns:
//   - FirstWeekMetrics: Performance data for days 0-7
//   - error: Any error encountered during the API call
func GetFirstWeekMetrics(ctx context.Context, videoID string, publishDate time.Time) (FirstWeekMetrics, error) {
	client := getClient(ctx)

	// Initialize YouTube Analytics API service
	analyticsService, err := youtubeanalytics.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return FirstWeekMetrics{}, fmt.Errorf("failed to create YouTube Analytics service: %w", err)
	}

	// Calculate first week date range
	startDate := publishDate
	endDate := publishDate.AddDate(0, 0, 7) // 7 days after publish

	// Don't query future dates
	now := time.Now()
	if endDate.After(now) {
		endDate = now
	}

	// Format dates for API request (YYYY-MM-DD)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Get channel ID from settings
	channelID := configuration.GlobalSettings.YouTube.ChannelId
	if channelID == "" {
		return FirstWeekMetrics{}, fmt.Errorf("YouTube channel ID not configured in settings.yaml")
	}

	// Fetch first-week analytics for specific video
	// Note: Using filters parameter to restrict to single video
	analyticsCall := analyticsService.Reports.Query().
		Ids("channel==" + channelID).
		StartDate(startDateStr).
		EndDate(endDateStr).
		Metrics("views,likes,comments,cardClickRate").
		Filters("video==" + videoID) // Filter to specific video

	analyticsResponse, err := analyticsCall.Do()
	if err != nil {
		return FirstWeekMetrics{}, fmt.Errorf("failed to fetch first-week analytics for video %s: %w", videoID, err)
	}

	// Check if we have any data
	if analyticsResponse.Rows == nil || len(analyticsResponse.Rows) == 0 {
		// Video may be too new or have no views in first week
		return FirstWeekMetrics{}, nil
	}

	// Parse the single row of data
	row := analyticsResponse.Rows[0]
	if len(row) < 4 {
		return FirstWeekMetrics{}, fmt.Errorf("unexpected response format: expected 4 fields, got %d", len(row))
	}

	metrics := FirstWeekMetrics{
		Views:    int64(row[0].(float64)),
		Likes:    int64(row[1].(float64)),
		Comments: int64(row[2].(float64)),
		CTR:      row[3].(float64),
	}

	return metrics, nil
}

// EnrichWithFirstWeekMetrics fetches first-week performance data for each video
// in the analytics array and populates the FirstWeek* fields.
//
// This function makes N API calls (one per video) to get accurate first-week metrics,
// enabling apples-to-apples comparison across videos regardless of age.
//
// Parameters:
//   - ctx: Context for the API calls
//   - analytics: Array of VideoAnalytics to enrich
//
// Returns:
//   - []VideoAnalytics: Enriched array with FirstWeek* fields populated
//   - error: Any error encountered during API calls
func EnrichWithFirstWeekMetrics(ctx context.Context, analytics []VideoAnalytics) ([]VideoAnalytics, error) {
	enriched := make([]VideoAnalytics, len(analytics))

	for i, video := range analytics {
		enriched[i] = video

		// Skip if first-week data already populated (e.g., from tests or pre-processing)
		if video.FirstWeekViews > 0 {
			continue
		}

		// Fetch first-week metrics for this video
		firstWeek, err := GetFirstWeekMetrics(ctx, video.VideoID, video.PublishedAt)
		if err != nil {
			// Log error but continue with other videos
			fmt.Printf("Warning: Failed to fetch first-week metrics for video %s: %v\n", video.VideoID, err)
			continue
		}

		// Populate first-week fields
		enriched[i].FirstWeekViews = firstWeek.Views
		enriched[i].FirstWeekLikes = firstWeek.Likes
		enriched[i].FirstWeekComments = firstWeek.Comments
		enriched[i].FirstWeekCTR = firstWeek.CTR
	}

	return enriched, nil
}

// EnrichWithTimingData extracts timing information from PublishedAt and calculates
// engagement metrics based on first-week performance data.
//
// This function should be called after EnrichWithFirstWeekMetrics() to ensure
// FirstWeek* fields are populated.
//
// Parameters:
//   - analytics: Array of VideoAnalytics to enrich (should have FirstWeek* fields populated)
//
// Returns:
//   - []VideoAnalytics: Enriched array with DayOfWeek, TimeOfDay, and FirstWeekEngagement populated
func EnrichWithTimingData(analytics []VideoAnalytics) []VideoAnalytics {
	enriched := make([]VideoAnalytics, len(analytics))

	for i, video := range analytics {
		enriched[i] = video

		// Extract day of week (Monday, Tuesday, etc.)
		enriched[i].DayOfWeek = video.PublishedAt.Weekday().String()

		// Extract time of day in UTC (HH:MM format)
		enriched[i].TimeOfDay = video.PublishedAt.UTC().Format("15:04")

		// Calculate first-week engagement rate as percentage
		if video.FirstWeekViews > 0 {
			totalEngagement := float64(video.FirstWeekLikes + video.FirstWeekComments)
			enriched[i].FirstWeekEngagement = (totalEngagement / float64(video.FirstWeekViews)) * 100
		} else {
			enriched[i].FirstWeekEngagement = 0.0
		}
	}

	return enriched
}

// TimeSlot represents a unique day/time combination for grouping videos
type TimeSlot struct {
	DayOfWeek string // "Monday", "Tuesday", etc.
	TimeOfDay string // "16:00", "09:00" (UTC, 24-hour format)
}

// String returns a formatted string representation: "Monday 16:00 UTC"
func (ts TimeSlot) String() string {
	return fmt.Sprintf("%s %s UTC", ts.DayOfWeek, ts.TimeOfDay)
}

// GroupByTimeSlot groups videos by their publish day/time.
//
// This enables analysis of performance patterns for specific time slots
// (e.g., all videos published on "Monday 16:00 UTC").
//
// Parameters:
//   - analytics: Array of VideoAnalytics with DayOfWeek and TimeOfDay populated
//
// Returns:
//   - map[TimeSlot][]VideoAnalytics: Videos grouped by time slot
func GroupByTimeSlot(analytics []VideoAnalytics) map[TimeSlot][]VideoAnalytics {
	grouped := make(map[TimeSlot][]VideoAnalytics)

	for _, video := range analytics {
		slot := TimeSlot{
			DayOfWeek: video.DayOfWeek,
			TimeOfDay: video.TimeOfDay,
		}
		grouped[slot] = append(grouped[slot], video)
	}

	return grouped
}

// TimeSlotPerformance holds aggregated metrics for a time slot
type TimeSlotPerformance struct {
	Slot                    TimeSlot
	VideoCount              int
	AvgFirstWeekViews       float64
	AvgFirstWeekCTR         float64
	AvgFirstWeekEngagement  float64
	TotalFirstWeekViews     int64
}

// CalculateTimeSlotPerformance computes aggregate metrics for grouped videos.
//
// This provides summary statistics for each time slot to help identify
// high-performing and low-performing publish times.
//
// Parameters:
//   - grouped: Map of videos grouped by time slot
//
// Returns:
//   - []TimeSlotPerformance: Array of performance metrics per time slot
func CalculateTimeSlotPerformance(grouped map[TimeSlot][]VideoAnalytics) []TimeSlotPerformance {
	results := make([]TimeSlotPerformance, 0, len(grouped))

	for slot, videos := range grouped {
		if len(videos) == 0 {
			continue
		}

		perf := TimeSlotPerformance{
			Slot:       slot,
			VideoCount: len(videos),
		}

		var totalViews, totalCTR, totalEngagement float64

		for _, video := range videos {
			totalViews += float64(video.FirstWeekViews)
			totalCTR += video.FirstWeekCTR
			totalEngagement += video.FirstWeekEngagement
			perf.TotalFirstWeekViews += video.FirstWeekViews
		}

		count := float64(len(videos))
		perf.AvgFirstWeekViews = totalViews / count
		perf.AvgFirstWeekCTR = totalCTR / count
		perf.AvgFirstWeekEngagement = totalEngagement / count

		results = append(results, perf)
	}

	return results
}

// isShort checks if a video is a YouTube Short based on its duration.
// Shorts are defined as videos with duration ≤ 60 seconds.
//
// Parameters:
//   - duration: ISO 8601 duration string (e.g., "PT1M30S", "PT45S", "PT10M5S")
//
// Returns:
//   - bool: true if video is a Short (≤ 60 seconds), false otherwise
func isShort(duration string) bool {
	// Parse ISO 8601 duration format (e.g., "PT1M30S" = 1 minute 30 seconds)
	// Format: PT[hours]H[minutes]M[seconds]S

	if duration == "" {
		return false
	}

	// Use strings.Contains to determine which format to parse
	hasHours := strings.Contains(duration, "H")
	hasMinutes := strings.Contains(duration, "M")
	hasSeconds := strings.Contains(duration, "S")

	var hours, minutes, seconds int

	switch {
	case hasHours && hasMinutes && hasSeconds:
		fmt.Sscanf(duration, "PT%dH%dM%dS", &hours, &minutes, &seconds)
	case hasHours && hasMinutes:
		fmt.Sscanf(duration, "PT%dH%dM", &hours, &minutes)
	case hasHours && hasSeconds:
		fmt.Sscanf(duration, "PT%dH%dS", &hours, &seconds)
	case hasMinutes && hasSeconds:
		fmt.Sscanf(duration, "PT%dM%dS", &minutes, &seconds)
	case hasHours:
		fmt.Sscanf(duration, "PT%dH", &hours)
	case hasMinutes:
		fmt.Sscanf(duration, "PT%dM", &minutes)
	case hasSeconds:
		fmt.Sscanf(duration, "PT%dS", &seconds)
	default:
		return false
	}

	totalSeconds := hours*3600 + minutes*60 + seconds
	return totalSeconds <= 60
}

// ChannelDemographics holds age and gender distribution data for sponsor page analytics
type ChannelDemographics struct {
	AgeGroups []AgeGroupData
	Gender    []GenderData
}

// AgeGroupData represents viewer percentage for a specific age group
type AgeGroupData struct {
	AgeGroup   string  // "age13-17", "age18-24", "age25-34", "age35-44", "age45-54", "age55-64", "age65-"
	Percentage float64 // viewerPercentage (0-100)
}

// GenderData represents viewer percentage for a specific gender
type GenderData struct {
	Gender     string  // "male", "female", "user_specified"
	Percentage float64 // viewerPercentage (0-100)
}

// GeographicDistribution holds top countries by views
type GeographicDistribution struct {
	Countries []CountryData
}

// CountryData represents view data for a specific country
type CountryData struct {
	CountryCode string  // ISO 3166-1 alpha-2 code (e.g., "US", "GB", "IN")
	Views       int64   // Total views from this country
	Percentage  float64 // Percentage of total views (calculated)
}

// ChannelStatistics holds channel-level metrics from YouTube Data API
type ChannelStatistics struct {
	SubscriberCount   int64
	TotalViews        int64
	VideoCount        int64
	HiddenSubscribers bool // true if channel hides subscriber count
}

// GetChannelDemographics fetches age and gender distribution from YouTube Analytics API.
// Data is aggregated over the last 90 days for statistical significance.
//
// Returns:
//   - ChannelDemographics: Age and gender distribution data
//   - error: Any error encountered during the API call
func GetChannelDemographics(ctx context.Context) (ChannelDemographics, error) {
	client := getClient(ctx)

	analyticsService, err := youtubeanalytics.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return ChannelDemographics{}, fmt.Errorf("failed to create YouTube Analytics service: %w", err)
	}

	// Use last 90 days for statistical significance
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -90)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	channelID := configuration.GlobalSettings.YouTube.ChannelId
	if channelID == "" {
		return ChannelDemographics{}, fmt.Errorf("YouTube channel ID not configured in settings.yaml")
	}

	// Fetch demographics data with ageGroup and gender dimensions
	analyticsCall := analyticsService.Reports.Query().
		Ids("channel==" + channelID).
		StartDate(startDateStr).
		EndDate(endDateStr).
		Dimensions("ageGroup,gender").
		Metrics("viewerPercentage")

	response, err := analyticsCall.Do()
	if err != nil {
		return ChannelDemographics{}, fmt.Errorf("failed to fetch demographics data: %w", err)
	}

	demographics := ChannelDemographics{
		AgeGroups: []AgeGroupData{},
		Gender:    []GenderData{},
	}

	if response.Rows == nil || len(response.Rows) == 0 {
		return demographics, nil
	}

	// Aggregate data by age group and gender
	ageGroupTotals := make(map[string]float64)
	genderTotals := make(map[string]float64)

	for _, row := range response.Rows {
		if len(row) < 3 {
			continue
		}

		ageGroup, _ := row[0].(string)
		gender, _ := row[1].(string)
		percentage, _ := row[2].(float64)

		ageGroupTotals[ageGroup] += percentage
		genderTotals[gender] += percentage
	}

	// Convert maps to slices
	for ageGroup, percentage := range ageGroupTotals {
		demographics.AgeGroups = append(demographics.AgeGroups, AgeGroupData{
			AgeGroup:   ageGroup,
			Percentage: percentage,
		})
	}

	for gender, percentage := range genderTotals {
		demographics.Gender = append(demographics.Gender, GenderData{
			Gender:     gender,
			Percentage: percentage,
		})
	}

	return demographics, nil
}

// GetGeographicDistribution fetches top countries by views from YouTube Analytics API.
// Returns the top 10 countries by view count over the last 90 days.
//
// Returns:
//   - GeographicDistribution: Top countries with view counts and percentages
//   - error: Any error encountered during the API call
func GetGeographicDistribution(ctx context.Context) (GeographicDistribution, error) {
	client := getClient(ctx)

	analyticsService, err := youtubeanalytics.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return GeographicDistribution{}, fmt.Errorf("failed to create YouTube Analytics service: %w", err)
	}

	// Use last 90 days for statistical significance
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -90)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	channelID := configuration.GlobalSettings.YouTube.ChannelId
	if channelID == "" {
		return GeographicDistribution{}, fmt.Errorf("YouTube channel ID not configured in settings.yaml")
	}

	// Fetch geographic data with country dimension, sorted by views descending
	analyticsCall := analyticsService.Reports.Query().
		Ids("channel==" + channelID).
		StartDate(startDateStr).
		EndDate(endDateStr).
		Dimensions("country").
		Metrics("views").
		Sort("-views").
		MaxResults(10)

	response, err := analyticsCall.Do()
	if err != nil {
		return GeographicDistribution{}, fmt.Errorf("failed to fetch geographic data: %w", err)
	}

	distribution := GeographicDistribution{
		Countries: []CountryData{},
	}

	if response.Rows == nil || len(response.Rows) == 0 {
		return distribution, nil
	}

	// Calculate total views for percentage calculation
	var totalViews int64
	for _, row := range response.Rows {
		if len(row) >= 2 {
			views := int64(row[1].(float64))
			totalViews += views
		}
	}

	// Parse country data
	for _, row := range response.Rows {
		if len(row) < 2 {
			continue
		}

		countryCode, _ := row[0].(string)
		views := int64(row[1].(float64))

		var percentage float64
		if totalViews > 0 {
			percentage = float64(views) / float64(totalViews) * 100
		}

		distribution.Countries = append(distribution.Countries, CountryData{
			CountryCode: countryCode,
			Views:       views,
			Percentage:  percentage,
		})
	}

	return distribution, nil
}

// GetChannelStatistics fetches channel-level statistics from YouTube Data API.
// Returns subscriber count, total views, and video count.
//
// Returns:
//   - ChannelStatistics: Channel statistics data
//   - error: Any error encountered during the API call
func GetChannelStatistics(ctx context.Context) (ChannelStatistics, error) {
	client := getClient(ctx)

	youtubeService, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return ChannelStatistics{}, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	channelID := configuration.GlobalSettings.YouTube.ChannelId
	if channelID == "" {
		return ChannelStatistics{}, fmt.Errorf("YouTube channel ID not configured in settings.yaml")
	}

	// Fetch channel statistics
	channelsCall := youtubeService.Channels.List([]string{"statistics"}).Id(channelID)
	response, err := channelsCall.Do()
	if err != nil {
		return ChannelStatistics{}, fmt.Errorf("failed to fetch channel statistics: %w", err)
	}

	if len(response.Items) == 0 {
		return ChannelStatistics{}, fmt.Errorf("channel not found: %s", channelID)
	}

	channel := response.Items[0]
	stats := channel.Statistics

	return ChannelStatistics{
		SubscriberCount:   int64(stats.SubscriberCount),
		TotalViews:        int64(stats.ViewCount),
		VideoCount:        int64(stats.VideoCount),
		HiddenSubscribers: stats.HiddenSubscriberCount,
	}, nil
}
