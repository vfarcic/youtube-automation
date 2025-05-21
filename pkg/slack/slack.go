package slack

import (
	"fmt"
	"log"
	"time"

	"github.com/slack-go/slack"

	"devopstoolkitseries/youtube-automation/internal/storage"
)

// Config holds the configuration for Slack API
type Config struct {
	Token         string   // Slack API token
	DefaultChannel string  // Default channel for posting
	Reactions     []string // Reactions to add to the post
}

// GetConfig retrieves Slack configuration from the provided settings
func GetConfig(token, defaultChannel string, reactions []string) Config {
	// Check environment variable for token first
	return Config{
		Token:         token,
		DefaultChannel: defaultChannel,
		Reactions:     reactions,
	}
}

// ValidateConfig validates the Slack configuration
func ValidateConfig(config Config) error {
	if config.Token == "" {
		return fmt.Errorf("Slack token is required for posting")
	}

	if config.DefaultChannel == "" {
		log.Println("Warning: No default Slack channel configured, will need to specify channel for each post")
	}

	return nil
}

// PostMessage posts the video information to Slack
func PostMessage(config Config, video storage.Video) error {
	// Validate configuration
	if err := ValidateConfig(config); err != nil {
		return err
	}

	// Initialize Slack client
	client := slack.New(config.Token)
	
	// Determine appropriate channels
	channels := getChannelsForCategory(video.Category)
	if len(channels) == 0 {
		// Default channel if no category-specific ones are found
		channels = []string{config.DefaultChannel}
	}
	
	// Create message blocks with rich formatting
	blocks := createSlackMessageBlocks(video)
	
	// Post to each channel
	var lastError error
	successCount := 0
	postedChannels := []string{}
	
	for _, channel := range channels {
		// Post the message
		resp, timestamp, err := client.PostMessage(
			channel,
			slack.MsgOptionBlocks(blocks...),
			slack.MsgOptionAsUser(true),
		)
		
		if err != nil {
			lastError = fmt.Errorf("failed to post to channel %s: %w", channel, err)
			log.Printf("Error posting to Slack channel %s: %s", channel, err)
			continue
		}
		
		// Add reactions if configured
		if len(config.Reactions) > 0 {
			for _, reaction := range config.Reactions {
				err := client.AddReaction(reaction, slack.ItemRef{
					Channel:   resp,
					Timestamp: timestamp,
				})
				if err != nil {
					log.Printf("Warning: failed to add reaction %s: %s", reaction, err)
				}
			}
		}
		
		successCount++
		postedChannels = append(postedChannels, channel)
		log.Printf("Posted to Slack channel %s successfully", channel)
	}
	
	// Update video metadata if at least one post succeeded
	if successCount > 0 {
		video.SlackPosted = true
		video.SlackPostTimestamp = time.Now().Format(time.RFC3339)
		video.SlackPostChannels = postedChannels
		
		return nil
	}
	
	return lastError
}

// getChannelsForCategory returns a list of channels appropriate for the video category
// This is a placeholder implementation, can be extended with category->channel mappings
func getChannelsForCategory(category string) []string {
	// This can be extended with more sophisticated channel mapping logic
	return []string{}
}

// createSlackMessageBlocks creates rich Slack message blocks with video information
func createSlackMessageBlocks(video storage.Video) []slack.Block {
	youtubeURL := fmt.Sprintf("https://youtu.be/%s", video.VideoId)
	
	// Create text sections
	headerText := slack.NewTextBlockObject(
		slack.MarkdownType,
		fmt.Sprintf("*New Video:* %s", video.Title),
		false, false,
	)
	
	descriptionText := slack.NewTextBlockObject(
		slack.MarkdownType,
		getSlackSummary(video.Description),
		false, false,
	)
	
	linkText := slack.NewTextBlockObject(
		slack.MarkdownType,
		fmt.Sprintf("*Watch on YouTube:* <%s|%s>", youtubeURL, video.Title),
		false, false,
	)
	
	// Create blocks
	blocks := []slack.Block{
		slack.NewSectionBlock(headerText, nil, nil),
		slack.NewDividerBlock(),
		slack.NewSectionBlock(descriptionText, nil, nil),
		slack.NewSectionBlock(linkText, nil, nil),
	}
	
	// Add thumbnail if available
	if len(video.Thumbnail) > 0 {
		// Get thumbnail URL from video if available
		thumbnailURL := getThumbnailURL(video)
		if thumbnailURL != "" {
			accessory := slack.NewAccessory(
				slack.NewImageBlockElement(thumbnailURL, "Video Thumbnail"),
			)
			blocks = append(blocks, slack.NewSectionBlock(
				slack.NewTextBlockObject(slack.PlainTextType, "Thumbnail", false, false),
				nil,
				accessory,
			))
		}
	}
	
	return blocks
}

// getSlackSummary returns a shortened version of the description for Slack
func getSlackSummary(description string) string {
	// For now, just return the first 300 chars if longer
	if len(description) > 300 {
		return description[:297] + "..."
	}
	return description
}

// getThumbnailURL attempts to get a URL for the video thumbnail
// This is a placeholder - actual implementation would depend on how thumbnails are stored
func getThumbnailURL(video storage.Video) string {
	// In a real implementation, this could:
	// 1. Return a YouTube thumbnail URL based on VideoId
	// 2. Upload the local thumbnail to a CDN and return the URL
	// 3. Use a pre-uploaded URL stored in the video metadata
	
	// For now, just use YouTube's thumbnail if we have a VideoId
	if video.VideoId != "" {
		return fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", video.VideoId)
	}
	
	return ""
}