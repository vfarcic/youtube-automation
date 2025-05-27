package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	// No schema import needed here as we expect plain text.
)

// newLLMClientFuncForDescriptionTags is a variable to allow mocking in tests.
// It should be initialized to the actual openai.New function.
var newLLMClientFuncForDescriptionTags = func(options ...openai.Option) (llms.Model, error) {
	return openai.New(options...)
}

// SuggestDescriptionTags generates a space-separated string of exactly three tags,
// each starting with '#', based on the provided manuscript.
func SuggestDescriptionTags(ctx context.Context, manuscript string, config AITitleGeneratorConfig) (string, error) {
	if strings.TrimSpace(manuscript) == "" {
		return "", errors.New("manuscript content is empty, cannot suggest description tags")
	}
	if !isAIConfigComplete(config) {
		return "", errors.New("AI configuration (Endpoint, DeploymentName, APIKey, APIVersion) is not fully set for description tags")
	}

	llm, err := newLLMClientFuncForDescriptionTags(
		openai.WithAPIType(openai.APITypeAzure),
		openai.WithToken(config.APIKey),
		openai.WithBaseURL(config.Endpoint),
		openai.WithModel(config.DeploymentName),
		openai.WithAPIVersion(config.APIVersion),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create LangChainGo client for description tags: %w", err)
	}

	prompt := fmt.Sprintf(
		"Given the following manuscript, suggest exactly three relevant tags for a video description. "+
			"Each tag MUST start with a '#' character and all three tags MUST be separated by a single space. "+
			"Do not add any other text, explanation, or formatting. "+
			"Example response: #keyword1 #anotherkeyword #thirdtag\n\nManuscript:\n%s",
		manuscript,
	)

	var responseContent string
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		llmResponse, err := llms.GenerateFromSinglePrompt(ctx, llm, prompt)
		if err != nil {
			if attempt == maxRetries {
				return "", fmt.Errorf("error generating description tags after %d attempts: %w", maxRetries, err)
			}
			time.Sleep(time.Duration(attempt) * time.Second) // Exponential backoff
			continue
		}
		responseContent = strings.TrimSpace(llmResponse)
		break
	}

	if responseContent == "" {
		return "", errors.New("AI returned an empty response for description tags")
	}

	// Basic validation: count '#' and spaces to infer tag count.
	// This is a simple check, more robust validation could be added if needed.
	if strings.Count(responseContent, "#") != 3 || strings.Count(responseContent, " ") != 2 {
		// Potentially retry or return an error if format is critical.
		// For now, we'll return what we got but log a warning or let user correct.
		// Or, we could return an error to force AI to try again (if retry loop is modified).
		// Let's return an error to indicate the AI didn't follow instructions.
		return "", fmt.Errorf("AI did not return exactly three space-separated tags starting with '#'. Got: %s", responseContent)
	}

	// No specific length truncation needed here as the request is for 3 short tags.

	return responseContent, nil
}

// isAIConfigComplete checks if all necessary fields in AITitleGeneratorConfig are set.
// This function can be shared or moved to a common place if used by multiple AI services.
// For now, keeping it simple. Duplicated from titles.go for now.
func isAIConfigComplete(config AITitleGeneratorConfig) bool {
	return config.Endpoint != "" &&
		config.DeploymentName != "" &&
		config.APIKey != "" &&
		config.APIVersion != ""
}
