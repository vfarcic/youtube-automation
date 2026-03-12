package gdrive

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// mockDriveService implements DriveService for testing.
type mockDriveService struct {
	uploadedFilename string
	uploadedMimeType string
	uploadedFolderID string
	uploadedContent  string
	returnFileID     string
	returnErr        error
}

func (m *mockDriveService) FindOrCreateFolder(_ context.Context, _ string, parentID string) (string, error) {
	if m.returnErr != nil {
		return "", m.returnErr
	}
	return parentID + "-subfolder", nil
}

func (m *mockDriveService) GetFile(_ context.Context, _ string) (io.ReadCloser, string, string, error) {
	return nil, "", "", nil
}

func (m *mockDriveService) ListFilesInFolder(_ context.Context, _ string) ([]DriveFileInfo, error) {
	return nil, m.returnErr
}

func (m *mockDriveService) UploadFile(_ context.Context, filename string, content io.Reader, mimeType string, folderID string) (string, error) {
	m.uploadedFilename = filename
	m.uploadedMimeType = mimeType
	m.uploadedFolderID = folderID
	if content != nil {
		b, _ := io.ReadAll(content)
		m.uploadedContent = string(b)
	}
	return m.returnFileID, m.returnErr
}

func TestMockDriveService_UploadFile(t *testing.T) {
	mock := &mockDriveService{
		returnFileID: "abc123",
	}

	fileID, err := mock.UploadFile(context.Background(), "thumb.png", strings.NewReader("image data"), "image/png", "folder-id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileID != "abc123" {
		t.Errorf("expected fileID 'abc123', got '%s'", fileID)
	}
	if mock.uploadedFilename != "thumb.png" {
		t.Errorf("expected filename 'thumb.png', got '%s'", mock.uploadedFilename)
	}
	if mock.uploadedMimeType != "image/png" {
		t.Errorf("expected mimeType 'image/png', got '%s'", mock.uploadedMimeType)
	}
	if mock.uploadedFolderID != "folder-id" {
		t.Errorf("expected folderID 'folder-id', got '%s'", mock.uploadedFolderID)
	}
	if mock.uploadedContent != "image data" {
		t.Errorf("expected content 'image data', got '%s'", mock.uploadedContent)
	}
}

func TestMockDriveService_UploadFile_Error(t *testing.T) {
	mock := &mockDriveService{
		returnErr: context.DeadlineExceeded,
	}

	_, err := mock.UploadFile(context.Background(), "thumb.png", strings.NewReader("data"), "image/png", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestMockDriveService_UploadFile_EmptyFolder(t *testing.T) {
	mock := &mockDriveService{
		returnFileID: "xyz789",
	}

	fileID, err := mock.UploadFile(context.Background(), "thumb.jpg", strings.NewReader("jpeg data"), "image/jpeg", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fileID != "xyz789" {
		t.Errorf("expected fileID 'xyz789', got '%s'", fileID)
	}
	if mock.uploadedFolderID != "" {
		t.Errorf("expected empty folderID, got '%s'", mock.uploadedFolderID)
	}
}

func newTestDriveService(t *testing.T, handler http.Handler) *driveService {
	t.Helper()
	ts := httptest.NewServer(handler)
	t.Cleanup(ts.Close)

	srv, err := drive.NewService(context.Background(),
		option.WithHTTPClient(ts.Client()),
		option.WithEndpoint(ts.URL),
	)
	if err != nil {
		t.Fatalf("creating test drive service: %v", err)
	}
	return &driveService{service: srv}
}

func TestListFilesInFolder_SinglePage(t *testing.T) {
	ds := newTestDriveService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"files": []map[string]string{
				{"id": "f1", "name": "image1.png", "mimeType": "image/png"},
				{"id": "f2", "name": "image2.jpg", "mimeType": "image/jpeg"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	files, err := ds.ListFilesInFolder(context.Background(), "folder-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
	if files[0].ID != "f1" || files[0].Name != "image1.png" {
		t.Errorf("unexpected first file: %+v", files[0])
	}
	if files[1].ID != "f2" || files[1].Name != "image2.jpg" {
		t.Errorf("unexpected second file: %+v", files[1])
	}
}

func TestListFilesInFolder_Pagination(t *testing.T) {
	callCount := 0
	ds := newTestDriveService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp map[string]any
		if r.URL.Query().Get("pageToken") == "" {
			resp = map[string]any{
				"nextPageToken": "page2-token",
				"files": []map[string]string{
					{"id": "f1", "name": "a.png", "mimeType": "image/png"},
				},
			}
		} else {
			resp = map[string]any{
				"files": []map[string]string{
					{"id": "f2", "name": "b.png", "mimeType": "image/png"},
					{"id": "f3", "name": "c.png", "mimeType": "image/png"},
				},
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	files, err := ds.ListFilesInFolder(context.Background(), "folder-xyz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", callCount)
	}
	if files[2].ID != "f3" || files[2].Name != "c.png" {
		t.Errorf("unexpected third file: %+v", files[2])
	}
}

func TestListFilesInFolder_Error(t *testing.T) {
	ds := newTestDriveService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"message": "internal error", "code": 500}}`))
	}))

	_, err := ds.ListFilesInFolder(context.Background(), "folder-bad")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unable to list files in folder") {
		t.Errorf("expected wrapped error message, got: %v", err)
	}
}

func TestListFilesInFolder_EmptyFolder(t *testing.T) {
	ds := newTestDriveService(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"files": []map[string]string{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))

	files, err := ds.ListFilesInFolder(context.Background(), "empty-folder")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestNewDriveService_WithClient(t *testing.T) {
	// NewDriveService with a valid (but unauthenticated) client should not error
	ds, err := NewDriveService(context.Background(), &http.Client{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds == nil {
		t.Error("expected non-nil DriveService")
	}
}
