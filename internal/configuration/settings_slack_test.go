package configuration

import (
	"os"
	"testing"
)

func TestSlackChannelIDsFromEnv(t *testing.T) {
	// Save original value
	original := os.Getenv("SLACK_CHANNEL_IDS")
	defer func() {
		if original != "" {
			os.Setenv("SLACK_CHANNEL_IDS", original)
		} else {
			os.Unsetenv("SLACK_CHANNEL_IDS")
		}
	}()

	tests := []struct {
		name     string
		envValue string
		want     []string
	}{
		{
			name:     "single channel",
			envValue: "C123456789",
			want:     []string{"C123456789"},
		},
		{
			name:     "multiple channels",
			envValue: "C123456789,C987654321",
			want:     []string{"C123456789", "C987654321"},
		},
		{
			name:     "channels with spaces",
			envValue: " C123456789 , C987654321 ",
			want:     []string{"C123456789", "C987654321"},
		},
		{
			name:     "empty values filtered",
			envValue: "C123456789,,C987654321",
			want:     []string{"C123456789", "C987654321"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset GlobalSettings
			GlobalSettings = Settings{}

			os.Setenv("SLACK_CHANNEL_IDS", tt.envValue)

			if err := InitGlobalSettings(); err != nil {
				t.Fatalf("InitGlobalSettings() error = %v", err)
			}

			if len(GlobalSettings.Slack.TargetChannelIDs) != len(tt.want) {
				t.Errorf("got %d channels, want %d", len(GlobalSettings.Slack.TargetChannelIDs), len(tt.want))
			}

			for i, want := range tt.want {
				if i >= len(GlobalSettings.Slack.TargetChannelIDs) {
					t.Errorf("missing channel at index %d", i)
					continue
				}
				if GlobalSettings.Slack.TargetChannelIDs[i] != want {
					t.Errorf("channel[%d] = %q, want %q", i, GlobalSettings.Slack.TargetChannelIDs[i], want)
				}
			}
		})
	}
}
