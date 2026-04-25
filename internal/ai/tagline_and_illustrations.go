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

//go:embed templates/tagline_and_illustrations.md
var taglineAndIllustrationsTemplate string

// TaglineAndIllustrationsResult holds the AI-suggested taglines and illustrations.
type TaglineAndIllustrationsResult struct {
	Taglines      []string `json:"taglines"`
	Illustrations []string `json:"illustrations"`
}

type taglineAndIllustrationsTemplateData struct {
	Manuscript string
}

// SuggestTaglineAndIllustrations generates tagline options and illustration ideas
// for a thumbnail based on the video's manuscript using the configured text AI provider.
func SuggestTaglineAndIllustrations(ctx context.Context, manuscript string) (*TaglineAndIllustrationsResult, error) {
	if strings.TrimSpace(manuscript) == "" {
		return nil, fmt.Errorf("manuscript content is empty")
	}

	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	tmpl, err := template.New("tagline_and_illustrations").Parse(taglineAndIllustrationsTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse tagline and illustrations template: %w", err)
	}

	data := taglineAndIllustrationsTemplateData{
		Manuscript: manuscript,
	}

	var promptBuf bytes.Buffer
	if err := tmpl.Execute(&promptBuf, data); err != nil {
		return nil, fmt.Errorf("failed to execute tagline and illustrations template: %w", err)
	}

	responseContent, err := provider.GenerateContent(ctx, promptBuf.String(), 512)
	if err != nil {
		return nil, fmt.Errorf("AI tagline and illustration suggestion failed: %w", err)
	}

	return parseTaglineAndIllustrationsResponse(responseContent)
}

// parseTaglineAndIllustrationsResponse parses the AI response into taglines and illustrations.
// It handles plain JSON objects, JSON wrapped in markdown code fences, and AI responses
// that include explanatory text before/after the JSON.
func parseTaglineAndIllustrationsResponse(text string) (*TaglineAndIllustrationsResult, error) {
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
	var result TaglineAndIllustrationsResult
	if err := json.Unmarshal([]byte(cleanedText), &result); err == nil {
		if len(result.Taglines) == 0 {
			return nil, fmt.Errorf("AI returned an empty list of taglines")
		}
		if len(result.Illustrations) == 0 {
			return nil, fmt.Errorf("AI returned an empty list of illustrations")
		}
		return &result, nil
	}

	// Fallback: extract a JSON object from mixed text by finding the first '{' and last '}'
	startIdx := strings.Index(cleanedText, "{")
	endIdx := strings.LastIndex(cleanedText, "}")
	if startIdx >= 0 && endIdx > startIdx {
		candidate := cleanedText[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(candidate), &result); err == nil {
			if len(result.Taglines) == 0 {
				return nil, fmt.Errorf("AI returned an empty list of taglines")
			}
			if len(result.Illustrations) == 0 {
				return nil, fmt.Errorf("AI returned an empty list of illustrations")
			}
			return &result, nil
		}
	}

	return nil, fmt.Errorf("failed to parse JSON response from AI output")
}
