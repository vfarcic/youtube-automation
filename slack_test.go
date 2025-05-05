package main

import (
	"testing"

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

	// Test cases
	tests := []struct {
		name     string
		videoId  string
		wantClip string
	}{
		{
			name:     "Basic video ID",
			videoId:  "abc123",
			wantClip: "https://youtu.be/abc123",
		},
		{
			name:     "Empty video ID",
			videoId:  "",
			wantClip: "https://youtu.be/",
		},
		{
			name:     "Video ID with special characters",
			videoId:  "xyz-789_ABC",
			wantClip: "https://youtu.be/xyz-789_ABC",
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
