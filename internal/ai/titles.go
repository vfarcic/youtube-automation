package ai

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/template"
)

//go:embed templates/titles.md
var defaultTitlesTemplate string

// SuggestedTitle is no longer a struct, AI will return a simple list of strings.
// type SuggestedTitle struct {
// 	Title       string `json:"title"`
// 	Explanation string `json:"explanation"` // This will be removed
// }

// titlesTemplateData holds the data for the titles template
type titlesTemplateData struct {
	ManuscriptContent string
}

// SuggestTitles generates video title suggestions using the configured AI provider.
// It returns a simple JSON array of strings.
func SuggestTitles(ctx context.Context, manuscriptContent string, optionalConfig ...interface{}) ([]string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Load titles.md from working directory (user-owned, editable template)
	titlesTemplate, err := LoadTitlesTemplate()
	if err != nil {
		return nil, err
	}

	// Parse and execute template for title generation prompt
	tmpl, err := template.New("titles").Parse(titlesTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse titles template: %w", err)
	}

	data := titlesTemplateData{
		ManuscriptContent: manuscriptContent,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute titles template: %w", err)
	}

	prompt := promptBuf.String()

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
	cleanedResponse = strings.TrimSpace(cleanedResponse)

	var titles []string
	if err := json.Unmarshal([]byte(cleanedResponse), &titles); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from AI (expected array of strings): %w. Response: %s", err, cleanedResponse)
	}

	return titles, nil
}

// LoadTitlesTemplate reads titles.md from the working directory.
// Returns an error with instructions if the file doesn't exist.
func LoadTitlesTemplate() (string, error) {
	content, err := os.ReadFile("titles.md")
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf(
				"titles.md not found in the current directory.\n\n"+
					"To create it, either:\n"+
					"  1. Run 'Analyze → Titles' to generate one from your A/B test data\n"+
					"  2. Create titles.md manually with the following default content:\n\n"+
					"--- START default titles.md ---\n%s\n--- END default titles.md ---",
				defaultTitlesTemplate,
			)
		}
		return "", fmt.Errorf("failed to read titles.md: %w", err)
	}
	return string(content), nil
}

// TEMPORARY: Compatibility function for old app module - returns empty struct
func GetAIConfig() (interface{}, error) {
	return struct{}{}, nil
}

