// Package thumbnail provides thumbnail resolution and upload utilities.
package thumbnail

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"devopstoolkit/youtube-automation/internal/storage"
)

// Service-level errors
var (
	ErrNoThumbnail      = errors.New("video has no thumbnail")
	ErrOpenViewerFailed = errors.New("failed to open default viewer")
)

// GetOriginalThumbnailPath extracts the thumbnail path from a video.
// It uses the first non-empty variant path,
// falling back to the deprecated Thumbnail field.
func GetOriginalThumbnailPath(video *storage.Video) (string, error) {
	for _, variant := range video.ThumbnailVariants {
		if variant.Path != "" {
			return variant.Path, nil
		}
	}

	// Fallback: deprecated Thumbnail field
	if video.Thumbnail != "" {
		return video.Thumbnail, nil
	}

	return "", ErrNoThumbnail
}

// OpenInDefaultViewer opens a file in the OS default application.
// This is cross-platform: macOS uses "open", Linux uses "xdg-open", Windows uses "start".
// The function returns immediately (non-blocking) after starting the viewer.
func OpenInDefaultViewer(filePath string) error {
	return openInDefaultViewerWithRunner(filePath, defaultCommandRunner{})
}

// commandRunner is an interface for executing commands, allowing for testing.
type commandRunner interface {
	Start(cmd *exec.Cmd) error
}

// defaultCommandRunner is the production implementation that actually runs commands.
type defaultCommandRunner struct{}

func (r defaultCommandRunner) Start(cmd *exec.Cmd) error {
	return cmd.Start()
}

// openInDefaultViewerWithRunner is the testable version that accepts a command runner.
func openInDefaultViewerWithRunner(filePath string, runner commandRunner) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", filePath)
	default:
		return fmt.Errorf("%w: unsupported OS %s", ErrOpenViewerFailed, runtime.GOOS)
	}

	if err := runner.Start(cmd); err != nil {
		return fmt.Errorf("%w: %v", ErrOpenViewerFailed, err)
	}

	return nil
}
