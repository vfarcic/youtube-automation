package thumbnail

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

var (
	ErrGPTImageNoAPIKey            = errors.New("OPENAI_API_KEY is not set")
	ErrGPTImageEmptyPrompt         = errors.New("prompt cannot be empty")
	ErrGPTImageNoPhotos            = errors.New("at least one reference photo is required")
	ErrGPTImageAPIError            = errors.New("openai API error")
	ErrGPTImageNoImage             = errors.New("no image returned by OpenAI")
	ErrGPTImageContentFiltered     = errors.New("image generation was filtered by content policy")
	ErrGPTImageUnexpectedContentType = errors.New("unexpected response content type")
)

const (
	gptImageBaseURL = "https://api.openai.com/v1/images/edits"

	// gptImageHTTPTimeout is the timeout for OpenAI API requests.
	// Image generation can take 30-90s, so we use 120s to allow headroom.
	gptImageHTTPTimeout = 120 * time.Second

	// gptImageMaxResponseBytes is the maximum response body size (50MB).
	// This prevents OOM from malicious or buggy responses returning unbounded data.
	gptImageMaxResponseBytes = 50 * 1024 * 1024
)

// GPTImageClient implements ImageGenerator using the OpenAI image generation API.
type GPTImageClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string // overridable for testing
}

// NewGPTImageClient creates a new OpenAI image generation client.
// If httpClient is nil, a client with a sensible timeout is created.
func NewGPTImageClient(apiKey, model string, httpClient *http.Client) (*GPTImageClient, error) {
	if apiKey == "" {
		return nil, ErrGPTImageNoAPIKey
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: gptImageHTTPTimeout}
	}
	return &GPTImageClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: httpClient,
		baseURL:    gptImageBaseURL,
	}, nil
}

func (g *GPTImageClient) Name() string {
	return "gpt-image"
}

// gptImageResponse is the top-level response from the OpenAI images API.
type gptImageResponse struct {
	Data  []gptImageData `json:"data"`
	Error *gptImageError `json:"error,omitempty"`
}

type gptImageData struct {
	B64JSON string `json:"b64_json"`
}

type gptImageError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func (g *GPTImageClient) GenerateImage(ctx context.Context, prompt string, photos [][]byte) ([]byte, error) {
	if prompt == "" {
		return nil, ErrGPTImageEmptyPrompt
	}
	if len(photos) == 0 {
		return nil, ErrGPTImageNoPhotos
	}

	body, contentType, err := g.buildMultipartRequest(prompt, photos)
	if err != nil {
		return nil, fmt.Errorf("building openai request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL, body)
	if err != nil {
		return nil, fmt.Errorf("creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Validate response content-type before reading/parsing the body.
	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.HasPrefix(ct, "application/json") {
		return nil, fmt.Errorf("%w: expected application/json, got %s", ErrGPTImageUnexpectedContentType, ct)
	}

	// Limit response body reads to prevent OOM from oversized responses.
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, gptImageMaxResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("reading openai response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Parse error response to detect content policy violations and avoid
		// leaking raw upstream response bodies in error messages.
		var errResp gptImageResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != nil {
			if errResp.Error.Code == "content_policy_violation" {
				return nil, ErrGPTImageContentFiltered
			}
			return nil, fmt.Errorf("%w: HTTP %d: %s (code: %s)", ErrGPTImageAPIError, resp.StatusCode, errResp.Error.Message, errResp.Error.Code)
		}
		return nil, fmt.Errorf("%w: HTTP %d", ErrGPTImageAPIError, resp.StatusCode)
	}

	return g.parseResponse(respBody)
}

func (g *GPTImageClient) buildMultipartRequest(prompt string, photos [][]byte) (*bytes.Buffer, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	if err := writer.WriteField("model", g.model); err != nil {
		return nil, "", fmt.Errorf("writing model field: %w", err)
	}

	if err := writer.WriteField("prompt", prompt); err != nil {
		return nil, "", fmt.Errorf("writing prompt field: %w", err)
	}

	if err := writer.WriteField("size", "1536x1024"); err != nil {
		return nil, "", fmt.Errorf("writing size field: %w", err)
	}

	if err := writer.WriteField("quality", "high"); err != nil {
		return nil, "", fmt.Errorf("writing quality field: %w", err)
	}

	for i, photo := range photos {
		mimeType := detectMimeType(photo)
		ext := extensionFromMimeType(mimeType)
		filename := fmt.Sprintf("photo_%d.%s", i, ext)

		// Use CreatePart with explicit MIME type instead of CreateFormFile
		// which hardcodes application/octet-stream (rejected by OpenAI).
		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image[]"; filename="%s"`, filename))
		header.Set("Content-Type", mimeType)
		part, err := writer.CreatePart(header)
		if err != nil {
			return nil, "", fmt.Errorf("creating form file: %w", err)
		}
		if _, err := part.Write(photo); err != nil {
			return nil, "", fmt.Errorf("writing photo data: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("closing multipart writer: %w", err)
	}

	return &buf, writer.FormDataContentType(), nil
}

// extensionFromMimeType returns the file extension for a MIME type.
func extensionFromMimeType(mimeType string) string {
	switch mimeType {
	case "image/png":
		return "png"
	case "image/webp":
		return "webp"
	default:
		return "jpg"
	}
}

func (g *GPTImageClient) parseResponse(body []byte) ([]byte, error) {
	var resp gptImageResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing openai response: %w", err)
	}

	if resp.Error != nil {
		if resp.Error.Code == "content_policy_violation" {
			return nil, ErrGPTImageContentFiltered
		}
		return nil, fmt.Errorf("%w: %s (type: %s, code: %s)", ErrGPTImageAPIError, resp.Error.Message, resp.Error.Type, resp.Error.Code)
	}

	if len(resp.Data) == 0 {
		return nil, ErrGPTImageNoImage
	}

	if resp.Data[0].B64JSON == "" {
		return nil, ErrGPTImageNoImage
	}

	imgBytes, err := base64.StdEncoding.DecodeString(resp.Data[0].B64JSON)
	if err != nil {
		return nil, fmt.Errorf("decoding openai image data: %w", err)
	}

	return imgBytes, nil
}
