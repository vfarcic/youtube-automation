// Package thumbnail provides AI-powered thumbnail localization using Google Gemini.
package thumbnail

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"devopstoolkit/youtube-automation/internal/gdrive"
	"devopstoolkit/youtube-automation/internal/storage"
)

// Service-level errors
var (
	ErrNoThumbnail       = errors.New("video has no thumbnail")
	ErrNoTagline         = errors.New("video has no tagline")
	ErrSaveFailed        = errors.New("failed to save thumbnail")
	ErrOpenViewerFailed  = errors.New("failed to open default viewer")
)

// ThumbnailGenerator defines the interface for generating localized thumbnails.
// This allows for easy mocking in tests.
type ThumbnailGenerator interface {
	GenerateLocalizedThumbnail(ctx context.Context, imagePath, tagline, targetLang string) ([]byte, error)
}

// LocalizeThumbnail generates a localized thumbnail and saves it to disk.
// It reads the original thumbnail path and tagline from the video,
// calls the generator to create the localized version,
// and saves the result to disk with a language suffix.
// Returns the path to the saved thumbnail.
func LocalizeThumbnail(ctx context.Context, generator ThumbnailGenerator, video *storage.Video, langCode string, driveService gdrive.DriveService) (string, error) {
	// Validate language
	if !IsSupportedLanguage(langCode) {
		return "", fmt.Errorf("%w: %s", ErrUnsupportedLang, langCode)
	}

	// Resolve the thumbnail reference (supports both local and Drive-hosted)
	ref, err := ResolveThumbnail(video)
	if err != nil {
		return "", err
	}

	// Get the tagline
	if video.Tagline == "" {
		return "", ErrNoTagline
	}

	// Determine output path from local path if available, or construct from video name
	var outputBase string
	if localPath, pathErr := GetOriginalThumbnailPath(video); pathErr == nil {
		outputBase = localPath
	} else {
		// Drive-only thumbnail: construct output path from video name
		outputBase = filepath.Join(".", video.Name+"-thumbnail.png")
	}
	outputPath := GetLocalizedThumbnailPath(outputBase, langCode)

	// Generate the localized thumbnail (may download from Drive first)
	var imageBytes []byte
	genErr := WithThumbnailFile(ctx, ref, driveService, func(localPath string) error {
		var genInnerErr error
		imageBytes, genInnerErr = generator.GenerateLocalizedThumbnail(ctx, localPath, video.Tagline, langCode)
		return genInnerErr
	})
	if genErr != nil {
		return "", fmt.Errorf("failed to generate localized thumbnail: %w", genErr)
	}

	// Save the generated image
	if err := os.WriteFile(outputPath, imageBytes, 0644); err != nil {
		return "", fmt.Errorf("%w: %v", ErrSaveFailed, err)
	}

	return outputPath, nil
}

// GetLocalizedThumbnailPath constructs the output path for a localized thumbnail.
// It inserts the language code before the file extension.
// e.g., "/path/to/thumbnail.png" + "es" -> "/path/to/thumbnail-es.png"
func GetLocalizedThumbnailPath(originalPath, langCode string) string {
	ext := filepath.Ext(originalPath)
	base := strings.TrimSuffix(originalPath, ext)
	return fmt.Sprintf("%s-%s%s", base, langCode, ext)
}

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
