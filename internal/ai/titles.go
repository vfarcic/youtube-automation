package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SuggestedTitle is no longer a struct, AI will return a simple list of strings.
// type SuggestedTitle struct {
// 	Title       string `json:"title"`
// 	Explanation string `json:"explanation"` // This will be removed
// }

// SuggestTitles generates video title suggestions using the configured AI provider.
// It returns a simple JSON array of strings.
func SuggestTitles(ctx context.Context, manuscriptContent string, optionalConfig ...interface{}) ([]string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Create prompt for title generation
	prompt := fmt.Sprintf(`You are an expert YouTube title generator. Your task is to generate 5 compelling and SEO-friendly titles for a video based on the provided manuscript. Each title MUST be 70 characters or less. Return the output as a simple JSON array of strings. For example: ["Example Title 1 (Max 70 Chars)", "Example Title 2 (Max 70 Chars)", ...]

Video Manuscript:
%s`, manuscriptContent)

	// Generate content using the provider
	responseContent, err := provider.GenerateContent(ctx, prompt, 512)
	if err != nil {
		return nil, fmt.Errorf("AI title generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return nil, fmt.Errorf("AI returned an empty response for titles")
	}

	// Strip Markdown code fences if present
	cleanedResponse := responseContent
	if strings.HasPrefix(cleanedResponse, "```json\n") && strings.HasSuffix(cleanedResponse, "\n```") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json\n")
		cleanedResponse = strings.TrimSuffix(cleanedResponse, "\n```")
	} else if strings.HasPrefix(cleanedResponse, "```") && strings.HasSuffix(cleanedResponse, "```") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```")
		cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
	}

	var titles []string
	if err := json.Unmarshal([]byte(cleanedResponse), &titles); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from AI (expected array of strings): %w. Response: %s", err, cleanedResponse)
	}

	return titles, nil
}

// TEMPORARY: Compatibility function for old app module - returns empty struct
func GetAIConfig() (interface{}, error) {
	return struct{}{}, nil
}

