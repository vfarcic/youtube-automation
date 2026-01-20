package dubbing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// MaxFileSizeBytes is the ElevenLabs file size limit (1GB)
	MaxFileSizeBytes = 1024 * 1024 * 1024
	// TargetSizeMB is the target compressed size (~900MB to stay safely under 1GB)
	TargetSizeMB = 900
	// MaxDurationFor4K is the max duration (seconds) where 4K compression is attempted
	MaxDurationFor4K = 25 * 60 // 25 minutes
	// DefaultCRF1080p is the CRF used for 1080p compression
	DefaultCRF1080p = 26
	// MinCRF is the minimum CRF (highest quality)
	MinCRF = 23
	// MaxCRF is the maximum CRF before switching to 1080p
	MaxCRF = 30
)

// Errors returned by compression functions
var (
	ErrFFmpegNotFound  = errors.New("ffmpeg not found in PATH")
	ErrFFprobeNotFound = errors.New("ffprobe not found in PATH")
	ErrFileNotFound    = errors.New("video file not found")
	ErrCompressionFailed = errors.New("video compression failed")
)

// VideoInfo holds metadata extracted via FFprobe
type VideoInfo struct {
	Duration float64 // seconds
	Size     int64   // bytes
	Width    int     // pixels
	Height   int     // pixels
	FilePath string
}

// ffprobeOutput represents the JSON output from ffprobe
type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
	} `json:"format"`
	Streams []struct {
		Width  int `json:"width,omitempty"`
		Height int `json:"height,omitempty"`
	} `json:"streams"`
}

// CommandExecutor is an interface for executing commands (allows mocking in tests)
type CommandExecutor interface {
	ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error)
	ExecuteCommandWithStderr(ctx context.Context, name string, args ...string) ([]byte, []byte, error)
}

// RealCommandExecutor executes actual system commands
type RealCommandExecutor struct{}

// ExecuteCommand runs a command and returns stdout
func (r *RealCommandExecutor) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

// ExecuteCommandWithStderr runs a command and returns both stdout and stderr
func (r *RealCommandExecutor) ExecuteCommandWithStderr(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// defaultExecutor is used when no executor is provided
var defaultExecutor CommandExecutor = &RealCommandExecutor{}

// GetVideoInfo extracts video metadata using FFprobe
func GetVideoInfo(filePath string) (*VideoInfo, error) {
	return GetVideoInfoWithExecutor(context.Background(), filePath, defaultExecutor)
}

// GetVideoInfoWithExecutor extracts video metadata using a custom executor (for testing)
func GetVideoInfoWithExecutor(ctx context.Context, filePath string, executor CommandExecutor) (*VideoInfo, error) {
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Run ffprobe to get video info
	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	output, err := executor.ExecuteCommand(ctx, "ffprobe", args...)
	if err != nil {
		if isCommandNotFound(err) {
			return nil, ErrFFprobeNotFound
		}
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	var probeResult ffprobeOutput
	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Parse duration
	var duration float64
	if probeResult.Format.Duration != "" {
		fmt.Sscanf(probeResult.Format.Duration, "%f", &duration)
	}

	// Find video stream dimensions
	var width, height int
	for _, stream := range probeResult.Streams {
		if stream.Width > 0 && stream.Height > 0 {
			width = stream.Width
			height = stream.Height
			break
		}
	}

	return &VideoInfo{
		Duration: duration,
		Size:     fileInfo.Size(),
		Width:    width,
		Height:   height,
		FilePath: filePath,
	}, nil
}

// CalculateOptimalCRF determines the best CRF value to hit target size while maintaining quality
// CRF range: 23 (high quality) to 30 (more compression)
// Returns crf value and whether to use 1080p instead of 4K
func CalculateOptimalCRF(durationSec float64, currentSizeMB float64, targetSizeMB int) (crf int, use1080p bool) {
	// For longer videos, always use 1080p
	if durationSec > MaxDurationFor4K {
		return DefaultCRF1080p, true
	}

	// Calculate compression ratio needed
	compressionRatio := currentSizeMB / float64(targetSizeMB)

	// Estimate CRF based on compression ratio
	// CRF increases logarithmically with compression ratio
	// Each +6 CRF roughly halves the file size
	if compressionRatio <= 1 {
		return MinCRF, false // No compression needed, use high quality
	}

	// Calculate CRF: start at 23, add based on how much compression is needed
	// log2(compressionRatio) * 6 gives us the CRF increase needed
	crfIncrease := int(log2(compressionRatio) * 6)
	crf = MinCRF + crfIncrease

	// If CRF would exceed max, switch to 1080p
	if crf > MaxCRF {
		return DefaultCRF1080p, true
	}

	return crf, false
}

// log2 calculates log base 2
func log2(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// log2(x) = ln(x) / ln(2)
	// Using approximation: count how many times we can divide by 2
	result := 0.0
	for x >= 2 {
		x /= 2
		result++
	}
	// Add fractional part
	if x > 1 {
		result += (x - 1)
	}
	return result
}

// CompressForDubbing compresses video to fit under 1GB limit while maximizing quality
// Returns the path to use (original if no compression needed, or compressed file path)
func CompressForDubbing(ctx context.Context, inputPath string) (string, error) {
	return CompressForDubbingWithExecutor(ctx, inputPath, defaultExecutor)
}

// CompressForDubbingWithExecutor compresses video using a custom executor (for testing)
func CompressForDubbingWithExecutor(ctx context.Context, inputPath string, executor CommandExecutor) (string, error) {
	// Get video info
	info, err := GetVideoInfoWithExecutor(ctx, inputPath, executor)
	if err != nil {
		return "", err
	}

	// If file is already under limit, return original path
	if info.Size <= MaxFileSizeBytes {
		return inputPath, nil
	}

	// Calculate compression parameters
	currentSizeMB := float64(info.Size) / (1024 * 1024)
	crf, use1080p := CalculateOptimalCRF(info.Duration, currentSizeMB, TargetSizeMB)

	// Build output path
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)
	outputPath := filepath.Join(filepath.Dir(inputPath), baseName+"_compressed.mp4")

	// Build FFmpeg arguments
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-crf", fmt.Sprintf("%d", crf),
		"-preset", "medium",
		"-c:a", "aac",
		"-b:a", "128k",
	}

	// Add resolution scaling if using 1080p
	if use1080p {
		args = append(args, "-vf", "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2")
	}

	// Add output file (overwrite if exists)
	args = append(args, "-y", outputPath)

	// Execute FFmpeg
	_, stderr, err := executor.ExecuteCommandWithStderr(ctx, "ffmpeg", args...)
	if err != nil {
		if isCommandNotFound(err) {
			return "", ErrFFmpegNotFound
		}
		return "", fmt.Errorf("%w: %s", ErrCompressionFailed, string(stderr))
	}

	// Verify output file exists and is under the limit
	outputInfo, err := os.Stat(outputPath)
	if err != nil {
		return "", fmt.Errorf("compressed file not created: %w", err)
	}

	if outputInfo.Size() > MaxFileSizeBytes {
		// Compression didn't achieve target - try again with 1080p if we haven't already
		if !use1080p {
			os.Remove(outputPath) // Clean up failed attempt
			return compressTo1080p(ctx, inputPath, outputPath, executor)
		}
		// Already tried 1080p, return what we have (it's closer to the limit)
	}

	return outputPath, nil
}

// compressTo1080p forces 1080p compression when 4K compression didn't achieve target
func compressTo1080p(ctx context.Context, inputPath, outputPath string, executor CommandExecutor) (string, error) {
	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-crf", fmt.Sprintf("%d", DefaultCRF1080p),
		"-preset", "medium",
		"-vf", "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2",
		"-c:a", "aac",
		"-b:a", "128k",
		"-y", outputPath,
	}

	_, stderr, err := executor.ExecuteCommandWithStderr(ctx, "ffmpeg", args...)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrCompressionFailed, string(stderr))
	}

	return outputPath, nil
}

// isCommandNotFound checks if an error is due to command not found
func isCommandNotFound(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "executable file not found") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "no such file or directory")
}

// NeedsCompression checks if a video file needs compression for ElevenLabs
func NeedsCompression(filePath string) (bool, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return false, err
	}
	return info.Size() > MaxFileSizeBytes, nil
}
