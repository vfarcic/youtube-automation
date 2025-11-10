package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/publishing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAnalysisFiles_Success(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()

	mockAnalytics := []publishing.VideoAnalytics{
		{
			VideoID:             "test123",
			Title:               "Test Video 1",
			Views:               1000,
			CTR:                 5.5,
			AverageViewDuration: 120.0,
			Likes:               50,
			Comments:            10,
			PublishedAt:         time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			VideoID:             "test456",
			Title:               "Test Video 2",
			Views:               2000,
			CTR:                 6.5,
			AverageViewDuration: 180.0,
			Likes:               100,
			Comments:            20,
			PublishedAt:         time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}

	mockAnalysis := "# Test Analysis\n\nHigh-performing titles use numbers and questions."
	channelID := "UC1234567890"

	// Act
	files, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis, tempDir, channelID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, files)

	// Verify JSON file was created
	assert.FileExists(t, files.JSONPath)
	assert.Contains(t, files.JSONPath, "youtube-analytics-")
	assert.Contains(t, files.JSONPath, ".json")

	// Verify JSON content
	jsonData, err := os.ReadFile(files.JSONPath)
	require.NoError(t, err)

	var savedAnalytics []publishing.VideoAnalytics
	err = json.Unmarshal(jsonData, &savedAnalytics)
	require.NoError(t, err)
	assert.Equal(t, 2, len(savedAnalytics))
	assert.Equal(t, "test123", savedAnalytics[0].VideoID)
	assert.Equal(t, "Test Video 1", savedAnalytics[0].Title)
	assert.Equal(t, int64(1000), savedAnalytics[0].Views)

	// Verify Markdown file was created
	assert.FileExists(t, files.MDPath)
	assert.Contains(t, files.MDPath, "title-analysis-")
	assert.Contains(t, files.MDPath, ".md")

	// Verify Markdown content
	mdData, err := os.ReadFile(files.MDPath)
	require.NoError(t, err)
	mdContent := string(mdData)

	assert.Contains(t, mdContent, "# YouTube Title Analysis")
	assert.Contains(t, mdContent, "**Videos Analyzed**: 2")
	assert.Contains(t, mdContent, "**Date Range**: Last 365 days")
	assert.Contains(t, mdContent, "**Channel ID**: UC1234567890")
	assert.Contains(t, mdContent, "High-performing titles use numbers and questions")
}

func TestSaveAnalysisFiles_EmptyAnalytics(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	mockAnalysis := "# Test Analysis"
	channelID := "UC1234567890"

	// Act
	files, err := SaveAnalysisFiles([]publishing.VideoAnalytics{}, mockAnalysis, tempDir, channelID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "no analytics data to save")
}

func TestSaveAnalysisFiles_EmptyAnalysis(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	mockAnalytics := []publishing.VideoAnalytics{
		{VideoID: "test123", Title: "Test Video", Views: 1000},
	}
	channelID := "UC1234567890"

	// Act
	files, err := SaveAnalysisFiles(mockAnalytics, "", tempDir, channelID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "no analysis content to save")
}

func TestSaveAnalysisFiles_EmptyOutputDir(t *testing.T) {
	// Arrange
	mockAnalytics := []publishing.VideoAnalytics{
		{VideoID: "test123", Title: "Test Video", Views: 1000},
	}
	mockAnalysis := "# Test Analysis"
	channelID := "UC1234567890"

	// Act
	files, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis, "", channelID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "output directory not specified")
}

func TestSaveAnalysisFiles_InvalidOutputDir(t *testing.T) {
	// Arrange
	invalidDir := "/nonexistent/invalid/directory/path"
	mockAnalytics := []publishing.VideoAnalytics{
		{VideoID: "test123", Title: "Test Video", Views: 1000},
	}
	mockAnalysis := "# Test Analysis"
	channelID := "UC1234567890"

	// Act
	files, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis, invalidDir, channelID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, files)
	assert.Contains(t, err.Error(), "failed to write")
}

func TestSaveAnalysisFiles_JSONStructure(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	mockAnalytics := []publishing.VideoAnalytics{
		{
			VideoID:             "abc123",
			Title:               "Complete Test",
			Views:               5000,
			CTR:                 7.5,
			AverageViewDuration: 250.5,
			Likes:               200,
			Comments:            30,
			PublishedAt:         time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
		},
	}
	mockAnalysis := "# Analysis"
	channelID := "UCTEST"

	// Act
	files, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis, tempDir, channelID)

	// Assert
	require.NoError(t, err)

	// Verify JSON is pretty-printed (has indentation)
	jsonData, err := os.ReadFile(files.JSONPath)
	require.NoError(t, err)
	jsonStr := string(jsonData)

	// Pretty-printed JSON should have newlines and spaces
	assert.Contains(t, jsonStr, "\n")
	assert.Contains(t, jsonStr, "  ") // 2-space indentation

	// Verify all fields are present (fields are PascalCase as defined in struct)
	assert.Contains(t, jsonStr, `"VideoID"`)
	assert.Contains(t, jsonStr, `"Title"`)
	assert.Contains(t, jsonStr, `"Views"`)
	assert.Contains(t, jsonStr, `"CTR"`)
	assert.Contains(t, jsonStr, `"AverageViewDuration"`)
	assert.Contains(t, jsonStr, `"Likes"`)
	assert.Contains(t, jsonStr, `"Comments"`)
	assert.Contains(t, jsonStr, `"PublishedAt"`)
}

func TestSaveAnalysisFiles_FilenameFormat(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	mockAnalytics := []publishing.VideoAnalytics{
		{VideoID: "test", Title: "Test", Views: 100},
	}
	mockAnalysis := "# Test"
	channelID := "UC123"

	// Act
	files, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis, tempDir, channelID)

	// Assert
	require.NoError(t, err)

	// Verify filename format (YYYY-MM-DD)
	expectedDate := time.Now().Format("2006-01-02")
	expectedJSONName := "youtube-analytics-" + expectedDate + ".json"
	expectedMDName := "title-analysis-" + expectedDate + ".md"

	assert.Equal(t, filepath.Join(tempDir, expectedJSONName), files.JSONPath)
	assert.Equal(t, filepath.Join(tempDir, expectedMDName), files.MDPath)
}

func TestSaveAnalysisFiles_MarkdownMetadata(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	mockAnalytics := []publishing.VideoAnalytics{
		{VideoID: "test1", Title: "Test 1", Views: 100},
		{VideoID: "test2", Title: "Test 2", Views: 200},
		{VideoID: "test3", Title: "Test 3", Views: 300},
	}
	mockAnalysis := "Analysis content here"
	channelID := "UC999"

	// Act
	files, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis, tempDir, channelID)

	// Assert
	require.NoError(t, err)

	mdData, err := os.ReadFile(files.MDPath)
	require.NoError(t, err)
	mdContent := string(mdData)

	// Verify metadata is present and accurate
	assert.Contains(t, mdContent, "**Generated**:")
	assert.Contains(t, mdContent, "**Videos Analyzed**: 3") // Exact count
	assert.Contains(t, mdContent, "**Date Range**: Last 365 days")
	assert.Contains(t, mdContent, "**Channel ID**: UC999") // Exact channel ID
	assert.Contains(t, mdContent, "---") // Separator
	assert.Contains(t, mdContent, "Analysis content here") // Actual analysis
}

func TestSaveAnalysisFiles_OverwriteExisting(t *testing.T) {
	// Arrange
	tempDir := t.TempDir()
	mockAnalytics := []publishing.VideoAnalytics{
		{VideoID: "test", Title: "Test", Views: 100},
	}
	mockAnalysis1 := "# First Analysis"
	mockAnalysis2 := "# Second Analysis"
	channelID := "UC123"

	// Act - Save first time
	files1, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis1, tempDir, channelID)
	require.NoError(t, err)

	// Act - Save second time (same day, should overwrite)
	files2, err := SaveAnalysisFiles(mockAnalytics, mockAnalysis2, tempDir, channelID)
	require.NoError(t, err)

	// Assert - Paths should be the same
	assert.Equal(t, files1.JSONPath, files2.JSONPath)
	assert.Equal(t, files1.MDPath, files2.MDPath)

	// Assert - Content should be from second save
	mdData, err := os.ReadFile(files2.MDPath)
	require.NoError(t, err)
	assert.Contains(t, string(mdData), "Second Analysis")
	assert.NotContains(t, string(mdData), "First Analysis")
}
