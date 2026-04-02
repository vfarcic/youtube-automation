package publishing

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/storage"
)

// CopyImagesFromDrive downloads images referenced in the manuscript from Google Drive
// into the post directory. If drive is nil, it gracefully skips (returns nil).
// Missing files are logged as warnings but do not cause an error.
func CopyImagesFromDrive(ctx context.Context, drive gdrive.DriveService, videoName, driveFolderID, destDir string, filenames []string) error {
	if drive == nil || driveFolderID == "" || len(filenames) == 0 {
		return nil
	}

	// Find the video's subfolder in Drive
	videoFolderID, err := drive.FindOrCreateFolder(ctx, videoName, driveFolderID)
	if err != nil {
		return fmt.Errorf("finding video folder in Drive: %w", err)
	}

	// List files in the video folder
	driveFiles, err := drive.ListFilesInFolder(ctx, videoFolderID)
	if err != nil {
		return fmt.Errorf("listing files in Drive folder: %w", err)
	}

	// Build a map of filename -> DriveFileInfo for quick lookup
	fileMap := make(map[string]gdrive.DriveFileInfo)
	for _, f := range driveFiles {
		fileMap[f.Name] = f
	}

	// Download each referenced image
	for _, filename := range filenames {
		// Use just the base name for Drive lookup
		baseName := filepath.Base(filename)
		driveFile, found := fileMap[baseName]
		if !found {
			fmt.Printf("Warning: image %q not found in Drive folder for %q\n", baseName, videoName)
			continue
		}

		content, _, _, err := drive.GetFile(ctx, driveFile.ID)
		if err != nil {
			fmt.Printf("Warning: failed to download image %q from Drive: %v\n", baseName, err)
			continue
		}

		destPath := filepath.Join(destDir, baseName)
		if err := writeFromReader(destPath, content); err != nil {
			content.Close()
			fmt.Printf("Warning: failed to write image %q: %v\n", baseName, err)
			continue
		}
		content.Close()
	}

	return nil
}

// CopyThumbnailFromDrive downloads the thumbnail (variant index 0) from Google Drive
// and saves it as thumbnail.jpg in the post directory.
// Falls back to copying from the local path if Drive is not available.
func CopyThumbnailFromDrive(ctx context.Context, drive gdrive.DriveService, thumbnailVariants []storage.ThumbnailVariant, destDir string) error {
	if len(thumbnailVariants) == 0 {
		return nil
	}

	// Use the first variant (index 0 in the slice)
	variant := thumbnailVariants[0]

	destPath := filepath.Join(destDir, "thumbnail.jpg")

	// Try DriveFileID first (requires drive service)
	if variant.DriveFileID != "" && drive != nil {
		content, _, _, err := drive.GetFile(ctx, variant.DriveFileID)
		if err != nil {
			return fmt.Errorf("downloading thumbnail from Drive: %w", err)
		}
		defer content.Close()
		return writeFromReader(destPath, content)
	}

	// Fallback to local path
	if variant.Path != "" {
		srcContent, err := os.Open(variant.Path)
		if err != nil {
			return fmt.Errorf("opening local thumbnail: %w", err)
		}
		defer srcContent.Close()
		return writeFromReader(destPath, srcContent)
	}

	return fmt.Errorf("thumbnail variant has no DriveFileID or Path")
}

func writeFromReader(destPath string, r io.Reader) error {
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r)
	return err
}
