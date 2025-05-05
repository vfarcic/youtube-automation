package examples

import (
	"errors"
	"testing"
)

// YouTubeAPI defines the interface for interacting with the YouTube API
type YouTubeAPI interface {
	UploadVideo(title, description string, tags []string, categoryID, videoPath, thumbnailPath string) (string, error)
	GetVideoCategories() (map[string]string, error)
}

// MockYouTubeAPI implements the YouTubeAPI interface for testing
type MockYouTubeAPI struct {
	// Response values to return
	UploadVideoResponse string
	UploadVideoError    error
	Categories          map[string]string
	CategoryError       error

	// Call tracking for verification
	UploadCalled        bool
	UploadTitleParam    string
	UploadDescParam     string
	UploadCategoryParam string
	UploadTagsParam     []string
	UploadVideoParam    string
	UploadThumbParam    string
}

// UploadVideo implements the YouTubeAPI interface for tests
func (m *MockYouTubeAPI) UploadVideo(title, description string, tags []string, categoryID, videoPath, thumbnailPath string) (string, error) {
	// Track that this method was called and with what parameters
	m.UploadCalled = true
	m.UploadTitleParam = title
	m.UploadDescParam = description
	m.UploadTagsParam = tags
	m.UploadCategoryParam = categoryID
	m.UploadVideoParam = videoPath
	m.UploadThumbParam = thumbnailPath

	// Return the pre-configured response
	return m.UploadVideoResponse, m.UploadVideoError
}

// GetVideoCategories implements the YouTubeAPI interface for tests
func (m *MockYouTubeAPI) GetVideoCategories() (map[string]string, error) {
	return m.Categories, m.CategoryError
}

// NewMockYouTubeAPI creates a new mock with default success values
func NewMockYouTubeAPI() *MockYouTubeAPI {
	return &MockYouTubeAPI{
		UploadVideoResponse: "mock-video-id-123",
		UploadVideoError:    nil,
		Categories: map[string]string{
			"1":  "Film & Animation",
			"10": "Music",
			"22": "People & Blogs",
		},
		CategoryError: nil,
	}
}

// Example of a function that uses the YouTube API
func PublishVideo(youtube YouTubeAPI, metadata VideoMetadata, videoPath, thumbnailPath string) (string, error) {
	// Format the metadata
	formattedTitle := FormatVideoTitle(metadata)

	// Validate category
	categories, err := youtube.GetVideoCategories()
	if err != nil {
		return "", errors.New("failed to get video categories: " + err.Error())
	}

	_, exists := categories[metadata.Category]
	if !exists {
		return "", errors.New("invalid video category")
	}

	// Upload the video
	videoID, err := youtube.UploadVideo(
		formattedTitle,
		metadata.Description,
		metadata.Tags,
		metadata.Category,
		videoPath,
		thumbnailPath,
	)

	if err != nil {
		return "", errors.New("failed to upload video: " + err.Error())
	}

	return videoID, nil
}

// TestPublishVideo demonstrates how to use the mock in a test
func TestPublishVideo(t *testing.T) {
	// Test case 1: Successful upload
	t.Run("Successful upload", func(t *testing.T) {
		// Create a mock with success responses
		mockAPI := NewMockYouTubeAPI()

		// Create test data
		metadata := VideoMetadata{
			Title:       "Test Video",
			Description: "This is a test video",
			Tags:        []string{"test", "demo"},
			Category:    "22", // People & Blogs
		}

		// Call the function being tested
		videoID, err := PublishVideo(mockAPI, metadata, "test.mp4", "thumb.jpg")

		// Assert results
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if videoID != "mock-video-id-123" {
			t.Errorf("Expected video ID 'mock-video-id-123', got: %s", videoID)
		}

		// Verify the mock was called with expected parameters
		if !mockAPI.UploadCalled {
			t.Error("Expected UploadVideo to be called")
		}

		if mockAPI.UploadTitleParam != "Test Video" {
			t.Errorf("Expected title 'Test Video', got: %s", mockAPI.UploadTitleParam)
		}
	})

	// Test case 2: Invalid category
	t.Run("Invalid category", func(t *testing.T) {
		// Create a mock with success responses
		mockAPI := NewMockYouTubeAPI()

		// Create test data with invalid category
		metadata := VideoMetadata{
			Title:       "Test Video",
			Description: "This is a test video",
			Tags:        []string{"test", "demo"},
			Category:    "999", // Invalid category
		}

		// Call the function being tested
		_, err := PublishVideo(mockAPI, metadata, "test.mp4", "thumb.jpg")

		// Assert error is returned
		if err == nil {
			t.Error("Expected an error for invalid category, got nil")
		}

		// Verify the upload method was NOT called
		if mockAPI.UploadCalled {
			t.Error("UploadVideo should not be called with invalid category")
		}
	})

	// Test case 3: Upload error
	t.Run("Upload error", func(t *testing.T) {
		// Create a mock with an upload error
		mockAPI := NewMockYouTubeAPI()
		mockAPI.UploadVideoError = errors.New("network error")

		// Create test data
		metadata := VideoMetadata{
			Title:       "Test Video",
			Description: "This is a test video",
			Tags:        []string{"test", "demo"},
			Category:    "22", // People & Blogs
		}

		// Call the function being tested
		_, err := PublishVideo(mockAPI, metadata, "test.mp4", "thumb.jpg")

		// Assert error is returned
		if err == nil {
			t.Error("Expected an error for upload failure, got nil")
		}
	})
}
