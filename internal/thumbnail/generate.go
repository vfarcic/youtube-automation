package thumbnail

import "context"

// ImageGenerator is the interface for AI image generation providers.
// Each provider (Gemini, GPT Image, etc.) implements this to generate
// thumbnail images from a text prompt and reference photos.
type ImageGenerator interface {
	// Name returns the provider name (e.g., "gemini", "gpt-image").
	Name() string

	// GenerateImage sends a prompt and reference photos to the provider
	// and returns the generated image bytes.
	GenerateImage(ctx context.Context, prompt string, photos [][]byte) ([]byte, error)
}
