package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"
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
//   - analytics: Data to save as JSON (any JSON-serializable value)
//   - analysis: AI-generated analysis text to save as Markdown
//   - outputDir: Directory where files should be saved (typically "tmp")
//   - channelID: YouTube channel ID to include in metadata
//
// Returns:
//   - AnalysisFiles: Paths to the created files
//   - error: Any error encountered during file operations
func SaveAnalysisFiles(analytics interface{}, analysis string, outputDir string, channelID string) (*AnalysisFiles, error) {
	if isEmptyAnalytics(analytics) {
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

	// Count items for metadata
	count := reflectLen(analytics)

	// Build markdown with metadata header
	mdContent := fmt.Sprintf(`# YouTube Title Analysis

**Generated**: %s
**Videos Analyzed**: %d
**Date Range**: Last 365 days
**Channel ID**: %s

---

%s
`, time.Now().Format("2006-01-02 15:04:05"), count, channelID, analysis)

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

// CompleteAnalysisFiles represents all files saved for a complete analysis
type CompleteAnalysisFiles struct {
	AnalyticsPath string // 01-analytics.json
	ResponsePath  string // 02-ai-response.txt
	ResultPath    string // 03-result.md
}

// SaveCompleteAnalysis saves a complete analysis with all audit trail files.
// This creates a timestamped directory with 3 files for full traceability.
// Note: The prompt is saved separately before the LLM call (in the ai package)
// so it's available for inspection even when the LLM call or parsing fails.
//
// Parameters:
//   - analysisType: Type of analysis (e.g., "title-analysis", "timing-analysis")
//   - analytics: Data to save as JSON (any JSON-serializable value)
//   - rawResponse: Raw AI response (saved as 02-ai-response.txt)
//   - formattedResult: User-friendly formatted result (saved as 03-result.md)
//   - outputDir: Base directory where analysis folder will be created
//   - channelID: YouTube channel ID for metadata
//
// Returns:
//   - CompleteAnalysisFiles: Paths to all created files
//   - error: Any error encountered during file operations
func SaveCompleteAnalysis(
	analysisType string,
	analytics interface{},
	rawResponse string,
	formattedResult string,
	outputDir string,
	channelID string,
) (*CompleteAnalysisFiles, error) {
	if isEmptyAnalytics(analytics) {
		return nil, fmt.Errorf("no analytics data to save")
	}
	if rawResponse == "" {
		return nil, fmt.Errorf("no raw response to save")
	}
	if formattedResult == "" {
		return nil, fmt.Errorf("no formatted result to save")
	}
	if outputDir == "" {
		return nil, fmt.Errorf("output directory not specified")
	}
	if analysisType == "" {
		return nil, fmt.Errorf("analysis type not specified")
	}

	// Create timestamped analysis directory
	timestamp := time.Now().Format("2006-01-02")
	analysisDir := filepath.Join(outputDir, fmt.Sprintf("%s-%s", analysisType, timestamp))

	// Create directory if it doesn't exist
	if err := os.MkdirAll(analysisDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create analysis directory: %w", err)
	}

	// Define file paths
	analyticsPath := filepath.Join(analysisDir, "01-analytics.json")
	responsePath := filepath.Join(analysisDir, "02-ai-response.txt")
	resultPath := filepath.Join(analysisDir, "03-result.md")

	// Save 01-analytics.json
	analyticsJSON, err := json.MarshalIndent(analytics, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analytics: %w", err)
	}
	if err := os.WriteFile(analyticsPath, analyticsJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write analytics file: %w", err)
	}

	// Save 02-ai-response.txt
	if err := os.WriteFile(responsePath, []byte(rawResponse), 0644); err != nil {
		return nil, fmt.Errorf("failed to write response file: %w", err)
	}

	// Save 03-result.md
	if err := os.WriteFile(resultPath, []byte(formattedResult), 0644); err != nil {
		return nil, fmt.Errorf("failed to write result file: %w", err)
	}

	return &CompleteAnalysisFiles{
		AnalyticsPath: analyticsPath,
		ResponsePath:  responsePath,
		ResultPath:    resultPath,
	}, nil
}

// isEmptyAnalytics checks if the analytics value is nil or an empty slice/array.
func isEmptyAnalytics(analytics interface{}) bool {
	if analytics == nil {
		return true
	}
	v := reflect.ValueOf(analytics)
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		return v.Len() == 0
	}
	return false
}

// reflectLen returns the length of a slice/array, or 0 for other types.
func reflectLen(v interface{}) int {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		return rv.Len()
	}
	return 0
}
