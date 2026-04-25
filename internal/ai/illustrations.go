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

//go:embed templates/illustrations.md
var illustrationsTemplate string

// illustrationsTemplateData holds the data for the illustrations template.
type illustrationsTemplateData struct {
	Manuscript string
	Tagline    string
}

// SuggestIllustrations generates illustration ideas for a thumbnail based on
// the video's manuscript and tagline using the configured text AI provider.
func SuggestIllustrations(ctx context.Context, manuscript, tagline string) ([]string, error) {
	if strings.TrimSpace(manuscript) == "" {
		return nil, fmt.Errorf("manuscript content is empty")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	tmpl, err := template.New("illustrations").Parse(illustrationsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse illustrations template: %w", err)
	}

	data := illustrationsTemplateData{
		Manuscript: manuscript,
		Tagline:    tagline,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute illustrations template: %w", err)
	}

	responseContent, err := provider.GenerateContent(ctx, promptBuf.String(), 512)
	if err != nil {
		return nil, fmt.Errorf("AI illustration suggestion failed: %w", err)
	}

	return parseIllustrationsResponse(responseContent)
}

// parseIllustrationsResponse parses the AI response into a list of illustration ideas.
// It handles plain JSON arrays, JSON wrapped in markdown code fences (```json or ```),
// and AI responses that include explanatory text before/after the JSON array.
func parseIllustrationsResponse(text string) ([]string, error) {
	cleanedText := text

	// Handle markdown code fences (```json ... ``` or ``` ... ```)
	if strings.Contains(cleanedText, "```json") {
		parts := strings.Split(cleanedText, "```json")
		if len(parts) > 1 {
			cleanedText = parts[1]
			if strings.Contains(cleanedText, "```") {
				cleanedText = strings.Split(cleanedText, "```")[0]
			}
		}
	} else if strings.Contains(cleanedText, "```") {
		parts := strings.Split(cleanedText, "```")
		if len(parts) > 1 {
			cleanedText = parts[1]
			if strings.Contains(cleanedText, "```") {
				cleanedText = strings.Split(cleanedText, "```")[0]
			}
		}
	}

	cleanedText = strings.TrimSpace(cleanedText)

	// Try direct parse first
	var illustrations []string
	if err := json.Unmarshal([]byte(cleanedText), &illustrations); err == nil {
		if len(illustrations) == 0 {
			return nil, fmt.Errorf("AI returned an empty list of illustrations")
		}
		return illustrations, nil
	}

	// Fallback: extract a JSON array from mixed text by finding the first '[' and last ']'
	startIdx := strings.Index(cleanedText, "[")
	endIdx := strings.LastIndex(cleanedText, "]")
	if startIdx >= 0 && endIdx > startIdx {
		candidate := cleanedText[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(candidate), &illustrations); err == nil {
			if len(illustrations) == 0 {
				return nil, fmt.Errorf("AI returned an empty list of illustrations")
			}
			return illustrations, nil
		}
	}

	return nil, fmt.Errorf("failed to parse JSON response from AI output")
}
