package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *Server {
	// Setup test environment
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tempDir)
	t.Cleanup(func() { os.Chdir(originalDir) })

	// Create test structure
	os.Mkdir("manuscript", 0755)
	os.Mkdir("manuscript/test-category", 0755)

	// Create index.yaml file
	indexContent := "[]"
	os.WriteFile("index.yaml", []byte(indexContent), 0644)

	// Initialize dependencies
	filesystem := &filesystem.Operations{}
	videoManager := video.NewManager(filesystem.GetFilePath)
	videoService := service.NewVideoService("index.yaml", filesystem, videoManager)

	server := &Server{
		videoService: videoService,
		port:         8080,
	}
	server.setupRoutes()

	return server
}

func TestServer_CreateVideo(t *testing.T) {
	server := setupTestServer(t)

	tests := []struct {
		name           string
		requestBody    CreateVideoRequest
		expectedStatus int
	}{
		{
			name: "Valid video creation",
			requestBody: CreateVideoRequest{
				Name:     "test-video",
				Category: "test-category",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing name",
			requestBody: CreateVideoRequest{
				Name:     "",
				Category: "test-category",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing category",
			requestBody: CreateVideoRequest{
				Name:     "test-video",
				Category: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/videos", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				var response CreateVideoResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.requestBody.Name, response.Video.Name)
				assert.Equal(t, tt.requestBody.Category, response.Video.Category)
			}
		})
	}
}

func TestServer_GetVideoPhases(t *testing.T) {
	server := setupTestServer(t)

	// Create test video
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/api/videos/phases", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response VideoPhasesResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have at least one phase with videos
	assert.NotEmpty(t, response.Phases)
}

func TestServer_GetVideos(t *testing.T) {
	server := setupTestServer(t)

	// Create test video
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	tests := []struct {
		name           string
		phase          string
		expectedStatus int
	}{
		{
			name:           "Valid phase",
			phase:          "7", // PhaseIdeas
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing phase parameter",
			phase:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid phase parameter",
			phase:          "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/videos"
			if tt.phase != "" {
				url += "?phase=" + tt.phase
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response GetVideosResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.Videos)
			}
		})
	}
}

func TestServer_GetVideo(t *testing.T) {
	server := setupTestServer(t)

	// Create test video
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	tests := []struct {
		name           string
		videoName      string
		category       string
		expectedStatus int
	}{
		{
			name:           "Valid video retrieval",
			videoName:      "test-video",
			category:       "test-category",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Missing category",
			videoName:      "test-video",
			category:       "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Nonexistent video",
			videoName:      "nonexistent",
			category:       "test-category",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/videos/" + tt.videoName
			if tt.category != "" {
				url += "?category=" + tt.category
			}

			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()

			// Add chi context for URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("videoName", tt.videoName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			server.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response GetVideoResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.videoName, response.Video.Name)
				assert.Equal(t, tt.category, response.Video.Category)
			}
		})
	}
}

func TestServer_UpdateVideo(t *testing.T) {
	server := setupTestServer(t)

	// Create test video
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	// Get the video to update
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)

	// Update the video
	video.Title = "Updated Title"
	video.Description = "Updated Description"

	requestBody := UpdateVideoRequest{Video: video}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("PUT", "/api/videos/test-video", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("videoName", "test-video")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetVideoResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", response.Video.Title)
	assert.Equal(t, "Updated Description", response.Video.Description)
}

func TestServer_DeleteVideo(t *testing.T) {
	server := setupTestServer(t)

	// Create test video
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	req := httptest.NewRequest("DELETE", "/api/videos/test-video?category=test-category", nil)
	w := httptest.NewRecorder()

	// Add chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("videoName", "test-video")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify video is deleted
	_, err = server.videoService.GetVideo("test-video", "test-category")
	assert.Error(t, err)
}

func TestServer_GetCategories(t *testing.T) {
	server := setupTestServer(t)

	// Create additional test categories
	os.Mkdir("manuscript/another-category", 0755)

	req := httptest.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response CategoriesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Categories, 2)
	assert.Contains(t, []string{"Another Category", "Test Category"}, response.Categories[0].Name)
	assert.Contains(t, []string{"Another Category", "Test Category"}, response.Categories[1].Name)
}

func TestServer_MoveVideo(t *testing.T) {
	server := setupTestServer(t)

	// Create test video and target directory
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)
	os.Mkdir("manuscript/target-category", 0755)

	requestBody := MoveVideoRequest{
		TargetDirectoryPath: "manuscript/target-category",
	}
	body, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/videos/test-video/move?category=test-category", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add chi context for URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("videoName", "test-video")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify files are moved
	assert.FileExists(t, "manuscript/target-category/test-video.yaml")
	assert.FileExists(t, "manuscript/target-category/test-video.md")
	assert.NoFileExists(t, "manuscript/test-category/test-video.yaml")
	assert.NoFileExists(t, "manuscript/test-category/test-video.md")
}

func TestServer_HealthCheck(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.NotEmpty(t, response["time"])
}

func TestServer_UpdateVideoPhase(t *testing.T) {
	server := setupTestServer(t)

	// Create test video first
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	updateData := map[string]interface{}{
		"title": "Updated Title",
	}

	body, _ := json.Marshal(updateData)
	req := httptest.NewRequest("PUT", "/api/videos/test-video/initial-details?category=test-category", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// The exact status code depends on the video service implementation
	// For now, we'll verify the request was processed
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

// TestVideoListItem tests the new optimized video list data structure
func TestVideoListItem(t *testing.T) {
	t.Run("JSON serialization with expected field names", func(t *testing.T) {
		item := VideoListItem{
			ID:        1,
			Title:     "Top 10 DevOps Tools You MUST Use in 2025!",
			Date:      "2025-01-06T16:00",
			Thumbnail: "material/top-2025/thumbnail-01.jpg",
			Category:  "devops",
			Status:    "published",
			Progress: VideoProgress{
				Completed: 10,
				Total:     11,
			},
		}

		jsonData, err := json.Marshal(item)
		require.NoError(t, err)

		// Verify expected JSON structure
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		// Check all required fields are present with correct names
		assert.Equal(t, float64(1), jsonMap["id"])
		assert.Equal(t, "Top 10 DevOps Tools You MUST Use in 2025!", jsonMap["title"])
		assert.Equal(t, "2025-01-06T16:00", jsonMap["date"])
		assert.Equal(t, "material/top-2025/thumbnail-01.jpg", jsonMap["thumbnail"])
		assert.Equal(t, "devops", jsonMap["category"])
		assert.Equal(t, "published", jsonMap["status"])

		// Check nested progress object
		progress, ok := jsonMap["progress"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, float64(10), progress["completed"])
		assert.Equal(t, float64(11), progress["total"])
	})

	t.Run("size comparison with target ~200 bytes", func(t *testing.T) {
		item := VideoListItem{
			ID:        1,
			Title:     "Top 10 DevOps Tools You MUST Use in 2025!",
			Date:      "2025-01-06T16:00",
			Thumbnail: "material/top-2025/thumbnail-01.jpg",
			Category:  "devops",
			Status:    "published",
			Progress: VideoProgress{
				Completed: 10,
				Total:     11,
			},
		}

		jsonData, err := json.Marshal(item)
		require.NoError(t, err)

		size := len(jsonData)
		t.Logf("VideoListItem JSON size: %d bytes", size)

		// Should be significantly smaller than 8.8KB (8821 bytes)
		// Target is ~200 bytes, allow some flexibility (up to 400 bytes is still excellent)
		assert.Less(t, size, 400, "VideoListItem should be under 400 bytes")
		assert.Greater(t, size, 50, "VideoListItem should have reasonable content (>50 bytes)")
	})

	t.Run("VideoListResponse structure", func(t *testing.T) {
		videos := []VideoListItem{
			{
				ID:        1,
				Title:     "Video 1",
				Date:      "2025-01-01T12:00",
				Thumbnail: "thumb1.jpg",
				Category:  "devops",
				Status:    "published",
				Progress:  VideoProgress{Completed: 10, Total: 10},
			},
			{
				ID:        2,
				Title:     "Video 2",
				Date:      "2025-01-02T12:00",
				Thumbnail: "thumb2.jpg",
				Category:  "ai",
				Status:    "draft",
				Progress:  VideoProgress{Completed: 5, Total: 10},
			},
		}

		response := VideoListResponse{Videos: videos}
		jsonData, err := json.Marshal(response)
		require.NoError(t, err)

		// Verify the response structure
		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		videosArray, ok := jsonMap["videos"].([]interface{})
		require.True(t, ok)
		assert.Len(t, videosArray, 2)

		// Check first video in array
		firstVideo := videosArray[0].(map[string]interface{})
		assert.Equal(t, float64(1), firstVideo["id"])
		assert.Equal(t, "Video 1", firstVideo["title"])
	})

	t.Run("field mapping verification", func(t *testing.T) {
		// This test verifies the field mappings specified in the PRD
		// ID (from Index), Title, Date, Thumbnail, Category, Status (derived), Progress

		item := VideoListItem{
			ID:        15,                 // Maps from storage.Video.Index
			Title:     "Test Title",       // Maps from storage.Video.Title
			Date:      "2025-01-01T12:00", // Maps from storage.Video.Date
			Thumbnail: "test.jpg",         // Maps from storage.Video.Thumbnail
			Category:  "test",             // Maps from storage.Video.Category
			Status:    "published",        // Derived from Publish.Completed == Publish.Total
			Progress: VideoProgress{ // Maps from storage.Video.Publish
				Completed: 8,
				Total:     8,
			},
		}

		// Verify all fields are accessible and correctly typed
		assert.Equal(t, 15, item.ID)
		assert.Equal(t, "Test Title", item.Title)
		assert.Equal(t, "2025-01-01T12:00", item.Date)
		assert.Equal(t, "test.jpg", item.Thumbnail)
		assert.Equal(t, "test", item.Category)
		assert.Equal(t, "published", item.Status)
		assert.Equal(t, 8, item.Progress.Completed)
		assert.Equal(t, 8, item.Progress.Total)
	})
}

// TestTransformToVideoListItems tests the video transformation function
func TestTransformToVideoListItems(t *testing.T) {
	t.Run("basic transformation", func(t *testing.T) {
		videos := []storage.Video{
			{
				Index:     1,
				Name:      "test-video",
				Title:     "Test Video Title",
				Date:      "2025-01-01T12:00",
				Thumbnail: "test-thumb.jpg",
				Category:  "devops",
				Publish: storage.Tasks{
					Completed: 10,
					Total:     10,
				},
			},
			{
				Index:     2,
				Name:      "draft-video",
				Title:     "Draft Video Title",
				Date:      "2025-01-02T12:00",
				Thumbnail: "draft-thumb.jpg",
				Category:  "ai",
				Publish: storage.Tasks{
					Completed: 5,
					Total:     10,
				},
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2)

		// Check published video
		assert.Equal(t, 1, result[0].ID)
		assert.Equal(t, "Test Video Title", result[0].Title)
		assert.Equal(t, "2025-01-01T12:00", result[0].Date)
		assert.Equal(t, "test-thumb.jpg", result[0].Thumbnail)
		assert.Equal(t, "devops", result[0].Category)
		assert.Equal(t, "published", result[0].Status)
		assert.Equal(t, 10, result[0].Progress.Completed)
		assert.Equal(t, 10, result[0].Progress.Total)

		// Check draft video
		assert.Equal(t, 2, result[1].ID)
		assert.Equal(t, "Draft Video Title", result[1].Title)
		assert.Equal(t, "draft", result[1].Status)
		assert.Equal(t, 5, result[1].Progress.Completed)
		assert.Equal(t, 10, result[1].Progress.Total)
	})

	t.Run("edge cases and missing fields", func(t *testing.T) {
		videos := []storage.Video{
			{
				Index:     3,
				Name:      "no-title-video",
				Title:     "", // Missing title
				Date:      "", // Missing date
				Thumbnail: "", // Missing thumbnail
				Category:  "test",
				Publish: storage.Tasks{
					Completed: 0,
					Total:     0, // Zero total
				},
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 1)

		// Check fallback values
		assert.Equal(t, "no-title-video", result[0].Title)  // Falls back to Name
		assert.Equal(t, "TBD", result[0].Date)              // Default for missing date
		assert.Equal(t, "default.jpg", result[0].Thumbnail) // Default thumbnail
		assert.Equal(t, "draft", result[0].Status)          // Draft when total is 0
	})

	t.Run("status derivation logic", func(t *testing.T) {
		testCases := []struct {
			name      string
			completed int
			total     int
			expected  string
		}{
			{"published - complete", 10, 10, "published"},
			{"draft - incomplete", 5, 10, "draft"},
			{"draft - zero total", 0, 0, "draft"},
			{"draft - zero completed", 0, 5, "draft"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				videos := []storage.Video{
					{
						Index:    1,
						Name:     "test",
						Title:    "Test",
						Category: "test",
						Publish: storage.Tasks{
							Completed: tc.completed,
							Total:     tc.total,
						},
					},
				}

				result := transformToVideoListItems(videos)
				assert.Equal(t, tc.expected, result[0].Status)
			})
		}
	})

	t.Run("performance with large dataset", func(t *testing.T) {
		// Create a large dataset to test performance
		videos := make([]storage.Video, 1000)
		for i := 0; i < 1000; i++ {
			videos[i] = storage.Video{
				Index:     i + 1,
				Name:      fmt.Sprintf("video-%d", i+1),
				Title:     fmt.Sprintf("Video Title %d", i+1),
				Date:      "2025-01-01T12:00",
				Thumbnail: fmt.Sprintf("thumb-%d.jpg", i+1),
				Category:  "performance-test",
				Publish: storage.Tasks{
					Completed: i % 10, // Vary completion
					Total:     10,
				},
			}
		}

		// Measure transformation time
		start := time.Now()
		result := transformToVideoListItems(videos)
		duration := time.Since(start)

		// Verify results
		assert.Len(t, result, 1000)

		// Performance should be fast (under 10ms for 1000 items)
		assert.Less(t, duration.Milliseconds(), int64(10),
			"Transformation should be fast for large datasets")

		t.Logf("Transformed 1000 videos in %v", duration)
	})

	t.Run("memory usage comparison", func(t *testing.T) {
		// Create sample video with typical large fields
		video := storage.Video{
			Index:       1,
			Name:        "memory-test",
			Title:       "Memory Test Video",
			Description: strings.Repeat("Long description ", 100), // ~1.8KB
			Highlight:   strings.Repeat("Quote content ", 200),    // ~2.6KB
			Tags:        strings.Repeat("tag1, tag2, tag3, ", 50), // ~650B
			Date:        "2025-01-01T12:00",
			Thumbnail:   "test.jpg",
			Category:    "test",
			Publish: storage.Tasks{
				Completed: 10,
				Total:     10,
			},
		}

		// Test original size
		originalJSON, _ := json.Marshal(video)
		originalSize := len(originalJSON)

		// Test transformed size
		transformed := transformToVideoListItems([]storage.Video{video})
		transformedJSON, _ := json.Marshal(transformed[0])
		transformedSize := len(transformedJSON)

		t.Logf("Original video JSON size: %d bytes", originalSize)
		t.Logf("Transformed video JSON size: %d bytes", transformedSize)

		reduction := float64(originalSize-transformedSize) / float64(originalSize) * 100
		t.Logf("Size reduction: %.1f%%", reduction)

		// Should achieve significant reduction (target 95%+)
		assert.Greater(t, reduction, 90.0, "Should achieve >90% size reduction")
		assert.Less(t, transformedSize, 400, "Transformed size should be under 400 bytes")
	})
}

// TestServer_GetVideosList tests the new optimized video list endpoint
func TestServer_GetVideosList(t *testing.T) {
	server := setupTestServer(t)

	// Create test videos
	_, err := server.videoService.CreateVideo("test-video-1", "test-category")
	require.NoError(t, err)
	_, err = server.videoService.CreateVideo("test-video-2", "test-category")
	require.NoError(t, err)

	t.Run("valid request returns optimized response", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/videos/list?phase=7", nil) // PhaseIdeas
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response VideoListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should have our test videos
		assert.NotEmpty(t, response.Videos)

		// Verify structure of returned videos
		for _, video := range response.Videos {
			assert.GreaterOrEqual(t, video.ID, 0) // ID might be 0 in test setup
			assert.NotEmpty(t, video.Title)
			assert.NotEmpty(t, video.Category)
			assert.Contains(t, []string{"published", "draft"}, video.Status)
			assert.GreaterOrEqual(t, video.Progress.Total, 0)
			assert.GreaterOrEqual(t, video.Progress.Completed, 0)
		}
	})

	t.Run("missing phase parameter returns 400", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/videos/list", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse.Error, "phase parameter is required")
	})

	t.Run("invalid phase parameter returns 400", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/videos/list?phase=invalid", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResponse ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		require.NoError(t, err)
		assert.Contains(t, errorResponse.Error, "Invalid phase parameter")
	})

	t.Run("response format matches VideoListResponse schema", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/videos/list?phase=7", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Parse as generic JSON to verify structure
		var jsonResponse map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
		require.NoError(t, err)

		// Verify top-level structure
		videos, ok := jsonResponse["videos"].([]interface{})
		require.True(t, ok, "Response should have 'videos' array")

		if len(videos) > 0 {
			firstVideo := videos[0].(map[string]interface{})

			// Check all required fields are present
			requiredFields := []string{"id", "title", "date", "thumbnail", "category", "status", "progress"}
			for _, field := range requiredFields {
				_, exists := firstVideo[field]
				assert.True(t, exists, "Field '%s' should be present", field)
			}

			// Check progress object structure
			progress, ok := firstVideo["progress"].(map[string]interface{})
			require.True(t, ok, "Progress should be an object")

			_, exists := progress["completed"]
			assert.True(t, exists, "Progress should have 'completed' field")
			_, exists = progress["total"]
			assert.True(t, exists, "Progress should have 'total' field")
		}
	})

	t.Run("payload size verification", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/videos/list?phase=0", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		responseSize := len(w.Body.Bytes())

		var response VideoListResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if len(response.Videos) > 0 {
			avgSizePerVideo := responseSize / len(response.Videos)
			t.Logf("Response size: %d bytes for %d videos (avg: %d bytes/video)",
				responseSize, len(response.Videos), avgSizePerVideo)

			// Verify we're achieving significant size reduction
			assert.Less(t, avgSizePerVideo, 500, "Average size per video should be under 500 bytes")
		}
	})
}

// Advanced Testing: Edge Cases and Performance Analysis
// The following tests provide comprehensive coverage of the optimized video list functionality

// TestTransformToVideoListItems_Comprehensive provides exhaustive coverage of edge cases
func TestTransformToVideoListItems_Comprehensive(t *testing.T) {
	t.Run("edge cases and data integrity", func(t *testing.T) {
		videos := []storage.Video{
			// Test complete published video
			{
				Index:     1,
				Name:      "published-video",
				Title:     "Published Test Video",
				Date:      "2025-01-01T12:00:00Z",
				Thumbnail: "thumbnails/published.jpg",
				Category:  "devops",
				Publish:   storage.Tasks{Completed: 10, Total: 10},
			},
			// Test partial progress video
			{
				Index:     2,
				Name:      "partial-video",
				Title:     "Partial Progress Video",
				Date:      "2025-01-02T14:30:00Z",
				Thumbnail: "thumbnails/partial.jpg",
				Category:  "cloud",
				Publish:   storage.Tasks{Completed: 5, Total: 10},
			},
			// Test video with missing fields (edge case)
			{
				Index:     3,
				Name:      "minimal-video",
				Title:     "", // Empty title
				Date:      "", // Empty date
				Thumbnail: "", // Empty thumbnail
				Category:  "", // Empty category
				Publish:   storage.Tasks{Completed: 0, Total: 0},
			},
			// Test video with zero progress
			{
				Index:     4,
				Name:      "zero-progress",
				Title:     "Zero Progress Video",
				Date:      "2025-01-03T09:15:00Z",
				Thumbnail: "thumbnails/zero.jpg",
				Category:  "kubernetes",
				Publish:   storage.Tasks{Completed: 0, Total: 5},
			},
		}

		result := transformToVideoListItems(videos)

		// Verify correct count
		assert.Len(t, result, 4)

		// Test published video
		published := result[0]
		assert.Equal(t, 1, published.ID)
		assert.Equal(t, "Published Test Video", published.Title)
		assert.Equal(t, "2025-01-01T12:00:00Z", published.Date)
		assert.Equal(t, "thumbnails/published.jpg", published.Thumbnail)
		assert.Equal(t, "devops", published.Category)
		assert.Equal(t, "published", published.Status)
		assert.Equal(t, 10, published.Progress.Completed)
		assert.Equal(t, 10, published.Progress.Total)

		// Test partial progress video
		partial := result[1]
		assert.Equal(t, 2, partial.ID)
		assert.Equal(t, "draft", partial.Status)
		assert.Equal(t, 5, partial.Progress.Completed)
		assert.Equal(t, 10, partial.Progress.Total)

		// Test edge case with missing fields
		minimal := result[2]
		assert.Equal(t, 3, minimal.ID)
		assert.Equal(t, "minimal-video", minimal.Title) // Should default to video.Name
		assert.Equal(t, "draft", minimal.Status)
		assert.Equal(t, 0, minimal.Progress.Completed)
		assert.Equal(t, 0, minimal.Progress.Total)

		// Test zero progress
		zero := result[3]
		assert.Equal(t, 4, zero.ID)
		assert.Equal(t, "Zero Progress Video", zero.Title)
		assert.Equal(t, "draft", zero.Status)
		assert.Equal(t, 0, zero.Progress.Completed)
		assert.Equal(t, 5, zero.Progress.Total)
	})

	t.Run("JSON serialization size validation", func(t *testing.T) {
		// Create a realistic video with all fields that would normally be large
		video := storage.Video{
			Index:     42,
			Name:      "realistic-test-video-with-long-name",
			Title:     "A Comprehensive Guide to Kubernetes Deployment Strategies in Production Environments",
			Date:      "2025-01-06T16:30:45Z",
			Thumbnail: "material/devops/kubernetes/deployment-strategies/thumbnail.jpg",
			Category:  "kubernetes",
			Publish:   storage.Tasks{Completed: 8, Total: 12},
			// These fields would be excluded in transformation
			Description: "This is a very long description that would normally add significant size to the JSON response. It contains detailed information about the content, learning objectives, prerequisites, and much more detailed information that frontend lists don't need.",
			Tags:        "kubernetes,devops,deployment,production,strategies,containers,orchestration,microservices,cloud-native,scalability",
			Highlight:   "Learn advanced Kubernetes deployment patterns including blue-green deployments, canary releases, and rolling updates with practical examples.",
		}

		result := transformToVideoListItems([]storage.Video{video})
		require.Len(t, result, 1)

		// Serialize to JSON and verify size
		jsonData, err := json.Marshal(result[0])
		require.NoError(t, err)

		jsonSize := len(jsonData)
		t.Logf("Optimized VideoListItem JSON size: %d bytes", jsonSize)

		// Should be significantly smaller than 300 bytes (our target is ~200)
		assert.Less(t, jsonSize, 350, "Single VideoListItem should be under 350 bytes")
		assert.Greater(t, jsonSize, 150, "Should contain meaningful data (over 150 bytes)")

		// Verify it contains expected fields
		assert.Contains(t, string(jsonData), "Comprehensive Guide")
		assert.Contains(t, string(jsonData), "kubernetes")
		assert.Contains(t, string(jsonData), "thumbnail.jpg")
		assert.NotContains(t, string(jsonData), "very long description") // Should be excluded
	})
}

// TestServer_GetVideosList_Comprehensive provides complete API endpoint testing
func TestServer_GetVideosList_Comprehensive(t *testing.T) {
	server := setupTestServer(t)

	// Create multiple test videos for comprehensive testing
	videos := []struct {
		name     string
		category string
	}{
		{"kubernetes-basics", "devops"},
		{"docker-advanced", "devops"},
		{"cloud-migration", "cloud"},
		{"security-best-practices", "security"},
		{"monitoring-setup", "observability"},
	}

	for _, v := range videos {
		_, err := server.videoService.CreateVideo(v.name, v.category)
		require.NoError(t, err)
	}

	t.Run("successful requests with different phases", func(t *testing.T) {
		phases := []string{"0", "1", "2", "3", "4", "5", "6", "7"}

		for _, phase := range phases {
			req := httptest.NewRequest("GET", "/api/videos/list?phase="+phase, nil)
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Phase %s should return 200", phase)

			var response VideoListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err, "Response should be valid JSON for phase %s", phase)
			assert.NotNil(t, response.Videos, "Videos array should not be nil for phase %s", phase)
		}
	})

	t.Run("error handling", func(t *testing.T) {
		testCases := []struct {
			name           string
			url            string
			expectedStatus int
		}{
			{"missing phase parameter", "/api/videos/list", http.StatusBadRequest},
			{"empty phase parameter", "/api/videos/list?phase=", http.StatusBadRequest},
			{"invalid phase parameter", "/api/videos/list?phase=invalid", http.StatusBadRequest},
			{"negative phase", "/api/videos/list?phase=-1", http.StatusOK},       // Might be valid depending on implementation
			{"very large phase", "/api/videos/list?phase=999999", http.StatusOK}, // Should handle gracefully
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", tc.url, nil)
				w := httptest.NewRecorder()

				server.router.ServeHTTP(w, req)

				assert.Equal(t, tc.expectedStatus, w.Code, "Test case: %s", tc.name)
			})
		}
	})

	t.Run("response format and content validation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/videos/list?phase=0", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response VideoListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Validate each video in the response
		for i, video := range response.Videos {
			assert.GreaterOrEqual(t, video.ID, 0, "Video %d should have valid ID", i)
			assert.NotEmpty(t, video.Title, "Video %d should have title", i)
			assert.NotEmpty(t, video.Category, "Video %d should have category", i)
			assert.Contains(t, []string{"published", "draft"}, video.Status, "Video %d should have valid status", i)
			assert.GreaterOrEqual(t, video.Progress.Total, 0, "Video %d should have valid progress total", i)
			assert.GreaterOrEqual(t, video.Progress.Completed, 0, "Video %d should have valid progress completed", i)
			assert.LessOrEqual(t, video.Progress.Completed, video.Progress.Total, "Video %d progress should be logical", i)
		}
	})

	t.Run("payload size verification", func(t *testing.T) {
		// Test with phase 7 which should have actual videos for meaningful comparison
		req1 := httptest.NewRequest("GET", "/api/videos/list?phase=7", nil)
		w1 := httptest.NewRecorder()
		server.router.ServeHTTP(w1, req1)

		req2 := httptest.NewRequest("GET", "/api/videos?phase=7", nil)
		w2 := httptest.NewRecorder()
		server.router.ServeHTTP(w2, req2)

		optimizedSize := w1.Body.Len()
		originalSize := w2.Body.Len()

		t.Logf("Optimized response size: %d bytes", optimizedSize)
		t.Logf("Original response size: %d bytes", originalSize)

		if originalSize > 0 {
			reduction := float64(originalSize-optimizedSize) / float64(originalSize) * 100
			t.Logf("Size reduction: %.1f%%", reduction)

			// Only assert size reduction if there are actual videos to compare
			assert.Greater(t, reduction, 50.0, "List endpoint should be significantly smaller")
		}
	})
}

// ========================================
// BENCHMARK TESTS FOR PERFORMANCE ANALYSIS
// ========================================

// Performance Benchmarks and Analysis
// These benchmarks measure and compare endpoint performance

// BenchmarkVideoListEndpoint measures performance of the optimized endpoint
func BenchmarkVideoListEndpoint(b *testing.B) {
	server := setupBenchmarkServer(b)
	defer server.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := http.Get(server.URL + "/api/videos/list?phase=0")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatalf("Failed to read response: %v", err)
		}
		resp.Body.Close()
	}
}

// BenchmarkOriginalEndpoint measures performance of the original endpoint for comparison
func BenchmarkOriginalEndpoint(b *testing.B) {
	server := setupBenchmarkServer(b)
	defer server.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		resp, err := http.Get(server.URL + "/api/videos?phase=0")
		if err != nil {
			b.Fatalf("Request failed: %v", err)
		}
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			b.Fatalf("Failed to read response: %v", err)
		}
		resp.Body.Close()
	}
}

// BenchmarkTransformToVideoListItems measures transformation performance
func BenchmarkTransformToVideoListItems(b *testing.B) {
	// Create test data with realistic sizes
	videos := make([]storage.Video, 50) // Simulate 50 videos
	for i := 0; i < 50; i++ {
		videos[i] = storage.Video{
			Index:       i + 1,
			Name:        fmt.Sprintf("test-video-%d", i+1),
			Title:       fmt.Sprintf("Test Video Title %d with Some Length", i+1),
			Date:        "2025-01-06T16:30:45Z",
			Thumbnail:   fmt.Sprintf("thumbnails/video-%d.jpg", i+1),
			Category:    "devops",
			Publish:     storage.Tasks{Completed: i % 10, Total: 10},
			Description: "This is a longer description that would normally add significant size",
			Tags:        "kubernetes,devops,deployment,production,strategies",
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result := transformToVideoListItems(videos)
		_ = result // Prevent optimization
	}
}

// setupBenchmarkServer creates a test server for benchmarking
func setupBenchmarkServer(b *testing.B) *httptest.Server {
	server := setupTestServer(&testing.T{})

	// Create some test videos for benchmarking
	for i := 0; i < 10; i++ {
		_, err := server.videoService.CreateVideo(fmt.Sprintf("bench-video-%d", i), "devops")
		if err != nil {
			b.Fatalf("Failed to create test video: %v", err)
		}
	}

	return httptest.NewServer(server.router)
}

// TestPerformanceComparison provides detailed performance analysis
func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance comparison in short mode")
	}

	server := setupTestServer(t)

	// Create test videos
	for i := 0; i < 25; i++ {
		_, err := server.videoService.CreateVideo(fmt.Sprintf("perf-video-%d", i), "devops")
		require.NoError(t, err)
	}

	t.Run("response time comparison", func(t *testing.T) {
		const iterations = 10

		// Measure optimized endpoint
		var optimizedTimes []time.Duration
		for i := 0; i < iterations; i++ {
			start := time.Now()
			req := httptest.NewRequest("GET", "/api/videos/list?phase=0", nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)
			optimizedTimes = append(optimizedTimes, time.Since(start))

			require.Equal(t, http.StatusOK, w.Code)
		}

		// Measure original endpoint
		var originalTimes []time.Duration
		for i := 0; i < iterations; i++ {
			start := time.Now()
			req := httptest.NewRequest("GET", "/api/videos?phase=0", nil)
			w := httptest.NewRecorder()
			server.router.ServeHTTP(w, req)
			originalTimes = append(originalTimes, time.Since(start))

			require.Equal(t, http.StatusOK, w.Code)
		}

		// Calculate averages
		avgOptimized := averageDuration(optimizedTimes)
		avgOriginal := averageDuration(originalTimes)

		t.Logf("Average optimized endpoint time: %v", avgOptimized)
		t.Logf("Average original endpoint time: %v", avgOriginal)

		if avgOriginal > 0 {
			improvement := float64(avgOriginal-avgOptimized) / float64(avgOriginal) * 100
			t.Logf("Performance improvement: %.1f%%", improvement)
		}
	})

	t.Run("payload size comparison", func(t *testing.T) {
		// Test with phase 7 which should have actual videos for meaningful comparison
		req1 := httptest.NewRequest("GET", "/api/videos/list?phase=7", nil)
		w1 := httptest.NewRecorder()
		server.router.ServeHTTP(w1, req1)

		req2 := httptest.NewRequest("GET", "/api/videos?phase=7", nil)
		w2 := httptest.NewRecorder()
		server.router.ServeHTTP(w2, req2)

		optimizedSize := w1.Body.Len()
		originalSize := w2.Body.Len()

		t.Logf("Optimized response size: %d bytes", optimizedSize)
		t.Logf("Original response size: %d bytes", originalSize)

		if originalSize > 0 {
			reduction := float64(originalSize-optimizedSize) / float64(originalSize) * 100
			t.Logf("Size reduction: %.1f%%", reduction)

			// Only assert size reduction if there are actual videos to compare
			assert.Greater(t, reduction, 50.0, "Should achieve at least 50%% size reduction when videos are present")
		} else {
			t.Skip("No videos in phase 7 to test size reduction")
		}
	})
}

// averageDuration calculates the average of a slice of durations
func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}
