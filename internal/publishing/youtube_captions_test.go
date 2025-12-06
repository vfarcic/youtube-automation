package publishing

import (
	"strings"
	"testing"
)

func TestGetTranscript(t *testing.T) {
	tests := []struct {
		name        string
		videoID     string
		wantErr     bool
		errContains string
	}{
		{
			name:        "empty video ID",
			videoID:     "",
			wantErr:     true,
			errContains: "video ID cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetTranscript(tt.videoID)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetTranscript() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetTranscript() error = %v, want error containing %v", err, tt.errContains)
				}
			} else if err != nil {
				t.Errorf("GetTranscript() unexpected error = %v", err)
			}
		})
	}
}

func TestErrNoCaptions(t *testing.T) {
	if ErrNoCaptions == nil {
		t.Error("ErrNoCaptions should not be nil")
	}
	if ErrNoCaptions.Error() != "no captions available for this video" {
		t.Errorf("ErrNoCaptions message = %v, want 'no captions available for this video'", ErrNoCaptions.Error())
	}
}

func TestFormatSRTTime(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
		want    string
	}{
		{
			name:    "zero seconds",
			seconds: 0,
			want:    "00:00:00,000",
		},
		{
			name:    "simple seconds",
			seconds: 5.5,
			want:    "00:00:05,500",
		},
		{
			name:    "minutes and seconds",
			seconds: 125.75,
			want:    "00:02:05,750",
		},
		{
			name:    "hours minutes seconds",
			seconds: 3725.123,
			want:    "01:02:05,123",
		},
		{
			name:    "fraction handling",
			seconds: 1.999,
			want:    "00:00:01,999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSRTTime(tt.seconds)
			if got != tt.want {
				t.Errorf("formatSRTTime(%v) = %v, want %v", tt.seconds, got, tt.want)
			}
		})
	}
}
