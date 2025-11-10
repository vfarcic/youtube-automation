package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devopstoolkit/youtube-automation/internal/publishing"
)

// AnalysisFiles represents the paths to saved analysis files
type AnalysisFiles struct {
	JSONPath string
	MDPath   string
}

// SaveAnalysisFiles saves video analytics and AI analysis to files.
// This is a pure function that's easy to test.
//
// Parameters:
//   - analytics: Video analytics data to save as JSON
//   - analysis: AI-generated analysis text to save as Markdown
//   - outputDir: Directory where files should be saved (typically "tmp")
//   - channelID: YouTube channel ID to include in metadata
//
// Returns:
//   - AnalysisFiles: Paths to the created files
//   - error: Any error encountered during file operations
func SaveAnalysisFiles(analytics []publishing.VideoAnalytics, analysis string, outputDir string, channelID string) (*AnalysisFiles, error) {
	if len(analytics) == 0 {
		return nil, fmt.Errorf("no analytics data to save")
	}

	if analysis == "" {
		return nil, fmt.Errorf("no analysis content to save")
	}

	if outputDir == "" {
		return nil, fmt.Errorf("output directory not specified")
	}

	// Generate timestamp-based filenames
	timestamp := time.Now().Format("2006-01-02")
	jsonPath := filepath.Join(outputDir, fmt.Sprintf("youtube-analytics-%s.json", timestamp))
	mdPath := filepath.Join(outputDir, fmt.Sprintf("title-analysis-%s.md", timestamp))

	// Save raw analytics data as JSON
	jsonData, err := json.MarshalIndent(analytics, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analytics data: %w", err)
	}

	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write JSON file %s: %w", jsonPath, err)
	}

	// Build markdown with metadata header
	mdContent := fmt.Sprintf(`# YouTube Title Analysis

**Generated**: %s
**Videos Analyzed**: %d
**Date Range**: Last 365 days
**Channel ID**: %s

---

%s
`, time.Now().Format("2006-01-02 15:04:05"), len(analytics), channelID, analysis)

	// Save analysis as Markdown
	err = os.WriteFile(mdPath, []byte(mdContent), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write Markdown file %s: %w", mdPath, err)
	}

	return &AnalysisFiles{
		JSONPath: jsonPath,
		MDPath:   mdPath,
	}, nil
}
