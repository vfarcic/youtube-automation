package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	fsOps := filesystem.NewOperations()
	videoManager := video.NewManager(fsOps.GetFilePath, nil)
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
			vi, err := service.CreateVideo(tt.videoName, tt.category, "")

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

				// Verify Gist is auto-populated from Path
				expectedGist := strings.Replace(video.Path, ".yaml", ".md", 1)
				assert.Equal(t, expectedGist, video.Gist, "Gist should be auto-populated")
			}
		})
	}
}

func TestVideoService_GetVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video first
	_, err := service.CreateVideo("test-video", "test-category", "")
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
	_, err := service.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Get the video to update
	video, err := service.GetVideo("test-video", "test-category")
	require.NoError(t, err)

	// Update some fields
	video.Titles = []storage.TitleVariant{{Index: 1, Text: "Updated Title"}}
	video.Description = "Updated Description"
	video.Head = true

	err = service.UpdateVideo(video)
	assert.NoError(t, err)

	// Verify update persisted
	updatedVideo, err := service.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedVideo.GetUploadTitle())
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
	_, err := service.CreateVideo("test-video-1", "test-category", "")
	require.NoError(t, err)
	_, err = service.CreateVideo("test-video-2", "test-category", "")
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
	_, err := service.CreateVideo("delayed-video", "test-category", "")
	require.NoError(t, err)

	_, err = service.CreateVideo("normal-video", "test-category", "")
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
	_, err := service.CreateVideo("delayed-video", "test-category", "")
	require.NoError(t, err)

	_, err = service.CreateVideo("normal-video", "test-category", "")
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

func TestVideoService_SearchVideos(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create videos with different fields
	_, err := service.CreateVideo("kubernetes-basics", "test-category", "2025-01-01T10:00")
	require.NoError(t, err)
	_, err = service.CreateVideo("docker-intro", "test-category", "2025-02-01T10:00")
	require.NoError(t, err)
	_, err = service.CreateVideo("terraform-guide", "category-02", "2025-03-01T10:00")
	require.NoError(t, err)

	// Update one video with description
	v, err := service.GetVideo("docker-intro", "test-category")
	require.NoError(t, err)
	v.Description = "Learn about container orchestration"
	err = service.UpdateVideo(v)
	require.NoError(t, err)

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantFirst string
	}{
		{name: "match by name", query: "kubernetes", wantCount: 1, wantFirst: "kubernetes-basics"},
		{name: "match by category", query: "category-02", wantCount: 1, wantFirst: "terraform-guide"},
		{name: "match by description", query: "orchestration", wantCount: 1, wantFirst: "docker-intro"},
		{name: "case insensitive", query: "DOCKER", wantCount: 1, wantFirst: "docker-intro"},
		{name: "no match", query: "nonexistent", wantCount: 0},
		{name: "empty query", query: "", wantCount: 0},
		{name: "partial match", query: "terra", wantCount: 1, wantFirst: "terraform-guide"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := service.SearchVideos(tt.query)
			assert.NoError(t, err)
			assert.Len(t, results, tt.wantCount)
			if tt.wantCount > 0 && len(results) > 0 {
				assert.Equal(t, tt.wantFirst, results[0].Name)
			}
		})
	}
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

func TestVideoService_MoveVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create additional target directory
	os.Mkdir("manuscript/target-category", 0755)

	// Create a test video
	_, err := service.CreateVideo("test-video", "test-category", "")
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
		_, err := service.CreateVideo(name, "test-category", "")
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
		_, err := service.CreateVideo("delayed-video-a", "test-category", "")
		require.NoError(t, err)
		_, err = service.CreateVideo("delayed-video-b", "test-category", "")
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

func TestVideoService_OnMutate_CreateVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	var mutateMessages []string
	done := make(chan struct{}, 1)
	service.SetOnMutate(func(msg string) error {
		mutateMessages = append(mutateMessages, msg)
		done <- struct{}{}
		return nil
	})

	_, err := service.CreateVideo("callback-test", "test-category", "")
	require.NoError(t, err)
	<-done
	assert.Len(t, mutateMessages, 1)
	assert.Contains(t, mutateMessages[0], "create video")
}

func TestVideoService_OnMutate_UpdateVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("update-cb-test", "test-category", "")
	require.NoError(t, err)

	var mutateMessages []string
	done := make(chan struct{}, 1)
	service.SetOnMutate(func(msg string) error {
		mutateMessages = append(mutateMessages, msg)
		done <- struct{}{}
		return nil
	})

	video, err := service.GetVideo("update-cb-test", "test-category")
	require.NoError(t, err)
	video.Description = "updated"
	err = service.UpdateVideo(video)
	require.NoError(t, err)
	<-done
	assert.Len(t, mutateMessages, 1)
	assert.Contains(t, mutateMessages[0], "update video")
}

func TestVideoService_OnMutate_DeleteVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	_, err := service.CreateVideo("delete-cb-test", "test-category", "")
	require.NoError(t, err)

	var mutateMessages []string
	done := make(chan struct{}, 1)
	service.SetOnMutate(func(msg string) error {
		mutateMessages = append(mutateMessages, msg)
		done <- struct{}{}
		return nil
	})

	err = service.DeleteVideo("delete-cb-test", "test-category")
	require.NoError(t, err)
	<-done
	assert.Len(t, mutateMessages, 1)
	assert.Contains(t, mutateMessages[0], "delete video")
}

func TestVideoService_OnMutate_NotCalledOnFailure(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	var called bool
	service.SetOnMutate(func(msg string) error {
		called = true
		return nil
	})

	// Empty name should fail validation before mutation
	_, err := service.CreateVideo("", "test-category", "")
	assert.Error(t, err)
	assert.False(t, called, "onMutate should NOT be called when the mutation itself fails")
}

func TestVideoService_OnMutate_NilCallbackWorks(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// No callback set — should work without panic
	_, err := service.CreateVideo("no-cb-test", "test-category", "")
	assert.NoError(t, err)
}

func TestVideoService_OnMutate_ErrorIsLogged(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	done := make(chan struct{}, 1)
	service.SetOnMutate(func(msg string) error {
		defer func() { done <- struct{}{} }()
		return fmt.Errorf("push failed")
	})

	// Should succeed even though callback returns error (sync is async)
	_, err := service.CreateVideo("error-cb-test", "test-category", "")
	assert.NoError(t, err)
	<-done
	// Allow the goroutine in notifyMutation to store the error after the callback returns
	assert.Eventually(t, func() bool {
		return service.LastSyncError() != nil
	}, time.Second, 10*time.Millisecond, "expected LastSyncError to be set after async callback failure")
}

func TestVideoService_SanitizedNamesIntegration(t *testing.T) {
	service, tempDir, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Test creating a video with unsanitized name
	originalName := "Test Video With Spaces & Special!"
	category := "test-category"

	// Create video
	videoIndex, err := service.CreateVideo(originalName, category, "")
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

	// Verify GetVideosByPhase returns sanitized names
	videos, err := service.GetVideosByPhase(7) // Phase 7 is Ideas (new videos start here)
	require.NoError(t, err)
	require.Len(t, videos, 1)
	assert.Equal(t, expectedSanitizedName, videos[0].Name, "GetVideosByPhase should return sanitized names")
}

func TestVideoService_GetVideoManuscript(t *testing.T) {
	service, tempDir, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video first
	_, err := service.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Create a test manuscript file
	manuscriptDir := filepath.Join(tempDir, "manuscript", "test-category")
	err = os.MkdirAll(manuscriptDir, 0755)
	require.NoError(t, err)

	manuscriptPath := filepath.Join(manuscriptDir, "test-video.md")
	testManuscriptContent := "# Test Video\n\nThis is a test manuscript content for AI processing."
	err = os.WriteFile(manuscriptPath, []byte(testManuscriptContent), 0644)
	require.NoError(t, err)

	// Get the video and set its Gist field to point to the manuscript file
	video, err := service.GetVideo("test-video", "test-category")
	require.NoError(t, err)

	video.Gist = manuscriptPath
	err = service.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name            string
		videoName       string
		category        string
		expectError     bool
		errorMsg        string
		expectedContent string
	}{
		{
			name:            "Valid manuscript retrieval",
			videoName:       "test-video",
			category:        "test-category",
			expectError:     false,
			expectedContent: testManuscriptContent,
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
			content, err := service.GetVideoManuscript(tt.videoName, tt.category)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Empty(t, content)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContent, content)
			}
		})
	}
}

func TestVideoService_GetVideoManuscript_EmptyGistField(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video then clear its Gist field
	_, err := service.CreateVideo("test-video-no-gist", "test-category", "")
	require.NoError(t, err)

	video, err := service.GetVideo("test-video-no-gist", "test-category")
	require.NoError(t, err)
	video.Gist = ""
	err = service.UpdateVideo(video)
	require.NoError(t, err)

	// The Gist field is now empty
	content, err := service.GetVideoManuscript("test-video-no-gist", "test-category")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "gist field is empty")
	assert.Empty(t, content)
}

func TestVideoService_GetVideoManuscript_NonExistentFile(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create a test video with invalid Gist path
	_, err := service.CreateVideo("test-video-bad-gist", "test-category", "")
	require.NoError(t, err)

	// Get the video and set a non-existent Gist path
	video, err := service.GetVideo("test-video-bad-gist", "test-category")
	require.NoError(t, err)

	video.Gist = "/non/existent/path/to/manuscript.md"
	err = service.UpdateVideo(video)
	require.NoError(t, err)

	content, err := service.GetVideoManuscript("test-video-bad-gist", "test-category")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read manuscript file")
	assert.Empty(t, content)
}

func TestVideoService_ArchiveVideo(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create test videos with dates
	_, err := service.CreateVideo("video-2024", "test-category", "2024-06-15T10:00")
	require.NoError(t, err)
	_, err = service.CreateVideo("video-2025", "test-category", "2025-01-20T14:00")
	require.NoError(t, err)

	tests := []struct {
		name        string
		videoName   string
		category    string
		date        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid archive 2024",
			videoName:   "video-2024",
			category:    "test-category",
			date:        "2024-06-15T10:00",
			expectError: false,
		},
		{
			name:        "Valid archive 2025",
			videoName:   "video-2025",
			category:    "test-category",
			date:        "2025-01-20T14:00",
			expectError: false,
		},
		{
			name:        "Empty name",
			videoName:   "",
			category:    "test-category",
			date:        "2024-01-01T10:00",
			expectError: true,
			errorMsg:    "name and category are required",
		},
		{
			name:        "Empty category",
			videoName:   "some-video",
			category:    "",
			date:        "2024-01-01T10:00",
			expectError: true,
			errorMsg:    "name and category are required",
		},
		{
			name:        "Empty date",
			videoName:   "some-video",
			category:    "test-category",
			date:        "",
			expectError: true,
			errorMsg:    "video has no valid date",
		},
		{
			name:        "Invalid short date",
			videoName:   "some-video",
			category:    "test-category",
			date:        "202",
			expectError: true,
			errorMsg:    "video has no valid date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ArchiveVideo(tt.videoName, tt.category, tt.date)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)

				// Verify archive file was created
				year := tt.date[:4]
				archivePath := filepath.Join("index", year+".yaml")
				assert.FileExists(t, archivePath)

				// Verify video is in archive index
				archivedIndex, err := service.readArchiveIndex(archivePath)
				require.NoError(t, err)

				found := false
				for _, vi := range archivedIndex {
					if vi.Name == tt.videoName && vi.Category == tt.category {
						found = true
						break
					}
				}
				assert.True(t, found, "Video should be in archive index")

				// Verify removed from main index
				mainIndex, err := service.yamlStorage.GetIndex()
				require.NoError(t, err)

				for _, vi := range mainIndex {
					assert.False(t,
						vi.Name == tt.videoName && vi.Category == tt.category,
						"Video should be removed from main index")
				}
			}
		})
	}
}

func TestVideoService_ArchiveVideo_MultipleToSameYear(t *testing.T) {
	service, _, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Create multiple test videos for the same year
	_, err := service.CreateVideo("video-jan", "test-category", "2024-01-15T10:00")
	require.NoError(t, err)
	_, err = service.CreateVideo("video-jun", "test-category", "2024-06-20T14:00")
	require.NoError(t, err)
	_, err = service.CreateVideo("video-dec", "test-category", "2024-12-25T16:00")
	require.NoError(t, err)

	// Archive all three
	err = service.ArchiveVideo("video-jan", "test-category", "2024-01-15T10:00")
	require.NoError(t, err)
	err = service.ArchiveVideo("video-jun", "test-category", "2024-06-20T14:00")
	require.NoError(t, err)
	err = service.ArchiveVideo("video-dec", "test-category", "2024-12-25T16:00")
	require.NoError(t, err)

	// Verify all three are in the same archive file
	archivePath := filepath.Join("index", "2024.yaml")
	assert.FileExists(t, archivePath)

	archivedIndex, err := service.readArchiveIndex(archivePath)
	require.NoError(t, err)
	assert.Len(t, archivedIndex, 3)

	expectedNames := []string{"video-jan", "video-jun", "video-dec"}
	for _, expectedName := range expectedNames {
		found := false
		for _, vi := range archivedIndex {
			if vi.Name == expectedName {
				found = true
				break
			}
		}
		assert.True(t, found, "Video '%s' should be in archive", expectedName)
	}

	// Verify main index is empty
	mainIndex, err := service.yamlStorage.GetIndex()
	require.NoError(t, err)
	assert.Len(t, mainIndex, 0, "Main index should be empty after archiving all videos")
}

func TestVideoService_ExtractYearFromDate(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		expected string
	}{
		{"Full ISO date", "2024-06-15T10:00", "2024"},
		{"Different year", "2025-01-20T14:30", "2025"},
		{"Old year", "2020-12-31T23:59", "2020"},
		{"Empty string", "", ""},
		{"Too short", "202", ""},
		{"Exactly 4 chars", "2024", "2024"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractYearFromDate(tt.date)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVideoService_ArchiveVideo_LegacyUnsanitizedNames(t *testing.T) {
	service, tempDir, cleanup := setupTestVideoService(t)
	defer cleanup()

	// Simulate a legacy index entry with unsanitized name (capitals, spaces)
	legacyName := "AI for Devs"
	sanitizedName := "ai-for-devs"
	category := "test-category"
	date := "2024-03-15T10:00"

	// Create the video file with sanitized filename (as the system would expect)
	videoContent := "name: " + legacyName + "\ndate: " + date
	videoPath := filepath.Join(tempDir, "manuscript", category, sanitizedName+".yaml")
	err := os.WriteFile(videoPath, []byte(videoContent), 0644)
	require.NoError(t, err)

	// Create a legacy index entry with unsanitized name
	legacyIndex := []storage.VideoIndex{
		{Name: legacyName, Category: category},
	}
	err = service.yamlStorage.WriteIndex(legacyIndex)
	require.NoError(t, err)

	// Archive using sanitized name (as the menu handler would pass)
	err = service.ArchiveVideo(sanitizedName, category, date)
	require.NoError(t, err)

	// Verify the video was removed from main index
	mainIndex, err := service.yamlStorage.GetIndex()
	require.NoError(t, err)
	assert.Len(t, mainIndex, 0, "Main index should be empty after archiving legacy entry")

	// Verify the video is in the archive
	archivePath := filepath.Join("index", "2024.yaml")
	assert.FileExists(t, archivePath)

	archivedIndex, err := service.readArchiveIndex(archivePath)
	require.NoError(t, err)
	assert.Len(t, archivedIndex, 1)
	assert.Equal(t, sanitizedName, archivedIndex[0].Name)
}
