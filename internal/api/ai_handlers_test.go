package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AI Request/Response types are defined in handlers.go

// Test that AI endpoints are properly registered and handle requests
func TestAIEndpointsRouting(t *testing.T) {
	server := NewServer()

	// Test that all AI endpoints are properly registered (should return 400 for empty manuscript, not 404)
	endpoints := []string{
		"/api/ai/titles",
		"/api/ai/description",
		"/api/ai/tags",
		"/api/ai/tweets",
		"/api/ai/description-tags",
	}

	for _, endpoint := range endpoints {
		t.Run("Route exists for "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("POST", endpoint, bytes.NewBufferString(`{"manuscript": ""}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Should return 400 (bad request for empty manuscript), not 404 (route not found)
			assert.Equal(t, http.StatusBadRequest, w.Code, "Route should exist and return 400 for empty manuscript")
		})
	}
}

// Test JSON validation and error handling
func TestAIEndpointsValidation(t *testing.T) {
	server := NewServer()

	endpoints := []string{
		"/api/ai/titles",
		"/api/ai/description",
		"/api/ai/tags",
		"/api/ai/tweets",
		"/api/ai/description-tags",
	}

	for _, endpoint := range endpoints {
		t.Run("Invalid JSON for "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("POST", endpoint, bytes.NewBufferString(`{"invalid": json}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Empty manuscript for "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("POST", endpoint, bytes.NewBufferString(`{"manuscript": ""}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})

		t.Run("Missing manuscript for "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("POST", endpoint, bytes.NewBufferString(`{}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

// Test that wrong HTTP methods return 405
func TestAIEndpointsMethodNotAllowed(t *testing.T) {
	server := NewServer()

	endpoints := []string{
		"/api/ai/titles",
		"/api/ai/description",
		"/api/ai/tags",
		"/api/ai/tweets",
		"/api/ai/description-tags",
	}

	for _, endpoint := range endpoints {
		t.Run("GET not allowed for "+endpoint, func(t *testing.T) {
			req := httptest.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// Test request/response JSON structure
func TestAIEndpointsJSONStructure(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name     string
		endpoint string
		request  AIRequest
	}{
		{
			name:     "Valid JSON structure for titles",
			endpoint: "/api/ai/titles",
			request:  AIRequest{Manuscript: "Test manuscript content"},
		},
		{
			name:     "Valid JSON structure for description",
			endpoint: "/api/ai/description",
			request:  AIRequest{Manuscript: "Test manuscript content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", tt.endpoint, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// We expect either 500 (no AI config) or 200 (if AI config exists)
			// The important thing is that JSON parsing worked (not 400)
			assert.Contains(t, []int{http.StatusInternalServerError, http.StatusOK}, w.Code,
				"Should parse JSON correctly and either succeed or fail with AI config issue")
		})
	}
}

// Test specific endpoint response formats
func TestAITitlesEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		manuscript     string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid manuscript",
			manuscript:     "This is a comprehensive tutorial about React hooks and state management in modern web applications.",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AITitlesResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Titles, "Should return at least one title")
				for _, title := range resp.Titles {
					assert.NotEmpty(t, title, "Each title should not be empty")
					assert.True(t, len(title) > 10, "Title should be reasonably long")
				}
			},
		},
		{
			name:           "Short manuscript",
			manuscript:     "Quick tip about CSS.",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AITitlesResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.Titles, "Should handle short manuscripts")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: tt.manuscript})
			req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available (expected in test environment)
			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test description endpoint
func TestAIDescriptionEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		manuscript     string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid manuscript",
			manuscript:     "This tutorial covers advanced Docker concepts including multi-stage builds, networking, and orchestration.",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AIDescriptionResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Description, "Should return a description")
				assert.True(t, len(resp.Description) > 50, "Description should be substantial")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: tt.manuscript})
			req := httptest.NewRequest("POST", "/api/ai/description", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test tags endpoint
func TestAITagsEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		manuscript     string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid manuscript",
			manuscript:     "Learn Kubernetes deployment strategies, including blue-green and canary deployments for production environments.",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AITagsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Tags, "Should return tags")
				for _, tag := range resp.Tags {
					assert.NotEmpty(t, tag, "Each tag should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: tt.manuscript})
			req := httptest.NewRequest("POST", "/api/ai/tags", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test tweets endpoint
func TestAITweetsEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		manuscript     string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid manuscript",
			manuscript:     "Explore the latest features in Go 1.21 including improved performance and new standard library additions.",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AITweetsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Tweets, "Should return tweets")
				for _, tweet := range resp.Tweets {
					assert.NotEmpty(t, tweet, "Each tweet should not be empty")
					assert.True(t, len(tweet) <= 280, "Tweet should respect character limit")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: tt.manuscript})
			req := httptest.NewRequest("POST", "/api/ai/tweets", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}


// Test description-tags endpoint
func TestAIDescriptionTagsEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		manuscript     string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid manuscript",
			manuscript:     "Master microservices architecture with Spring Boot, including service discovery, load balancing, and distributed tracing.",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AIDescriptionTagsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.DescriptionTags, "Should return description tags")
				// Note: This endpoint returns a single string, not an array
				assert.Len(t, resp.DescriptionTags, 1, "Should return exactly one description-tags string")
				assert.NotEmpty(t, resp.DescriptionTags[0], "Description tags should not be empty")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: tt.manuscript})
			req := httptest.NewRequest("POST", "/api/ai/description-tags", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test edge cases and error scenarios
func TestAIEndpointsEdgeCases(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		endpoint       string
		manuscript     string
		expectedStatus int
		description    string
	}{
		{
			name:           "Very long manuscript",
			endpoint:       "/api/ai/titles",
			manuscript:     strings.Repeat("This is a very long manuscript content. ", 1000),
			expectedStatus: http.StatusOK,
			description:    "Should handle very long manuscripts",
		},
		{
			name:           "Manuscript with special characters",
			endpoint:       "/api/ai/description",
			manuscript:     "Tutorial about C++ templates & STL containers: vector<T>, map<K,V>, and unique_ptr<T>.",
			expectedStatus: http.StatusOK,
			description:    "Should handle special characters",
		},
		{
			name:           "Manuscript with unicode",
			endpoint:       "/api/ai/tags",
			manuscript:     "Learn about 日本語 programming concepts and émojis 🚀 in modern development.",
			expectedStatus: http.StatusOK,
			description:    "Should handle unicode characters",
		},
		{
			name:           "Minimal manuscript",
			endpoint:       "/api/ai/tweets",
			manuscript:     "AI tutorial.",
			expectedStatus: http.StatusOK,
			description:    "Should handle minimal content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: tt.manuscript})
			req := httptest.NewRequest("POST", tt.endpoint, bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)
		})
	}
}

// Test CORS headers
func TestAIEndpointsCORS(t *testing.T) {
	server := NewServer()

	req := httptest.NewRequest("OPTIONS", "/api/ai/titles", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

// Test content type validation
func TestAIEndpointsContentType(t *testing.T) {
	server := NewServer()

	// Test without Content-Type header
	req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBufferString(`{"manuscript": "test"}`))
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Should still work (Go's JSON decoder is lenient)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError, http.StatusOK}, w.Code)

	// Test with wrong Content-Type
	req = httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBufferString(`{"manuscript": "test"}`))
	req.Header.Set("Content-Type", "text/plain")
	w = httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Should handle gracefully
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError, http.StatusOK}, w.Code)
}

// Test additional error scenarios and edge cases
func TestAIEndpointsAdvancedErrorHandling(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		endpoint       string
		requestBody    string
		contentType    string
		expectedStatus int
		description    string
	}{
		{
			name:           "Malformed JSON - missing quotes",
			endpoint:       "/api/ai/titles",
			requestBody:    `{manuscript: "test content"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject malformed JSON",
		},
		{
			name:           "Malformed JSON - trailing comma",
			endpoint:       "/api/ai/description",
			requestBody:    `{"manuscript": "test content",}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject JSON with trailing comma",
		},
		{
			name:           "Empty request body",
			endpoint:       "/api/ai/tags",
			requestBody:    "",
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject empty request body",
		},
		{
			name:           "Non-JSON content type with JSON body",
			endpoint:       "/api/ai/tweets",
			requestBody:    `{"manuscript": "test content"}`,
			contentType:    "text/xml",
			expectedStatus: http.StatusBadRequest,
			description:    "Should handle content type mismatch gracefully",
		},
		{
			name:           "Very large request body",
			endpoint:       "/api/ai/titles",
			requestBody:    `{"manuscript": "` + strings.Repeat("Very long content. ", 10000) + `"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK, // Should handle large content
			description:    "Should handle very large manuscripts",
		},
		{
			name:           "Null manuscript value",
			endpoint:       "/api/ai/description-tags",
			requestBody:    `{"manuscript": null}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject null manuscript",
		},
		{
			name:           "Manuscript with only whitespace",
			endpoint:       "/api/ai/titles",
			requestBody:    `{"manuscript": "   \n\t   "}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject manuscript with only whitespace",
		},
		{
			name:           "Extra fields in request",
			endpoint:       "/api/ai/description",
			requestBody:    `{"manuscript": "test content", "extra_field": "should be ignored", "another": 123}`,
			contentType:    "application/json",
			expectedStatus: http.StatusOK, // Should ignore extra fields
			description:    "Should ignore extra fields in request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", tt.endpoint, bytes.NewBufferString(tt.requestBody))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip AI configuration errors in test environment
			if w.Code == http.StatusInternalServerError && tt.expectedStatus != http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code, tt.description)

			// Verify error responses have proper JSON format
			if w.Code >= 400 && w.Code < 500 {
				var errorResp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err, "Error response should be valid JSON")
				assert.Contains(t, errorResp, "error", "Error response should contain error field")
			}
		})
	}
}

// Test concurrent requests to AI endpoints
func TestAIEndpointsConcurrency(t *testing.T) {
	server := NewServer()

	// Test concurrent requests to the same endpoint
	t.Run("Concurrent requests to same endpoint", func(t *testing.T) {
		const numRequests = 10
		var wg sync.WaitGroup
		results := make(chan int, numRequests)

		for i := 0; i < numRequests; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				reqBody := fmt.Sprintf(`{"manuscript": "Test manuscript %d for concurrent processing"}`, id)
				req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBufferString(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.router.ServeHTTP(w, req)
				results <- w.Code
			}(i)
		}

		wg.Wait()
		close(results)

		// Collect results
		statusCodes := make(map[int]int)
		for code := range results {
			statusCodes[code]++
		}

		// All requests should either succeed (200) or fail with AI config error (500)
		// No requests should fail with validation errors (400) since all have valid input
		for code, count := range statusCodes {
			assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, code,
				"Concurrent requests should handle properly, got %d status codes with value %d", count, code)
		}
	})

	// Test concurrent requests to different endpoints
	t.Run("Concurrent requests to different endpoints", func(t *testing.T) {
		endpoints := []string{
			"/api/ai/titles",
			"/api/ai/description",
			"/api/ai/tags",
			"/api/ai/tweets",
				"/api/ai/description-tags",
		}

		var wg sync.WaitGroup
		results := make(chan struct {
			endpoint string
			status   int
		}, len(endpoints))

		for _, endpoint := range endpoints {
			wg.Add(1)
			go func(ep string) {
				defer wg.Done()

				reqBody := `{"manuscript": "Concurrent test manuscript for different endpoints"}`
				req := httptest.NewRequest("POST", ep, bytes.NewBufferString(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.router.ServeHTTP(w, req)
				results <- struct {
					endpoint string
					status   int
				}{ep, w.Code}
			}(endpoint)
		}

		wg.Wait()
		close(results)

		// Verify all endpoints handled concurrent requests properly
		for result := range results {
			assert.Contains(t, []int{http.StatusOK, http.StatusInternalServerError}, result.status,
				"Endpoint %s should handle concurrent requests properly", result.endpoint)
		}
	})
}

// Test AI endpoints with diverse manuscript content types and sizes
func TestAIEndpointsManuscriptContent(t *testing.T) {
	server := NewServer()

	// Test various content types and scenarios
	contentTests := map[string]string{
		"technical_tutorial": "This comprehensive tutorial covers advanced Docker containerization techniques, including multi-stage builds, layer optimization, security best practices, and orchestration with Kubernetes. We'll explore practical examples and real-world deployment scenarios.",
		"programming_guide":  "Learn Python's advanced features including decorators, context managers, metaclasses, and async programming. This guide provides hands-on examples and best practices for writing clean, efficient Python code.",
		"devops_workflow":    "Setting up CI/CD pipelines with GitHub Actions, automated testing, deployment strategies, monitoring, and infrastructure as code using Terraform and Ansible.",
		"web_development":    "Building modern web applications with React, TypeScript, and Node.js. Covers component architecture, state management, API integration, and performance optimization techniques.",
		"cloud_architecture": "Designing scalable cloud solutions on AWS, including microservices architecture, serverless computing, database design, and cost optimization strategies.",
		"security_practices": "Implementing cybersecurity best practices in software development, including secure coding, vulnerability assessment, penetration testing, and compliance frameworks.",
	}

	endpoints := []struct {
		path     string
		name     string
		validate func(t *testing.T, body []byte)
	}{
		{
			path: "/api/ai/titles",
			name: "titles",
			validate: func(t *testing.T, body []byte) {
				var resp AITitlesResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Titles, "Should return titles")
				assert.True(t, len(resp.Titles) >= 1, "Should return at least one title")
				for _, title := range resp.Titles {
					assert.NotEmpty(t, title, "Each title should not be empty")
					assert.True(t, len(title) > 5, "Title should be reasonably long")
					assert.True(t, len(title) < 200, "Title should not be excessively long")
				}
			},
		},
		{
			path: "/api/ai/description",
			name: "description",
			validate: func(t *testing.T, body []byte) {
				var resp AIDescriptionResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Description, "Should return description")
				assert.True(t, len(resp.Description) > 20, "Description should be substantial")
			},
		},
		{
			path: "/api/ai/tags",
			name: "tags",
			validate: func(t *testing.T, body []byte) {
				var resp AITagsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Tags, "Should return tags")
				for _, tag := range resp.Tags {
					assert.NotEmpty(t, tag, "Each tag should not be empty")
					assert.True(t, len(tag) > 1, "Tag should be reasonably long")
				}
			},
		},
	}

	for _, endpoint := range endpoints {
		for domain, manuscript := range contentTests {
			t.Run(fmt.Sprintf("%s_%s", endpoint.name, domain), func(t *testing.T) {
				reqBody, _ := json.Marshal(AIRequest{Manuscript: manuscript})
				req := httptest.NewRequest("POST", endpoint.path, bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.router.ServeHTTP(w, req)

				// Skip validation if AI config is not available
				if w.Code == http.StatusInternalServerError {
					t.Skip("AI configuration not available in test environment")
				}

				assert.Equal(t, http.StatusOK, w.Code, "Should handle %s content for %s", domain, endpoint.name)

				if w.Code == http.StatusOK && endpoint.validate != nil {
					endpoint.validate(t, w.Body.Bytes())
				}
			})
		}
	}
}

// Test new AI titles endpoint that uses URL parameters instead of JSON payload
func TestAITitlesWithVideoParams(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video with manuscript
	_, err := server.videoService.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Create a test manuscript file
	manuscriptContent := "# Test Video Tutorial\n\nThis is a comprehensive tutorial about Docker containerization and Kubernetes orchestration."
	manuscriptPath := server.videoService.GetManuscriptPath("test-video", "test-category")
	err = os.WriteFile(manuscriptPath, []byte(manuscriptContent), 0644)
	require.NoError(t, err)

	// Update the video to set the Gist field
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	video.Gist = manuscriptPath
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid video with manuscript",
			videoName:      "test-video",
			category:       "test-category",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AITitlesResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Titles, "Should return at least one title")
				for _, title := range resp.Titles {
					assert.NotEmpty(t, title, "Each title should not be empty")
					assert.True(t, len(title) > 10, "Title should be reasonably long")
				}
			},
		},
		{
			name:           "Non-existent video",
			videoName:      "non-existent",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty video name",
			videoName:      "",
			category:       "test-category",
			expectedStatus: http.StatusNotFound, // Chi router returns 404 when path param is empty
		},
		{
			name:           "Empty category",
			videoName:      "test-video",
			category:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/ai/titles/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available (expected in test environment)
			if w.Code == http.StatusInternalServerError && tt.expectedStatus == http.StatusOK {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test new AI titles endpoint error scenarios
func TestAITitlesWithVideoParamsErrors(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video without manuscript
	_, err := server.videoService.CreateVideo("video-no-manuscript", "test-category", "")
	require.NoError(t, err)

	// Create a test video with empty Gist field
	_, err = server.videoService.CreateVideo("video-empty-gist", "test-category", "")
	require.NoError(t, err)

	// Create a test video with non-existent manuscript file
	_, err = server.videoService.CreateVideo("video-bad-gist", "test-category", "")
	require.NoError(t, err)
	video, err := server.videoService.GetVideo("video-bad-gist", "test-category")
	require.NoError(t, err)
	video.Gist = "/non/existent/path/to/manuscript.md"
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		errorContains  string
	}{
		{
			name:           "Video with empty Gist field",
			videoName:      "video-empty-gist",
			category:       "test-category",
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Video manuscript not configured",
		},
		{
			name:           "Video with non-existent manuscript file",
			videoName:      "video-bad-gist",
			category:       "test-category",
			expectedStatus: http.StatusInternalServerError,
			errorContains:  "Failed to read manuscript",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/ai/titles/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.errorContains != "" {
				var errorResp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err, "Error response should be valid JSON")
				assert.Contains(t, errorResp, "error", "Error response should contain error field")
				assert.Contains(t, errorResp["error"], tt.errorContains, "Error message should contain expected text")
			}
		})
	}
}

// Test that wrong HTTP methods return 405 for new endpoint
func TestAITitlesWithVideoParamsMethodNotAllowed(t *testing.T) {
	server := setupTestServer(t)

	methods := []string{"GET", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			url := "/api/ai/titles/test-video?category=test-category"
			req := httptest.NewRequest(method, url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// Test new AI description endpoint that uses URL parameters instead of JSON payload
func TestAIDescriptionWithVideoParams(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video with manuscript
	_, err := server.videoService.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Create a test manuscript file
	manuscriptContent := "# Test Video Tutorial\n\nThis is a comprehensive tutorial about Docker containerization and Kubernetes orchestration."
	manuscriptPath := server.videoService.GetManuscriptPath("test-video", "test-category")
	err = os.WriteFile(manuscriptPath, []byte(manuscriptContent), 0644)
	require.NoError(t, err)

	// Update the video to set the Gist field
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	video.Gist = manuscriptPath
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid video with manuscript",
			videoName:      "test-video",
			category:       "test-category",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AIDescriptionResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Description, "Should return a description")
				assert.True(t, len(resp.Description) > 10, "Description should be reasonably long")
			},
		},
		{
			name:           "Non-existent video",
			videoName:      "non-existent",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty video name",
			videoName:      "",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty category",
			videoName:      "test-video",
			category:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/ai/description/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available (expected in test environment)
			if w.Code == http.StatusInternalServerError && tt.expectedStatus == http.StatusOK {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test new AI tags endpoint that uses URL parameters instead of JSON payload
func TestAITagsWithVideoParams(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video with manuscript
	_, err := server.videoService.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Create a test manuscript file
	manuscriptContent := "# Test Video Tutorial\n\nThis is a comprehensive tutorial about Docker containerization and Kubernetes orchestration."
	manuscriptPath := server.videoService.GetManuscriptPath("test-video", "test-category")
	err = os.WriteFile(manuscriptPath, []byte(manuscriptContent), 0644)
	require.NoError(t, err)

	// Update the video to set the Gist field
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	video.Gist = manuscriptPath
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid video with manuscript",
			videoName:      "test-video",
			category:       "test-category",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AITagsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Tags, "Should return at least one tag")
				for _, tag := range resp.Tags {
					assert.NotEmpty(t, tag, "Each tag should not be empty")
				}
			},
		},
		{
			name:           "Non-existent video",
			videoName:      "non-existent",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty video name",
			videoName:      "",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty category",
			videoName:      "test-video",
			category:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/ai/tags/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available (expected in test environment)
			if w.Code == http.StatusInternalServerError && tt.expectedStatus == http.StatusOK {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test new AI tweets endpoint that uses URL parameters instead of JSON payload
func TestAITweetsWithVideoParams(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video with manuscript
	_, err := server.videoService.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Create a test manuscript file
	manuscriptContent := "# Test Video Tutorial\n\nThis is a comprehensive tutorial about Docker containerization and Kubernetes orchestration."
	manuscriptPath := server.videoService.GetManuscriptPath("test-video", "test-category")
	err = os.WriteFile(manuscriptPath, []byte(manuscriptContent), 0644)
	require.NoError(t, err)

	// Update the video to set the Gist field
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	video.Gist = manuscriptPath
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid video with manuscript",
			videoName:      "test-video",
			category:       "test-category",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AITweetsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Tweets, "Should return at least one tweet")
				for _, tweet := range resp.Tweets {
					assert.NotEmpty(t, tweet, "Each tweet should not be empty")
				}
			},
		},
		{
			name:           "Non-existent video",
			videoName:      "non-existent",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty video name",
			videoName:      "",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty category",
			videoName:      "test-video",
			category:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/ai/tweets/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available (expected in test environment)
			if w.Code == http.StatusInternalServerError && tt.expectedStatus == http.StatusOK {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}


// Test new AI description-tags endpoint that uses URL parameters instead of JSON payload
func TestAIDescriptionTagsWithVideoParams(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video with manuscript
	_, err := server.videoService.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Create a test manuscript file
	manuscriptContent := "# Test Video Tutorial\n\nThis is a comprehensive tutorial about Docker containerization and Kubernetes orchestration."
	manuscriptPath := server.videoService.GetManuscriptPath("test-video", "test-category")
	err = os.WriteFile(manuscriptPath, []byte(manuscriptContent), 0644)
	require.NoError(t, err)

	// Update the video to set the Gist field
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	video.Gist = manuscriptPath
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid video with manuscript",
			videoName:      "test-video",
			category:       "test-category",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AIDescriptionTagsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.DescriptionTags, "Should return at least one description tag")
				for _, tag := range resp.DescriptionTags {
					assert.NotEmpty(t, tag, "Each description tag should not be empty")
				}
			},
		},
		{
			name:           "Non-existent video",
			videoName:      "non-existent",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty video name",
			videoName:      "",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty category",
			videoName:      "test-video",
			category:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/ai/description-tags/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available (expected in test environment)
			if w.Code == http.StatusInternalServerError && tt.expectedStatus == http.StatusOK {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test new animations endpoint that uses URL parameters and manuscript parsing (no AI)
func TestAnimationsWithVideoParams(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video with manuscript
	_, err := server.videoService.CreateVideo("test-video", "test-category", "")
	require.NoError(t, err)

	// Create a test manuscript file with TODO comments and sections
	manuscriptContent := `# Test Video Tutorial

This is a comprehensive tutorial about Docker containerization.

## Introduction

Welcome to this tutorial.

TODO: Show Docker logo animation
TODO: Display terminal with typing effect

## Docker Basics

Let's start with the basics.

TODO: Animate container creation process
TODO: Show port mapping visualization

## Setup

Installation steps here.

## Conclusion

That's all for today.

TODO: Show subscribe button animation

## Destroy

Cleanup steps.
`
	manuscriptPath := server.videoService.GetManuscriptPath("test-video", "test-category")
	err = os.WriteFile(manuscriptPath, []byte(manuscriptContent), 0644)
	require.NoError(t, err)

	// Update the video to set the Gist field
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)
	video.Gist = manuscriptPath
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid video with manuscript containing TODO comments",
			videoName:      "test-video",
			category:       "test-category",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AnimationsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")

				// Should contain the parsed animations
				assert.NotEmpty(t, resp.Animations, "Should return animations")

				// Check that TODO comments are parsed correctly
				animationsText := strings.Join(resp.Animations, "\n")
				assert.Contains(t, animationsText, "Show Docker logo animation", "Should contain TODO comment")
				assert.Contains(t, animationsText, "Display terminal with typing effect", "Should contain TODO comment")
				assert.Contains(t, animationsText, "Animate container creation process", "Should contain TODO comment")
				assert.Contains(t, animationsText, "Show port mapping visualization", "Should contain TODO comment")
				assert.Contains(t, animationsText, "Show subscribe button animation", "Should contain TODO comment")

				// Check that sections are parsed correctly (excluding Setup, Intro, Destroy)
				assert.Contains(t, animationsText, "Docker Basics", "Should contain section header")
				assert.Contains(t, animationsText, "Conclusion", "Should contain section header")

				// Should NOT contain excluded sections
				assert.NotContains(t, animationsText, "Introduction", "Should exclude Introduction section")
				assert.NotContains(t, animationsText, "Setup", "Should exclude Setup section")
				assert.NotContains(t, animationsText, "Destroy", "Should exclude Destroy section")
			},
		},
		{
			name:           "Non-existent video",
			videoName:      "non-existent",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Empty video name",
			videoName:      "",
			category:       "test-category",
			expectedStatus: http.StatusNotFound, // Chi router returns 404 when path param is empty
		},
		{
			name:           "Empty category",
			videoName:      "test-video",
			category:       "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/animations/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test animations endpoint error scenarios
func TestAnimationsWithVideoParamsErrors(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video without manuscript
	_, err := server.videoService.CreateVideo("video-no-manuscript", "test-category", "")
	require.NoError(t, err)

	// Create a test video with empty Gist field
	_, err = server.videoService.CreateVideo("video-empty-gist", "test-category", "")
	require.NoError(t, err)

	// Create a test video with non-existent manuscript file
	_, err = server.videoService.CreateVideo("video-bad-gist", "test-category", "")
	require.NoError(t, err)
	video, err := server.videoService.GetVideo("video-bad-gist", "test-category")
	require.NoError(t, err)
	video.Gist = "/non/existent/path/to/manuscript.md"
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	// Create a test video with manuscript but no TODO comments or sections
	_, err = server.videoService.CreateVideo("video-no-animations", "test-category", "")
	require.NoError(t, err)
	manuscriptPath := server.videoService.GetManuscriptPath("video-no-animations", "test-category")
	err = os.WriteFile(manuscriptPath, []byte("Just plain content without TODO or sections"), 0644)
	require.NoError(t, err)
	video, err = server.videoService.GetVideo("video-no-animations", "test-category")
	require.NoError(t, err)
	video.Gist = manuscriptPath
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
		errorContains  string
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Video with empty Gist field",
			videoName:      "video-empty-gist",
			category:       "test-category",
			expectedStatus: http.StatusBadRequest,
			errorContains:  "Video manuscript not configured",
		},
		{
			name:           "Video with non-existent manuscript file",
			videoName:      "video-bad-gist",
			category:       "test-category",
			expectedStatus: http.StatusInternalServerError,
			errorContains:  "Failed to read manuscript",
		},
		{
			name:           "Video with manuscript but no animations content",
			videoName:      "video-no-animations",
			category:       "test-category",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AnimationsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err)
				// Should return empty array, not error
				assert.Empty(t, resp.Animations, "Should return empty animations array")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := fmt.Sprintf("/api/animations/%s?category=%s", tt.videoName, tt.category)
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.errorContains != "" {
				var errorResp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResp)
				assert.NoError(t, err, "Error response should be valid JSON")
				assert.Contains(t, errorResp, "error", "Error response should contain error field")
				assert.Contains(t, errorResp["error"], tt.errorContains, "Error message should contain expected text")
			}

			if tt.validateResp != nil && w.Code == http.StatusOK {
				tt.validateResp(t, w.Body.Bytes())
			}
		})
	}
}

// Test that wrong HTTP methods return 405 for animations endpoint
func TestAnimationsWithVideoParamsMethodNotAllowed(t *testing.T) {
	server := setupTestServer(t)

	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			url := "/api/animations/test-video?category=test-category"
			req := httptest.NewRequest(method, url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		})
	}
}

// Test animations endpoint routing
func TestAnimationsEndpointRouting(t *testing.T) {
	server := NewServer()

	// Test that animations endpoint is properly registered (should return 400 for missing category, not 404)
	req := httptest.NewRequest("GET", "/api/animations/test-video", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Should return 400 (bad request for missing category), not 404 (route not found)
	assert.Equal(t, http.StatusBadRequest, w.Code, "Route should exist and return 400 for missing category")
}
