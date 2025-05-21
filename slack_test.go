package main

import (
	"testing"

	"devopstoolkitseries/youtube-automation/internal/configuration"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	// Initialize styles for testing
	confirmationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
}

func TestPostSlack(t *testing.T) {
	// Save original clipboard content to restore later
	originalContent, err := clipboard.ReadAll()
	if err != nil {
		t.Skipf("Clipboard access failed, skipping test: %v", err)
	}

	defer func() {
		// Restore original clipboard content
		clipboard.WriteAll(originalContent)
	}()

	// Save original Slack token to restore after tests
	originalToken := configuration.GlobalSettings.Slack.Token
	defer func() {
		configuration.GlobalSettings.Slack.Token = originalToken
	}()

	// Test cases
	tests := []struct {
		name     string
		videoId  string
		wantClip string
		token    string // Slack token to use
	}{
		{
			name:     "Basic video ID without token",
			videoId:  "abc123",
			wantClip: "https://youtu.be/abc123",
			token:    "", // No token means it should fall back to clipboard
		},
		{
			name:     "Empty video ID",
			videoId:  "",
			wantClip: "https://youtu.be/",
			token:    "",
		},
		{
			name:     "Video ID with special characters",
			videoId:  "xyz-789_ABC",
			wantClip: "https://youtu.be/xyz-789_ABC",
			token:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up token for this test
			configuration.GlobalSettings.Slack.Token = tt.token
			
			// Call the function being tested
			postSlack(tt.videoId)

			// Check what was set to clipboard
			gotClip, err := clipboard.ReadAll()
			if err != nil {
				t.Fatalf("Failed to read clipboard: %v", err)
			}

			if gotClip != tt.wantClip {
				t.Errorf("postSlack() set clipboard to %q, want %q", gotClip, tt.wantClip)
			}
		})
	}
}
