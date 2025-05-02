package bluesky

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
