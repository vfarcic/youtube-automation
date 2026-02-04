package thumbnail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/configuration"
)

// createTestImage creates a minimal valid PNG file for testing
func createTestImage(t *testing.T, dir string) string {
	t.Helper()
	// Minimal 1x1 red PNG (valid PNG format)
	pngData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR chunk
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41, // IDAT chunk
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x03, 0x00, 0x01, 0x00, 0x05, 0xFE,
		0xD4, 0xEF, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, // IEND chunk
		0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	imgPath := filepath.Join(dir, "test_thumbnail.png")
	if err := os.WriteFile(imgPath, pngData, 0644); err != nil {
		t.Fatalf("failed to create test image: %v", err)
	}
	return imgPath
}

// mockGeminiResponse creates a valid Gemini API response with an image
func mockGeminiResponse(imageData string, mimeType string) string {
	resp := geminiResponse{
		Candidates: []geminiCandidate{
			{
				Content: geminiContent{
					Parts: []geminiPart{
						{Text: "Here is the translated thumbnail:"},
						{
							InlineData: &geminiInlineData{
								MimeType: mimeType,
								Data:     imageData,
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// mockGeminiErrorResponse creates an error response from Gemini API
func mockGeminiErrorResponse(code int, message string) string {
	resp := geminiResponse{
		Error: &geminiError{
			Code:    code,
			Message: message,
		},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

// mockGeminiEmptyResponse creates a response with no candidates
func mockGeminiEmptyResponse() string {
	resp := geminiResponse{
		Candidates: []geminiCandidate{},
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name       string
		envAPIKey  string
		settingsModel string
		wantErr    error
		wantModel  string
	}{
		{
			name:       "success with env var and settings",
			envAPIKey:  "test-api-key",
			settingsModel: "gemini-2.5-flash-image",
			wantErr:    nil,
			wantModel:  "gemini-2.5-flash-image",
		},
		{
			name:       "success with default model",
			envAPIKey:  "test-api-key",
			settingsModel: "",
			wantErr:    nil,
			wantModel:  "gemini-3-pro-image-preview",
		},
		{
			name:      "error when API key not set",
			envAPIKey: "",
			wantErr:   ErrAPIKeyNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore environment
			origAPIKey := os.Getenv("GEMINI_API_KEY")
			defer os.Setenv("GEMINI_API_KEY", origAPIKey)

			// Save and restore settings
			origSettings := configuration.GlobalSettings
			defer func() { configuration.GlobalSettings = origSettings }()

			// Set test values
			os.Setenv("GEMINI_API_KEY", tt.envAPIKey)
			configuration.GlobalSettings.Gemini.Model = tt.settingsModel

			client, err := NewClient()

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("NewClient() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if err != tt.wantErr {
					t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
				return
			}

			if client.config.Model != tt.wantModel {
				t.Errorf("NewClient() model = %v, want %v", client.config.Model, tt.wantModel)
			}

			// Verify HTTP client has timeout set
			if client.httpClient.Timeout != 30*time.Second {
				t.Errorf("NewClient() httpClient.Timeout = %v, want %v", client.httpClient.Timeout, 30*time.Second)
			}
		})
	}
}

func TestNewClientWithHTTPClient(t *testing.T) {
	tests := []struct {
		name       string
		config     Config
		baseURL    string
		wantURL    string
	}{
		{
			name: "custom base URL",
			config: Config{
				APIKey: "test-key",
				Model:  "test-model",
			},
			baseURL: "http://localhost:8080",
			wantURL: "http://localhost:8080",
		},
		{
			name: "default base URL when empty",
			config: Config{
				APIKey: "test-key",
				Model:  "test-model",
			},
			baseURL: "",
			wantURL: defaultBaseURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientWithHTTPClient(tt.config, &http.Client{}, tt.baseURL)

			if client.baseURL != tt.wantURL {
				t.Errorf("baseURL = %v, want %v", client.baseURL, tt.wantURL)
			}
			if client.config.APIKey != tt.config.APIKey {
				t.Errorf("APIKey = %v, want %v", client.config.APIKey, tt.config.APIKey)
			}
		})
	}
}

func TestGenerateLocalizedThumbnail(t *testing.T) {
	// Create test image
	tmpDir := t.TempDir()
	imagePath := createTestImage(t, tmpDir)

	// Sample base64 encoded image for response
	sampleImageData := base64.StdEncoding.EncodeToString([]byte("fake-image-bytes"))

	tests := []struct {
		name           string
		tagline        string
		targetLang     string
		serverResponse string
		statusCode     int
		wantErr        bool
		wantErrType    error
	}{
		{
			name:           "successful generation for Spanish",
			tagline:        "Test Your Code",
			targetLang:     "es",
			serverResponse: mockGeminiResponse(sampleImageData, "image/png"),
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "successful generation for Portuguese",
			tagline:        "Cloud Native",
			targetLang:     "pt",
			serverResponse: mockGeminiResponse(sampleImageData, "image/jpeg"),
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:        "unsupported language",
			tagline:     "Test",
			targetLang:  "xx",
			wantErr:     true,
			wantErrType: ErrUnsupportedLang,
		},
		{
			name:           "API error response",
			tagline:        "Test",
			targetLang:     "es",
			serverResponse: mockGeminiErrorResponse(400, "Invalid request"),
			statusCode:     http.StatusBadRequest,
			wantErr:        true,
			wantErrType:    ErrAPIRequestFailed,
		},
		{
			name:           "empty response - no candidates",
			tagline:        "Test",
			targetLang:     "es",
			serverResponse: mockGeminiEmptyResponse(),
			statusCode:     http.StatusOK,
			wantErr:        true,
			wantErrType:    ErrNoImageGenerated,
		},
		{
			name:           "API error in successful HTTP response",
			tagline:        "Test",
			targetLang:     "es",
			serverResponse: mockGeminiErrorResponse(500, "Internal error"),
			statusCode:     http.StatusOK,
			wantErr:        true,
			wantErrType:    ErrAPIRequestFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}

				// Verify content type
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", ct)
				}

				// Verify URL contains model
				if !strings.Contains(r.URL.Path, "/models/") {
					t.Errorf("URL should contain /models/, got %s", r.URL.Path)
				}

				// Verify API key in query
				if r.URL.Query().Get("key") != "test-api-key" {
					t.Errorf("expected API key in query, got %s", r.URL.Query().Get("key"))
				}

				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.serverResponse))
			}))
			defer server.Close()

			// Create client with mock server
			config := Config{
				APIKey: "test-api-key",
				Model:  "gemini-3-pro-image-preview",
			}
			client := NewClientWithHTTPClient(config, server.Client(), server.URL)

			// Execute
			ctx := context.Background()
			result, err := client.GenerateLocalizedThumbnail(ctx, imagePath, tt.tagline, tt.targetLang)

			// Verify
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
					return
				}
				if tt.wantErrType != nil && !strings.Contains(err.Error(), tt.wantErrType.Error()) {
					t.Errorf("error = %v, want error containing %v", err, tt.wantErrType)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestGenerateLocalizedThumbnail_ImageNotFound(t *testing.T) {
	config := Config{
		APIKey: "test-api-key",
		Model:  "gemini-3-pro-image-preview",
	}
	client := NewClientWithHTTPClient(config, &http.Client{}, "")

	ctx := context.Background()
	_, err := client.GenerateLocalizedThumbnail(ctx, "/nonexistent/path/image.png", "Test", "es")

	if err == nil {
		t.Error("expected error for nonexistent image")
		return
	}

	if !strings.Contains(err.Error(), ErrImageReadFailed.Error()) {
		t.Errorf("error = %v, want error containing %v", err, ErrImageReadFailed)
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/path/to/image.png", "image/png"},
		{"/path/to/image.PNG", "image/png"},
		{"/path/to/image.jpg", "image/jpeg"},
		{"/path/to/image.jpeg", "image/jpeg"},
		{"/path/to/image.JPEG", "image/jpeg"},
		{"/path/to/image.webp", "image/webp"},
		{"/path/to/image.gif", "image/gif"},
		{"/path/to/image.unknown", "image/png"}, // default
		{"/path/to/image", "image/png"},         // no extension
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getMimeType(tt.path)
			if result != tt.expected {
				t.Errorf("getMimeType(%s) = %s, want %s", tt.path, result, tt.expected)
			}
		})
	}
}

func TestBuildPrompt(t *testing.T) {
	prompt := buildPrompt("Test Your Code", "Spanish")

	// Verify key elements are present
	if !strings.Contains(prompt, "Test Your Code") {
		t.Error("prompt should contain the tagline")
	}
	if !strings.Contains(prompt, "Spanish") {
		t.Error("prompt should contain the target language")
	}
	if !strings.Contains(prompt, "Translate") {
		t.Error("prompt should contain translation instruction")
	}
	if !strings.Contains(prompt, "same") {
		t.Error("prompt should contain instruction to keep other elements same")
	}
}

func TestGetLanguageName(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"es", "Spanish"},
		{"pt", "Portuguese"},
		{"de", "German"},
		{"fr", "French"},
		{"it", "Italian"},
		{"ja", "Japanese"},
		{"ko", "Korean"},
		{"zh", "Chinese"},
		{"xx", ""},       // unsupported
		{"", ""},         // empty
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := GetLanguageName(tt.code)
			if result != tt.expected {
				t.Errorf("GetLanguageName(%s) = %s, want %s", tt.code, result, tt.expected)
			}
		})
	}
}

func TestIsSupportedLanguage(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{"es", true},
		{"pt", true},
		{"de", true},
		{"fr", true},
		{"xx", false},
		{"", false},
		{"en", false}, // English is source, not target
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := IsSupportedLanguage(tt.code)
			if result != tt.expected {
				t.Errorf("IsSupportedLanguage(%s) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

func TestEncodeImageToBase64(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		// Create test file
		testData := []byte("test image content")
		testPath := filepath.Join(tmpDir, "test.png")
		if err := os.WriteFile(testPath, testData, 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		encoded, mimeType, err := encodeImageToBase64(testPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if mimeType != "image/png" {
			t.Errorf("mimeType = %s, want image/png", mimeType)
		}

		// Verify encoding
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			t.Fatalf("failed to decode: %v", err)
		}
		if string(decoded) != string(testData) {
			t.Error("decoded data doesn't match original")
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, _, err := encodeImageToBase64("/nonexistent/file.png")
		if err == nil {
			t.Error("expected error for nonexistent file")
		}
	})
}

func TestExtractImageFromResponse(t *testing.T) {
	sampleData := "dGVzdCBpbWFnZSBkYXRh" // base64 of "test image data"

	tests := []struct {
		name    string
		resp    *geminiResponse
		wantErr bool
	}{
		{
			name: "success - image in first part",
			resp: &geminiResponse{
				Candidates: []geminiCandidate{
					{
						Content: geminiContent{
							Parts: []geminiPart{
								{
									InlineData: &geminiInlineData{
										MimeType: "image/png",
										Data:     sampleData,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success - image in second part after text",
			resp: &geminiResponse{
				Candidates: []geminiCandidate{
					{
						Content: geminiContent{
							Parts: []geminiPart{
								{Text: "Here is your image"},
								{
									InlineData: &geminiInlineData{
										MimeType: "image/jpeg",
										Data:     sampleData,
									},
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "error - no candidates",
			resp: &geminiResponse{
				Candidates: []geminiCandidate{},
			},
			wantErr: true,
		},
		{
			name: "error - no image in parts",
			resp: &geminiResponse{
				Candidates: []geminiCandidate{
					{
						Content: geminiContent{
							Parts: []geminiPart{
								{Text: "Just text, no image"},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "error - invalid base64",
			resp: &geminiResponse{
				Candidates: []geminiCandidate{
					{
						Content: geminiContent{
							Parts: []geminiPart{
								{
									InlineData: &geminiInlineData{
										MimeType: "image/png",
										Data:     "not-valid-base64!!!",
									},
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractImageFromResponse(tt.resp)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) == 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}
