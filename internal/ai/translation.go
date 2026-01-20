package ai

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates/translate_metadata.md
var translateMetadataTemplate string

// VideoMetadataInput holds the input fields for translation.
type VideoMetadataInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        string   `json:"tags"`
	Timecodes   string   `json:"timecodes"`
	ShortTitles []string `json:"shortTitles,omitempty"` // Titles of YouTube Shorts to translate
}

// VideoMetadataOutput holds the translated fields.
type VideoMetadataOutput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Tags        string   `json:"tags"`
	Timecodes   string   `json:"timecodes"`
	ShortTitles []string `json:"shortTitles,omitempty"` // Translated Short titles (same order as input)
}

// translateMetadataTemplateData holds the data for the translation template.
type translateMetadataTemplateData struct {
	TargetLanguage string
	InputJSON      string
}

// TranslateVideoMetadata translates video metadata (title, description, tags, timecodes)
// to the target language using the configured AI provider.
// It returns all translated fields in a single API call for consistency and efficiency.
func TranslateVideoMetadata(ctx context.Context, input VideoMetadataInput, targetLanguage string) (*VideoMetadataOutput, error) {
	if targetLanguage == "" {
		return nil, fmt.Errorf("target language is required")
	}

	// Check if there's anything to translate
	if input.Title == "" && input.Description == "" && input.Tags == "" && input.Timecodes == "" && len(input.ShortTitles) == 0 {
		return nil, fmt.Errorf("at least one field (title, description, tags, timecodes, or shortTitles) must be provided")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	// Marshal input to JSON (handles escaping of quotes, newlines, etc.)
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input to JSON: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New("translate_metadata").Parse(translateMetadataTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse translation template: %w", err)
	}

	data := translateMetadataTemplateData{
		TargetLanguage: targetLanguage,
		InputJSON:      string(inputJSON),
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute translation template: %w", err)
	}

	prompt := promptBuf.String()

	// Generate content using the provider
	// Use higher max tokens since we're translating multiple fields
	responseContent, err := provider.GenerateContent(ctx, prompt, 2048)
	if err != nil {
		return nil, fmt.Errorf("AI translation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return nil, fmt.Errorf("AI returned an empty response for translation")
	}

	// Strip Markdown code fences if present
	cleanedResponse := stripCodeFences(responseContent)

	// Parse the JSON response
	var output VideoMetadataOutput
	if err := json.Unmarshal([]byte(cleanedResponse), &output); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from AI: %w. Response: %s", err, cleanedResponse)
	}

	return &output, nil
}

// stripCodeFences removes markdown code fences from AI responses.
func stripCodeFences(response string) string {
	// Trim leading/trailing whitespace first
	cleaned := strings.TrimSpace(response)

	// Remove ```json ... ``` fences
	if strings.HasPrefix(cleaned, "```json\n") && strings.HasSuffix(cleaned, "\n```") {
		cleaned = strings.TrimPrefix(cleaned, "```json\n")
		cleaned = strings.TrimSuffix(cleaned, "\n```")
	} else if strings.HasPrefix(cleaned, "```json") && strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimSuffix(cleaned, "```")
	} else if strings.HasPrefix(cleaned, "```\n") && strings.HasSuffix(cleaned, "\n```") {
		cleaned = strings.TrimPrefix(cleaned, "```\n")
		cleaned = strings.TrimSuffix(cleaned, "\n```")
	} else if strings.HasPrefix(cleaned, "```") && strings.HasSuffix(cleaned, "```") {
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
	}

	return strings.TrimSpace(cleaned)
}
