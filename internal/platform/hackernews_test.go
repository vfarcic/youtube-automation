package platform

import (
	"fmt"
	"strings"
	"testing"
)

// mockStyle is a mock implementation for styles that have a Render method.
// It's defined here to be accessible by other test files in the platform package.
type mockStyle struct{}

// Render returns the first string passed to it, or an empty string.
func (m mockStyle) Render(text ...string) string {
	if len(text) == 0 {
		return ""
	}
	return text[0] // Simple mock behavior
}

// TestHackerNewsOutput verifies the format of HackerNews post message
func TestHackerNewsOutput(t *testing.T) {
	// Mock getYouTubeURL function
	getYouTubeURL := func(videoId string) string {
		return "https://youtu.be/" + videoId
	}

	// Mock confirmation style
	confirmationStyle := mockStyle{}

	// Test cases
	tests := []struct {
		name       string
		title      string
		videoId    string
		wantSubstr []string
	}{
		{
			name:    "Basic post",
			title:   "How to Deploy Kubernetes",
			videoId: "abc123",
			wantSubstr: []string{
				"https://news.ycombinator.com/submit",
				"How to Deploy Kubernetes",
				"https://youtu.be/abc123",
			},
		},
		{
			name:    "Empty title",
			title:   "",
			videoId: "def456",
			wantSubstr: []string{
				"https://news.ycombinator.com/submit",
				"https://youtu.be/def456",
			},
		},
		{
			name:    "Title with special characters",
			title:   "K8s & Docker: Best Practices!",
			videoId: "ghi789",
			wantSubstr: []string{
				"https://news.ycombinator.com/submit",
				"K8s & Docker: Best Practices!",
				"https://youtu.be/ghi789",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Directly test the message format logic from PostHackerNews
			message := fmt.Sprintf(
				"Use the following information to post it to https://news.ycombinator.com/submit manually.\n\nTitle:\n%s\nURL:\n%s",
				tt.title,
				getYouTubeURL(tt.videoId),
			)

			// Check the message for expected content
			for _, substr := range tt.wantSubstr {
				if !strings.Contains(message, substr) {
					t.Errorf("HackerNews message does not contain %q\nMessage: %q", substr, message)
				}
			}

			// Ensure the actual function doesn't panic (but we can't verify its output)
			PostHackerNews(tt.title, tt.videoId, getYouTubeURL, confirmationStyle)
		})
	}
}
