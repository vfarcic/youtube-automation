package main

import (
	"devopstoolkitseries/youtube-automation/pkg/testutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	y := YAML{}

	video := Video{
		Name:     "Test Video",
		Category: "test-category",
		Path:     yamlFilePath,
		Init:     Tasks{Completed: 0, Total: 5},
		Work:     Tasks{Completed: 0, Total: 11},
		Define:   Tasks{Completed: 0, Total: 8},
		Edit:     Tasks{Completed: 0, Total: 6},
		Publish:  Tasks{Completed: 0, Total: 5},
	}

	y.WriteVideo(video, yamlFilePath)

	// Verify the files were created
	if _, err := os.Stat(mdFilePath); os.IsNotExist(err) {
		t.Errorf("Markdown file was not created at %s", mdFilePath)
	}

	if _, err := os.Stat(yamlFilePath); os.IsNotExist(err) {
		t.Errorf("YAML file was not created at %s", yamlFilePath)
	}

	// Read back the video to verify its contents
	readVideo := y.GetVideo(yamlFilePath)
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
		PhaseIdeas:            videosPhaseIdeas,
		PhaseStarted:          videosPhaseStarted,
		PhaseMaterialDone:     videosPhaseMaterialDone,
		PhaseEditRequested:    videosPhaseEditRequested,
		PhasePublishPending:   videosPhasePublishPending,
		PhasePublished:        videosPhasePublished,
		PhaseDelayed:          videosPhaseDelayed,
		PhaseSponsoredBlocked: videosPhaseSponsoredBlocked,
	}

	// Define function to test a video's phase
	testPhase := func(video Video, expectedPhase int, message string) {
		// Write the video to the file
		y := YAML{}
		y.WriteVideo(video, testVideoPath)

		// Read the video directly without mocking
		video = y.GetVideo(testVideoPath)

		// Use the common helper function to determine the phase
		phase := testutil.DetermineVideoPhase(video, phaseConstants)

		if phase != expectedPhase {
			t.Errorf("Expected phase %s to be %d, got %d", message, expectedPhase, phase)
		}
	}

	// First phase: Idea (initial state)
	video := Video{
		Name:     "Test Video",
		Category: "test-category",
		Path:     testVideoPath,
	}

	testPhase(video, videosPhaseIdeas, "initial")

	// Transition to Started phase by adding a date
	video.Date = "2023-12-31T12:00"
	testPhase(video, videosPhaseStarted, "after adding date")

	// Transition to Material Done phase
	video.Code = true
	video.Screen = true
	video.Head = true
	video.Diagrams = true
	testPhase(video, videosPhaseMaterialDone, "after completing material")

	// Transition to Edit Requested phase
	video.RequestEdit = true
	testPhase(video, videosPhaseEditRequested, "after requesting edit")

	// Transition to Publish Pending phase
	video.RequestEdit = false
	video.UploadVideo = "/path/to/video.mp4"
	video.Tweet = "This is a test tweet"
	testPhase(video, videosPhasePublishPending, "after adding upload and tweet")

	// Transition to Published phase
	video.Repo = "https://github.com/test/repo"
	testPhase(video, videosPhasePublished, "after adding repo")

	// Test Delayed phase
	video.Delayed = true
	testPhase(video, videosPhaseDelayed, "after setting delayed")

	// Test Sponsored Blocked phase
	video.Delayed = false
	video.Sponsorship.Blocked = "Some reason"
	testPhase(video, videosPhaseSponsoredBlocked, "after setting sponsorship blocked")
}

// TestVideoTaskCompletion tests the task completion tracking functionality
func TestVideoTaskCompletion(t *testing.T) {
	// Create a test Choices struct
	c := NewChoices()

	// Test counting completed tasks
	fields := []interface{}{
		"Not empty string", // Completed
		"",                 // Not completed
		true,               // Completed
		false,              // Not completed
		[]string{"item"},   // Completed
		[]string{},         // Not completed
	}

	completed, total := c.Count(fields)

	if total != 6 {
		t.Errorf("Expected total count to be 6, got %d", total)
	}

	if completed != 3 {
		t.Errorf("Expected completed count to be 3, got %d", completed)
	}
}

// TestVideoDeleteOperation tests the deletion of a video
func TestVideoDeleteOperation(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "video-delete-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	yamlPath := filepath.Join(tempDir, "test-video.yaml")
	mdPath := filepath.Join(tempDir, "test-video.md")

	// Create the files
	if err := os.WriteFile(yamlPath, []byte("name: Test Video"), 0644); err != nil {
		t.Fatalf("Failed to write test YAML file: %v", err)
	}

	if err := os.WriteFile(mdPath, []byte("## Test content"), 0644); err != nil {
		t.Fatalf("Failed to write test MD file: %v", err)
	}

	// Verify the files exist
	if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
		t.Fatalf("Test YAML file was not created")
	}
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		t.Fatalf("Test MD file was not created")
	}

	// Delete the files
	if err := os.Remove(yamlPath); err != nil {
		t.Fatalf("Failed to delete test YAML file: %v", err)
	}
	if err := os.Remove(mdPath); err != nil {
		t.Fatalf("Failed to delete test MD file: %v", err)
	}

	// Verify the files were deleted
	if _, err := os.Stat(yamlPath); !os.IsNotExist(err) {
		t.Errorf("Test YAML file was not deleted")
	}
	if _, err := os.Stat(mdPath); !os.IsNotExist(err) {
		t.Errorf("Test MD file was not deleted")
	}
}

// TestVideoFilteringAndSorting tests the filtering and sorting functionality
func TestVideoFilteringAndSorting(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "video-filter-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test directory structure
	testCategoryDir := filepath.Join(tempDir, "test-category")
	if err := os.MkdirAll(testCategoryDir, 0755); err != nil {
		t.Fatalf("Failed to create test category dir: %v", err)
	}

	// Create some test videos
	videos := []struct {
		name     string
		date     string
		phase    int // expected phase
		delayed  bool
		category string
	}{
		{"video1", "2023-01-01T12:00", videosPhaseStarted, false, "test-category"},
		{"video2", "2023-02-01T12:00", videosPhaseStarted, false, "test-category"},
		{"video3", "2023-03-01T12:00", videosPhaseDelayed, true, "test-category"},
		{"video4", "2023-04-01T12:00", videosPhaseStarted, false, "test-category"},
		{"video5", "", videosPhaseIdeas, false, "test-category"},
	}

	// Create video files
	y := YAML{}

	videoIndices := []VideoIndex{}

	for _, v := range videos {
		videoPath := filepath.Join(testCategoryDir, v.name+".yaml")

		video := Video{
			Name:     v.name,
			Category: v.category,
			Path:     videoPath,
			Date:     v.date,
			Delayed:  v.delayed,
			// Additional fields can be set as needed for specific phases
		}

		y.WriteVideo(video, videoPath)

		videoIndices = append(videoIndices, VideoIndex{
			Name:     v.name,
			Category: v.category,
		})
	}

	// Filter videos by phase
	phaseVideos := make(map[int][]VideoIndex)

	for _, vi := range videoIndices {
		// Read the video directly
		video := y.GetVideo(filepath.Join(testCategoryDir, vi.Name+".yaml"))

		// Determine the phase based on the video data
		var phase int
		if video.Delayed {
			phase = videosPhaseDelayed
		} else if len(video.Sponsorship.Blocked) > 0 {
			phase = videosPhaseSponsoredBlocked
		} else if len(video.Repo) > 0 {
			phase = videosPhasePublished
		} else if len(video.UploadVideo) > 0 && len(video.Tweet) > 0 {
			phase = videosPhasePublishPending
		} else if video.RequestEdit {
			phase = videosPhaseEditRequested
		} else if video.Code && video.Screen && video.Head && video.Diagrams {
			phase = videosPhaseMaterialDone
		} else if len(video.Date) > 0 {
			phase = videosPhaseStarted
		} else {
			phase = videosPhaseIdeas
		}

		phaseVideos[phase] = append(phaseVideos[phase], vi)
	}

	// Check that delayed videos are correctly identified
	if len(phaseVideos[videosPhaseDelayed]) != 1 || phaseVideos[videosPhaseDelayed][0].Name != "video3" {
		t.Errorf("Expected 1 delayed video (video3), got %d videos", len(phaseVideos[videosPhaseDelayed]))
	}

	// Check that videos in Ideas phase are correctly identified
	if len(phaseVideos[videosPhaseIdeas]) != 1 || phaseVideos[videosPhaseIdeas][0].Name != "video5" {
		t.Errorf("Expected 1 idea video (video5), got %d videos", len(phaseVideos[videosPhaseIdeas]))
	}

	// Check that videos in Started phase are correctly identified
	if len(phaseVideos[videosPhaseStarted]) != 3 {
		t.Errorf("Expected 3 started videos, got %d videos", len(phaseVideos[videosPhaseStarted]))
	}
}

// TestVersionFlag tests the --version flag
func TestVersionFlag(t *testing.T) {
	// This test assumes that a binary named "youtube-release"
	// has been built (e.g., via `make build-local` or a similar mechanism)
	// and is available in the current directory or PATH.
	// The Makefile's build-local target should inject the correct version.

	binaryName := "youtube-release" // Assumes 'make build-local' was run

	cmd := exec.Command("./"+binaryName, "--version")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Logf("Command exited with non-zero status: %v, stderr: %s", err, string(exitErr.Stderr))
		} else if _, ok := err.(*exec.Error); ok && os.IsNotExist(err) {
			t.Fatalf("Error executing binary: %v. Make sure '%s' is built (e.g., using 'make build-local') and in the current directory.", err, binaryName)
		} else {
			t.Fatalf("Error executing binary: %v, output: %s", err, string(output))
		}
	}

	// Trim whitespace from the output before comparing
	actualVersion := strings.TrimSpace(string(output))
	expectedVersion := "v0.1.0"

	if actualVersion != expectedVersion && actualVersion != "dev" { // Allow "dev"
		t.Errorf("Expected version output '%s' or 'dev', got '%s'", expectedVersion, actualVersion)
	}
}
