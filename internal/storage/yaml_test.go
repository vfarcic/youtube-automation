package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	if err := y.WriteVideo(testVideo, testPath); err != nil {
		t.Fatalf("Failed to write test video YAML in TestGetVideo: %v", err)
	}

	// Read the YAML file
	video, err := y.GetVideo(testPath)
	if err != nil {
		t.Fatalf("GetVideo returned an error: %v", err)
	}

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
	if err := y.WriteVideo(testVideo, testPath); err != nil {
		t.Fatalf("Failed to write test video YAML for TestWriteVideo: %v", err)
	}

	// Verify the file was created
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		t.Errorf("Expected file %s to exist, but it doesn't", testPath)
	}

	// Read the file back
	readVideo, err := y.GetVideo(testPath)
	if err != nil {
		t.Fatalf("GetVideo returned an error during read back: %v", err)
	}

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
	index, err := y.GetIndex()
	if err != nil {
		t.Fatalf("GetIndex returned an error: %v", err)
	}

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
	readIndex, err := y.GetIndex()
	if err != nil {
		t.Fatalf("GetIndex returned an error during read back: %v", err)
	}

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

	// Test that NewYAML creates a YAML struct with the correct IndexPath
	testIndexPath := "/test/path/index.json"
	newY := NewYAML(testIndexPath)
	if newY.IndexPath != testIndexPath {
		t.Errorf("Expected IndexPath to be '%s', got '%s'", testIndexPath, newY.IndexPath)
	}
}

func TestGetVideo_FileNotFound(t *testing.T) {
	y := YAML{}
	_, err := y.GetVideo("non_existent_path.yaml")
	if err == nil {
		t.Fatalf("Expected GetVideo to return an error for non-existent file, but got nil")
	}
	// Check if the error is an os.PathError, which is what os.ReadFile returns for non-existent files
	if !os.IsNotExist(err) {
		// It might be wrapped, so check unwrap
		type unwrap interface {
			Unwrap() error
		}
		if unwrapErr, ok := err.(unwrap); ok {
			if !os.IsNotExist(unwrapErr.Unwrap()) {
				t.Errorf("Expected GetVideo to return an os.IsNotExist error, got %T: %v", err, err)
			}
		} else {
			t.Errorf("Expected GetVideo to return an os.IsNotExist error, got %T: %v", err, err)
		}
	}
}

func TestGetVideo_InvalidYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invalid-yaml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidYAMLPath := filepath.Join(tempDir, "invalid.yaml")
	if err := os.WriteFile(invalidYAMLPath, []byte("name: Test Video\ncategory: testing\n  badlyIndentedKey: true"), 0644); err != nil {
		t.Fatalf("Failed to write invalid YAML file: %v", err)
	}

	y := YAML{}
	_, err = y.GetVideo(invalidYAMLPath)
	if err == nil {
		t.Fatalf("Expected GetVideo to return an error for invalid YAML, but got nil")
	}
	// We expect an error from yaml.Unmarshal, check for it.
	// The error message from our function is "failed to unmarshal video data from %s: %w"
	expectedErrorMsgPart := "failed to unmarshal video data"
	if !strings.Contains(err.Error(), expectedErrorMsgPart) {
		t.Errorf("Expected GetVideo error to contain '%s', got '%s'", expectedErrorMsgPart, err.Error())
	}
}

func TestGetIndex_FileNotFound(t *testing.T) {
	y := YAML{IndexPath: "non_existent_index.json"}
	_, err := y.GetIndex()
	if err == nil {
		t.Fatalf("Expected GetIndex to return an error for non-existent file, but got nil")
	}
	if !os.IsNotExist(err) {
		// It might be wrapped, so check unwrap
		type unwrap interface {
			Unwrap() error
		}
		if unwrapErr, ok := err.(unwrap); ok {
			if !os.IsNotExist(unwrapErr.Unwrap()) {
				t.Errorf("Expected GetIndex to return an os.IsNotExist error, got %T: %v", err, err)
			}
		} else {
			t.Errorf("Expected GetIndex to return an os.IsNotExist error, got %T: %v", err, err)
		}
	}
}

func TestGetIndex_InvalidYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invalid-index-yaml-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	invalidIndexYAMLPath := filepath.Join(tempDir, "invalid_index.yaml")
	if err := os.WriteFile(invalidIndexYAMLPath, []byte("[{\"name\": \"Test Video 1\", \"category\": \"testing\"}, {invalid_json]"), 0644); err != nil {
		t.Fatalf("Failed to write invalid index YAML file: %v", err)
	}

	y := YAML{IndexPath: invalidIndexYAMLPath}
	_, err = y.GetIndex()
	if err == nil {
		t.Fatalf("Expected GetIndex to return an error for invalid YAML, but got nil")
	}
	expectedErrorMsgPart := "failed to unmarshal video index"
	if !strings.Contains(err.Error(), expectedErrorMsgPart) {
		t.Errorf("Expected GetIndex error to contain '%s', got '%s'", expectedErrorMsgPart, err.Error())
	}
}
