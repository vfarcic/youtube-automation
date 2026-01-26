package dubbing

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockCommandExecutor is a mock implementation of CommandExecutor for testing
type MockCommandExecutor struct {
	ExecuteCommandFunc           func(ctx context.Context, name string, args ...string) ([]byte, error)
	ExecuteCommandWithStderrFunc func(ctx context.Context, name string, args ...string) ([]byte, []byte, error)
}

func (m *MockCommandExecutor) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.ExecuteCommandFunc != nil {
		return m.ExecuteCommandFunc(ctx, name, args...)
	}
	return nil, nil
}

func (m *MockCommandExecutor) ExecuteCommandWithStderr(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
	if m.ExecuteCommandWithStderrFunc != nil {
		return m.ExecuteCommandWithStderrFunc(ctx, name, args...)
	}
	return nil, nil, nil
}

func TestCalculateCompressionParams(t *testing.T) {
	tests := []struct {
		name             string
		durationSec      float64
		targetSizeBytes  int64
		wantUse1080p     bool
		wantMinBitrate   int // minimum expected bitrate
		wantMaxBitrate   int // maximum expected bitrate
	}{
		{
			name:            "short video high bitrate keeps 4K",
			durationSec:     600, // 10 min
			targetSizeBytes: 900 * 1024 * 1024,
			wantUse1080p:    false,
			wantMinBitrate:  10000000, // ~10+ Mbps for 10 min video at 900MB
			wantMaxBitrate:  15000000,
		},
		{
			name:            "medium video moderate bitrate keeps 4K",
			durationSec:     1800, // 30 min
			targetSizeBytes: 900 * 1024 * 1024,
			wantUse1080p:    false,
			wantMinBitrate:  3500000, // ~4 Mbps for 30 min video at 900MB
			wantMaxBitrate:  5000000,
		},
		{
			name:            "long video low bitrate uses 1080p",
			durationSec:     7200, // 2 hours
			targetSizeBytes: 900 * 1024 * 1024,
			wantUse1080p:    true, // below 2 Mbps threshold
			wantMinBitrate:  800000,
			wantMaxBitrate:  1200000,
		},
		{
			name:            "very long video very low bitrate uses 1080p",
			durationSec:     14400, // 4 hours
			targetSizeBytes: 900 * 1024 * 1024,
			wantUse1080p:    true,
			wantMinBitrate:  300000,
			wantMaxBitrate:  600000,
		},
		{
			name:            "1 hour video borderline uses 1080p",
			durationSec:     3600, // 1 hour - bitrate ~2 Mbps, borderline
			targetSizeBytes: 900 * 1024 * 1024,
			wantUse1080p:    true, // just under 2 Mbps threshold
			wantMinBitrate:  1800000,
			wantMaxBitrate:  2100000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := CalculateCompressionParams(tt.durationSec, tt.targetSizeBytes)

			if params.Use1080p != tt.wantUse1080p {
				t.Errorf("Use1080p = %v, want %v", params.Use1080p, tt.wantUse1080p)
			}

			if params.VideoBitrate < tt.wantMinBitrate || params.VideoBitrate > tt.wantMaxBitrate {
				t.Errorf("VideoBitrate = %d, want between %d and %d", params.VideoBitrate, tt.wantMinBitrate, tt.wantMaxBitrate)
			}
		})
	}
}

func TestCalculateCompressionParams_BitrateCalculation(t *testing.T) {
	// Test the exact bitrate calculation
	// For 900MB target and 1800 seconds (30 min):
	// Total bitrate = (900 * 1024 * 1024 * 8) / 1800 = 4,194,304 bps
	// Video bitrate = 4,194,304 - 128,000 = 4,066,304 bps

	params := CalculateCompressionParams(1800, 900*1024*1024)

	expectedTotalBitrate := int((float64(900*1024*1024) * 8) / 1800)
	expectedVideoBitrate := expectedTotalBitrate - AudioBitrate

	if params.VideoBitrate != expectedVideoBitrate {
		t.Errorf("VideoBitrate = %d, want %d", params.VideoBitrate, expectedVideoBitrate)
	}
}

func TestGetVideoInfoWithExecutor(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   func(t *testing.T) string
		mockOutput  string
		mockErr     error
		wantErr     bool
		wantErrType error
		wantInfo    *VideoInfo
	}{
		{
			name: "success",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.mp4")
				if err := os.WriteFile(path, make([]byte, 1024), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			mockOutput: `{
				"format": {"duration": "120.5", "size": "1024"},
				"streams": [{"width": 1920, "height": 1080}]
			}`,
			wantErr: false,
			wantInfo: &VideoInfo{
				Duration: 120.5,
				Size:     1024,
				Width:    1920,
				Height:   1080,
			},
		},
		{
			name: "file not found",
			setupFile: func(t *testing.T) string {
				return "/nonexistent/path/video.mp4"
			},
			wantErr:     true,
			wantErrType: ErrFileNotFound,
		},
		{
			name: "ffprobe not found",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.mp4")
				os.WriteFile(path, make([]byte, 1024), 0644)
				return path
			},
			mockErr:     errors.New("executable file not found"),
			wantErr:     true,
			wantErrType: ErrFFprobeNotFound,
		},
		{
			name: "invalid json output",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.mp4")
				os.WriteFile(path, make([]byte, 1024), 0644)
				return path
			},
			mockOutput: `{invalid json}`,
			wantErr:    true,
		},
		{
			name: "no video streams",
			setupFile: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.mp4")
				os.WriteFile(path, make([]byte, 2048), 0644)
				return path
			},
			mockOutput: `{
				"format": {"duration": "60.0", "size": "2048"},
				"streams": []
			}`,
			wantErr: false,
			wantInfo: &VideoInfo{
				Duration: 60.0,
				Size:     2048,
				Width:    0,
				Height:   0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := tt.setupFile(t)

			mock := &MockCommandExecutor{
				ExecuteCommandFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
					if tt.mockErr != nil {
						return nil, tt.mockErr
					}
					return []byte(tt.mockOutput), nil
				},
			}

			info, err := GetVideoInfoWithExecutor(context.Background(), filePath, mock)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
					t.Errorf("expected error type %v, got %v", tt.wantErrType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if info.Duration != tt.wantInfo.Duration {
				t.Errorf("Duration = %f, want %f", info.Duration, tt.wantInfo.Duration)
			}
			if info.Width != tt.wantInfo.Width {
				t.Errorf("Width = %d, want %d", info.Width, tt.wantInfo.Width)
			}
			if info.Height != tt.wantInfo.Height {
				t.Errorf("Height = %d, want %d", info.Height, tt.wantInfo.Height)
			}
			if info.FilePath != filePath {
				t.Errorf("FilePath = %s, want %s", info.FilePath, filePath)
			}
		})
	}
}

func TestCompressForDubbingWithExecutor(t *testing.T) {
	tests := []struct {
		name                 string
		fileSize             int64
		ffprobeOutput        string
		ffmpegErr            error
		wantCompressed       bool
		wantErr              bool
		wantErrType          error
		createCompressedFile bool
	}{
		{
			name:     "file under 1GB - no compression",
			fileSize: 500 * 1024 * 1024, // 500MB
			ffprobeOutput: `{
				"format": {"duration": "600", "size": "524288000"},
				"streams": [{"width": 3840, "height": 2160}]
			}`,
			wantCompressed: false,
			wantErr:        false,
		},
		{
			name:     "file over 1GB - needs compression",
			fileSize: 2 * 1024 * 1024 * 1024, // 2GB
			ffprobeOutput: `{
				"format": {"duration": "600", "size": "2147483648"},
				"streams": [{"width": 3840, "height": 2160}]
			}`,
			wantCompressed:       true,
			wantErr:              false,
			createCompressedFile: true,
		},
		{
			name:     "ffmpeg not found",
			fileSize: 2 * 1024 * 1024 * 1024,
			ffprobeOutput: `{
				"format": {"duration": "600", "size": "2147483648"},
				"streams": [{"width": 3840, "height": 2160}]
			}`,
			ffmpegErr:   errors.New("executable file not found"),
			wantErr:     true,
			wantErrType: ErrFFmpegNotFound,
		},
		{
			name:     "ffmpeg execution error",
			fileSize: 2 * 1024 * 1024 * 1024,
			ffprobeOutput: `{
				"format": {"duration": "600", "size": "2147483648"},
				"streams": [{"width": 3840, "height": 2160}]
			}`,
			ffmpegErr:   errors.New("encoding failed"),
			wantErr:     true,
			wantErrType: ErrCompressionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with specified size
			tmpDir := t.TempDir()
			inputPath := filepath.Join(tmpDir, "input.mp4")

			// Create a sparse file to avoid actually writing GB of data
			f, err := os.Create(inputPath)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}
			if err := f.Truncate(tt.fileSize); err != nil {
				f.Close()
				t.Fatalf("failed to set file size: %v", err)
			}
			f.Close()

			ffmpegCallCount := 0
			mock := &MockCommandExecutor{
				ExecuteCommandFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
					// ffprobe call
					return []byte(tt.ffprobeOutput), nil
				},
				ExecuteCommandWithStderrFunc: func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
					// ffmpeg call (two-pass, so called twice)
					ffmpegCallCount++
					if tt.ffmpegErr != nil {
						return nil, []byte("error output"), tt.ffmpegErr
					}
					// Create the compressed output file on second pass
					if tt.createCompressedFile && ffmpegCallCount == 2 {
						outputPath := filepath.Join(tmpDir, "input_compressed.mp4")
						// Create a file under 1GB
						f, _ := os.Create(outputPath)
						f.Truncate(800 * 1024 * 1024) // 800MB
						f.Close()
					}
					return nil, nil, nil
				},
			}

			outputPath, err := CompressForDubbingWithExecutor(context.Background(), inputPath, mock)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrType != nil && !errors.Is(err, tt.wantErrType) {
					t.Errorf("expected error type %v, got %v", tt.wantErrType, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantCompressed {
				expectedPath := filepath.Join(tmpDir, "input_compressed.mp4")
				if outputPath != expectedPath {
					t.Errorf("outputPath = %s, want %s", outputPath, expectedPath)
				}
			} else {
				if outputPath != inputPath {
					t.Errorf("outputPath = %s, want original %s", outputPath, inputPath)
				}
			}
		})
	}
}

func TestCompressForDubbingWithExecutor_TwoPassEncoding(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "video.mp4")

	// Create a 2GB file
	f, _ := os.Create(inputPath)
	f.Truncate(2 * 1024 * 1024 * 1024)
	f.Close()

	var ffmpegCalls [][]string
	mock := &MockCommandExecutor{
		ExecuteCommandFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			// Return a 10 min video (short enough for 4K)
			return []byte(`{
				"format": {"duration": "600", "size": "2147483648"},
				"streams": [{"width": 3840, "height": 2160}]
			}`), nil
		},
		ExecuteCommandWithStderrFunc: func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
			ffmpegCalls = append(ffmpegCalls, args)
			// Create compressed file on second call
			if len(ffmpegCalls) == 2 {
				outputPath := filepath.Join(tmpDir, "video_compressed.mp4")
				f, _ := os.Create(outputPath)
				f.Truncate(800 * 1024 * 1024)
				f.Close()
			}
			return nil, nil, nil
		},
	}

	_, err := CompressForDubbingWithExecutor(context.Background(), inputPath, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify two-pass encoding
	if len(ffmpegCalls) != 2 {
		t.Fatalf("expected 2 ffmpeg calls (two-pass), got %d", len(ffmpegCalls))
	}

	// Verify pass 1 uses -pass 1 and outputs to /dev/null
	pass1Found := false
	nullOutputFound := false
	for i, arg := range ffmpegCalls[0] {
		if arg == "-pass" && i+1 < len(ffmpegCalls[0]) && ffmpegCalls[0][i+1] == "1" {
			pass1Found = true
		}
		if arg == "/dev/null" {
			nullOutputFound = true
		}
	}
	if !pass1Found {
		t.Error("pass 1 did not include -pass 1")
	}
	if !nullOutputFound {
		t.Error("pass 1 did not output to /dev/null")
	}

	// Verify pass 2 uses -pass 2
	pass2Found := false
	for i, arg := range ffmpegCalls[1] {
		if arg == "-pass" && i+1 < len(ffmpegCalls[1]) && ffmpegCalls[1][i+1] == "2" {
			pass2Found = true
		}
	}
	if !pass2Found {
		t.Error("pass 2 did not include -pass 2")
	}
}

func TestCompressForDubbingWithExecutor_4KPreserved(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "short_4k.mp4")

	// Create a 2GB file
	f, _ := os.Create(inputPath)
	f.Truncate(2 * 1024 * 1024 * 1024)
	f.Close()

	var ffmpegArgs []string
	mock := &MockCommandExecutor{
		ExecuteCommandFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			// Return a 10 min 4K video - should preserve 4K at ~12 Mbps
			return []byte(`{
				"format": {"duration": "600", "size": "2147483648"},
				"streams": [{"width": 3840, "height": 2160}]
			}`), nil
		},
		ExecuteCommandWithStderrFunc: func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
			ffmpegArgs = args // Capture second pass args
			outputPath := filepath.Join(tmpDir, "short_4k_compressed.mp4")
			f, _ := os.Create(outputPath)
			f.Truncate(800 * 1024 * 1024)
			f.Close()
			return nil, nil, nil
		},
	}

	outputPath, err := CompressForDubbingWithExecutor(context.Background(), inputPath, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify 4K is preserved (no scale filter)
	for _, arg := range ffmpegArgs {
		if strings.Contains(arg, "scale=1920:1080") {
			t.Error("short 4K video should NOT be downscaled to 1080p")
		}
	}

	expectedPath := filepath.Join(tmpDir, "short_4k_compressed.mp4")
	if outputPath != expectedPath {
		t.Errorf("outputPath = %s, want %s", outputPath, expectedPath)
	}
}

func TestCompressForDubbingWithExecutor_LongVideoDownscaled(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "long_video.mp4")

	// Create a 2GB file
	f, _ := os.Create(inputPath)
	f.Truncate(2 * 1024 * 1024 * 1024)
	f.Close()

	var ffmpegArgs []string
	mock := &MockCommandExecutor{
		ExecuteCommandFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			// Return a 2 hour video - bitrate ~1 Mbps, should downscale
			return []byte(`{
				"format": {"duration": "7200", "size": "2147483648"},
				"streams": [{"width": 3840, "height": 2160}]
			}`), nil
		},
		ExecuteCommandWithStderrFunc: func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
			ffmpegArgs = args
			outputPath := filepath.Join(tmpDir, "long_video_compressed.mp4")
			f, _ := os.Create(outputPath)
			f.Truncate(600 * 1024 * 1024)
			f.Close()
			return nil, nil, nil
		},
	}

	outputPath, err := CompressForDubbingWithExecutor(context.Background(), inputPath, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify 1080p scaling was used for long video
	foundScale := false
	for i, arg := range ffmpegArgs {
		if arg == "-vf" && i+1 < len(ffmpegArgs) {
			if ffmpegArgs[i+1] == "scale=1920:1080:force_original_aspect_ratio=decrease,pad=1920:1080:(ow-iw)/2:(oh-ih)/2" {
				foundScale = true
			}
		}
	}
	if !foundScale {
		t.Error("expected 1080p scaling for long video with low bitrate")
	}

	expectedPath := filepath.Join(tmpDir, "long_video_compressed.mp4")
	if outputPath != expectedPath {
		t.Errorf("outputPath = %s, want %s", outputPath, expectedPath)
	}
}

func TestNeedsCompression(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		want     bool
	}{
		{
			name:     "file under 1GB",
			fileSize: 500 * 1024 * 1024,
			want:     false,
		},
		{
			name:     "file exactly 1GB",
			fileSize: 1024 * 1024 * 1024,
			want:     false,
		},
		{
			name:     "file over 1GB",
			fileSize: 1024*1024*1024 + 1,
			want:     true,
		},
		{
			name:     "large file",
			fileSize: 5 * 1024 * 1024 * 1024,
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.mp4")

			// Create sparse file
			f, _ := os.Create(path)
			f.Truncate(tt.fileSize)
			f.Close()

			got, err := NeedsCompression(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("NeedsCompression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNeedsCompression_FileNotFound(t *testing.T) {
	_, err := NeedsCompression("/nonexistent/file.mp4")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestIsCommandNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "executable file not found",
			err:  errors.New("executable file not found in $PATH"),
			want: true,
		},
		{
			name: "not found",
			err:  errors.New("command not found"),
			want: true,
		},
		{
			name: "no such file or directory",
			err:  errors.New("no such file or directory"),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("permission denied"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCommandNotFound(tt.err)
			if got != tt.want {
				t.Errorf("isCommandNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConstants(t *testing.T) {
	// Verify constants are set correctly
	if MaxFileSizeBytes != 1024*1024*1024 {
		t.Errorf("MaxFileSizeBytes = %d, want 1GB", MaxFileSizeBytes)
	}
	if TargetSizeBytes != 900*1024*1024 {
		t.Errorf("TargetSizeBytes = %d, want 900MB", TargetSizeBytes)
	}
	if AudioBitrate != 128000 {
		t.Errorf("AudioBitrate = %d, want 128000", AudioBitrate)
	}
	if MinVideoBitrate != 500000 {
		t.Errorf("MinVideoBitrate = %d, want 500000", MinVideoBitrate)
	}
}

func TestVideoInfoStruct(t *testing.T) {
	info := VideoInfo{
		Duration: 120.5,
		Size:     1024 * 1024 * 500,
		Width:    1920,
		Height:   1080,
		FilePath: "/path/to/video.mp4",
	}

	if info.Duration != 120.5 {
		t.Errorf("Duration = %f, want 120.5", info.Duration)
	}
	if info.Size != 1024*1024*500 {
		t.Errorf("Size = %d, want %d", info.Size, 1024*1024*500)
	}
	if info.Width != 1920 {
		t.Errorf("Width = %d, want 1920", info.Width)
	}
	if info.Height != 1080 {
		t.Errorf("Height = %d, want 1080", info.Height)
	}
	if info.FilePath != "/path/to/video.mp4" {
		t.Errorf("FilePath = %s, want /path/to/video.mp4", info.FilePath)
	}
}

func TestCompressionParamsStruct(t *testing.T) {
	params := CompressionParams{
		VideoBitrate: 5000000,
		Use1080p:     false,
	}

	if params.VideoBitrate != 5000000 {
		t.Errorf("VideoBitrate = %d, want 5000000", params.VideoBitrate)
	}
	if params.Use1080p != false {
		t.Errorf("Use1080p = %v, want false", params.Use1080p)
	}
}
