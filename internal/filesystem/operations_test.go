package filesystem

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOperations(t *testing.T) {
	ops := NewOperations()
	assert.NotNil(t, ops, "NewOperations should return a non-nil Operations struct")
}

func TestGetDirPath(t *testing.T) {
	ops := NewOperations()

	tests := []struct {
		name     string
		category string
		expected string
	}{
		{"lowercase no space", "series", "manuscript/series"},
		{"lowercase with space", "my series", "manuscript/my-series"},
		{"mixed case with space", "My Awesome Series", "manuscript/my-awesome-series"},
		{"empty category", "", "manuscript/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ops.GetDirPath(tt.category)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetFilePath(t *testing.T) {
	ops := NewOperations()

	tests := []struct {
		name      string
		category  string
		videoName string
		extension string
		expected  string
	}{
		{
			name:      "simple case",
			category:  "tutorials",
			videoName: "My First Video",
			extension: "md",
			expected:  "manuscript/tutorials/my-first-video.md",
		},
		{
			name:      "name with question mark",
			category:  "faq",
			videoName: "What is Go?",
			extension: "yaml",
			expected:  "manuscript/faq/what-is-go.yaml",
		},
		{
			name:      "empty name",
			category:  "general",
			videoName: "",
			extension: "txt",
			expected:  "manuscript/general/.txt",
		},
		{
			name:      "category with spaces",
			category:  "long form content",
			videoName: "Deep Dive",
			extension: "md",
			expected:  "manuscript/long-form-content/deep-dive.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := ops.GetFilePath(tt.category, tt.videoName, tt.extension)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

// TestGetAnimations tests the GetAnimations function
func TestGetAnimations(t *testing.T) {
	ops := NewOperations()

	type testCase struct {
		name             string
		fileContent      string
		filePath         string // if non-empty, use this instead of creating temp file
		expectedAnims    []string
		expectedSections []string
		expectError      bool
		expectedErrorMsg string
	}

	// Helper function to create a temporary file with content
	createTempFile := func(content string) (string, func()) {
		tmpFile, err := os.CreateTemp(t.TempDir(), "test_gist_*.md")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}
		return tmpFile.Name(), func() { /* os.Remove(tmpFile.Name()) handled by t.TempDir() */ }
	}

	testCases := []testCase{
		{
			name:             "file not found",
			filePath:         "non_existent_file.md",
			expectedAnims:    nil,
			expectedSections: nil,
			expectError:      true,
			expectedErrorMsg: "failed to open file non_existent_file.md:",
		},
		{
			name:             "empty file",
			fileContent:      "",
			expectedAnims:    []string{},
			expectedSections: []string{},
			expectError:      false,
		},
		{
			name:             "only TODOs",
			fileContent:      "TODO: First animation\nTODO: Second animation with spaces  ",
			expectedAnims:    []string{"First animation", "Second animation with spaces"},
			expectedSections: []string{},
			expectError:      false,
		},
		{
			name:             "only sections",
			fileContent:      "## Section One\n  ##   Section Two  \n## Section Three",
			expectedAnims:    []string{"Section: Section One", "Section: Section Two", "Section: Section Three"},
			expectedSections: []string{"Section: Section One", "Section: Section Two", "Section: Section Three"},
			expectError:      false,
		},
		{
			name:             "mix of TODOs and sections",
			fileContent:      "TODO: Anim 1\n## Section Alpha\nTODO: Anim 2 in section\n## Section Beta",
			expectedAnims:    []string{"Anim 1", "Section: Section Alpha", "Anim 2 in section", "Section: Section Beta"},
			expectedSections: []string{"Section: Section Alpha", "Section: Section Beta"},
			expectError:      false,
		},
		{
			name:             "ignored sections but process their TODOs",
			fileContent:      "## Intro\nTODO: Intro anim\n## Setup\nTODO: Setup anim\n## Real Section\nTODO: Real anim\n## Destroy\nTODO: Destroy anim",
			expectedAnims:    []string{"Intro anim", "Setup anim", "Section: Real Section", "Real anim", "Destroy anim"},
			expectedSections: []string{"Section: Real Section"},
			expectError:      false,
		},
		{
			name:             "lines with leading trailing spaces and non-breaking spaces",
			fileContent:      "  TODO:  Anim with spaces  \n\t##\u00a0Section\u00a0with\u00a0NBSP  ",
			expectedAnims:    []string{"Anim with spaces", "Section: Section with NBSP"},
			expectedSections: []string{"Section: Section with NBSP"},
			expectError:      false,
		},
		{
			name:             "empty lines and comments",
			fileContent:      "# This is a comment\n\nTODO: Actual task\n\n## Real Section\n// Another comment style",
			expectedAnims:    []string{"Actual task", "Section: Real Section"},
			expectedSections: []string{"Section: Real Section"},
			expectError:      false,
		},
		{
			name:             "no relevant lines",
			fileContent:      "Just some random text.\nAnother line without TODO or section.",
			expectedAnims:    []string{},
			expectedSections: []string{},
			expectError:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := tc.filePath
			if filePath == "" { // Create temp file if no specific path is given
				var cleanup func()
				filePath, cleanup = createTempFile(tc.fileContent)
				defer cleanup() // t.TempDir() handles actual file removal
			}

			anims, sections, err := ops.GetAnimations(filePath)

			if tc.expectError {
				assert.Error(t, err, "Expected an error")
				if tc.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tc.expectedErrorMsg, "Error message should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Did not expect an error")
			}

			assert.Equal(t, tc.expectedAnims, anims, "Animations did not match")
			assert.Equal(t, tc.expectedSections, sections, "Sections did not match")
		})
	}
}
