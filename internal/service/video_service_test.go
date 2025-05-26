package service

import (
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestVideoService(t *testing.T) (*VideoService, string, func()) {
	// Create temporary directory
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)

	// Create manuscript directory structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)
	os.Mkdir("manuscript/category-02", 0755)

	// Create empty index.yaml file
	indexContent := "[]"
	os.WriteFile("index.yaml", []byte(indexContent), 0644)

	// Initialize service dependencies
	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	cleanup := func() {
		os.Chdir(originalDir)
	}

	return service, tempDir, cleanup
}

func TestVideoService_CreateVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	tests := []struct {
		name        string
		videoName   string
		category    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid video creation",
			videoName:   "test-video",
			category:    "test-category",
			expectError: false,
		},
		{
			name:        "Empty name",
			videoName:   "",
			category:    "test-category",
			expectError: true,
			errorMsg:    "name and category are required",
		},
		{
			name:        "Empty category",
			videoName:   "test-video",
			category:    "",
			expectError: true,
			errorMsg:    "name and category are required",
		},
		{
			name:        "Both empty",
			videoName:   "",
			category:    "",
			expectError: true,
			errorMsg:    "name and category are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vi, err := service.CreateVideo(tt.videoName, tt.category)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Empty(t, vi.Name)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.videoName, vi.Name)
				assert.Equal(t, tt.category, vi.Category)

				// Verify files were created
				yamlPath := filepath.Join("manuscript", tt.category, tt.videoName+".yaml")
				mdPath := filepath.Join("manuscript", tt.category, tt.videoName+".md")
				
				assert.FileExists(t, yamlPath)
				assert.FileExists(t, mdPath)

				// Verify YAML content has proper structure
				video, err := service.GetVideo(tt.videoName, tt.category)
				assert.NoError(t, err)
				assert.Equal(t, tt.videoName, video.Name)
				assert.Equal(t, tt.category, video.Category)
			}
		})
	}
}

func TestVideoService_GetVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video first
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	tests := []struct {
		name        string
		videoName   string
		category    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid video retrieval",
			videoName:   "test-video",
			category:    "test-category",
			expectError: false,
		},
		{
			name:        "Non-existent video",
			videoName:   "non-existent",
			category:    "test-category",
			expectError: true,
			errorMsg:    "failed to get video non-existent",
		},
		{
			name:        "Empty name",
			videoName:   "",
			category:    "test-category",
			expectError: true,
			errorMsg:    "name and category are required",
		},
		{
			name:        "Empty category",
			videoName:   "test-video",
			category:    "",
			expectError: true,
			errorMsg:    "name and category are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			video, err := service.GetVideo(tt.videoName, tt.category)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.videoName, video.Name)
				assert.Equal(t, tt.category, video.Category)
				assert.NotEmpty(t, video.Path)
			}
		})
	}
}

func TestVideoService_UpdateVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video first
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	// Get the video to update
	video, err := service.GetVideo("test-video", "test-category")
	require.NoError(t, err)

	// Update some fields
	video.Title = "Updated Title"
	video.Description = "Updated Description"
	video.Head = true

	err = service.UpdateVideo(video)
	assert.NoError(t, err)

	// Verify update persisted
	updatedVideo, err := service.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedVideo.Title)
	assert.Equal(t, "Updated Description", updatedVideo.Description)
	assert.True(t, updatedVideo.Head)

	// Test with empty path
	video.Path = ""
	err = service.UpdateVideo(video)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "video path is required")
}

func TestVideoService_DeleteVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create test videos
	_, err := service.CreateVideo("test-video-1", "test-category")
	require.NoError(t, err)
	_, err = service.CreateVideo("test-video-2", "test-category")
	require.NoError(t, err)

	tests := []struct {
		name        string
		videoName   string
		category    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid video deletion",
			videoName:   "test-video-1",
			category:    "test-category",
			expectError: false,
		},
		{
			name:        "Empty name",
			videoName:   "",
			category:    "test-category",
			expectError: true,
			errorMsg:    "name and category are required",
		},
		{
			name:        "Empty category",
			videoName:   "test-video-2",
			category:    "",
			expectError: true,
			errorMsg:    "name and category are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeleteVideo(tt.videoName, tt.category)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// Verify files were deleted
				yamlPath := filepath.Join("manuscript", tt.category, tt.videoName+".yaml")
				mdPath := filepath.Join("manuscript", tt.category, tt.videoName+".md")
				
				assert.NoFileExists(t, yamlPath)
				assert.NoFileExists(t, mdPath)

				// Verify video can't be retrieved anymore
				_, err := service.GetVideo(tt.videoName, tt.category)
				assert.Error(t, err)
			}
		})
	}
}

func TestVideoService_GetVideosByPhase(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create test videos with different characteristics
	_, err := service.CreateVideo("delayed-video", "test-category")
	require.NoError(t, err)
	
	_, err = service.CreateVideo("normal-video", "test-category")
	require.NoError(t, err)

	// Update one video to be delayed (phase 7)
	delayedVideo, err := service.GetVideo("delayed-video", "test-category")
	require.NoError(t, err)
	delayedVideo.Delayed = true
	err = service.UpdateVideo(delayedVideo)
	require.NoError(t, err)

	tests := []struct {
		name         string
		phase        int
		expectedLen  int
		expectError  bool
	}{
		{
			name:        "Phase 7 (delayed videos)",
			phase:       7,
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Phase 1 (normal new videos)",
			phase:       1,
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Phase 3 (non-existent)",
			phase:       3,
			expectedLen: 0,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			videos, err := service.GetVideosByPhase(tt.phase)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, videos, tt.expectedLen)
				
				for _, video := range videos {
					assert.NotEmpty(t, video.Name)
					assert.NotEmpty(t, video.Category)
					assert.NotEmpty(t, video.Path)
				}
			}
		})
	}
}

func TestVideoService_GetVideoPhases(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create videos in different phases
	_, err := service.CreateVideo("delayed-video", "test-category")
	require.NoError(t, err)
	
	_, err = service.CreateVideo("normal-video", "test-category")
	require.NoError(t, err)

	// Update one video to be delayed
	delayedVideo, err := service.GetVideo("delayed-video", "test-category")
	require.NoError(t, err)
	delayedVideo.Delayed = true
	err = service.UpdateVideo(delayedVideo)
	require.NoError(t, err)

	phases, err := service.GetVideoPhases()
	assert.NoError(t, err)
	assert.NotNil(t, phases)

	// Verify we have the expected phases structure
	expectedPhases := []int{1, 2, 3, 4, 5, 6, 7, 8}
	for _, phase := range expectedPhases {
		count, exists := phases[phase]
		assert.True(t, exists, "Phase %d should exist", phase)
		assert.GreaterOrEqual(t, count, 0, "Phase %d count should be >= 0", phase)
	}

	// Verify we have videos in phases 1 and 7
	assert.Equal(t, 1, phases[1], "Should have 1 video in phase 1")
	assert.Equal(t, 1, phases[7], "Should have 1 video in phase 7")
}

func TestVideoService_GetCategories(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	categories, err := service.GetCategories()
	assert.NoError(t, err)
	assert.NotNil(t, categories)

	// Should have at least the test categories we created
	assert.GreaterOrEqual(t, len(categories), 2)

	// Verify category structure
	found := false
	for _, cat := range categories {
		if cat.Name == "Test Category" {
			found = true
			assert.Contains(t, cat.Path, "test-category")
		}
		assert.NotEmpty(t, cat.Name)
		assert.NotEmpty(t, cat.Path)
	}
	assert.True(t, found, "Should find 'Test Category' in the list")
}

func TestVideoService_UpdateVideoPhase_InitialDetails(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"projectName":    "Test Project",
		"projectURL":     "https://example.com",
		"publishDate":    "2023-12-01T10:00",
		"gistPath":       "https://gist.github.com/example",
		"delayed":        false,
	}

	updatedVideo, err := service.UpdateVideoPhase("test-video", "test-category", "initial-details", updateData)
	assert.NoError(t, err)
	assert.Equal(t, "Test Project", updatedVideo.ProjectName)
	assert.Equal(t, "https://example.com", updatedVideo.ProjectURL)
	assert.Equal(t, "2023-12-01T10:00", updatedVideo.Date)
	assert.Equal(t, "https://gist.github.com/example", updatedVideo.Gist)
	assert.False(t, updatedVideo.Delayed)

	// Verify completion calculation
	assert.Greater(t, updatedVideo.Init.Total, 0)
	assert.GreaterOrEqual(t, updatedVideo.Init.Completed, 0)
}

func TestVideoService_UpdateVideoPhase_WorkProgress(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"codeDone":             true,
		"talkingHeadDone":      true,
		"screenRecordingDone":  true,
		"relatedVideos":        "video1,video2",
		"thumbnailsDone":       true,
		"diagramsDone":         false,
		"screenshotsDone":      true,
		"filesLocation":        "/path/to/files",
		"tagline":              "Amazing video tagline",
		"taglineIdeas":         "idea1,idea2,idea3",
		"otherLogosAssets":     "logo1.png,logo2.png",
	}

	updatedVideo, err := service.UpdateVideoPhase("test-video", "test-category", "work-progress", updateData)
	assert.NoError(t, err)
	assert.True(t, updatedVideo.Code)
	assert.True(t, updatedVideo.Head)
	assert.True(t, updatedVideo.Screen)
	assert.Equal(t, "video1,video2", updatedVideo.RelatedVideos)
	assert.True(t, updatedVideo.Thumbnails)
	assert.False(t, updatedVideo.Diagrams)
	assert.True(t, updatedVideo.Screenshots)
	assert.Equal(t, "/path/to/files", updatedVideo.Location)
	assert.Equal(t, "Amazing video tagline", updatedVideo.Tagline)
	assert.Equal(t, "idea1,idea2,idea3", updatedVideo.TaglineIdeas)
	assert.Equal(t, "logo1.png,logo2.png", updatedVideo.OtherLogos)

	// Verify completion calculation
	assert.Greater(t, updatedVideo.Work.Total, 0)
	assert.Greater(t, updatedVideo.Work.Completed, 0)
}

func TestVideoService_UpdateVideoPhase_Definition(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"title":                        "Amazing Video Title",
		"description":                  "This is an amazing video description",
		"highlight":                    "Key highlight of the video",
		"tags":                         "tag1,tag2,tag3",
		"descriptionTags":              "desc1,desc2",
		"tweetText":                    "Check out this amazing video!",
		"animationsScript":             "Animation script content",
		"requestThumbnailGeneration":   true,
	}

	updatedVideo, err := service.UpdateVideoPhase("test-video", "test-category", "definition", updateData)
	assert.NoError(t, err)
	assert.Equal(t, "Amazing Video Title", updatedVideo.Title)
	assert.Equal(t, "This is an amazing video description", updatedVideo.Description)
	assert.Equal(t, "Key highlight of the video", updatedVideo.Highlight)
	assert.Equal(t, "tag1,tag2,tag3", updatedVideo.Tags)
	assert.Equal(t, "desc1,desc2", updatedVideo.DescriptionTags)
	assert.Equal(t, "Check out this amazing video!", updatedVideo.Tweet)
	assert.Equal(t, "Animation script content", updatedVideo.Animations)
	assert.True(t, updatedVideo.RequestThumbnail)

	// Verify completion calculation
	assert.Greater(t, updatedVideo.Define.Total, 0)
	assert.Greater(t, updatedVideo.Define.Completed, 0)
}

func TestVideoService_UpdateVideoPhase_PostProduction(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"thumbnailPath": "/path/to/thumbnail.jpg",
		"members":       "member1,member2",
		"requestEdit":   true,
		"timecodes":     "00:00 - Intro, 01:30 - Main content",
		"movieDone":     true,
		"slidesDone":    false,
	}

	updatedVideo, err := service.UpdateVideoPhase("test-video", "test-category", "post-production", updateData)
	assert.NoError(t, err)
	assert.Equal(t, "/path/to/thumbnail.jpg", updatedVideo.Thumbnail)
	assert.Equal(t, "member1,member2", updatedVideo.Members)
	assert.True(t, updatedVideo.RequestEdit)
	assert.Equal(t, "00:00 - Intro, 01:30 - Main content", updatedVideo.Timecodes)
	assert.True(t, updatedVideo.Movie)
	assert.False(t, updatedVideo.Slides)

	// Verify completion calculation
	assert.Greater(t, updatedVideo.Edit.Total, 0)
	assert.GreaterOrEqual(t, updatedVideo.Edit.Completed, 0)
}

func TestVideoService_UpdateVideoPhase_Publishing(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"videoFilePath":   "/path/to/video.mp4",
		"uploadToYouTube": true,
		"createHugoPost":  true,
	}

	updatedVideo, err := service.UpdateVideoPhase("test-video", "test-category", "publishing", updateData)
	assert.NoError(t, err)
	assert.Equal(t, "/path/to/video.mp4", updatedVideo.UploadVideo)
	assert.Equal(t, "placeholder-youtube-id", updatedVideo.VideoId)
	assert.Equal(t, "placeholder-hugo-path", updatedVideo.HugoPath)

	// Verify completion calculation
	assert.Greater(t, updatedVideo.Publish.Total, 0)
	assert.Greater(t, updatedVideo.Publish.Completed, 0)
}

func TestVideoService_UpdateVideoPhase_PostPublish(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"blueSkyPostSent":              true,
		"linkedInPostSent":             true,
		"slackPostSent":                false,
		"youTubeHighlightCreated":      true,
		"youTubePinnedCommentAdded":    true,
		"repliedToYouTubeComments":     false,
		"gdeAdvocuPostSent":            true,
		"codeRepositoryURL":            "https://github.com/example/repo",
		"notifiedSponsors":             false,
	}

	updatedVideo, err := service.UpdateVideoPhase("test-video", "test-category", "post-publish", updateData)
	assert.NoError(t, err)
	assert.True(t, updatedVideo.BlueSkyPosted)
	assert.True(t, updatedVideo.LinkedInPosted)
	assert.False(t, updatedVideo.SlackPosted)
	assert.True(t, updatedVideo.YouTubeHighlight)
	assert.True(t, updatedVideo.YouTubeComment)
	assert.False(t, updatedVideo.YouTubeCommentReply)
	assert.True(t, updatedVideo.GDE)
	assert.Equal(t, "https://github.com/example/repo", updatedVideo.Repo)
	assert.False(t, updatedVideo.NotifiedSponsors)

	// Verify completion calculation
	assert.Greater(t, updatedVideo.PostPublish.Total, 0)
	assert.GreaterOrEqual(t, updatedVideo.PostPublish.Completed, 0)
}

func TestVideoService_UpdateVideoPhase_InvalidPhase(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"someField": "someValue",
	}

	_, err = service.UpdateVideoPhase("test-video", "test-category", "invalid-phase", updateData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown phase: invalid-phase")
}

func TestVideoService_UpdateVideoPhase_NonExistentVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	updateData := map[string]interface{}{
		"projectName": "Test Project",
	}

	_, err := service.UpdateVideoPhase("non-existent", "test-category", "initial-details", updateData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get video")
}

func TestVideoService_MoveVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create additional target directory
	os.Mkdir("manuscript/target-category", 0755)

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	tests := []struct {
		name        string
		videoName   string
		category    string
		targetDir   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid move",
			videoName:   "test-video",
			category:    "test-category",
			targetDir:   "manuscript/target-category",
			expectError: false,
		},
		{
			name:        "Empty name",
			videoName:   "",
			category:    "test-category",
			targetDir:   "manuscript/target-category",
			expectError: true,
			errorMsg:    "name, category, and target directory are required",
		},
		{
			name:        "Empty category",
			videoName:   "test-video",
			category:    "",
			targetDir:   "manuscript/target-category",
			expectError: true,
			errorMsg:    "name, category, and target directory are required",
		},
		{
			name:        "Empty target directory",
			videoName:   "test-video",
			category:    "test-category",
			targetDir:   "",
			expectError: true,
			errorMsg:    "name, category, and target directory are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.MoveVideo(tt.videoName, tt.category, tt.targetDir)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// For successful moves, verify the video is accessible from new location
				targetCategoryName := filepath.Base(tt.targetDir)
				video, err := service.GetVideo(tt.videoName, targetCategoryName)
				assert.NoError(t, err)
				assert.Equal(t, tt.videoName, video.Name)
				assert.Equal(t, targetCategoryName, video.Category)
			}
		})
	}
}