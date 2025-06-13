package service

import (
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/storage"
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
	fsOps := &filesystem.Operations{}
	videoManager := video.NewManager(fsOps.GetFilePath)
	service := NewVideoService("index.yaml", fsOps, videoManager)

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

func TestVideoService_GetVideo_SanitizesNameFromFile(t *testing.T) {
	service, tempDir, cleanup := setupTestVideoService(t)
	defer cleanup()

	videoFileName := "my-video-file"
	videoDisplayName := "My Video Display Name"
	category := "test-category"

	// Create a video file with a different name in content
	videoContent := "name: \"" + videoDisplayName + "\""
	videoPath := filepath.Join(tempDir, "manuscript", category, videoFileName+".yaml")
	err := os.WriteFile(videoPath, []byte(videoContent), 0644)
	require.NoError(t, err)

	// Create index entry for the video
	index, err := service.yamlStorage.GetIndex()
	require.NoError(t, err)
	index = append(index, storage.VideoIndex{Name: videoFileName, Category: category})
	err = service.yamlStorage.WriteIndex(index)
	require.NoError(t, err)

	// Get the video
	video, err := service.GetVideo(videoFileName, category)
	require.NoError(t, err)

	// Assert that the name is sanitized for consistency with filenames
	// The YAML content "My Video Display Name" should be sanitized to "my-video-display-name"
	expectedSanitizedName := "my-video-display-name"
	assert.Equal(t, expectedSanitizedName, video.Name)
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

	// Update one video to be delayed (phase 5)
	delayedVideo, err := service.GetVideo("delayed-video", "test-category")
	require.NoError(t, err)
	delayedVideo.Delayed = true
	err = service.UpdateVideo(delayedVideo)
	require.NoError(t, err)

	tests := []struct {
		name        string
		phase       int
		expectedLen int
		expectError bool
	}{
		{
			name:        "Phase 5 (delayed videos)",
			phase:       5,
			expectedLen: 1,
			expectError: false,
		},
		{
			name:        "Phase 7 (ideas - normal new videos)",
			phase:       7,
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

	// Update one video to be delayed (phase 5)
	delayedVideo, err := service.GetVideo("delayed-video", "test-category")
	require.NoError(t, err)
	delayedVideo.Delayed = true
	err = service.UpdateVideo(delayedVideo)
	require.NoError(t, err)

	phases, err := service.GetVideoPhases()
	assert.NoError(t, err)
	assert.NotNil(t, phases)

	// Verify we have the expected phases structure (0-7 only)
	expectedPhases := []int{0, 1, 2, 3, 4, 5, 6, 7}
	for _, phase := range expectedPhases {
		count, exists := phases[phase]
		assert.True(t, exists, "Phase %d should exist", phase)
		assert.GreaterOrEqual(t, count, 0, "Phase %d count should be >= 0", phase)
	}

	// Verify we have videos in phases 5 (delayed) and 7 (ideas)
	assert.Equal(t, 1, phases[5], "Should have 1 video in phase 5 (delayed)")
	assert.Equal(t, 1, phases[7], "Should have 1 video in phase 7 (ideas)")
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

	_, err := service.CreateVideo("test-initial", "test-category")
	require.NoError(t, err)

	videoToUpdate, err := service.GetVideo("test-initial", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"projectName":              "Test Project",
		"projectURL":               "http://example.com",
		"publishDate":              "2024-01-01T10:00",
		"gistPath":                 "path/to/gist.md",
		"delayed":                  false,
		"sponsorshipAmount":        "100",
		"sponsorshipEmails":        "sponsor@example.com",
		"sponsorshipBlockedReason": "",
	}

	videoAfterUpdate, err := service.UpdateVideoPhase(&videoToUpdate, "initial-details", updateData)
	require.NoError(t, err)
	require.NotNil(t, videoAfterUpdate)

	assert.Equal(t, "Test Project", videoAfterUpdate.ProjectName)
	assert.Equal(t, "http://example.com", videoAfterUpdate.ProjectURL)
	assert.Equal(t, "2024-01-01T10:00", videoAfterUpdate.Date)
	assert.Equal(t, "path/to/gist.md", videoAfterUpdate.Gist)
	assert.False(t, videoAfterUpdate.Delayed)
	assert.Equal(t, "100", videoAfterUpdate.Sponsorship.Amount)
	assert.Equal(t, "sponsor@example.com", videoAfterUpdate.Sponsorship.Emails)
	assert.Equal(t, "", videoAfterUpdate.Sponsorship.Blocked)
}

func TestVideoService_UpdateVideoPhase_WorkProgress(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("test-work", "test-category")
	require.NoError(t, err)

	videoToUpdate, err := service.GetVideo("test-work", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"codeDone":            true,
		"talkingHeadDone":     true,
		"screenRecordingDone": true,
		"relatedVideos":       "video1,video2",
		"thumbnailsDone":      true,
		"diagramsDone":        false,
		"screenshotsDone":     true,
		"filesLocation":       "/path/to/files",
		"tagline":             "Amazing video tagline",
		"taglineIdeas":        "idea1,idea2,idea3",
		"otherLogosAssets":    "logo1.png,logo2.png",
	}

	videoAfterUpdate, err := service.UpdateVideoPhase(&videoToUpdate, "work-progress", updateData)
	require.NoError(t, err)
	require.NotNil(t, videoAfterUpdate)

	assert.True(t, videoAfterUpdate.Code)
	assert.True(t, videoAfterUpdate.Head)
	assert.True(t, videoAfterUpdate.Screen)
	assert.Equal(t, "video1,video2", videoAfterUpdate.RelatedVideos)
	assert.True(t, videoAfterUpdate.Thumbnails)
	assert.False(t, videoAfterUpdate.Diagrams)
	assert.True(t, videoAfterUpdate.Screenshots)
	assert.Equal(t, "/path/to/files", videoAfterUpdate.Location)
	assert.Equal(t, "Amazing video tagline", videoAfterUpdate.Tagline)
	assert.Equal(t, "idea1,idea2,idea3", videoAfterUpdate.TaglineIdeas)
	assert.Equal(t, "logo1.png,logo2.png", videoAfterUpdate.OtherLogos)
}

func TestVideoService_UpdateVideoPhase_Definition(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("test-define", "test-category")
	require.NoError(t, err)

	videoToUpdate, err := service.GetVideo("test-define", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"title":                      "New Title",
		"description":                "New Description",
		"highlight":                  "New Highlight",
		"tags":                       "new,tags",
		"descriptionTags":            "#new #description #tags",
		"tweetText":                  "New Tweet Text",
		"animationsScript":           "New Animations Script",
		"requestThumbnailGeneration": true,
		"gistPath":                   "new/gist/path.md",
	}

	videoAfterUpdate, err := service.UpdateVideoPhase(&videoToUpdate, "definition", updateData)
	require.NoError(t, err)
	require.NotNil(t, videoAfterUpdate)

	assert.Equal(t, "New Title", videoAfterUpdate.Title)
	assert.Equal(t, "New Description", videoAfterUpdate.Description)
	assert.Equal(t, "New Highlight", videoAfterUpdate.Highlight)
	assert.Equal(t, "new,tags", videoAfterUpdate.Tags)
	assert.Equal(t, "#new #description #tags", videoAfterUpdate.DescriptionTags)
	assert.Equal(t, "New Tweet Text", videoAfterUpdate.Tweet)
	assert.Equal(t, "New Animations Script", videoAfterUpdate.Animations)
	assert.True(t, videoAfterUpdate.RequestThumbnail)
	assert.Equal(t, "new/gist/path.md", videoAfterUpdate.Gist)

	fsOps := &filesystem.Operations{}
	localVideoManager := video.NewManager(fsOps.GetFilePath)
	defineCompleted, defineTotal := localVideoManager.CalculateDefinePhaseCompletion(*videoAfterUpdate)
	assert.Greater(t, defineTotal, 0)
	assert.Equal(t, 9, defineCompleted)
}

func TestVideoService_UpdateVideoPhase_PostProduction(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("test-postprod", "test-category")
	require.NoError(t, err)

	videoToUpdate, err := service.GetVideo("test-postprod", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"thumbnailPath": "/path/to/thumbnail.jpg",
		"members":       "member1,member2",
		"requestEdit":   true,
		"timecodes":     "00:00 - Intro, 01:30 - Main content",
		"movieDone":     true,
		"slidesDone":    false,
	}

	videoAfterUpdate, err := service.UpdateVideoPhase(&videoToUpdate, "post-production", updateData)
	require.NoError(t, err)
	require.NotNil(t, videoAfterUpdate)

	assert.Equal(t, "/path/to/thumbnail.jpg", videoAfterUpdate.Thumbnail)
	assert.Equal(t, "member1,member2", videoAfterUpdate.Members)
	assert.True(t, videoAfterUpdate.RequestEdit)
	assert.Equal(t, "00:00 - Intro, 01:30 - Main content", videoAfterUpdate.Timecodes)
	assert.True(t, videoAfterUpdate.Movie)
	assert.False(t, videoAfterUpdate.Slides)
}

func TestVideoService_UpdateVideoPhase_Publishing(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("test-publishing", "test-category")
	require.NoError(t, err)

	videoToUpdate, err := service.GetVideo("test-publishing", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"videoFilePath":   "/path/to/video.mp4",
		"uploadToYouTube": true,
		"createHugoPost":  true,
	}

	videoAfterUpdate, err := service.UpdateVideoPhase(&videoToUpdate, "publishing", updateData)
	require.NoError(t, err)
	require.NotNil(t, videoAfterUpdate)

	assert.Equal(t, "/path/to/video.mp4", videoAfterUpdate.UploadVideo)
	assert.Equal(t, "placeholder-youtube-id", videoAfterUpdate.VideoId)
	assert.Equal(t, "placeholder-hugo-path", videoAfterUpdate.HugoPath)
}

func TestVideoService_UpdateVideoPhase_PostPublish(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("test-postpublish", "test-category")
	require.NoError(t, err)

	videoToUpdate, err := service.GetVideo("test-postpublish", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"dotPosted":                 true,
		"blueSkyPostSent":           true,
		"linkedInPostSent":          true,
		"slackPostSent":             false,
		"youTubeHighlightCreated":   true,
		"youTubePinnedCommentAdded": true,
		"repliedToYouTubeComments":  false,
		"gdeAdvocuPostSent":         true,
		"codeRepositoryURL":         "https://github.com/example/repo",
		"notifiedSponsors":          false,
	}

	videoAfterUpdate, err := service.UpdateVideoPhase(&videoToUpdate, "post-publish", updateData)
	require.NoError(t, err)
	require.NotNil(t, videoAfterUpdate)

	assert.True(t, videoAfterUpdate.DOTPosted)
	assert.True(t, videoAfterUpdate.BlueSkyPosted)
	assert.True(t, videoAfterUpdate.LinkedInPosted)
	assert.False(t, videoAfterUpdate.SlackPosted)
	assert.True(t, videoAfterUpdate.YouTubeHighlight)
	assert.True(t, videoAfterUpdate.YouTubeComment)
	assert.False(t, videoAfterUpdate.YouTubeCommentReply)
	assert.True(t, videoAfterUpdate.GDE)
	assert.Equal(t, "https://github.com/example/repo", videoAfterUpdate.Repo)
	assert.False(t, videoAfterUpdate.NotifiedSponsors)
}

func TestVideoService_UpdateVideoPhase_InvalidPhase(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("test-invalid-phase", "test-category")
	require.NoError(t, err)

	videoToUpdate, err := service.GetVideo("test-invalid-phase", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"someField": "someValue",
	}

	_, err = service.UpdateVideoPhase(&videoToUpdate, "invalid-phase", updateData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown phase: invalid-phase")
}

func TestVideoService_UpdateVideoPhase_NonExistentVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	updateData := map[string]interface{}{
		"projectName": "Test Project",
	}

	// Test calling with a nil video pointer
	var nilVideo *storage.Video
	_, err := service.UpdateVideoPhase(nilVideo, "initial-details", updateData)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "video to update cannot be nil")
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

func TestVideoService_GetVideosByPhase_IdeasRandomization(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create multiple test videos in Ideas phase (phase 7)
	videoNames := []string{"video-01", "video-02", "video-03", "video-04", "video-05"}
	for _, name := range videoNames {
		_, err := service.CreateVideo(name, "test-category")
		require.NoError(t, err)
	}

	t.Run("Ideas phase should return videos in random order", func(t *testing.T) {
		// Get videos multiple times to check for randomization
		var orders [][]string
		iterations := 10

		for i := 0; i < iterations; i++ {
			videos, err := service.GetVideosByPhase(7) // Phase 7 is Ideas
			require.NoError(t, err)
			require.Len(t, videos, len(videoNames))

			// Extract video names in order
			var order []string
			for _, video := range videos {
				order = append(order, video.Name)
			}
			orders = append(orders, order)
		}

		// Check that not all orders are identical (randomization working)
		firstOrder := orders[0]
		foundDifferentOrder := false
		for i := 1; i < len(orders); i++ {
			if !slicesEqual(firstOrder, orders[i]) {
				foundDifferentOrder = true
				break
			}
		}

		assert.True(t, foundDifferentOrder, "Videos should be returned in different orders across multiple calls (randomization)")

		// Verify all expected videos are always present
		for _, order := range orders {
			assert.ElementsMatch(t, videoNames, order, "All videos should be present in each call")
		}
	})

	t.Run("Non-Ideas phases should maintain deterministic sorting", func(t *testing.T) {
		// Create a delayed video (phase 5) with dates for comparison
		_, err := service.CreateVideo("delayed-video-a", "test-category")
		require.NoError(t, err)
		_, err = service.CreateVideo("delayed-video-b", "test-category")
		require.NoError(t, err)

		// Update both to be delayed but with different dates
		videoA, err := service.GetVideo("delayed-video-a", "test-category")
		require.NoError(t, err)
		videoA.Delayed = true
		videoA.Date = "2024-01-01T10:00"
		err = service.UpdateVideo(videoA)
		require.NoError(t, err)

		videoB, err := service.GetVideo("delayed-video-b", "test-category")
		require.NoError(t, err)
		videoB.Delayed = true
		videoB.Date = "2024-01-02T10:00"
		err = service.UpdateVideo(videoB)
		require.NoError(t, err)

		// Get delayed videos multiple times
		var orders [][]string
		for i := 0; i < 5; i++ {
			videos, err := service.GetVideosByPhase(5) // Phase 5 is Delayed
			require.NoError(t, err)

			var order []string
			for _, video := range videos {
				order = append(order, video.Name)
			}
			orders = append(orders, order)
		}

		// All orders should be identical (deterministic sorting by date)
		expectedOrder := orders[0]
		for i := 1; i < len(orders); i++ {
			assert.Equal(t, expectedOrder, orders[i], "Non-Ideas phases should maintain consistent date-based sorting")
		}

		// Should be sorted by date (earliest first)
		assert.Equal(t, "delayed-video-a", orders[0][0], "Earlier dated video should come first")
		assert.Equal(t, "delayed-video-b", orders[0][1], "Later dated video should come second")
	})
}

// Helper function to compare string slices
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestVideoService_SanitizedNamesIntegration(t *testing.T) {
	service, tempDir, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Test creating a video with unsanitized name
	originalName := "Test Video With Spaces & Special!"
	category := "test-category"

	// Create video
	videoIndex, err := service.CreateVideo(originalName, category)
	require.NoError(t, err)

	// Verify the stored name is sanitized
	expectedSanitizedName := "test-video-with-spaces-&-special!"
	assert.Equal(t, expectedSanitizedName, videoIndex.Name, "CreateVideo should store sanitized name")

	// Verify we can retrieve the video using the sanitized name
	retrievedVideo, err := service.GetVideo(expectedSanitizedName, category)
	require.NoError(t, err)
	assert.Equal(t, expectedSanitizedName, retrievedVideo.Name, "GetVideo should return sanitized name")

	// Verify the actual file was created with sanitized filename
	expectedFilePath := filepath.Join(tempDir, "manuscript", category, expectedSanitizedName+".yaml")
	_, err = os.Stat(expectedFilePath)
	assert.NoError(t, err, "File should be created with sanitized filename")

	// Verify GetAllVideos returns sanitized names
	allVideos, err := service.GetAllVideos()
	require.NoError(t, err)
	require.Len(t, allVideos, 1)
	assert.Equal(t, expectedSanitizedName, allVideos[0].Name, "GetAllVideos should return sanitized names")
}
