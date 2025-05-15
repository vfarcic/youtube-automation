package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// MoveFile safely moves a file from src to dst.
// It creates the destination directory if it doesn't exist.
// It returns an error if the destination file already exists.
// It attempts to preserve the source file's permissions on the destination file.
func MoveFile(src, dst string) error {
	// Create target directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", filepath.Dir(dst), err)
	}

	// Check if destination file already exists
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("destination file already exists: %s", dst)
	} else if !os.IsNotExist(err) {
		// If os.Stat returned an error other than NotExist, it's a problem.
		return fmt.Errorf("failed to check destination file %s: %w", dst, err)
	}

	// Get source file info to preserve permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source file does not exist: %s", src)
		}
		return fmt.Errorf("failed to get source file info for %s: %w", src, err)
	}

	// Move the file
	if err := os.Rename(src, dst); err != nil {
		return fmt.Errorf("failed to move file from %s to %s: %w", src, dst, err)
	}

	// Ensure permissions are maintained
	if err := os.Chmod(dst, srcInfo.Mode()); err != nil {
		// If chmod fails, the file is moved but permissions are not set.
		// This might be acceptable in some cases, but we should return an error
		// to indicate that not all operations were successful.
		// Consider logging a warning and not returning error if partial success is ok.
		return fmt.Errorf("failed to set permissions on destination file %s: %w", dst, err)
	}

	return nil
}
