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

// TitlePattern represents a pattern found in title analysis
type TitlePattern struct {
	Pattern     string   `json:"pattern"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Examples    []string `json:"examples"`
}

// TitleLengthAnalysis holds findings about optimal title length
type TitleLengthAnalysis struct {
	OptimalRange string `json:"optimalRange"`
	Finding      string `json:"finding"`
	Data         string `json:"data"`
}

// ContentTypeAnalysis holds findings about content types and topics
type ContentTypeAnalysis struct {
	Finding       string   `json:"finding"`
	TopPerformers []string `json:"topPerformers"`
	Data          string   `json:"data"`
}

// EngagementPatterns holds findings about engagement metrics
type EngagementPatterns struct {
	Finding          string `json:"finding"`
	LikesPattern     string `json:"likesPattern"`
	CommentsPattern  string `json:"commentsPattern"`
	WatchTimePattern string `json:"watchTimePattern"`
}

// TitleRecommendation represents an actionable recommendation
type TitleRecommendation struct {
	Recommendation string `json:"recommendation"`
	Evidence       string `json:"evidence"`
	Example        string `json:"example"`
}

// TitleAnalysisResult holds the structured analysis results from AI
type TitleAnalysisResult struct {
	HighPerformingPatterns []TitlePattern          `json:"highPerformingPatterns"`
	LowPerformingPatterns  []TitlePattern          `json:"lowPerformingPatterns"`
	TitleLengthAnalysis    TitleLengthAnalysis     `json:"titleLengthAnalysis"`
	ContentTypeAnalysis    ContentTypeAnalysis     `json:"contentTypeAnalysis"`
	EngagementPatterns     EngagementPatterns      `json:"engagementPatterns"`
	Recommendations        []TitleRecommendation   `json:"recommendations"`
	PromptSuggestions      []string                `json:"promptSuggestions"`
}

// AnalyzeTitles analyzes video performance data and generates recommendations
// for improving title generation based on what actually works for the channel.
//
// Parameters:
//   - ctx: Context for the AI provider call
//   - analytics: Video performance data from YouTube Analytics API
//
// Returns:
//   - TitleAnalysisResult: Parsed analysis results
//   - string: The prompt sent to AI (for audit trail)
//   - string: Raw AI response (for audit trail)
//   - error: Any error encountered during template rendering, AI generation, or parsing
func AnalyzeTitles(ctx context.Context, analytics []publishing.VideoAnalytics) (TitleAnalysisResult, string, string, error) {
	var emptyResult TitleAnalysisResult

	if len(analytics) == 0 {
		return emptyResult, "", "", fmt.Errorf("no analytics data provided for analysis")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return emptyResult, "", "", fmt.Errorf("failed to get AI provider: %w", err)
	}

	// Parse embedded template
	tmpl, err := template.New("analyze-titles").Parse(analyzeTitlesTemplate)
	if err != nil {
		return emptyResult, "", "", fmt.Errorf("failed to parse template: %w", err)
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
		return emptyResult, "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	prompt := promptBuf.String()

	// Generate analysis using AI provider
	// Use a large token limit since we want comprehensive analysis
	rawResponse, err := provider.GenerateContent(ctx, prompt, 4096)
	if err != nil {
		return emptyResult, prompt, "", fmt.Errorf("AI analysis generation failed: %w", err)
	}

	if len(rawResponse) == 0 {
		return emptyResult, prompt, "", fmt.Errorf("AI returned empty analysis")
	}

	// Parse JSON response
	var result TitleAnalysisResult
	if err := ParseJSONResponse(rawResponse, &result); err != nil {
		return emptyResult, prompt, rawResponse, fmt.Errorf("failed to parse title analysis JSON: %w", err)
	}

	return result, prompt, rawResponse, nil
}
