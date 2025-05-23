package slack

import (
	"fmt"
	"strings"

	// "devopstoolkit/youtube-automation/internal/configuration" // Removed as GetTargetChannels is in current (slack) package
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/slack-go/slack"
)

// PostStatus struct removed as it's unused.

// FailedAttempt struct removed as it's unused

// SlackService orchestrates Slack posting.
type SlackService struct {
	config *SlackConfig
	client *SlackClient
}

// NewSlackService creates a new service with the given config.
func NewSlackService(cfg *SlackConfig) (*SlackService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("invalid Slack configuration: config is nil")
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Slack configuration: %w", err)
	}

	auth, err := NewSlackAuth(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create Slack auth: %w", err)
	}

	client, err := NewSlackClient(auth)
	if err != nil {
		return nil, fmt.Errorf("failed to create Slack client: %w", err)
	}

	return &SlackService{
		config: cfg,
		client: client,
	}, nil
}

// PostVideo posts a video to appropriate Slack channels.
func (s *SlackService) PostVideo(video *storage.Video, videoPath string) error {
	// Get channels using global configuration
	channels := GetTargetChannels()
	if len(channels) == 0 {
		// If no channels are configured, this might be a configuration error or an expected state.
		// For now, returning an error. Could also log a warning and return nil if appropriate.
		LogSlackWarn("No target Slack channels configured globally.")
		return fmt.Errorf("no target Slack channels configured")
	}

	// Message construction similar to messages.PostVideoThumbnail
	if video.VideoId == "" {
		return fmt.Errorf("cannot post to Slack: VideoId is empty for video %q", video.Name)
	}
	// Consider if Thumbnail check is needed here or if it's handled by the client/message options
	// if video.Thumbnail == "" {
	// \treturn fmt.Errorf("cannot post to Slack: Thumbnail URL is empty for video %q (ID: %s)", video.Name, video.VideoId)
	// }

	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.VideoId)
	videoTitle := video.Name // Using Name as title, as per storage.Video struct
	if videoTitle == "" {
		videoTitle = "New Video Released!" // Default title
	}
	// thumbnailURL := video.Thumbnail // Commented out as ImageURL is removed

	messageText := fmt.Sprintf("📺 New Video\\n%s", videoURL) // URL now part of main message text, using actual TV emoji
	attachment := slack.Attachment{
		Fallback:  fmt.Sprintf("%s - %s", videoTitle, videoURL),
		Title:     videoTitle,
		TitleLink: videoURL, // Slack will use this to unfurl and show a preview/thumbnail
		// ImageURL:  thumbnailURL, // Removed to let Slack unfurl the TitleLink
		// ThumbURL: thumbnailURL, // Some Slack SDK versions might use ThumbURL for attachments
	}

	// Simplified status tracking, focusing on whether any post succeeded.
	var anyPostSucceeded bool // Default is false

	var lastError error
	successCount := 0

	for _, channelID := range channels { // Changed loop variable to channelID for clarity
		_, timestamp, err := s.client.PostMessage(
			channelID,
			slack.MsgOptionText(messageText, false),
			slack.MsgOptionAttachments(attachment),
		)

		if err != nil {
			lastError = err
			slackErr := CategorizeError(err)
			LogSlackError(slackErr, fmt.Sprintf("Failed to post to channel %s", channelID))
			// RecordFailedPost(status, channelID, slackErr.Type, slackErr.Message) // Call removed as FailedAttempts is gone
			continue
		}

		anyPostSucceeded = true // Mark true if any post succeeds
		// Construct message URL (placeholder TEAM_ID needs to be configured or discovered)
		// This format might need adjustment for different Slack workspace URL structures or if using Enterprise Grid.
		// A more robust way might involve using response data if available, or a configured workspace URL.
		// For now, using a known placeholder format, which may not always work.
		// Consider making TEAM_ID part of SlackConfig.
		messageLink := fmt.Sprintf("https://slack.com/archives/%s/p%s", channelID, strings.ReplaceAll(timestamp, ".", ""))

		// RecordSuccessfulPost call removed as PostStatus struct is gone
		successCount++
		LogSlackInfo(fmt.Sprintf("Posted to Slack channel %s (msg ID: %s, link: %s) successfully", channelID, timestamp, messageLink))
	}

	// Update video metadata using the existing function from status.go
	// This requires videoPath to be passed to PostVideo.
	if anyPostSucceeded { // Only update if at least one post was successful
		if err := UpdateSlackPostStatus(video, true, videoPath); err != nil {
			LogSlackWarn(fmt.Sprintf("Failed to update video metadata (SlackPosted=true) for %s: %s", videoPath, err.Error()))
			// Decide if this should contribute to lastError or be returned differently
		}
	} else if lastError != nil { // If all posts failed (anyPostSucceeded is false and there was an error)
		if err := UpdateSlackPostStatus(video, false, videoPath); err != nil {
			LogSlackWarn(fmt.Sprintf("Failed to update video metadata (SlackPosted=false) for %s after all posts failed: %s", videoPath, err.Error()))
		}
	}

	if successCount == 0 && lastError != nil {
		return fmt.Errorf("failed to post to any Slack channel: %w", lastError)
	}
	if successCount == 0 && lastError == nil { // Should ideally not happen if channels list is not empty
		return fmt.Errorf("failed to post to any Slack channel, no specific error reported (target channels: %v)", channels)
	}

	return nil
}

// Placeholder functions for status tracking - these should ideally be in status.go
// func RecordFailedPost(status *PostStatus, channel string, errorType ErrorType, message string) { // Function removed
// }

// func RecordSuccessfulPost(status *PostStatus, channel, timestamp, messageURL string) { // Function removed
// }

// MessageBuilder and its methods are removed as PostVideo now handles message construction directly.
// ChannelMapping struct and its methods are removed as PostVideo now uses configuration.GetTargetChannels().
