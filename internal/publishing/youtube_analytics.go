package publishing

import (
	"context"
	"fmt"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"

	"golang.org/x/oauth2"
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
	// Create OAuth client with analytics scope
	client := getClient(ctx, &oauth2.Config{
		Scopes: []string{
			youtube.YoutubeReadonlyScope,
			"https://www.googleapis.com/auth/yt-analytics.readonly",
		},
	})

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
	analyticsCall := analyticsService.Reports.Query().
		Ids("channel=="+channelID).
		StartDate(startDateStr).
		EndDate(endDateStr).
		Metrics("views,estimatedMinutesWatched,likes,comments,averageViewDuration,cardClickRate").
		Dimensions("video").
		Sort("-views"). // Sort by views descending
		MaxResults(200) // Fetch up to 200 videos

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

	// Fetch video metadata (titles and publish dates) from YouTube Data API
	videoMetadata := make(map[string]struct {
		Title       string
		PublishedAt time.Time
	})

	// YouTube Data API allows up to 50 video IDs per request
	// Split into chunks if needed
	for i := 0; i < len(videoIDs); i += 50 {
		end := i + 50
		if end > len(videoIDs) {
			end = len(videoIDs)
		}
		chunk := videoIDs[i:end]

		videosCall := youtubeService.Videos.List([]string{"snippet"}).Id(chunk...)
		videosResponse, err := videosCall.Do()
		if err != nil {
			return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
		}

		for _, video := range videosResponse.Items {
			publishedAt, _ := time.Parse(time.RFC3339, video.Snippet.PublishedAt)
			videoMetadata[video.Id] = struct {
				Title       string
				PublishedAt time.Time
			}{
				Title:       video.Snippet.Title,
				PublishedAt: publishedAt,
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
	// Create OAuth client with analytics scope
	client := getClient(ctx, &oauth2.Config{
		Scopes: []string{
			youtube.YoutubeReadonlyScope,
			"https://www.googleapis.com/auth/yt-analytics.readonly",
		},
	})

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
