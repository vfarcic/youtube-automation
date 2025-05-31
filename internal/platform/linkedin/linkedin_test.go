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

func TestPostToLinkedInWithConfig(t *testing.T) {
	testCases := []struct {
		name                string
		video               *storage.Video
		config              *Config
		expectError         bool
		expectPersonalURL   bool
		expectedProfileID   string
	}{
		{
			name: "successful post with default config",
			video: &storage.Video{
				Title:       "Test Video",
				Description: "Test Description",
				VideoId:     "test123",
			},
			config: &Config{
				AccessToken: "test-token",
				APIUrl:      "https://api.linkedin.com/v2",
				UsePersonal: false,
				ProfileID:   "",
			},
			expectError:       false,
			expectPersonalURL: false,
		},
		{
			name: "successful post with personal profile",
			video: &storage.Video{
				Title:       "Test Video",
				Description: "Test Description",
				VideoId:     "test123",
			},
			config: &Config{
				AccessToken: "test-token",
				APIUrl:      "https://api.linkedin.com/v2",
				UsePersonal: true,
				ProfileID:   "viktorfarcic",
			},
			expectError:       false,
			expectPersonalURL: true,
			expectedProfileID: "viktorfarcic",
		},
		{
			name:        "nil config",
			video:       &storage.Video{VideoId: "test123"},
			config:      nil,
			expectError: true,
		},
		{
			name: "empty access token",
			video: &storage.Video{
				VideoId: "test123",
			},
			config: &Config{
				AccessToken: "",
				APIUrl:      "https://api.linkedin.com/v2",
			},
			expectError: true,
		},
		{
			name: "usePersonal true but empty profileID",
			video: &storage.Video{
				VideoId: "test123",
			},
			config: &Config{
				AccessToken: "test-token",
				APIUrl:      "https://api.linkedin.com/v2",
				UsePersonal: true,
				ProfileID:   "",
			},
			expectError:       false,
			expectPersonalURL: false, // Should default to feed URL when profileID is empty
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test the function
			err := PostToLinkedInWithConfig(tc.video, tc.config)

			// Check the error result
			if tc.expectError && err == nil {
				t.Errorf("Expected error but got nil")
				return
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}
			
			// Skip further checks if we expected an error
			if tc.expectError {
				return
			}

			// Check the video was updated correctly
			if !tc.video.LinkedInPosted {
				t.Error("Expected LinkedInPosted to be true")
			}
			if tc.video.LinkedInPostURL == "" {
				t.Error("Expected LinkedInPostURL to be populated")
			}
			
			// Check URL format based on config
			if tc.expectPersonalURL {
				expectedPrefix := "https://www.linkedin.com/in/" + tc.expectedProfileID + "/detail/"
				if tc.video.LinkedInPostURL[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("Expected URL to start with %s but got %s", expectedPrefix, tc.video.LinkedInPostURL)
				}
			} else {
				expectedPrefix := "https://www.linkedin.com/feed/update/"
				if tc.video.LinkedInPostURL[:len(expectedPrefix)] != expectedPrefix {
					t.Errorf("Expected URL to start with %s but got %s", expectedPrefix, tc.video.LinkedInPostURL)
				}
			}
			
			// Check timestamp
			if tc.video.LinkedInPostTimestamp == "" {
				t.Error("Expected LinkedInPostTimestamp to be populated")
			}
			
			// Check if timestamp format is valid
			_, err = time.Parse(time.RFC3339, tc.video.LinkedInPostTimestamp)
			if err != nil {
				t.Errorf("LinkedInPostTimestamp is not in valid RFC3339 format: %v", err)
			}
		})
	}
}