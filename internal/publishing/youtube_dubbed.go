package publishing

import (
	"context"
	"fmt"
	"os"
	"strings"

	"devopstoolkit/youtube-automation/internal/storage"

	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

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
			PrivacyStatus: "private", // Start as private, user can change later
		},
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

	fmt.Printf("Dubbed video uploaded successfully! Video ID: %v\n", response.Id)
	return response.Id, nil
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
