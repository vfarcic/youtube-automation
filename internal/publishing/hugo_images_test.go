package publishing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDrive implements gdrive.DriveService for image/thumbnail tests.
type mockDrive struct {
	folders    map[string]string                   // name -> folderID
	files      map[string][]gdrive.DriveFileInfo   // folderID -> files
	fileData   map[string]string                   // fileID -> content
	findErr    error
	listErr    error
	getFileErr error
}

func newMockDrive() *mockDrive {
	return &mockDrive{
		folders:  make(map[string]string),
		files:    make(map[string][]gdrive.DriveFileInfo),
		fileData: make(map[string]string),
	}
}

func (m *mockDrive) UploadFile(_ context.Context, _ string, _ io.Reader, _ string, _ string) (string, error) {
	return "", nil
}

func (m *mockDrive) FindOrCreateFolder(_ context.Context, name string, parentID string) (string, error) {
	if m.findErr != nil {
		return "", m.findErr
	}
	key := parentID + "/" + name
	if id, ok := m.folders[key]; ok {
		return id, nil
	}
	return parentID + "-" + name, nil
}

func (m *mockDrive) GetFile(_ context.Context, fileID string) (io.ReadCloser, string, string, error) {
	if m.getFileErr != nil {
		return nil, "", "", m.getFileErr
	}
	data, ok := m.fileData[fileID]
	if !ok {
		return nil, "", "", fmt.Errorf("file not found: %s", fileID)
	}
	return io.NopCloser(bytes.NewReader([]byte(data))), "image/png", fileID, nil
}

func (m *mockDrive) ListFilesInFolder(_ context.Context, folderID string) ([]gdrive.DriveFileInfo, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.files[folderID], nil
}

func TestCopyImagesFromDrive(t *testing.T) {
	t.Run("nil drive skips gracefully", func(t *testing.T) {
		err := CopyImagesFromDrive(context.Background(), nil, "video", "folder", t.TempDir(), []string{"img.png"})
		assert.NoError(t, err)
	})

	t.Run("empty filenames skips", func(t *testing.T) {
		err := CopyImagesFromDrive(context.Background(), newMockDrive(), "video", "folder", t.TempDir(), nil)
		assert.NoError(t, err)
	})

	t.Run("downloads matching images", func(t *testing.T) {
		drive := newMockDrive()
		drive.folders["root-folder/my-video"] = "video-folder-id"
		drive.files["video-folder-id"] = []gdrive.DriveFileInfo{
			{ID: "file1", Name: "diagram.png", MimeType: "image/png"},
			{ID: "file2", Name: "screenshot.jpg", MimeType: "image/jpeg"},
		}
		drive.fileData["file1"] = "png-data"
		drive.fileData["file2"] = "jpg-data"

		destDir := t.TempDir()
		err := CopyImagesFromDrive(context.Background(), drive, "my-video", "root-folder", destDir, []string{"diagram.png", "screenshot.jpg"})
		require.NoError(t, err)

		data1, err := os.ReadFile(filepath.Join(destDir, "diagram.png"))
		require.NoError(t, err)
		assert.Equal(t, "png-data", string(data1))

		data2, err := os.ReadFile(filepath.Join(destDir, "screenshot.jpg"))
		require.NoError(t, err)
		assert.Equal(t, "jpg-data", string(data2))
	})

	t.Run("missing image warns but continues", func(t *testing.T) {
		drive := newMockDrive()
		drive.folders["root/video"] = "vfolder"
		drive.files["vfolder"] = []gdrive.DriveFileInfo{}

		destDir := t.TempDir()
		err := CopyImagesFromDrive(context.Background(), drive, "video", "root", destDir, []string{"missing.png"})
		assert.NoError(t, err) // Should not error
		assert.NoFileExists(t, filepath.Join(destDir, "missing.png"))
	})

	t.Run("find folder error", func(t *testing.T) {
		drive := newMockDrive()
		drive.findErr = fmt.Errorf("drive error")

		err := CopyImagesFromDrive(context.Background(), drive, "video", "root", t.TempDir(), []string{"img.png"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "finding video folder")
	})

	t.Run("list files error", func(t *testing.T) {
		drive := newMockDrive()
		drive.listErr = fmt.Errorf("list error")

		err := CopyImagesFromDrive(context.Background(), drive, "video", "root", t.TempDir(), []string{"img.png"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "listing files")
	})
}

func TestCopyThumbnailFromDrive(t *testing.T) {
	t.Run("nil drive skips gracefully", func(t *testing.T) {
		err := CopyThumbnailFromDrive(context.Background(), nil, []storage.ThumbnailVariant{{DriveFileID: "x"}}, t.TempDir())
		assert.NoError(t, err)
	})

	t.Run("empty variants skips", func(t *testing.T) {
		err := CopyThumbnailFromDrive(context.Background(), newMockDrive(), nil, t.TempDir())
		assert.NoError(t, err)
	})

	t.Run("downloads from DriveFileID", func(t *testing.T) {
		drive := newMockDrive()
		drive.fileData["thumb-id"] = "thumbnail-content"

		destDir := t.TempDir()
		variants := []storage.ThumbnailVariant{{Index: 1, DriveFileID: "thumb-id"}}

		err := CopyThumbnailFromDrive(context.Background(), drive, variants, destDir)
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(destDir, "thumbnail.jpg"))
		require.NoError(t, err)
		assert.Equal(t, "thumbnail-content", string(data))
	})

	t.Run("falls back to local path", func(t *testing.T) {
		// Create a local thumbnail file
		srcDir := t.TempDir()
		srcPath := filepath.Join(srcDir, "thumb.png")
		require.NoError(t, os.WriteFile(srcPath, []byte("local-thumb"), 0644))

		destDir := t.TempDir()
		variants := []storage.ThumbnailVariant{{Index: 1, Path: srcPath}}

		err := CopyThumbnailFromDrive(context.Background(), newMockDrive(), variants, destDir)
		require.NoError(t, err)

		data, err := os.ReadFile(filepath.Join(destDir, "thumbnail.jpg"))
		require.NoError(t, err)
		assert.Equal(t, "local-thumb", string(data))
	})

	t.Run("drive download error", func(t *testing.T) {
		drive := newMockDrive()
		drive.getFileErr = fmt.Errorf("download failed")

		variants := []storage.ThumbnailVariant{{Index: 1, DriveFileID: "bad-id"}}
		err := CopyThumbnailFromDrive(context.Background(), drive, variants, t.TempDir())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "downloading thumbnail")
	})

	t.Run("local path not found error", func(t *testing.T) {
		variants := []storage.ThumbnailVariant{{Index: 1, Path: "/nonexistent/thumb.png"}}
		err := CopyThumbnailFromDrive(context.Background(), newMockDrive(), variants, t.TempDir())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "opening local thumbnail")
	})
}
