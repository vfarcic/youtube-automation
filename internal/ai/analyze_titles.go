package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	"devopstoolkit/youtube-automation/internal/publishing"
)

//go:embed templates/analyze-titles.md
var analyzeTitlesTemplate string

// TitleAnalysisData holds the data passed to the analysis template
type TitleAnalysisData struct {
	Videos    []publishing.VideoAnalytics
	StartDate string
	EndDate   string
}

// AnalyzeTitles analyzes video performance data and generates recommendations
// for improving title generation based on what actually works for the channel.
//
// Parameters:
//   - ctx: Context for the AI provider call
//   - analytics: Video performance data from YouTube Analytics API
//
// Returns:
//   - string: Markdown-formatted analysis with recommendations
//   - error: Any error encountered during template rendering or AI generation
func AnalyzeTitles(ctx context.Context, analytics []publishing.VideoAnalytics) (string, error) {
	if len(analytics) == 0 {
		return "", fmt.Errorf("no analytics data provided for analysis")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to get AI provider: %w", err)
	}

	// Parse embedded template
	tmpl, err := template.New("analyze-titles").Parse(analyzeTitlesTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Calculate date range from analytics data
	startDate := analytics[0].PublishedAt
	endDate := analytics[0].PublishedAt
	for _, video := range analytics {
		if video.PublishedAt.Before(startDate) {
			startDate = video.PublishedAt
		}
		if video.PublishedAt.After(endDate) {
			endDate = video.PublishedAt
		}
	}

	// Prepare template data
	data := TitleAnalysisData{
		Videos:    analytics,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}

	// Execute template to generate prompt
	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	prompt := promptBuf.String()

	// Generate analysis using AI provider
	// Use a large token limit since we want comprehensive analysis
	responseContent, err := provider.GenerateContent(ctx, prompt, 4096)
	if err != nil {
		return "", fmt.Errorf("AI analysis generation failed: %w", err)
	}

	if len(responseContent) == 0 {
		return "", fmt.Errorf("AI returned empty analysis")
	}

	return responseContent, nil
}
