package publishing

import (
	"os"
	"strings"
	"testing"
	"time"

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

func TestParseVideoDate(t *testing.T) {
	tests := []struct {
		name    string
		dateStr string
		wantErr bool
	}{
		{
			name:    "standard format without seconds",
			dateStr: "2025-01-15T10:30",
			wantErr: false,
		},
		{
			name:    "format with seconds and Z",
			dateStr: "2025-01-15T10:30:00Z",
			wantErr: false,
		},
		{
			name:    "RFC3339 format",
			dateStr: "2025-01-15T10:30:00+00:00",
			wantErr: false,
		},
		{
			name:    "invalid format",
			dateStr: "15-01-2025",
			wantErr: true,
		},
		{
			name:    "empty string",
			dateStr: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseVideoDate(tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseVideoDate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetPublishStatus(t *testing.T) {
	// Create a future date (1 year from now)
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02T15:04")
	// Create a past date (1 year ago)
	pastDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02T15:04")

	tests := []struct {
		name              string
		scheduledDate     string
		wantPrivacy       string
		wantPublishAtSet  bool
	}{
		{
			name:              "future date - should schedule",
			scheduledDate:     futureDate,
			wantPrivacy:       "private",
			wantPublishAtSet:  true,
		},
		{
			name:              "past date - should be public",
			scheduledDate:     pastDate,
			wantPrivacy:       "public",
			wantPublishAtSet:  false,
		},
		{
			name:              "empty date - should be public",
			scheduledDate:     "",
			wantPrivacy:       "public",
			wantPublishAtSet:  false,
		},
		{
			name:              "invalid date - should be public",
			scheduledDate:     "not-a-date",
			wantPrivacy:       "public",
			wantPublishAtSet:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privacy, publishAt := getPublishStatus(tt.scheduledDate)
			if privacy != tt.wantPrivacy {
				t.Errorf("getPublishStatus() privacy = %v, want %v", privacy, tt.wantPrivacy)
			}
			if (publishAt != "") != tt.wantPublishAtSet {
				t.Errorf("getPublishStatus() publishAt set = %v, want %v", publishAt != "", tt.wantPublishAtSet)
			}
		})
	}
}

func TestCalculateShortPublishStatus(t *testing.T) {
	// Create a future date (1 year from now)
	futureDate := time.Now().AddDate(1, 0, 0).Format("2006-01-02T15:04")
	// Create a past date (1 year ago)
	pastDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02T15:04")

	tests := []struct {
		name              string
		mainVideoDate     string
		shortIndex        int
		wantPrivacy       string
		wantPublishAtSet  bool
	}{
		{
			name:              "future main video - short 0 should schedule",
			mainVideoDate:     futureDate,
			shortIndex:        0,
			wantPrivacy:       "private",
			wantPublishAtSet:  true,
		},
		{
			name:              "future main video - short 2 should schedule",
			mainVideoDate:     futureDate,
			shortIndex:        2,
			wantPrivacy:       "private",
			wantPublishAtSet:  true,
		},
		{
			name:              "past main video - short should be public",
			mainVideoDate:     pastDate,
			shortIndex:        0,
			wantPrivacy:       "public",
			wantPublishAtSet:  false,
		},
		{
			name:              "empty main video date - short should be public",
			mainVideoDate:     "",
			shortIndex:        0,
			wantPrivacy:       "public",
			wantPublishAtSet:  false,
		},
		{
			name:              "invalid main video date - short should be public",
			mainVideoDate:     "invalid",
			shortIndex:        0,
			wantPrivacy:       "public",
			wantPublishAtSet:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privacy, publishAt := calculateShortPublishStatus(tt.mainVideoDate, tt.shortIndex)
			if privacy != tt.wantPrivacy {
				t.Errorf("calculateShortPublishStatus() privacy = %v, want %v", privacy, tt.wantPrivacy)
			}
			if (publishAt != "") != tt.wantPublishAtSet {
				t.Errorf("calculateShortPublishStatus() publishAt set = %v, want %v", publishAt != "", tt.wantPublishAtSet)
			}
		})
	}
}

func TestBuildDubbedShortDescription(t *testing.T) {
	tests := []struct {
		name            string
		title           string
		originalVideoID string
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:            "with original video ID",
			title:           "Mi Short en Espanol",
			originalVideoID: "abc123",
			wantContains: []string{
				"Mi Short en Espanol",
				"Video completo",
				"https://youtu.be/abc123",
				"#Shorts",
			},
		},
		{
			name:            "without original video ID",
			title:           "Short sin video original",
			originalVideoID: "",
			wantContains: []string{
				"Short sin video original",
				"#Shorts",
			},
			wantNotContains: []string{
				"Video completo",
				"youtu.be",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDubbedShortDescription(tt.title, tt.originalVideoID)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("buildDubbedShortDescription() result missing expected string %q\nGot: %s", want, result)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(result, notWant) {
					t.Errorf("buildDubbedShortDescription() result should not contain %q\nGot: %s", notWant, result)
				}
			}
		})
	}
}

func TestUploadDubbedShort_Validation(t *testing.T) {
	// Create a temp file for testing file existence
	tempFile, err := os.CreateTemp("", "test-dubbed-short-*.mp4")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	tests := []struct {
		name       string
		video      *storage.Video
		shortIndex int
		wantErr    string
	}{
		{
			name: "negative short index",
			video: &storage.Video{
				Shorts: []storage.Short{{ID: "short1", Title: "Test"}},
			},
			shortIndex: -1,
			wantErr:    "invalid short index: -1",
		},
		{
			name: "short index out of range",
			video: &storage.Video{
				Shorts: []storage.Short{{ID: "short1", Title: "Test"}},
			},
			shortIndex: 5,
			wantErr:    "invalid short index: 5",
		},
		{
			name: "no shorts in video",
			video: &storage.Video{
				Shorts: []storage.Short{},
			},
			shortIndex: 0,
			wantErr:    "invalid short index: 0",
		},
		{
			name: "missing dubbing info for short",
			video: &storage.Video{
				Shorts:  []storage.Short{{ID: "short1", Title: "Test"}},
				Dubbing: map[string]storage.DubbingInfo{},
			},
			shortIndex: 0,
			wantErr:    "no dubbing info found for short: es:short1",
		},
		{
			name: "empty dubbed short path",
			video: &storage.Video{
				Shorts: []storage.Short{{ID: "short1", Title: "Test"}},
				Dubbing: map[string]storage.DubbingInfo{
					"es:short1": {
						DubbedVideoPath: "",
						Title:           "Titulo",
					},
				},
			},
			shortIndex: 0,
			wantErr:    "dubbed short video path is empty",
		},
		{
			name: "both translated and original title empty",
			video: &storage.Video{
				Shorts: []storage.Short{{ID: "short1", Title: ""}}, // original title empty
				Dubbing: map[string]storage.DubbingInfo{
					"es:short1": {
						DubbedVideoPath: "/some/path.mp4",
						Title:           "", // translated title also empty
					},
				},
			},
			shortIndex: 0,
			wantErr:    "short title is empty",
		},
		{
			name: "dubbed short file does not exist",
			video: &storage.Video{
				Shorts: []storage.Short{{ID: "short1", Title: "Test"}},
				Dubbing: map[string]storage.DubbingInfo{
					"es:short1": {
						DubbedVideoPath: "/nonexistent/path/short.mp4",
						Title:           "Titulo",
					},
				},
			},
			shortIndex: 0,
			wantErr:    "dubbed short file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UploadDubbedShort(tt.video, tt.shortIndex)
			if err == nil {
				t.Errorf("UploadDubbedShort() expected error containing %q, got nil", tt.wantErr)
				return
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("UploadDubbedShort() error = %q, want error containing %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestUploadDubbedShort_EmptyChannelID(t *testing.T) {
	// Create a temp file for testing
	tempFile, err := os.CreateTemp("", "test-dubbed-short-*.mp4")
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
		Shorts: []storage.Short{{ID: "short1", Title: "Test"}},
		Dubbing: map[string]storage.DubbingInfo{
			"es:short1": {
				DubbedVideoPath: tempFile.Name(),
				Title:           "Titulo en Espanol",
			},
		},
	}

	_, err = UploadDubbedShort(video, 0)
	if err == nil {
		t.Error("UploadDubbedShort() expected error for empty channel ID, got nil")
		return
	}
	if !strings.Contains(err.Error(), "Spanish channel ID is not configured") {
		t.Errorf("UploadDubbedShort() error = %q, want error about channel ID not configured", err.Error())
	}
}
