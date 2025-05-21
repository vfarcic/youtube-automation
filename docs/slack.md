# Slack Integration

The YouTube Automation tool includes integration with Slack for automated posting of new videos to designated Slack channels.

## Configuration

### In settings.yaml

```yaml
slack:
  token: xoxb-your-slack-token
  defaultChannel: general
  reactions:
    - thumbsup
    - rocket
```

### Environment Variables

You can also configure Slack using environment variables:

```bash
export SLACK_TOKEN="xoxb-your-slack-token"
```

### Command Line Flags

```bash
youtube-release --slack-token="xoxb-your-slack-token" --slack-default-channel="general"
```

## Creating a Slack App

To use the Slack integration, you need to create a Slack app with the appropriate permissions:

1. Go to [https://api.slack.com/apps](https://api.slack.com/apps)
2. Click "Create New App" → "From scratch"
3. Name your app and select your workspace
4. Under "OAuth & Permissions", add the following scopes:
   - `chat:write` - To post messages
   - `chat:write.public` - To post messages in channels the app is not a member of
   - `reactions:write` - To add reactions to messages
5. Install the app to your workspace
6. Copy the "Bot User OAuth Token" (starts with `xoxb-`) for use in the configuration

## Usage

Once configured, when a video is marked as "Slack posted" in the tool's interface, it will automatically:

1. Post the video information to the configured Slack channel(s)
2. Add any configured reactions to the post
3. Update the video's metadata with the posting information
4. Record the timestamp of the post

If the Slack token is not configured, the tool will fall back to copying the video URL to the clipboard for manual posting.

## Message Format

The Slack message includes:
- Video title
- Brief description
- Link to the YouTube video
- Thumbnail image (if available)

## Channel Selection

By default, videos are posted to the `defaultChannel` specified in the configuration. In future versions, channel selection based on video category will be supported.

## Troubleshooting

If you encounter issues with the Slack posting:

1. Verify your Slack token is correct and has the required permissions
2. Check that the default channel exists and the bot has access to post in it
3. Look for error messages in the console output
4. Ensure your Slack workspace allows bot users to post messages

For any other issues, please refer to the Slack API documentation at [https://api.slack.com/](https://api.slack.com/).