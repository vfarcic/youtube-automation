package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// SuggestTweets generates 5 tweet suggestions based on the provided manuscript.
// Each tweet should be a maximum of 280 characters.
func SuggestTweets(ctx context.Context, manuscript string, optionalConfig ...interface{}) ([]string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	if strings.TrimSpace(manuscript) == "" {
		return nil, fmt.Errorf("manuscript content is empty, cannot suggest tweets")
	}

	prompt := fmt.Sprintf(`Based on the following video manuscript, generate 5 engaging tweet suggestions to promote the video. Each tweet MUST be 280 characters or less. Return the output as a simple JSON array of strings.

Video Manuscript:
%s

Response format: ["Tweet 1 (max 280 chars)", "Tweet 2 (max 280 chars)", ...]`, manuscript)

	responseContent, err := provider.GenerateContent(ctx, prompt, 400)
	if err != nil {
		return nil, fmt.Errorf("AI tweet generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return nil, fmt.Errorf("AI returned an empty response for tweets")
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

	var tweets []string
	if err := json.Unmarshal([]byte(cleanedResponse), &tweets); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from AI (expected array of strings): %w. Response: %s", err, cleanedResponse)
	}

	return tweets, nil
}