package slack

import (
	"fmt"

	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/slack-go/slack"
)

// PostVideoThumbnail posts a simple message with the video thumbnail linking to YouTube.
// It uses the provided internal Slack client and video details.
func PostVideoThumbnail(client *SlackClient, channelID string, videoDetails storage.Video) error {
	videoTitle := videoDetails.GetUploadTitle()
	if videoTitle == "" {
		return fmt.Errorf("cannot post to Slack: video title is empty")
	}
	if videoDetails.VideoId == "" {
		return fmt.Errorf("cannot post to Slack: VideoId is empty for video %q", videoTitle)
	}
	if videoDetails.Thumbnail == "" {
		// We could still post without a thumbnail, but the task is focused on thumbnail posting.
		// For now, let's consider it an issue if the thumbnail URL is missing.
		return fmt.Errorf("cannot post to Slack: Thumbnail URL is empty for video %q (ID: %s)", videoTitle, videoDetails.VideoId)
	}

	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoDetails.VideoId)
	thumbnailURL := videoDetails.Thumbnail

	// The message text that appears above the attachment
	messageText := fmt.Sprintf("📺 New Video: %s\n%s", videoTitle, videoURL)

	attachment := slack.Attachment{
		// Fallback text for notifications
		Fallback: fmt.Sprintf("%s - %s", videoTitle, videoURL),
		// Title of the attachment block
		Title: videoTitle,
		// Link for the attachment title
		TitleLink: videoURL,
		// ImageURL is the URL of the image to display as a thumbnail
		ImageURL: thumbnailURL,
		// pretext (text that appears above attachment, but slack.MsgOptionText is better for main text)
		// text (main text of attachment - we are putting main text in MsgOptionText)
	}

	// Our internal SlackClient.PostMessage expects attachments as a slice
	// and uses slack.MsgOption for message construction.
	_, _, err := client.PostMessage(
		channelID,
		slack.MsgOptionText(messageText, false), // false for not using markdown escaping
		slack.MsgOptionAttachments(attachment),  // MsgOptionAttachments takes a variadic number of attachments
	)
	if err != nil {
		return fmt.Errorf("failed to post Slack message for video ID %s to channel %s: %w", videoDetails.VideoId, channelID, err)
	}

	return nil
}
