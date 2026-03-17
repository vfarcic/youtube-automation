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
	"net/textproto"
	"testing"

	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/storage"
)

// mockDriveService implements gdrive.DriveService for testing.
type mockDriveService struct {
	returnFileID    string
	returnErr       error
	lastFilename    string
	lastMimeType    string
	lastFolderID    string
	getFileContent  string
	getFileMIME     string
	getFileName     string
	getFileErr      error
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

func (m *mockDriveService) ListFilesInFolder(_ context.Context, _ string) ([]gdrive.DriveFileInfo, error) {
	return nil, m.returnErr
}

func (m *mockDriveService) GetFile(_ context.Context, _ string) (io.ReadCloser, string, string, error) {
	if m.getFileErr != nil {
		return nil, "", "", m.getFileErr
	}
	content := m.getFileContent
	if content == "" {
		content = "fake-video-data"
	}
	mime := m.getFileMIME
	if mime == "" {
		mime = "video/mp4"
	}
	name := m.getFileName
	if name == "" {
		name = "video.mp4"
	}
	return io.NopCloser(bytes.NewReader([]byte(content))), mime, name, nil
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

func createMultipartBodyWithMIME(t *testing.T, fieldName, filename, content, mimeType string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, filename))
	h.Set("Content-Type", mimeType)
	part, err := writer.CreatePart(h)
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

func TestHandleDriveUploadThumbnail_AutoCreateVariant(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "drive-new-123"}
	env.server.SetDriveService(mock, "")

	// Seed a video with NO thumbnail variants
	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	body, contentType := createMultipartBody(t, "thumbnail", "thumb.png", "fake-image-data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/thumbnail/test-video?category=devops&variantIndex=0", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify a variant was auto-created with the DriveFileID
	v, err := env.server.videoService.GetVideo("test-video", "devops")
	if err != nil {
		t.Fatalf("failed to get video: %v", err)
	}
	if len(v.ThumbnailVariants) != 1 {
		t.Fatalf("expected 1 variant, got %d", len(v.ThumbnailVariants))
	}
	if v.ThumbnailVariants[0].DriveFileID != "drive-new-123" {
		t.Errorf("expected DriveFileID 'drive-new-123', got '%s'", v.ThumbnailVariants[0].DriveFileID)
	}
	if v.ThumbnailVariants[0].Index != 1 {
		t.Errorf("expected Index 1, got %d", v.ThumbnailVariants[0].Index)
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

// --- Video Upload Tests ---

func TestHandleDriveUploadVideo_Success(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "video-drive-123"}
	env.server.SetDriveService(mock, "folder-abc")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Titles: []storage.TitleVariant{
			{Index: 1, Text: "My Great Video (2024)"},
		},
	})

	body, contentType := createMultipartBodyWithMIME(t, "video", "recording.mp4", "fake-video-data", "video/mp4")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/video/test-video?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["driveFileId"] != "video-drive-123" {
		t.Errorf("expected driveFileId 'video-drive-123', got '%v'", resp["driveFileId"])
	}
	if resp["videoFile"] != "drive://video-drive-123" {
		t.Errorf("expected videoFile 'drive://video-drive-123', got '%v'", resp["videoFile"])
	}

	// Verify the video was updated
	v, err := env.server.videoService.GetVideo("test-video", "devops")
	if err != nil {
		t.Fatalf("failed to get video: %v", err)
	}
	if v.VideoDriveFileID != "video-drive-123" {
		t.Errorf("expected VideoDriveFileID 'video-drive-123', got '%s'", v.VideoDriveFileID)
	}
	if v.VideoFile != "drive://video-drive-123" {
		t.Errorf("expected VideoFile 'drive://video-drive-123', got '%s'", v.VideoFile)
	}

	// Verify filename is derived from the sanitized title + .mp4 extension
	if mock.lastFilename != "my-great-video-2024.mp4" {
		t.Errorf("expected filename 'my-great-video-2024.mp4', got '%s'", mock.lastFilename)
	}
}

func TestHandleDriveUploadVideo_NoTitlesFallback(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "video-drive-456"}
	env.server.SetDriveService(mock, "folder-abc")

	seedVideo(t, env, storage.Video{
		Name:     "notitle-video",
		Category: "devops",
	})

	body, contentType := createMultipartBodyWithMIME(t, "video", "recording.mp4", "fake-video-data", "video/mp4")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/video/notitle-video?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify filename falls back to "video.mp4" when no titles
	if mock.lastFilename != "video.mp4" {
		t.Errorf("expected filename 'video.mp4', got '%s'", mock.lastFilename)
	}
}

func TestHandleDriveUploadVideo_NoDriveService(t *testing.T) {
	env := setupTestEnv(t)

	body, contentType := createMultipartBody(t, "video", "recording.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/video/test-video?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleDriveUploadVideo_MissingCategory(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	body, contentType := createMultipartBody(t, "video", "recording.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/video/test-video", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadVideo_MissingFile(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/video/test-video?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadVideo_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	body, contentType := createMultipartBody(t, "video", "recording.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/video/nonexistent?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadVideo_DriveError(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnErr: fmt.Errorf("drive quota exceeded")}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
	})

	body, contentType := createMultipartBody(t, "video", "recording.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/video/test-video?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Video Download Tests ---

func TestHandleDriveDownloadVideo_Success(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{
		returnFileID:   "abc",
		getFileContent: "video-bytes",
		getFileMIME:    "video/mp4",
		getFileName:    "my-video.mp4",
	}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:             "test-video",
		Category:         "devops",
		VideoDriveFileID: "drive-vid-456",
	})

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/video/test-video?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "video/mp4" {
		t.Errorf("expected Content-Type 'video/mp4', got '%s'", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "video-bytes" {
		t.Errorf("expected body 'video-bytes', got '%s'", w.Body.String())
	}
}

func TestHandleDriveDownloadVideo_NoFileID(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		// VideoDriveFileID is empty
	})

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/video/test-video?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveDownloadVideo_NoDriveService(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/video/test-video?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestVideoExtensionFromMIME(t *testing.T) {
	tests := []struct {
		mime string
		want string
	}{
		{"video/mp4", ".mp4"},
		{"video/webm", ".webm"},
		{"video/quicktime", ".mov"},
		{"video/x-msvideo", ".avi"},
		{"application/octet-stream", ""},
		{"", ""},
	}
	for _, tt := range tests {
		if got := videoExtensionFromMIME(tt.mime); got != tt.want {
			t.Errorf("videoExtensionFromMIME(%q) = %q, want %q", tt.mime, got, tt.want)
		}
	}
}

// --- Short Upload Tests ---

func TestHandleDriveUploadShort_Success(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "short-drive-123"}
	env.server.SetDriveService(mock, "folder-abc")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One"},
		},
	})

	body, contentType := createMultipartBodyWithMIME(t, "short", "short1.mp4", "fake-short-data", "video/mp4")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/short/test-video/short1?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["driveFileId"] != "short-drive-123" {
		t.Errorf("expected driveFileId 'short-drive-123', got '%v'", resp["driveFileId"])
	}
	if resp["filePath"] != "drive://short-drive-123" {
		t.Errorf("expected filePath 'drive://short-drive-123', got '%v'", resp["filePath"])
	}

	// Verify the video was updated
	v, err := env.server.videoService.GetVideo("test-video", "devops")
	if err != nil {
		t.Fatalf("failed to get video: %v", err)
	}
	if v.Shorts[0].DriveFileID != "short-drive-123" {
		t.Errorf("expected DriveFileID 'short-drive-123', got '%s'", v.Shorts[0].DriveFileID)
	}
	if v.Shorts[0].FilePath != "drive://short-drive-123" {
		t.Errorf("expected FilePath 'drive://short-drive-123', got '%s'", v.Shorts[0].FilePath)
	}

	// Verify nested folder: video folder → shorts subfolder
	// mock returns parentID + "-subfolder" so: folder-abc-subfolder (video) → folder-abc-subfolder-subfolder (shorts)
	if mock.lastFolderID != "folder-abc-subfolder-subfolder" {
		t.Errorf("expected folder 'folder-abc-subfolder-subfolder', got '%s'", mock.lastFolderID)
	}
	if mock.lastFilename != "short1.mp4" {
		t.Errorf("expected filename 'short1.mp4', got '%s'", mock.lastFilename)
	}
}

func TestHandleDriveUploadShort_NoDriveService(t *testing.T) {
	env := setupTestEnv(t)

	body, contentType := createMultipartBody(t, "short", "short1.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/short/test-video/short1?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleDriveUploadShort_MissingCategory(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	body, contentType := createMultipartBody(t, "short", "short1.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/short/test-video/short1", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadShort_MissingFile(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/short/test-video/short1?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadShort_VideoNotFound(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	body, contentType := createMultipartBody(t, "short", "short1.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/short/nonexistent/short1?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadShort_ShortNotFound(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One"},
		},
	})

	body, contentType := createMultipartBody(t, "short", "short1.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/short/test-video/nonexistent?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveUploadShort_DriveError(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnErr: fmt.Errorf("drive quota exceeded")}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One"},
		},
	})

	body, contentType := createMultipartBody(t, "short", "short1.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/api/drive/upload/short/test-video/short1?category=devops", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

// --- Short Download Tests ---

func TestHandleDriveDownloadShort_Success(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{
		returnFileID:   "abc",
		getFileContent: "short-video-bytes",
		getFileMIME:    "video/mp4",
		getFileName:    "short1.mp4",
	}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One", DriveFileID: "drive-short-456"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/short/test-video/short1?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("Content-Type") != "video/mp4" {
		t.Errorf("expected Content-Type 'video/mp4', got '%s'", w.Header().Get("Content-Type"))
	}
	if w.Body.String() != "short-video-bytes" {
		t.Errorf("expected body 'short-video-bytes', got '%s'", w.Body.String())
	}
}

func TestHandleDriveDownloadShort_NoFileID(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/short/test-video/short1?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveDownloadShort_NoDriveService(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/short/test-video/short1?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotImplemented {
		t.Errorf("expected 501, got %d", w.Code)
	}
}

func TestHandleDriveDownloadShort_ShortNotFound(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{returnFileID: "abc"}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One", DriveFileID: "drive-short-456"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/short/test-video/nonexistent?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDriveDownloadShort_DriveError(t *testing.T) {
	env := setupTestEnv(t)
	mock := &mockDriveService{
		returnFileID: "abc",
		getFileErr:   fmt.Errorf("drive API unavailable"),
	}
	env.server.SetDriveService(mock, "")

	seedVideo(t, env, storage.Video{
		Name:     "test-video",
		Category: "devops",
		Shorts: []storage.Short{
			{ID: "short1", Title: "Short One", DriveFileID: "drive-short-456"},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/drive/download/short/test-video/short1?category=devops", nil)
	w := httptest.NewRecorder()

	env.server.Router().ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d: %s", w.Code, w.Body.String())
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
