package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/publishing"
)

//go:embed templates/analyze-timing.md
var analyzeTimingTemplate string

// TimingAnalysisData holds the data passed to the timing analysis template
type TimingAnalysisData struct {
	CurrentPattern    []TimeSlotSummary
	PerformanceBySlot []publishing.TimeSlotPerformance
	TotalVideos       int
}

// TimeSlotSummary represents a summary of current publishing patterns
type TimeSlotSummary struct {
	DayOfWeek  string
	TimeOfDay  string
	Count      int
	Percentage float64
}

// GenerateTimingRecommendations analyzes video performance data and generates
// timing recommendations for optimal publishing schedules.
//
// Parameters:
//   - ctx: Context for the AI provider call
//   - analytics: Video performance data from YouTube Analytics API
//
// Returns:
//   - []configuration.TimingRecommendation: Array of 6-8 timing recommendations
//   - string: The prompt sent to AI (for audit trail)
//   - string: Raw AI response (for audit trail)
//   - error: Any error encountered during analysis or parsing
func GenerateTimingRecommendations(ctx context.Context, analytics []publishing.VideoAnalytics) ([]configuration.TimingRecommendation, string, string, error) {
	if len(analytics) == 0 {
		return nil, "", "", fmt.Errorf("no analytics data provided for timing analysis")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to get AI provider: %w", err)
	}

	// First, fetch first-week metrics for all videos (critical for accurate comparison)
	enrichedWithMetrics, err := publishing.EnrichWithFirstWeekMetrics(ctx, analytics)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to fetch first-week metrics: %w", err)
	}

	// Then enrich with timing data (day/time extraction and engagement calculation)
	enriched := publishing.EnrichWithTimingData(enrichedWithMetrics)

	// Group by time slot and calculate performance
	grouped := publishing.GroupByTimeSlot(enriched)
	performance := publishing.CalculateTimeSlotPerformance(grouped)

	// Filter out time slots with insufficient data (< 3 videos)
	// Statistical significance requires minimum sample size
	filteredPerformance := filterTimeSlotsByMinVideos(performance, 3)

	// Calculate current pattern summary (only for slots with enough data)
	currentPattern := calculateCurrentPattern(grouped, len(analytics))

	// Prepare template data
	data := TimingAnalysisData{
		CurrentPattern:    currentPattern,
		PerformanceBySlot: filteredPerformance,
		TotalVideos:       len(analytics),
	}

	// Parse embedded template
	tmpl, err := template.New("analyze-timing").Parse(analyzeTimingTemplate)
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template to generate prompt
	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return nil, "", "", fmt.Errorf("failed to execute template: %w", err)
	}

	prompt := promptBuf.String()

	// Generate recommendations using AI provider
	// Use 4096 tokens to allow for detailed reasoning
	rawResponse, err := provider.GenerateContent(ctx, prompt, 4096)
	if err != nil {
		return nil, prompt, "", fmt.Errorf("AI timing analysis failed: %w", err)
	}

	if len(rawResponse) == 0 {
		return nil, prompt, "", fmt.Errorf("AI returned empty timing recommendations")
	}

	// Parse JSON response
	recommendations, err := parseTimingRecommendations(rawResponse)
	if err != nil {
		return nil, prompt, rawResponse, fmt.Errorf("failed to parse AI recommendations: %w", err)
	}

	// Validate recommendations
	if err := validateRecommendations(recommendations); err != nil {
		return nil, prompt, rawResponse, fmt.Errorf("invalid recommendations: %w", err)
	}

	return recommendations, prompt, rawResponse, nil
}

// calculateCurrentPattern summarizes the current publishing pattern
func calculateCurrentPattern(grouped map[publishing.TimeSlot][]publishing.VideoAnalytics, totalVideos int) []TimeSlotSummary {
	summaries := make([]TimeSlotSummary, 0, len(grouped))

	for slot, videos := range grouped {
		count := len(videos)
		percentage := (float64(count) / float64(totalVideos)) * 100

		summaries = append(summaries, TimeSlotSummary{
			DayOfWeek:  slot.DayOfWeek,
			TimeOfDay:  slot.TimeOfDay,
			Count:      count,
			Percentage: percentage,
		})
	}

	return summaries
}

// filterTimeSlotsByMinVideos filters out time slots with insufficient video count
// for statistical significance.
//
// Parameters:
//   - performance: Array of time slot performance data
//   - minVideos: Minimum number of videos required per slot
//
// Returns:
//   - []publishing.TimeSlotPerformance: Filtered array with only slots meeting minimum
func filterTimeSlotsByMinVideos(performance []publishing.TimeSlotPerformance, minVideos int) []publishing.TimeSlotPerformance {
	filtered := make([]publishing.TimeSlotPerformance, 0, len(performance))

	for _, perf := range performance {
		if perf.VideoCount >= minVideos {
			filtered = append(filtered, perf)
		}
	}

	return filtered
}

// parseTimingRecommendations extracts JSON from AI response and parses into recommendations
func parseTimingRecommendations(response string) ([]configuration.TimingRecommendation, error) {
	var recommendations []configuration.TimingRecommendation
	err := ParseJSONResponse(response, &recommendations)
	if err != nil {
		return nil, err
	}
	return recommendations, nil
}

// validateRecommendations ensures recommendations meet requirements
func validateRecommendations(recommendations []configuration.TimingRecommendation) error {
	if len(recommendations) < 6 || len(recommendations) > 8 {
		return fmt.Errorf("expected 6-8 recommendations, got %d", len(recommendations))
	}

	validDays := map[string]bool{
		"Monday": true, "Tuesday": true, "Wednesday": true, "Thursday": true,
		"Friday": true, "Saturday": true, "Sunday": true,
	}

	for i, rec := range recommendations {
		// Validate day
		if !validDays[rec.Day] {
			return fmt.Errorf("recommendation %d has invalid day: %s", i+1, rec.Day)
		}

		// Validate time format (HH:MM)
		if !isValidTimeFormat(rec.Time) {
			return fmt.Errorf("recommendation %d has invalid time format: %s (expected HH:MM)", i+1, rec.Time)
		}

		// Validate reasoning exists
		if strings.TrimSpace(rec.Reasoning) == "" {
			return fmt.Errorf("recommendation %d is missing reasoning", i+1)
		}
	}

	return nil
}

// isValidTimeFormat checks if time string is in HH:MM format
func isValidTimeFormat(timeStr string) bool {
	if len(timeStr) != 5 {
		return false
	}

	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}

	// Check hours (00-23)
	if len(parts[0]) != 2 {
		return false
	}
	hour := 0
	if _, err := fmt.Sscanf(parts[0], "%d", &hour); err != nil || hour < 0 || hour > 23 {
		return false
	}

	// Check minutes (00-59)
	if len(parts[1]) != 2 {
		return false
	}
	minute := 0
	if _, err := fmt.Sscanf(parts[1], "%d", &minute); err != nil || minute < 0 || minute > 59 {
		return false
	}

	return true
}
