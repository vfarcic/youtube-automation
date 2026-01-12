package dubbing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	defaultBaseURL = "https://api.elevenlabs.io"
)

// Errors returned by the client
var (
	ErrDubbingFailed     = errors.New("dubbing job failed")
	ErrDubbingInProgress = errors.New("dubbing still in progress")
	ErrInvalidAPIKey     = errors.New("invalid API key")
	ErrDubbingNotFound   = errors.New("dubbing job not found")
)

// Client is the ElevenLabs API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	config     Config
}

// NewClient creates a new ElevenLabs API client
func NewClient(apiKey string, config Config) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{},
		config:     config,
	}
}

// NewClientWithHTTPClient creates a new ElevenLabs API client with a custom HTTP client
// This is useful for testing with mock servers
func NewClientWithHTTPClient(apiKey string, config Config, httpClient *http.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
		config:     config,
	}
}

// CreateDubFromURL initiates a dubbing job using a video URL (e.g., YouTube)
// POST /v1/dubbing with URL and target language
func (c *Client) CreateDubFromURL(ctx context.Context, videoURL, sourceLang, targetLang string) (*DubbingJob, error) {
	// Create multipart form with URL instead of file
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add the video URL
	if err := writer.WriteField("source_url", videoURL); err != nil {
		return nil, fmt.Errorf("failed to write source_url: %w", err)
	}

	// Add target language
	if err := writer.WriteField("target_lang", targetLang); err != nil {
		return nil, fmt.Errorf("failed to write target_lang: %w", err)
	}

	// Add source language if provided
	if sourceLang != "" {
		if err := writer.WriteField("source_lang", sourceLang); err != nil {
			return nil, fmt.Errorf("failed to write source_lang: %w", err)
		}
	}

	// Add number of speakers
	numSpeakers := c.config.NumSpeakers
	if numSpeakers <= 0 {
		numSpeakers = 1
	}
	if err := writer.WriteField("num_speakers", strconv.Itoa(numSpeakers)); err != nil {
		return nil, fmt.Errorf("failed to write num_speakers: %w", err)
	}

	// Add drop_background_audio
	if err := writer.WriteField("drop_background_audio", strconv.FormatBool(c.config.DropBackgroundAudio)); err != nil {
		return nil, fmt.Errorf("failed to write drop_background_audio: %w", err)
	}

	// Add test mode settings
	if err := writer.WriteField("watermark", strconv.FormatBool(c.config.TestMode)); err != nil {
		return nil, fmt.Errorf("failed to write watermark: %w", err)
	}
	if err := writer.WriteField("highest_resolution", strconv.FormatBool(!c.config.TestMode)); err != nil {
		return nil, fmt.Errorf("failed to write highest_resolution: %w", err)
	}

	// Add start/end time if specified
	if c.config.StartTime > 0 {
		if err := writer.WriteField("start_time", strconv.Itoa(c.config.StartTime)); err != nil {
			return nil, fmt.Errorf("failed to write start_time: %w", err)
		}
	}
	if c.config.EndTime > 0 {
		if err := writer.WriteField("end_time", strconv.Itoa(c.config.EndTime)); err != nil {
			return nil, fmt.Errorf("failed to write end_time: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Create the request
	url := c.baseURL + "/v1/dubbing"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("xi-api-key", c.apiKey)

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAPIKey
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if json.Unmarshal(respBody, &errResp) == nil {
			if msg := errResp.GetMessage(); msg != "" {
				return nil, fmt.Errorf("dubbing request failed (status %d): %s", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("dubbing request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var createResp createDubbingResponse
	if err := json.Unmarshal(respBody, &createResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &DubbingJob{
		DubbingID:        createResp.DubbingID,
		Status:           StatusDubbing,
		TargetLanguages:  []string{targetLang},
		ExpectedDuration: createResp.ExpectedDuration,
	}, nil
}

// GetDubbingStatus checks job status
// GET /v1/dubbing/{dubbing_id}
func (c *Client) GetDubbingStatus(ctx context.Context, dubbingID string) (*DubbingJob, error) {
	url := fmt.Sprintf("%s/v1/dubbing/%s", c.baseURL, dubbingID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAPIKey
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrDubbingNotFound
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if json.Unmarshal(body, &errResp) == nil {
			if msg := errResp.GetMessage(); msg != "" {
				return nil, fmt.Errorf("get status failed (status %d): %s", resp.StatusCode, msg)
			}
		}
		return nil, fmt.Errorf("get status failed with status %d: %s", resp.StatusCode, string(body))
	}

	var job DubbingJob
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &job, nil
}

// DownloadDubbedAudio downloads the dubbed video
// GET /v1/dubbing/{dubbing_id}/audio/{language_code}
func (c *Client) DownloadDubbedAudio(ctx context.Context, dubbingID, langCode, outputPath string) error {
	// First check the status to ensure dubbing is complete
	job, err := c.GetDubbingStatus(ctx, dubbingID)
	if err != nil {
		return fmt.Errorf("failed to check dubbing status: %w", err)
	}

	if job.Status == StatusDubbing {
		return ErrDubbingInProgress
	}

	if job.Status == StatusFailed {
		if job.Error != "" {
			return fmt.Errorf("%w: %s", ErrDubbingFailed, job.Error)
		}
		return ErrDubbingFailed
	}

	// Download the dubbed audio
	url := fmt.Sprintf("%s/v1/dubbing/%s/audio/%s", c.baseURL, dubbingID, langCode)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("xi-api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrInvalidAPIKey
	}

	if resp.StatusCode == http.StatusNotFound {
		return ErrDubbingNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Stream the response to the file
	if _, err := io.Copy(outFile, resp.Body); err != nil {
		return fmt.Errorf("failed to write dubbed audio to file: %w", err)
	}

	return nil
}
