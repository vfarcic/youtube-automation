package platform

import (
	"testing"

	"github.com/atotto/clipboard"
)

func TestPostLinkedIn(t *testing.T) {
	// Save original clipboard content to restore later
	originalContent, err := clipboard.ReadAll()
	if err != nil {
		t.Skipf("Clipboard access failed, skipping test: %v", err)
	}

	defer func() {
		// Restore original clipboard content
		clipboard.WriteAll(originalContent)
	}()

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
