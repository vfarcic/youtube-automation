package publishing

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/storage"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// dateFormats lists the date formats to try when parsing video dates
var dateFormats = []string{
	"2006-01-02T15:04",
	"2006-01-02T15:04:05Z",
	time.RFC3339,
}

// parseVideoDate attempts to parse a video date string using known formats
func parseVideoDate(dateStr string) (time.Time, error) {
	for _, format := range dateFormats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// getPublishStatus determines the privacy status and publish date for a dubbed video.
// If the scheduled date is in the future, returns "private" with the date.
// If the date is in the past or empty, returns "public" for immediate availability.
func getPublishStatus(scheduledDate string) (privacyStatus string, publishAt string) {
	if scheduledDate == "" {
		return "public", ""
	}

	parsedDate, err := parseVideoDate(scheduledDate)
	if err != nil {
		// Can't parse date, make it public immediately
		return "public", ""
	}

	if parsedDate.After(time.Now()) {
		// Future date - schedule it
		return "private", FormatScheduleISO(parsedDate)
	}

	// Past date - make it public immediately
	return "public", ""
}

// UploadDubbedVideo uploads a dubbed video to the appropriate channel for the given language.
// Currently only supports "es" (Spanish). Returns the YouTube video ID of the uploaded video.
//
// Parameters:
//   - video: The original video with dubbing info
//   - langCode: Language code (e.g., "es" for Spanish)
//
// Returns:
//   - string: The YouTube video ID of the uploaded dubbed video
//   - error: Any error that occurred during upload
func UploadDubbedVideo(video *storage.Video, langCode string) (string, error) {
	// Validate language code - only Spanish supported for now
	if langCode != "es" {
		return "", fmt.Errorf("unsupported language code: %s (only 'es' is currently supported)", langCode)
	}

	// Get dubbing info for the language
	dubbingInfo, exists := video.Dubbing[langCode]
	if !exists {
		return "", fmt.Errorf("no dubbing info found for language: %s", langCode)
	}

	// Validate required fields
	if dubbingInfo.DubbedVideoPath == "" {
		return "", fmt.Errorf("dubbed video path is empty")
	}
	if dubbingInfo.Title == "" {
		return "", fmt.Errorf("translated title is empty")
	}

	// Verify dubbed video file exists
	if _, err := os.Stat(dubbingInfo.DubbedVideoPath); os.IsNotExist(err) {
		return "", fmt.Errorf("dubbed video file does not exist: %s", dubbingInfo.DubbedVideoPath)
	}

	// Get channel ID first (before trying to authenticate)
	channelID := GetSpanishChannelID()
	if channelID == "" {
		return "", fmt.Errorf("Spanish channel ID is not configured in settings.yaml")
	}

	// Get the appropriate client for the language
	ctx := context.Background()
	client := GetSpanishChannelClient(ctx) // For now, only Spanish

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("error creating YouTube client: %w", err)
	}

	// Build description for dubbed video
	description := buildDubbedDescription(dubbingInfo, video.VideoId)

	// Determine publish status based on original video's date
	privacyStatus, publishAt := getPublishStatus(video.Date)

	// Prepare upload request
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:                dubbingInfo.Title,
			Description:          description,
			CategoryId:           "28", // Science & Technology
			ChannelId:            channelID,
			DefaultLanguage:      langCode,
			DefaultAudioLanguage: langCode,
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: privacyStatus,
		},
	}

	// Set scheduled publish time if in the future
	if publishAt != "" {
		upload.Status.PublishAt = publishAt
	}

	// Add tags if available
	if dubbingInfo.Tags != "" {
		upload.Snippet.Tags = strings.Split(dubbingInfo.Tags, ",")
	}

	// Open video file
	file, err := os.Open(dubbingInfo.DubbedVideoPath)
	if err != nil {
		return "", fmt.Errorf("error opening dubbed video file %s: %w", dubbingInfo.DubbedVideoPath, err)
	}
	defer file.Close()

	// Upload
	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	response, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("error uploading dubbed video to YouTube: %w", err)
	}

	if publishAt != "" {
		fmt.Printf("Dubbed video uploaded and scheduled! Video ID: %v, Publish at: %s\n", response.Id, publishAt)
	} else {
		fmt.Printf("Dubbed video uploaded and made public! Video ID: %v\n", response.Id)
	}

	// Upload localized thumbnail if available (same pattern as English videos)
	if dubbingInfo.ThumbnailPath != "" {
		if _, statErr := os.Stat(dubbingInfo.ThumbnailPath); statErr == nil {
			thumbnailFile, openErr := os.Open(dubbingInfo.ThumbnailPath)
			if openErr != nil {
				fmt.Printf("Warning: could not open thumbnail file %s: %v\n", dubbingInfo.ThumbnailPath, openErr)
			} else {
				defer thumbnailFile.Close()
				thumbnailCall := service.Thumbnails.Set(response.Id)
				thumbnailResp, thumbnailErr := thumbnailCall.Media(thumbnailFile).Do()
				if thumbnailErr != nil {
					fmt.Printf("Warning: could not upload thumbnail: %v\n", thumbnailErr)
				} else {
					fmt.Printf("Thumbnail uploaded, URL: %s\n", thumbnailResp.Items[0].Default.Url)
				}
			}
		}
	}

	return response.Id, nil
}

// UploadDubbedShort uploads a dubbed short to the Spanish channel with interval-based scheduling.
// Shorts are scheduled relative to the main video's publish date, maintaining the same
// day intervals as the original shorts (Short 1 = +1 day, Short 2 = +2 days, etc.).
//
// Parameters:
//   - video: The original video with dubbing info for the short
//   - shortIndex: 0-based index of the short (used to calculate day offset: index+1 days after main video)
//
// Returns:
//   - string: The YouTube video ID of the uploaded dubbed short
//   - error: Any error that occurred during upload
func UploadDubbedShort(video *storage.Video, shortIndex int) (string, error) {
	// Validate short index
	if shortIndex < 0 || shortIndex >= len(video.Shorts) {
		return "", fmt.Errorf("invalid short index: %d (video has %d shorts)", shortIndex, len(video.Shorts))
	}

	// Get the dubbing key for this short
	dubbingKey := fmt.Sprintf("es:short%d", shortIndex+1)

	// Get dubbing info for the short
	dubbingInfo, exists := video.Dubbing[dubbingKey]
	if !exists {
		return "", fmt.Errorf("no dubbing info found for short: %s", dubbingKey)
	}

	// Validate required fields
	if dubbingInfo.DubbedVideoPath == "" {
		return "", fmt.Errorf("dubbed short video path is empty")
	}

	// Use translated title if available, otherwise use original short title
	shortTitle := dubbingInfo.Title
	if shortTitle == "" {
		shortTitle = video.Shorts[shortIndex].Title
	}
	if shortTitle == "" {
		return "", fmt.Errorf("short title is empty (no translated or original title)")
	}

	// Verify dubbed video file exists
	if _, err := os.Stat(dubbingInfo.DubbedVideoPath); os.IsNotExist(err) {
		return "", fmt.Errorf("dubbed short file does not exist: %s", dubbingInfo.DubbedVideoPath)
	}

	// Get channel ID
	channelID := GetSpanishChannelID()
	if channelID == "" {
		return "", fmt.Errorf("Spanish channel ID is not configured in settings.yaml")
	}

	// Get the appropriate client
	ctx := context.Background()
	client := GetSpanishChannelClient(ctx)

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", fmt.Errorf("error creating YouTube client: %w", err)
	}

	// Calculate scheduled date for this short
	// Shorts are scheduled: main video date + (shortIndex + 1) days
	privacyStatus, publishAt := calculateShortPublishStatus(video.Date, shortIndex)

	// Build description for dubbed short
	description := buildDubbedShortDescription(shortTitle, video.VideoId)

	// Prepare upload request
	upload := &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:                shortTitle,
			Description:          description,
			CategoryId:           "28", // Science & Technology
			ChannelId:            channelID,
			DefaultLanguage:      "es",
			DefaultAudioLanguage: "es",
		},
		Status: &youtube.VideoStatus{
			PrivacyStatus: privacyStatus,
		},
	}

	// Set scheduled publish time if applicable
	if publishAt != "" {
		upload.Status.PublishAt = publishAt
	}

	// Open video file
	file, err := os.Open(dubbingInfo.DubbedVideoPath)
	if err != nil {
		return "", fmt.Errorf("error opening dubbed short file %s: %w", dubbingInfo.DubbedVideoPath, err)
	}
	defer file.Close()

	// Upload
	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	response, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("error uploading dubbed short to YouTube: %w", err)
	}

	if publishAt != "" {
		fmt.Printf("Dubbed short uploaded and scheduled! Video ID: %v, Publish at: %s\n", response.Id, publishAt)
	} else {
		fmt.Printf("Dubbed short uploaded and made public! Video ID: %v\n", response.Id)
	}
	return response.Id, nil
}

// calculateShortPublishStatus determines the privacy status and publish date for a dubbed short.
// Shorts are scheduled at main video date + (shortIndex + 1) days.
// If the calculated date is in the past, the short is made public immediately.
func calculateShortPublishStatus(mainVideoDate string, shortIndex int) (privacyStatus string, publishAt string) {
	if mainVideoDate == "" {
		return "public", ""
	}

	parsedDate, err := parseVideoDate(mainVideoDate)
	if err != nil {
		return "public", ""
	}

	// Calculate short's scheduled date: main video + (index + 1) days
	daysAfter := shortIndex + 1
	shortDate := parsedDate.AddDate(0, 0, daysAfter)

	if shortDate.After(time.Now()) {
		// Future date - schedule it with random hour/minute like original shorts
		schedules := CalculateShortsSchedule(parsedDate, shortIndex+1)
		if len(schedules) > shortIndex {
			return "private", FormatScheduleISO(schedules[shortIndex])
		}
		return "private", FormatScheduleISO(shortDate)
	}

	// Past date - make it public immediately
	return "public", ""
}

// buildDubbedShortDescription creates the description for a dubbed short.
func buildDubbedShortDescription(title string, originalMainVideoID string) string {
	mainVideoURL := ""
	if originalMainVideoID != "" {
		mainVideoURL = fmt.Sprintf("\nVideo completo: %s\n", GetYouTubeURL(originalMainVideoID))
	}
	return fmt.Sprintf("%s%s\n#Shorts", title, mainVideoURL)
}

// buildDubbedDescription creates the description for a dubbed video.
// It includes the translated description, timecodes, and a link to the original video.
func buildDubbedDescription(dubbingInfo storage.DubbingInfo, originalVideoID string) string {
	var parts []string

	// Add translated description
	if dubbingInfo.Description != "" {
		parts = append(parts, dubbingInfo.Description)
	}

	// Add timecodes if available
	if dubbingInfo.Timecodes != "" && dubbingInfo.Timecodes != "N/A" {
		timecodeSection := fmt.Sprintf("▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬\n%s", dubbingInfo.Timecodes)
		parts = append(parts, timecodeSection)
	}

	// Add link to original video
	if originalVideoID != "" {
		originalLink := fmt.Sprintf("▬▬▬▬▬▬ 🔗 Original Video 🔗 ▬▬▬▬▬▬\n🎬 English version: %s", GetYouTubeURL(originalVideoID))
		parts = append(parts, originalLink)
	}

	return strings.Join(parts, "\n\n")
}
