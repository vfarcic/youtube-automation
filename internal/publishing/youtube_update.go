package publishing

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// VideoMetadata holds the metadata for a YouTube video
type VideoMetadata struct {
	Title       string
	Description string
	Tags        []string
	PublishedAt string // ISO 8601 format (e.g., "2025-12-06T15:00:00Z")
}

// boilerplateDelimiter is the separator between custom content and boilerplate
const boilerplateDelimiter = "▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬"

// TimecodesHeader is the header for the timecodes section.
// It also serves as the idempotency marker for AMA processing — when
// present in a video's description, the video has already been processed.
const TimecodesHeader = "▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬"

// videoListDoer wraps the chained .Do() of a Videos.List call.
type videoListDoer interface {
	Do(opts ...googleapi.CallOption) (*youtube.VideoListResponse, error)
}

// videosClient abstracts youtube.VideosService for testing. The ctx is
// forwarded to .Context(ctx) on the underlying call so callers can cancel
// the in-flight HTTP request via the standard context plumbing. Reuses the
// videoUpdateDoer interface declared in youtube.go.
type videosClient interface {
	List(ctx context.Context, part []string, videoID string) videoListDoer
	Update(ctx context.Context, part []string, video *youtube.Video) videoUpdateDoer
}

// realVideosClient adapts *youtube.VideosService to videosClient.
type realVideosClient struct {
	svc *youtube.VideosService
}

func (r *realVideosClient) List(ctx context.Context, part []string, videoID string) videoListDoer {
	return r.svc.List(part).Id(videoID).Context(ctx)
}

func (r *realVideosClient) Update(ctx context.Context, part []string, video *youtube.Video) videoUpdateDoer {
	return r.svc.Update(part, video).Context(ctx)
}

// buildVideosClient constructs a videosClient from an authenticated *http.Client.
// Split out so tests can exercise service construction without a real OAuth flow.
func buildVideosClient(ctx context.Context, client *http.Client) (videosClient, error) {
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}
	return &realVideosClient{svc: service.Videos}, nil
}

// newVideosClient constructs an authenticated videosClient. Tests may override
// this to inject a mock.
var newVideosClient = func() (videosClient, error) {
	ctx := context.Background()
	client, err := getClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("OAuth failed: %w", err)
	}
	return buildVideosClient(ctx, client)
}

// GetVideoMetadata fetches the current metadata for a YouTube video. The ctx
// is forwarded to the underlying YouTube API call so callers can cancel the
// in-flight HTTP request.
func GetVideoMetadata(ctx context.Context, videoID string) (*VideoMetadata, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID cannot be empty")
	}
	client, err := newVideosClient()
	if err != nil {
		return nil, err
	}
	return getVideoMetadata(ctx, client, videoID)
}

// getVideoMetadata is the testable inner implementation of GetVideoMetadata.
func getVideoMetadata(ctx context.Context, client videosClient, videoID string) (*VideoMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	response, err := client.List(ctx, []string{"snippet"}, videoID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	video := response.Items[0]
	if video == nil || video.Snippet == nil {
		return nil, fmt.Errorf("video %s missing snippet metadata", videoID)
	}
	return &VideoMetadata{
		Title:       video.Snippet.Title,
		Description: video.Snippet.Description,
		Tags:        video.Snippet.Tags,
		PublishedAt: video.Snippet.PublishedAt,
	}, nil
}

// UpdateAMAVideo updates a YouTube video with AMA-specific content.
// It merges the new description with existing boilerplate and appends timecodes.
// The ctx is forwarded to the underlying YouTube API calls so callers can
// cancel the in-flight HTTP requests.
func UpdateAMAVideo(ctx context.Context, videoID, title, description, tags, timecodes string) error {
	if videoID == "" {
		return fmt.Errorf("video ID cannot be empty")
	}
	client, err := newVideosClient()
	if err != nil {
		return err
	}
	return updateAMAVideo(ctx, client, videoID, title, description, tags, timecodes)
}

// updateAMAVideo is the testable inner implementation of UpdateAMAVideo.
func updateAMAVideo(ctx context.Context, client videosClient, videoID, title, description, tags, timecodes string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	listResponse, err := client.List(ctx, []string{"snippet"}, videoID).Do()
	if err != nil {
		return fmt.Errorf("failed to fetch video: %w", err)
	}

	if len(listResponse.Items) == 0 {
		return fmt.Errorf("video not found: %s", videoID)
	}

	currentVideo := listResponse.Items[0]
	if currentVideo == nil || currentVideo.Snippet == nil {
		return fmt.Errorf("video %s missing snippet metadata", videoID)
	}

	// Build new description by merging with existing boilerplate
	newDescription := buildAMADescription(description, currentVideo.Snippet.Description, timecodes)

	// Prepare update request
	updateVideo := &youtube.Video{
		Id: videoID,
		Snippet: &youtube.VideoSnippet{
			Title:       title,
			Description: newDescription,
			CategoryId:  currentVideo.Snippet.CategoryId, // Required field - preserve existing
		},
	}

	// Only update title if provided
	if title == "" {
		updateVideo.Snippet.Title = currentVideo.Snippet.Title
	}

	// Parse and set tags if provided
	if tags != "" {
		updateVideo.Snippet.Tags = parseTags(tags)
	} else {
		updateVideo.Snippet.Tags = currentVideo.Snippet.Tags
	}

	if _, err := client.Update(ctx, []string{"snippet"}, updateVideo).Do(); err != nil {
		return fmt.Errorf("failed to update video: %w", err)
	}

	return nil
}

// buildAMADescription constructs the final description by:
// 1. Using the new description as the first section
// 2. Preserving existing boilerplate (content after the first ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬)
// 3. Appending timecodes section at the end
func buildAMADescription(newDescription, currentDescription, timecodes string) string {
	var result strings.Builder

	// Add new description
	if newDescription != "" {
		result.WriteString(strings.TrimSpace(newDescription))
		result.WriteString("\n\n")
	}

	// Extract and preserve boilerplate from current description
	boilerplate := extractBoilerplate(currentDescription)
	if boilerplate != "" {
		result.WriteString(boilerplate)
	}

	// Add timecodes section if provided
	if timecodes != "" {
		// Remove any existing timecodes section from result
		currentResult := result.String()
		if idx := strings.Index(currentResult, TimecodesHeader); idx != -1 {
			currentResult = strings.TrimSpace(currentResult[:idx])
			result.Reset()
			result.WriteString(currentResult)
		}

		// Ensure empty line before timecodes header
		result.WriteString("\n\n")
		result.WriteString(TimecodesHeader)
		result.WriteString("\n")
		result.WriteString(strings.TrimSpace(timecodes))
	}

	return strings.TrimSpace(result.String())
}

// extractBoilerplate extracts the boilerplate section from a description.
// The boilerplate starts at the first occurrence of ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
// but excludes any existing timecodes section.
func extractBoilerplate(description string) string {
	idx := strings.Index(description, boilerplateDelimiter)
	if idx == -1 {
		return ""
	}

	boilerplate := description[idx:]

	// Remove existing timecodes section from boilerplate
	if timecodesIdx := strings.Index(boilerplate, TimecodesHeader); timecodesIdx != -1 {
		boilerplate = strings.TrimSpace(boilerplate[:timecodesIdx])
	}

	return boilerplate
}

// parseTags splits a comma-separated string into individual tags
func parseTags(tagsStr string) []string {
	if tagsStr == "" {
		return nil
	}

	parts := strings.Split(tagsStr, ",")
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}
