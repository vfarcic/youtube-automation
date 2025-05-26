package data

import (
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVideoService_CreateVideo(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create manuscript directory
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)

	// Create index.yaml file
	indexContent := "[]"
	os.WriteFile("index.yaml", []byte(indexContent), 0644)

	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	tests := []struct {
		name        string
		videoName   string
		category    string
		expectError bool
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
		},
		{
			name:        "Empty category",
			videoName:   "test-video",
			category:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			video, err := service.CreateVideo(tt.videoName, tt.category)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.videoName, video.Name)
			assert.Equal(t, tt.category, video.Category)
			
			// Check if files were created
			mdPath := filepath.Join("manuscript", tt.category, tt.videoName+".md")
			assert.FileExists(t, mdPath)
		})
	}
}

func TestVideoService_GetVideosByPhase(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create test structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)

	// Create index.yaml with test data
	indexContent := `- name: test-video-1
  category: test-category
- name: test-video-2
  category: test-category`
	os.WriteFile("index.yaml", []byte(indexContent), 0644)

	// Create video YAML files
	videoContent1 := `date: "2023-01-01T12:00"
delayed: false
code: true
screen: true
head: true
diagrams: true`
	os.WriteFile("manuscript/test-category/test-video-1.yaml", []byte(videoContent1), 0644)

	videoContent2 := `date: ""
delayed: false`
	os.WriteFile("manuscript/test-category/test-video-2.yaml", []byte(videoContent2), 0644)

	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	// Test getting videos in material done phase
	videos, err := service.GetVideosByPhase(workflow.PhaseMaterialDone)
	require.NoError(t, err)
	
	// Should have one video in material done phase (test-video-1)
	assert.Len(t, videos, 1)
	assert.Equal(t, "test-video-1", videos[0].Name)
}

func TestVideoService_GetVideoPhases(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create test structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)

	// Create index.yaml with test data
	indexContent := `- name: test-video-1
  category: test-category
- name: test-video-2
  category: test-category`
	os.WriteFile("index.yaml", []byte(indexContent), 0644)

	// Create video YAML files with different phases
	videoContent1 := `date: "2023-01-01T12:00"
delayed: false
code: true
screen: true
head: true
diagrams: true`
	os.WriteFile("manuscript/test-category/test-video-1.yaml", []byte(videoContent1), 0644)

	videoContent2 := `date: ""
delayed: false`
	os.WriteFile("manuscript/test-category/test-video-2.yaml", []byte(videoContent2), 0644)

	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	phases, err := service.GetVideoPhases()
	require.NoError(t, err)
	
	// Should have videos in different phases
	assert.Equal(t, 1, phases[workflow.PhaseMaterialDone]) // test-video-1
	assert.Equal(t, 1, phases[workflow.PhaseIdeas])       // test-video-2
}

func TestVideoService_GetVideo(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create test structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)

	// Create video YAML file
	videoContent := `title: "Test Video Title"
description: "Test video description"
date: "2023-01-01T12:00"`
	os.WriteFile("manuscript/test-category/test-video.yaml", []byte(videoContent), 0644)

	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	tests := []struct {
		name        string
		videoName   string
		category    string
		expectError bool
	}{
		{
			name:        "Valid video retrieval",
			videoName:   "test-video",
			category:    "test-category",
			expectError: false,
		},
		{
			name:        "Nonexistent video",
			videoName:   "nonexistent",
			category:    "test-category",
			expectError: true,
		},
		{
			name:        "Empty name",
			videoName:   "",
			category:    "test-category",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			video, err := service.GetVideo(tt.videoName, tt.category)
			
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			require.NoError(t, err)
			assert.Equal(t, tt.videoName, video.Name)
			assert.Equal(t, tt.category, video.Category)
			assert.Equal(t, "Test Video Title", video.Title)
		})
	}
}

func TestVideoService_UpdateVideo(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create test structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)

	// Create video YAML file
	videoContent := `title: "Original Title"
description: "Original description"`
	videoPath := "manuscript/test-category/test-video.yaml"
	os.WriteFile(videoPath, []byte(videoContent), 0644)

	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	// Update video
	video := storage.Video{
		Name:        "test-video",
		Category:    "test-category",
		Path:        videoPath,
		Title:       "Updated Title",
		Description: "Updated description",
	}

	err := service.UpdateVideo(video)
	require.NoError(t, err)

	// Verify update
	updatedVideo, err := service.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", updatedVideo.Title)
	assert.Equal(t, "Updated description", updatedVideo.Description)
}

func TestVideoService_DeleteVideo(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create test structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)

	// Create index.yaml with test data
	indexContent := `- name: test-video
  category: test-category
- name: other-video
  category: test-category`
	os.WriteFile("index.yaml", []byte(indexContent), 0644)

	// Create video files
	os.WriteFile("manuscript/test-category/test-video.yaml", []byte("title: Test"), 0644)
	os.WriteFile("manuscript/test-category/test-video.md", []byte("# Test"), 0644)
	os.WriteFile("manuscript/test-category/other-video.yaml", []byte("title: Other"), 0644)
	os.WriteFile("manuscript/test-category/other-video.md", []byte("# Other"), 0644)

	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	// Delete video
	err := service.DeleteVideo("test-video", "test-category")
	require.NoError(t, err)

	// Verify files are deleted
	assert.NoFileExists(t, "manuscript/test-category/test-video.yaml")
	assert.NoFileExists(t, "manuscript/test-category/test-video.md")
	
	// Verify other video still exists
	assert.FileExists(t, "manuscript/test-category/other-video.yaml")
	assert.FileExists(t, "manuscript/test-category/other-video.md")

	// Verify index is updated
	yaml := storage.NewYAML("index.yaml")
	index, err := yaml.GetIndex()
	require.NoError(t, err)
	assert.Len(t, index, 1)
	assert.Equal(t, "other-video", index[0].Name)
}

func TestVideoService_GetCategories(t *testing.T) {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(originalDir)

	// Create test directory structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/category-one", 0755)
	os.Mkdir("manuscript/category-two", 0755)
	os.WriteFile("manuscript/somefile.txt", []byte("not a directory"), 0644)

	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	service := NewVideoService("index.yaml", filesystem, videoManager)

	categories, err := service.GetCategories()
	require.NoError(t, err)
	
	assert.Len(t, categories, 2)
	
	// Categories should be sorted by name
	assert.Equal(t, "Category One", categories[0].Name)
	assert.Equal(t, "Category Two", categories[1].Name)
	assert.Contains(t, categories[0].Path, "category-one")
	assert.Contains(t, categories[1].Path, "category-two")
}