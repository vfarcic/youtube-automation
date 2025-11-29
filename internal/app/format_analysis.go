package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/ai"
	"devopstoolkit/youtube-automation/internal/configuration"
)

// FormatTitleAnalysisMarkdown converts TitleAnalysisResult to user-friendly markdown
func FormatTitleAnalysisMarkdown(result ai.TitleAnalysisResult, videoCount int, channelID string) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf(`# YouTube Title Analysis

**Generated**: %s
**Videos Analyzed**: %d
**Channel ID**: %s

---

## High-Performing Title Patterns

`, time.Now().Format("2006-01-02 15:04:05"), videoCount, channelID))

	for i, pattern := range result.HighPerformingPatterns {
		md.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, pattern.Pattern))
		md.WriteString(fmt.Sprintf("**Description**: %s\n\n", pattern.Description))
		md.WriteString(fmt.Sprintf("**Impact**: %s\n\n", pattern.Impact))
		if len(pattern.Examples) > 0 {
			md.WriteString("**Examples**:\n")
			for _, ex := range pattern.Examples {
				md.WriteString(fmt.Sprintf("- \"%s\"\n", ex))
			}
			md.WriteString("\n")
		}
	}

	md.WriteString("## Low-Performing Title Patterns\n\n")
	for i, pattern := range result.LowPerformingPatterns {
		md.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, pattern.Pattern))
		md.WriteString(fmt.Sprintf("**Description**: %s\n\n", pattern.Description))
		md.WriteString(fmt.Sprintf("**Impact**: %s\n\n", pattern.Impact))
		if len(pattern.Examples) > 0 {
			md.WriteString("**Examples**:\n")
			for _, ex := range pattern.Examples {
				md.WriteString(fmt.Sprintf("- \"%s\"\n", ex))
			}
			md.WriteString("\n")
		}
	}

	md.WriteString("## Title Length Analysis\n\n")
	md.WriteString(fmt.Sprintf("**Optimal Range**: %s\n\n", result.TitleLengthAnalysis.OptimalRange))
	md.WriteString(fmt.Sprintf("**Finding**: %s\n\n", result.TitleLengthAnalysis.Finding))
	md.WriteString(fmt.Sprintf("**Data**: %s\n\n", result.TitleLengthAnalysis.Data))

	md.WriteString("## Content Type Analysis\n\n")
	md.WriteString(fmt.Sprintf("**Finding**: %s\n\n", result.ContentTypeAnalysis.Finding))
	if len(result.ContentTypeAnalysis.TopPerformers) > 0 {
		md.WriteString("**Top Performers**:\n")
		for _, tp := range result.ContentTypeAnalysis.TopPerformers {
			md.WriteString(fmt.Sprintf("- %s\n", tp))
		}
		md.WriteString("\n")
	}
	md.WriteString(fmt.Sprintf("**Data**: %s\n\n", result.ContentTypeAnalysis.Data))

	md.WriteString("## Engagement Patterns\n\n")
	md.WriteString(fmt.Sprintf("**Finding**: %s\n\n", result.EngagementPatterns.Finding))
	md.WriteString(fmt.Sprintf("**Likes Pattern**: %s\n\n", result.EngagementPatterns.LikesPattern))
	md.WriteString(fmt.Sprintf("**Comments Pattern**: %s\n\n", result.EngagementPatterns.CommentsPattern))
	md.WriteString(fmt.Sprintf("**Watch Time Pattern**: %s\n\n", result.EngagementPatterns.WatchTimePattern))

	md.WriteString("## Actionable Recommendations\n\n")
	for i, rec := range result.Recommendations {
		md.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, rec.Recommendation))
		md.WriteString(fmt.Sprintf("**Evidence**: %s\n\n", rec.Evidence))
		md.WriteString(fmt.Sprintf("**Example**: %s\n\n", rec.Example))
	}

	md.WriteString("## Prompt Engineering Suggestions\n\n")
	for i, suggestion := range result.PromptSuggestions {
		md.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
	}
	md.WriteString("\n")

	md.WriteString(`---

## Next Steps

1. **Review recommendations**: Evaluate which patterns apply to your content strategy
2. **Update title generation**: Modify internal/ai/titles.go with insights
3. **Test improved titles**: Generate titles using updated patterns
4. **Monitor results**: Re-run analysis in 3-6 months to measure improvement
`)

	return md.String()
}

// FormatTimingRecommendationsMarkdown converts timing recommendations to user-friendly markdown
func FormatTimingRecommendationsMarkdown(recommendations []configuration.TimingRecommendation, videoCount int, channelID string) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf(`# YouTube Timing Analysis & Recommendations

**Generated**: %s
**Videos Analyzed**: %d
**Channel ID**: %s

---

## Timing Recommendations

The following publish times have been recommended based on your channel's performance data and experimental diversity goals:

`, time.Now().Format("2006-01-02 15:04:05"), videoCount, channelID))

	// Add each recommendation to markdown
	for i, rec := range recommendations {
		md.WriteString(fmt.Sprintf("### %d. %s %s UTC\n\n", i+1, rec.Day, rec.Time))
		md.WriteString(fmt.Sprintf("**Reasoning**: %s\n\n", rec.Reasoning))
	}

	md.WriteString(`---

## Next Steps

1. **Review recommendations**: These times have been saved to settings.yaml
2. **Apply to videos**: Use the "Apply Random Timing" button in video Initial Details form
3. **Monitor results**: Re-run this analysis in 3-6 months to see which times performed best
4. **Iterate**: Successful times will be kept, poor performers replaced with new experiments

## How to Apply

When editing a video in the Initial Details phase:
- Click "Apply Random Timing" button
- A random recommendation will be selected
- Date will be updated to match the selected day/time within the same week
- Review and save the video
`)

	return md.String()
}

// FormatTimingRecommendationsJSON converts timing recommendations to formatted JSON string
func FormatTimingRecommendationsJSON(recommendations []configuration.TimingRecommendation) (string, error) {
	jsonBytes, err := json.MarshalIndent(recommendations, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal recommendations: %w", err)
	}
	return string(jsonBytes), nil
}
