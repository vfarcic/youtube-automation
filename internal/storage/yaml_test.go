package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestVideo_JSONConsistency(t *testing.T) {
	t.Run("Video struct should serialize to camelCase JSON", func(t *testing.T) {
		video := Video{
			Name:        "test-video",
			ProjectName: "Test Project",
			ProjectURL:  "https://example.com",
			Sponsorship: Sponsorship{
				Amount:  "1000",
				Emails:  "sponsor@example.com",
				Blocked: "false",
			},
		}

		// Test serialization (GET response behavior)
		jsonData, err := json.Marshal(video)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		// Should be camelCase, not PascalCase
		assert.Equal(t, "Test Project", jsonMap["projectName"])
		assert.Equal(t, "https://example.com", jsonMap["projectURL"])

		// Sponsorship nested fields should also be camelCase
		sponsorship, ok := jsonMap["sponsorship"].(map[string]interface{})
		require.True(t, ok, "sponsorship should be a JSON object")
		assert.Equal(t, "1000", sponsorship["amount"])
		assert.Equal(t, "sponsor@example.com", sponsorship["emails"])
		assert.Equal(t, "false", sponsorship["blocked"])

		// These PascalCase fields should NOT exist
		assert.NotContains(t, jsonMap, "ProjectName")
		assert.NotContains(t, jsonMap, "ProjectURL")
	})

	t.Run("Video struct should deserialize from camelCase JSON", func(t *testing.T) {
		// Test deserialization (PUT request behavior)
		jsonData := `{
			"name": "test-video",
			"projectName": "Test Project",
			"projectURL": "https://example.com",
			"sponsorship": {
				"amount": "1000",
				"emails": "sponsor@example.com",
				"blocked": "false"
			}
		}`

		var video Video
		err := json.Unmarshal([]byte(jsonData), &video)
		require.NoError(t, err)

		assert.Equal(t, "test-video", video.Name)
		assert.Equal(t, "Test Project", video.ProjectName)
		assert.Equal(t, "https://example.com", video.ProjectURL)
		assert.Equal(t, "1000", video.Sponsorship.Amount)
		assert.Equal(t, "sponsor@example.com", video.Sponsorship.Emails)
		assert.Equal(t, "false", video.Sponsorship.Blocked)
	})

}

func TestGetUploadTitle(t *testing.T) {
	tests := []struct {
		name     string
		video    Video
		expected string
	}{
		{
			name: "Single title in Titles array",
			video: Video{
				Titles: []TitleVariant{
					{Index: 1, Text: "Primary Title", Share: 0},
				},
			},
			expected: "Primary Title",
		},
		{
			name: "Multiple titles, finds Index 1",
			video: Video{
				Titles: []TitleVariant{
					{Index: 2, Text: "Variant A", Share: 32.5},
					{Index: 1, Text: "Uploaded Title", Share: 45.2},
					{Index: 3, Text: "Variant B", Share: 22.3},
				},
			},
			expected: "Uploaded Title",
		},
		{
			name: "Empty Titles array returns empty string",
			video: Video{
				Titles: []TitleVariant{},
			},
			expected: "",
		},
		{
			name: "Index 1 missing returns empty string",
			video: Video{
				Titles: []TitleVariant{
					{Index: 2, Text: "Only Variant", Share: 100},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.video.GetUploadTitle()
			if got != tt.expected {
				t.Errorf("GetUploadTitle() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestVideoWithoutTitles(t *testing.T) {
	// Create a temp file without titles (GetUploadTitle should return empty string)
	tempDir := t.TempDir()
	videoPath := tempDir + "/test-video.yaml"

	yamlWithoutTitles := `name: Test Video
path: /path/to/video
category: testing
description: Test description
`
	err := os.WriteFile(videoPath, []byte(yamlWithoutTitles), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load the video
	y := NewYAML(tempDir + "/index.yaml")
	video, err := y.GetVideo(videoPath)
	if err != nil {
		t.Fatalf("GetVideo failed: %v", err)
	}

	// Verify Titles array is empty and GetUploadTitle returns empty string
	if len(video.Titles) != 0 {
		t.Errorf("Expected Titles array to be empty, got %d elements", len(video.Titles))
	}
	if video.GetUploadTitle() != "" {
		t.Errorf("Expected GetUploadTitle() to return empty string, got %q", video.GetUploadTitle())
	}
}

func TestTitlesArrayLoading(t *testing.T) {
	// Create a temp file with Titles array
	tempDir := t.TempDir()
	videoPath := tempDir + "/test-video.yaml"

	titlesYAML := `name: Test Video
path: /path/to/video
category: testing
titles:
  - index: 1
    text: New Format Title
    share: 0
  - index: 2
    text: Variant Title
    share: 0
`
	err := os.WriteFile(videoPath, []byte(titlesYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load the video
	y := NewYAML(tempDir + "/index.yaml")
	video, err := y.GetVideo(videoPath)
	if err != nil {
		t.Fatalf("GetVideo failed: %v", err)
	}

	// Verify Titles array is loaded correctly
	if len(video.Titles) != 2 {
		t.Errorf("Expected Titles array to have 2 elements, got %d", len(video.Titles))
	}
	if len(video.Titles) > 0 && video.Titles[0].Text != "New Format Title" {
		t.Errorf("Expected first title to be 'New Format Title', got %q", video.Titles[0].Text)
	}
}

func TestTitleVariantSerialization(t *testing.T) {
	t.Run("TitleVariant array serializes with share percentages", func(t *testing.T) {
		video := Video{
			Name: "test-video",
			Titles: []TitleVariant{
				{Index: 1, Text: "Primary Title", Share: 45.2},
				{Index: 2, Text: "Variant A", Share: 32.8},
				{Index: 3, Text: "Variant B", Share: 22.0},
			},
		}

		jsonData, err := json.Marshal(video)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		titles, ok := jsonMap["titles"].([]interface{})
		require.True(t, ok, "titles should be an array")
		assert.Len(t, titles, 3)

		title1 := titles[0].(map[string]interface{})
		assert.Equal(t, float64(1), title1["index"])
		assert.Equal(t, "Primary Title", title1["text"])
		assert.Equal(t, 45.2, title1["share"])
	})

	t.Run("TitleVariant array deserializes correctly", func(t *testing.T) {
		jsonData := `{
			"name": "test-video",
			"titles": [
				{"index": 1, "text": "Primary Title", "share": 45.2},
				{"index": 2, "text": "Variant A", "share": 32.8}
			]
		}`

		var video Video
		err := json.Unmarshal([]byte(jsonData), &video)
		require.NoError(t, err)

		assert.Len(t, video.Titles, 2)
		assert.Equal(t, 1, video.Titles[0].Index)
		assert.Equal(t, "Primary Title", video.Titles[0].Text)
		assert.Equal(t, 45.2, video.Titles[0].Share)
	})
}

// TestShortStruct tests the Short struct serialization and deserialization
func TestShortStruct(t *testing.T) {
	t.Run("Short struct serializes to JSON correctly", func(t *testing.T) {
		short := Short{
			ID:            "short1",
			Title:         "Quick Kubernetes Tip",
			Text:          "Here's a quick tip about Kubernetes...",
			ScheduledDate: "2025-01-15T14:30:00Z",
			YouTubeID:     "abc123xyz",
		}

		jsonData, err := json.Marshal(short)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		assert.Equal(t, "short1", jsonMap["id"])
		assert.Equal(t, "Quick Kubernetes Tip", jsonMap["title"])
		assert.Equal(t, "Here's a quick tip about Kubernetes...", jsonMap["text"])
		assert.Equal(t, "2025-01-15T14:30:00Z", jsonMap["scheduled_date"])
		assert.Equal(t, "abc123xyz", jsonMap["youtube_id"])
	})

	t.Run("Short struct deserializes from JSON correctly", func(t *testing.T) {
		jsonData := `{
			"id": "short2",
			"title": "DevOps Best Practice",
			"text": "One important DevOps practice is...",
			"scheduled_date": "2025-01-16T10:00:00Z",
			"youtube_id": "def456uvw"
		}`

		var short Short
		err := json.Unmarshal([]byte(jsonData), &short)
		require.NoError(t, err)

		assert.Equal(t, "short2", short.ID)
		assert.Equal(t, "DevOps Best Practice", short.Title)
		assert.Equal(t, "One important DevOps practice is...", short.Text)
		assert.Equal(t, "2025-01-16T10:00:00Z", short.ScheduledDate)
		assert.Equal(t, "def456uvw", short.YouTubeID)
	})

	t.Run("Short struct omits empty YouTubeID in JSON", func(t *testing.T) {
		short := Short{
			ID:            "short3",
			Title:         "Pending Short",
			Text:          "This short hasn't been uploaded yet",
			ScheduledDate: "2025-01-17T12:00:00Z",
			YouTubeID:     "", // Empty, should be omitted
		}

		jsonData, err := json.Marshal(short)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		// YouTubeID should not be present when empty (omitempty)
		_, exists := jsonMap["youtube_id"]
		assert.False(t, exists, "youtube_id should be omitted when empty")
	})

	t.Run("Short struct serializes to YAML correctly", func(t *testing.T) {
		short := Short{
			ID:            "short1",
			Title:         "Quick Tip",
			Text:          "Here's a quick tip...",
			ScheduledDate: "2025-01-15T14:30:00Z",
			YouTubeID:     "abc123",
		}

		yamlData, err := yaml.Marshal(short)
		require.NoError(t, err)

		var parsedShort Short
		err = yaml.Unmarshal(yamlData, &parsedShort)
		require.NoError(t, err)

		assert.Equal(t, short.ID, parsedShort.ID)
		assert.Equal(t, short.Title, parsedShort.Title)
		assert.Equal(t, short.Text, parsedShort.Text)
		assert.Equal(t, short.ScheduledDate, parsedShort.ScheduledDate)
		assert.Equal(t, short.YouTubeID, parsedShort.YouTubeID)
	})
}

// TestVideoWithShorts tests Video struct with Shorts field
func TestVideoWithShorts(t *testing.T) {
	t.Run("Video with Shorts serializes to JSON correctly", func(t *testing.T) {
		video := Video{
			Name:     "test-video",
			Category: "testing",
			Shorts: []Short{
				{
					ID:            "short1",
					Title:         "First Short",
					Text:          "First short content",
					ScheduledDate: "2025-01-15T14:30:00Z",
					YouTubeID:     "abc123",
				},
				{
					ID:            "short2",
					Title:         "Second Short",
					Text:          "Second short content",
					ScheduledDate: "2025-01-16T10:00:00Z",
					YouTubeID:     "",
				},
			},
		}

		jsonData, err := json.Marshal(video)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		shorts, ok := jsonMap["shorts"].([]interface{})
		require.True(t, ok, "shorts should be an array")
		assert.Len(t, shorts, 2)

		short1 := shorts[0].(map[string]interface{})
		assert.Equal(t, "short1", short1["id"])
		assert.Equal(t, "First Short", short1["title"])
	})

	t.Run("Video without Shorts omits field in JSON", func(t *testing.T) {
		video := Video{
			Name:     "test-video",
			Category: "testing",
			Shorts:   nil,
		}

		jsonData, err := json.Marshal(video)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		_, exists := jsonMap["shorts"]
		assert.False(t, exists, "shorts should be omitted when nil")
	})

	t.Run("Video with Shorts persists to YAML and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		videoPath := filepath.Join(tempDir, "test-video.yaml")

		originalVideo := Video{
			Name:     "Test Video with Shorts",
			Category: "testing",
			Path:     "/path/to/video",
			Shorts: []Short{
				{
					ID:            "short1",
					Title:         "Quick Tip",
					Text:          "Here's a quick tip about testing...",
					ScheduledDate: "2025-01-15T14:30:00Z",
					YouTubeID:     "xyz789",
				},
				{
					ID:            "short2",
					Title:         "Another Tip",
					Text:          "Another useful tip...",
					ScheduledDate: "2025-01-16T10:00:00Z",
					YouTubeID:     "", // Not yet uploaded
				},
			},
		}

		y := NewYAML(filepath.Join(tempDir, "index.yaml"))

		// Write the video
		err := y.WriteVideo(originalVideo, videoPath)
		require.NoError(t, err)

		// Read it back
		loadedVideo, err := y.GetVideo(videoPath)
		require.NoError(t, err)

		// Verify Shorts were persisted correctly
		assert.Len(t, loadedVideo.Shorts, 2)
		assert.Equal(t, "short1", loadedVideo.Shorts[0].ID)
		assert.Equal(t, "Quick Tip", loadedVideo.Shorts[0].Title)
		assert.Equal(t, "Here's a quick tip about testing...", loadedVideo.Shorts[0].Text)
		assert.Equal(t, "2025-01-15T14:30:00Z", loadedVideo.Shorts[0].ScheduledDate)
		assert.Equal(t, "xyz789", loadedVideo.Shorts[0].YouTubeID)

		assert.Equal(t, "short2", loadedVideo.Shorts[1].ID)
		assert.Equal(t, "Another Tip", loadedVideo.Shorts[1].Title)
		assert.Equal(t, "", loadedVideo.Shorts[1].YouTubeID)
	})

	t.Run("Video YAML without Shorts loads with empty Shorts slice", func(t *testing.T) {
		tempDir := t.TempDir()
		videoPath := filepath.Join(tempDir, "no-shorts-video.yaml")

		// Write YAML without shorts field
		yamlContent := `name: Video Without Shorts
category: testing
path: /path/to/video
`
		err := os.WriteFile(videoPath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		y := NewYAML(filepath.Join(tempDir, "index.yaml"))
		loadedVideo, err := y.GetVideo(videoPath)
		require.NoError(t, err)

		// Shorts should be nil/empty
		assert.Empty(t, loadedVideo.Shorts)
	})
}

// TestDubbingInfo tests DubbingInfo struct serialization
func TestDubbingInfo(t *testing.T) {
	t.Run("DubbingInfo with ThumbnailPath serializes to YAML correctly", func(t *testing.T) {
		info := DubbingInfo{
			DubbingID:       "dub123",
			DubbedVideoPath: "/path/to/dubbed-video-es.mp4",
			Title:           "Título en Español",
			Description:     "Descripción del video",
			Tags:            "tag1,tag2,tag3",
			UploadedVideoID: "yt123",
			DubbingStatus:   "dubbed",
			ThumbnailPath:   "/path/to/thumbnail-es.png",
		}

		yamlData, err := yaml.Marshal(info)
		require.NoError(t, err)

		var parsed DubbingInfo
		err = yaml.Unmarshal(yamlData, &parsed)
		require.NoError(t, err)

		assert.Equal(t, info.DubbingID, parsed.DubbingID)
		assert.Equal(t, info.DubbedVideoPath, parsed.DubbedVideoPath)
		assert.Equal(t, info.Title, parsed.Title)
		assert.Equal(t, info.Description, parsed.Description)
		assert.Equal(t, info.Tags, parsed.Tags)
		assert.Equal(t, info.UploadedVideoID, parsed.UploadedVideoID)
		assert.Equal(t, info.DubbingStatus, parsed.DubbingStatus)
		assert.Equal(t, info.ThumbnailPath, parsed.ThumbnailPath)
	})

	t.Run("DubbingInfo without ThumbnailPath omits field in YAML", func(t *testing.T) {
		info := DubbingInfo{
			DubbingID:     "dub123",
			DubbingStatus: "dubbed",
		}

		yamlData, err := yaml.Marshal(info)
		require.NoError(t, err)

		yamlStr := string(yamlData)
		assert.NotContains(t, yamlStr, "thumbnailPath", "thumbnailPath should be omitted when empty")
	})

	t.Run("DubbingInfo with ThumbnailPath serializes to JSON correctly", func(t *testing.T) {
		info := DubbingInfo{
			DubbingID:     "dub123",
			DubbingStatus: "dubbed",
			ThumbnailPath: "/path/to/thumbnail-es.png",
		}

		jsonData, err := json.Marshal(info)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		assert.Equal(t, "/path/to/thumbnail-es.png", jsonMap["thumbnailPath"])
	})
}

// TestVideoWithDubbing tests Video struct with Dubbing field including ThumbnailPath
func TestVideoWithDubbing(t *testing.T) {
	t.Run("Video with Dubbing persists to YAML and loads correctly", func(t *testing.T) {
		tempDir := t.TempDir()
		videoPath := filepath.Join(tempDir, "test-video.yaml")

		originalVideo := Video{
			Name:     "Test Video with Dubbing",
			Category: "testing",
			Path:     videoPath,
			Dubbing: map[string]DubbingInfo{
				"es": {
					DubbingID:       "dub-es-123",
					DubbedVideoPath: "/path/to/video-es.mp4",
					Title:           "Título en Español",
					Description:     "Descripción del video",
					UploadedVideoID: "yt-es-123",
					DubbingStatus:   "dubbed",
					ThumbnailPath:   "/path/to/thumbnail-es.png",
				},
				"pt": {
					DubbingID:       "dub-pt-456",
					DubbedVideoPath: "/path/to/video-pt.mp4",
					Title:           "Título em Português",
					DubbingStatus:   "dubbing",
					// ThumbnailPath intentionally empty
				},
			},
		}

		y := NewYAML(filepath.Join(tempDir, "index.yaml"))
		err := y.WriteVideo(originalVideo, videoPath)
		require.NoError(t, err)

		// Read it back
		loadedVideo, err := y.GetVideo(videoPath)
		require.NoError(t, err)

		// Verify Dubbing was persisted correctly
		require.Len(t, loadedVideo.Dubbing, 2)

		esInfo := loadedVideo.Dubbing["es"]
		assert.Equal(t, "dub-es-123", esInfo.DubbingID)
		assert.Equal(t, "/path/to/video-es.mp4", esInfo.DubbedVideoPath)
		assert.Equal(t, "Título en Español", esInfo.Title)
		assert.Equal(t, "dubbed", esInfo.DubbingStatus)
		assert.Equal(t, "/path/to/thumbnail-es.png", esInfo.ThumbnailPath)

		ptInfo := loadedVideo.Dubbing["pt"]
		assert.Equal(t, "dub-pt-456", ptInfo.DubbingID)
		assert.Equal(t, "dubbing", ptInfo.DubbingStatus)
		assert.Empty(t, ptInfo.ThumbnailPath, "ThumbnailPath should be empty for pt")
	})

	t.Run("Video YAML without Dubbing loads with nil Dubbing map", func(t *testing.T) {
		tempDir := t.TempDir()
		videoPath := filepath.Join(tempDir, "no-dubbing-video.yaml")

		yamlContent := `name: Video Without Dubbing
category: testing
path: /path/to/video
`
		err := os.WriteFile(videoPath, []byte(yamlContent), 0644)
		require.NoError(t, err)

		y := NewYAML(filepath.Join(tempDir, "index.yaml"))
		loadedVideo, err := y.GetVideo(videoPath)
		require.NoError(t, err)

		assert.Empty(t, loadedVideo.Dubbing)
	})
}
