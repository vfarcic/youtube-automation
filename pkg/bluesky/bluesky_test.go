package bluesky

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCreateBlueskyPostWithYouTubeThumbnail(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Mock authentication response
			session := Session{
				AccessJWT:  "test-jwt",
				RefreshJWT: "test-refresh-jwt",
				Handle:     "test.bsky.social",
				DID:        "did:test",
			}
			json.NewEncoder(w).Encode(session)
			return
		}

		if r.URL.Path == "/com.atproto.repo.createRecord" {
			// Parse the request body
			var req createPostRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request: %v", err)
			}

			// Verify the post content
			expectedText := "Test post https://youtu.be/test123"
			if req.Record.Text != expectedText {
				t.Errorf("Unexpected post text: %s", req.Record.Text)
			}

			// Verify the embed
			if req.Record.Embed == nil {
				t.Error("Expected embed to be present")
			} else {
				embed := req.Record.Embed
				if embed.Type != "app.bsky.embed.external" {
					t.Errorf("Unexpected embed type: %s", embed.Type)
				}
				if embed.External.URI != "https://youtu.be/test123" {
					t.Errorf("Unexpected embed URI: %s", embed.External.URI)
				}
				if embed.External.Thumb != "https://img.youtube.com/vi/test123/maxresdefault.jpg" {
					t.Errorf("Unexpected thumbnail URL: %s", embed.External.Thumb)
				}
			}

			// Mock the response with a post URI
			response := map[string]string{
				"uri": "at://did:test/app.bsky.feed.post/3k7qmjev5lr2s",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Create test post
	post := Post{
		Text:       "Test post https://youtu.be/test123",
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Create the post
	postURL, err := CreatePost(config, post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}

	expectedURL := "https://bsky.app/profile/test.bsky.social/post/3k7qmjev5lr2s"
	if postURL != expectedURL {
		t.Errorf("Expected post URL %s, got %s", expectedURL, postURL)
	}
}

func TestSendPost(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Mock authentication response
			session := Session{
				AccessJWT:  "test-jwt",
				RefreshJWT: "test-refresh-jwt",
				Handle:     "test.bsky.social",
				DID:        "did:test",
			}
			json.NewEncoder(w).Encode(session)
			return
		}

		if r.URL.Path == "/com.atproto.repo.createRecord" {
			// Parse the request body
			var req createPostRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request: %v", err)
			}

			// Verify the post content
			expectedText := "Test post https://youtu.be/test123"
			if req.Record.Text != expectedText {
				t.Errorf("Unexpected post text: %s", req.Record.Text)
			}

			// Verify the embed
			if req.Record.Embed == nil {
				t.Error("Expected embed to be present")
			} else {
				embed := req.Record.Embed
				if embed.Type != "app.bsky.embed.external" {
					t.Errorf("Unexpected embed type: %s", embed.Type)
				}
				if embed.External.URI != "https://youtu.be/test123" {
					t.Errorf("Unexpected embed URI: %s", embed.External.URI)
				}
				if embed.External.Thumb != "https://img.youtube.com/vi/test123/maxresdefault.jpg" {
					t.Errorf("Unexpected thumbnail URL: %s", embed.External.Thumb)
				}
			}

			// Mock the response with a post URI
			response := map[string]string{
				"uri": "at://did:test/app.bsky.feed.post/3k7qmjev5lr2s",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Test posting
	err := SendPost(config, "Test post [YOUTUBE]", "test123")
	if err != nil {
		t.Fatalf("Failed to post to Bluesky: %v", err)
	}
}

func TestSendPostValidation(t *testing.T) {
	// Create test config
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        "http://example.com",
	}

	tests := []struct {
		name     string
		text     string
		videoID  string
		expected string
	}{
		{
			name:     "Missing YouTube placeholder",
			text:     "Test post without placeholder",
			videoID:  "test123",
			expected: "text does not contain [YOUTUBE] placeholder",
		},
		{
			name:     "Missing video ID",
			text:     "Test post [YOUTUBE]",
			videoID:  "",
			expected: "YouTube video ID is required",
		},
		{
			name:     "Text too long",
			text:     strings.Repeat("a", 301) + " [YOUTUBE]",
			videoID:  "test123",
			expected: "text exceeds Bluesky's 300 character limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendPost(config, tt.text, tt.videoID)
			if err == nil {
				t.Error("Expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}

// TestAuthenticationFailure tests error handling when authentication fails
func TestAuthenticationFailure(t *testing.T) {
	// Create a test server that returns an error for authentication
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Return a 401 Unauthorized error
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "InvalidLogin", "message": "Invalid identifier or password"}`))
			return
		}
		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config with invalid credentials
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "wrong-password",
		URL:        server.URL,
	}

	// Create test post
	post := Post{
		Text:       "Test post https://youtu.be/test123",
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Attempt to create the post, should fail with auth error
	_, err := CreatePost(config, post)
	if err == nil {
		t.Fatalf("Expected authentication failure, but got success")
	}

	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("Expected error message to contain 'authentication failed', got: %s", err.Error())
	}
}

// TestRateLimiting tests the scenario when the API returns a rate limit error
func TestRateLimiting(t *testing.T) {
	// Create a test server that returns a rate limit error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Return successful auth
			session := Session{
				AccessJWT:  "test-jwt",
				RefreshJWT: "test-refresh-jwt",
				Handle:     "test.bsky.social",
				DID:        "did:test",
			}
			json.NewEncoder(w).Encode(session)
			return
		}

		if r.URL.Path == "/com.atproto.repo.createRecord" {
			// Return a 429 Too Many Requests error
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": "RateLimitExceeded", "message": "Too many requests. Try again later."}`))
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Create test post
	post := Post{
		Text:       "Test post https://youtu.be/test123",
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Attempt to create the post, should fail with rate limit error
	_, err := CreatePost(config, post)
	if err == nil {
		t.Fatalf("Expected rate limit failure, but got success")
	}

	if !strings.Contains(err.Error(), "post creation failed with status 429") {
		t.Errorf("Expected error message to contain rate limit status code, got: %s", err.Error())
	}
}

// TestNetworkFailure tests error handling when the network connection fails
func TestNetworkFailure(t *testing.T) {
	// Create a server that will immediately close the connection
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Close the connection immediately to simulate network failure
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatalf("webserver doesn't support hijacking")
		}
		conn, _, err := hj.Hijack()
		if err != nil {
			t.Fatalf("failed to hijack connection: %v", err)
		}
		conn.Close()
	}))
	server.Close() // Close the server immediately to ensure connection will fail

	// Use the closed server's URL to ensure network failure
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Create test post
	post := Post{
		Text:       "Test post https://youtu.be/test123",
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Attempt to create the post, should fail with network error
	_, err := CreatePost(config, post)
	if err == nil {
		t.Fatalf("Expected network failure, but got success")
	}
}

// TestServerError tests error handling when the server returns an internal error
func TestServerError(t *testing.T) {
	// Create a test server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Return successful auth
			session := Session{
				AccessJWT:  "test-jwt",
				RefreshJWT: "test-refresh-jwt",
				Handle:     "test.bsky.social",
				DID:        "did:test",
			}
			json.NewEncoder(w).Encode(session)
			return
		}

		if r.URL.Path == "/com.atproto.repo.createRecord" {
			// Return a 500 Internal Server Error
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "InternalServerError", "message": "Something went wrong"}`))
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Create test post
	post := Post{
		Text:       "Test post https://youtu.be/test123",
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Attempt to create the post, should fail with server error
	_, err := CreatePost(config, post)
	if err == nil {
		t.Fatalf("Expected server error, but got success")
	}

	if !strings.Contains(err.Error(), "post creation failed with status 500") {
		t.Errorf("Expected error message to contain server error status code, got: %s", err.Error())
	}
}

// TestInvalidResponseFormat tests error handling when the server returns invalid JSON
func TestInvalidResponseFormat(t *testing.T) {
	// Create a test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Return successful auth
			session := Session{
				AccessJWT:  "test-jwt",
				RefreshJWT: "test-refresh-jwt",
				Handle:     "test.bsky.social",
				DID:        "did:test",
			}
			json.NewEncoder(w).Encode(session)
			return
		}

		if r.URL.Path == "/com.atproto.repo.createRecord" {
			// Return invalid JSON
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"uri": "at://did:test/app.bsky.feed.post/3k7qmjev5lr2s"`)) // Missing closing bracket
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Create test post
	post := Post{
		Text:       "Test post https://youtu.be/test123",
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Attempt to create the post, should fail with parsing error
	_, err := CreatePost(config, post)
	if err == nil {
		t.Fatalf("Expected parsing error, but got success")
	}

	if !strings.Contains(err.Error(), "error decoding response") {
		t.Errorf("Expected error message to contain decoding error, got: %s", err.Error())
	}
}

// TestConfigValidation tests the ValidateConfig function more thoroughly
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		shouldError bool
		errorSubstr string
	}{
		{
			name: "Valid config",
			config: Config{
				Identifier: "test.bsky.social",
				Password:   "valid-password",
				URL:        "https://example.com",
			},
			shouldError: false,
		},
		{
			name: "Missing identifier",
			config: Config{
				Identifier: "",
				Password:   "valid-password",
				URL:        "https://example.com",
			},
			shouldError: true,
			errorSubstr: "Bluesky credentials not configured",
		},
		{
			name: "Missing password",
			config: Config{
				Identifier: "test.bsky.social",
				Password:   "",
				URL:        "https://example.com",
			},
			shouldError: true,
			errorSubstr: "Bluesky password is required",
		},
		{
			name: "Both identifier and password missing",
			config: Config{
				Identifier: "",
				Password:   "",
				URL:        "https://example.com",
			},
			shouldError: false, // This is fine, means Bluesky is not being used
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)

			if tt.shouldError && err == nil {
				t.Errorf("Expected error but got nil")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if tt.shouldError && err != nil && !strings.Contains(err.Error(), tt.errorSubstr) {
				t.Errorf("Expected error message to contain '%s', got: %s", tt.errorSubstr, err.Error())
			}
		})
	}
}

// TestGetConfig tests the GetConfig function with environment variable
func TestGetConfig(t *testing.T) {
	// Save original env var
	origEnv := os.Getenv("BLUESKY_PASSWORD")
	defer os.Setenv("BLUESKY_PASSWORD", origEnv)

	// Test with env var set
	os.Setenv("BLUESKY_PASSWORD", "env-password")

	config := GetConfig("test.user", "default-password", "https://example.com")

	if config.Password != "env-password" {
		t.Errorf("Expected password to be from env var 'env-password', got: %s", config.Password)
	}

	// Test with env var unset
	os.Setenv("BLUESKY_PASSWORD", "")

	config = GetConfig("test.user", "default-password", "https://example.com")

	if config.Password != "default-password" {
		t.Errorf("Expected password to be 'default-password', got: %s", config.Password)
	}
}

// TestCreatePostLongText tests handling of long post texts
func TestCreatePostLongText(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Return successful auth
			session := Session{
				AccessJWT:  "test-jwt",
				RefreshJWT: "test-refresh-jwt",
				Handle:     "test.bsky.social",
				DID:        "did:test",
			}
			json.NewEncoder(w).Encode(session)
			return
		}

		if r.URL.Path == "/com.atproto.repo.createRecord" {
			// Parse the request body
			var req createPostRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode request: %v", err)
			}

			// Verify the text length is within limit
			if len(req.Record.Text) > 300 {
				t.Errorf("Text length exceeds 300 characters: %d", len(req.Record.Text))
			}

			// Check for truncation
			if !strings.HasSuffix(req.Record.Text, "...") {
				t.Errorf("Expected truncated text to end with '...', got: %s", req.Record.Text)
			}

			// Mock the response with a post URI
			response := map[string]string{
				"uri": "at://did:test/app.bsky.feed.post/3k7qmjev5lr2s",
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config
	config := Config{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Create a post with text that exceeds the limit
	post := Post{
		Text:       strings.Repeat("a", 310) + " https://youtu.be/test123", // Exceeds 300 chars
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Create the post
	_, err := CreatePost(config, post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}
}
