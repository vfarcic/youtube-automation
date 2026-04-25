package thumbnail

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewGPTImageClient(t *testing.T) {
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
			model:   "gpt-image-1",
			wantErr: nil,
		},
		{
			name:    "empty API key",
			apiKey:  "",
			model:   "gpt-image-1",
			wantErr: ErrGPTImageNoAPIKey,
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
			client, err := NewGPTImageClient(tt.apiKey, tt.model, tt.httpClient)

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

			if client.Name() != "gpt-image" {
				t.Errorf("Name() = %q, want %q", client.Name(), "gpt-image")
			}
			if client.model != tt.model {
				t.Errorf("model = %q, want %q", client.model, tt.model)
			}
		})
	}
}

func TestNewGPTImageClient_DefaultTimeout(t *testing.T) {
	client, err := NewGPTImageClient("test-key", "test-model", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.httpClient.Timeout != gptImageHTTPTimeout {
		t.Errorf("default HTTP timeout = %v, want %v", client.httpClient.Timeout, gptImageHTTPTimeout)
	}
}

func TestGPTImageClient_GenerateImage(t *testing.T) {
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
				// Verify Authorization header
				auth := r.Header.Get("Authorization")
				if auth != "Bearer test-key" {
					t.Errorf("Authorization = %q, want %q", auth, "Bearer test-key")
				}

				// Verify it's multipart form data
				ct := r.Header.Get("Content-Type")
				if !strings.HasPrefix(ct, "multipart/form-data") {
					t.Errorf("Content-Type = %q, want multipart/form-data", ct)
				}

				// Parse multipart form
				mediaType, params, err := mime.ParseMediaType(ct)
				if err != nil {
					t.Errorf("failed to parse content type: %v", err)
				}
				if mediaType != "multipart/form-data" {
					t.Errorf("media type = %q, want multipart/form-data", mediaType)
				}

				reader := multipart.NewReader(r.Body, params["boundary"])
				fields := make(map[string]string)
				var fileCount int
				for {
					part, err := reader.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Errorf("error reading part: %v", err)
						break
					}
					if part.FileName() != "" {
						fileCount++
					} else {
						data, _ := io.ReadAll(part)
						fields[part.FormName()] = string(data)
					}
					part.Close()
				}

				if fields["model"] != "test-model" {
					t.Errorf("model = %q, want %q", fields["model"], "test-model")
				}
				if fields["prompt"] != "Generate a thumbnail" {
					t.Errorf("prompt = %q, want %q", fields["prompt"], "Generate a thumbnail")
				}
				if fields["size"] != "1536x1024" {
					t.Errorf("size = %q, want %q", fields["size"], "1536x1024")
				}
				if fields["quality"] != "high" {
					t.Errorf("quality = %q, want %q", fields["quality"], "high")
				}
				if fileCount != 1 {
					t.Errorf("file count = %d, want 1", fileCount)
				}

				resp := gptImageResponse{
					Data: []gptImageData{
						{B64JSON: encodedImage},
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
				ct := r.Header.Get("Content-Type")
				_, params, _ := mime.ParseMediaType(ct)
				reader := multipart.NewReader(r.Body, params["boundary"])

				var fileCount int
				var fileNames []string
				for {
					part, err := reader.NextPart()
					if err == io.EOF {
						break
					}
					if err != nil {
						break
					}
					if part.FileName() != "" {
						fileCount++
						fileNames = append(fileNames, part.FileName())
					}
					part.Close()
				}

				if fileCount != 2 {
					t.Errorf("file count = %d, want 2", fileCount)
				}
				// Verify file extensions match detected MIME types
				if len(fileNames) >= 2 {
					if !strings.HasSuffix(fileNames[0], ".jpg") {
						t.Errorf("photo 0 filename = %q, want .jpg suffix", fileNames[0])
					}
					if !strings.HasSuffix(fileNames[1], ".png") {
						t.Errorf("photo 1 filename = %q, want .png suffix", fileNames[1])
					}
				}

				resp := gptImageResponse{
					Data: []gptImageData{
						{B64JSON: encodedImage},
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
			wantErr: ErrGPTImageEmptyPrompt,
		},
		{
			name:    "no photos",
			prompt:  "Generate a thumbnail",
			photos:  [][]byte{},
			handler: nil,
			wantErr: ErrGPTImageNoPhotos,
		},
		{
			name:    "nil photos",
			prompt:  "Generate a thumbnail",
			photos:  nil,
			handler: nil,
			wantErr: ErrGPTImageNoPhotos,
		},
		{
			name:   "API returns HTTP error",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":{"message":"rate limit exceeded","type":"rate_limit_error","code":"rate_limit_exceeded"}}`))
			},
			wantErr: ErrGPTImageAPIError,
		},
		{
			name:   "non-200 content policy violation",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":{"message":"Your request was rejected as a result of our safety system.","type":"invalid_request_error","code":"content_policy_violation"}}`))
			},
			wantErr: ErrGPTImageContentFiltered,
		},
		{
			name:   "non-200 unparseable body",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`not json at all`))
			},
			wantErr: ErrGPTImageAPIError,
		},
		{
			name:   "API returns error in body",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := gptImageResponse{
					Error: &gptImageError{
						Message: "invalid request",
						Type:    "invalid_request_error",
						Code:    "invalid_api_key",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGPTImageAPIError,
		},
		{
			name:   "content policy violation",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := gptImageResponse{
					Error: &gptImageError{
						Message: "Your request was rejected as a result of our safety system.",
						Type:    "invalid_request_error",
						Code:    "content_policy_violation",
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGPTImageContentFiltered,
		},
		{
			name:   "empty data array",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := gptImageResponse{
					Data: []gptImageData{},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGPTImageNoImage,
		},
		{
			name:   "empty b64_json in data",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := gptImageResponse{
					Data: []gptImageData{
						{B64JSON: ""},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErr: ErrGPTImageNoImage,
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
			wantErr: ErrGPTImageUnexpectedContentType,
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
			wantErrStr: "parsing openai response",
		},
		{
			name:   "invalid base64 in response",
			prompt: "Generate a thumbnail",
			photos: [][]byte{{0xFF}},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := gptImageResponse{
					Data: []gptImageData{
						{B64JSON: "not-valid-base64!!!"},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(resp)
			},
			wantErrStr: "decoding openai image data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.handler != nil {
				server = httptest.NewServer(tt.handler)
				defer server.Close()
			}

			client, err := NewGPTImageClient("test-key", "test-model", nil)
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

func TestGPTImageClient_GenerateImage_SanitizedErrors(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.HandlerFunc
		wantContains string
		wantAbsent   string
	}{
		{
			name: "parsed error uses message and code only",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"error":{"message":"invalid image format","type":"invalid_request_error","code":"bad_request"}}`))
			},
			wantContains: "invalid image format (code: bad_request)",
			wantAbsent:   `"type"`,
		},
		{
			name: "unparseable body omits raw content",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadGateway)
				w.Write([]byte(`<html>secret internal proxy error</html>`))
			},
			wantContains: "HTTP 502",
			wantAbsent:   "secret internal proxy error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client, _ := NewGPTImageClient("test-key", "test-model", nil)
			client.baseURL = server.URL

			_, err := client.GenerateImage(context.Background(), "test prompt", [][]byte{{0xFF}})
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			errMsg := err.Error()
			if !strings.Contains(errMsg, tt.wantContains) {
				t.Errorf("error %q should contain %q", errMsg, tt.wantContains)
			}
			if tt.wantAbsent != "" && strings.Contains(errMsg, tt.wantAbsent) {
				t.Errorf("error %q should NOT contain %q", errMsg, tt.wantAbsent)
			}
		})
	}
}

func TestGPTImageClient_GenerateImage_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	client, err := NewGPTImageClient("test-key", "test-model", nil)
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

func TestGPTImageClient_AuthorizationHeader(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		resp := gptImageResponse{
			Data: []gptImageData{
				{B64JSON: base64.StdEncoding.EncodeToString([]byte("img"))},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewGPTImageClient("my-secret-key", "gpt-image-1", nil)
	client.baseURL = server.URL

	_, err := client.GenerateImage(context.Background(), "test", [][]byte{{0xFF}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Bearer my-secret-key"
	if capturedAuth != expected {
		t.Errorf("Authorization = %q, want %q", capturedAuth, expected)
	}
}

func TestGPTImageClient_RequestMethod(t *testing.T) {
	var capturedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		resp := gptImageResponse{
			Data: []gptImageData{
				{B64JSON: base64.StdEncoding.EncodeToString([]byte("img"))},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewGPTImageClient("test-key", "test-model", nil)
	client.baseURL = server.URL

	client.GenerateImage(context.Background(), "test", [][]byte{{0xFF}})

	if capturedMethod != http.MethodPost {
		t.Errorf("Method = %q, want %q", capturedMethod, http.MethodPost)
	}
}

func TestExtensionFromMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		want     string
	}{
		{"image/png", "png"},
		{"image/webp", "webp"},
		{"image/jpeg", "jpg"},
		{"image/gif", "jpg"}, // unknown defaults to jpg
		{"", "jpg"},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			got := extensionFromMimeType(tt.mimeType)
			if got != tt.want {
				t.Errorf("extensionFromMimeType(%q) = %q, want %q", tt.mimeType, got, tt.want)
			}
		})
	}
}

func TestGPTImageClient_ImplementsInterface(t *testing.T) {
	// Compile-time check that GPTImageClient implements ImageGenerator.
	var _ ImageGenerator = (*GPTImageClient)(nil)
}
