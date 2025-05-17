# Development Guide

This guide provides instructions for setting up your development environment, building, running, testing, and contributing to the YouTube Automation Tool.

## Starting a Development Session

When starting a new development session with your AI assistant, use the following prompt:

```
Retrieve and process all information from the `memory` MCP knowledge graph to guide our session.
```

## Prerequisites

- Go 1.x
- YouTube API credentials (client_secret.json)
- Azure OpenAI API key
- Email account for notifications
- Hugo site repository (optional)
- Bluesky account (optional)

## Configuration

The tool uses a `settings.yaml` file for configuration:

```yaml
email:
  from: your-email@example.com
  thumbnailTo: thumbnail-creator@example.com
  editTo: video-editor@example.com
  financeTo: finance@example.com
ai:
  endpoint: https://your-openai-instance.openai.azure.com
  deployment: gpt-4-1106-preview
hugo:
  path: ../path-to-hugo-site/
youtube:
  channelId: YOUR_CHANNEL_ID
bluesky:
  identifier: username.bsky.social
  url: https://bsky.social/xrpc
```

Environment variables can also be used for sensitive information:
- `EMAIL_PASSWORD`
- `AI_KEY`
- `YOUTUBE_API_KEY`
- `BLUESKY_PASSWORD`

## Running the Tool

### With Devbox

```bash
devbox run run
```

### Build from Source

```bash
devbox run build
```

Or use the Makefile:

```bash
make build
```

## Usage

```
youtube-release [flags]
```

### Required Flags

- `--email-from` - Sender email address
- `--email-thumbnail-to` - Email for thumbnail requests
- `--email-edit-to` - Email for editing requests
- `--email-finance-to` - Email for financial matters
- `--email-password` - Email password
- `--ai-endpoint` - Azure OpenAI endpoint
- `--ai-key` - Azure OpenAI API key
- `--ai-deployment` - Azure OpenAI deployment name
- `--youtube-api-key` - YouTube API key
- `--hugo-path` - Path to Hugo site repository

### Optional Flags

- `--bluesky-identifier` - Bluesky username
- `--bluesky-password` - Bluesky password
- `--bluesky-url` - Bluesky API URL

## Project Structure

- `main.go` - Entry point
- `cli.go` - Command-line interface setup
- `youtube.go` - YouTube API integration
- `email.go` - Email notification system
- `hugo.go` - Hugo integration
- `bluesky.go` - Bluesky social media integration

## Contributing

[Add contribution guidelines]

(If you have specific contribution guidelines, they can be detailed here. This might include coding standards, commit message formats, PR processes, etc.)

## Testing and Code Coverage

This project aims for a test coverage goal of 80%. To check current test coverage, run:

```bash
./scripts/coverage.sh
```

This will generate a detailed coverage report, an HTML visualization, and identify areas needing improvement.

For comprehensive testing documentation including guidelines, examples, and best practices, see the [Testing Guide](docs/testing.md). Additional testing examples and patterns can be found in the [examples directory](docs/examples/). 