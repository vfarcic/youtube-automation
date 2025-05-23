package slack

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/slack-go/slack"
)

const (
	defaultSummary        = "No summary available."
	thumbnailURLFormat    = "https://i.ytimg.com/vi/%s/hqdefault.jpg"
	thumbnailAltText      = "Video thumbnail"
	maxDescriptionLength  = 3000
	dateFormat            = "Jan 2, 2006"
	inputDateFormat       = "2006-01-02T15:04"
	tagSeparator          = " | "
	tagInputSeparator     = ","
	youtubeWatchURLFormat = "https://www.youtube.com/watch?v=%s"
)

// BuildHeaderBlock creates a Slack Header block for a video.
// Actual implementation will be in Subtask 4.3.
func BuildHeaderBlock(videoDetails storage.Video) (*slack.HeaderBlock, error) {
	if videoDetails.Title == "" {
		return nil, errors.New("video title cannot be empty for header block")
	}
	headerText := slack.NewTextBlockObject(slack.PlainTextType, videoDetails.Title, false, false)
	headerBlock := slack.NewHeaderBlock(headerText)
	return headerBlock, nil
}

// BuildSectionBlockWithThumbnail creates a Slack Section block with text and a thumbnail image.
func BuildSectionBlockWithThumbnail(videoDetails storage.Video) (*slack.SectionBlock, error) {
	if videoDetails.VideoId == "" {
		return nil, errors.New("VideoId cannot be empty for section block with thumbnail")
	}

	var textContent string
	if videoDetails.Highlight != "" {
		textContent = videoDetails.Highlight
	} else if videoDetails.Description != "" {
		textContent = videoDetails.Description
	} else {
		textContent = defaultSummary
	}

	// Truncate textContent if it exceeds Slack's limit for section text
	if len(textContent) > maxDescriptionLength {
		textContent = textContent[:maxDescriptionLength-3] + "..."
	}

	textBlock := slack.NewTextBlockObject(slack.MarkdownType, textContent, false, false)

	thumbnailURL := fmt.Sprintf(thumbnailURLFormat, videoDetails.VideoId)
	imageAltText := videoDetails.Title
	if imageAltText == "" { // Use default alt text if title is empty
		imageAltText = thumbnailAltText
	}
	imageAccessory := slack.NewAccessory(slack.NewImageBlockElement(thumbnailURL, imageAltText))

	sectionBlock := slack.NewSectionBlock(textBlock, nil, imageAccessory)
	return sectionBlock, nil
}

// BuildContextBlock creates a Slack Context block with information like view count, likes, etc.
func BuildContextBlock(videoDetails storage.Video) (*slack.ContextBlock, error) {
	elements := []slack.MixedElement{}

	if videoDetails.Date != "" {
		parsedTime, err := time.Parse(inputDateFormat, videoDetails.Date)
		if err == nil {
			dateStr := parsedTime.Format(dateFormat)
			dateElement := slack.NewTextBlockObject(slack.MarkdownType, dateStr, false, false)
			elements = append(elements, dateElement)
		}
	}

	if videoDetails.Category != "" {
		categoryElement := slack.NewTextBlockObject(slack.PlainTextType, videoDetails.Category, false, false)
		elements = append(elements, categoryElement)
	}

	if videoDetails.Tags != "" {
		tags := strings.Split(videoDetails.Tags, tagInputSeparator)
		var validTags []string
		for _, tag := range tags {
			trimmedTag := strings.TrimSpace(tag)
			if trimmedTag != "" {
				validTags = append(validTags, trimmedTag)
			}
		}
		if len(validTags) > 0 {
			tagsStr := strings.Join(validTags, tagSeparator)
			tagsElement := slack.NewTextBlockObject(slack.PlainTextType, tagsStr, false, false)
			elements = append(elements, tagsElement)
		}
	}

	if len(elements) == 0 {
		return nil, nil
	}

	contextBlock := slack.NewContextBlock("", elements...)
	return contextBlock, nil
}

// BuildActionsBlock creates a Slack Actions block with buttons (e.g., "Watch Video").
func BuildActionsBlock(videoDetails storage.Video) (*slack.ActionBlock, error) {
	if videoDetails.VideoId == "" {
		return nil, errors.New("VideoId is required to create an ActionsBlock with a Watch Video button")
	}

	buttons := []slack.BlockElement{}

	// Watch Video button
	watchURL := fmt.Sprintf(youtubeWatchURLFormat, videoDetails.VideoId)
	watchText := slack.NewTextBlockObject(slack.PlainTextType, "▶️ Watch Video", true, false)
	watchButton := slack.NewButtonBlockElement("watch_video_button", videoDetails.VideoId, watchText)
	watchButton.URL = watchURL
	watchButton.Style = slack.StylePrimary
	buttons = append(buttons, watchButton)

	// Project Details button
	if videoDetails.ProjectURL != "" {
		projectText := slack.NewTextBlockObject(slack.PlainTextType, "Project Details", false, false)
		projectButton := slack.NewButtonBlockElement("project_details_button", videoDetails.ProjectURL, projectText)
		projectButton.URL = videoDetails.ProjectURL
		buttons = append(buttons, projectButton)
	}

	// This check should ideally not be hit if VideoId is mandatory as per the first check.
	// However, if logic changes, it's a safeguard.
	if len(buttons) == 0 {
		return nil, nil // Or an error indicating no actionable content
	}

	actionBlock := slack.NewActionBlock("", buttons...)
	return actionBlock, nil
}

// BuildMessage orchestrates the creation of all blocks for a Slack message.
// It will call the other Build*Block functions.
func BuildMessage(videoDetails storage.Video) ([]slack.Block, error) {
	var blocks []slack.Block

	header, err := BuildHeaderBlock(videoDetails)
	if err != nil {
		return nil, err
	}
	if header != nil {
		blocks = append(blocks, header)
	}

	section, err := BuildSectionBlockWithThumbnail(videoDetails)
	if err != nil {
		return nil, err
	}
	if section != nil {
		blocks = append(blocks, section)
	}

	context, err := BuildContextBlock(videoDetails)
	if err != nil {
		return nil, err
	}
	if context != nil {
		blocks = append(blocks, context)
	}

	actions, err := BuildActionsBlock(videoDetails)
	if err != nil {
		return nil, err
	}
	if actions != nil {
		blocks = append(blocks, actions)
	}

	return blocks, nil
}
