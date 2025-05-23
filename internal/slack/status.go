package slack

import (
	"fmt"

	"devopstoolkit/youtube-automation/internal/storage"
)

// updateSlackPostStatus_actual updates the video's SlackPosted status and saves it to disk.
func updateSlackPostStatus_actual(video *storage.Video, posted bool, videoPath string) error {
	// Update video metadata field
	video.SlackPosted = posted

	// Save updated video metadata
	yamlWriter := storage.YAML{IndexPath: ""}
	if err := yamlWriter.WriteVideo(*video, videoPath); err != nil {
		return fmt.Errorf("failed to update video metadata in %s: %w", videoPath, err)
	}

	return nil
}

// UpdateSlackPostStatus is a function variable for updating Slack post status.
// This allows it to be replaced for mocking in tests.
var UpdateSlackPostStatus = updateSlackPostStatus_actual
