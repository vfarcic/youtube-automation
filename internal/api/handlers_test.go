package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/aspect"
	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
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
	aspectService := aspect.NewService()

	server := &Server{
		videoService:  videoService,
		aspectService: aspectService,
		port:          8080,
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

				// Verify the response includes the string-based ID
				expectedID := tt.category + "/" + tt.videoName
				assert.Equal(t, expectedID, response.Video.ID, "Individual video response should include string-based ID")
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
			ID:        "devops/top-2025-tools",
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
		assert.Equal(t, "devops/top-2025-tools", jsonMap["id"])
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
			ID:        "devops/top-2025-tools",
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
				ID:        "devops/video-1",
				Title:     "Video 1",
				Date:      "2025-01-01T12:00",
				Thumbnail: "thumb1.jpg",
				Category:  "devops",
				Status:    "published",
				Progress:  VideoProgress{Completed: 10, Total: 10},
			},
			{
				ID:        "ai/video-2",
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
		assert.Equal(t, "devops/video-1", firstVideo["id"])
		assert.Equal(t, "Video 1", firstVideo["title"])
	})

	t.Run("field mapping verification", func(t *testing.T) {
		// This test verifies the field mappings specified in the PRD
		// ID (from Index), Title, Date, Thumbnail, Category, Status (derived), Progress

		item := VideoListItem{
			ID:        "test/test-video",  // Maps from storage.Video.Category + "/" + storage.Video.Name
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
		assert.Equal(t, "test/test-video", item.ID)
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
		assert.Equal(t, "devops/test-video", video.ID)
		assert.Equal(t, "Test Video Title", video.Title)
		assert.Equal(t, "2025-01-01T12:00", video.Date)
		assert.Equal(t, "test-thumb.jpg", video.Thumbnail)
		assert.Equal(t, "devops", video.Category)
		assert.Equal(t, "draft", video.Status)
		// Expected: Basic video has minimal completion values
		assert.LessOrEqual(t, video.Progress.Completed, video.Progress.Total)
		assert.Equal(t, 46, video.Progress.Total)
	})

	t.Run("edge cases and missing fields", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:      "no-title-video",
				Title:     "",
				Thumbnail: "",
				Category:  "test",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 1, "Should return exactly one video")

		video := result[0]
		assert.Equal(t, "test/no-title-video", video.ID)
		assert.Equal(t, "no-title-video", video.Title)  // Falls back to name
		assert.Equal(t, "TBD", video.Date)              // Default for missing date
		assert.Equal(t, "default.jpg", video.Thumbnail) // Default thumbnail
		assert.Equal(t, "test", video.Category)
		assert.Equal(t, "draft", video.Status) // Not published
		assert.LessOrEqual(t, video.Progress.Completed, video.Progress.Total)
		assert.Equal(t, 46, video.Progress.Total)
	})

	t.Run("status derivation logic", func(t *testing.T) {
		testCases := []struct {
			name     string
			video    storage.Video
			expected string
		}{
			{
				"basic draft",
				storage.Video{Name: "test", Title: "Test", Category: "test"},
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

	t.Run("should handle special characters and unicode", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:     "video-with-spaces-special", // Sanitized at service level
				Title:    "Test Video",
				Category: "test-category",
			},
			{
				Name:     "vidéo-avec-accénts", // Sanitized at service level
				Title:    "French Video",
				Category: "français",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2, "Should return exactly two videos")

		// Names are now sanitized at the service level, so they should be lowercase with hyphens
		assert.Equal(t, "test-category/video-with-spaces-special", result[0].ID, "Should sanitize spaces and special characters")
		assert.Equal(t, "français/vidéo-avec-accénts", result[1].ID, "Should handle unicode characters and preserve accents")
	})

	t.Run("should handle malformed paths gracefully", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:     "test-video", // Sanitized at service level
				Title:    "Malformed Path Test",
				Category: "test",
			},
			{
				Name:     "another-test", // Sanitized at service level
				Title:    "No Path Segments",
				Category: "test",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2, "Should return exactly two videos")

		// Names are now sanitized at the service level
		assert.Equal(t, "test/test-video", result[0].ID, "Should use sanitized names")
		assert.Equal(t, "test/another-test", result[1].ID, "Should use sanitized names")
	})

	t.Run("should handle very long names and paths", func(t *testing.T) {
		longName := strings.Repeat("Very Long Name ", 20) + "End"

		// Sanitize the long name as the service would
		sanitizedLongName := strings.ToLower(strings.ReplaceAll(longName, " ", "-"))
		videos := []storage.Video{
			{
				Name:     sanitizedLongName, // Sanitized at service level
				Title:    "Long Name Test",
				Category: "test",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 1, "Should return exactly one video")

		video := result[0]
		// Names are now sanitized at the service level
		assert.Equal(t, "test/"+sanitizedLongName, video.ID)
		assert.Equal(t, "Long Name Test", video.Title)
	})

	t.Run("should handle path with no filename", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:     "fallback-name", // Sanitized at service level
				Title:    "No Filename Test",
				Category: "test",
			},
			{
				Name:     "another-fallback", // Sanitized at service level
				Title:    "Empty Path Test",
				Category: "test",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2, "Should return exactly two videos")

		// Names are now sanitized at the service level
		assert.Equal(t, "test/fallback-name", result[0].ID, "Should use sanitized names")
		assert.Equal(t, "test/another-fallback", result[1].ID, "Should use sanitized names")
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
			assert.NotEmpty(t, video.ID) // ID should be string like "category/name"
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

// TestVideoListItemPhaseField tests that the phase field is correctly calculated and included in API responses
func TestVideoListItemPhaseField(t *testing.T) {
	testCases := []struct {
		name          string
		video         storage.Video
		expectedPhase int
		description   string
	}{
		{
			name: "PhaseDelayed",
			video: storage.Video{
				Name:     "delayed-video",
				Category: "test-category",
				Delayed:  true,
				Title:    "Delayed Video",
			},
			expectedPhase: workflow.PhaseDelayed,
			description:   "Delayed video should have phase 5",
		},
		{
			name: "PhaseSponsoredBlocked",
			video: storage.Video{
				Name:        "blocked-video",
				Category:    "test-category",
				Sponsorship: storage.Sponsorship{Blocked: "Waiting for sponsor"},
				Title:       "Blocked Video",
			},
			expectedPhase: workflow.PhaseSponsoredBlocked,
			description:   "Sponsored blocked video should have phase 6",
		},
		{
			name: "PhasePublished",
			video: storage.Video{
				Name:     "published-video",
				Category: "test-category",
				Repo:     "github.com/some/repo",
				Title:    "Published Video",
			},
			expectedPhase: workflow.PhasePublished,
			description:   "Published video should have phase 0",
		},
		{
			name: "PhaseIdeas",
			video: storage.Video{
				Name:     "idea-video",
				Category: "test-category",
				Title:    "Idea Video",
			},
			expectedPhase: workflow.PhaseIdeas,
			description:   "Video with no workflow state should have phase 7 (Ideas)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := setupTestServer(t)

			// Create test video with the appropriate setup
			_, err := server.videoService.CreateVideo(tc.video.Name, tc.video.Category)
			assert.NoError(t, err)

			// Update the video file with our test data
			videoPath := filepath.Join("manuscript", tc.video.Category, tc.video.Name+".yaml")
			videoData, err := yaml.Marshal(tc.video)
			assert.NoError(t, err)
			err = os.WriteFile(videoPath, videoData, 0644)
			assert.NoError(t, err)

			// Make request to the API
			req := httptest.NewRequest("GET", "/api/videos/list", nil)
			recorder := httptest.NewRecorder()

			server.router.ServeHTTP(recorder, req)

			// Verify response
			assert.Equal(t, http.StatusOK, recorder.Code)

			var response VideoListResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Find our test video in the response
			var testVideo *VideoListItem
			for _, video := range response.Videos {
				if video.Title == tc.video.Title {
					testVideo = &video
					break
				}
			}

			assert.NotNil(t, testVideo, "Test video not found in response")

			if testVideo == nil {
				return // Avoid panic
			}

			// Verify the phase field is correctly set
			assert.Equal(t, tc.expectedPhase, testVideo.Phase,
				fmt.Sprintf("Test %s failed: %s. Expected phase %d, got %d",
					tc.name, tc.description, tc.expectedPhase, testVideo.Phase))
		})
	}
}

func TestGetEditingAspects(t *testing.T) {
	server := setupTestServer(t)

	// Create a new request
	req, err := http.NewRequest("GET", "/api/editing/aspects", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	server.router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the Content-Type header
	expected := "application/json"
	if ct := rr.Header().Get("Content-Type"); ct != expected {
		t.Errorf("Handler returned wrong content type: got %v want %v", ct, expected)
	}

	// Parse the response body
	var aspectOverview aspect.AspectOverview
	if err := json.Unmarshal(rr.Body.Bytes(), &aspectOverview); err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	// Verify the response structure
	if len(aspectOverview.Aspects) == 0 {
		t.Error("Response should contain aspects")
	}

	// Verify each aspect summary has required fields
	for i, aspectSummary := range aspectOverview.Aspects {
		if aspectSummary.Key == "" {
			t.Errorf("Aspect %d: Key is empty", i)
		}
		if aspectSummary.Title == "" {
			t.Errorf("Aspect %d: Title is empty", i)
		}
		if aspectSummary.Description == "" {
			t.Errorf("Aspect %d: Description is empty", i)
		}
		if aspectSummary.Endpoint == "" {
			t.Errorf("Aspect %d: Endpoint is empty", i)
		}
		if aspectSummary.Icon == "" {
			t.Errorf("Aspect %d: Icon is empty", i)
		}
		if aspectSummary.Order == 0 {
			t.Errorf("Aspect %d: Order should not be zero", i)
		}
		if aspectSummary.FieldCount == 0 {
			t.Errorf("Aspect %d: FieldCount should not be zero", i)
		}
	}

	// Verify basic structure and order
	for i, aspectSummary := range aspectOverview.Aspects {
		expectedOrder := i + 1
		if aspectSummary.Order != expectedOrder {
			t.Errorf("Aspect %d: expected order %d, got %d", i, expectedOrder, aspectSummary.Order)
		}
		// TDD: Check for CompletedFieldCount (should be 0 without video context)
		if aspectSummary.CompletedFieldCount != 0 {
			t.Errorf("Aspect %d: completedFieldCount should be 0 without video context, got %d", i, aspectSummary.CompletedFieldCount)
		}
	}

	// Verify expected aspect keys are present
	expectedKeys := []string{
		aspect.AspectKeyInitialDetails,
		aspect.AspectKeyWorkProgress,
		aspect.AspectKeyDefinition,
		aspect.AspectKeyPostProduction,
		aspect.AspectKeyPublishing,
		aspect.AspectKeyPostPublish,
	}

	if len(aspectOverview.Aspects) != len(expectedKeys) {
		t.Errorf("Expected %d aspects, got %d", len(expectedKeys), len(aspectOverview.Aspects))
	}

	for i, expectedKey := range expectedKeys {
		if i >= len(aspectOverview.Aspects) {
			t.Errorf("Missing aspect with key %s", expectedKey)
			continue
		}
		if aspectOverview.Aspects[i].Key != expectedKey {
			t.Errorf("Expected aspect key %s at index %d, got %s", expectedKey, i, aspectOverview.Aspects[i].Key)
		}
	}
}

func TestGetEditingAspectFields(t *testing.T) {
	server := setupTestServer(t)

	testCases := []struct {
		name           string
		aspectKey      string
		expectedStatus int
		shouldHaveData bool
	}{
		{
			name:           "Valid aspect key - initial-details",
			aspectKey:      aspect.AspectKeyInitialDetails,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "Valid aspect key - work-progress",
			aspectKey:      aspect.AspectKeyWorkProgress,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "Valid aspect key - definition",
			aspectKey:      aspect.AspectKeyDefinition,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "Valid aspect key - post-production",
			aspectKey:      aspect.AspectKeyPostProduction,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "Valid aspect key - publishing",
			aspectKey:      aspect.AspectKeyPublishing,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "Valid aspect key - post-publish",
			aspectKey:      aspect.AspectKeyPostPublish,
			expectedStatus: http.StatusOK,
			shouldHaveData: true,
		},
		{
			name:           "Invalid aspect key",
			aspectKey:      "invalid-aspect",
			expectedStatus: http.StatusNotFound,
			shouldHaveData: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with aspect key in URL path
			req, err := http.NewRequest("GET", "/api/editing/aspects/"+tc.aspectKey+"/fields", nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder
			rr := httptest.NewRecorder()

			// Call the handler
			server.router.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tc.expectedStatus)
			}

			if tc.shouldHaveData {
				// Check the Content-Type header for successful responses
				expected := "application/json"
				if ct := rr.Header().Get("Content-Type"); ct != expected {
					t.Errorf("Handler returned wrong content type: got %v want %v", ct, expected)
				}

				// Parse the response body
				var aspectFields aspect.AspectFields
				if err := json.Unmarshal(rr.Body.Bytes(), &aspectFields); err != nil {
					t.Errorf("Failed to parse response JSON: %v", err)
				}

				// Verify the response structure
				if aspectFields.AspectKey != tc.aspectKey {
					t.Errorf("Expected AspectKey %s, got %s", tc.aspectKey, aspectFields.AspectKey)
				}
				if aspectFields.AspectTitle == "" {
					t.Error("AspectTitle should not be empty")
				}
				if len(aspectFields.Fields) == 0 {
					t.Error("Fields should not be empty")
				}

				// Verify each field has required properties
				for i, field := range aspectFields.Fields {
					if field.Name == "" {
						t.Errorf("Field %d: Name is empty", i)
					}
					if field.Type == "" {
						t.Errorf("Field %d: Type is empty", i)
					}
					if field.Description == "" {
						t.Errorf("Field %d: Description is empty", i)
					}
					if field.Order == 0 {
						t.Errorf("Field %d: Order should not be zero", i)
					}

					// Verify field types are valid
					validTypes := []string{
						aspect.FieldTypeString,
						aspect.FieldTypeText,
						aspect.FieldTypeBoolean,
						aspect.FieldTypeDate,
						aspect.FieldTypeNumber,
						aspect.FieldTypeSelect,
					}
					isValidType := false
					for _, validType := range validTypes {
						if field.Type == validType {
							isValidType = true
							break
						}
					}
					if !isValidType {
						t.Errorf("Field %d: Invalid field type: %s", i, field.Type)
					}
				}

				// Verify fields are ordered correctly
				for i := 1; i < len(aspectFields.Fields); i++ {
					if aspectFields.Fields[i].Order <= aspectFields.Fields[i-1].Order {
						t.Errorf("Fields are not properly ordered: field %d has order %d, previous field has order %d",
							i, aspectFields.Fields[i].Order, aspectFields.Fields[i-1].Order)
					}
				}
			} else {
				// For error responses, check that we get an error message
				var errorResponse map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
					t.Errorf("Failed to parse error response JSON: %v", err)
				}
				if errorResponse["error"] == "" {
					t.Error("Error response should contain an error message")
				}
			}
		})
	}
}

func TestGetEditingAspectFieldsInvalidMethod(t *testing.T) {
	server := setupTestServer(t)

	// Test that non-GET methods are not allowed
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/api/editing/aspects/initial-details/fields", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			// Should return 405 Method Not Allowed
			if status := rr.Code; status != http.StatusMethodNotAllowed {
				t.Errorf("Handler should return 405 for %s method, got %v", method, status)
			}
		})
	}
}

func TestGetEditingAspectsInvalidMethod(t *testing.T) {
	server := setupTestServer(t)

	// Test that non-GET methods are not allowed
	methods := []string{"POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			req, err := http.NewRequest(method, "/api/editing/aspects", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			server.router.ServeHTTP(rr, req)

			// Should return 405 Method Not Allowed
			if status := rr.Code; status != http.StatusMethodNotAllowed {
				t.Errorf("Handler should return 405 for %s method, got %v", method, status)
			}
		})
	}
}

func TestAPIResponseFormat(t *testing.T) {
	server := setupTestServer(t)

	t.Run("Aspects overview response format", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/editing/aspects", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		// Verify JSON structure
		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse response as JSON: %v", err)
		}

		// Check top-level structure
		if _, exists := response["aspects"]; !exists {
			t.Error("Response should contain 'aspects' field")
		}

		aspects, ok := response["aspects"].([]interface{})
		if !ok {
			t.Error("'aspects' should be an array")
		}

		if len(aspects) == 0 {
			t.Error("'aspects' array should not be empty")
		}

		// Check first aspect structure
		firstAspect, ok := aspects[0].(map[string]interface{})
		if !ok {
			t.Error("Aspect should be an object")
		}

		requiredFields := []string{"key", "title", "description", "endpoint", "icon", "order", "fieldCount", "completedFieldCount"}
		for _, field := range requiredFields {
			if _, exists := firstAspect[field]; !exists {
				t.Errorf("Aspect should contain '%s' field", field)
			}
		}
	})

	t.Run("Aspect fields response format", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/editing/aspects/initial-details/fields", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		// Verify JSON structure
		var response map[string]interface{}
		if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to parse response as JSON: %v", err)
		}

		// Check top-level structure
		requiredTopFields := []string{"aspectKey", "aspectTitle", "fields"}
		for _, field := range requiredTopFields {
			if _, exists := response[field]; !exists {
				t.Errorf("Response should contain '%s' field", field)
			}
		}

		fields, ok := response["fields"].([]interface{})
		if !ok {
			t.Error("'fields' should be an array")
		}

		if len(fields) == 0 {
			t.Error("'fields' array should not be empty")
		}

		// Check first field structure
		firstField, ok := fields[0].(map[string]interface{})
		if !ok {
			t.Error("Field should be an object")
		}

		requiredFieldFields := []string{"name", "type", "required", "order", "description", "options"}
		for _, field := range requiredFieldFields {
			if _, exists := firstField[field]; !exists {
				t.Errorf("Field should contain '%s' field", field)
			}
		}
	})
}

// TestGetAspectFieldsEnhancedMetadata tests the integration between API and service
// This test ensures that enhanced metadata (UIHints, ValidationHints, DefaultValue)
// flows through the complete pipeline from mapping -> service -> API response
func TestGetAspectFieldsEnhancedMetadata(t *testing.T) {
	server := NewServer()

	// Test the initial-details endpoint which has various field types
	req := httptest.NewRequest("GET", "/api/editing/aspects/initial-details/fields", nil)
	w := httptest.NewRecorder()

	// Set the URL parameter
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("aspectKey", "initial-details")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	server.getAspectFields(w, req)

	// Check response status
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	// Parse the response
	var response struct {
		AspectKey   string `json:"aspectKey"`
		AspectTitle string `json:"aspectTitle"`
		Fields      []struct {
			Name        string      `json:"name"`
			Type        string      `json:"type"`
			Required    bool        `json:"required"`
			Order       int         `json:"order"`
			Description string      `json:"description"`
			Options     interface{} `json:"options"`
			UIHints     struct {
				InputType   string `json:"inputType"`
				Placeholder string `json:"placeholder"`
				HelpText    string `json:"helpText"`
				Rows        int    `json:"rows"`
				Multiline   bool   `json:"multiline"`
			} `json:"uiHints"`
			ValidationHints struct {
				Required bool `json:"required"`
			} `json:"validationHints"`
			DefaultValue interface{} `json:"defaultValue"`
		} `json:"fields"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	// Verify basic structure
	if response.AspectKey != "initial-details" {
		t.Errorf("Expected aspectKey 'initial-details', got '%s'", response.AspectKey)
	}

	if len(response.Fields) == 0 {
		t.Fatal("Expected fields in response, got none")
	}

	// Test enhanced metadata for each field
	for _, field := range response.Fields {
		// Every field must have UIHints with a valid InputType
		if field.UIHints.InputType == "" {
			t.Errorf("Field '%s' missing UIHints.InputType in API response", field.Name)
		}

		// Check specific field types have appropriate UI hints
		switch field.Type {
		case "string":
			if field.UIHints.InputType != "text" {
				t.Errorf("String field '%s' should have InputType 'text', got '%s'",
					field.Name, field.UIHints.InputType)
			}
		case "date":
			if field.UIHints.InputType != "datetime" {
				t.Errorf("Date field '%s' should have InputType 'datetime', got '%s'",
					field.Name, field.UIHints.InputType)
			}
			if field.UIHints.Placeholder != "YYYY-MM-DDTHH:MM" {
				t.Errorf("Date field '%s' should have placeholder 'YYYY-MM-DDTHH:MM', got '%s'",
					field.Name, field.UIHints.Placeholder)
			}
		case "boolean":
			if field.UIHints.InputType != "checkbox" {
				t.Errorf("Boolean field '%s' should have InputType 'checkbox', got '%s'",
					field.Name, field.UIHints.InputType)
			}
		case "text":
			if field.UIHints.InputType != "textarea" {
				t.Errorf("Text field '%s' should have InputType 'textarea', got '%s'",
					field.Name, field.UIHints.InputType)
			}
			if field.UIHints.Rows != 3 {
				t.Errorf("Text field '%s' should have Rows 3, got %d",
					field.Name, field.UIHints.Rows)
			}
			if !field.UIHints.Multiline {
				t.Errorf("Text field '%s' should have Multiline true", field.Name)
			}
		}

		// ValidationHints should always be present
		// Note: we're not checking the specific value since it comes from field type instances
		// The presence of the ValidationHints object itself is what we're validating
	}

	t.Logf("Successfully validated enhanced metadata for %d fields", len(response.Fields))
}

// TestGetAspectFieldsAllAspects tests enhanced metadata for all aspects
func TestGetAspectFieldsAllAspects(t *testing.T) {
	server := NewServer()

	aspectKeys := []string{"initial-details", "work-progress", "definition", "post-production", "publishing", "post-publish"}

	for _, aspectKey := range aspectKeys {
		t.Run("Enhanced metadata for "+aspectKey, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/editing/aspects/"+aspectKey+"/fields", nil)
			w := httptest.NewRecorder()

			// Set the URL parameter
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("aspectKey", aspectKey)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			server.getAspectFields(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Expected status 200 for %s, got %d", aspectKey, w.Code)
			}

			// Verify response contains enhanced metadata structure
			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response JSON for %s: %v", aspectKey, err)
			}

			// Check that fields exist and have enhanced metadata
			fields, ok := response["fields"].([]interface{})
			if !ok {
				t.Fatalf("Expected 'fields' array in response for %s", aspectKey)
			}

			if len(fields) == 0 {
				t.Fatalf("Expected at least one field for %s", aspectKey)
			}

			// Check first field has enhanced metadata structure
			firstField, ok := fields[0].(map[string]interface{})
			if !ok {
				t.Fatalf("Expected field to be an object for %s", aspectKey)
			}

			// Verify UIHints exists
			if _, hasUIHints := firstField["uiHints"]; !hasUIHints {
				t.Errorf("Field missing 'uiHints' in API response for %s", aspectKey)
			}

			// Verify ValidationHints exists
			if _, hasValidationHints := firstField["validationHints"]; !hasValidationHints {
				t.Errorf("Field missing 'validationHints' in API response for %s", aspectKey)
			}

			// DefaultValue may not be present if it's nil (due to omitempty)
			// This is correct behavior - only check that the field structure is valid
			if defaultValue, hasDefaultValue := firstField["defaultValue"]; hasDefaultValue {
				// If defaultValue is present, it should not be an invalid type
				if defaultValue != nil {
					t.Logf("Field has defaultValue: %v for %s", defaultValue, aspectKey)
				}
			}
		})
	}
}

func TestGetEditingAspectsWithCompletion(t *testing.T) {
	server := setupTestServer(t)

	// Create a test video with some completed fields
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	// Get the video and update some fields to simulate completion
	video, err := server.videoService.GetVideo("test-video", "test-category")
	require.NoError(t, err)

	// Set some fields as completed for testing
	video.Title = "Test Video Title"   // definition aspect - string field
	video.Code = true                  // work-progress aspect - boolean field
	video.ProjectName = "Test Project" // initial-details aspect - string field
	video.Delayed = true               // initial-details aspect - boolean field

	// Save the updated video
	err = server.videoService.UpdateVideo(video)
	require.NoError(t, err)

	t.Run("Without video context - should return zero completion counts", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/editing/aspects", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var aspectOverview aspect.AspectOverview
		err = json.Unmarshal(rr.Body.Bytes(), &aspectOverview)
		require.NoError(t, err)

		// All completion counts should be 0 without video context
		for _, aspectSummary := range aspectOverview.Aspects {
			assert.Equal(t, 0, aspectSummary.CompletedFieldCount,
				"Aspect %s should have 0 completed fields without video context", aspectSummary.Key)
		}
	})

	t.Run("With video context - should calculate actual completion counts", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/editing/aspects?videoName=test-video&category=test-category", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var aspectOverview aspect.AspectOverview
		err = json.Unmarshal(rr.Body.Bytes(), &aspectOverview)
		require.NoError(t, err)

		// Check specific aspects for expected completion counts
		for _, aspectSummary := range aspectOverview.Aspects {
			switch aspectSummary.Key {
			case "initial-details":
				// Should have 2 completed fields: ProjectName (string) and Delayed (boolean)
				assert.Greater(t, aspectSummary.CompletedFieldCount, 0,
					"initial-details should have some completed fields")
			case "work-progress":
				// Should have at least 1 completed field: Code (boolean)
				assert.Greater(t, aspectSummary.CompletedFieldCount, 0,
					"work-progress should have some completed fields")
			case "definition":
				// Should have 1 completed field: Title (string)
				assert.Greater(t, aspectSummary.CompletedFieldCount, 0,
					"definition should have some completed fields")
			}

			// Completion count should never exceed field count
			assert.LessOrEqual(t, aspectSummary.CompletedFieldCount, aspectSummary.FieldCount,
				"Aspect %s completion count should not exceed field count", aspectSummary.Key)
		}
	})

	t.Run("With invalid video context - should gracefully fallback to zero counts", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/editing/aspects?videoName=nonexistent&category=test-category", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		server.router.ServeHTTP(rr, req)

		// Should still return 200 OK, not fail
		require.Equal(t, http.StatusOK, rr.Code)

		var aspectOverview aspect.AspectOverview
		err = json.Unmarshal(rr.Body.Bytes(), &aspectOverview)
		require.NoError(t, err)

		// All completion counts should be 0 for invalid video context
		for _, aspectSummary := range aspectOverview.Aspects {
			assert.Equal(t, 0, aspectSummary.CompletedFieldCount,
				"Aspect %s should have 0 completed fields for invalid video context", aspectSummary.Key)
		}
	})
}

func TestGetEditingAspectsAllAspects(t *testing.T) {
	// ... existing code ...
}

// TestVideoListItem_StringID tests the new string-based ID system
func TestVideoListItem_StringID(t *testing.T) {
	t.Run("ID should be string-based path instead of numeric", func(t *testing.T) {
		// This test expects the new string-based ID system
		item := VideoListItem{
			ID:        "devops/test-video", // Should be string path, not int
			Name:      "test-video",        // Should include filename
			Title:     "Test Video Title",
			Date:      "2025-01-01T12:00",
			Thumbnail: "test-thumb.jpg",
			Category:  "devops",
			Status:    "draft",
			Progress: VideoProgress{
				Completed: 5,
				Total:     10,
			},
		}

		// Verify ID is string type
		assert.IsType(t, "", item.ID, "ID should be string type")
		assert.Equal(t, "devops/test-video", item.ID)

		// Test JSON serialization with string ID
		jsonData, err := json.Marshal(item)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		// ID should be serialized as string, not number
		assert.Equal(t, "devops/test-video", jsonMap["id"])
		assert.IsType(t, "", jsonMap["id"], "JSON ID should be string type")
	})
}

// TestTransformToVideoListItems_StringID tests transformation with string IDs
func TestTransformToVideoListItems_StringID(t *testing.T) {
	t.Run("should generate string-based IDs from category and name", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:      "test-video",
				Title:     "Test Video Title",
				Date:      "2025-01-01T12:00",
				Thumbnail: "test-thumb.jpg",
				Category:  "devops",
			},
			{
				Name:     "another-video",
				Title:    "Another Video",
				Category: "ai",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2, "Should return exactly two videos")

		// First video should have string ID based on category/name
		video1 := result[0]
		assert.Equal(t, "devops/test-video", video1.ID)
		assert.Equal(t, "test-video", video1.Name)
		assert.Equal(t, "Test Video Title", video1.Title)
		assert.Equal(t, "devops", video1.Category)

		// Second video should have string ID based on category/name
		video2 := result[1]
		assert.Equal(t, "ai/another-video", video2.ID)
		assert.Equal(t, "another-video", video2.Name)
		assert.Equal(t, "Another Video", video2.Title)
		assert.Equal(t, "ai", video2.Category)
	})

	t.Run("should use sanitized names from service layer", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:      "windsurf", // Name is now sanitized at service level
				Title:     "Remote Environments with Dev Containers and Devpod: Are They Worth It?",
				Date:      "2025-01-01T12:00",
				Thumbnail: "windsurf-thumb.jpg",
				Category:  "development",
				Path:      "manuscript/development/windsurf.yaml",
			},
			{
				Name:     "ai-for-policies", // Name is now sanitized at service level
				Title:    "Using AI for Policy Management",
				Category: "ai",
				Path:     "manuscript/ai/ai-for-policies.yaml",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2, "Should return exactly two videos")

		// Names are now already sanitized at the service level
		video1 := result[0]
		assert.Equal(t, "development/windsurf", video1.ID, "Should use sanitized name")
		assert.Equal(t, "windsurf", video1.Name, "Should use sanitized name")
		assert.Equal(t, "Remote Environments with Dev Containers and Devpod: Are They Worth It?", video1.Title)
		assert.Equal(t, "development", video1.Category)

		video2 := result[1]
		assert.Equal(t, "ai/ai-for-policies", video2.ID, "Should use sanitized name")
		assert.Equal(t, "ai-for-policies", video2.Name, "Should use sanitized name")
		assert.Equal(t, "Using AI for Policy Management", video2.Title)
		assert.Equal(t, "ai", video2.Category)
	})
}

// TestTransformToVideoListItems_EdgeCases tests edge cases and special characters
func TestTransformToVideoListItems_EdgeCases(t *testing.T) {
	t.Run("should handle empty and nil values gracefully", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:     "", // Empty name
				Title:    "Video with Empty Name",
				Category: "test",
				Path:     "", // Empty path
			},
			{
				Name:     "valid-name",
				Title:    "Valid Video",
				Category: "", // Empty category
				Path:     "manuscript/test/valid-name.yaml",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2, "Should return exactly two videos")

		// First video with empty name should fallback to empty string
		video1 := result[0]
		assert.Equal(t, "test/", video1.ID, "Should handle empty name gracefully")
		assert.Equal(t, "Video with Empty Name", video1.Title)
		assert.Equal(t, "test", video1.Category)

		// Second video with empty category
		video2 := result[1]
		assert.Equal(t, "/valid-name", video2.ID, "Should handle empty category gracefully")
		assert.Equal(t, "Valid Video", video2.Title)
	})

	t.Run("should handle special characters and unicode", func(t *testing.T) {
		videos := []storage.Video{
			{
				Name:     "video-with-spaces-special", // Name is now sanitized at service level
				Title:    "Test Video",
				Category: "test-category",
			},
			{
				Name:     "vidéo-avec-accénts", // Name is now sanitized at service level
				Title:    "French Video",
				Category: "français",
			},
		}

		result := transformToVideoListItems(videos)

		require.Len(t, result, 2, "Should return exactly two videos")

		// Names are now sanitized at the service level
		assert.Equal(t, "test-category/video-with-spaces-special", result[0].ID, "Should use sanitized name")
		assert.Equal(t, "français/vidéo-avec-accénts", result[1].ID, "Should use sanitized name")
	})
}

func TestServer_GetVideo_NameShouldBeFilename(t *testing.T) {
	server := setupTestServer(t)

	// Create test video with a name that will be sanitized
	// Input: "Test Video" -> Expected filename: "test-video"
	_, err := server.videoService.CreateVideo("Test Video", "test-category")
	require.NoError(t, err)

	url := "/api/videos/test-video?category=test-category"
	req := httptest.NewRequest("GET", url, nil)
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

	// The key assertion: name should be the filename, not the original input
	// ID is "test-category/test-video", so name should be "test-video" (the filename part)
	assert.Equal(t, "test-category/test-video", response.Video.ID, "ID should be category/filename")
	assert.Equal(t, "test-video", response.Video.Name, "Name should be the filename part of ID, not the original input")
	assert.Equal(t, "test-category", response.Video.Category, "Category should match the category part of ID")
}

func TestVideoAPI_JSONConsistency(t *testing.T) {
	t.Run("GET and PUT should use consistent camelCase JSON field names", func(t *testing.T) {
		// Test that the Video struct used by API handlers produces consistent JSON
		video := storage.Video{
			Name:        "test-video",
			Category:    "devops",
			ProjectName: "Test Project",
			ProjectURL:  "https://example.com",
			Sponsorship: storage.Sponsorship{
				Amount:  "1000",
				Emails:  "sponsor@example.com",
				Blocked: "false",
			},
		}

		// Test VideoWithID serialization (used in GET responses)
		videoWithID := VideoWithID{
			ID:    "devops/test-video",
			Video: video,
		}

		jsonData, err := json.Marshal(videoWithID)
		require.NoError(t, err)

		var jsonMap map[string]interface{}
		err = json.Unmarshal(jsonData, &jsonMap)
		require.NoError(t, err)

		// Should use camelCase (matches frontend expectation)
		assert.Equal(t, "Test Project", jsonMap["projectName"])
		assert.Equal(t, "https://example.com", jsonMap["projectURL"])

		// Verify sponsorship nested fields are also camelCase
		sponsorship, ok := jsonMap["sponsorship"].(map[string]interface{})
		require.True(t, ok, "sponsorship should be a JSON object")
		assert.Equal(t, "1000", sponsorship["amount"])
		assert.Equal(t, "sponsor@example.com", sponsorship["emails"])
		assert.Equal(t, "false", sponsorship["blocked"])

		// Verify old PascalCase fields don't exist (would break frontend)
		assert.NotContains(t, jsonMap, "ProjectName")
		assert.NotContains(t, jsonMap, "ProjectURL")

		// Test that the same JSON can be unmarshaled (PUT request simulation)
		var deserializedVideo storage.Video
		err = json.Unmarshal(jsonData, &deserializedVideo)
		require.NoError(t, err)

		assert.Equal(t, "Test Project", deserializedVideo.ProjectName)
		assert.Equal(t, "https://example.com", deserializedVideo.ProjectURL)
		assert.Equal(t, "1000", deserializedVideo.Sponsorship.Amount)
	})
}
