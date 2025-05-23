package slack

import (
	"testing"

	"devopstoolkit/youtube-automation/internal/configuration"

	"github.com/stretchr/testify/assert"
)

func TestGetTargetChannels(t *testing.T) {
	tests := []struct {
		name             string
		setupChannels    []string
		expectedChannels []string
		emptyIfNil       bool
	}{
		{
			name:             "Single channel ID",
			setupChannels:    []string{"C123"},
			expectedChannels: []string{"C123"},
		},
		{
			name:             "Multiple channel IDs",
			setupChannels:    []string{"C123", "C456", "C789"},
			expectedChannels: []string{"C123", "C456", "C789"},
		},
		{
			name:             "No channel IDs configured (nil slice)",
			setupChannels:    nil,
			expectedChannels: nil,  // Or []string{} depending on desired behavior for nil
			emptyIfNil:       true, // if GlobalSettings.Slack.TargetChannelIDs is nil, GetTargetChannels returns nil
		},
		{
			name:             "No channel IDs configured (empty slice)",
			setupChannels:    []string{},
			expectedChannels: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily set the global configuration for this test case
			originalSlackSettings := configuration.GlobalSettings.Slack
			configuration.GlobalSettings.Slack = configuration.SettingsSlack{
				TargetChannelIDs: tt.setupChannels,
			}
			defer func() {
				configuration.GlobalSettings.Slack = originalSlackSettings
			}()

			actualChannels := GetTargetChannels()

			if tt.emptyIfNil && tt.setupChannels == nil {
				assert.Nil(t, actualChannels, "Expected nil for nil setupChannels when emptyIfNil is true")
			} else if len(tt.expectedChannels) == 0 && len(actualChannels) == 0 {
				// Handles both nil and empty slice being effectively the same for "no channels"
				assert.Empty(t, actualChannels, "Expected empty slice")
			} else {
				assert.Equal(t, tt.expectedChannels, actualChannels)
			}
		})
	}
}
