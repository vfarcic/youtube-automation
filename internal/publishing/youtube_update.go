package publishing

import (
	"context"
	"fmt"
	"strings"

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

// timecodesHeader is the header for the timecodes section
const timecodesHeader = "▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬"

// GetVideoMetadata fetches the current metadata for a YouTube video
func GetVideoMetadata(videoID string) (*VideoMetadata, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID cannot be empty")
	}

	ctx := context.Background()
	client := getClient(ctx)

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	call := service.Videos.List([]string{"snippet"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	video := response.Items[0]
	return &VideoMetadata{
		Title:       video.Snippet.Title,
		Description: video.Snippet.Description,
		Tags:        video.Snippet.Tags,
		PublishedAt: video.Snippet.PublishedAt,
	}, nil
}

// UpdateAMAVideo updates a YouTube video with AMA-specific content.
// It merges the new description with existing boilerplate and appends timecodes.
func UpdateAMAVideo(videoID, title, description, tags, timecodes string) error {
	if videoID == "" {
		return fmt.Errorf("video ID cannot be empty")
	}

	ctx := context.Background()
	client := getClient(ctx)

	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %w", err)
	}

	// Fetch current video to get existing description and required fields
	listCall := service.Videos.List([]string{"snippet"}).Id(videoID)
	listResponse, err := listCall.Do()
	if err != nil {
		return fmt.Errorf("failed to fetch video: %w", err)
	}

	if len(listResponse.Items) == 0 {
		return fmt.Errorf("video not found: %s", videoID)
	}

	currentVideo := listResponse.Items[0]

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

	// Perform the update
	updateCall := service.Videos.Update([]string{"snippet"}, updateVideo)
	_, err = updateCall.Do()
	if err != nil {
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
		if idx := strings.Index(currentResult, timecodesHeader); idx != -1 {
			currentResult = strings.TrimSpace(currentResult[:idx])
			result.Reset()
			result.WriteString(currentResult)
		}

		// Ensure empty line before timecodes header
		result.WriteString("\n\n")
		result.WriteString(timecodesHeader)
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
	if timecodesIdx := strings.Index(boilerplate, timecodesHeader); timecodesIdx != -1 {
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
