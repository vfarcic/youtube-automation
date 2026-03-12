package gdrive

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
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
