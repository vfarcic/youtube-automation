package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLParsing(t *testing.T) {
	// Simple direct test of yaml library functionality
	yamlContent := []byte("name: Test Video\ncategory: testing\npath: /path/to/video.yaml\n")
	var video Video
	err := yaml.Unmarshal(yamlContent, &video)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Print the video struct to debug
	fmt.Printf("Parsed Video: %+v\n", video)

	if video.Name != "Test Video" {
		t.Errorf("Expected Name to be 'Test Video', got '%s'", video.Name)
	}
	if video.Category != "testing" {
		t.Errorf("Expected Category to be 'testing', got '%s'", video.Category)
	}

	// Try with struct literals to verify the Video struct is working
	directVideo := Video{
		Name:     "Test Video",
		Category: "testing",
	}

	if directVideo.Name != "Test Video" {
		t.Errorf("Direct assignment test failed. Expected Name to be 'Test Video', got '%s'", directVideo.Name)
	}
}

// TestExportedFieldParsing tests if the issue might be with lowercase vs uppercase field names
func TestExportedFieldParsing(t *testing.T) {
	// Test structure without explicit yaml tags - relying on yaml library's auto-conversion
	type TestVideo struct {
		Name     string
		Category string
		Path     string
	}

	yamlContent := []byte("name: Test Video\ncategory: testing\npath: /path/to/video.yaml\n")
	var video TestVideo
	err := yaml.Unmarshal(yamlContent, &video)
	if err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	fmt.Printf("Parsed TestVideo: %+v\n", video)

	if video.Name != "Test Video" {
		t.Errorf("Expected Name to be 'Test Video', got '%s'", video.Name)
	}
}

// TestGetVideo tests the GetVideo functionality
func TestGetVideo(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "yaml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test YAML file
	testPath := filepath.Join(tempDir, "test-video.yaml")
	testVideo := Video{
		Name:     "Test Video",
		Category: "testing",
		Path:     "/path/to/video.yaml",
	}

	// Write the YAML file
	y := YAML{}
	y.WriteVideo(testVideo, testPath)

	// Read the YAML file
	video := y.GetVideo(testPath)

	// Verify the video was read correctly
	if video.Name != "Test Video" {
		t.Errorf("Expected video name to be 'Test Video', got '%s'", video.Name)
	}
	if video.Category != "testing" {
		t.Errorf("Expected video category to be 'testing', got '%s'", video.Category)
	}
	if video.Path != "/path/to/video.yaml" {
		t.Errorf("Expected video path to be '/path/to/video.yaml', got '%s'", video.Path)
	}
}

// TestWriteVideo tests the WriteVideo functionality
func TestWriteVideo(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "yaml-write-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test video
	testPath := filepath.Join(tempDir, "test-write-video.yaml")
	testVideo := Video{
		Name:     "Test Write Video",
		Category: "testing",
		Path:     "/path/to/written/video.yaml",
	}

	// Write the video to YAML
	y := YAML{}
	y.WriteVideo(testVideo, testPath)

	// Verify the file was created
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it doesn't", testPath)
	}

	// Read the file back
	readVideo := y.GetVideo(testPath)

	// Verify the contents
	if readVideo.Name != "Test Write Video" {
		t.Errorf("Expected video name to be 'Test Write Video', got '%s'", readVideo.Name)
	}
	if readVideo.Category != "testing" {
		t.Errorf("Expected video category to be 'testing', got '%s'", readVideo.Category)
	}
	if readVideo.Path != "/path/to/written/video.yaml" {
		t.Errorf("Expected video path to be '/path/to/written/video.yaml', got '%s'", readVideo.Path)
	}
}

// TestGetIndex tests the GetIndex functionality
func TestGetIndex(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "yaml-index-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test index file
	testPath := filepath.Join(tempDir, "index.json")

	// Create a simple index file
	indexContent := `[
		{"name": "Test Video 1", "category": "testing"},
		{"name": "Test Video 2", "category": "testing"}
	]`
	err = os.WriteFile(testPath, []byte(indexContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test index file: %v", err)
	}

	// Read the index
	y := YAML{
		IndexPath: testPath,
	}
	index := y.GetIndex()

	// Verify the index was read correctly
	if len(index) != 2 {
		t.Errorf("Expected index to have 2 entries, got %d", len(index))
	}
	if index[0].Name != "Test Video 1" {
		t.Errorf("Expected first video name to be 'Test Video 1', got '%s'", index[0].Name)
	}
	if index[1].Name != "Test Video 2" {
		t.Errorf("Expected second video name to be 'Test Video 2', got '%s'", index[1].Name)
	}
}

// TestWriteIndex tests the WriteIndex functionality
func TestWriteIndex(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "yaml-write-index-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test index
	testPath := filepath.Join(tempDir, "write-index.json")
	testIndex := []VideoIndex{
		{Name: "Test Write Video 1", Category: "testing"},
		{Name: "Test Write Video 2", Category: "testing"},
	}

	// Write the index
	y := YAML{
		IndexPath: testPath,
	}
	y.WriteIndex(testIndex)

	// Verify the file was created
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it doesn't", testPath)
	}

	// Read the file back
	readIndex := y.GetIndex()

	// Verify the contents
	if len(readIndex) != 2 {
		t.Errorf("Expected index to have 2 entries, got %d", len(readIndex))
	}
	if readIndex[0].Name != "Test Write Video 1" {
		t.Errorf("Expected first video name to be 'Test Write Video 1', got '%s'", readIndex[0].Name)
	}
	if readIndex[1].Name != "Test Write Video 2" {
		t.Errorf("Expected second video name to be 'Test Write Video 2', got '%s'", readIndex[1].Name)
	}
}

// TestNewYAML tests the NewYAML functionality
func TestNewYAML(t *testing.T) {
	// Create a YAML instance
	indexPath := "test-index.yaml"
	y := NewYAML(indexPath)

	// Verify it's not nil
	if y == nil {
		t.Errorf("Expected NewYAML to return a non-nil instance")
	}

	// Verify the index path is set correctly
	if y.IndexPath != indexPath {
		t.Errorf("Expected IndexPath to be '%s', got '%s'", indexPath, y.IndexPath)
	}
}
