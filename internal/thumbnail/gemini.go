// Package thumbnail provides AI-powered thumbnail localization using Google Gemini.
package thumbnail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"devopstoolkit/youtube-automation/internal/configuration"
)

// Default Gemini API base URL
const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta"

// Errors returned by the Gemini client
var (
	ErrAPIKeyNotSet     = errors.New("GEMINI_API_KEY environment variable not set")
	ErrImageReadFailed  = errors.New("failed to read image file")
	ErrAPIRequestFailed = errors.New("Gemini API request failed")
	ErrNoImageGenerated = errors.New("no image generated in response")
	ErrUnsupportedLang  = errors.New("unsupported target language")
)

// languageNames maps language codes to full language names for prompts
var languageNames = map[string]string{
	"es": "Spanish",
	"pt": "Portuguese",
	"de": "German",
	"fr": "French",
	"it": "Italian",
	"ja": "Japanese",
	"ko": "Korean",
	"zh": "Chinese",
}

// Config holds Gemini API configuration
type Config struct {
	APIKey string
	Model  string // e.g., "gemini-3-pro-image-preview" or "gemini-2.5-flash-image"
}

// Client is the Google Gemini API client for image generation
type Client struct {
	config     Config
	httpClient *http.Client
	baseURL    string
}

// geminiRequest is the request structure for Gemini generateContent API
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string           `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"` // base64 encoded
}

type geminiGenerationConfig struct {
	ResponseModalities []string `json:"responseModalities"`
}

// geminiResponse is the response structure from Gemini API
type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewClient creates a new Gemini client for production use.
// It reads the API key from GEMINI_API_KEY environment variable
// and the model from configuration settings.
func NewClient() (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, ErrAPIKeyNotSet
	}

	model := configuration.GlobalSettings.Gemini.Model
	if model == "" {
		model = "gemini-3-pro-image-preview"
	}

	return &Client{
		config: Config{
			APIKey: apiKey,
			Model:  model,
		},
		httpClient: &http.Client{},
		baseURL:    defaultBaseURL,
	}, nil
}

// NewClientWithHTTPClient creates a client with injectable HTTP client and base URL.
// This is primarily used for testing with mock HTTP servers.
func NewClientWithHTTPClient(config Config, httpClient *http.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		config:     config,
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// GenerateLocalizedThumbnail generates a thumbnail with translated text.
// It takes the original image path, English tagline, and target language code.
// The Gemini model handles both translation and image generation in a single call.
// Returns the generated image bytes.
func (c *Client) GenerateLocalizedThumbnail(ctx context.Context, imagePath, tagline, targetLang string) ([]byte, error) {
	// Validate language
	langName, ok := languageNames[targetLang]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLang, targetLang)
	}

	// Read and encode the image
	imageData, mimeType, err := encodeImageToBase64(imagePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrImageReadFailed, err)
	}

	// Build the prompt
	prompt := buildPrompt(tagline, langName)

	// Build the request
	req := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
					{
						InlineData: &geminiInlineData{
							MimeType: mimeType,
							Data:     imageData,
						},
					},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseModalities: []string{"TEXT", "IMAGE"},
		},
	}

	// Make the API request
	respBody, err := c.doRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var resp geminiResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return nil, fmt.Errorf("%w: failed to parse response: %v", ErrAPIRequestFailed, err)
	}

	// Check for API error
	if resp.Error != nil {
		return nil, fmt.Errorf("%w: %s (code: %d)", ErrAPIRequestFailed, resp.Error.Message, resp.Error.Code)
	}

	// Extract the generated image
	imageBytes, err := extractImageFromResponse(&resp)
	if err != nil {
		return nil, err
	}

	return imageBytes, nil
}

// doRequest makes the HTTP request to the Gemini API
func (c *Client) doRequest(ctx context.Context, reqBody geminiRequest) ([]byte, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to marshal request: %v", ErrAPIRequestFailed, err)
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.config.Model, c.config.APIKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("%w: failed to create request: %v", ErrAPIRequestFailed, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrAPIRequestFailed, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read response: %v", ErrAPIRequestFailed, err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error message from response
		var errResp geminiResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != nil {
			return nil, fmt.Errorf("%w: %s (code: %d)", ErrAPIRequestFailed, errResp.Error.Message, errResp.Error.Code)
		}
		return nil, fmt.Errorf("%w: HTTP %d: %s", ErrAPIRequestFailed, resp.StatusCode, string(body))
	}

	return body, nil
}

// extractImageFromResponse extracts the base64 image data from the API response
func extractImageFromResponse(resp *geminiResponse) ([]byte, error) {
	if len(resp.Candidates) == 0 {
		return nil, ErrNoImageGenerated
	}

	for _, candidate := range resp.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && strings.HasPrefix(part.InlineData.MimeType, "image/") {
				imageBytes, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
				if err != nil {
					return nil, fmt.Errorf("%w: failed to decode image: %v", ErrNoImageGenerated, err)
				}
				return imageBytes, nil
			}
		}
	}

	return nil, ErrNoImageGenerated
}

// encodeImageToBase64 reads an image file and returns its base64 encoding and MIME type
func encodeImageToBase64(imagePath string) (string, string, error) {
	data, err := os.ReadFile(imagePath)
	if err != nil {
		return "", "", err
	}

	mimeType := getMimeType(imagePath)
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, mimeType, nil
}

// getMimeType returns the MIME type based on file extension
func getMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	default:
		return "image/png"
	}
}

// buildPrompt constructs the prompt for Gemini to translate and replace text
func buildPrompt(tagline, targetLang string) string {
	return fmt.Sprintf(`You are given a YouTube thumbnail image. The image contains the text: "%s"

Translate that text to %s and replace it in the image.

Keep everything else exactly the same:
- Same colors, fonts, and styling
- Same positioning and layout
- Same background and all other elements
- Only the specified text should change

Generate the modified image.`, tagline, targetLang)
}

// GetLanguageName returns the full language name for a language code.
// Returns empty string if the language is not supported.
func GetLanguageName(langCode string) string {
	return languageNames[langCode]
}

// IsSupportedLanguage checks if a language code is supported for thumbnail localization
func IsSupportedLanguage(langCode string) bool {
	_, ok := languageNames[langCode]
	return ok
}
