package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

//nolint:lll
var newLLMClientFuncForDescriptions = func(options ...openai.Option) (llms.Model, error) {
	return openai.New(options...)
}

// SuggestDescription contacts Azure OpenAI via LangChainGo to get a description suggestion.
// It expects the AI to return a single plain text string.
func SuggestDescription(ctx context.Context, manuscriptContent string, aiConfig AITitleGeneratorConfig) (string, error) {
	if aiConfig.Endpoint == "" || aiConfig.DeploymentName == "" || aiConfig.APIKey == "" || aiConfig.APIVersion == "" {
		return "", fmt.Errorf("AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set")
	}

	// Ensure the base URL for Azure is just the endpoint, LangChainGo handles the rest.
	// Example: "https://your-resource.openai.azure.com/"
	baseURL := strings.TrimSuffix(aiConfig.Endpoint, "/")

	llm, err := newLLMClientFuncForDescriptions(
		openai.WithToken(aiConfig.APIKey),
		openai.WithBaseURL(baseURL),
		openai.WithModel(aiConfig.DeploymentName), // For Azure, model is the deployment name
		openai.WithAPIVersion(aiConfig.APIVersion),
		openai.WithAPIType(openai.APITypeAzure),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create Azure OpenAI client: %w", err)
	}

	prompt := fmt.Sprintf(
		"Based on the following video manuscript, generate a concise and engaging video description. "+
			"The description should be one or two paragraphs long. "+
			"Return only the description text, with no additional formatting or commentary.\n\nMANUSCRIPT:\n%s\n\nDESCRIPTION:",
		manuscriptContent,
	)

	var responseContent string
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		completion, err := llms.GenerateFromSinglePrompt(
			ctx,
			llm,
			prompt,
			llms.WithTemperature(0.7), // Adjust for creativity
			llms.WithMaxTokens(400),   // Enough for a couple of paragraphs
			// llms.WithJSONMode(), // Not using JSON mode, expect plain text
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating description (attempt %d/%d): %v\n", i+1, maxRetries, err)
			if i == maxRetries-1 {
				return "", fmt.Errorf("failed to generate description after %d attempts: %w", maxRetries, err)
			}
			continue // Retry
		}
		responseContent = strings.TrimSpace(completion)
		break // Success
	}

	// Unlike titles, we expect plain text directly, no JSON parsing or stripping of fences needed
	// unless the model consistently adds them anyway. For now, assume it respects the prompt.

	if responseContent == "" {
		return "", fmt.Errorf("AI returned an empty description")
	}

	return responseContent, nil
}
