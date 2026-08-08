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
	// GenerateContent generates content with extended (adaptive) thinking
	// enabled for higher-quality output. On Anthropic (Sonnet 5+) this runs
	// adaptive thinking; providers without a thinking concept ignore it.
	// maxTokens is a ceiling covering thinking + output, so callers should
	// budget generously.
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
		model = "claude-sonnet-5"
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

// GenerateContent for Anthropic. The thinking field is left unset, which enables
// adaptive thinking by default on Sonnet 5+ (the vendored SDK predates the
// explicit adaptive-thinking parameter), trading extra tokens and latency for
// higher-quality output. Thinking tokens count against max_tokens, so callers
// budget generously.
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

	return extractAnthropicText(message.Content)
}

// extractAnthropicText returns the concatenated text from all text blocks in the
// response, skipping thinking blocks. When thinking is enabled the response
// leads with thinking block(s) that carry no text content, so reading only the
// first block is not safe.
func extractAnthropicText(content []anthropic.ContentBlockUnion) (string, error) {
	if len(content) == 0 {
		return "", fmt.Errorf("Anthropic returned empty response")
	}

	var sb strings.Builder
	for _, block := range content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}

	text := strings.TrimSpace(sb.String())
	if text == "" {
		return "", fmt.Errorf("Anthropic response contains no text content")
	}
	return text, nil
}