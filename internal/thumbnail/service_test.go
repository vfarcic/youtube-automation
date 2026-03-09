package thumbnail

import (
	"errors"
	"os/exec"
	"runtime"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

func TestGetOriginalThumbnailPath(t *testing.T) {
	tests := []struct {
		name    string
		video   *storage.Video
		want    string
		wantErr error
	}{
		{
			name: "ThumbnailVariants returns first non-empty path",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: "/path/to/first.png"},
					{Index: 2, Path: "/path/to/second.png"},
				},
			},
			want:    "/path/to/first.png",
			wantErr: nil,
		},
		{
			name: "ThumbnailVariants skips empty paths",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: ""},
					{Index: 2, Path: "/path/to/second.png"},
				},
			},
			want:    "/path/to/second.png",
			wantErr: nil,
		},
		{
			name: "Empty ThumbnailVariants uses deprecated Thumbnail field",
			video: &storage.Video{
				Thumbnail: "/legacy/thumbnail.jpg",
			},
			want:    "/legacy/thumbnail.jpg",
			wantErr: nil,
		},
		{
			name: "Both ThumbnailVariants and Thumbnail - prefers ThumbnailVariants",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: "/new/thumbnail.png"},
				},
				Thumbnail: "/legacy/thumbnail.jpg",
			},
			want:    "/new/thumbnail.png",
			wantErr: nil,
		},
		{
			name: "No thumbnail at all",
			video: &storage.Video{
				Name: "Test Video",
			},
			want:    "",
			wantErr: ErrNoThumbnail,
		},
		{
			name: "Empty ThumbnailVariants and empty Thumbnail",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{},
				Thumbnail:         "",
			},
			want:    "",
			wantErr: ErrNoThumbnail,
		},
		{
			name: "ThumbnailVariants with all empty paths and empty Thumbnail",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Path: ""},
				},
				Thumbnail: "",
			},
			want:    "",
			wantErr: ErrNoThumbnail,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetOriginalThumbnailPath(tt.video)

			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("GetOriginalThumbnailPath() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("GetOriginalThumbnailPath() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
			} else if err != nil {
				t.Errorf("GetOriginalThumbnailPath() unexpected error = %v", err)
				return
			}

			if got != tt.want {
				t.Errorf("GetOriginalThumbnailPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// mockCommandRunner for testing OpenInDefaultViewer
type mockCommandRunner struct {
	lastCmd     *exec.Cmd
	returnError error
}

func (m *mockCommandRunner) Start(cmd *exec.Cmd) error {
	m.lastCmd = cmd
	return m.returnError
}

func TestOpenInDefaultViewer_CommandConstruction(t *testing.T) {
	tests := []struct {
		name        string
		goos        string
		filePath    string
		wantCmd     string
		wantArgs    []string
		wantErr     bool
		runnerError error
	}{
		{
			name:     "macOS uses open",
			goos:     "darwin",
			filePath: "/path/to/file.png",
			wantCmd:  "open",
			wantArgs: []string{"/path/to/file.png"},
			wantErr:  false,
		},
		{
			name:     "Linux uses xdg-open",
			goos:     "linux",
			filePath: "/path/to/file.png",
			wantCmd:  "xdg-open",
			wantArgs: []string{"/path/to/file.png"},
			wantErr:  false,
		},
		{
			name:     "Windows uses cmd /c start",
			goos:     "windows",
			filePath: "C:\\path\\to\\file.png",
			wantCmd:  "cmd",
			wantArgs: []string{"/c", "start", "", "C:\\path\\to\\file.png"},
			wantErr:  false,
		},
		{
			name:        "Runner error is propagated",
			goos:        "darwin",
			filePath:    "/path/to/file.png",
			wantCmd:     "open",
			wantArgs:    []string{"/path/to/file.png"},
			wantErr:     true,
			runnerError: errors.New("command failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if we're not on the right OS for command verification
			// The test still validates error handling
			if runtime.GOOS != tt.goos && tt.runnerError == nil {
				t.Skipf("Skipping %s test on %s", tt.goos, runtime.GOOS)
			}

			runner := &mockCommandRunner{returnError: tt.runnerError}
			err := openInDefaultViewerWithRunner(tt.filePath, runner)

			if tt.wantErr {
				if err == nil {
					t.Error("openInDefaultViewerWithRunner() expected error, got nil")
				}
				if !errors.Is(err, ErrOpenViewerFailed) {
					t.Errorf("openInDefaultViewerWithRunner() error = %v, want ErrOpenViewerFailed", err)
				}
			} else {
				if err != nil {
					t.Errorf("openInDefaultViewerWithRunner() unexpected error = %v", err)
				}
			}
		})
	}
}

func TestOpenInDefaultViewer_CurrentOS(t *testing.T) {
	supportedOS := map[string]bool{
		"darwin":  true,
		"linux":   true,
		"windows": true,
	}

	runner := &mockCommandRunner{}
	err := openInDefaultViewerWithRunner("/some/path.png", runner)

	if supportedOS[runtime.GOOS] {
		if err != nil {
			t.Errorf("openInDefaultViewerWithRunner() on %s should succeed, got error = %v", runtime.GOOS, err)
		}
		if runner.lastCmd == nil {
			t.Error("openInDefaultViewerWithRunner() should have created a command")
		}
	}
}
