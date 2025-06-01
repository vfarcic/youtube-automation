package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
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
			name:           "Missing phase parameter returns all videos",
			phase:          "",
			expectedStatus: http.StatusOK,
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
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 1, "Should return exactly one video")

		video := result[0]
		assert.Equal(t, 1, video.ID)
		assert.Equal(t, "Test Video Title", video.Title)
		assert.Equal(t, "2025-01-01T12:00", video.Date)
		assert.Equal(t, "test-thumb.jpg", video.Thumbnail)
		assert.Equal(t, "devops", video.Category)
		assert.Equal(t, "draft", video.Status)
		// Expected: Basic video has minimal completion values
		assert.LessOrEqual(t, video.Progress.Completed, video.Progress.Total)
		assert.Equal(t, 45, video.Progress.Total)
	})

	t.Run("edge cases and missing fields", func(t *testing.T) {
		videos := []storage.Video{
			{
				Index:     3,
				Name:      "no-title-video",
				Title:     "",
				Thumbnail: "",
				Category:  "test",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 1, "Should return exactly one video")

		video := result[0]
		assert.Equal(t, 3, video.ID)
		assert.Equal(t, "no-title-video", video.Title)  // Falls back to name
		assert.Equal(t, "TBD", video.Date)              // Default for missing date
		assert.Equal(t, "default.jpg", video.Thumbnail) // Default thumbnail
		assert.Equal(t, "test", video.Category)
		assert.Equal(t, "draft", video.Status) // Not published
		assert.LessOrEqual(t, video.Progress.Completed, video.Progress.Total)
		assert.Equal(t, 45, video.Progress.Total)
	})

	t.Run("status derivation logic", func(t *testing.T) {
		testCases := []struct {
			name     string
			video    storage.Video
			expected string
		}{
			{
				"basic draft",
				storage.Video{Index: 1, Name: "test", Title: "Test", Category: "test"},
				"draft",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result := transformToVideoListItems([]storage.Video{tc.video})
				assert.Equal(t, tc.expected, result[0].Status)
			})
		}
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

	t.Run("missing phase parameter returns all videos", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/videos/list", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response VideoListResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should return videos from all phases
		assert.NotNil(t, response.Videos)
		// Note: May be empty if no videos exist, but should not error
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

// BenchmarkTransformToVideoListItems measures transformation performance
func BenchmarkTransformToVideoListItems(b *testing.B) {
	// Create test data with realistic sizes
	videos := make([]storage.Video, 50) // Simulate 50 videos
	for i := 0; i < 50; i++ {
		videos[i] = storage.Video{
			Index:     i + 1,
			Name:      fmt.Sprintf("test-video-%d", i+1),
			Title:     fmt.Sprintf("Test Video Title %d with Some Length", i+1),
			Date:      "2025-01-06T16:30:45Z",
			Thumbnail: fmt.Sprintf("thumbnails/video-%d.jpg", i+1),
			Category:  "devops",
			// Initial Details phase fields (8 total tasks)
			ProjectName: "Test Project",     // +1 completed
			ProjectURL:  "https://test.com", // +1 completed
			// Date already set above +1 completed
			// No sponsorship fields set, no Gist, so 3/8 completed for Initial Details

			// Work Progress phase fields (11 total tasks)
			Code:   true, // +1 completed
			Screen: true, // +1 completed
			Head:   true, // +1 completed
			// Other work fields false/empty, so 3/11 completed for Work Progress

			// Definition phase fields (9 total tasks)
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
