package linkedin

import (
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestPostToLinkedIn(t *testing.T) {
	testCases := []struct {
		name        string
		video       *storage.Video
		accessToken string
		expectError bool
	}{
		{
			name: "successful post",
			video: &storage.Video{
				Title:       "Test Video",
				Description: "Test Description",
				VideoId:     "test123",
			},
			accessToken: "test-token",
			expectError: false,
		},
		{
			name:        "nil video",
			video:       nil,
			accessToken: "test-token",
			expectError: true,
		},
		{
			name: "empty access token",
			video: &storage.Video{
				Title:       "Test Video",
				Description: "Test Description",
				VideoId:     "test123",
			},
			accessToken: "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the function
			err := PostToLinkedIn(tc.video, tc.accessToken)

			// Check the error result
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check the video was updated correctly if successful
			if !tc.expectError && err == nil && tc.video != nil {
				if !tc.video.LinkedInPosted {
					t.Error("Expected LinkedInPosted to be true")
				}
				if tc.video.LinkedInPostURL == "" {
					t.Error("Expected LinkedInPostURL to be populated")
				}
				if tc.video.LinkedInPostTimestamp == "" {
					t.Error("Expected LinkedInPostTimestamp to be populated")
				}

				// Check if timestamp format is valid
				_, err := time.Parse(time.RFC3339, tc.video.LinkedInPostTimestamp)
				if err != nil {
					t.Errorf("LinkedInPostTimestamp is not in valid RFC3339 format: %v", err)
				}
			}
		})
	}
}