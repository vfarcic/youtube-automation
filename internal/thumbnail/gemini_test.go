package thumbnail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewGeminiClient(t *testing.T) {
	tests := []struct {
		name       string
		apiKey     string
		model      string
		httpClient *http.Client
		wantErr    error
	}{
		{
			name:    "valid config",
			apiKey:  "test-key",
			model:   "gemini-2.0-flash-preview-image-generation",
			wantErr: nil,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			model:   "gemini-2.0-flash-preview-image-generation",
			wantErr: ErrGeminiNoAPIKey,
		},
		{
			name:       "custom HTTP client",
			apiKey:     "test-key",
			model:      "test-model",
			httpClient: &http.Client{},
			wantErr:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewGeminiClient(tt.apiKey, tt.model, tt.httpClient)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if client.Name() != "gemini" {
				t.Errorf("Name() = %q, want %q", client.Name(), "gemini")
			}
			if client.model != tt.model {
				t.Errorf("model = %q, want %q", client.model, tt.model)
			}
		})
	}
}

func TestNewGeminiClient_DefaultTimeout(t *testing.T) {
	client, err := NewGeminiClient("test-key", "test-model", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.httpClient.Timeout != geminiHTTPTimeout {
		t.Errorf("default HTTP timeout = %v, want %v", client.httpClient.Timeout, geminiHTTPTimeout)
	}
}

func TestGeminiClient_GenerateImage(t *testing.T) {
	fakeImageData := []byte("fake-png-image-bytes")
	encodedImage := base64.StdEncoding.EncodeToString(fakeImageData)

	tests := []struct {
		name       string
		prompt     string
		photos     [][]byte
		handler    http.HandlerFunc
		wantErr    error
		wantErrStr string
		wantImage  []byte
	}{
		{
			name:   "successful generation",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF, 0xD8, 0xFF, 0xE0}}, // JPEG magic bytes
			handler: func(w http.ResponseWriter, r *http.Request) {
				// Verify request structure
				var req geminiRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request: %v", err)
				}
				if len(req.Contents) != 1 {
					t.Errorf("expected 1 content, got %d", len(req.Contents))
				}
				if len(req.Contents[0].Parts) != 2 {
					t.Errorf("expected 2 parts (text + photo), got %d", len(req.Contents[0].Parts))
				}
				if req.Contents[0].Parts[0].Text != "Generate a thumbnail" {
					t.Errorf("unexpected prompt text: %s", req.Contents[0].Parts[0].Text)
				}
				if req.Contents[0].Parts[1].InlineData.MimeType != "image/jpeg" {
					t.Errorf("expected image/jpeg, got %s", req.Contents[0].Parts[1].InlineData.MimeType)
				}
				if req.GenerationConfig.ImageConfig.AspectRatio != "16:9" {
					t.Errorf("expected 16:9 aspect ratio, got %s", req.GenerationConfig.ImageConfig.AspectRatio)
				}

				resp := geminiResponse{
					Candidates: []geminiCandidate{
						{
							Content: geminiContent{
								Parts: []geminiPart{
									{InlineData: &geminiInlineData{
										MimeType: "image/png",
										Data:     encodedImage,
									}},
								},
							},
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantImage: fakeImageData,
		},
		{
			name:   "multiple photos",
			prompt: "Generate a thumbnail",
			photos: [][]byte{
				{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG
				{0x89, 0x50, 0x4E, 0x47}, // PNG
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				var req geminiRequest
				json.NewDecoder(r.Body).Decode(&req)
				if len(req.Contents[0].Parts) != 3 {
					t.Errorf("expected 3 parts (text + 2 photos), got %d", len(req.Contents[0].Parts))
				}
				if req.Contents[0].Parts[1].InlineData.MimeType != "image/jpeg" {
					t.Errorf("photo 1: expected image/jpeg, got %s", req.Contents[0].Parts[1].InlineData.MimeType)
				}
				if req.Contents[0].Parts[2].InlineData.MimeType != "image/png" {
					t.Errorf("photo 2: expected image/png, got %s", req.Contents[0].Parts[2].InlineData.MimeType)
				}

				resp := geminiResponse{
					Candidates: []geminiCandidate{
						{Content: geminiContent{Parts: []geminiPart{
							{InlineData: &geminiInlineData{MimeType: "image/png", Data: encodedImage}},
						}}},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantImage: fakeImageData,
		},
		{
			name:    "empty prompt",
			prompt:  "",
			photos:  [][]byte{{0xFF}},
			handler: nil,
			wantErr: ErrGeminiEmptyPrompt,
		},
		{
			name:    "no photos",
			prompt:  "Generate a thumbnail",
			photos:  [][]byte{},
			handler: nil,
			wantErr: ErrGeminiNoPhotos,
		},
		{
			name:    "nil photos",
			prompt:  "Generate a thumbnail",
			photos:  nil,
			handler: nil,
			wantErr: ErrGeminiNoPhotos,
		},
		{
			name:   "API returns HTTP error",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":{"code":429,"message":"rate limit exceeded","status":"RESOURCE_EXHAUSTED"}}`))
			},
			wantErr: ErrGeminiAPIError,
		},
		{
			name:   "API returns error in body",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := geminiResponse{
					Error: &geminiError{
						Code:    400,
						Message: "invalid request",
						Status:  "INVALID_ARGUMENT",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGeminiAPIError,
		},
		{
			name:   "no image in response",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := geminiResponse{
					Candidates: []geminiCandidate{
						{Content: geminiContent{Parts: []geminiPart{
							{Text: "I cannot generate this image"},
						}}},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGeminiNoImage,
		},
		{
			name:   "empty candidates",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := geminiResponse{
					Candidates: []geminiCandidate{},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGeminiNoImage,
		},
		{
			name:   "safety filtered response",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := geminiResponse{
					Candidates: []geminiCandidate{
						{
							Content:      geminiContent{Parts: []geminiPart{}},
							FinishReason: "SAFETY",
						},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGeminiImageFiltered,
		},
		{
			name:   "unexpected content type",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("<html>error</html>"))
			},
			wantErr: ErrGeminiUnexpectedContentType,
		},
		{
			name:   "invalid JSON response",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("not json"))
			},
			wantErrStr: "parsing gemini response",
		},
		{
			name:   "invalid base64 in response",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := geminiResponse{
					Candidates: []geminiCandidate{
						{Content: geminiContent{Parts: []geminiPart{
							{InlineData: &geminiInlineData{MimeType: "image/png", Data: "not-valid-base64!!!"}},
						}}},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErrStr: "decoding gemini image data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.handler != nil {
				server = httptest.NewServer(tt.handler)
				defer server.Close()
			}

			client, err := NewGeminiClient("test-key", "test-model", nil)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			if server != nil {
				client.baseURL = server.URL
			}

			got, err := client.GenerateImage(context.Background(), tt.prompt, tt.photos)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if tt.wantErrStr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErrStr)
				}
				if !strings.Contains(err.Error(), tt.wantErrStr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErrStr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != string(tt.wantImage) {
				t.Errorf("image bytes mismatch: got %d bytes, want %d bytes", len(got), len(tt.wantImage))
			}
		})
	}
}

func TestGeminiClient_GenerateImage_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until request context is canceled
		<-r.Context().Done()
	}))
	defer server.Close()

	client, err := NewGeminiClient("test-key", "test-model", nil)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	client.baseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err = client.GenerateImage(ctx, "test prompt", [][]byte{{0xFF}})
	if err == nil {
		t.Fatal("expected error from canceled context, got nil")
	}
}

func TestDetectMimeType(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{
			name: "JPEG magic bytes",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00},
			want: "image/jpeg",
		},
		{
			name: "PNG magic bytes",
			data: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A},
			want: "image/png",
		},
		{
			name: "WEBP magic bytes",
			data: []byte{'R', 'I', 'F', 'F', 0x00, 0x00, 0x00, 0x00, 'W', 'E', 'B', 'P'},
			want: "image/webp",
		},
		{
			name: "unknown defaults to JPEG",
			data: []byte{0x00, 0x01, 0x02, 0x03, 0x04},
			want: "image/jpeg",
		},
		{
			name: "short data defaults to JPEG",
			data: []byte{0x00, 0x01},
			want: "image/jpeg",
		},
		{
			name: "empty data defaults to JPEG",
			data: []byte{},
			want: "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectMimeType(tt.data)
			if got != tt.want {
				t.Errorf("detectMimeType() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGeminiClient_RequestURL(t *testing.T) {
	var capturedURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedURL = r.URL.String()
		resp := geminiResponse{
			Candidates: []geminiCandidate{
				{Content: geminiContent{Parts: []geminiPart{
					{InlineData: &geminiInlineData{
						MimeType: "image/png",
						Data:     base64.StdEncoding.EncodeToString([]byte("img")),
					}},
				}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewGeminiClient("my-api-key", "gemini-2.0-flash", nil)
	client.baseURL = server.URL

	client.GenerateImage(context.Background(), "test", [][]byte{{0xFF}})

	expected := "/gemini-2.0-flash:generateContent?key=my-api-key"
	if capturedURL != expected {
		t.Errorf("URL = %q, want %q", capturedURL, expected)
	}
}
