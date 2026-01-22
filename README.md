# YouTube Automation

This project automates various aspects of managing a YouTube channel with a CLI interface.

## Demo Manifests and Code Used in DevOps Toolkit Videos

[![My Workflow With AI: How I Code, Test, and Deploy Faster Than Ever](https://img.youtube.com/vi/2E610yzqQwg/0.jpg)](https://youtu.be/2E610yzqQwg)
[![How I Fixed My Lazy Vibe Coding Habits with Taskmaster](https://img.youtube.com/vi/0WtCBbIHoKE/0.jpg)](https://youtu.be/0WtCBbIHoKE)

## Features

*   **Video Lifecycle Management**: Complete workflow from ideas to publishing and post-publish activities
*   **CLI Interface**: Interactive command-line interface for video management
*   **AI Content Generation**: AI-powered generation of titles, descriptions, tags, and tweets
*   **Video Uploads**: Automated YouTube video uploads
*   **Thumbnail Management**: Thumbnail creation and upload workflow  
*   **Metadata Handling**: Video titles, descriptions, tags, and metadata
*   **Social Media Integration**: BlueSky, LinkedIn, and Slack posting
*   **Hugo Integration**: Blog post generation and management
*   **Sponsorship Management**: Sponsor tracking and notification system

## Getting Started

### Prerequisites

*   Go (version 1.20 or higher recommended)
*   Google Cloud Project with YouTube Data API v3 enabled
*   OAuth 2.0 Client ID credentials (client_secret.json)

### Installation & Setup

1.  Clone the repository.
2.  Place your `client_secret.json` in the root directory.
3.  Build the executable: `go build`

### Configuration

For detailed configuration options, including setting default video languages, please see [docs/configuration.md](docs/configuration.md).

Global settings can be managed via `settings.yaml` in the root directory and command-line flags. See `internal/configuration/cli.go` for all available flags and their corresponding YAML paths.

## Usage

```bash
./youtube-automation --help
```

Interactive video management through terminal interface.

## Development

For development guidelines, project structure, and contribution information, please refer to [docs/development.md](docs/development.md).

## Contributing

(TODO: Add contribution guidelines)

## License

(TODO: Add license information)

<!-- Test comment for release automation -->
