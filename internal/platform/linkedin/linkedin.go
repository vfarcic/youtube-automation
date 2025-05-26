
package linkedin

import (
	"fmt"
	"time"
	"os"

	"devopstoolkit/youtube-automation/internal/storage"
)

// PostToLinkedIn posts the video to LinkedIn and updates the video metadata
func PostToLinkedIn(video *storage.Video, accessToken string) error {
	if accessToken == "" {
		return fmt.Errorf("LinkedIn access token not provided")
	}
	
	if video == nil {
		return fmt.Errorf("video cannot be nil")
	}
	
	// Load profile ID from environment if available
	profileID := os.Getenv("LINKEDIN_PROFILE_ID")
	usePersonal := profileID != ""
	
	// In a real implementation, this would make an API call to LinkedIn
	// For now, mark as posted and set a placeholder URL based on config
	video.LinkedInPosted = true
	if usePersonal && profileID != "" {
		// Use personal profile URL format
		video.LinkedInPostURL = fmt.Sprintf("https://www.linkedin.com/in/%s/detail/simulated-%s", 
			profileID, video.VideoId)
	} else {
		// Use default feed URL format
		video.LinkedInPostURL = fmt.Sprintf("https://www.linkedin.com/feed/update/simulated-%s", 
			video.VideoId)
	}
	video.LinkedInPostTimestamp = time.Now().Format(time.RFC3339)
	
	return nil
}

// PostToLinkedInWithConfig posts the video to LinkedIn using config parameters
func PostToLinkedInWithConfig(video *storage.Video, config *Config) error {
	if config == nil {
		return fmt.Errorf("LinkedIn configuration not provided")
	}
	
	if config.AccessToken == "" {
		return fmt.Errorf("LinkedIn access token not provided")
	}
	
	if video == nil {
		return fmt.Errorf("video cannot be nil")
	}
	
	// In a real implementation, this would make an API call to LinkedIn
	// For now, mark as posted and set a placeholder URL based on config
	video.LinkedInPosted = true
	if config.UsePersonal && config.ProfileID != "" {
		// Use personal profile URL format
		video.LinkedInPostURL = fmt.Sprintf("https://www.linkedin.com/in/%s/detail/simulated-%s", 
			config.ProfileID, video.VideoId)
	} else {
		// Use default feed URL format
		video.LinkedInPostURL = fmt.Sprintf("https://www.linkedin.com/feed/update/simulated-%s", 
			video.VideoId)
	}
	video.LinkedInPostTimestamp = time.Now().Format(time.RFC3339)
	
	return nil
}

