package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

// AIHighlightResponse matches the expected JSON structure from the AI for highlights.
// It might be { "suggested_highlights": ["phrase1", "phrase2"] }
// or directly ["phrase1", "phrase2"]. The code will try to handle both.
type AIHighlightResponse struct {
	SuggestedHighlights []string `json:"suggested_highlights"`
}

// SuggestHighlights contacts Azure OpenAI to get suggestions for words or phrases to highlight in a manuscript.
// It expects the AI to return a JSON array of strings, potentially wrapped in an object.
func SuggestHighlights(ctx context.Context, manuscriptContent string, aiConfig AITitleGeneratorConfig) ([]string, error) {
	if aiConfig.Endpoint == "" || aiConfig.DeploymentName == "" || aiConfig.APIKey == "" || aiConfig.APIVersion == "" {
		return nil, fmt.Errorf("AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set")
	}

	baseURL := strings.TrimSuffix(aiConfig.Endpoint, "/")

	llm, err := openai.New(
		openai.WithToken(aiConfig.APIKey),
		openai.WithBaseURL(baseURL),
		openai.WithModel(aiConfig.DeploymentName),
		openai.WithAPIVersion(aiConfig.APIVersion),
		openai.WithAPIType(openai.APITypeAzure),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure OpenAI client: %w", err)
	}

	prompt := fmt.Sprintf(
		"Analyze the following video manuscript. Identify specific words or short, exact phrases (2-5 words) that are excellent candidates for highlighting (e.g., making bold in Markdown). "+
			"These should be key terms, commands, important concepts, or impactful statements. "+
			"Do not suggest phrases that are already part of Markdown headings (lines starting with #). "+
			"Do not suggest entire sentences. Focus on concise, impactful selections. "+
			"Return your suggestions as a JSON array of strings. Each string in the array should be an exact quote from the manuscript. "+
			"Example JSON output: [\"exact phrase from manuscript\", \"another key term\"]\n\n"+
			"MANUSCRIPT:\n%s\n\nSUGGESTED HIGHLIGHTS (JSON ARRAY OF STRINGS):",
		manuscriptContent,
	)

	var responseContent string
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		completion, genErr := llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			prompt,
			llms.WithTemperature(0.5), // Lower temperature for more precise extraction
			llms.WithMaxTokens(500),   // Allow for a decent number of suggestions
			llms.WithJSONMode(),       // Request JSON output
		)
		if genErr != nil {
			fmt.Fprintf(os.Stderr, "Error generating highlights (attempt %d/%d): %v\n", i+1, maxRetries, genErr)
			if i == maxRetries-1 {
				return nil, fmt.Errorf("failed to generate highlights after %d attempts: %w", maxRetries, genErr)
			}
			continue // Retry
		}
		responseContent = completion
		break // Success
	}

	if responseContent == "" {
		return nil, fmt.Errorf("AI returned an empty response for highlights")
	}

	// Attempt to strip Markdown code fences if present, as seen with title suggestions
	cleanedResponse := strings.TrimSpace(responseContent)
	if strings.HasPrefix(cleanedResponse, "```json") {
		cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
		cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
		cleanedResponse = strings.TrimSpace(cleanedResponse)
	}

	var highlights []string
	// First, try to unmarshal as a direct array of strings
	errDirect := json.Unmarshal([]byte(cleanedResponse), &highlights)
	if errDirect == nil {
		return highlights, nil // Successfully unmarshalled as a direct array
	}

	// If direct unmarshal failed, try unmarshalling as an object { "suggested_highlights": [...] }
	var responseObj AIHighlightResponse
	if errObj := json.Unmarshal([]byte(cleanedResponse), &responseObj); errObj != nil {
		// If both failed, return a combined error message or the more specific one if possible.
		return nil, fmt.Errorf(
			"failed to unmarshal highlights JSON from AI response. Direct array error: %v. Object error: %v. Response: %s",
			errDirect, // Include the error from the direct attempt
			errObj,    // Include the error from the object attempt
			cleanedResponse,
		)
	}

	// If unmarshalling as an object succeeded, use its content.
	return responseObj.SuggestedHighlights, nil
}
