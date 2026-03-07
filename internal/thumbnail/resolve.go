package thumbnail

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/storage"
)

var (
	ErrEmptyRef      = errors.New("thumbnail reference is empty")
	ErrNoDriveService = errors.New("drive service is required for Drive-hosted thumbnails")
)

// ThumbnailRef holds either a local path or a Google Drive file ID for a thumbnail.
type ThumbnailRef struct {
	Path        string
	DriveFileID string
}

// IsEmpty returns true if neither path nor drive file ID is set.
func (r ThumbnailRef) IsEmpty() bool {
	return r.Path == "" && r.DriveFileID == ""
}

// ResolveThumbnail extracts the best thumbnail reference from a video.
// Priority: DriveFileID on variants > Path on variants > deprecated Thumbnail field.
func ResolveThumbnail(video *storage.Video) (ThumbnailRef, error) {
	// Check variants for DriveFileID first
	for _, v := range video.ThumbnailVariants {
		if v.DriveFileID != "" {
			return ThumbnailRef{DriveFileID: v.DriveFileID, Path: v.Path}, nil
		}
	}

	// Check variants for local Path
	for _, v := range video.ThumbnailVariants {
		if v.Path != "" {
			return ThumbnailRef{Path: v.Path}, nil
		}
	}

	// Fallback: deprecated Thumbnail field
	if video.Thumbnail != "" {
		return ThumbnailRef{Path: video.Thumbnail}, nil
	}

	return ThumbnailRef{}, ErrNoThumbnail
}

// WithThumbnailFile resolves a ThumbnailRef to a local file path and calls fn with it.
// If the ref has a local Path, fn is called directly.
// If the ref has a DriveFileID, the file is downloaded to a temp file, fn is called,
// and the temp file is cleaned up afterward.
func WithThumbnailFile(ctx context.Context, ref ThumbnailRef, drive gdrive.DriveService, fn func(localPath string) error) error {
	if ref.IsEmpty() {
		return ErrEmptyRef
	}

	// If we have a DriveFileID, download it
	if ref.DriveFileID != "" {
		if drive == nil {
			return fmt.Errorf("%w: cannot resolve Drive file ID %s", ErrNoDriveService, ref.DriveFileID)
		}

		content, _, filename, err := drive.GetFile(ctx, ref.DriveFileID)
		if err != nil {
			return fmt.Errorf("failed to download thumbnail from Drive: %w", err)
		}
		defer content.Close()

		// Determine extension from filename
		ext := filepath.Ext(filename)
		if ext == "" {
			ext = ".png"
		}

		tmpFile, err := os.CreateTemp("", "thumbnail-*"+ext)
		if err != nil {
			return fmt.Errorf("failed to create temp file for thumbnail: %w", err)
		}
		tmpPath := tmpFile.Name()
		defer os.Remove(tmpPath)

		if _, err := io.Copy(tmpFile, content); err != nil {
			tmpFile.Close()
			return fmt.Errorf("failed to write thumbnail to temp file: %w", err)
		}
		tmpFile.Close()

		return fn(tmpPath)
	}

	// Local path — use directly
	return fn(ref.Path)
}
