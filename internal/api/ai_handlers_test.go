package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
		"/api/ai/highlights",
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
		"/api/ai/highlights",
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
		"/api/ai/highlights",
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

// Test highlights endpoint
func TestAIHighlightsEndpoint(t *testing.T) {
	server := NewServer()

	tests := []struct {
		name           string
		manuscript     string
		expectedStatus int
		validateResp   func(t *testing.T, body []byte)
	}{
		{
			name:           "Valid manuscript",
			manuscript:     "This comprehensive guide covers advanced Python concepts including decorators, context managers, and metaclasses for professional development.",
			expectedStatus: http.StatusOK,
			validateResp: func(t *testing.T, body []byte) {
				var resp AIHighlightsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Highlights, "Should return highlights")
				for _, highlight := range resp.Highlights {
					assert.NotEmpty(t, highlight, "Each highlight should not be empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: tt.manuscript})
			req := httptest.NewRequest("POST", "/api/ai/highlights", bytes.NewReader(reqBody))
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
			endpoint:       "/api/ai/highlights",
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
			"/api/ai/highlights",
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
func TestAIEndpointsDiverseContent(t *testing.T) {
	server := NewServer()

	// Define various manuscript content types for testing
	testManuscripts := map[string]string{
		"Technical Tutorial": `
# Advanced Docker Multi-Stage Builds

## Introduction
Docker multi-stage builds are a powerful feature that allows you to optimize your Docker images by using multiple FROM statements in your Dockerfile. This technique helps reduce the final image size by separating the build environment from the runtime environment.

## Key Benefits
1. **Reduced Image Size**: Only necessary runtime dependencies are included in the final image
2. **Security**: Build tools and source code are not present in the production image
3. **Efficiency**: Faster deployment and reduced storage costs

## Implementation Example
Here's a practical example using a Go application:

` + "```dockerfile" + `
# Build stage
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Production stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
` + "```" + `

## Best Practices
- Use specific base image tags for reproducibility
- Minimize the number of layers in your final stage
- Use .dockerignore to exclude unnecessary files
- Consider using distroless images for even smaller footprints

## Advanced Techniques
Multi-stage builds can also be used for:
- Running tests in a separate stage
- Building multiple variants of your application
- Creating development and production images from the same Dockerfile

This approach significantly improves your container security posture while maintaining development efficiency.`,

		"Entertainment Content": `
# The Ultimate Guide to Streaming Setup for Content Creators

## Getting Started with Streaming
Whether you're planning to stream gaming content, tutorials, or just chatting with your audience, having the right setup is crucial for success. This comprehensive guide will walk you through everything you need to know to create professional-quality streams that engage your viewers.

## Essential Hardware
**Camera**: A good webcam or DSLR camera is your window to the audience. Consider lighting conditions and background when choosing your setup.

**Microphone**: Audio quality can make or break your stream. Invest in a decent USB microphone or audio interface with an XLR mic.

**Computer**: Streaming is resource-intensive. Ensure your CPU and GPU can handle both your content and the encoding process.

## Software Recommendations
- **OBS Studio**: Free, powerful, and widely supported
- **Streamlabs**: User-friendly with built-in alerts and widgets
- **XSplit**: Professional features with a polished interface

## Engagement Strategies
1. **Consistent Schedule**: Viewers need to know when to find you
2. **Interactive Elements**: Polls, Q&A sessions, and viewer games
3. **Community Building**: Discord servers and social media presence
4. **Quality Content**: Plan your streams and have backup activities

## Monetization Options
- Subscriptions and donations
- Sponsorships and brand partnerships
- Merchandise sales
- Affiliate marketing

## Common Mistakes to Avoid
- Ignoring chat and viewer interaction
- Inconsistent streaming schedule
- Poor audio quality
- Overcomplicated overlays and alerts

Remember, building a successful streaming career takes time, patience, and consistent effort. Focus on creating content you're passionate about, and the audience will follow!`,

		"Educational Material": `
# Understanding Climate Change: Science, Impacts, and Solutions

## Introduction to Climate Science
Climate change refers to long-term shifts in global temperatures and weather patterns. While climate variations occur naturally, scientific evidence overwhelmingly shows that human activities have been the dominant driver of climate change since the mid-20th century.

## The Greenhouse Effect
The greenhouse effect is a natural process that warms Earth's surface. When the Sun's energy reaches Earth's atmosphere, some of it is reflected back to space and the rest is absorbed and re-radiated by greenhouse gases.

### Key Greenhouse Gases:
- **Carbon Dioxide (CO2)**: Primary driver, mainly from fossil fuel combustion
- **Methane (CH4)**: From agriculture, landfills, and natural gas production
- **Nitrous Oxide (N2O)**: From agriculture and fossil fuel combustion
- **Fluorinated Gases**: From industrial processes and refrigeration

## Observable Impacts
**Temperature Changes**: Global average temperature has risen by approximately 1.1°C since the late 19th century.

**Ice Loss**: Arctic sea ice is declining at a rate of 13% per decade. Glaciers worldwide are retreating.

**Sea Level Rise**: Global sea level has risen about 8-9 inches since 1880, with the rate of rise accelerating in recent decades.

**Weather Patterns**: Increased frequency of extreme weather events including heatwaves, droughts, and intense storms.

## Regional Variations
Climate change affects different regions differently:
- **Arctic**: Warming twice as fast as the global average
- **Small Island Nations**: Facing existential threats from sea level rise
- **Sub-Saharan Africa**: Increased drought and desertification
- **Coastal Cities**: Flooding and storm surge risks

## Solutions and Mitigation
**Renewable Energy**: Solar, wind, and hydroelectric power to replace fossil fuels

**Energy Efficiency**: Improving building insulation, LED lighting, and efficient appliances

**Transportation**: Electric vehicles, public transit, and sustainable urban planning

**Carbon Capture**: Technologies to remove CO2 from the atmosphere

**Nature-Based Solutions**: Reforestation, wetland restoration, and sustainable agriculture

## Individual Actions
- Reduce energy consumption at home
- Choose sustainable transportation options
- Support renewable energy initiatives
- Make informed consumer choices
- Advocate for climate policies

Understanding climate change is the first step toward addressing this global challenge. Through collective action and technological innovation, we can work toward a sustainable future.`,

		"Short Technical Tip": `
# Quick Git Tip: Interactive Rebase

Use git rebase -i HEAD~3 to interactively rebase the last 3 commits. This allows you to:
- Squash commits together
- Edit commit messages
- Reorder commits
- Drop unwanted commits

Perfect for cleaning up your commit history before merging!`,

		"Very Long Content": strings.Repeat(`
This is a comprehensive tutorial about advanced software engineering practices, design patterns, and architectural principles that every developer should understand to build scalable, maintainable, and robust applications.

## Design Patterns
Design patterns are reusable solutions to commonly occurring problems in software design. They represent best practices and provide a common vocabulary for developers.

### Creational Patterns
- Singleton: Ensures a class has only one instance
- Factory: Creates objects without specifying exact classes
- Builder: Constructs complex objects step by step

### Structural Patterns
- Adapter: Allows incompatible interfaces to work together
- Decorator: Adds behavior to objects dynamically
- Facade: Provides a simplified interface to complex subsystems

### Behavioral Patterns
- Observer: Defines one-to-many dependency between objects
- Strategy: Defines family of algorithms and makes them interchangeable
- Command: Encapsulates requests as objects

## SOLID Principles
1. Single Responsibility Principle
2. Open/Closed Principle
3. Liskov Substitution Principle
4. Interface Segregation Principle
5. Dependency Inversion Principle

## Architectural Patterns
- Model-View-Controller (MVC)
- Model-View-ViewModel (MVVM)
- Microservices Architecture
- Event-Driven Architecture
- Hexagonal Architecture

`, 50), // Repeat 50 times to create very long content

		"Code-Heavy Content": `
# Advanced JavaScript Patterns and Techniques

## Closures and Scope
` + "```javascript" + `
function createCounter() {
    let count = 0;
    return function() {
        return ++count;
    };
}

const counter = createCounter();
console.log(counter()); // 1
console.log(counter()); // 2
` + "```" + `

## Async/Await Patterns
` + "```javascript" + `
async function fetchUserData(userId) {
    try {
        const response = await fetch('/api/users/' + userId);
        if (!response.ok) {
            throw new Error('Failed to fetch user data');
        }
        return await response.json();
    } catch (error) {
        console.error('Error fetching user:', error);
        throw error;
    }
}

// Parallel execution
async function fetchMultipleUsers(userIds) {
    const promises = userIds.map(id => fetchUserData(id));
    return await Promise.all(promises);
}
` + "```" + `

## Advanced Object Patterns
` + "```javascript" + `
class EventEmitter {
    constructor() {
        this.events = {};
    }
    
    on(event, callback) {
        if (!this.events[event]) {
            this.events[event] = [];
        }
        this.events[event].push(callback);
    }
    
    emit(event, data) {
        if (this.events[event]) {
            this.events[event].forEach(callback => callback(data));
        }
    }
    
    off(event, callback) {
        if (this.events[event]) {
            this.events[event] = this.events[event].filter(cb => cb !== callback);
        }
    }
}
` + "```" + `

## Functional Programming Concepts
` + "```javascript" + `
// Higher-order functions
const compose = (...fns) => (value) => fns.reduceRight((acc, fn) => fn(acc), value);

const addOne = x => x + 1;
const double = x => x * 2;
const square = x => x * x;

const transform = compose(square, double, addOne);
console.log(transform(3)); // ((3 + 1) * 2)^2 = 64
` + "```" + `

These patterns form the foundation of modern JavaScript development and enable writing more maintainable and scalable code.`,

		"Minimal Content": "Quick tip: Use console.table() to display arrays and objects in a nice table format in the browser console.",

		"Unicode and Special Characters": `
# Internationalization (i18n) Best Practices 🌍

## Supporting Multiple Languages
When building applications for global audiences, proper internationalization is crucial. Here are key considerations:

### Character Encoding
Always use UTF-8 encoding to support characters from different languages:
- 中文 (Chinese)
- العربية (Arabic)  
- Русский (Russian)
- हिन्दी (Hindi)
- 日本語 (Japanese)
- Español (Spanish)
- Français (French)

### Currency and Number Formatting
Different regions have different conventions:
- US: $1,234.56
- EU: €1.234,56
- India: ₹1,23,456.78

### Date and Time Formats
- US: MM/DD/YYYY
- EU: DD/MM/YYYY
- ISO: YYYY-MM-DD

### Special Characters in Code
Handle special characters properly:
- Quotes: "smart quotes" vs 'straight quotes'
- Dashes: – (en dash) vs — (em dash) vs - (hyphen)
- Symbols: © ® ™ § ¶ † ‡ • …

### Emoji Support 📱
Modern applications should handle emoji properly:
🚀 🎯 💡 🔥 ⭐ 🎉 💻 📊 🌟 ✨

Remember to test your application with various character sets and input methods to ensure proper functionality across all supported languages and regions.`,
	}

	// Test each endpoint with different content types
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
		{
			path: "/api/ai/tweets",
			name: "tweets",
			validate: func(t *testing.T, body []byte) {
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
		{
			path: "/api/ai/highlights",
			name: "highlights",
			validate: func(t *testing.T, body []byte) {
				var resp AIHighlightsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.Highlights, "Should return highlights")
				for _, highlight := range resp.Highlights {
					assert.NotEmpty(t, highlight, "Each highlight should not be empty")
				}
			},
		},
		{
			path: "/api/ai/description-tags",
			name: "description-tags",
			validate: func(t *testing.T, body []byte) {
				var resp AIDescriptionTagsResponse
				err := json.Unmarshal(body, &resp)
				require.NoError(t, err, "Response should be valid JSON")
				assert.NotEmpty(t, resp.DescriptionTags, "Should return description tags")
				assert.Len(t, resp.DescriptionTags, 1, "Should return exactly one description-tags string")
			},
		},
	}

	// Test each endpoint with each content type
	for contentType, manuscript := range testManuscripts {
		for _, endpoint := range endpoints {
			t.Run(fmt.Sprintf("%s with %s", endpoint.name, contentType), func(t *testing.T) {
				reqBody, _ := json.Marshal(AIRequest{Manuscript: manuscript})
				req := httptest.NewRequest("POST", endpoint.path, bytes.NewReader(reqBody))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()

				server.router.ServeHTTP(w, req)

				// Skip validation if AI config is not available (expected in test environment)
				if w.Code == http.StatusInternalServerError {
					t.Skip("AI configuration not available in test environment")
				}

				assert.Equal(t, http.StatusOK, w.Code, "Should handle %s content successfully", contentType)

				if w.Code == http.StatusOK && endpoint.validate != nil {
					endpoint.validate(t, w.Body.Bytes())
				}
			})
		}
	}
}

// Test AI endpoints with various manuscript sizes
func TestAIEndpointsManuscriptSizes(t *testing.T) {
	server := NewServer()

	sizeTests := []struct {
		name       string
		manuscript string
		expectOK   bool
	}{
		{
			name:       "Very short manuscript",
			manuscript: "AI tip.",
			expectOK:   true,
		},
		{
			name:       "Short manuscript",
			manuscript: "This is a quick tutorial about using Docker containers for development.",
			expectOK:   true,
		},
		{
			name:       "Medium manuscript",
			manuscript: strings.Repeat("This is a comprehensive tutorial about software development best practices. ", 20),
			expectOK:   true,
		},
		{
			name:       "Large manuscript",
			manuscript: strings.Repeat("This is a detailed technical tutorial covering advanced concepts in software engineering, including design patterns, architectural principles, and best practices for building scalable applications. ", 100),
			expectOK:   true,
		},
		{
			name:       "Very large manuscript",
			manuscript: strings.Repeat("This is an extremely comprehensive tutorial covering every aspect of modern software development, from basic concepts to advanced architectural patterns, including detailed code examples, best practices, performance optimization techniques, security considerations, testing strategies, deployment procedures, and maintenance guidelines. ", 500),
			expectOK:   true,
		},
	}

	// Test with titles endpoint as representative
	for _, test := range sizeTests {
		t.Run(test.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: test.manuscript})
			req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available
			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			if test.expectOK {
				assert.Equal(t, http.StatusOK, w.Code, "Should handle manuscript of size %d characters", len(test.manuscript))
			} else {
				assert.NotEqual(t, http.StatusOK, w.Code, "Should reject manuscript of size %d characters", len(test.manuscript))
			}

			t.Logf("Manuscript size: %d characters, Response status: %d", len(test.manuscript), w.Code)
		})
	}
}

// Test AI endpoints with content containing various technical domains
func TestAIEndpointsTechnicalDomains(t *testing.T) {
	server := NewServer()

	domainTests := map[string]string{
		"DevOps/Infrastructure": `
# Kubernetes Deployment Strategies

## Blue-Green Deployment
Blue-green deployment is a technique that reduces downtime and risk by running two identical production environments called Blue and Green.

## Rolling Updates
Kubernetes rolling updates allow you to update your application with zero downtime by gradually replacing old pods with new ones.

## Canary Deployments
Canary deployments allow you to roll out changes to a small subset of users before rolling it out to the entire infrastructure.

Key tools: kubectl, Helm, ArgoCD, Istio, Prometheus, Grafana
`,

		"Data Science/AI": `
# Machine Learning Model Deployment Pipeline

## Data Preprocessing
- Feature engineering and selection
- Data cleaning and normalization
- Handling missing values and outliers

## Model Training
- Cross-validation strategies
- Hyperparameter tuning
- Model evaluation metrics

## MLOps Best Practices
- Version control for models and data
- Automated testing and validation
- Continuous integration/deployment
- Model monitoring and drift detection

Tools: Python, scikit-learn, TensorFlow, PyTorch, MLflow, Kubeflow
`,

		"Web Development": `
# Modern React Development Patterns

## Component Architecture
- Functional components with hooks
- Custom hooks for reusable logic
- Component composition patterns

## State Management
- useState and useReducer for local state
- Context API for global state
- External libraries: Redux, Zustand, Jotai

## Performance Optimization
- React.memo for component memoization
- useMemo and useCallback for expensive computations
- Code splitting with React.lazy

## Testing Strategies
- Unit tests with Jest and React Testing Library
- Integration tests for component interactions
- End-to-end tests with Cypress or Playwright
`,

		"Cybersecurity": `
# Application Security Best Practices

## Authentication and Authorization
- Multi-factor authentication (MFA)
- OAuth 2.0 and OpenID Connect
- Role-based access control (RBAC)

## Data Protection
- Encryption at rest and in transit
- Secure key management
- Data anonymization and pseudonymization

## Vulnerability Management
- Regular security assessments
- Dependency scanning and updates
- Static and dynamic code analysis

## Incident Response
- Security monitoring and alerting
- Incident response procedures
- Forensic analysis and recovery

Tools: OWASP ZAP, Burp Suite, Nessus, Wireshark, Metasploit
`,

		"Mobile Development": `
# Cross-Platform Mobile Development with React Native

## Architecture Patterns
- Component-based architecture
- State management with Redux or Context
- Navigation with React Navigation

## Performance Optimization
- Image optimization and lazy loading
- Memory management and leak prevention
- Bundle size optimization

## Native Integration
- Bridging to native modules
- Platform-specific code with Platform API
- Third-party library integration

## Testing and Deployment
- Unit testing with Jest
- E2E testing with Detox
- CI/CD with Fastlane and App Center

Platforms: iOS, Android, Windows, macOS
`,
	}

	// Test titles endpoint with different technical domains
	for domain, manuscript := range domainTests {
		t.Run(fmt.Sprintf("Titles for %s", domain), func(t *testing.T) {
			reqBody, _ := json.Marshal(AIRequest{Manuscript: manuscript})
			req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			// Skip validation if AI config is not available
			if w.Code == http.StatusInternalServerError {
				t.Skip("AI configuration not available in test environment")
			}

			assert.Equal(t, http.StatusOK, w.Code, "Should handle %s content", domain)

			if w.Code == http.StatusOK {
				var resp AITitlesResponse
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.Titles, "Should generate titles for %s content", domain)

				// Log the generated titles for manual inspection
				t.Logf("Generated titles for %s: %v", domain, resp.Titles)
			}
		})
	}
}
