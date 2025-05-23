package filesystem

import (
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
