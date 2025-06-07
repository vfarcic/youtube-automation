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
- `GET /api/videos/list?phase={id}` - Optimized lightweight video list with phase data (0-7) for frontend grids
- `GET /api/videos/{name}?category={cat}` - Get video details
- `PUT /api/videos/{name}` - Update video
- `DELETE /api/videos/{name}?category={cat}` - Delete video
- `PUT /api/videos/{name}/{phase}` - Update specific phase

**Editing aspects metadata:**
- `GET /api/editing/aspects` - **NEW**: Get editing aspects overview (lightweight, ~1KB)
- `GET /api/editing/aspects/{aspectKey}/fields` - **NEW**: Get detailed field metadata for dynamic form generation

**Phase-specific endpoints:**
- `/initial-details` - Project information and sponsorship
- `/work-progress` - Content creation tasks
- `/definition` - Title, description, metadata
- `/post-production` - Editing and thumbnails
- `/publishing` - Video upload and Hugo posts
- `/post-publish` - Social media and follow-up tasks

## Frontend Integration

### Fetching Editing Aspects

```javascript
// Get all available aspects (lightweight overview)
const aspectsResponse = await fetch('/api/editing/aspects');
const { aspects } = await aspectsResponse.json();

// Get detailed fields for a specific aspect
const fieldsResponse = await fetch(`/api/editing/aspects/${aspectKey}/fields`);
const { aspectKey, aspectTitle, fields } = await fieldsResponse.json();
```

### Dynamic UI Rendering

Use the aspects metadata to render dynamic editing interfaces:

```javascript
// Render navigation tabs
function renderAspectTabs(aspects) {
  return aspects.map(aspect => ({
    key: aspect.key,
    title: aspect.title,
    description: aspect.description,
    icon: aspect.icon,
    fieldCount: aspect.fieldCount,
    order: aspect.order
  }));
}

// Generate form fields from metadata
function renderFormFields(fields) {
  return fields.map(field => ({
    name: field.name,
    type: field.type,
    required: field.required,
    description: field.description,
    uiHints: field.uiHints,
    defaultValue: field.defaultValue
  }));
}
```

### Error Handling

```javascript
try {
  const response = await fetch('/api/editing/aspects/invalid-key/fields');
  if (!response.ok) {
    const error = await response.json();
    console.error('API Error:', error.error); // "aspect not found"
  }
} catch (error) {
  console.error('Network Error:', error);
}
```

**Benefits:**
- **93% smaller payload** for overview endpoint (~1KB vs ~15KB)
- **Dynamic form generation** from field metadata
- **Consistent field ordering** and validation rules
- **Rich UI hints** for optimal user experience

For complete API documentation and testing examples, see [docs/api-manual-testing.md](docs/api-manual-testing.md).

## Development

For development guidelines, project structure, and contribution information, please refer to [docs/development.md](docs/development.md).

## Contributing

(TODO: Add contribution guidelines)

## License

(TODO: Add license information)

<!-- Test comment for release automation -->
