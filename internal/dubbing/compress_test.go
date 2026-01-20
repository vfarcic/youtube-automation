package dubbing

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

func TestCalculateOptimalCRF(t *testing.T) {
	tests := []struct {
		name          string
		durationSec   float64
		currentSizeMB float64
		targetSizeMB  int
		wantCRF       int
		wantUse1080p  bool
	}{
		{
			name:          "small file no compression needed",
			durationSec:   600, // 10 min
			currentSizeMB: 500,
			targetSizeMB:  900,
			wantCRF:       MinCRF,
			wantUse1080p:  false,
		},
		{
			name:          "moderate compression 2x ratio",
			durationSec:   600, // 10 min
			currentSizeMB: 1800,
			targetSizeMB:  900,
			wantCRF:       29, // 23 + 6 (log2(2) * 6)
			wantUse1080p:  false,
		},
		{
			name:          "high compression 4x ratio",
			durationSec:   600, // 10 min
			currentSizeMB: 3600,
			targetSizeMB:  900,
			wantCRF:       DefaultCRF1080p,
			wantUse1080p:  true, // CRF would be 35, exceeds max
		},
		{
			name:          "long video always 1080p",
			durationSec:   1800, // 30 min (> 25 min threshold)
			currentSizeMB: 1000,
			targetSizeMB:  900,
			wantCRF:       DefaultCRF1080p,
			wantUse1080p:  true,
		},
		{
			name:          "exactly at 25 min threshold",
			durationSec:   1500, // exactly 25 min
			currentSizeMB: 1800,
			targetSizeMB:  900,
			wantCRF:       29,
			wantUse1080p:  false,
		},
		{
			name:          "just over 25 min threshold",
			durationSec:   1501, // just over 25 min
			currentSizeMB: 1800,
			targetSizeMB:  900,
			wantCRF:       DefaultCRF1080p,
			wantUse1080p:  true,
		},
		{
			name:          "very large file needs 1080p",
			durationSec:   900, // 15 min
			currentSizeMB: 17000,
			targetSizeMB:  900,
			wantCRF:       DefaultCRF1080p,
			wantUse1080p:  true,
		},
		{
			name:          "borderline compression 1.5x ratio",
			durationSec:   600,
			currentSizeMB: 1350,
			targetSizeMB:  900,
			wantCRF:       26, // 23 + ~3 (log2(1.5) * 6)
			wantUse1080p:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCRF, gotUse1080p := CalculateOptimalCRF(tt.durationSec, tt.currentSizeMB, tt.targetSizeMB)

			if gotUse1080p != tt.wantUse1080p {
				t.Errorf("use1080p = %v, want %v", gotUse1080p, tt.wantUse1080p)
			}

			if gotCRF != tt.wantCRF {
				t.Errorf("crf = %d, want %d", gotCRF, tt.wantCRF)
			}
		})
	}
}

func TestLog2(t *testing.T) {
	tests := []struct {
		input float64
		want  float64
	}{
		{1, 0},
		{2, 1},
		{4, 2},
		{8, 3},
		{0, 0},
		{-1, 0},
	}

	for _, tt := range tests {
		got := log2(tt.input)
		if got != tt.want {
			t.Errorf("log2(%f) = %f, want %f", tt.input, got, tt.want)
		}
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
		name               string
		fileSize           int64
		ffprobeOutput      string
		ffmpegErr          error
		wantCompressed     bool
		wantErr            bool
		wantErrType        error
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

			mock := &MockCommandExecutor{
				ExecuteCommandFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
					// ffprobe call
					return []byte(tt.ffprobeOutput), nil
				},
				ExecuteCommandWithStderrFunc: func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
					// ffmpeg call
					if tt.ffmpegErr != nil {
						return nil, []byte("error output"), tt.ffmpegErr
					}
					// Create the compressed output file if requested
					if tt.createCompressedFile {
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

func TestCompressForDubbingWithExecutor_LongVideo(t *testing.T) {
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "long_video.mp4")

	// Create a 2GB file
	f, _ := os.Create(inputPath)
	f.Truncate(2 * 1024 * 1024 * 1024)
	f.Close()

	var ffmpegArgs []string
	mock := &MockCommandExecutor{
		ExecuteCommandFunc: func(ctx context.Context, name string, args ...string) ([]byte, error) {
			// Return a 30 min video (over 25 min threshold)
			return []byte(`{
				"format": {"duration": "1800", "size": "2147483648"},
				"streams": [{"width": 3840, "height": 2160}]
			}`), nil
		},
		ExecuteCommandWithStderrFunc: func(ctx context.Context, name string, args ...string) ([]byte, []byte, error) {
			ffmpegArgs = args
			// Create compressed file
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
		t.Error("expected 1080p scaling for long video")
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
	if TargetSizeMB != 900 {
		t.Errorf("TargetSizeMB = %d, want 900", TargetSizeMB)
	}
	if MaxDurationFor4K != 25*60 {
		t.Errorf("MaxDurationFor4K = %d, want 1500 (25 min)", MaxDurationFor4K)
	}
	if DefaultCRF1080p != 26 {
		t.Errorf("DefaultCRF1080p = %d, want 26", DefaultCRF1080p)
	}
	if MinCRF != 23 {
		t.Errorf("MinCRF = %d, want 23", MinCRF)
	}
	if MaxCRF != 30 {
		t.Errorf("MaxCRF = %d, want 30", MaxCRF)
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
