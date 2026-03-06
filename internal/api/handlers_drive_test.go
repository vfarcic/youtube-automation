package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

// mockDriveService implements gdrive.DriveService for testing.
type mockDriveService struct {
	returnFileID string
	returnErr    error
	lastFilename string
	lastMimeType string
	lastFolderID string
}

func (m *mockDriveService) UploadFile(_ context.Context, filename string, _ io.Reader, mimeType string, folderID string) (string, error) {
	m.lastFilename = filename
	m.lastMimeType = mimeType
	m.lastFolderID = folderID
	return m.returnFileID, m.returnErr
}

func (m *mockDriveService) FindOrCreateFolder(_ context.Context, _ string, parentID string) (string, error) {
	// Return a predictable subfolder ID based on the parent
	if m.returnErr != nil {
		return "", m.returnErr
	}
	return parentID + "-subfolder", nil
}

func createMultipartBody(t *testing.T, fieldName, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatal(err)
	}
	part.Write([]byte(content))
	writer.Close()
	return body, writer.FormDataContentType()
}

func TestHandleDriveUploadThumbnail_Success(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "drive-file-123"}
	env.server.SetDriveService(mock, "folder-abc")

	// Seed a video with one thumbnail variant
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Path: "/local/thumb.png"},
		},
	})

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "fake-image-data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?category=devops&variantIndex=0", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["driveFileId"] != "drive-file-123" {
		t.Errorf("expected driveFileId 'drive-file-123', got '%v'", resp["driveFileId"])
	}

	// Verify the video was updated with the DriveFileID
	v, err := env.server.videoService.GetVideo("test-video", "devops")
	if err != nil {
		t.Fatalf("failed to get video: %v", err)
	}
	if v.ThumbnailVariants[0].DriveFileID != "drive-file-123" {
		t.Errorf("expected DriveFileID 'drive-file-123', got '%s'", v.ThumbnailVariants[0].DriveFileID)
	}
	if mock.lastFolderID != "folder-abc-subfolder" {
		t.Errorf("expected folder 'folder-abc-subfolder', got '%s'", mock.lastFolderID)
	}
}

func TestHandleDriveUploadThumbnail_NoDriveService(t *testing.T) {
	env := setupTestEnv(t)
	// driveService is nil by default

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?category=devops&variantIndex=0", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleDriveUploadThumbnail_MissingCategory(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?variantIndex=0", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadThumbnail_MissingVariantIndex(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadThumbnail_MissingFile(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	// Send request without multipart body
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?category=devops&variantIndex=0", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadThumbnail_InvalidVariantIndex(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Path: "/local/thumb.png"},
		},
	})

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?category=devops&variantIndex=5", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadThumbnail_DriveUploadError(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnErr: fmt.Errorf("drive quota exceeded")}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Path: "/local/thumb.png"},
		},
	})

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?category=devops&variantIndex=0", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadThumbnail_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/nonexistent?category=devops&variantIndex=0", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestExtensionFromMIME(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"image/png", ".png"},
		{"image/jpeg", ".jpg"},
		{"image/webp", ".webp"},
		{"image/gif", ".gif"},
		{"application/octet-stream", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := extensionFromMIME(tt.mime); got != tt.want {
			t.Errorf("extensionFromMIME(%q) = %q, want %q", tt.mime, got, tt.want)
		}
	}
}
