package ai

import (
	"context"
	"fmt"
	"strings"
)

// SuggestTags generates comma-separated tags using the configured AI provider.
// It expects the AI to return a single comma-separated string of tags,
// with a total character limit of 450.
func SuggestTags(ctx context.Context, manuscriptContent string, optionalConfig ...interface{}) (string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}
	if strings.TrimSpace(manuscriptContent) == "" {
		return "", fmt.Errorf("manuscript content is empty, cannot generate tags")
	}

	prompt := fmt.Sprintf(
		`Based on the following manuscript, generate a comma-separated list of relevant tags.
The total length of the comma-separated string of tags MUST NOT exceed 450 characters.
Provide ONLY the comma-separated string of tags, without any additional explanation, preamble, or markdown formatting.

Manuscript:
---
%s
---

Tags (comma-separated, max 450 chars):`,
		manuscriptContent,
	)

	responseContent, err := provider.GenerateContent(ctx, prompt, 200)
	if err != nil {
		return "", fmt.Errorf("AI tag generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty response for tags")
	}

	// Ensure the response does not exceed the character limit.
	// This is a safeguard, the AI should respect the prompt.
	if len(responseContent) > 450 {
		// Attempt to truncate intelligently by finding the last comma before 450 chars
		if idx := strings.LastIndex(responseContent[:450], ","); idx != -1 {
			responseContent = responseContent[:idx]
		} else {
			// If no comma, just hard truncate (less ideal)
			responseContent = responseContent[:450]
		}
	}

	return strings.TrimSpace(responseContent), nil
}