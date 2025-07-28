package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// AIHighlightResponse matches the expected JSON structure from the AI for highlights.
// It might be { "suggested_highlights": ["phrase1", "phrase2"] }
// or directly ["phrase1", "phrase2"]. The code will try to handle both.
type AIHighlightResponse struct {
	SuggestedHighlights []string `json:"suggested_highlights"`
}

// SuggestHighlights generates suggestions for words or phrases to highlight in a manuscript.
// It expects the AI to return a JSON array of strings, potentially wrapped in an object.
func SuggestHighlights(ctx context.Context, manuscriptContent string, optionalConfig ...interface{}) ([]string, error) {
	provider, err := GetAIProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create AI provider: %w", err)
	}

	prompt := fmt.Sprintf(`Based on the following manuscript, identify 5-10 key words or phrases that should be highlighted to emphasize the most important concepts. Return the output as a simple JSON array of strings.

Manuscript:
%s

Response format: ["key phrase 1", "important term 2", ...]`, manuscriptContent)

	responseContent, err := provider.GenerateContent(ctx, prompt, 300)
	if err != nil {
		return nil, fmt.Errorf("AI highlight generation failed: %w", err)
	}

	if strings.TrimSpace(responseContent) == "" {
		return nil, fmt.Errorf("AI returned an empty response for highlights")
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

	// Try to parse as direct array first
	var highlights []string
	if err := json.Unmarshal([]byte(cleanedResponse), &highlights); err == nil {
		return highlights, nil
	}

	// If that fails, try to parse as wrapped object
	var response AIHighlightResponse
	if err := json.Unmarshal([]byte(cleanedResponse), &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response from AI: %w. Response: %s", err, cleanedResponse)
	}

	return response.SuggestedHighlights, nil
}