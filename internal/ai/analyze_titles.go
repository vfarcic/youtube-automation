package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/analyze-titles.md
var analyzeTitlesTemplate string

// TitleAnalysisData holds the data passed to the analysis template
type TitleAnalysisData struct {
	ABData     string
	VideoCount int
}

// TitlePattern represents a pattern found in title analysis
type TitlePattern struct {
	Pattern     string   `json:"pattern"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Examples    []string `json:"examples"`
}

// TitleRecommendation represents an actionable recommendation
type TitleRecommendation struct {
	Recommendation string `json:"recommendation"`
	Evidence       string `json:"evidence"`
	Example        string `json:"example"`
}

// TitleAnalysisResult holds the structured analysis results from AI
type TitleAnalysisResult struct {
	HighPerformingPatterns []TitlePattern        `json:"highPerformingPatterns"`
	LowPerformingPatterns  []TitlePattern        `json:"lowPerformingPatterns"`
	Recommendations        []TitleRecommendation `json:"recommendations"`
	TitlesMDContent        string                `json:"titlesMdContent"`
}

// AnalyzeTitles analyzes video A/B test data and generates recommendations
// for improving title generation based on what actually works for the channel.
//
// Parameters:
//   - ctx: Context for the AI provider call
//   - videos: Video A/B test data enriched with analytics
//
// Returns:
//   - TitleAnalysisResult: Parsed analysis results
//   - string: Raw AI response (for audit trail)
//   - error: Any error encountered during template rendering, AI generation, or parsing
func AnalyzeTitles(ctx context.Context, videos []VideoABData) (TitleAnalysisResult, string, error) {
	var emptyResult TitleAnalysisResult

	if len(videos) == 0 {
		return emptyResult, "", fmt.Errorf("no video data provided for analysis")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return emptyResult, "", fmt.Errorf("failed to get AI provider: %w", err)
	}

	// Parse embedded template
	tmpl, err := template.New("analyze-titles").Parse(analyzeTitlesTemplate)
	if err != nil {
		return emptyResult, "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare template data
	data := TitleAnalysisData{
		ABData:     FormatABDataForPrompt(videos),
		VideoCount: len(videos),
	}

	// Execute template to generate prompt
	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return emptyResult, "", fmt.Errorf("failed to execute template: %w", err)
	}

	prompt := promptBuf.String()

	// Save prompt to audit trail before sending to LLM (only if tmp/ already exists)
	promptDir := filepath.Join(".", "tmp")
	if info, err := os.Stat(promptDir); err == nil && info.IsDir() {
		_ = os.WriteFile(filepath.Join(promptDir, "title-analysis-prompt.md"), []byte(prompt), 0644)
	}

	// Generate analysis using AI provider
	// Use a large token limit since we want comprehensive analysis with titles.md content
	rawResponse, err := provider.GenerateContent(ctx, prompt, 8192)
	if err != nil {
		return emptyResult, "", fmt.Errorf("AI analysis generation failed: %w", err)
	}

	if len(rawResponse) == 0 {
		return emptyResult, "", fmt.Errorf("AI returned empty analysis")
	}

	// Parse JSON response
	var result TitleAnalysisResult
	if err := ParseJSONResponse(rawResponse, &result); err != nil {
		return emptyResult, rawResponse, fmt.Errorf("failed to parse title analysis JSON: %w", err)
	}

	return result, rawResponse, nil
}
