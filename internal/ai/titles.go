package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai" // Using LangChainGo's OpenAI client for Azure

	"devopstoolkit/youtube-automation/internal/configuration"
)

// SuggestedTitle is no longer a struct, AI will return a simple list of strings.
// type SuggestedTitle struct {
// 	Title       string `json:"title"`
// 	Explanation string `json:"explanation"` // This will be removed
// }

// AITitleGeneratorConfig holds the necessary configuration.
// For LangChainGo, APIKey, Endpoint, and DeploymentName (as Model) are key.
type AITitleGeneratorConfig struct {
	Endpoint       string
	DeploymentName string // This will be used as the Model name for LangChainGo
	APIKey         string
	APIVersion     string
}

// SuggestTitles contacts Azure OpenAI via LangChainGo to get title suggestions.
// It now expects the AI to return a simple JSON array of strings.
func SuggestTitles(ctx context.Context, manuscriptContent string, aiConfig AITitleGeneratorConfig) ([]string, error) { // Return type changed to []string
	if aiConfig.APIKey == "" {
		return nil, fmt.Errorf("Azure OpenAI API key is not configured")
	}
	if aiConfig.Endpoint == "" {
		return nil, fmt.Errorf("Azure OpenAI Endpoint is not configured")
	}
	if aiConfig.DeploymentName == "" {
		return nil, fmt.Errorf("Azure OpenAI DeploymentName (model) is not configured")
	}
	if aiConfig.APIVersion == "" {
		// Default to a known working version if not set, but prefer explicit configuration
		aiConfig.APIVersion = "2023-07-01-preview" // Defaulting based on curl tests
		fmt.Fprintf(os.Stderr, "Warning: Azure OpenAI APIVersion not set in config, defaulting to %s\n", aiConfig.APIVersion)
	}

	// Use the root Azure OpenAI endpoint as the BaseURL.
	// LangChainGo's Azure APIType setting should handle the rest of the path construction.
	azureBaseURL := aiConfig.Endpoint

	llm, err := openai.New(
		openai.WithModel(aiConfig.DeploymentName), // Still pass the model; Azure might need it in the body too
		openai.WithAPIType(openai.APITypeAzure),
		openai.WithToken(aiConfig.APIKey),
		openai.WithBaseURL(azureBaseURL), // Use the constructed Azure-specific base URL
		openai.WithAPIVersion(aiConfig.APIVersion),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LangChainGo client: %w", err)
	}

	// Updated prompt to request only a JSON array of strings (titles)
	systemMessage := fmt.Sprintf(`You are an expert YouTube title generator. Your task is to generate 5 compelling and SEO-friendly titles for a video based on the provided manuscript. Each title MUST be 70 characters or less. Return the output as a simple JSON array of strings. For example: ["Example Title 1 (Max 70 Chars)", "Example Title 2 (Max 70 Chars)", ...]

Video Manuscript:
%s`, manuscriptContent)

	// Using GenerateFromSinglePrompt for simplicity, assuming the model understands the JSON instruction well.
	// For more complex chat interactions or explicit message roles, llm.Call or llm.Generate could be used.
	// We need to ensure the model correctly returns JSON in the content of its response.
	responseContent, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		systemMessage,
		llms.WithTemperature(0.7), // Adjust creativity
		llms.WithMaxTokens(512),   // Reduced max tokens as we only need titles
		// It's crucial the LLM is prompted to return JSON. LangchainGo doesn't have a direct 'ResponseFormat: JSON' like the raw SDK for all models.
		// The instruction to return JSON is in the systemMessage.
	)

	if err != nil {
		return nil, fmt.Errorf("LangChainGo title generation failed: %w", err)
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

	var titles []string // Changed from []SuggestedTitle to []string
	if err := json.Unmarshal([]byte(cleanedResponse), &titles); err != nil {
		// Log the raw response for debugging if JSON parsing fails
		fmt.Fprintf(os.Stderr, "Failed to parse JSON response (expected array of strings) from AI. Cleaned response attempt: %s\nRaw response: %s\n", cleanedResponse, responseContent)
		return nil, fmt.Errorf("failed to parse JSON response from AI (expected array of strings): %w. Cleaned response attempt: %s", err, cleanedResponse)
	}

	return titles, nil // Return []string
}

// GetAIConfig retrieves AI configuration from global settings.
func GetAIConfig() (AITitleGeneratorConfig, error) {
	apiKey := os.Getenv("AI_KEY") // Or your specific env var name
	if apiKey == "" {
		// Fallback or error if not in env, though settings should provide it
		if configuration.GlobalSettings.AI.Key != "" {
			apiKey = configuration.GlobalSettings.AI.Key
		} else {
			return AITitleGeneratorConfig{}, fmt.Errorf("AI_KEY environment variable not set and no key in settings")
		}
	}

	if configuration.GlobalSettings.AI.Endpoint == "" || configuration.GlobalSettings.AI.Deployment == "" {
		return AITitleGeneratorConfig{}, fmt.Errorf("AI endpoint or deployment not configured in settings.yaml")
	}

	return AITitleGeneratorConfig{
		Endpoint:       configuration.GlobalSettings.AI.Endpoint,
		DeploymentName: configuration.GlobalSettings.AI.Deployment,
		APIKey:         apiKey,
		APIVersion:     configuration.GlobalSettings.AI.APIVersion,
	}, nil
}
