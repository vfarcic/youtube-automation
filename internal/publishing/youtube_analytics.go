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
	Views              int64
	CTR                float64 // Click-through rate (percentage)
	AverageViewDuration float64 // In seconds
	Likes              int64
	Comments           int64
	PublishedAt        time.Time
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
	// Metrics: views, averageViewDuration, likes, comments
	// Dimensions: video (group by video ID)
	// Note: CTR (cardClickRate) requires cards to be present, so we'll fetch impressionClickThroughRate instead
	analyticsCall := analyticsService.Reports.Query().
		Ids("channel=="+channelID).
		StartDate(startDateStr).
		EndDate(endDateStr).
		Metrics("views,estimatedMinutesWatched,likes,comments,averageViewDuration").
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
		if len(row) < 6 {
			continue // Skip malformed rows
		}

		videoID, _ := row[0].(string)
		views := int64(row[1].(float64))
		// estimatedMinutesWatched := row[2].(float64) // Not currently used
		likes := int64(row[3].(float64))
		comments := int64(row[4].(float64))
		avgViewDuration := row[5].(float64)

		// Get video metadata
		metadata, exists := videoMetadata[videoID]
		if !exists {
			// Skip videos without metadata (shouldn't happen)
			continue
		}

		// Calculate CTR estimate (we don't have actual CTR, so this is a placeholder)
		// In reality, CTR requires impressions data which isn't always available
		ctr := 0.0 // Placeholder - actual CTR would require impressions data

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
