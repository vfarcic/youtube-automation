package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

//nolint:lll
var newLLMClientFuncForTags = func(options ...openai.Option) (llms.Model, error) {
	return openai.New(options...)
}

// SuggestTags contacts Azure OpenAI via LangChainGo to get tag suggestions.
// It expects the AI to return a single comma-separated string of tags,
// with a total character limit of 450.
func SuggestTags(ctx context.Context, manuscriptContent string, aiConfig AITitleGeneratorConfig) (string, error) {
	if aiConfig.Endpoint == "" || aiConfig.DeploymentName == "" || aiConfig.APIKey == "" || aiConfig.APIVersion == "" {
		return "", fmt.Errorf("AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set. Please check your settings.yaml or environment variables (AI_KEY)")
	}
	if strings.TrimSpace(manuscriptContent) == "" {
		return "", fmt.Errorf("manuscript content is empty, cannot generate tags")
	}

	baseURL := strings.TrimSuffix(aiConfig.Endpoint, "/")

	llm, err := newLLMClientFuncForTags(
		openai.WithAPIType(openai.APITypeAzure),
		openai.WithToken(aiConfig.APIKey),
		openai.WithBaseURL(baseURL),
		openai.WithAPIVersion(aiConfig.APIVersion),
		openai.WithModel(aiConfig.DeploymentName), // In Azure, Deployment Name often serves as the model identifier for the endpoint
	)
	if err != nil {
		return "", fmt.Errorf("failed to create LangChainGo client for tags: %w", err)
	}

	prompt := fmt.Sprintf(
		`Based on the following manuscript, generate a comma-separated list of relevant tags.
The total length of the comma-separated string of tags MUST NOT exceed 450 characters.
Provide ONLY the comma-separated string of tags, without any additional explanation, preamble, or markdown formatting.

Manuscript:
---
%s
---

Tags (comma-separated, max 450 chars):`,
		manuscriptContent,
	)

	var responseContent string
	maxRetries := 3
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryDelay)
			fmt.Fprintf(os.Stderr, "Retrying AI call for tags (%d/%d)...\n", i, maxRetries-1)
		}

		responseContent, err = llms.GenerateFromSinglePrompt(ctx, llm, prompt, llms.WithTemperature(0.7))
		if err == nil {
			break // Success
		}
		fmt.Fprintf(os.Stderr, "Attempt %d for tags failed: %v\n", i, err)
	}

	if err != nil {
		return "", fmt.Errorf("LangChainGo tag generation failed after %d retries: %w", maxRetries, err)
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
