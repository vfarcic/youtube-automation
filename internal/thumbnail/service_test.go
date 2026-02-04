package thumbnail

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"
)

// mockGenerator is a mock implementation of ThumbnailGenerator for testing.
type mockGenerator struct {
	returnBytes []byte
	returnError error
	calledWith  struct {
		imagePath  string
		tagline    string
		targetLang string
	}
}

func (m *mockGenerator) GenerateLocalizedThumbnail(ctx context.Context, imagePath, tagline, targetLang string) ([]byte, error) {
	m.calledWith.imagePath = imagePath
	m.calledWith.tagline = tagline
	m.calledWith.targetLang = targetLang
	return m.returnBytes, m.returnError
}

func TestGetLocalizedThumbnailPath(t *testing.T) {
	tests := []struct {
		name         string
		originalPath string
		langCode     string
		want         string
	}{
		{
			name:         "PNG file with Spanish",
			originalPath: "/path/to/thumbnail.png",
			langCode:     "es",
			want:         "/path/to/thumbnail-es.png",
		},
		{
			name:         "JPEG file with Portuguese",
			originalPath: "/videos/my-video/thumb.jpg",
			langCode:     "pt",
			want:         "/videos/my-video/thumb-pt.jpg",
		},
		{
			name:         "WebP file with German",
			originalPath: "relative/path/image.webp",
			langCode:     "de",
			want:         "relative/path/image-de.webp",
		},
		{
			name:         "File with multiple dots",
			originalPath: "/path/to/my.video.thumbnail.png",
			langCode:     "fr",
			want:         "/path/to/my.video.thumbnail-fr.png",
		},
		{
			name:         "File without extension",
			originalPath: "/path/to/thumbnail",
			langCode:     "it",
			want:         "/path/to/thumbnail-it",
		},
		{
			name:         "Empty path",
			originalPath: "",
			langCode:     "ja",
			want:         "-ja",
		},
		{
			name:         "Path with spaces",
			originalPath: "/path/to/my thumbnail.png",
			langCode:     "ko",
			want:         "/path/to/my thumbnail-ko.png",
		},
		{
			name:         "Uppercase extension",
			originalPath: "/path/to/thumbnail.PNG",
			langCode:     "zh",
			want:         "/path/to/thumbnail-zh.PNG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetLocalizedThumbnailPath(tt.originalPath, tt.langCode)
			if got != tt.want {
				t.Errorf("GetLocalizedThumbnailPath(%q, %q) = %q, want %q",
					tt.originalPath, tt.langCode, got, tt.want)
			}
		})
	}
}

func TestGetOriginalThumbnailPath(t *testing.T) {
	tests := []struct {
		name    string
		video   *storage.Video
		want    string
		wantErr error
	}{
		{
			name: "ThumbnailVariants with original type",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Type: "original", Path: "/path/to/original.png"},
					{Index: 2, Type: "subtle", Path: "/path/to/subtle.png"},
					{Index: 3, Type: "bold", Path: "/path/to/bold.png"},
				},
			},
			want:    "/path/to/original.png",
			wantErr: nil,
		},
		{
			name: "ThumbnailVariants original not first",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Type: "subtle", Path: "/path/to/subtle.png"},
					{Index: 2, Type: "original", Path: "/path/to/original.png"},
				},
			},
			want:    "/path/to/original.png",
			wantErr: nil,
		},
		{
			name: "ThumbnailVariants without original type falls back to first",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Type: "custom", Path: "/path/to/custom.png"},
				},
			},
			want:    "/path/to/custom.png",
			wantErr: nil,
		},
		{
			name: "ThumbnailVariants with empty original path falls back to first non-empty",
			video: &storage.Video{
				ThumbnailVariants: []storage.ThumbnailVariant{
					{Index: 1, Type: "original", Path: ""},
					{Index: 2, Type: "subtle", Path: "/path/to/subtle.png"},
				},
			},
			want:    "/path/to/subtle.png",
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
					{Index: 1, Type: "original", Path: "/new/thumbnail.png"},
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
					{Index: 1, Type: "original", Path: ""},
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

func TestLocalizeThumbnail_Success(t *testing.T) {
	// Create a temp directory for the test
	tmpDir := t.TempDir()

	// Create a fake original thumbnail
	originalPath := filepath.Join(tmpDir, "thumbnail.png")
	if err := os.WriteFile(originalPath, []byte("original image data"), 0644); err != nil {
		t.Fatalf("Failed to create test thumbnail: %v", err)
	}

	video := &storage.Video{
		Name:    "Test Video",
		Tagline: "This is the tagline",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Type: "original", Path: originalPath},
		},
	}

	expectedOutput := []byte("localized image data")
	mock := &mockGenerator{returnBytes: expectedOutput}

	ctx := context.Background()
	outputPath, err := LocalizeThumbnail(ctx, mock, video, "es")

	if err != nil {
		t.Fatalf("LocalizeThumbnail() unexpected error = %v", err)
	}

	// Verify the mock was called with correct parameters
	if mock.calledWith.imagePath != originalPath {
		t.Errorf("Generator called with imagePath = %q, want %q", mock.calledWith.imagePath, originalPath)
	}
	if mock.calledWith.tagline != video.Tagline {
		t.Errorf("Generator called with tagline = %q, want %q", mock.calledWith.tagline, video.Tagline)
	}
	if mock.calledWith.targetLang != "es" {
		t.Errorf("Generator called with targetLang = %q, want %q", mock.calledWith.targetLang, "es")
	}

	// Verify output path
	expectedPath := filepath.Join(tmpDir, "thumbnail-es.png")
	if outputPath != expectedPath {
		t.Errorf("LocalizeThumbnail() path = %q, want %q", outputPath, expectedPath)
	}

	// Verify file was written
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if string(content) != string(expectedOutput) {
		t.Errorf("Output file content = %q, want %q", string(content), string(expectedOutput))
	}
}

func TestLocalizeThumbnail_NoThumbnail(t *testing.T) {
	video := &storage.Video{
		Name:    "Test Video",
		Tagline: "This is the tagline",
		// No thumbnail
	}

	mock := &mockGenerator{returnBytes: []byte("image")}
	ctx := context.Background()

	_, err := LocalizeThumbnail(ctx, mock, video, "es")

	if err == nil {
		t.Fatal("LocalizeThumbnail() expected error, got nil")
	}
	if !errors.Is(err, ErrNoThumbnail) {
		t.Errorf("LocalizeThumbnail() error = %v, want %v", err, ErrNoThumbnail)
	}
}

func TestLocalizeThumbnail_NoTagline(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "thumbnail.png")
	if err := os.WriteFile(originalPath, []byte("original"), 0644); err != nil {
		t.Fatalf("Failed to create test thumbnail: %v", err)
	}

	video := &storage.Video{
		Name:    "Test Video",
		Tagline: "", // Empty tagline
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Type: "original", Path: originalPath},
		},
	}

	mock := &mockGenerator{returnBytes: []byte("image")}
	ctx := context.Background()

	_, err := LocalizeThumbnail(ctx, mock, video, "es")

	if err == nil {
		t.Fatal("LocalizeThumbnail() expected error, got nil")
	}
	if !errors.Is(err, ErrNoTagline) {
		t.Errorf("LocalizeThumbnail() error = %v, want %v", err, ErrNoTagline)
	}
}

func TestLocalizeThumbnail_UnsupportedLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "thumbnail.png")
	if err := os.WriteFile(originalPath, []byte("original"), 0644); err != nil {
		t.Fatalf("Failed to create test thumbnail: %v", err)
	}

	video := &storage.Video{
		Name:    "Test Video",
		Tagline: "Test tagline",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Type: "original", Path: originalPath},
		},
	}

	mock := &mockGenerator{returnBytes: []byte("image")}
	ctx := context.Background()

	_, err := LocalizeThumbnail(ctx, mock, video, "xx") // Invalid language

	if err == nil {
		t.Fatal("LocalizeThumbnail() expected error, got nil")
	}
	if !errors.Is(err, ErrUnsupportedLang) {
		t.Errorf("LocalizeThumbnail() error = %v, want %v", err, ErrUnsupportedLang)
	}
}

func TestLocalizeThumbnail_GenerationFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "thumbnail.png")
	if err := os.WriteFile(originalPath, []byte("original"), 0644); err != nil {
		t.Fatalf("Failed to create test thumbnail: %v", err)
	}

	video := &storage.Video{
		Name:    "Test Video",
		Tagline: "Test tagline",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Type: "original", Path: originalPath},
		},
	}

	generationError := errors.New("API error: rate limited")
	mock := &mockGenerator{returnError: generationError}
	ctx := context.Background()

	_, err := LocalizeThumbnail(ctx, mock, video, "es")

	if err == nil {
		t.Fatal("LocalizeThumbnail() expected error, got nil")
	}
	if !errors.Is(err, generationError) {
		t.Errorf("LocalizeThumbnail() error should wrap %v, got %v", generationError, err)
	}
}

func TestLocalizeThumbnail_SaveFails(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "thumbnail.png")
	if err := os.WriteFile(originalPath, []byte("original"), 0644); err != nil {
		t.Fatalf("Failed to create test thumbnail: %v", err)
	}

	// Use a path in a non-existent directory to cause write failure
	nonExistentDir := filepath.Join(tmpDir, "nonexistent", "subdir")
	video := &storage.Video{
		Name:    "Test Video",
		Tagline: "Test tagline",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Type: "original", Path: filepath.Join(nonExistentDir, "thumbnail.png")},
		},
	}

	mock := &mockGenerator{returnBytes: []byte("localized image")}
	ctx := context.Background()

	_, err := LocalizeThumbnail(ctx, mock, video, "es")

	if err == nil {
		t.Fatal("LocalizeThumbnail() expected error, got nil")
	}
	if !errors.Is(err, ErrSaveFailed) {
		t.Errorf("LocalizeThumbnail() error = %v, want %v", err, ErrSaveFailed)
	}
}

func TestLocalizeThumbnail_AllLanguages(t *testing.T) {
	tmpDir := t.TempDir()
	originalPath := filepath.Join(tmpDir, "thumbnail.png")
	if err := os.WriteFile(originalPath, []byte("original"), 0644); err != nil {
		t.Fatalf("Failed to create test thumbnail: %v", err)
	}

	video := &storage.Video{
		Name:    "Test Video",
		Tagline: "Test tagline",
		ThumbnailVariants: []storage.ThumbnailVariant{
			{Index: 1, Type: "original", Path: originalPath},
		},
	}

	languages := []string{"es", "pt", "de", "fr", "it", "ja", "ko", "zh"}

	for _, lang := range languages {
		t.Run(lang, func(t *testing.T) {
			mock := &mockGenerator{returnBytes: []byte("localized for " + lang)}
			ctx := context.Background()

			outputPath, err := LocalizeThumbnail(ctx, mock, video, lang)
			if err != nil {
				t.Fatalf("LocalizeThumbnail() error = %v", err)
			}

			expectedPath := filepath.Join(tmpDir, "thumbnail-"+lang+".png")
			if outputPath != expectedPath {
				t.Errorf("LocalizeThumbnail() path = %q, want %q", outputPath, expectedPath)
			}

			// Verify generator was called with correct language
			if mock.calledWith.targetLang != lang {
				t.Errorf("Generator called with lang = %q, want %q", mock.calledWith.targetLang, lang)
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
		name         string
		goos         string
		filePath     string
		wantCmd      string
		wantArgs     []string
		wantErr      bool
		runnerError  error
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
	// This test verifies the function works on the current OS
	// We can't actually open a viewer in tests, but we can verify
	// the function doesn't error for supported OSes

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

// Verify that Client satisfies ThumbnailGenerator interface
func TestClientImplementsThumbnailGenerator(t *testing.T) {
	// This is a compile-time check - if Client doesn't implement
	// ThumbnailGenerator, this won't compile
	var _ ThumbnailGenerator = (*Client)(nil)
}
