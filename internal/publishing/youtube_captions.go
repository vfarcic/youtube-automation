package publishing

import (
	"fmt"
	"strings"

	"github.com/horiagug/youtube-transcript-api-go/pkg/yt_transcript"
)

// ErrNoCaptions is returned when a video has no caption tracks available
var ErrNoCaptions = fmt.Errorf("no captions available for this video")

// GetTranscript fetches the transcript for a video using the youtube-transcript-api-go library.
// This works for auto-generated (ASR) captions which cannot be downloaded via the official API.
// Returns the transcript in SRT-like format with timestamps.
func GetTranscript(videoID string) (string, error) {
	if videoID == "" {
		return "", fmt.Errorf("video ID cannot be empty")
	}

	client := yt_transcript.NewClient()

	// Try English first, then fallback to any available language
	transcripts, err := client.GetTranscripts(videoID, []string{"en"})
	if err != nil {
		// Try without language preference
		transcripts, err = client.GetTranscripts(videoID, []string{})
		if err != nil {
			if strings.Contains(err.Error(), "no transcript") ||
				strings.Contains(err.Error(), "Transcript is disabled") ||
				strings.Contains(err.Error(), "disabled") {
				return "", ErrNoCaptions
			}
			return "", fmt.Errorf("failed to fetch transcript: %w", err)
		}
	}

	if len(transcripts) == 0 {
		return "", ErrNoCaptions
	}

	// Convert to SRT-like format
	var result strings.Builder
	for _, transcript := range transcripts {
		for i, line := range transcript.Lines {
			startTime := formatSRTTime(line.Start)
			endTime := formatSRTTime(line.Start + line.Duration)

			result.WriteString(fmt.Sprintf("%d\n", i+1))
			result.WriteString(fmt.Sprintf("%s --> %s\n", startTime, endTime))
			result.WriteString(line.Text + "\n\n")
		}
	}

	return result.String(), nil
}

// formatSRTTime converts seconds to SRT timestamp format (HH:MM:SS,mmm)
func formatSRTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}
