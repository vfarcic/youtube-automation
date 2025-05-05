package examples

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// VideoMetadata represents a YouTube video metadata
type VideoMetadata struct {
	Title       string
	Description string
	Tags        []string
	Category    string
}

// FormatVideoTitle formats a video title according to platform guidelines
func FormatVideoTitle(metadata VideoMetadata) string {
	if metadata.Title == "" {
		return "Untitled Video"
	}

	// Truncate title if too long (YouTube limit is 100 characters)
	if len(metadata.Title) > 100 {
		return metadata.Title[:97] + "..."
	}

	return metadata.Title
}

// TestFormatVideoTitle demonstrates a well-structured unit test
func TestFormatVideoTitle(t *testing.T) {
	// Table-driven tests with descriptive names and expected outcomes
	tests := []struct {
		name     string
		metadata VideoMetadata
		expected string
	}{
		{
			name: "Standard title",
			metadata: VideoMetadata{
				Title: "How to Build a Go Application",
			},
			expected: "How to Build a Go Application",
		},
		{
			name: "Empty title",
			metadata: VideoMetadata{
				Title: "",
			},
			expected: "Untitled Video",
		},
		{
			name: "Overly long title gets truncated",
			metadata: VideoMetadata{
				Title: "This is an extremely long video title that exceeds the YouTube character limit and should be truncated by our formatting function",
			},
			expected: "This is an extremely long video title that exceeds the YouTube character limit and should be truncat...",
		},
	}

	// Execute each test case
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act - call the function being tested
			result := FormatVideoTitle(tt.metadata)

			// Assert - verify the results match expectations
			assert.Equal(t, tt.expected, result, "FormatVideoTitle should properly format the video title")
		})
	}
}
