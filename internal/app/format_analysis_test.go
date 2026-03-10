package app

import (
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/ai"
)

func TestFormatTitleAnalysisMarkdown(t *testing.T) {
	tests := []struct {
		name       string
		result     ai.TitleAnalysisResult
		videoCount int
		channelID  string
		wantParts  []string
		dontWant   []string
	}{
		{
			name: "Full result with all fields",
			result: ai.TitleAnalysisResult{
				HighPerformingPatterns: []ai.TitlePattern{
					{Pattern: "Numbers in titles", Description: "Titles with numbers get more share", Impact: "67% avg share", Examples: []string{"Top 5 Tools (70%)", "3 Ways (65%)"}},
				},
				LowPerformingPatterns: []ai.TitlePattern{
					{Pattern: "Vague titles", Description: "Generic titles lose share", Impact: "32% avg share", Examples: []string{"Guide to Tools (30%)"}},
				},
				Recommendations: []ai.TitleRecommendation{
					{Recommendation: "Use numbers", Evidence: "67% share with numbers", Example: "Top 5 X instead of Guide to X"},
				},
				TitlesMDContent: "# Title Prompt\n\nUse numbers.\n\n{{.ManuscriptContent}}",
			},
			videoCount: 25,
			channelID:  "UC123",
			wantParts: []string{
				"# YouTube Title Analysis",
				"**Videos Analyzed**: 25",
				"**Channel ID**: UC123",
				"## High-Performing Title Patterns",
				"Numbers in titles",
				"67% avg share",
				"## Low-Performing Title Patterns",
				"Vague titles",
				"## Actionable Recommendations",
				"Use numbers",
				"## Proposed titles.md Update",
				"# Title Prompt",
				"{{.ManuscriptContent}}",
			},
		},
		{
			name: "Empty TitlesMDContent omits section",
			result: ai.TitleAnalysisResult{
				HighPerformingPatterns: []ai.TitlePattern{},
				LowPerformingPatterns:  []ai.TitlePattern{},
				Recommendations:        []ai.TitleRecommendation{},
				TitlesMDContent:        "",
			},
			videoCount: 5,
			channelID:  "UC456",
			wantParts:  []string{"# YouTube Title Analysis"},
			dontWant:   []string{"## Proposed titles.md Update"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTitleAnalysisMarkdown(tt.result, tt.videoCount, tt.channelID)

			for _, part := range tt.wantParts {
				if !strings.Contains(result, part) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", part, result)
				}
			}

			for _, part := range tt.dontWant {
				if strings.Contains(result, part) {
					t.Errorf("Expected output NOT to contain %q, but it did.\nOutput:\n%s", part, result)
				}
			}
		})
	}
}
