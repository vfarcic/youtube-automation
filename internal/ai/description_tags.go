package ai

import (
	"context"
	"fmt"
	"strings"
)

// SuggestDescriptionTags generates a space-separated string of exactly three tags,
// each starting with '#', based on the provided manuscript.
func SuggestDescriptionTags(ctx context.Context, manuscript string, optionalConfig ...interface{}) (string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	if strings.TrimSpace(manuscript) == "" {
		return "", fmt.Errorf("manuscript content is empty, cannot suggest description tags")
	}

	prompt := fmt.Sprintf(`Based on the following manuscript, generate exactly 3 hashtags for a video description. Each hashtag should start with '#' and be relevant to the video content. Return them as a single line separated by spaces.

Manuscript:
%s

Response format: #hashtag1 #hashtag2 #hashtag3`, manuscript)

	responseContent, err := provider.GenerateContent(ctx, prompt, 100)
	if err != nil {
		return "", fmt.Errorf("AI description tags generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty response for description tags")
	}

	// Clean up the response - ensure tags start with # and are space-separated
	tags := strings.Fields(strings.TrimSpace(responseContent))
	var cleanedTags []string
	
	for _, tag := range tags {
		if !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}
		cleanedTags = append(cleanedTags, tag)
		if len(cleanedTags) >= 3 {
			break // Limit to 3 tags
		}
	}

	if len(cleanedTags) == 0 {
		return "", fmt.Errorf("AI did not return any valid tags")
	}

	return strings.Join(cleanedTags, " "), nil
}