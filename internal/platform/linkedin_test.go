package platform

import (
	"os"
	"testing"

	"github.com/atotto/clipboard"
)

func TestPostLinkedIn(t *testing.T) {
	// Save original clipboard content to restore later
	originalContent, err := clipboard.ReadAll()
	if err != nil {
		t.Skipf("Clipboard access failed, skipping test: %v", err)
	}

	// Save original LINKEDIN_ACCESS_TOKEN environment variable to restore later
	originalToken := os.Getenv("LINKEDIN_ACCESS_TOKEN")
	
	defer func() {
		// Restore original clipboard content
		clipboard.WriteAll(originalContent)
		
		// Restore the original environment variable
		os.Setenv("LINKEDIN_ACCESS_TOKEN", originalToken)
	}()
	
	// Ensure no token is set to force clipboard behavior
	os.Unsetenv("LINKEDIN_ACCESS_TOKEN")

	// Mock getYouTubeURL function
	getYouTubeURL := func(videoId string) string {
		return "https://youtu.be/" + videoId
	}

	// Mock confirmation style
	confirmationStyle := mockStyle{}

	// Test cases
	tests := []struct {
		name     string
		message  string
		videoId  string
		wantClip string
	}{
		{
			name:     "Basic message with YouTube link placeholder",
			message:  "Check out my new video: [YouTube Link]",
			videoId:  "abc123",
			wantClip: "Check out my new video: https://youtu.be/abc123",
		},
		{
			name:     "Message without placeholder",
			message:  "Just a regular message",
			videoId:  "xyz789",
			wantClip: "Just a regular message",
		},
		{
			name:     "Multiple placeholders",
			message:  "Link: [YouTube Link] and again [YouTube Link]",
			videoId:  "def456",
			wantClip: "Link: https://youtu.be/def456 and again https://youtu.be/def456",
		},
		{
			name:     "Empty message",
			message:  "",
			videoId:  "ghi789",
			wantClip: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the function being tested
			PostLinkedIn(tt.message, tt.videoId, getYouTubeURL, confirmationStyle)

			// Check what was set to clipboard
			gotClip, err := clipboard.ReadAll()
			if err != nil {
				t.Fatalf("Failed to read clipboard: %v", err)
			}

			if gotClip != tt.wantClip {
				t.Errorf("PostLinkedIn() set clipboard to %q, want %q", gotClip, tt.wantClip)
			}
		})
	}
}

func TestPostLinkedInWithToken(t *testing.T) {
	// Skip this test in CI environments where we can't set real tokens
	if os.Getenv("CI") != "" {
		t.Skip("Skipping LinkedIn token test in CI environment")
	}

	// Save original environment variable to restore later
	originalToken := os.Getenv("LINKEDIN_ACCESS_TOKEN")
	
	defer func() {
		// Restore the original environment variable
		os.Setenv("LINKEDIN_ACCESS_TOKEN", originalToken)
	}()
	
	// Set a fake token for testing the API path
	os.Setenv("LINKEDIN_ACCESS_TOKEN", "test-token")

	// Mock getYouTubeURL function
	getYouTubeURL := func(videoId string) string {
		return "https://youtu.be/" + videoId
	}

	// Mock confirmation style
	confirmationStyle := mockStyle{}

	// Test automated posting
	message := "Check out my new video: [YouTube Link]"
	videoId := "test123"
	
	// Call the function being tested
	PostLinkedIn(message, videoId, getYouTubeURL, confirmationStyle)
	
	// Success is indicated by not panicking and the message from confirmationStyle
	// In a more comprehensive test, we could mock the LinkedIn package as well
}
