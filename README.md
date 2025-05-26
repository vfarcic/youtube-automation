# YouTube Automation

This project automates various aspects of managing a YouTube channel.

## Features

*   Video Uploads
*   Thumbnail Management
*   Metadata Handling
*   REST API for programmatic interaction
*   Playlist Management (TODO)
*   Comments Interaction (TODO)

## Getting Started

### Prerequisites

*   Go (version 1.20 or higher recommended)
*   Google Cloud Project with YouTube Data API v3 enabled
*   OAuth 2.0 Client ID credentials (client_secret.json)

### Installation & Setup

1.  Clone the repository.
2.  Place your `client_secret.json` in the root directory.
3.  Build the executable: `go build`

### Usage

#### CLI Mode
Run the application with required configuration flags:
```bash
./youtube-automation --email-from="..." --email-thumbnail-to="..." [other flags]
```

#### API Mode
Run the application in API mode:
```bash
./youtube-automation --api --api-port=8080 [other required flags]
```

For detailed API documentation, see [docs/api_reference.md](docs/api_reference.md).

### Configuration

For detailed configuration options, including setting default video languages, please see [docs/configuration.md](docs/configuration.md).

Global settings can be managed via `settings.yaml` in the root directory and command-line flags. See `internal/configuration/cli.go` for all available flags and their corresponding YAML paths.

## Usage

```bash
./youtube-automation --help
```

(More specific usage examples to be added)

## Development

For development guidelines, project structure, and contribution information, please refer to [docs/development.md](docs/development.md).

## Contributing

(TODO: Add contribution guidelines)

## License

(TODO: Add license information)

<!-- Test comment for release automation -->
