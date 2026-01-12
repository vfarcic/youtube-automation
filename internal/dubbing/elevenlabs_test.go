package dubbing

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-api-key"
	config := Config{
		TestMode:    true,
		NumSpeakers: 1,
	}

	client := NewClient(apiKey, config)

	if client == nil {
		t.Fatal("expected client to be non-nil")
	}
	if client.apiKey != apiKey {
		t.Errorf("expected apiKey %q, got %q", apiKey, client.apiKey)
	}
	if client.baseURL != defaultBaseURL {
		t.Errorf("expected baseURL %q, got %q", defaultBaseURL, client.baseURL)
	}
	if client.config.TestMode != true {
		t.Error("expected TestMode to be true")
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	apiKey := "test-api-key"
	config := Config{}
	httpClient := &http.Client{}
	customBaseURL := "https://custom.api.com"

	client := NewClientWithHTTPClient(apiKey, config, httpClient, customBaseURL)

	if client.baseURL != customBaseURL {
		t.Errorf("expected baseURL %q, got %q", customBaseURL, client.baseURL)
	}

	// Test with empty baseURL defaults to defaultBaseURL
	client2 := NewClientWithHTTPClient(apiKey, config, httpClient, "")
	if client2.baseURL != defaultBaseURL {
		t.Errorf("expected baseURL %q, got %q", defaultBaseURL, client2.baseURL)
	}
}

func TestCreateDubFromURL(t *testing.T) {
	tests := []struct {
		name           string
		videoURL       string
		serverResponse string
		statusCode     int
		sourceLang     string
		targetLang     string
		config         Config
		wantErr        bool
		wantErrType    error
		wantDubbingID  string
	}{
		{
			name:           "success with YouTube URL",
			videoURL:       "https://www.youtube.com/watch?v=abc123",
			serverResponse: `{"dubbing_id":"dub_yt_123","expected_duration_sec":300.0}`,
			statusCode:     http.StatusOK,
			sourceLang:     "en",
			targetLang:     "es",
			config:         Config{NumSpeakers: 1},
			wantErr:        false,
			wantDubbingID:  "dub_yt_123",
		},
		{
			name:           "success without source lang",
			videoURL:       "https://youtu.be/xyz789",
			serverResponse: `{"dubbing_id":"dub_yt_456"}`,
			statusCode:     http.StatusOK,
			sourceLang:     "",
			targetLang:     "es",
			config:         Config{NumSpeakers: 1},
			wantErr:        false,
			wantDubbingID:  "dub_yt_456",
		},
		{
			name:           "success with test mode",
			videoURL:       "https://www.youtube.com/watch?v=test123",
			serverResponse: `{"dubbing_id":"dub_yt_test"}`,
			statusCode:     http.StatusOK,
			sourceLang:     "en",
			targetLang:     "es",
			config:         Config{TestMode: true, StartTime: 0, EndTime: 60},
			wantErr:        false,
			wantDubbingID:  "dub_yt_test",
		},
		{
			name:           "unauthorized",
			videoURL:       "https://www.youtube.com/watch?v=abc123",
			serverResponse: `{"detail":{"message":"Invalid API key"}}`,
			statusCode:     http.StatusUnauthorized,
			sourceLang:     "en",
			targetLang:     "es",
			config:         Config{},
			wantErr:        true,
			wantErrType:    ErrInvalidAPIKey,
		},
		{
			name:           "server error",
			videoURL:       "https://www.youtube.com/watch?v=abc123",
			serverResponse: `{"detail":{"message":"Failed to fetch video"}}`,
			statusCode:     http.StatusInternalServerError,
			sourceLang:     "en",
			targetLang:     "es",
			config:         Config{},
			wantErr:        true,
		},
		{
			name:           "rate limited",
			videoURL:       "https://www.youtube.com/watch?v=abc123",
			serverResponse: `{"detail":{"message":"Rate limit exceeded"}}`,
			statusCode:     http.StatusTooManyRequests,
			sourceLang:     "en",
			targetLang:     "es",
			config:         Config{},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/v1/dubbing" {
					t.Errorf("expected path /v1/dubbing, got %s", r.URL.Path)
				}
				if r.Header.Get("xi-api-key") == "" {
					t.Error("expected xi-api-key header")
				}
				if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
					t.Errorf("expected multipart/form-data content type, got %s", r.Header.Get("Content-Type"))
				}

				// Parse multipart form to verify fields
				if err := r.ParseMultipartForm(10 << 20); err != nil {
					t.Errorf("failed to parse multipart form: %v", err)
				}

				// Check source_url field
				if r.FormValue("source_url") != tt.videoURL {
					t.Errorf("expected source_url %q, got %q", tt.videoURL, r.FormValue("source_url"))
				}

				// Check required fields
				if r.FormValue("target_lang") != tt.targetLang {
					t.Errorf("expected target_lang %q, got %q", tt.targetLang, r.FormValue("target_lang"))
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewClientWithHTTPClient("test-api-key", tt.config, server.Client(), server.URL)

			job, err := client.CreateDubFromURL(context.Background(), tt.videoURL, tt.sourceLang, tt.targetLang)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrType != nil && err != tt.wantErrType {
					if !strings.Contains(err.Error(), tt.wantErrType.Error()) {
						t.Errorf("expected error type %v, got %v", tt.wantErrType, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if job.DubbingID != tt.wantDubbingID {
				t.Errorf("expected DubbingID %q, got %q", tt.wantDubbingID, job.DubbingID)
			}
			if job.Status != StatusDubbing {
				t.Errorf("expected Status %q, got %q", StatusDubbing, job.Status)
			}
		})
	}
}

func TestCreateDubFromURL_WithStartEndTime(t *testing.T) {
	var receivedStartTime, receivedEndTime string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		receivedStartTime = r.FormValue("start_time")
		receivedEndTime = r.FormValue("end_time")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dubbing_id":"dub_segment"}`))
	}))
	defer server.Close()

	config := Config{
		StartTime: 30,
		EndTime:   90,
	}
	client := NewClientWithHTTPClient("test-api-key", config, server.Client(), server.URL)

	_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedStartTime != "30" {
		t.Errorf("expected start_time 30, got %q", receivedStartTime)
	}
	if receivedEndTime != "90" {
		t.Errorf("expected end_time 90, got %q", receivedEndTime)
	}
}

func TestCreateDubFromURL_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dubbing_id":"dub_123"}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.CreateDubFromURL(ctx, "https://www.youtube.com/watch?v=abc123", "en", "es")

	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestCreateDubFromURL_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("expected parse error, got %v", err)
	}
}

func TestCreateDubFromURL_DefaultNumSpeakers(t *testing.T) {
	var receivedNumSpeakers string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		receivedNumSpeakers = r.FormValue("num_speakers")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dubbing_id":"dub_test"}`))
	}))
	defer server.Close()

	// Config with NumSpeakers = 0 should default to 1
	config := Config{NumSpeakers: 0}
	client := NewClientWithHTTPClient("test-api-key", config, server.Client(), server.URL)

	_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedNumSpeakers != "1" {
		t.Errorf("expected num_speakers to default to 1, got %q", receivedNumSpeakers)
	}
}

func TestCreateDubFromURL_NegativeNumSpeakers(t *testing.T) {
	var receivedNumSpeakers string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		receivedNumSpeakers = r.FormValue("num_speakers")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dubbing_id":"dub_test"}`))
	}))
	defer server.Close()

	// Negative NumSpeakers should default to 1
	config := Config{NumSpeakers: -5}
	client := NewClientWithHTTPClient("test-api-key", config, server.Client(), server.URL)

	_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedNumSpeakers != "1" {
		t.Errorf("expected num_speakers to default to 1 for negative value, got %q", receivedNumSpeakers)
	}
}

func TestCreateDubFromURL_DropBackgroundAudio(t *testing.T) {
	var receivedDropBackground string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		receivedDropBackground = r.FormValue("drop_background_audio")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dubbing_id":"dub_test"}`))
	}))
	defer server.Close()

	config := Config{DropBackgroundAudio: true}
	client := NewClientWithHTTPClient("test-api-key", config, server.Client(), server.URL)

	_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedDropBackground != "true" {
		t.Errorf("expected drop_background_audio to be true, got %q", receivedDropBackground)
	}
}

func TestCreateDubFromURL_WatermarkAndResolution(t *testing.T) {
	tests := []struct {
		name               string
		testMode           bool
		wantWatermark      string
		wantHighResolution string
	}{
		{
			name:               "test mode enabled",
			testMode:           true,
			wantWatermark:      "true",
			wantHighResolution: "false",
		},
		{
			name:               "test mode disabled",
			testMode:           false,
			wantWatermark:      "false",
			wantHighResolution: "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var receivedWatermark, receivedHighRes string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if err := r.ParseMultipartForm(10 << 20); err != nil {
					t.Errorf("failed to parse form: %v", err)
				}
				receivedWatermark = r.FormValue("watermark")
				receivedHighRes = r.FormValue("highest_resolution")

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"dubbing_id":"dub_test"}`))
			}))
			defer server.Close()

			config := Config{TestMode: tt.testMode}
			client := NewClientWithHTTPClient("test-api-key", config, server.Client(), server.URL)

			_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if receivedWatermark != tt.wantWatermark {
				t.Errorf("expected watermark %q, got %q", tt.wantWatermark, receivedWatermark)
			}
			if receivedHighRes != tt.wantHighResolution {
				t.Errorf("expected highest_resolution %q, got %q", tt.wantHighResolution, receivedHighRes)
			}
		})
	}
}

func TestCreateDubFromURL_StartTimeOnlyNoEndTime(t *testing.T) {
	var receivedStartTime, receivedEndTime string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		receivedStartTime = r.FormValue("start_time")
		receivedEndTime = r.FormValue("end_time")

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dubbing_id":"dub_test"}`))
	}))
	defer server.Close()

	// Only set start time, no end time
	config := Config{StartTime: 10, EndTime: 0}
	client := NewClientWithHTTPClient("test-api-key", config, server.Client(), server.URL)

	_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedStartTime != "10" {
		t.Errorf("expected start_time 10, got %q", receivedStartTime)
	}
	if receivedEndTime != "" {
		t.Errorf("expected no end_time, got %q", receivedEndTime)
	}
}

func TestCreateDubFromURL_ErrorResponseWithoutDetail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`plain error message`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	_, err := client.CreateDubFromURL(context.Background(), "https://www.youtube.com/watch?v=abc123", "en", "es")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "plain error message") {
		t.Errorf("expected raw error message, got %v", err)
	}
}

func TestGetDubbingStatus(t *testing.T) {
	tests := []struct {
		name           string
		dubbingID      string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantErrType    error
		wantStatus     string
	}{
		{
			name:           "dubbing in progress",
			dubbingID:      "dub_123",
			serverResponse: `{"dubbing_id":"dub_123","status":"dubbing","target_languages":["es"]}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantStatus:     StatusDubbing,
		},
		{
			name:           "dubbing complete",
			dubbingID:      "dub_456",
			serverResponse: `{"dubbing_id":"dub_456","status":"dubbed","target_languages":["es"]}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantStatus:     StatusDubbed,
		},
		{
			name:           "dubbing failed",
			dubbingID:      "dub_789",
			serverResponse: `{"dubbing_id":"dub_789","status":"failed","error":"Audio processing error"}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
			wantStatus:     StatusFailed,
		},
		{
			name:           "not found",
			dubbingID:      "nonexistent",
			serverResponse: `{"detail":{"message":"Dubbing not found"}}`,
			statusCode:     http.StatusNotFound,
			wantErr:        true,
			wantErrType:    ErrDubbingNotFound,
		},
		{
			name:           "unauthorized",
			dubbingID:      "dub_123",
			serverResponse: `{"detail":{"message":"Invalid API key"}}`,
			statusCode:     http.StatusUnauthorized,
			wantErr:        true,
			wantErrType:    ErrInvalidAPIKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				expectedPath := "/v1/dubbing/" + tt.dubbingID
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Header.Get("xi-api-key") == "" {
					t.Error("expected xi-api-key header")
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

			job, err := client.GetDubbingStatus(context.Background(), tt.dubbingID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrType != nil && err != tt.wantErrType {
					t.Errorf("expected error type %v, got %v", tt.wantErrType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if job.Status != tt.wantStatus {
				t.Errorf("expected Status %q, got %q", tt.wantStatus, job.Status)
			}
		})
	}
}

func TestGetDubbingStatus_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"detail":{"message":"Internal server error"}}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	_, err := client.GetDubbingStatus(context.Background(), "dub_123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "Internal server error") {
		t.Errorf("expected error message containing 'Internal server error', got %v", err)
	}
}

func TestGetDubbingStatus_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	_, err := client.GetDubbingStatus(context.Background(), "dub_123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("expected parse error, got %v", err)
	}
}

func TestGetDubbingStatus_ErrorResponseWithoutDetail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`plain error message`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	_, err := client.GetDubbingStatus(context.Background(), "dub_123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "plain error message") {
		t.Errorf("expected raw error message, got %v", err)
	}
}

func TestGetDubbingStatus_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"dubbing_id":"dub_123","status":"dubbed"}`))
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetDubbingStatus(ctx, "dub_123")

	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestDownloadDubbedAudio(t *testing.T) {
	tests := []struct {
		name             string
		dubbingID        string
		langCode         string
		statusResponse   string
		statusStatusCode int
		audioContent     string
		audioStatusCode  int
		wantErr          bool
		wantErrType      error
	}{
		{
			name:             "success",
			dubbingID:        "dub_123",
			langCode:         "es",
			statusResponse:   `{"dubbing_id":"dub_123","status":"dubbed"}`,
			statusStatusCode: http.StatusOK,
			audioContent:     "fake audio content",
			audioStatusCode:  http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "dubbing in progress",
			dubbingID:        "dub_456",
			langCode:         "es",
			statusResponse:   `{"dubbing_id":"dub_456","status":"dubbing"}`,
			statusStatusCode: http.StatusOK,
			wantErr:          true,
			wantErrType:      ErrDubbingInProgress,
		},
		{
			name:             "dubbing failed",
			dubbingID:        "dub_789",
			langCode:         "es",
			statusResponse:   `{"dubbing_id":"dub_789","status":"failed","error":"Processing error"}`,
			statusStatusCode: http.StatusOK,
			wantErr:          true,
			wantErrType:      ErrDubbingFailed,
		},
		{
			name:             "status not found",
			dubbingID:        "nonexistent",
			langCode:         "es",
			statusResponse:   `{"detail":{"message":"Not found"}}`,
			statusStatusCode: http.StatusNotFound,
			wantErr:          true,
			wantErrType:      ErrDubbingNotFound,
		},
		{
			name:             "audio download unauthorized",
			dubbingID:        "dub_auth",
			langCode:         "es",
			statusResponse:   `{"dubbing_id":"dub_auth","status":"dubbed"}`,
			statusStatusCode: http.StatusOK,
			audioStatusCode:  http.StatusUnauthorized,
			wantErr:          true,
			wantErrType:      ErrInvalidAPIKey,
		},
		{
			name:             "audio not found",
			dubbingID:        "dub_noaudio",
			langCode:         "es",
			statusResponse:   `{"dubbing_id":"dub_noaudio","status":"dubbed"}`,
			statusStatusCode: http.StatusOK,
			audioStatusCode:  http.StatusNotFound,
			wantErr:          true,
			wantErrType:      ErrDubbingNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle status check
				if r.URL.Path == "/v1/dubbing/"+tt.dubbingID {
					w.WriteHeader(tt.statusStatusCode)
					w.Write([]byte(tt.statusResponse))
					return
				}

				// Handle audio download
				expectedAudioPath := "/v1/dubbing/" + tt.dubbingID + "/audio/" + tt.langCode
				if r.URL.Path == expectedAudioPath {
					if r.Header.Get("xi-api-key") == "" {
						t.Error("expected xi-api-key header")
					}
					w.WriteHeader(tt.audioStatusCode)
					if tt.audioStatusCode == http.StatusOK {
						w.Write([]byte(tt.audioContent))
					}
					return
				}

				t.Errorf("unexpected request path: %s", r.URL.Path)
			}))
			defer server.Close()

			client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

			tmpDir := t.TempDir()
			outputPath := filepath.Join(tmpDir, "output.mp4")

			err := client.DownloadDubbedAudio(context.Background(), tt.dubbingID, tt.langCode, outputPath)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrType != nil {
					if !strings.Contains(err.Error(), tt.wantErrType.Error()) {
						t.Errorf("expected error containing %v, got %v", tt.wantErrType, err)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify file was created and contains expected content
			content, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output file: %v", err)
			}
			if string(content) != tt.audioContent {
				t.Errorf("expected content %q, got %q", tt.audioContent, string(content))
			}
		})
	}
}

func TestDownloadDubbedAudio_CreatesDirectory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/audio/") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("audio content"))
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DubbingJob{DubbingID: "dub_123", Status: StatusDubbed})
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	// Create a nested path that doesn't exist
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "nested", "dirs", "output.mp4")

	err := client.DownloadDubbedAudio(context.Background(), "dub_123", "es", outputPath)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("expected output file to exist")
	}

	// Verify content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if string(content) != "audio content" {
		t.Errorf("expected content 'audio content', got %q", string(content))
	}
}

func TestDownloadDubbedAudio_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/audio/") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("audio content"))
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DubbingJob{DubbingID: "dub_123", Status: StatusDubbed})
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.DownloadDubbedAudio(ctx, "dub_123", "es", outputPath)

	if err == nil {
		t.Fatal("expected error due to cancelled context")
	}
}

func TestDownloadDubbedAudio_LargeFile(t *testing.T) {
	// Generate 1MB of content to test streaming
	largeContent := strings.Repeat("x", 1024*1024)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/audio/") {
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, largeContent)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DubbingJob{DubbingID: "dub_large", Status: StatusDubbed})
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "large_output.mp4")

	err := client.DownloadDubbedAudio(context.Background(), "dub_large", "es", outputPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file size
	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("failed to stat output file: %v", err)
	}
	if info.Size() != int64(len(largeContent)) {
		t.Errorf("expected file size %d, got %d", len(largeContent), info.Size())
	}
}

func TestDownloadDubbedAudio_DownloadError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/audio/") {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`server error`))
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DubbingJob{DubbingID: "dub_123", Status: StatusDubbed})
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	err := client.DownloadDubbedAudio(context.Background(), "dub_123", "es", outputPath)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "download failed") {
		t.Errorf("expected download failed error, got %v", err)
	}
}

func TestDownloadDubbedAudio_FailedWithoutErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(DubbingJob{DubbingID: "dub_123", Status: StatusFailed, Error: ""})
	}))
	defer server.Close()

	client := NewClientWithHTTPClient("test-api-key", Config{}, server.Client(), server.URL)

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.mp4")

	err := client.DownloadDubbedAudio(context.Background(), "dub_123", "es", outputPath)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrDubbingFailed {
		t.Errorf("expected ErrDubbingFailed, got %v", err)
	}
}

func TestDubbingStatusConstants(t *testing.T) {
	if StatusDubbing != "dubbing" {
		t.Errorf("expected StatusDubbing to be 'dubbing', got %q", StatusDubbing)
	}
	if StatusDubbed != "dubbed" {
		t.Errorf("expected StatusDubbed to be 'dubbed', got %q", StatusDubbed)
	}
	if StatusFailed != "failed" {
		t.Errorf("expected StatusFailed to be 'failed', got %q", StatusFailed)
	}
}

func TestDubbingJob_JSONMarshal(t *testing.T) {
	job := DubbingJob{
		DubbingID:        "dub_123",
		Name:             "Test Dub",
		Status:           StatusDubbed,
		TargetLanguages:  []string{"es", "fr"},
		ExpectedDuration: 120.5,
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled DubbingJob
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.DubbingID != job.DubbingID {
		t.Errorf("expected DubbingID %q, got %q", job.DubbingID, unmarshaled.DubbingID)
	}
	if unmarshaled.Name != job.Name {
		t.Errorf("expected Name %q, got %q", job.Name, unmarshaled.Name)
	}
}
