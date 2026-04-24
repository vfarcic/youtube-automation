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
	"strings"
	"time"
)

var (
	ErrGeminiNoAPIKey     = errors.New("GEMINI_API_KEY is not set")
	ErrGeminiEmptyPrompt  = errors.New("prompt cannot be empty")
	ErrGeminiNoPhotos     = errors.New("at least one reference photo is required")
	ErrGeminiAPIError     = errors.New("gemini API error")
	ErrGeminiNoImage      = errors.New("no image returned by Gemini")
	ErrGeminiImageFiltered = errors.New("image generation was filtered by safety settings")
)

var (
	ErrGeminiUnexpectedContentType = errors.New("unexpected response content type")
)

const (
	geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

	// geminiHTTPTimeout is the timeout for Gemini API requests.
	// Image generation can take 30-90s, so we use 120s to allow headroom.
	geminiHTTPTimeout = 120 * time.Second

	// geminiMaxResponseBytes is the maximum response body size (50MB).
	// This prevents OOM from malicious or buggy responses returning unbounded data.
	geminiMaxResponseBytes = 50 * 1024 * 1024
)

// GeminiClient implements ImageGenerator using the Gemini REST API.
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string // overridable for testing
}

// NewGeminiClient creates a new Gemini image generation client.
// If httpClient is nil, a client with a sensible timeout is created.
func NewGeminiClient(apiKey, model string, httpClient *http.Client) (*GeminiClient, error) {
	if apiKey == "" {
		return nil, ErrGeminiNoAPIKey
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: geminiHTTPTimeout}
	}
	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: httpClient,
		baseURL:    geminiBaseURL,
	}, nil
}

func (g *GeminiClient) Name() string {
	return "gemini"
}

// geminiRequest is the top-level request payload for the Gemini generateContent API.
type geminiRequest struct {
	Contents         []geminiContent         `json:"contents"`
	GenerationConfig geminiGenerationConfig  `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string          `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiGenerationConfig struct {
	ResponseModalities []string         `json:"responseModalities"`
	ImageConfig        geminiImageConfig `json:"imageConfig"`
}

type geminiImageConfig struct {
	AspectRatio string `json:"aspectRatio"`
	ImageSize   string `json:"imageSize"`
}

// geminiResponse is the top-level response from the Gemini generateContent API.
type geminiResponse struct {
	Candidates []geminiCandidate `json:"candidates"`
	Error      *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

func (g *GeminiClient) GenerateImage(ctx context.Context, prompt string, photos [][]byte) ([]byte, error) {
	if prompt == "" {
		return nil, ErrGeminiEmptyPrompt
	}
	if len(photos) == 0 {
		return nil, ErrGeminiNoPhotos
	}

	reqBody, err := g.buildRequest(prompt, photos)
	if err != nil {
		return nil, fmt.Errorf("building gemini request: %w", err)
	}

	// The Gemini REST API requires the API key as a query parameter.
	// It does not support Authorization header authentication for API keys.
	// https://ai.google.dev/gemini-api/docs/quickstart
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", g.baseURL, g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Validate response content-type before reading/parsing the body.
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		return nil, fmt.Errorf("%w: expected application/json, got %s", ErrGeminiUnexpectedContentType, ct)
	}

	// Limit response body reads to prevent OOM from oversized responses.
	body, err := io.ReadAll(io.LimitReader(resp.Body, geminiMaxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("reading gemini response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: HTTP %d: %s", ErrGeminiAPIError, resp.StatusCode, string(body))
	}

	return g.parseResponse(body)
}

func (g *GeminiClient) buildRequest(prompt string, photos [][]byte) ([]byte, error) {
	parts := []geminiPart{
		{Text: prompt},
	}

	for _, photo := range photos {
		mimeType := detectMimeType(photo)
		parts = append(parts, geminiPart{
			InlineData: &geminiInlineData{
				MimeType: mimeType,
				Data:     base64.StdEncoding.EncodeToString(photo),
			},
		})
	}

	reqPayload := geminiRequest{
		Contents: []geminiContent{
			{Parts: parts},
		},
		GenerationConfig: geminiGenerationConfig{
			ResponseModalities: []string{"IMAGE", "TEXT"},
			ImageConfig: geminiImageConfig{
				AspectRatio: "16:9",
				ImageSize:   "2K",
			},
		},
	}

	return json.Marshal(reqPayload)
}

func (g *GeminiClient) parseResponse(body []byte) ([]byte, error) {
	var resp geminiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing gemini response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("%w: %s (code %d, status %s)", ErrGeminiAPIError, resp.Error.Message, resp.Error.Code, resp.Error.Status)
	}

	for _, candidate := range resp.Candidates {
		// Gemini returns finishReason "SAFETY" when content is blocked by safety filters.
		if candidate.FinishReason == "SAFETY" {
			return nil, ErrGeminiImageFiltered
		}
		for _, part := range candidate.Content.Parts {
			if part.InlineData != nil && part.InlineData.Data != "" {
				imgBytes, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
				if err != nil {
					return nil, fmt.Errorf("decoding gemini image data: %w", err)
				}
				return imgBytes, nil
			}
		}
	}

	return nil, ErrGeminiNoImage
}

// detectMimeType returns the MIME type based on file magic bytes.
func detectMimeType(data []byte) string {
	if len(data) < 4 {
		return "image/jpeg"
	}
	// PNG: 89 50 4E 47
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	// WEBP: RIFF....WEBP
	if len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}
	// Default to JPEG
	return "image/jpeg"
}
