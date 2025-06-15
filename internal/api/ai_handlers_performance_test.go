//go:build performance
// +build performance

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// Performance test constants
const (
	performanceTestTimeout = 30 * time.Second
	concurrentRequests     = 100
	loadTestDuration       = 5 * time.Second
)

// setupPerformanceServer creates a server for performance testing
func setupPerformanceServer() *Server {
	// Create a server instance for testing
	return NewServer()
}

// generateTestPayload creates test payloads of various sizes
func generateTestPayload(size int) AIRequest {
	manuscript := strings.Repeat("Performance testing content. ", size/30)
	return AIRequest{
		Manuscript: manuscript,
	}
}

// BenchmarkAIEndpoints tests the performance of individual AI endpoints
func BenchmarkAITitlesEndpoint(b *testing.B) {
	server := setupPerformanceServer()
	payload := generateTestPayload(1000) // ~1KB payload
	jsonPayload, _ := json.Marshal(payload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		// Accept both 500 (no AI config) and 200 (success) as valid for performance testing
		if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
			b.Fatalf("Expected status 200 or 500, got %d", w.Code)
		}
	}
}

func BenchmarkAIDescriptionEndpoint(b *testing.B) {
	server := setupPerformanceServer()
	payload := generateTestPayload(1000)
	jsonPayload, _ := json.Marshal(payload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/ai/description", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
			b.Fatalf("Expected status 200 or 500, got %d", w.Code)
		}
	}
}

func BenchmarkAITagsEndpoint(b *testing.B) {
	server := setupPerformanceServer()
	payload := generateTestPayload(1000)
	jsonPayload, _ := json.Marshal(payload)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/ai/tags", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
			b.Fatalf("Expected status 200 or 500, got %d", w.Code)
		}
	}
}

// BenchmarkJSONParsing tests JSON parsing performance with various payload sizes
func BenchmarkJSONParsingSmall(b *testing.B) {
	payload := generateTestPayload(100) // ~100 bytes

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(payload)
		var parsed AIRequest
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONParsingMedium(b *testing.B) {
	payload := generateTestPayload(10000) // ~10KB

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(payload)
		var parsed AIRequest
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONParsingLarge(b *testing.B) {
	payload := generateTestPayload(100000) // ~100KB

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		jsonData, _ := json.Marshal(payload)
		var parsed AIRequest
		if err := json.Unmarshal(jsonData, &parsed); err != nil {
			b.Fatal(err)
		}
	}
}

// TestConcurrentRequests tests concurrent request handling
func TestConcurrentRequests(t *testing.T) {
	server := setupPerformanceServer()
	payload := generateTestPayload(1000)
	jsonPayload, _ := json.Marshal(payload)

	var wg sync.WaitGroup
	errors := make(chan error, concurrentRequests)
	responses := make(chan int, concurrentRequests)

	startTime := time.Now()

	// Launch concurrent requests
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			responses <- w.Code
			// Accept both 500 (no AI config) and 200 (success) as valid
			if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
				errors <- fmt.Errorf("expected status 200 or 500, got %d", w.Code)
			}
		}()
	}

	wg.Wait()
	close(responses)
	close(errors)

	duration := time.Since(startTime)

	// Check for errors
	errorCount := len(errors)
	if errorCount > 0 {
		t.Fatalf("Got %d errors out of %d requests", errorCount, concurrentRequests)
	}

	// Verify all responses were successful
	successCount := 0
	for code := range responses {
		if code == http.StatusOK || code == http.StatusInternalServerError {
			successCount++
		}
	}

	if successCount != concurrentRequests {
		t.Fatalf("Expected %d successful responses, got %d", concurrentRequests, successCount)
	}

	// Performance metrics
	avgResponseTime := duration / time.Duration(concurrentRequests)
	requestsPerSecond := float64(concurrentRequests) / duration.Seconds()

	t.Logf("Concurrent requests performance:")
	t.Logf("  Total requests: %d", concurrentRequests)
	t.Logf("  Total duration: %v", duration)
	t.Logf("  Average response time: %v", avgResponseTime)
	t.Logf("  Requests per second: %.2f", requestsPerSecond)

	// Performance assertions (relaxed for API layer testing)
	if avgResponseTime > 200*time.Millisecond {
		t.Errorf("Average response time too high: %v (expected < 200ms)", avgResponseTime)
	}

	if requestsPerSecond < 50 {
		t.Errorf("Requests per second too low: %.2f (expected > 50)", requestsPerSecond)
	}
}

// TestMemoryUsage tests memory usage patterns
func TestMemoryUsage(t *testing.T) {
	server := setupPerformanceServer()

	// Force garbage collection before test
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Generate requests with various payload sizes
	payloadSizes := []int{100, 1000, 10000, 50000}
	requestsPerSize := 50

	for _, size := range payloadSizes {
		payload := generateTestPayload(size)
		jsonPayload, _ := json.Marshal(payload)

		for i := 0; i < requestsPerSize; i++ {
			req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)

			// Accept both 500 (no AI config) and 200 (success) as valid
			if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
				t.Fatalf("Request failed with unexpected status %d", w.Code)
			}
		}
	}

	// Force garbage collection after test
	runtime.GC()
	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	// Calculate memory usage
	allocatedMemory := m2.TotalAlloc - m1.TotalAlloc
	memoryPerRequest := allocatedMemory / uint64(len(payloadSizes)*requestsPerSize)

	t.Logf("Memory usage analysis:")
	t.Logf("  Total requests: %d", len(payloadSizes)*requestsPerSize)
	t.Logf("  Total allocated memory: %d bytes", allocatedMemory)
	t.Logf("  Memory per request: %d bytes", memoryPerRequest)
	t.Logf("  Current heap size: %d bytes", m2.HeapInuse)

	// Memory usage assertions (reasonable limits for full server with middleware)
	maxMemoryPerRequest := uint64(100 * 1024) // 100KB per request (realistic for full server)
	if memoryPerRequest > maxMemoryPerRequest {
		t.Errorf("Memory usage per request too high: %d bytes (expected < %d bytes)",
			memoryPerRequest, maxMemoryPerRequest)
	}
}

// TestResponseTimeConsistency tests response time consistency
func TestResponseTimeConsistency(t *testing.T) {
	server := setupPerformanceServer()
	payload := generateTestPayload(1000)
	jsonPayload, _ := json.Marshal(payload)

	const numRequests = 100
	responseTimes := make([]time.Duration, numRequests)

	// Measure response times
	for i := 0; i < numRequests; i++ {
		start := time.Now()

		req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBuffer(jsonPayload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)

		responseTimes[i] = time.Since(start)

		// Accept both 500 (no AI config) and 200 (success) as valid
		if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
			t.Fatalf("Request %d failed with unexpected status %d", i, w.Code)
		}
	}

	// Calculate statistics
	var total time.Duration
	var min, max time.Duration = responseTimes[0], responseTimes[0]

	for _, rt := range responseTimes {
		total += rt
		if rt < min {
			min = rt
		}
		if rt > max {
			max = rt
		}
	}

	avg := total / time.Duration(numRequests)
	variance := max - min

	t.Logf("Response time consistency:")
	t.Logf("  Average: %v", avg)
	t.Logf("  Minimum: %v", min)
	t.Logf("  Maximum: %v", max)
	t.Logf("  Variance: %v", variance)

	// Consistency assertions (relaxed for full server)
	maxAcceptableAvg := 100 * time.Millisecond
	maxAcceptableVariance := 200 * time.Millisecond

	if avg > maxAcceptableAvg {
		t.Errorf("Average response time too high: %v (expected < %v)", avg, maxAcceptableAvg)
	}

	if variance > maxAcceptableVariance {
		t.Errorf("Response time variance too high: %v (expected < %v)", variance, maxAcceptableVariance)
	}
}

// TestLoadScalability tests system behavior under sustained load
func TestLoadScalability(t *testing.T) {
	server := setupPerformanceServer()
	payload := generateTestPayload(1000)
	jsonPayload, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), loadTestDuration)
	defer cancel()

	var requestCount int64
	var errorCount int64
	var totalResponseTime time.Duration
	var mu sync.Mutex

	// Launch multiple goroutines to generate load
	const numWorkers = 10
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					start := time.Now()

					req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBuffer(jsonPayload))
					req.Header.Set("Content-Type", "application/json")
					w := httptest.NewRecorder()
					server.router.ServeHTTP(w, req)

					responseTime := time.Since(start)

					mu.Lock()
					requestCount++
					totalResponseTime += responseTime
					// Accept both 500 (no AI config) and 200 (success) as valid
					if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
						errorCount++
					}
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Calculate metrics
	avgResponseTime := totalResponseTime / time.Duration(requestCount)
	requestsPerSecond := float64(requestCount) / loadTestDuration.Seconds()
	errorRate := float64(errorCount) / float64(requestCount) * 100

	t.Logf("Load scalability test results:")
	t.Logf("  Duration: %v", loadTestDuration)
	t.Logf("  Total requests: %d", requestCount)
	t.Logf("  Requests per second: %.2f", requestsPerSecond)
	t.Logf("  Average response time: %v", avgResponseTime)
	t.Logf("  Error count: %d", errorCount)
	t.Logf("  Error rate: %.2f%%", errorRate)

	// Performance assertions (relaxed for full server)
	minRequestsPerSecond := 100.0
	maxAvgResponseTime := 200 * time.Millisecond
	maxErrorRate := 1.0 // 1%

	if requestsPerSecond < minRequestsPerSecond {
		t.Errorf("Requests per second too low: %.2f (expected > %.2f)",
			requestsPerSecond, minRequestsPerSecond)
	}

	if avgResponseTime > maxAvgResponseTime {
		t.Errorf("Average response time too high: %v (expected < %v)",
			avgResponseTime, maxAvgResponseTime)
	}

	if errorRate > maxErrorRate {
		t.Errorf("Error rate too high: %.2f%% (expected < %.2f%%)",
			errorRate, maxErrorRate)
	}
}

// TestPayloadSizePerformance tests performance with various payload sizes
func TestPayloadSizePerformance(t *testing.T) {
	server := setupPerformanceServer()

	payloadSizes := []struct {
		name string
		size int
	}{
		{"Small", 100},
		{"Medium", 1000},
		{"Large", 10000},
		{"VeryLarge", 100000},
	}

	for _, ps := range payloadSizes {
		t.Run(ps.name, func(t *testing.T) {
			payload := generateTestPayload(ps.size)
			jsonPayload, _ := json.Marshal(payload)

			const numRequests = 10
			var totalTime time.Duration

			for i := 0; i < numRequests; i++ {
				start := time.Now()

				req := httptest.NewRequest("POST", "/api/ai/titles", bytes.NewBuffer(jsonPayload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				server.router.ServeHTTP(w, req)

				totalTime += time.Since(start)

				// Accept both 500 (no AI config) and 200 (success) as valid
				if w.Code != http.StatusInternalServerError && w.Code != http.StatusOK {
					t.Fatalf("Request failed with unexpected status %d", w.Code)
				}
			}

			avgTime := totalTime / numRequests
			t.Logf("Payload size %s (%d bytes): avg response time %v",
				ps.name, len(jsonPayload), avgTime)

			// Performance should scale reasonably with payload size (relaxed)
			maxTime := time.Duration(ps.size/500+100) * time.Millisecond
			if avgTime > maxTime {
				t.Errorf("Response time too high for %s payload: %v (expected < %v)",
					ps.name, avgTime, maxTime)
			}
		})
	}
}
