# Slack Integration (`internal/slack`)

This package handles all interactions with the Slack API for posting notifications about new YouTube videos.

## Features

*   Posts new video notifications to configured Slack channels.
*   The message includes "📺 New Video" and a direct link to the YouTube video, which Slack unfurls to show a preview.

## Configuration

Configuration for the Slack integration is managed through a combination of an environment variable and the main `settings.yaml` file.

### 1. Slack API Token (Environment Variable)

*   **`SLACK_API_TOKEN`**: Your Slack Bot User OAuth Token. This token typically starts with `xoxb-`.
    *   **Setup**: Export this environment variable in your shell:
        ```bash
        export SLACK_API_TOKEN="your-xoxb-token-here"
        ```
    *   For instructions on how to generate a new token, see the official Slack tutorial: [How to quickly get and use a Slack API bot token](https://api.slack.com/tutorials/tracks/getting-a-token).
    *   You can add this to your shell's configuration file (e.g., `.zshrc`, `.bashrc`) to make it permanent.

### 2. Target Slack Channels (`settings.yaml`)

The specific Slack channels to post to are defined in the main `settings.yaml` file, typically located in the project's data directory (e.g., `~/.youtube-automation/settings.yaml`). This is part of the global application settings.

*   **Structure within `settings.yaml`**:
    ```yaml
    # ... other global settings ...

    slack:
      # A list of Slack Channel IDs where notifications will be sent.
      # These are channel IDs (e.g., "C0123456789"), not channel names (e.g., "#general").
      targetChannelIDs:
        - "C0123ABCDEF" # Example: #videos channel
        - "C0456GHIJKL" # Example: #announcements channel
      # retryAttempts: 3 # This setting is available in internal/slack/config.go but not currently exposed for modification via settings.yaml
      # timeoutSeconds: 10 # This setting is available in internal/slack/config.go but not currently exposed for modification via settings.yaml

    # ... other global settings ...
    ```
*   **Note**: The `internal/slack/channels.go` file retrieves this list of `TargetChannelIDs`. All video notifications are currently sent to all channels in this list.

## Usage

Slack posting is integrated into the interactive video publishing workflow of the `youtube-automation` CLI.

1.  Run the `youtube-automation` CLI.
2.  Navigate through the menus to select a video for publishing.
3.  When you confirm the video details and choose to proceed with publishing actions, if Slack integration is configured correctly (API token set and `targetChannelIDs` defined in `settings.yaml`), the application will automatically attempt to post a notification to the configured Slack channels.
4.  Success or failure of the Slack posting will be logged to the console, and the video's YAML file will be updated to reflect the Slack posting status.

There is no separate `youtube-automation slack` subcommand; the functionality is embedded within the main publishing flow.

## Error Handling

The Slack posting service includes error handling for:
*   Authentication failures (e.g., invalid token).
*   Network errors.
*   API errors from Slack.
*   Detailed error logging is provided via `logrus`.

## Troubleshooting

### Authentication Issues
*   Verify your `SLACK_API_TOKEN` environment variable is correctly set and exported.
*   Ensure the token is a valid Bot User OAuth Token (starts with `xoxb-`).
*   Confirm the Slack app associated with the token has been installed to your workspace.
*   Required OAuth Scopes for the bot token:
    *   `chat:write`: To post messages.
    *   `links:read` & `links:write`: To allow Slack to unfurl the YouTube links. (Ensure these are enabled in your app's configuration on api.slack.com if unfurling isn't working).

### No Messages Posted
*   Check that `settings.yaml` (in your data directory, e.g., `~/.youtube-automation/`) contains the `slack.targetChannelIDs` list and that it has valid Slack channel IDs.
*   Verify that the `SLACK_API_TOKEN` is available in the environment where the `youtube-automation` application is running.
*   Look for error messages in the console output when the application attempts to post to Slack.

### Message Formatting or Unfurling Issues
*   If YouTube links are not unfurling (showing a preview), ensure your Slack app has the `links:read` and `links:write` OAuth scopes. Also, check Slack's link unfurling settings for your app on `api.slack.com`.
*   Slack's link unfurling can be cached. If you recently changed how links are posted or if the link was previously posted without unfurling, try posting a brand new, unique YouTube link to test.
*   You can use Slack's `/debug unfurl YOUR_YOUTUBE_LINK` command to diagnose unfurling issues for a specific link.

---
*Future Enhancement Note: The `internal/slack/config.go` defines a more granular `SlackConfig` struct (e.g. with `DefaultChannel` and `CategoryChannels` which are not currently used by `channels.go`). This was part of an earlier design for more complex channel routing. Task #23 has been created to simplify `config.go` to align with the current, simpler channel targeting mechanism.* 