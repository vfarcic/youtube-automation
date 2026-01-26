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
	// TargetSizeBytes is the target compressed size (~900MB to stay safely under 1GB)
	TargetSizeBytes = 900 * 1024 * 1024
	// AudioBitrate is the audio bitrate in bits per second
	AudioBitrate = 128000
	// MinVideoBitrate is the minimum video bitrate (500 kbps) below which we downscale
	MinVideoBitrate = 500000
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
		if n, _ := fmt.Sscanf(probeResult.Format.Duration, "%f", &duration); n != 1 {
			return nil, fmt.Errorf("failed to parse duration: %s", probeResult.Format.Duration)
		}
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

// CompressionParams holds the parameters for video compression
type CompressionParams struct {
	VideoBitrate int  // bits per second
	Use1080p     bool // whether to downscale to 1080p
}

// CalculateCompressionParams determines the optimal bitrate and resolution to hit target size
// Prefers original resolution (4K) and only downscales if bitrate would be too low for quality
func CalculateCompressionParams(durationSec float64, targetSizeBytes int64) CompressionParams {
	// Calculate total bitrate needed for target size
	// bitrate = (size_in_bits) / duration_seconds
	totalBitrate := int((float64(targetSizeBytes) * 8) / durationSec)

	// Subtract audio bitrate to get video bitrate
	videoBitrate := totalBitrate - AudioBitrate

	// If video bitrate is too low for acceptable 4K quality, downscale to 1080p
	// 1080p looks good at lower bitrates than 4K
	if videoBitrate < MinVideoBitrate {
		// Even for 1080p, use whatever bitrate we can get
		return CompressionParams{
			VideoBitrate: videoBitrate,
			Use1080p:     true,
		}
	}

	// For 4K, we want at least ~2 Mbps for decent quality
	// If bitrate is below 2 Mbps, 1080p will look better than low-bitrate 4K
	const min4KBitrate = 2000000 // 2 Mbps
	if videoBitrate < min4KBitrate {
		return CompressionParams{
			VideoBitrate: videoBitrate,
			Use1080p:     true,
		}
	}

	return CompressionParams{
		VideoBitrate: videoBitrate,
		Use1080p:     false,
	}
}

// CompressForDubbing compresses video to fit under 1GB limit while maximizing quality
// Returns the path to use (original if no compression needed, or compressed file path)
func CompressForDubbing(ctx context.Context, inputPath string) (string, error) {
	return CompressForDubbingWithExecutor(ctx, inputPath, defaultExecutor)
}

// CompressForDubbingWithExecutor compresses video using a custom executor (for testing)
// Uses two-pass bitrate encoding to maximize quality while hitting target size
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

	// Calculate compression parameters (bitrate-based for precise size control)
	params := CalculateCompressionParams(info.Duration, TargetSizeBytes)

	// Build output path
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)
	outputPath := filepath.Join(filepath.Dir(inputPath), baseName+"_compressed.mp4")

	// Create temp directory for two-pass log files
	tempDir := filepath.Dir(inputPath)
	passLogFile := filepath.Join(tempDir, "ffmpeg2pass")

	// Two-pass encoding for best quality at target bitrate
	// Pass 1: Analyze video (output to null)
	pass1Args := []string{
		"-y",
		"-i", inputPath,
		"-c:v", "libx264",
		"-b:v", fmt.Sprintf("%d", params.VideoBitrate),
		"-preset", "medium",
		"-pass", "1",
		"-passlogfile", passLogFile,
		"-an", // No audio in first pass
	}

	if params.Use1080p {
		pass1Args = append(pass1Args, "-vf", "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2")
	}

	// Output to null device for first pass
	pass1Args = append(pass1Args, "-f", "null", "/dev/null")

	_, stderr, err := executor.ExecuteCommandWithStderr(ctx, "ffmpeg", pass1Args...)
	if err != nil {
		if isCommandNotFound(err) {
			return "", ErrFFmpegNotFound
		}
		return "", fmt.Errorf("%w (pass 1): %s", ErrCompressionFailed, string(stderr))
	}

	// Pass 2: Actual encoding with target bitrate
	pass2Args := []string{
		"-y",
		"-i", inputPath,
		"-c:v", "libx264",
		"-b:v", fmt.Sprintf("%d", params.VideoBitrate),
		"-preset", "medium",
		"-pass", "2",
		"-passlogfile", passLogFile,
		"-c:a", "aac",
		"-b:a", "128k",
	}

	if params.Use1080p {
		pass2Args = append(pass2Args, "-vf", "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2")
	}

	pass2Args = append(pass2Args, outputPath)

	_, stderr, err = executor.ExecuteCommandWithStderr(ctx, "ffmpeg", pass2Args...)
	// Clean up pass log files regardless of outcome
	os.Remove(passLogFile + "-0.log")
	os.Remove(passLogFile + "-0.log.mbtree")

	if err != nil {
		return "", fmt.Errorf("%w (pass 2): %s", ErrCompressionFailed, string(stderr))
	}

	// Verify output file exists
	_, err = os.Stat(outputPath)
	if err != nil {
		return "", fmt.Errorf("compressed file not created: %w", err)
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
