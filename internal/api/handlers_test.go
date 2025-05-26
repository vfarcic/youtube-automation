package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"devopstoolkit/youtube-automation/internal/data"
	"devopstoolkit/youtube-automation/internal/filesystem"
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
	videoService := data.NewVideoService("index.yaml", filesystem, videoManager)

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
			req = req.WithContext(chi.URLCtxKey.WithValue(req.Context(), rctx))

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
	req = req.WithContext(chi.URLCtxKey.WithValue(req.Context(), rctx))

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
	req = req.WithContext(chi.URLCtxKey.WithValue(req.Context(), rctx))

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
	req = req.WithContext(chi.URLCtxKey.WithValue(req.Context(), rctx))

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

	// Create test video
	_, err := server.videoService.CreateVideo("test-video", "test-category")
	require.NoError(t, err)

	tests := []struct {
		name     string
		endpoint string
		updates  map[string]interface{}
	}{
		{
			name:     "Update initial details",
			endpoint: "/api/videos/test-video/initial-details?category=test-category",
			updates: map[string]interface{}{
				"projectName": "Test Project",
				"projectURL":  "https://example.com",
			},
		},
		{
			name:     "Update work progress",
			endpoint: "/api/videos/test-video/work-progress?category=test-category",
			updates: map[string]interface{}{
				"codeDone":       true,
				"talkingHeadDone": true,
			},
		},
		{
			name:     "Update definition",
			endpoint: "/api/videos/test-video/definition?category=test-category",
			updates: map[string]interface{}{
				"title":       "Updated Title",
				"description": "Updated Description",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.updates)
			req := httptest.NewRequest("PUT", tt.endpoint, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Add chi context for URL parameters
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("videoName", "test-video")
			req = req.WithContext(chi.URLCtxKey.WithValue(req.Context(), rctx))

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response GetVideoResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "test-video", response.Video.Name)
		})
	}
}