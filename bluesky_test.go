package main

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
			session := BlueskySession{
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
			if req.Record.Text != "Test post https://youtu.be/test123" {
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

			w.WriteHeader(http.StatusOK)
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Create test config
	config := BlueskyConfig{
		Identifier: "test.bsky.social",
		Password:   "test-password",
		URL:        server.URL,
	}

	// Create test post
	post := BlueskyPost{
		Text:       "Test post https://youtu.be/test123",
		YouTubeURL: "https://youtu.be/test123",
		VideoID:    "test123",
	}

	// Create the post
	err := CreateBlueskyPost(config, post)
	if err != nil {
		t.Fatalf("Failed to create post: %v", err)
	}
}

func TestPostToBluesky(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/com.atproto.server.createSession" {
			// Mock authentication response
			session := BlueskySession{
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
			expectedText := strings.ReplaceAll("Test post [YOUTUBE]", "[YOUTUBE]", "https://youtu.be/test123")
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

			w.WriteHeader(http.StatusOK)
			return
		}

		t.Errorf("Unexpected request path: %s", r.URL.Path)
	}))
	defer server.Close()

	// Set up test settings
	settings.Bluesky.Identifier = "test.bsky.social"
	settings.Bluesky.Password = "test-password"
	settings.Bluesky.URL = server.URL

	// Test posting
	err := PostToBluesky("Test post [YOUTUBE]", "test123")
	if err != nil {
		t.Fatalf("Failed to post to Bluesky: %v", err)
	}
}

func TestPostToBlueskyValidation(t *testing.T) {
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
			err := PostToBluesky(tt.text, tt.videoID)
			if err == nil {
				t.Error("Expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.expected, err.Error())
			}
		})
	}
}
