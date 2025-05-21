package main

import (
	"os"
	"path/filepath"
	"testing"

	"devopstoolkitseries/youtube-automation/internal/configuration"
	"devopstoolkitseries/youtube-automation/internal/storage"

	"github.com/atotto/clipboard"
)

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

	// Create a temporary index file for testing
	tmpDir, err := os.MkdirTemp("", "slack-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test cases
	tests := []struct {
		name     string
		videoId  string
		wantClip string
		setupEnv bool // Whether to set up the Slack token env var
	}{
		{
			name:     "Basic video ID",
			videoId:  "abc123",
			wantClip: "https://youtu.be/abc123",
			setupEnv: false,
		},
		{
			name:     "Empty video ID",
			videoId:  "",
			wantClip: "https://youtu.be/",
			setupEnv: false,
		},
		{
			name:     "Video ID with special characters",
			videoId:  "xyz-789_ABC",
			wantClip: "https://youtu.be/xyz-789_ABC",
			setupEnv: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
