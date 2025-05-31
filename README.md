# YouTube Automation

This project automates various aspects of managing a YouTube channel with both CLI and REST API interfaces.

## Features

*   **Video Lifecycle Management**: Complete workflow from ideas to publishing and post-publish activities
*   **CLI Interface**: Interactive command-line interface for video management
*   **REST API**: Comprehensive REST API for programmatic access
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

### CLI Mode (Default)
```bash
./youtube-automation --help
```

Interactive video management through terminal interface.

### API Server Mode
```bash
./youtube-automation --api-enabled --api-port 8080
```

Starts the REST API server. See [docs/api-manual-testing.md](docs/api-manual-testing.md) for API usage examples.

### API Endpoints
- `GET /health` - Health check
- `GET /api/categories` - List video categories
- `POST /api/videos` - Create new video
- `GET /api/videos/phases` - Get video phase summary
- `GET /api/videos?phase={id}` - List videos in phase
- `GET /api/videos/list?phase={id}` - **NEW**: Optimized lightweight video list for frontend grids
- `GET /api/videos/{name}?category={cat}` - Get video details
- `PUT /api/videos/{name}` - Update video
- `DELETE /api/videos/{name}?category={cat}` - Delete video
- `PUT /api/videos/{name}/{phase}` - Update specific phase

**Phase-specific endpoints:**
- `/initial-details` - Project information and sponsorship
- `/work-progress` - Content creation tasks
- `/definition` - Title, description, metadata
- `/post-production` - Editing and thumbnails
- `/publishing` - Video upload and Hugo posts
- `/post-publish` - Social media and follow-up tasks

## Development

For development guidelines, project structure, and contribution information, please refer to [docs/development.md](docs/development.md).

## Contributing

(TODO: Add contribution guidelines)

## License

(TODO: Add license information)

<!-- Test comment for release automation -->
