package ai

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"

	"devopstoolkit/youtube-automation/internal/configuration"
)

// AIProvider interface for different AI providers
type AIProvider interface {
	GenerateContent(ctx context.Context, prompt string, maxTokens int) (string, error)
}

// AzureProvider implements AIProvider for Azure OpenAI
type AzureProvider struct {
	client llms.Model
}

// AnthropicProvider implements AIProvider for Anthropic
type AnthropicProvider struct {
	client anthropic.Client
	model  string
}

// GetAIProvider creates the appropriate AI provider based on configuration
var GetAIProvider = func() (AIProvider, error) {
	switch configuration.GlobalSettings.AI.Provider {
	case "azure":
		return createAzureProvider()
	case "anthropic":
		return createAnthropicProvider()
	default:
		return nil, fmt.Errorf("unsupported AI provider: %s", configuration.GlobalSettings.AI.Provider)
	}
}

func createAzureProvider() (*AzureProvider, error) {
	config := configuration.GlobalSettings.AI.Azure
	
	// Get API key from environment or config
	apiKey := os.Getenv("AI_KEY")
	if apiKey == "" && config.Key != "" {
		apiKey = config.Key
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Azure OpenAI API key not configured")
	}

	if config.Endpoint == "" || config.Deployment == "" {
		return nil, fmt.Errorf("Azure OpenAI endpoint or deployment not configured")
	}

	// Default API version if not set
	apiVersion := config.APIVersion
	if apiVersion == "" {
		apiVersion = "2023-05-15"
	}

	baseURL := strings.TrimSuffix(config.Endpoint, "/")

	llm, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithBaseURL(baseURL),
		openai.WithModel(config.Deployment),
		openai.WithAPIVersion(apiVersion),
		openai.WithAPIType(openai.APITypeAzure),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure OpenAI client: %w", err)
	}

	return &AzureProvider{client: llm}, nil
}

func createAnthropicProvider() (*AnthropicProvider, error) {
	config := configuration.GlobalSettings.AI.Anthropic
	
	// Get API key from environment or config
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" && config.Key != "" {
		apiKey = config.Key
	}
	if apiKey == "" {
		return nil, fmt.Errorf("Anthropic API key not configured")
	}

	model := config.Model
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	client := anthropic.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &AnthropicProvider{
		client: client,
		model:  model,
	}, nil
}

// GenerateContent for Azure OpenAI
func (a *AzureProvider) GenerateContent(ctx context.Context, prompt string, maxTokens int) (string, error) {
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		a.client,
		prompt,
		llms.WithTemperature(0.7),
		llms.WithMaxTokens(maxTokens),
	)
	if err != nil {
		return "", fmt.Errorf("Azure OpenAI generation failed: %w", err)
	}
	
	return strings.TrimSpace(completion), nil
}

// GenerateContent for Anthropic
func (a *AnthropicProvider) GenerateContent(ctx context.Context, prompt string, maxTokens int) (string, error) {
	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(a.model),
		MaxTokens: int64(maxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	})
	if err != nil {
		return "", fmt.Errorf("Anthropic generation failed: %w", err)
	}

	if len(message.Content) == 0 {
		return "", fmt.Errorf("Anthropic returned empty response")
	}

	// Extract text from the first content block
	if len(message.Content) > 0 && message.Content[0].Text != "" {
		return strings.TrimSpace(message.Content[0].Text), nil
	}

	return "", fmt.Errorf("Anthropic response contains no text content")
}