package app

import (
	"os"
	"testing"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/storage"
)

// Note: Success path testing for HandleAnalyzeShorts is not possible in this layer
// because it requires interactive TTY input (huh forms). The AI analysis logic
// is thoroughly tested in internal/ai/shorts_test.go. Here we test validation
// and error handling paths that don't require TTY interaction.

func TestDisplayAndSelectShortCandidates_EmptyCandidates(t *testing.T) {
	handler := &MenuHandler{}

	_, err := handler.displayAndSelectShortCandidates([]ai.ShortCandidate{})
	if err == nil {
		t.Error("Expected error for empty candidates, got nil")
	}
}

func TestDisplayAndSelectShortCandidates_NilCandidates(t *testing.T) {
	handler := &MenuHandler{}

	_, err := handler.displayAndSelectShortCandidates(nil)
	if err == nil {
		t.Error("Expected error for nil candidates, got nil")
	}
}

func TestHandleAnalyzeShorts_EmptyGist(t *testing.T) {
	handler := &MenuHandler{}
	video := &storage.Video{
		Name: "Test Video",
		Gist: "", // Empty gist path
	}

	_, err := handler.HandleAnalyzeShorts(video)
	if err == nil {
		t.Error("Expected error for empty Gist path, got nil")
	}
}

func TestHandleAnalyzeShorts_NonExistentGist(t *testing.T) {
	handler := &MenuHandler{}
	video := &storage.Video{
		Name: "Test Video",
		Gist: "/nonexistent/path/to/manuscript.md",
	}

	_, err := handler.HandleAnalyzeShorts(video)
	if err == nil {
		t.Error("Expected error for non-existent Gist file, got nil")
	}
}

func TestHandleAnalyzeShorts_EmptyManuscript(t *testing.T) {
	handler := &MenuHandler{}

	// Create a temp file with empty content
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/empty.md"
	if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	video := &storage.Video{
		Name: "Test Video",
		Gist: tmpFile,
	}

	_, err := handler.HandleAnalyzeShorts(video)
	if err == nil {
		t.Error("Expected error for empty manuscript, got nil")
	}
}

func TestHandleAnalyzeShorts_WhitespaceOnlyManuscript(t *testing.T) {
	handler := &MenuHandler{}

	// Create a temp file with whitespace-only content
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/whitespace.md"
	if err := os.WriteFile(tmpFile, []byte("   \n\t  \n  "), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	video := &storage.Video{
		Name: "Test Video",
		Gist: tmpFile,
	}

	_, err := handler.HandleAnalyzeShorts(video)
	if err == nil {
		t.Error("Expected error for whitespace-only manuscript, got nil")
	}
}
