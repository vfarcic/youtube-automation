package main

import (
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/workflow"
	"devopstoolkit/youtube-automation/pkg/testutil"
	"os"
	"path/filepath"
	"testing"
)

// TestCreateVideo tests the creation of a new video
func TestCreateVideo(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "video-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	testCategoryDir := filepath.Join(tempDir, "test-category")
	if err := os.MkdirAll(testCategoryDir, 0755); err != nil {
		t.Fatalf("Failed to create test category dir: %v", err)
	}

	// Create a test video using direct file operations
	mdFilePath := filepath.Join(testCategoryDir, "test-video.md")
	if err := os.WriteFile(mdFilePath, []byte("## Test content"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create YAML file directly
	yamlFilePath := filepath.Join(testCategoryDir, "test-video.yaml")
	y := storage.YAML{}

	video := storage.Video{
		Name:     "Test Video",
		Category: "test-category",
		Path:     yamlFilePath,
		Init:     storage.Tasks{Completed: 0, Total: 5},
		Work:     storage.Tasks{Completed: 0, Total: 11},
		Edit:     storage.Tasks{Completed: 0, Total: 6},
		Publish:  storage.Tasks{Completed: 0, Total: 5},
	}

	if err := y.WriteVideo(video, yamlFilePath); err != nil {
		t.Fatalf("Failed to write initial test video YAML: %v", err)
	}

	// Verify the files were created
	if _, err := os.Stat(mdFilePath); os.IsNotExist(err) {
		t.Errorf("Markdown file was not created at %s", mdFilePath)
	}

	if _, err := os.Stat(yamlFilePath); os.IsNotExist(err) {
		t.Errorf("YAML file was not created at %s", yamlFilePath)
	}

	// Read back the video to verify its contents
	readVideo, err := y.GetVideo(yamlFilePath)
	if err != nil {
		t.Fatalf("Failed to read back video YAML: %v", err)
	}
	if readVideo.Name != "Test Video" {
		t.Errorf("Expected video name to be 'Test Video', got '%s'", readVideo.Name)
	}
	if readVideo.Category != "test-category" {
		t.Errorf("Expected video category to be 'test-category', got '%s'", readVideo.Category)
	}
}

// TestVideoPhaseTransitions tests the phase transition functionality
func TestVideoPhaseTransitions(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "video-phase-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files directory structure
	testCategoryDir := filepath.Join(tempDir, "test-category")
	if err := os.MkdirAll(testCategoryDir, 0755); err != nil {
		t.Fatalf("Failed to create test category dir: %v", err)
	}

	// Create a test video file
	testVideoPath := filepath.Join(testCategoryDir, "test-video.yaml")

	// Define phase constants
	phaseConstants := testutil.VideoPhaseConstants{
		PhaseIdeas:            workflow.PhaseIdeas,
		PhaseStarted:          workflow.PhaseStarted,
		PhaseMaterialDone:     workflow.PhaseMaterialDone,
		PhaseEditRequested:    workflow.PhaseEditRequested,
		PhasePublishPending:   workflow.PhasePublishPending,
		PhasePublished:        workflow.PhasePublished,
		PhaseDelayed:          workflow.PhaseDelayed,
		PhaseSponsoredBlocked: workflow.PhaseSponsoredBlocked,
	}

	// Define function to test a video's phase
	testPhase := func(video storage.Video, expectedPhase int, message string) {
		// Write the video to the file
		y := storage.YAML{}
		if err := y.WriteVideo(video, testVideoPath); err != nil {
			t.Fatalf("Failed to write test video YAML for phase '%s': %v", message, err)
		}

		// Read the video directly without mocking
		video, err = y.GetVideo(testVideoPath)
		if err != nil {
			t.Fatalf("Failed to read back video YAML in testPhase for '%s': %v", message, err)
		}

		// Use the common helper function to determine the phase
		phase := testutil.DetermineVideoPhase(video, phaseConstants)

		if phase != expectedPhase {
			t.Errorf("Expected phase %s to be %d, got %d", message, expectedPhase, phase)
		}
	}

	// First phase: Idea (initial state)
	video := storage.Video{
		Name:     "Test Video",
		Category: "test-category",
		Path:     testVideoPath,
	}

	testPhase(video, workflow.PhaseIdeas, "initial")

	// Transition to Started phase by adding a date
	video.Date = "2023-12-31T12:00"
	testPhase(video, workflow.PhaseStarted, "after adding date")

	// Transition to Material Done phase
	video.Code = true
	video.Screen = true
	video.Head = true
	video.Diagrams = true
	testPhase(video, workflow.PhaseMaterialDone, "after completing material")

	// Transition to Edit Requested phase
	video.RequestEdit = true
	testPhase(video, workflow.PhaseEditRequested, "after requesting edit")

	// Transition to Publish Pending phase
	video.RequestEdit = false
	video.UploadVideo = "/path/to/video.mp4"
	video.Tweet = "This is a test tweet"
	testPhase(video, workflow.PhasePublishPending, "after adding upload and tweet")

	// Transition to Published phase
	video.Repo = "https://github.com/test/repo"
	testPhase(video, workflow.PhasePublished, "after adding repo")

	// Test Delayed phase
	video.Delayed = true
	testPhase(video, workflow.PhaseDelayed, "after setting delayed")

	// Test Sponsored Blocked phase
	video.Delayed = false
	video.Sponsorship.Blocked = "Some reason"
	testPhase(video, workflow.PhaseSponsoredBlocked, "after setting sponsorship blocked")
}
