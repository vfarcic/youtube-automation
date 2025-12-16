package manuscript

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestInsertShortMarkers_SingleShort(t *testing.T) {
	// Create temp directory and manuscript file
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := `# Introduction

This is some intro text.

Here is an important point that should be a short. It contains valuable information.

And here is more content after.
`
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	shorts := []storage.Short{
		{
			ID:    "short1",
			Title: "Important Point",
			Text:  "Here is an important point that should be a short. It contains valuable information.",
		},
	}

	err := InsertShortMarkers(manuscriptPath, shorts)
	if err != nil {
		t.Fatalf("InsertShortMarkers failed: %v", err)
	}

	// Read result
	result, err := os.ReadFile(manuscriptPath)
	if err != nil {
		t.Fatalf("failed to read result: %v", err)
	}

	resultStr := string(result)

	// Check markers are present
	if !strings.Contains(resultStr, "TODO: Short (id: short1) (start)") {
		t.Error("start marker not found in result")
	}
	if !strings.Contains(resultStr, "TODO: Short (id: short1) (end)") {
		t.Error("end marker not found in result")
	}

	// Check original text is preserved
	if !strings.Contains(resultStr, "Here is an important point") {
		t.Error("original text was not preserved")
	}
}

func TestInsertShortMarkers_MultipleShorts(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := `# Video Script

First segment here with some content.

Second segment here with different content.

Third segment here with more content.
`
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	shorts := []storage.Short{
		{ID: "short1", Title: "First", Text: "First segment here with some content."},
		{ID: "short2", Title: "Second", Text: "Second segment here with different content."},
	}

	err := InsertShortMarkers(manuscriptPath, shorts)
	if err != nil {
		t.Fatalf("InsertShortMarkers failed: %v", err)
	}

	result, _ := os.ReadFile(manuscriptPath)
	resultStr := string(result)

	// Check both shorts have markers
	if !strings.Contains(resultStr, "TODO: Short (id: short1) (start)") {
		t.Error("short1 start marker not found")
	}
	if !strings.Contains(resultStr, "TODO: Short (id: short1) (end)") {
		t.Error("short1 end marker not found")
	}
	if !strings.Contains(resultStr, "TODO: Short (id: short2) (start)") {
		t.Error("short2 start marker not found")
	}
	if !strings.Contains(resultStr, "TODO: Short (id: short2) (end)") {
		t.Error("short2 end marker not found")
	}
}

func TestInsertShortMarkers_EmptyShorts(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := "Some content"
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	err := InsertShortMarkers(manuscriptPath, []storage.Short{})
	if err != nil {
		t.Errorf("expected no error for empty shorts, got: %v", err)
	}

	// Content should be unchanged
	result, _ := os.ReadFile(manuscriptPath)
	if string(result) != content {
		t.Error("content was modified for empty shorts")
	}
}

func TestInsertShortMarkers_TextNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := "Some manuscript content here."
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	shorts := []storage.Short{
		{ID: "short1", Title: "Missing", Text: "This text does not exist in manuscript"},
	}

	err := InsertShortMarkers(manuscriptPath, shorts)
	if err == nil {
		t.Error("expected error when text not found, got nil")
	}
	if !strings.Contains(err.Error(), "no short segments found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestInsertShortMarkers_PartialMatch(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := `First segment exists here.

Second segment does not exist.

Third segment exists here too.
`
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	shorts := []storage.Short{
		{ID: "short1", Title: "First", Text: "First segment exists here."},
		{ID: "short2", Title: "Missing", Text: "This text is nowhere to be found"},
		{ID: "short3", Title: "Third", Text: "Third segment exists here too."},
	}

	err := InsertShortMarkers(manuscriptPath, shorts)
	// Should return warning about short2 not found
	if err == nil {
		t.Error("expected warning error for partial match")
	}
	if !strings.Contains(err.Error(), "short2") {
		t.Errorf("error should mention short2: %v", err)
	}

	// But markers should still be inserted for found shorts
	result, _ := os.ReadFile(manuscriptPath)
	resultStr := string(result)

	if !strings.Contains(resultStr, "TODO: Short (id: short1)") {
		t.Error("short1 marker should be present")
	}
	if !strings.Contains(resultStr, "TODO: Short (id: short3)") {
		t.Error("short3 marker should be present")
	}
}

func TestInsertShortMarkers_FileNotFound(t *testing.T) {
	err := InsertShortMarkers("/nonexistent/path/manuscript.md", []storage.Short{
		{ID: "short1", Text: "some text"},
	})
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestInsertShortMarkers_WhitespaceNormalization(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	// Manuscript has different whitespace than the short text
	content := `Here is some   text with
multiple    spaces and
newlines in it.
`
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	// Short text has normalized whitespace
	shorts := []storage.Short{
		{ID: "short1", Title: "Test", Text: "Here is some text with multiple spaces and newlines in it."},
	}

	err := InsertShortMarkers(manuscriptPath, shorts)
	if err != nil {
		t.Fatalf("InsertShortMarkers failed with whitespace normalization: %v", err)
	}

	result, _ := os.ReadFile(manuscriptPath)
	resultStr := string(result)

	if !strings.Contains(resultStr, "TODO: Short (id: short1) (start)") {
		t.Error("start marker not found after whitespace normalization")
	}
}

func TestRemoveShortMarkers(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := `# Introduction

TODO: Short (id: short1) (start)

Here is important content.

TODO: Short (id: short1) (end)

More content here.

TODO: Short (id: short2) (start)

Another segment.

TODO: Short (id: short2) (end)
`
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	err := RemoveShortMarkers(manuscriptPath)
	if err != nil {
		t.Fatalf("RemoveShortMarkers failed: %v", err)
	}

	result, _ := os.ReadFile(manuscriptPath)
	resultStr := string(result)

	if strings.Contains(resultStr, "TODO: Short") {
		t.Error("markers were not removed")
	}
	if !strings.Contains(resultStr, "Here is important content.") {
		t.Error("content was incorrectly removed")
	}
	if !strings.Contains(resultStr, "Another segment.") {
		t.Error("content was incorrectly removed")
	}
}

func TestExtractShortText(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := `# Introduction

TODO: Short (id: short1) (start)

Here is the extracted content for short1.

TODO: Short (id: short1) (end)

Other content.

TODO: Short (id: short2) (start)

Content for short2 here.

TODO: Short (id: short2) (end)
`
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	// Extract short1
	text1, err := ExtractShortText(manuscriptPath, "short1")
	if err != nil {
		t.Fatalf("failed to extract short1: %v", err)
	}
	if text1 != "Here is the extracted content for short1." {
		t.Errorf("unexpected extracted text: %q", text1)
	}

	// Extract short2
	text2, err := ExtractShortText(manuscriptPath, "short2")
	if err != nil {
		t.Fatalf("failed to extract short2: %v", err)
	}
	if text2 != "Content for short2 here." {
		t.Errorf("unexpected extracted text: %q", text2)
	}
}

func TestExtractShortText_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := "No markers here."
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	_, err := ExtractShortText(manuscriptPath, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent short ID")
	}
}

func TestFindTextPosition_ExactMatch(t *testing.T) {
	content := "Hello world, this is a test."
	text := "this is a test"

	start, end, found := findTextPosition(content, text)
	if !found {
		t.Fatal("text should be found")
	}
	if content[start:end] != text {
		t.Errorf("extracted text %q doesn't match %q", content[start:end], text)
	}
}

func TestFindTextPosition_NotFound(t *testing.T) {
	content := "Hello world"
	text := "goodbye"

	_, _, found := findTextPosition(content, text)
	if found {
		t.Error("text should not be found")
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello  world", "hello world"},
		{"hello\nworld", "hello world"},
		{"hello\t\tworld", "hello world"},
		{"  hello world  ", "hello world"},
		{"hello\n\n\nworld", "hello world"},
		{"a   b   c", "a b c"},
	}

	for _, tt := range tests {
		result := normalizeWhitespace(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestInsertShortMarkers_MarkerOrder(t *testing.T) {
	// Test that markers appear in correct order relative to content
	tmpDir := t.TempDir()
	manuscriptPath := filepath.Join(tmpDir, "manuscript.md")

	content := `Before content.

Target segment here.

After content.
`
	if err := os.WriteFile(manuscriptPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test manuscript: %v", err)
	}

	shorts := []storage.Short{
		{ID: "short1", Title: "Target", Text: "Target segment here."},
	}

	err := InsertShortMarkers(manuscriptPath, shorts)
	if err != nil {
		t.Fatalf("InsertShortMarkers failed: %v", err)
	}

	result, _ := os.ReadFile(manuscriptPath)
	resultStr := string(result)

	// Check order: start marker -> content -> end marker
	startIdx := strings.Index(resultStr, "TODO: Short (id: short1) (start)")
	contentIdx := strings.Index(resultStr, "Target segment here.")
	endIdx := strings.Index(resultStr, "TODO: Short (id: short1) (end)")

	if startIdx == -1 || contentIdx == -1 || endIdx == -1 {
		t.Fatal("markers or content not found")
	}

	if !(startIdx < contentIdx && contentIdx < endIdx) {
		t.Errorf("incorrect order: start=%d, content=%d, end=%d", startIdx, contentIdx, endIdx)
	}
}
