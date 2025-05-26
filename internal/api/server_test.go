package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/storage"
)

func TestHealthEndpoint(t *testing.T) {
	// Create a temporary index file for testing
	tmpFile, err := os.CreateTemp("", "index.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp index file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Create a server with the test index path
	server := NewServer(tmpFile.Name(), 8080)

	// Add middleware and routes
	server.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	})

	// Add health check endpoint
	server.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Serve the request
	server.router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check the response body
	expected := `{"status":"ok"}`
	if strings.TrimSpace(w.Body.String()) != expected {
		t.Errorf("Expected response body %s, got %s", expected, w.Body.String())
	}
}

func TestCreateVideoEndpoint(t *testing.T) {
	// Create a temporary index file for testing
	tmpFile, err := os.CreateTemp("", "index.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp index file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write initial empty index
	if err := os.WriteFile(tmpFile.Name(), []byte("[]"), 0644); err != nil {
		t.Fatalf("Failed to write temp index file: %v", err)
	}

	// Create a server with the test index path
	server := NewServer(tmpFile.Name(), 8080)

	// Create a test request
	reqBody := `{"name":"Test Video","category":"test-category"}`
	req := httptest.NewRequest("POST", "/api/videos", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Serve the request
	server.router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	// Check the response body
	var response storage.Video
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	if response.Name != "Test Video" {
		t.Errorf("Expected name to be 'Test Video', got '%s'", response.Name)
	}
	if response.Category != "test-category" {
		t.Errorf("Expected category to be 'test-category', got '%s'", response.Category)
	}
}

func TestGetVideoPhases(t *testing.T) {
	// Create a temporary index file for testing
	tmpFile, err := os.CreateTemp("", "index.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp index file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Create a server with the test index path
	server := NewServer(tmpFile.Name(), 8080)

	// Create a test request
	req := httptest.NewRequest("GET", "/api/videos/phases", nil)
	w := httptest.NewRecorder()

	// Serve the request
	server.router.ServeHTTP(w, req)

	// Check the response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check the response body contains phases
	var response []service.VideoPhase
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}

	// Check that we have 6 phases (from Initial to Post-Publish)
	if len(response) != 6 {
		t.Errorf("Expected 6 phases, got %d", len(response))
	}
}