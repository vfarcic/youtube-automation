package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/tags.md
var tagsTemplate string

// tagsTemplateData holds the data for the tags template
type tagsTemplateData struct {
	Content string
}

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

	// Parse and execute template for tags generation prompt
	tmpl, err := template.New("tags").Parse(tagsTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse tags template: %w", err)
	}

	data := tagsTemplateData{
		Content: manuscriptContent,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute tags template: %w", err)
	}

	prompt := promptBuf.String()

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
