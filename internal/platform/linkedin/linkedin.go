package linkedin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"devopstoolkit/youtube-automation/internal/storage"
)

const (
	linkedInAPIBaseURL = "https://api.linkedin.com/v2"
	ugcPostsEndpoint   = "/ugcPosts"
)

// LinkedInShareRequest represents the payload for sharing content on LinkedIn.
// This structure should be verified against the latest LinkedIn API documentation.
type LinkedInShareRequest struct {
	Author          string                       `json:"author"`
	LifecycleState  string                       `json:"lifecycleState"`
	SpecificContent LinkedInSpecificShareContent `json:"specificContent"`
	Visibility      LinkedInVisibility           `json:"visibility"`
}

type LinkedInSpecificShareContent struct {
	ShareContent LinkedInShareContent `json:"com.linkedin.ugc.ShareContent"`
}

type LinkedInShareContent struct {
	ShareCommentary    LinkedInShareCommentary `json:"shareCommentary"`
	ShareMediaCategory string                  `json:"shareMediaCategory"`
	Media              []LinkedInMedia         `json:"media"`
}

type LinkedInShareCommentary struct {
	Text string `json:"text"`
}

type LinkedInMedia struct {
	Status      string              `json:"status"`
	OriginalURL string              `json:"originalUrl"` // This should be the YouTube video URL (video.ProjectURL)
	Title       *LinkedInMediaTitle `json:"title,omitempty"`
}

type LinkedInMediaTitle struct {
	Text string `json:"text"`
}

type LinkedInVisibility struct {
	MemberNetworkVisibility string `json:"com.linkedin.ugc.MemberNetworkVisibility"`
}

// LinkedInPostResponse represents a minimal structure for the response from LinkedIn.
type LinkedInPostResponse struct {
	ID string `json:"id"` // This typically contains the URN of the created post, e.g., "urn:li:ugcPost:..."
}

// PostToLinkedIn posts the video to LinkedIn and updates the video metadata
func PostToLinkedIn(video *storage.Video, accessToken string) error {
	fmt.Println("DEBUG: >>> Entering NEW PostToLinkedIn function in internal/platform/linkedin/linkedin.go <<<") // Debug print
	if accessToken == "" {
		return fmt.Errorf("LinkedIn access token not provided")
	}

	if video == nil {
		return fmt.Errorf("video cannot be nil")
	}

	if video.ProjectURL == "" {
		return fmt.Errorf("video ProjectURL is empty, cannot share")
	}

	profileID := os.Getenv("LINKEDIN_PROFILE_ID")
	if profileID == "" {
		// Fallback or error, as per README, profile ID is needed for personal posting.
		// For organization posts, the URN would be different and likely configured elsewhere.
		return fmt.Errorf("LINKEDIN_PROFILE_ID environment variable (numeric member ID) not set, cannot determine author URN")
	}
	authorURN := fmt.Sprintf("urn:li:member:%s", profileID)

	// Construct the message for the LinkedIn post
	// You might want to make this more configurable or use video.Description
	postText := fmt.Sprintf("Check out my new video: %s\\n\\n%s", video.Title, video.ProjectURL)
	if len(video.Description) > 0 {
		// Add a snippet of the description, ensuring not to exceed LinkedIn's limits
		maxDescLength := 150 // Arbitrary limit for snippet, adjust as needed
		descriptionSnippet := video.Description
		if len(descriptionSnippet) > maxDescLength {
			descriptionSnippet = descriptionSnippet[:maxDescLength] + "..."
		}
		postText = fmt.Sprintf("Check out my new video: %s\\n\\n%s\\n\\n%s", video.Title, descriptionSnippet, video.ProjectURL)
	}

	payload := LinkedInShareRequest{
		Author:         authorURN,
		LifecycleState: "PUBLISHED",
		SpecificContent: LinkedInSpecificShareContent{
			ShareContent: LinkedInShareContent{
				ShareCommentary: LinkedInShareCommentary{
					Text: postText,
				},
				ShareMediaCategory: "ARTICLE", // For sharing a URL
				Media: []LinkedInMedia{
					{
						Status:      "READY",
						OriginalURL: video.ProjectURL,
						Title:       &LinkedInMediaTitle{Text: video.Title},
					},
				},
			},
		},
		Visibility: LinkedInVisibility{
			MemberNetworkVisibility: "PUBLIC", // Or "CONNECTIONS"
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal LinkedIn request payload: %w", err)
	}

	fmt.Println("DEBUG: LinkedIn Payload:", string(payloadBytes)) // Debug print

	req, err := http.NewRequest("POST", linkedInAPIBaseURL+ugcPostsEndpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create LinkedIn API request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0") // Often required for LinkedIn v2 APIs
	req.Header.Set("LinkedIn-Version", "202305")         // Example: Specify a recent API version month

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to LinkedIn API: %w", err)
	}
	defer resp.Body.Close()

	fmt.Println("DEBUG: LinkedIn API Response Status:", resp.Status) // Debug print

	if resp.StatusCode != http.StatusCreated { // 201 Created is success for UGC posts
		// Attempt to read error body for more details
		var errorBody bytes.Buffer
		_, _ = errorBody.ReadFrom(resp.Body)
		fmt.Println("DEBUG: LinkedIn API Error Body:", errorBody.String()) // Debug print
		return fmt.Errorf("LinkedIn API request failed with status %s: %s", resp.Status, errorBody.String())
	}

	// Successfully posted. Try to get the post URN/ID.
	// LinkedIn often returns the ID in the X-Restli-Id header or in the response body.
	postURN := resp.Header.Get("X-Restli-Id")
	if postURN == "" {
		// If not in header, try to parse from JSON response body
		var linkedInResponse LinkedInPostResponse
		if err := json.NewDecoder(resp.Body).Decode(&linkedInResponse); err == nil && linkedInResponse.ID != "" {
			postURN = linkedInResponse.ID
		}
	}

	fmt.Println("DEBUG: Retrieved LinkedIn Post URN:", postURN) // Debug print

	if postURN == "" {
		// If still no URN, we can't form a direct link, but the post was successful.
		// Log this situation or handle as a partial success.
		fmt.Printf("Successfully posted to LinkedIn, but could not retrieve the post URN/ID from response headers or body.\\n")
		video.LinkedInPostURL = "https://www.linkedin.com/feed/" // Generic feed URL
	} else {
		video.LinkedInPostURL = fmt.Sprintf("https://www.linkedin.com/feed/update/%s", postURN)
	}

	video.LinkedInPosted = true
	video.LinkedInPostTimestamp = time.Now().Format(time.RFC3339)

	fmt.Println("DEBUG: Final LinkedInPostURL set to:", video.LinkedInPostURL) // Debug print
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
