package repository

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestGetAnimationsFromMarkdownLogic tests the animation extraction logic
// that will be used for markdown files via the GetAnimations method.
func TestGetAnimationsFromMarkdownLogic(t *testing.T) {
	tests := []struct {
		name               string
		markdownContent    string
		expectedAnimations []string
		expectedSections   []string
		expectError        bool
	}{
		{
			name:               "Valid markdown with sections and TODOs",
			markdownContent:    "## My First Section\nSome content\nTODO: Do this thing\n## My Second Section\nTODO: Another thing",
			expectedAnimations: []string{"Section: My First Section", "Do this thing", "Section: My Second Section", "Another thing"},
			expectedSections:   []string{"Section: My First Section", "Section: My Second Section"},
			expectError:        false,
		},
		{
			name:               "Markdown with only TODOs",
			markdownContent:    "TODO: Task 1\nSome other text\nTODO: Task 2",
			expectedAnimations: []string{"Task 1", "Task 2"},
			expectedSections:   []string{},
			expectError:        false,
		},
		{
			name:               "Markdown with only sections",
			markdownContent:    "## Section A\n## Section B",
			expectedAnimations: []string{"Section: Section A", "Section: Section B"},
			expectedSections:   []string{"Section: Section A", "Section: Section B"},
			expectError:        false,
		},
		{
			name:               "Markdown with ignored section headers",
			markdownContent:    "## Intro\nThis is intro.\n## My Real Section\n## Setup\n## Destroy\nTODO: A task here",
			expectedAnimations: []string{"Section: My Real Section", "A task here"},
			expectedSections:   []string{"Section: My Real Section"},
			expectError:        false,
		},
		{
			name:               "Empty markdown file",
			markdownContent:    "",
			expectedAnimations: []string{},
			expectedSections:   []string{},
			expectError:        false,
		},
		{
			name:               "Markdown with no relevant animation cues",
			markdownContent:    "# This is a H1\nJust some text.\nAnother line.",
			expectedAnimations: []string{},
			expectedSections:   []string{},
			expectError:        false,
		},
		{
			name:               "Markdown with TODOs needing trim",
			markdownContent:    "TODO:    Trimmed Task   \n##    Spaced Section   ",
			expectedAnimations: []string{"Trimmed Task", "Section: Spaced Section"},
			expectedSections:   []string{"Section: Spaced Section"},
			expectError:        false,
		},
		{
			name:               "File does not exist",
			markdownContent:    "", // Content doesn't matter, path will be non-existent
			expectedAnimations: nil,
			expectedSections:   nil,
			expectError:        true,
		},
	}

	repo := Repo{} // Changed to direct instantiation

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var filePath string
			var err error

			if tc.name == "File does not exist" {
				filePath = filepath.Join(t.TempDir(), "non_existent_file.md")
				// Ensure it really doesn't exist, though TempDir should be clean
				os.Remove(filePath)
			} else {
				tempFile, err := os.CreateTemp(t.TempDir(), "test_*.md")
				if err != nil {
					t.Fatalf("Failed to create temp file: %v", err)
				}
				filePath = tempFile.Name()
				if _, err := tempFile.Write([]byte(tc.markdownContent)); err != nil {
					tempFile.Close()
					t.Fatalf("Failed to write to temp file: %v", err)
				}
				tempFile.Close() // Close file before GetAnimations tries to open it
			}

			// We are testing the logic that GetAnimations will use for markdown.
			// Once GetAnimations is refactored to ONLY handle markdown (or directly call
			// a markdown-specific function like getAnimationsFromMarkdown), this call
			// will effectively test that unified logic.
			animations, sections, err := repo.GetAnimations(filePath)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
				// Optionally, check for specific error types or messages if needed
			} else {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
				if !reflect.DeepEqual(animations, tc.expectedAnimations) {
					t.Errorf("Unexpected animations.\nExpected: %v\nGot:      %v", tc.expectedAnimations, animations)
				}
				if !reflect.DeepEqual(sections, tc.expectedSections) {
					t.Errorf("Unexpected sections.\nExpected: %v\nGot:      %v", tc.expectedSections, sections)
				}
			}

			// TempDir will clean up the file if it was created in t.TempDir()
			// If we explicitly created a non-existent path, no cleanup needed for that path.
		})
	}
}
