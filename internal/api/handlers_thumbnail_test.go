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
	"path/filepath"
	"strings"
	"testing"
	"time"

	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/thumbnail"
)

// mockImageGenerator implements thumbnail.ImageGenerator for testing.
type mockImageGenerator struct {
	name string
	data []byte
	err  error
}

func (m *mockImageGenerator) Name() string { return m.name }
func (m *mockImageGenerator) GenerateImage(_ context.Context, _ string, _ [][]byte) ([]byte, error) {
	return m.data, m.err
}

// mockDriveWithScreenshots creates a mock Drive service with screenshot files
// in the video's subfolder. Used by thumbnail generation tests.
func mockDriveWithScreenshots() *mockDriveService {
	return &mockDriveService{
		returnFileID: "drive-id",
		listFiles: []gdrive.DriveFileInfo{
			{ID: "screenshot-01", Name: "screenshot-01.png", MimeType: "image/png"},
		},
		fileContents: map[string][]byte{
			"screenshot-01": []byte("fake-photo-data"),
		},
	}
}

// --- POST /api/thumbnails/generate ---

func TestHandleGenerateThumbnails_Success(t *testing.T) {
	env := setupTestEnv(t)

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "test-provider", data: []byte("fake-png-image")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")
	env.server.SetDriveService(mockDriveWithScreenshots(), "root-folder")

	seedVideo(t, env, storage.Video{
		Name:         "test-video",
		Category:     "devops",
		Tagline:      "Hello World",
		Illustration: "a robot",
	})

	body := `{"category":"devops","name":"test-video"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp ThumbnailGenerateResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Each provider generates 2 images (with/without illustration)
	if len(resp.Thumbnails) != 2 {
		t.Errorf("expected 2 thumbnails, got %d", len(resp.Thumbnails))
	}

	for _, m := range resp.Thumbnails {
		if m.ID == "" {
			t.Error("expected non-empty ID")
		}
		if m.Provider != "test-provider" {
			t.Errorf("expected provider 'test-provider', got %q", m.Provider)
		}
	}

	// Verify images are in the store
	if store.Len() != 2 {
		t.Errorf("expected 2 images in store, got %d", store.Len())
	}
}

func TestHandleGenerateThumbnails_WithDriveScreenshots(t *testing.T) {
	env := setupTestEnv(t)

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "gemini", data: []byte("image-bytes")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")
	// Mock Drive with multiple screenshots
	driveMock := &mockDriveService{
		returnFileID: "drive-id",
		listFiles: []gdrive.DriveFileInfo{
			{ID: "s1", Name: "screenshot-01.png", MimeType: "image/png"},
			{ID: "s2", Name: "screenshot-02.png", MimeType: "image/png"},
			{ID: "other", Name: "thumbnail.png", MimeType: "image/png"}, // should be filtered out
		},
		fileContents: map[string][]byte{
			"s1": []byte("photo-data-1"),
			"s2": []byte("photo-data-2"),
		},
	}
	env.server.SetDriveService(driveMock, "root-folder")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Tagline:  "Test",
	})

	body := `{"category":"devops","name":"test-video"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleGenerateThumbnails_LocalPhotoFallback(t *testing.T) {
	env := setupTestEnv(t)

	// No Drive configured — should fall back to local photoDir
	photoDir := filepath.Join(env.tmpDir, "photos")
	if err := os.MkdirAll(photoDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(photoDir, "creator.png"), []byte("photo-data"), 0644); err != nil {
		t.Fatal(err)
	}

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "gemini", data: []byte("image-bytes")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, photoDir)

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Tagline:  "Test",
	})

	body := `{"category":"devops","name":"test-video"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleGenerateThumbnails_NoPhotosAnywhere(t *testing.T) {
	env := setupTestEnv(t)

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "test", data: []byte("data")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")
	// Drive configured but no screenshots
	env.server.SetDriveService(&mockDriveService{returnFileID: "id"}, "root-folder")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Tagline:  "Test",
	})

	body := `{"category":"devops","name":"test-video"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "No creator photos found") {
		t.Errorf("expected 'No creator photos found' message, got: %s", w.Body.String())
	}
}

func TestHandleGenerateThumbnails_NoProviders(t *testing.T) {
	env := setupTestEnv(t)
	// No generators configured

	body := `{"category":"devops","name":"test-video"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleGenerateThumbnails_NoStore(t *testing.T) {
	env := setupTestEnv(t)
	gen := &mockImageGenerator{name: "test", data: []byte("data")}
	env.server.imageGenerators = []thumbnail.ImageGenerator{gen}
	// imageStore is nil

	body := `{"category":"devops","name":"test-video"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleGenerateThumbnails_InvalidBody(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "test", data: []byte("data")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")

	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleGenerateThumbnails_MissingFields(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{"missing category", `{"name":"test"}`},
		{"missing name", `{"category":"devops"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv(t)
			store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
			gen := &mockImageGenerator{name: "test", data: []byte("data")}
			env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")

			req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			env.server.Router().ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestHandleGenerateThumbnails_PathTraversal(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "test", data: []byte("data")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")

	body := `{"category":"../etc","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleGenerateThumbnails_AllProvidersFail(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "failing", err: fmt.Errorf("API error")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")
	env.server.SetDriveService(mockDriveWithScreenshots(), "root-folder")

	seedVideo(t, env, storage.Video{Name: "test", Category: "devops", Tagline: "Hello"})

	body := `{"category":"devops","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleGenerateThumbnails_PartialFailure(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	good := &mockImageGenerator{name: "good", data: []byte("image-data")}
	bad := &mockImageGenerator{name: "bad", err: fmt.Errorf("timeout")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{good, bad}, store, "")
	env.server.SetDriveService(mockDriveWithScreenshots(), "root-folder")

	seedVideo(t, env, storage.Video{Name: "test", Category: "devops", Tagline: "Hello"})

	body := `{"category":"devops","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for partial success, got %d: %s", w.Code, w.Body.String())
	}

	var resp ThumbnailGenerateResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if len(resp.Thumbnails) == 0 {
		t.Error("expected at least some thumbnails from successful provider")
	}
	if len(resp.Errors) == 0 {
		t.Error("expected errors from failing provider")
	}
}

// --- GET /api/thumbnails/generated/{id} ---

func TestHandleGetGeneratedThumbnail_Success(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	env.server.SetThumbnailGeneration(nil, store, "")

	// Add a PNG image to the store (PNG magic bytes)
	pngData := []byte("\x89PNG\r\n\x1a\nfake-image-data")
	id, err := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     pngData,
	})
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/thumbnails/generated/"+id, nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if w.Header().Get("Content-Type") != "image/png" {
		t.Errorf("expected Content-Type image/png, got %q", w.Header().Get("Content-Type"))
	}

	if !bytes.Equal(w.Body.Bytes(), pngData) {
		t.Errorf("response body doesn't match stored image")
	}
}

func TestHandleGetGeneratedThumbnail_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	env.server.SetThumbnailGeneration(nil, store, "")

	req := httptest.NewRequest(http.MethodGet, "/api/thumbnails/generated/nonexistent-id", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleGetGeneratedThumbnail_NoStore(t *testing.T) {
	env := setupTestEnv(t)
	// imageStore is nil

	req := httptest.NewRequest(http.MethodGet, "/api/thumbnails/generated/some-id", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleGetGeneratedThumbnail_PathTraversal(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	env.server.SetThumbnailGeneration(nil, store, "")

	req := httptest.NewRequest(http.MethodGet, "/api/thumbnails/generated/../../etc/passwd", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	// chi will route ../../etc/passwd differently, but validatePathParam should catch ".."
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("expected 400 or 404 for path traversal, got %d", w.Code)
	}
}

// --- POST /api/thumbnails/generated/{id}/select ---

func TestHandleSelectGeneratedThumbnail_Success(t *testing.T) {
	env := setupTestEnv(t)

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "drive-123"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "folder-root")

	// Seed a video with one existing variant
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Path: "/local/thumb.png"},
		},
	})

	// Add image to store
	id, err := store.Add(thumbnail.GeneratedImage{
		Provider: "gemini",
		Style:    "with illustration",
		Data:     []byte("fake-png-image-data"),
	})
	if err != nil {
		t.Fatal(err)
	}

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["driveFileId"] != "drive-123" {
		t.Errorf("expected driveFileId 'drive-123', got %v", resp["driveFileId"])
	}

	// Verify the video was updated
	v, err := env.server.videoService.GetVideo("test-video", "devops")
	if err != nil {
		t.Fatalf("failed to get video: %v", err)
	}
	if v.ThumbnailVariants[0].DriveFileID != "drive-123" {
		t.Errorf("expected DriveFileID 'drive-123', got %q", v.ThumbnailVariants[0].DriveFileID)
	}

	// Verify image was removed from store
	if _, found := store.Get(id); found {
		t.Error("expected image to be removed from store after selection")
	}
}

func TestHandleSelectGeneratedThumbnail_AutoCreateVariant(t *testing.T) {
	env := setupTestEnv(t)

	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "drive-456"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "folder-root")

	// Seed a video with NO variants
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "gpt-image",
		Style:    "without illustration",
		Data:     []byte("image-data"),
	})

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	v, _ := env.server.videoService.GetVideo("test-video", "devops")
	if len(v.ThumbnailVariants) != 1 {
		t.Errorf("expected 1 variant, got %d", len(v.ThumbnailVariants))
	}
}

func TestHandleSelectGeneratedThumbnail_NoDrive(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	env.server.SetThumbnailGeneration(nil, store, "")
	// driveService is nil

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleSelectGeneratedThumbnail_ImageNotFound(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/nonexistent/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleSelectGeneratedThumbnail_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	body := `{"category":"devops","name":"nonexistent-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestHandleSelectGeneratedThumbnail_InvalidVariantIndex(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	// variantIndex 5 is out of range for a video with 0 variants
	body := `{"category":"devops","name":"test-video","variantIndex":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSelectGeneratedThumbnail_DriveUploadFails(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnErr: fmt.Errorf("drive error")}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1},
		},
	})

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	// Image is consumed by Claim to prevent TOCTOU double-upload races.
	// After a failed upload the image is no longer in the store;
	// the user can regenerate thumbnails if needed.
	if _, found := store.Get(id); found {
		t.Error("expected image to be removed from store after Claim (TOCTOU prevention)")
	}
}

func TestHandleSelectGeneratedThumbnail_InvalidBody(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader("bad-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSelectGeneratedThumbnail_MissingFields(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	body := `{"name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleSelectGeneratedThumbnail_PathTraversal(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	body := `{"category":"../etc","name":"test","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- loadPhotos helper ---

func TestLoadPhotos_EmptyDir(t *testing.T) {
	photos, err := loadPhotos("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if photos != nil {
		t.Errorf("expected nil, got %d photos", len(photos))
	}
}

func TestLoadPhotos_NonexistentDir(t *testing.T) {
	photos, err := loadPhotos("/nonexistent/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if photos != nil {
		t.Errorf("expected nil, got %d photos", len(photos))
	}
}

func TestLoadPhotos_WithImages(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "photo1.png"), []byte("png-data"), 0644)
	os.WriteFile(filepath.Join(dir, "photo2.jpg"), []byte("jpg-data"), 0644)
	os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not an image"), 0644)
	os.MkdirAll(filepath.Join(dir, "subdir"), 0755)

	photos, err := loadPhotos(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(photos) != 2 {
		t.Errorf("expected 2 photos, got %d", len(photos))
	}
}

func TestLoadPhotos_WebpAndJpeg(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.webp"), []byte("webp"), 0644)
	os.WriteFile(filepath.Join(dir, "b.jpeg"), []byte("jpeg"), 0644)

	photos, err := loadPhotos(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(photos) != 2 {
		t.Errorf("expected 2 photos, got %d", len(photos))
	}
}

// mockDriveServiceForSelect extends mockDriveService to capture uploaded data.
type mockDriveServiceForSelect struct {
	mockDriveService
	uploadedData []byte
}

func (m *mockDriveServiceForSelect) UploadFile(_ context.Context, filename string, content io.Reader, mimeType string, folderID string) (string, error) {
	m.lastFilename = filename
	m.lastMimeType = mimeType
	m.lastFolderID = folderID
	data, _ := io.ReadAll(content)
	m.uploadedData = data
	return m.returnFileID, m.returnErr
}

func TestHandleSelectGeneratedThumbnail_VerifyUploadedData(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveServiceForSelect{
		mockDriveService: mockDriveService{returnFileID: "drive-789"},
	}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "folder-root")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1},
		},
	})

	imageData := []byte("my-image-bytes")
	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "gemini",
		Style:    "without illustration",
		Data:     imageData,
	})

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if !bytes.Equal(mockDrive.uploadedData, imageData) {
		t.Errorf("uploaded data doesn't match stored image")
	}

	if mockDrive.lastFolderID != "folder-root-subfolder" {
		t.Errorf("expected folder 'folder-root-subfolder', got %q", mockDrive.lastFolderID)
	}
}

func TestHandleSelectGeneratedThumbnail_NoStore(t *testing.T) {
	env := setupTestEnv(t)
	// No store, no drive

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/some-id/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

// --- Request body size limit tests ---

func TestHandleGenerateThumbnails_OversizedBody(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	gen := &mockImageGenerator{name: "test", data: []byte("data")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")

	// Create a body larger than 1MB
	oversized := strings.Repeat("x", 1<<20+1)
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized body, got %d", w.Code)
	}
}

func TestHandleSelectGeneratedThumbnail_OversizedBody(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	// Create a body larger than 1MB
	oversized := strings.Repeat("x", 1<<20+1)
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(oversized))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for oversized body, got %d", w.Code)
	}
}

// --- Error sanitization tests ---

func TestHandleGenerateThumbnails_ErrorsDoNotLeakInternalDetails(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)

	internalMsg := "secret API key xyz123 at /internal/path/config.json"
	gen := &mockImageGenerator{name: "failing", err: fmt.Errorf("%s", internalMsg)}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{gen}, store, "")
	env.server.SetDriveService(mockDriveWithScreenshots(), "root-folder")

	seedVideo(t, env, storage.Video{Name: "test", Category: "devops", Tagline: "Hello"})

	body := `{"category":"devops","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	responseBody := w.Body.String()
	if strings.Contains(responseBody, "xyz123") {
		t.Errorf("response should not contain internal error details, got: %s", responseBody)
	}
	if strings.Contains(responseBody, "/internal/path") {
		t.Errorf("response should not contain file paths, got: %s", responseBody)
	}
}

func TestHandleGenerateThumbnails_PartialFailureErrorsSanitized(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	good := &mockImageGenerator{name: "good", data: []byte("image-data")}
	bad := &mockImageGenerator{name: "bad", err: fmt.Errorf("connection refused to internal-host:8443")}
	env.server.SetThumbnailGeneration([]thumbnail.ImageGenerator{good, bad}, store, "")
	env.server.SetDriveService(mockDriveWithScreenshots(), "root-folder")

	seedVideo(t, env, storage.Video{Name: "test", Category: "devops", Tagline: "Hello"})

	body := `{"category":"devops","name":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp ThumbnailGenerateResponse
	json.NewDecoder(w.Body).Decode(&resp)

	for _, errMsg := range resp.Errors {
		if strings.Contains(errMsg, "internal-host") {
			t.Errorf("error message should not contain internal details: %s", errMsg)
		}
		if strings.Contains(errMsg, "8443") {
			t.Errorf("error message should not contain internal port: %s", errMsg)
		}
	}
}

func TestHandleSelectGeneratedThumbnail_DriveErrorDoesNotLeak(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnErr: fmt.Errorf("oauth2: token expired for user@internal.corp")}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1},
		},
	})

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	body := `{"category":"devops","name":"test-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}

	responseBody := w.Body.String()
	if strings.Contains(responseBody, "oauth2") {
		t.Errorf("response should not contain internal error details, got: %s", responseBody)
	}
	if strings.Contains(responseBody, "internal.corp") {
		t.Errorf("response should not contain internal hostnames, got: %s", responseBody)
	}
}

func TestHandleSelectGeneratedThumbnail_VideoNotFoundErrorSanitized(t *testing.T) {
	env := setupTestEnv(t)
	store := thumbnail.NewGeneratedImageStore(10 * time.Minute)
	mockDrive := &mockDriveService{returnFileID: "x"}
	env.server.SetThumbnailGeneration(nil, store, "")
	env.server.SetDriveService(mockDrive, "")

	id, _ := store.Add(thumbnail.GeneratedImage{
		Provider: "test",
		Style:    "with illustration",
		Data:     []byte("data"),
	})

	body := `{"category":"devops","name":"nonexistent-video","variantIndex":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/thumbnails/generated/"+id+"/select", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	// The response should not contain file system paths or storage details
	var errResp ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Detail != "" {
		t.Errorf("expected empty detail for video not found, got: %q", errResp.Detail)
	}
}
