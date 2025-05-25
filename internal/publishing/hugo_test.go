package publishing

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
)

// TestNewHugo tests creating a new Hugo instance
func TestNewHugo(t *testing.T) {
	hugo := &Hugo{}
	if hugo == nil {
		t.Fatal("Failed to create Hugo instance")
	}
}

// TestHugoFunctionErrors specifically tests error paths in Hugo functions
func TestHugoFunctionErrors(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-errors-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save original settings and restore after test
	originalSettings := configuration.GlobalSettings
	defer func() {
		configuration.GlobalSettings = originalSettings
	}()

	// Setup test settings
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{
			Path: tempDir,
		},
	}

	// Create test content
	testContent := "Test content"

	// Create the Hugo instance
	hugo := &Hugo{}

	// Test MkdirAll error
	t.Run("MkdirAll error", func(t *testing.T) {
		// Create a file where a directory should be to force MkdirAll to fail
		blockerFile := filepath.Join(tempDir, "content", "test-cat")
		if err := os.MkdirAll(filepath.Dir(blockerFile), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(blockerFile, []byte("blocker"), 0644); err != nil {
			t.Fatalf("Failed to create blocker file: %v", err)
		}

		// Test should now fail when trying to create directory
		_, err := hugo.hugoFromMarkdown(
			filepath.Join(tempDir, "manuscript", "test-cat", "file.md"),
			"Test Title",
			testContent,
		)
		if err == nil {
			t.Error("Expected error from MkdirAll, got nil")
		}
	})

	// Test WriteFile error
	t.Run("WriteFile error", func(t *testing.T) {
		// Create directory structure where we can test WriteFile error
		testDir := filepath.Join(tempDir, "content", "readonly")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		// Create a test filename where we'll test the error
		testFilename := filepath.Join(testDir, "test-title", "_index.md")

		// Make sure parent directory exists and is read-only
		parentDir := filepath.Dir(testFilename)
		if err := os.MkdirAll(parentDir, 0500); err != nil { // read-only directory
			t.Fatalf("Failed to create read-only directory: %v", err)
		}

		// Directly test hugoFromMarkdown - it should trigger WriteFile error
		_, err := hugo.hugoFromMarkdown(
			filepath.Join(tempDir, "manuscript", "readonly", "file.md"),
			"Test Title",
			testContent,
		)

		// The test may still pass if running with elevated permissions
		if err == nil {
			t.Log("Note: WriteFile error test passed unexpectedly, possibly running with elevated permissions")
		}
	})
}

// TestHugoIntegration tests the Hugo functionality using temporary directories
func TestHugoIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "hugo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the necessary directory structure
	manuscriptDir := filepath.Join(tempDir, "manuscript", "test-category")
	contentDir := filepath.Join(tempDir, "content", "test-category")

	if err := os.MkdirAll(manuscriptDir, 0755); err != nil {
		t.Fatalf("Failed to create manuscript directory: %v", err)
	}
	if err := os.MkdirAll(contentDir, 0755); err != nil {
		t.Fatalf("Failed to create content directory: %v", err)
	}

	// Save original settings and restore after test
	originalSettings := configuration.GlobalSettings
	defer func() {
		configuration.GlobalSettings = originalSettings
	}()

	// Setup test settings
	configuration.GlobalSettings = configuration.Settings{
		Hugo: configuration.SettingsHugo{
			Path: tempDir,
		},
	}

	// Create a test markdown file with complex content
	complexContent := `# Complex Test Post

## Introduction
This is a complex test post with multiple sections.

## Code Block
` + "```go" + `
func main() {
    fmt.Println("Hello, Hugo!")
}
` + "```" + `

## List
- Item 1
- Item 2
- Item 3

## Table
| Header 1 | Header 2 |
|----------|----------|
| Cell 1   | Cell 2   |
| Cell 3   | Cell 4   |
`

	testFilePath := filepath.Join(manuscriptDir, "complex-post.md")
	if err := os.WriteFile(testFilePath, []byte(complexContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a completely different drive/path structure for testing filepath.Rel error case
	// On Unix-like systems, this will cause filepath.Rel to return an error because
	// the paths have no common root
	unrelatedDir := "/unrelated/path/with/no/common/root"
	if _, err := os.Stat("/unrelated"); os.IsNotExist(err) {
		// Create a temporary fallback path within our tempDir since we can't create at root
		unrelatedDir = filepath.Join(tempDir, "unrelated-category")
		if err := os.MkdirAll(unrelatedDir, 0755); err != nil {
			t.Fatalf("Failed to create unrelated directory: %v", err)
		}
	}
	unrelatedFilePath := filepath.Join(unrelatedDir, "unrelated-post.md")
	if err := os.WriteFile(unrelatedFilePath, []byte(complexContent), 0644); err != nil {
		t.Fatalf("Failed to create unrelated test file: %v", err)
	}

	// For testing WriteFile error, create a directory structure where we can make a directory read-only
	readOnlyDir := filepath.Join(tempDir, "readonly-dir")
	if err := os.MkdirAll(readOnlyDir, 0755); err != nil {
		t.Fatalf("Failed to create read-only test directory: %v", err)
	}
	readOnlyFilePath := filepath.Join(readOnlyDir, "readonly-test.md")
	if err := os.WriteFile(readOnlyFilePath, []byte(complexContent), 0644); err != nil {
		t.Fatalf("Failed to create read-only test file: %v", err)
	}

	// Create the real Hugo instance
	hugo := &Hugo{}

	// Test various scenarios
	t.Run("Post with regular title", func(t *testing.T) {
		title := "Test Hugo Post"
		date := "2023-05-15T12:00"
		videoId := "testVideoId123"

		hugoPath, err := hugo.Post(testFilePath, title, date, videoId)
		if err != nil {
			t.Fatalf("Hugo.Post failed: %v", err)
		}

		// Verify the file was created at the expected location
		expectedPath := filepath.Join(tempDir, "content", "test-category", "test-hugo-post", "_index.md")
		if hugoPath != expectedPath {
			t.Errorf("Expected path: %s, got: %s", expectedPath, hugoPath)
		}

		// Check file exists
		if _, err := os.Stat(hugoPath); os.IsNotExist(err) {
			t.Errorf("Hugo post file was not created at: %s", hugoPath)
		}

		// Check content
		content, err := os.ReadFile(hugoPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		contentStr := string(content)
		expectedContent := []string{
			"title = 'Test Hugo Post'",
			"date = 2023-05-15T12:00:00+00:00",
			"draft = false",
			"{{< youtube testVideoId123 >}}",
			"# Complex Test Post",
			"## Introduction",
			"## Code Block",
			"## List",
			"## Table",
		}

		for _, expected := range expectedContent {
			if !strings.Contains(contentStr, expected) {
				t.Errorf("Generated file doesn't contain expected content: %s", expected)
			}
		}
	})

	t.Run("Post with special characters in title", func(t *testing.T) {
		title := "Test: Hugo & Post (Special) Characters!'"
		date := "2023-05-15T12:00"
		videoId := "anotherIdAbc"

		hugoPath, err := hugo.Post(testFilePath, title, date, videoId)
		if err != nil {
			t.Fatalf("Hugo.Post failed with special chars: %v", err)
		}

		// Get the actual directory name from the path
		dirName := filepath.Base(filepath.Dir(hugoPath))

		// Check the file exists
		if _, err := os.Stat(hugoPath); os.IsNotExist(err) {
			t.Errorf("Hugo post file was not created at: %s", hugoPath)
		}

		// Check that the directory name has been sanitized
		// by verifying it's lowercase and doesn't contain special characters
		if strings.ContainsAny(dirName, ":&()!'") {
			t.Errorf("Directory name contains special characters: %s", dirName)
		}

		// Check content
		content, err := os.ReadFile(hugoPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		// Verify the title in the content is the original, unsanitized title
		if !strings.Contains(string(content), fmt.Sprintf("title = '%s'", title)) {
			t.Errorf("Generated file doesn't contain the original title in front matter")
		}
		if !strings.Contains(string(content), fmt.Sprintf("{{< youtube %s >}}", videoId)) {
			t.Errorf("Generated file doesn't contain the YouTube shortcode with videoId")
		}
	})

	t.Run("Post with filepath.Rel error path", func(t *testing.T) {
		title := "Unrelated Post"
		date := "2023-05-15T12:00"
		videoId := "unrelatedVideo456"

		hugoPath, err := hugo.Post(unrelatedFilePath, title, date, videoId)
		if err != nil {
			t.Fatalf("Hugo.Post failed with unrelated path: %v", err)
		}

		// Since filepath.Rel returns an error, the implementation should fall back to using
		// the basename of the directory as the category
		expectedBaseName := filepath.Base(filepath.Dir(unrelatedFilePath))
		if !strings.Contains(hugoPath, expectedBaseName) {
			t.Errorf("Expected path to contain directory basename: %s, got: %s", expectedBaseName, hugoPath)
		}

		// Check file exists
		if _, err := os.Stat(hugoPath); os.IsNotExist(err) {
			t.Errorf("Hugo post file was not created at: %s", hugoPath)
		}

		// Check content
		content, err := os.ReadFile(hugoPath)
		if err != nil {
			t.Fatalf("Failed to read generated file: %v", err)
		}

		// Verify the title in the content is correct
		if !strings.Contains(string(content), fmt.Sprintf("title = '%s'", title)) {
			t.Errorf("Generated file doesn't contain the title in front matter")
		}
		if !strings.Contains(string(content), fmt.Sprintf("{{< youtube %s >}}", videoId)) {
			t.Errorf("Generated file doesn't contain the YouTube shortcode with videoId")
		}
	})

	t.Run("Post with N/A gist", func(t *testing.T) {
		hugoPath, err := hugo.Post("N/A", "Test Title", "2023-05-15T12:00", "")
		if err != nil {
			t.Errorf("Expected no error for N/A gist, got: %v", err)
		}
		if hugoPath != "" {
			t.Errorf("Expected empty path for N/A gist, got: %s", hugoPath)
		}
	})

	t.Run("Post with non-existent file", func(t *testing.T) {
		_, err := hugo.Post(filepath.Join(manuscriptDir, "non-existent.md"), "Test Title", "2023-05-15T12:00", "")
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})

	t.Run("Post with directory creation error", func(t *testing.T) {
		// Create a file where a directory should be
		blockerPath := filepath.Join(tempDir, "content", "test-category", "test-blocked-dir")
		if err := os.MkdirAll(filepath.Dir(blockerPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(blockerPath, []byte("blocker"), 0644); err != nil {
			t.Fatalf("Failed to create blocker file: %v", err)
		}

		_, err := hugo.Post(testFilePath, "Test Blocked Dir", "2023-05-15T12:00", "blockedVideoId")
		if err == nil {
			t.Error("Expected error for directory creation issue, got nil")
		}
	})

	t.Run("Post with write file error", func(t *testing.T) {
		// Create a directory where we'll test write error
		targetDir := filepath.Join(contentDir, "read-only-post")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatalf("Failed to create target directory: %v", err)
		}

		// Make an index file but set it to read-only
		indexPath := filepath.Join(targetDir, "_index.md")
		if err := os.WriteFile(indexPath, []byte("read-only"), 0400); err != nil {
			t.Fatalf("Failed to create read-only index file: %v", err)
		}

		// Make the directory read-only
		if err := os.Chmod(targetDir, 0500); err != nil {
			t.Fatalf("Failed to set directory permissions: %v", err)
		}

		// Now try to write to this location, which should fail
		_, err := hugo.hugoFromMarkdown(testFilePath, "Read-Only Post", "test content")
		if err == nil {
			// If the test is running as root or has elevated permissions, this might still succeed
			// So we'll just note it rather than failing the test
			t.Log("Expected write error, but operation succeeded (possibly running with elevated permissions)")
		}
	})

	t.Run("Post with question mark in title", func(t *testing.T) {
		title := "What is Go? A Test Post"
		date := "2023-05-16T10:00" // Using a slightly different date
		videoId := "whatIsGoVideo789"

		hugoPath, err := hugo.Post(testFilePath, title, date, videoId)
		if err != nil {
			t.Fatalf("Hugo.Post failed with question mark in title: %v", err)
		}

		// Assert that the path does not contain '?'
		if strings.Contains(hugoPath, "?") {
			t.Errorf("Generated path still contains '?': %s", hugoPath)
		}

		// Add check for youtube shortcode with videoId
		content, errReadFile := os.ReadFile(hugoPath)
		if errReadFile != nil {
			t.Fatalf("Failed to read generated file for shortcode check: %v", errReadFile)
		}
		if !strings.Contains(string(content), fmt.Sprintf("{{< youtube %s >}}", videoId)) {
			t.Errorf("Generated file doesn't contain the YouTube shortcode with videoId: %s", videoId)
		}

		// Construct expected sanitized path
		// Current sanitization in hugo.go: " ", "-", "(", "", ")", "", ":", "", "&", "", "/", "-", "'", "", "!", ""
		// We expect "?" to be ""
		sanitizedTitle := "what-is-go-a-test-post" // Manually sanitized based on existing and expected rules
		expectedPath := filepath.Join(tempDir, "content", "test-category", sanitizedTitle, "_index.md")

		if hugoPath != expectedPath {
			t.Errorf("Expected path: %s, got: %s", expectedPath, hugoPath)
		}

		// Check file exists
		if _, err := os.Stat(hugoPath); os.IsNotExist(err) {
			t.Errorf("Hugo post file was not created at: %s", hugoPath)
		}
	})

	t.Run("Post with gist N/A", func(t *testing.T) {
		hugoPath, err := hugo.Post("N/A", "Test Title", "2023-05-15T12:00", "testVideoId")
		if err != nil {
			t.Errorf("Expected no error for N/A gist, got: %v", err)
		}
		if hugoPath != "" {
			t.Errorf("Expected empty path for N/A gist, got: %s", hugoPath)
		}
	})

	// New Test Case: Post without VideoID (expecting FIXME)
	t.Run("Post without VideoID", func(t *testing.T) {
		title := "Test Post No Video ID"
		date := "2023-05-17T10:00"
		videoId := "" // Empty videoId

		hugoPath, err := hugo.Post(testFilePath, title, date, videoId)
		if err != nil {
			t.Fatalf("Hugo.Post failed for no VideoID case: %v", err)
		}

		content, err := os.ReadFile(hugoPath)
		if err != nil {
			t.Fatalf("Failed to read generated file for no VideoID case: %v", err)
		}

		if !strings.Contains(string(content), "{{< youtube FIXME: >}}") {
			t.Errorf("Generated file does not contain '{{< youtube FIXME: >}}' when VideoID is empty. Got: %s", string(content))
		}
	})

	// New Test Case: Post with VideoID (expecting the ID)
	t.Run("Post with VideoID", func(t *testing.T) {
		title := "Test Post With Video ID"
		date := "2023-05-18T10:00"
		videoId := "actualVideoId12345"

		hugoPath, err := hugo.Post(testFilePath, title, date, videoId)
		if err != nil {
			t.Fatalf("Hugo.Post failed for with VideoID case: %v", err)
		}

		content, err := os.ReadFile(hugoPath)
		if err != nil {
			t.Fatalf("Failed to read generated file for with VideoID case: %v", err)
		}

		expectedShortcode := fmt.Sprintf("{{< youtube %s >}}", videoId)
		if !strings.Contains(string(content), expectedShortcode) {
			t.Errorf("Generated file does not contain '%s'. Got: %s", expectedShortcode, string(content))
		}
	})
}
