package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestExtractDateFromISO(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid RFC3339 format",
			input:    "2025-12-06T15:30:00Z",
			expected: "2025-12-06",
		},
		{
			name:     "valid RFC3339 with timezone offset",
			input:    "2025-01-15T10:00:00+02:00",
			expected: "2025-01-15",
		},
		{
			name:     "valid RFC3339 with milliseconds",
			input:    "2024-06-20T08:45:30.123Z",
			expected: "2024-06-20",
		},
		{
			name:     "date-only string (fallback extraction)",
			input:    "2025-03-10",
			expected: "2025-03-10",
		},
		{
			name:     "empty string returns today",
			input:    "",
			expected: time.Now().UTC().Format("2006-01-02"),
		},
		{
			name:     "invalid format returns today",
			input:    "not-a-date",
			expected: time.Now().UTC().Format("2006-01-02"),
		},
		{
			name:     "partial date returns today",
			input:    "2025",
			expected: time.Now().UTC().Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDateFromISO(tt.input)
			if result != tt.expected {
				t.Errorf("extractDateFromISO(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSaveAMAFiles(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	// Create a minimal MenuHandler
	handler := &MenuHandler{}

	tests := []struct {
		name        string
		videoID     string
		title       string
		description string
		tags        string
		timecodes   string
		publishedAt string
		transcript  string
		wantErr     bool
	}{
		{
			name:        "successful save with all fields",
			videoID:     "test123",
			title:       "Test AMA Title",
			description: "Test description",
			tags:        "tag1, tag2, tag3",
			timecodes:   "00:00 Intro\n02:30 First question",
			publishedAt: "2025-12-06T15:30:00Z",
			transcript:  "This is the transcript content.",
			wantErr:     false,
		},
		{
			name:        "successful save with empty publishedAt uses today",
			videoID:     "video456",
			title:       "Another AMA",
			description: "Description here",
			tags:        "kubernetes, devops",
			timecodes:   "00:00 Start",
			publishedAt: "",
			transcript:  "Transcript text.",
			wantErr:     false,
		},
		{
			name:        "successful save with minimal fields",
			videoID:     "min789",
			title:       "",
			description: "",
			tags:        "",
			timecodes:   "",
			publishedAt: "2025-01-01T00:00:00Z",
			transcript:  "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.saveAMAFiles(tt.videoID, tt.title, tt.description, tt.tags, tt.timecodes, tt.publishedAt, tt.transcript)

			if (err != nil) != tt.wantErr {
				t.Errorf("saveAMAFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify files were created
				datePrefix := extractDateFromISO(tt.publishedAt)
				baseName := datePrefix + "-" + tt.videoID

				yamlPath := filepath.Join("manuscript/ama", baseName+".yaml")
				mdPath := filepath.Join("manuscript/ama", baseName+".md")

				// Check YAML file exists
				if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
					t.Errorf("YAML file not created: %s", yamlPath)
				}

				// Check MD file exists and contains expected content
				mdContent, err := os.ReadFile(mdPath)
				if err != nil {
					t.Errorf("Failed to read MD file: %v", err)
				} else {
					if tt.title != "" && !strings.Contains(string(mdContent), tt.title) {
						t.Errorf("MD file should contain title %q", tt.title)
					}
					if tt.videoID != "" && !strings.Contains(string(mdContent), tt.videoID) {
						t.Errorf("MD file should contain video ID %q", tt.videoID)
					}
					if tt.transcript != "" && !strings.Contains(string(mdContent), tt.transcript) {
						t.Errorf("MD file should contain transcript")
					}
				}
			}
		})
	}
}
