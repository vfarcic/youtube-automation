package platform

import (
	"fmt"
	"os"
	"strings"

	"devopstoolkit/youtube-automation/internal/platform/linkedin"
	"devopstoolkit/youtube-automation/internal/storage"
	"github.com/atotto/clipboard"
)

// PostLinkedIn posts content to LinkedIn
// It first attempts automated posting via the LinkedIn API
// If API access is not available, it falls back to the manual clipboard method
func PostLinkedIn(message, videoId string, getYouTubeURL func(string) string, confirmationStyle interface{ Render(...string) string }) {
	// Replace YouTube link placeholder with actual URL
	message = strings.ReplaceAll(message, "[YouTube Link]", getYouTubeURL(videoId))

	// Try to get LinkedIn access token from environment
	accessToken := os.Getenv("LINKEDIN_ACCESS_TOKEN")

	// If we don't have a token, fall back to clipboard method
	if accessToken == "" {
		// Fall back to manual clipboard method
		clipboard.WriteAll(message)
		println(confirmationStyle.Render("LinkedIn API access token not found. The message has been copied to clipboard. Please paste it into LinkedIn manually."))
		return
	}

	// Create a temporary video struct for the API call
	video := &storage.Video{
		Title:       "New Video", // This would normally be populated from the actual video
		Description: message,
		VideoId:     videoId,
	}

	// Attempt automated posting
	err := linkedin.PostToLinkedIn(video, accessToken)
	if err != nil {
		// Log the error
		fmt.Printf("LinkedIn API posting error: %v\n", err)
		
		// Fall back to manual clipboard method
		clipboard.WriteAll(message)
		println(confirmationStyle.Render(fmt.Sprintf("LinkedIn API posting failed: %v. The message has been copied to clipboard. Please paste it into LinkedIn manually.", err)))
		return
	}

	println(confirmationStyle.Render(fmt.Sprintf("Successfully posted to LinkedIn: %s", video.LinkedInPostURL)))
}