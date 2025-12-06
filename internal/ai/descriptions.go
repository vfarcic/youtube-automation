package ai

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/description.md
var descriptionTemplate string

// descriptionTemplateData holds the data for the description template
type descriptionTemplateData struct {
	Content string
}

// SuggestDescription generates video description using the configured AI provider.
// It expects the AI to return a single plain text string.
func SuggestDescription(ctx context.Context, manuscriptContent string, optionalConfig ...interface{}) (string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Parse and execute template for description generation prompt
	tmpl, err := template.New("description").Parse(descriptionTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse description template: %w", err)
	}

	data := descriptionTemplateData{
		Content: manuscriptContent,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return "", fmt.Errorf("failed to execute description template: %w", err)
	}

	prompt := promptBuf.String()

	responseContent, err := provider.GenerateContent(ctx, prompt, 400)
	if err != nil {
		return "", fmt.Errorf("AI description generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return "", fmt.Errorf("AI returned an empty description")
	}

	return strings.TrimSpace(responseContent), nil
}
