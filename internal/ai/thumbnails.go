package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	_ "embed"

	"github.com/anthropics/anthropic-sdk-go"
	constant "github.com/anthropics/anthropic-sdk-go/shared/constant"
)

//go:embed templates/thumbnail_variations.md
var thumbnailVariationsPrompt string

// VariationPrompts contains the two generated prompts for thumbnail variations
type VariationPrompts struct {
	Subtle string `json:"subtle_prompt"`
	Bold string `json:"bold_prompt"`
}

// GenerateThumbnailVariations analyzes an image and returns prompts for variations
func GenerateThumbnailVariations(ctx context.Context, imagePath string) (VariationPrompts, error) {
	// Read and encode image
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return VariationPrompts{}, fmt.Errorf("failed to read image file: %w", err)
	}
	
	// Detect media type
	ext := strings.ToLower(filepath.Ext(imagePath))
	mediaType := "image/jpeg"
	if ext == ".png" {
		mediaType = "image/png"
	} else if ext == ".webp" {
		mediaType = "image/webp"
	}

	encodedImage := base64.StdEncoding.EncodeToString(imageData)

	// Get AI provider
	provider, err := GetAIProvider()
	if err != nil {
		return VariationPrompts{}, err
	}

	// Check if provider supports vision (currently only implementing for Anthropic)
	anthropicProvider, ok := provider.(*AnthropicProvider)
	if !ok {
		return VariationPrompts{}, fmt.Errorf("vision analysis currently only supported with Anthropic provider")
	}

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(
			anthropic.NewTextBlock("Analyze this YouTube thumbnail and generate the two variation prompts."),
			anthropic.NewImageBlockBase64(mediaType, encodedImage),
		),
	}

	// Call Anthropic API directly since the generic GenerateContent doesn't support images yet
	resp, err := anthropicProvider.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(anthropicProvider.model),
		MaxTokens: int64(1024),
		System:    []anthropic.TextBlockParam{{Text: thumbnailVariationsPrompt, Type: constant.Text("text")}},
		Messages:  messages,
	})
	if err != nil {
		return VariationPrompts{}, fmt.Errorf("failed to analyze image: %w", err)
	}

	if len(resp.Content) == 0 || resp.Content[0].Text == "" {
		return VariationPrompts{}, fmt.Errorf("empty response from AI")
	}

	responseText := resp.Content[0].Text
	return parseVariationResponse(responseText)
}

func parseVariationResponse(text string) (VariationPrompts, error) {
	var prompts VariationPrompts
	
	// Attempt to clean the response if it contains markdown code blocks
	cleanedText := text
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
		}
	}
	
	cleanedText = strings.TrimSpace(cleanedText)

	err := json.Unmarshal([]byte(cleanedText), &prompts)
	if err != nil {
		return prompts, fmt.Errorf("failed to parse JSON response: %w. Raw response: %s", err, text)
	}

	if prompts.Subtle == "" || prompts.Bold == "" {
		return prompts, fmt.Errorf("JSON response missing required fields. Parsed object: %+v", prompts)
	}

	return prompts, nil
}
