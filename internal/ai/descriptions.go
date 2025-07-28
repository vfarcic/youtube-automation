package ai

import (
	"context"
	"fmt"
	"strings"
)

// SuggestDescription generates video description using the configured AI provider.
// It expects the AI to return a single plain text string.
func SuggestDescription(ctx context.Context, manuscriptContent string, optionalConfig ...interface{}) (string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	prompt := fmt.Sprintf(
		"Based on the following video manuscript, generate a concise and engaging video description. "+
			"The description should be one or two paragraphs long. "+
			"Return only the description text, with no additional formatting or commentary.\n\nMANUSCRIPT:\n%s\n\nDESCRIPTION:",
		manuscriptContent,
	)

	responseContent, err := provider.GenerateContent(ctx, prompt, 400)
	if err != nil {
		return "", fmt.Errorf("AI description generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty description")
	}

	return strings.TrimSpace(responseContent), nil
}
