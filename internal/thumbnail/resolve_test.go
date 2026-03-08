package thumbnail

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

// mockDriveService implements gdrive.DriveService for testing.
type mockDriveService struct {
	content  string
	mimeType string
	filename string
	err      error
}

func (m *mockDriveService) UploadFile(ctx context.Context, filename string, content io.Reader, mimeType string, folderID string) (string, error) {
	return "", nil
}

func (m *mockDriveService) FindOrCreateFolder(ctx context.Context, name string, parentID string) (string, error) {
	return "", nil
}

func (m *mockDriveService) GetFile(ctx context.Context, fileID string) (io.ReadCloser, string, string, error) {
	if m.err != nil {
		return nil, "", "", m.err
	}
	return io.NopCloser(strings.NewReader(m.content)), m.mimeType, m.filename, nil
}

func TestResolveThumbnail(t *testing.T) {
	tests := []struct {
		name        string
		video       *storage.Video
		wantDrive   string
		wantPath    string
		wantErr     error
	}{
		{
			name: "DriveFileID only",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, DriveFileID: "drive-abc"},
				},
			},
			wantDrive: "drive-abc",
			wantPath:  "",
		},
		{
			name: "Path only",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: "/local/thumb.png"},
				},
			},
			wantDrive: "",
			wantPath:  "/local/thumb.png",
		},
		{
			name: "Both Drive and Path - Drive wins",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: "/local/thumb.png", DriveFileID: "drive-xyz"},
				},
			},
			wantDrive: "drive-xyz",
			wantPath:  "/local/thumb.png",
		},
		{
			name: "Deprecated Thumbnail fallback",
			video: &storage.Video{
				Thumbnail: "/legacy/thumb.jpg",
			},
			wantDrive: "",
			wantPath:  "/legacy/thumb.jpg",
		},
		{
			name:    "No thumbnail - error",
			video:   &storage.Video{},
			wantErr: ErrNoThumbnail,
		},
		{
			name: "Empty variants and empty deprecated field",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: "", DriveFileID: ""},
				},
				Thumbnail: "",
			},
			wantErr: ErrNoThumbnail,
		},
		{
			name: "DriveFileID on second variant",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: ""},
					{Index: 2, DriveFileID: "drive-second"},
				},
			},
			wantDrive: "drive-second",
			wantPath:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ref, err := ResolveThumbnail(tt.video)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("ResolveThumbnail() expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("ResolveThumbnail() error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveThumbnail() unexpected error: %v", err)
			}
			if ref.DriveFileID != tt.wantDrive {
				t.Errorf("DriveFileID = %q, want %q", ref.DriveFileID, tt.wantDrive)
			}
			if ref.Path != tt.wantPath {
				t.Errorf("Path = %q, want %q", ref.Path, tt.wantPath)
			}
		})
	}
}

func TestWithThumbnailFile(t *testing.T) {
	ctx := context.Background()

	t.Run("local path passthrough", func(t *testing.T) {
		ref := ThumbnailRef{Path: "/some/local/path.png"}
		var gotPath string
		err := WithThumbnailFile(ctx, ref, nil, func(localPath string) error {
			gotPath = localPath
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotPath != "/some/local/path.png" {
			t.Errorf("fn called with path = %q, want %q", gotPath, "/some/local/path.png")
		}
	})

	t.Run("Drive download and temp file cleanup", func(t *testing.T) {
		drive := &mockDriveService{
			content:  "fake image data",
			mimeType: "image/png",
			filename: "thumbnail.png",
		}
		ref := ThumbnailRef{DriveFileID: "drive-123"}
		var gotPath string
		err := WithThumbnailFile(ctx, ref, drive, func(localPath string) error {
			gotPath = localPath
			// Verify file exists and has correct content
			data, readErr := os.ReadFile(localPath)
			if readErr != nil {
				return fmt.Errorf("failed to read temp file: %w", readErr)
			}
			if string(data) != "fake image data" {
				return fmt.Errorf("temp file content = %q, want %q", string(data), "fake image data")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Verify temp file was cleaned up
		if _, statErr := os.Stat(gotPath); !os.IsNotExist(statErr) {
			t.Errorf("temp file %q should have been cleaned up", gotPath)
		}
	})

	t.Run("cleanup on fn error", func(t *testing.T) {
		drive := &mockDriveService{
			content:  "data",
			mimeType: "image/png",
			filename: "thumb.png",
		}
		ref := ThumbnailRef{DriveFileID: "drive-456"}
		fnErr := errors.New("fn failed")
		var gotPath string
		err := WithThumbnailFile(ctx, ref, drive, func(localPath string) error {
			gotPath = localPath
			return fnErr
		})
		if !errors.Is(err, fnErr) {
			t.Errorf("error = %v, want %v", err, fnErr)
		}
		// Verify temp file was cleaned up even on error
		if _, statErr := os.Stat(gotPath); !os.IsNotExist(statErr) {
			t.Errorf("temp file %q should have been cleaned up after fn error", gotPath)
		}
	})

	t.Run("nil drive with DriveFileID returns error", func(t *testing.T) {
		ref := ThumbnailRef{DriveFileID: "drive-789"}
		err := WithThumbnailFile(ctx, ref, nil, func(localPath string) error {
			return nil
		})
		if err == nil {
			t.Fatal("expected error for nil drive, got nil")
		}
		if !errors.Is(err, ErrNoDriveService) {
			t.Errorf("error = %v, want ErrNoDriveService", err)
		}
	})

	t.Run("empty ref returns error", func(t *testing.T) {
		ref := ThumbnailRef{}
		err := WithThumbnailFile(ctx, ref, nil, func(localPath string) error {
			return nil
		})
		if !errors.Is(err, ErrEmptyRef) {
			t.Errorf("error = %v, want ErrEmptyRef", err)
		}
	})

	t.Run("Drive download error", func(t *testing.T) {
		driveErr := errors.New("network error")
		drive := &mockDriveService{err: driveErr}
		ref := ThumbnailRef{DriveFileID: "drive-fail"}
		err := WithThumbnailFile(ctx, ref, drive, func(localPath string) error {
			return nil
		})
		if err == nil {
			t.Fatal("expected error for Drive download failure, got nil")
		}
		if !errors.Is(err, driveErr) {
			t.Errorf("error should wrap %v, got %v", driveErr, err)
		}
	})

	t.Run("Drive file with no extension defaults to .png", func(t *testing.T) {
		drive := &mockDriveService{
			content:  "data",
			mimeType: "image/png",
			filename: "noext",
		}
		ref := ThumbnailRef{DriveFileID: "drive-noext"}
		var gotPath string
		err := WithThumbnailFile(ctx, ref, drive, func(localPath string) error {
			gotPath = localPath
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasSuffix(gotPath, ".png") {
			t.Errorf("temp file path %q should end with .png", gotPath)
		}
	})
}
