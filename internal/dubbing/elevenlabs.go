package dubbing

import (
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
	ErrVideoNotFound     = errors.New("video file not found")
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

// CreateDub initiates a dubbing job
// POST /v1/dubbing with video file and target language
func (c *Client) CreateDub(ctx context.Context, videoPath, sourceLang, targetLang string) (*DubbingJob, error) {
	// Verify video file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: %s", ErrVideoNotFound, videoPath)
	}

	// Open the video file
	file, err := os.Open(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video file: %w", err)
	}
	defer file.Close()

	// Create multipart form
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	// Write form data in a goroutine to avoid blocking
	errChan := make(chan error, 1)
	go func() {
		defer pw.Close()
		defer writer.Close()

		// Add the video file
		part, err := writer.CreateFormFile("file", filepath.Base(videoPath))
		if err != nil {
			errChan <- fmt.Errorf("failed to create form file: %w", err)
			return
		}
		if _, err := io.Copy(part, file); err != nil {
			errChan <- fmt.Errorf("failed to copy video file: %w", err)
			return
		}

		// Add target language
		if err := writer.WriteField("target_lang", targetLang); err != nil {
			errChan <- fmt.Errorf("failed to write target_lang: %w", err)
			return
		}

		// Add source language if provided
		if sourceLang != "" {
			if err := writer.WriteField("source_lang", sourceLang); err != nil {
				errChan <- fmt.Errorf("failed to write source_lang: %w", err)
				return
			}
		}

		// Add number of speakers
		numSpeakers := c.config.NumSpeakers
		if numSpeakers <= 0 {
			numSpeakers = 1
		}
		if err := writer.WriteField("num_speakers", strconv.Itoa(numSpeakers)); err != nil {
			errChan <- fmt.Errorf("failed to write num_speakers: %w", err)
			return
		}

		// Add drop_background_audio
		if err := writer.WriteField("drop_background_audio", strconv.FormatBool(c.config.DropBackgroundAudio)); err != nil {
			errChan <- fmt.Errorf("failed to write drop_background_audio: %w", err)
			return
		}

		// Add test mode settings
		if err := writer.WriteField("watermark", strconv.FormatBool(c.config.TestMode)); err != nil {
			errChan <- fmt.Errorf("failed to write watermark: %w", err)
			return
		}
		if err := writer.WriteField("highest_resolution", strconv.FormatBool(!c.config.TestMode)); err != nil {
			errChan <- fmt.Errorf("failed to write highest_resolution: %w", err)
			return
		}

		// Add start/end time if specified
		if c.config.StartTime > 0 {
			if err := writer.WriteField("start_time", strconv.Itoa(c.config.StartTime)); err != nil {
				errChan <- fmt.Errorf("failed to write start_time: %w", err)
				return
			}
		}
		if c.config.EndTime > 0 {
			if err := writer.WriteField("end_time", strconv.Itoa(c.config.EndTime)); err != nil {
				errChan <- fmt.Errorf("failed to write end_time: %w", err)
				return
			}
		}

		errChan <- nil
	}()

	// Create the request
	url := c.baseURL + "/v1/dubbing"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, pr)
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

	// Check for errors from the goroutine
	if writeErr := <-errChan; writeErr != nil {
		return nil, writeErr
	}

	// Handle response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, ErrInvalidAPIKey
	}

	if resp.StatusCode != http.StatusOK {
		var errResp errorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Detail.Message != "" {
			return nil, fmt.Errorf("dubbing request failed (status %d): %s", resp.StatusCode, errResp.Detail.Message)
		}
		return nil, fmt.Errorf("dubbing request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var createResp createDubbingResponse
	if err := json.Unmarshal(body, &createResp); err != nil {
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
		if json.Unmarshal(body, &errResp) == nil && errResp.Detail.Message != "" {
			return nil, fmt.Errorf("get status failed (status %d): %s", resp.StatusCode, errResp.Detail.Message)
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
