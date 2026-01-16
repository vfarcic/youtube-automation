package publishing

import (
	"os"
	"strings"
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"
	"devopstoolkit/youtube-automation/internal/storage"
)

func TestUploadDubbedVideo_Validation(t *testing.T) {
	// Create a temp file for testing file existence
	tempFile, err := os.CreateTemp("", "test-dubbed-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	tests := []struct {
		name    string
		video   *storage.Video
		lang    string
		wantErr string
	}{
		{
			name:    "unsupported language code",
			video:   &storage.Video{},
			lang:    "pt",
			wantErr: "unsupported language code: pt",
		},
		{
			name: "missing dubbing info for language",
			video: &storage.Video{
				Dubbing: map[string]storage.DubbingInfo{},
			},
			lang:    "es",
			wantErr: "no dubbing info found for language: es",
		},
		{
			name: "nil dubbing map",
			video: &storage.Video{
				Dubbing: nil,
			},
			lang:    "es",
			wantErr: "no dubbing info found for language: es",
		},
		{
			name: "empty dubbed video path",
			video: &storage.Video{
				Dubbing: map[string]storage.DubbingInfo{
					"es": {
						DubbedVideoPath: "",
						Title:           "Test Title",
					},
				},
			},
			lang:    "es",
			wantErr: "dubbed video path is empty",
		},
		{
			name: "empty translated title",
			video: &storage.Video{
				Dubbing: map[string]storage.DubbingInfo{
					"es": {
						DubbedVideoPath: "/some/path.mp4",
						Title:           "",
					},
				},
			},
			lang:    "es",
			wantErr: "translated title is empty",
		},
		{
			name: "dubbed video file does not exist",
			video: &storage.Video{
				Dubbing: map[string]storage.DubbingInfo{
					"es": {
						DubbedVideoPath: "/nonexistent/path/video.mp4",
						Title:           "Test Title",
					},
				},
			},
			lang:    "es",
			wantErr: "dubbed video file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UploadDubbedVideo(tt.video, tt.lang)
			if err == nil {
				t.Errorf("UploadDubbedVideo() expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("UploadDubbedVideo() error = %q, want error containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestUploadDubbedVideo_EmptyChannelID(t *testing.T) {
	// Create a temp file for testing
	tempFile, err := os.CreateTemp("", "test-dubbed-video-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Save original settings
	originalSettings := configuration.GlobalSettings.SpanishChannel
	defer func() {
		configuration.GlobalSettings.SpanishChannel = originalSettings
	}()

	// Set empty channel ID
	configuration.GlobalSettings.SpanishChannel = configuration.SettingsSpanishChannel{
		ChannelID: "",
	}

	video := &storage.Video{
		Dubbing: map[string]storage.DubbingInfo{
			"es": {
				DubbedVideoPath: tempFile.Name(),
				Title:           "Test Title",
			},
		},
	}

	_, err = UploadDubbedVideo(video, "es")
	if err == nil {
		t.Error("UploadDubbedVideo() expected error for empty channel ID, got nil")
		return
	}
	if !strings.Contains(err.Error(), "Spanish channel ID is not configured") {
		t.Errorf("UploadDubbedVideo() error = %q, want error about channel ID not configured", err.Error())
	}
}

func TestBuildDubbedDescription(t *testing.T) {
	tests := []struct {
		name            string
		dubbingInfo     storage.DubbingInfo
		originalVideoID string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name: "full description with all fields",
			dubbingInfo: storage.DubbingInfo{
				Description: "Esta es la descripcion del video.",
				Timecodes:   "0:00 Intro\n1:00 Contenido",
			},
			originalVideoID: "abc123",
			wantContains: []string{
				"Esta es la descripcion del video.",
				"Timecodes",
				"0:00 Intro",
				"1:00 Contenido",
				"Original Video",
				"https://youtu.be/abc123",
			},
		},
		{
			name: "description only - no timecodes",
			dubbingInfo: storage.DubbingInfo{
				Description: "Solo descripcion.",
				Timecodes:   "",
			},
			originalVideoID: "xyz789",
			wantContains: []string{
				"Solo descripcion.",
				"https://youtu.be/xyz789",
			},
			wantNotContains: []string{
				"Timecodes",
			},
		},
		{
			name: "N/A timecodes should be excluded",
			dubbingInfo: storage.DubbingInfo{
				Description: "Descripcion con timecodes N/A.",
				Timecodes:   "N/A",
			},
			originalVideoID: "def456",
			wantContains: []string{
				"Descripcion con timecodes N/A.",
				"https://youtu.be/def456",
			},
			wantNotContains: []string{
				"Timecodes",
			},
		},
		{
			name: "no original video ID",
			dubbingInfo: storage.DubbingInfo{
				Description: "Video sin original.",
				Timecodes:   "0:00 Start",
			},
			originalVideoID: "",
			wantContains: []string{
				"Video sin original.",
				"0:00 Start",
			},
			wantNotContains: []string{
				"Original Video",
				"youtu.be",
			},
		},
		{
			name: "empty description",
			dubbingInfo: storage.DubbingInfo{
				Description: "",
				Timecodes:   "0:00 Solo timecodes",
			},
			originalVideoID: "video123",
			wantContains: []string{
				"0:00 Solo timecodes",
				"https://youtu.be/video123",
			},
		},
		{
			name:            "all empty",
			dubbingInfo:     storage.DubbingInfo{},
			originalVideoID: "",
			wantContains:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDubbedDescription(tt.dubbingInfo, tt.originalVideoID)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("buildDubbedDescription() result missing expected string %q\nGot: %s", want, result)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("buildDubbedDescription() result should not contain %q\nGot: %s", notWant, result)
				}
			}
		})
	}
}

func TestBuildDubbedDescription_Formatting(t *testing.T) {
	dubbingInfo := storage.DubbingInfo{
		Description: "Test description",
		Timecodes:   "0:00 Intro",
	}

	result := buildDubbedDescription(dubbingInfo, "abc123")

	// Check that sections are separated by double newlines
	if !strings.Contains(result, "\n\n") {
		t.Error("buildDubbedDescription() should separate sections with double newlines")
	}

	// Check timecode section has proper header
	if !strings.Contains(result, "Timecodes") {
		t.Error("buildDubbedDescription() should include Timecodes header")
	}

	// Check original video section has proper header
	if !strings.Contains(result, "Original Video") {
		t.Error("buildDubbedDescription() should include Original Video header")
	}
}
