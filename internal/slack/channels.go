package slack

import "devopstoolkit/youtube-automation/internal/configuration"

// getTargetChannels_actual retrieves the predefined list of Slack channel IDs from the global configuration.
func getTargetChannels_actual() []string {
	return configuration.GlobalSettings.Slack.TargetChannelIDs
}

// GetTargetChannels is a function variable for retrieving target channels.
// This allows it to be replaced for mocking in tests.
var GetTargetChannels = getTargetChannels_actual
