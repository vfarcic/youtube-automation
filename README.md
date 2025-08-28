# YouTube Automation

This project automates various aspects of managing a YouTube channel with both CLI and REST API interfaces.

## Demo Manifests and Code Used in DevOps Toolkit Videos

[![My Workflow With AI: How I Code, Test, and Deploy Faster Than Ever](https://img.youtube.com/vi/2E610yzqQwg/0.jpg)](https://youtu.be/2E610yzqQwg)
[![How I Fixed My Lazy Vibe Coding Habits with Taskmaster](https://img.youtube.com/vi/0WtCBbIHoKE/0.jpg)](https://youtu.be/0WtCBbIHoKE)

## Features

*   **Video Lifecycle Management**: Complete workflow from ideas to publishing and post-publish activities
*   **CLI Interface**: Interactive command-line interface for video management
*   **REST API**: Comprehensive REST API for programmatic access
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
- `GET /api/videos/list?phase={id}` - Optimized lightweight video list with phase data (0-7) for frontend grids (includes string-based IDs)
- `GET /api/videos/{name}?category={cat}` - Get video details (includes string-based ID)
- `PUT /api/videos/{name}` - Update video
- `DELETE /api/videos/{name}?category={cat}` - Delete video
- `PUT /api/videos/{name}/{phase}` - Update specific phase

**AI Content Generation:**

*Traditional endpoints (JSON payload):*
- `POST /api/ai/titles` - Generate video titles from manuscript
- `POST /api/ai/description` - Generate video description from manuscript
- `POST /api/ai/tags` - Generate video tags from manuscript
- `POST /api/ai/tweets` - Generate social media tweets from manuscript
- `POST /api/ai/description-tags` - Generate description with hashtags from manuscript

*Optimized endpoints (URL parameters - recommended for existing videos):*
- `POST /api/ai/titles/{videoName}?category={cat}` - Generate titles for specific video
- `POST /api/ai/description/{videoName}?category={cat}` - Generate description for specific video
- `POST /api/ai/tags/{videoName}?category={cat}` - Generate tags for specific video
- `POST /api/ai/tweets/{videoName}?category={cat}` - Generate tweets for specific video
- `POST /api/ai/description-tags/{videoName}?category={cat}` - Generate description with hashtags for specific video

**Animations endpoint:**
- `GET /api/animations/{videoName}?category={cat}` - **NEW**: Get animation cues from video manuscript (non-AI parsing)

**Editing aspects metadata:**
- `GET /api/editing/aspects` - **NEW**: Get editing aspects overview (lightweight, ~1KB)
  - **NEW**: Optional query parameters: `?videoName={name}&category={cat}` for progress tracking
  - **NEW**: Returns `completedFieldCount` alongside `fieldCount` for progress indicators ("6/8 fields completed")
- `GET /api/editing/aspects/{aspectKey}/fields` - **NEW**: Get detailed field metadata for dynamic form generation

**Phase-specific endpoints:**
- `/initial-details` - Project information and sponsorship
- `/work-progress` - Content creation tasks
- `/definition` - Title, description, metadata
- `/post-production` - Editing and thumbnails
- `/publishing` - Video upload and Hugo posts
- `/post-publish` - Social media and follow-up tasks

## Frontend Integration

### String-Based Video IDs

**Important:** All video responses now include a string-based `id` field in the format `category/name` (e.g., `"tutorials/kubernetes-guide"`). This provides consistent, human-readable identifiers for frontend applications.

```javascript
// Example video object structure
{
  "id": "tutorials/kubernetes-guide",     // NEW: String-based ID
  "name": "kubernetes-guide",             // NEW: Filename for easy access
  "category": "tutorials",                // Category
  "title": "Kubernetes Guide",           // Display title
  // ... other fields
}
```

### Fetching Editing Aspects

```javascript
// Get all available aspects (lightweight overview)
const aspectsResponse = await fetch('/api/editing/aspects');
const { aspects } = await aspectsResponse.json();

// NEW: Get aspects with progress tracking for a specific video
const progressResponse = await fetch('/api/editing/aspects?videoName=my-video&category=tutorials');
const { aspects: aspectsWithProgress } = await progressResponse.json();

// Get detailed fields for a specific aspect
const fieldsResponse = await fetch(`/api/editing/aspects/${aspectKey}/fields`);
const { aspectKey, aspectTitle, fields } = await fieldsResponse.json();
```

### Dynamic UI Rendering

Use the aspects metadata to render dynamic editing interfaces:

```javascript
// Render navigation tabs with progress indicators
function renderAspectTabs(aspects) {
  return aspects.map(aspect => ({
    key: aspect.key,
    title: aspect.title,
    description: aspect.description,
    icon: aspect.icon,
    fieldCount: aspect.fieldCount,
    completedFieldCount: aspect.completedFieldCount, // NEW: Progress tracking
    order: aspect.order,
    progressPercentage: Math.round((aspect.completedFieldCount / aspect.fieldCount) * 100) // NEW: Calculate percentage
  }));
}

// NEW: Render progress indicators
function renderProgressIndicator(aspect) {
  const percentage = Math.round((aspect.completedFieldCount / aspect.fieldCount) * 100);
  return `
    <div class="progress-container">
      <div class="progress-bar" style="width: ${percentage}%"></div>
      <span class="progress-text">${aspect.completedFieldCount}/${aspect.fieldCount} fields completed</span>
    </div>
  `;
}

// Generate form fields from metadata
function renderFormFields(fields) {
  return fields.map(field => ({
    name: field.name,
    fieldName: field.fieldName,  // NEW: Use for mapping to video data
    type: field.type,
    required: field.required,
    description: field.description,
    uiHints: field.uiHints,
    defaultValue: field.defaultValue
  }));
}

// NEW: Map field values between UI and API
function mapFieldValues(fieldMetadata, videoData) {
  const mappedValues = {};
  fieldMetadata.forEach(field => {
    // Use fieldName to get the correct property from video data
    mappedValues[field.name] = videoData[field.fieldName];
  });
  return mappedValues;
}
```

### AI Content Generation Integration

The API provides two approaches for AI content generation:

#### Traditional Approach (JSON Payload)
Use when you have manuscript content from external sources or arbitrary text:

```javascript
// Generate video titles from manuscript content
async function generateTitles(manuscript) {
  const response = await fetch('/api/ai/titles', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ manuscript })
  });
  const { titles } = await response.json();
  return titles; // Array of 3 title suggestions
}
```

#### Optimized Approach (URL Parameters - Recommended)
Use when working with existing videos in your system:

```javascript
// Generate video titles for existing video (optimized)
async function generateTitlesOptimized(videoName, category) {
  const response = await fetch(`/api/ai/titles/${videoName}?category=${category}`, {
    method: 'POST'
  });
  const { titles } = await response.json();
  return titles; // Array of 3 title suggestions
}

// Generate complete video metadata (traditional approach)
async function generateVideoMetadata(manuscript) {
  const [titles, description, tags, tweets] = await Promise.all([
    fetch('/api/ai/titles', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ manuscript }) }),
    fetch('/api/ai/description', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ manuscript }) }),
    fetch('/api/ai/tags', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ manuscript }) }),
    fetch('/api/ai/tweets', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ manuscript }) })
  ]);

  return {
    titles: (await titles.json()).titles,
    description: (await description.json()).description,
    tags: (await tags.json()).tags,
    tweets: (await tweets.json()).tweets
  };
}

// Generate complete video metadata (optimized approach - recommended)
async function generateVideoMetadataOptimized(videoName, category) {
  const [titles, description, tags, tweets] = await Promise.all([
    fetch(`/api/ai/titles/${videoName}?category=${category}`, { method: 'POST' }),
    fetch(`/api/ai/description/${videoName}?category=${category}`, { method: 'POST' }),
    fetch(`/api/ai/tags/${videoName}?category=${category}`, { method: 'POST' }),
    fetch(`/api/ai/tweets/${videoName}?category=${category}`, { method: 'POST' })
  ]);

  return {
    titles: (await titles.json()).titles,
    description: (await description.json()).description,
    tags: (await tags.json()).tags,
    tweets: (await tweets.json()).tweets
  };
}

// Generate description with hashtags (traditional approach)
async function generateDescriptionWithTags(manuscript) {
  const response = await fetch('/api/ai/description-tags', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ manuscript })
  });
  const { description, tags } = await response.json();
  return { description, tags }; // Description text + hashtag string
}

// Generate description with hashtags (optimized approach - recommended)
async function generateDescriptionWithTagsOptimized(videoName, category) {
  const response = await fetch(`/api/ai/description-tags/${videoName}?category=${category}`, {
    method: 'POST'
  });
  const { description, tags } = await response.json();
  return { description, tags }; // Description text + hashtag string
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

// NEW: Handle progress tracking errors
try {
  const response = await fetch('/api/editing/aspects?videoName=test&category=missing');
  const { aspects } = await response.json();
  // Gracefully handles non-existent videos with 0 completion counts
} catch (error) {
  console.error('Progress tracking error:', error);
}

// AI API error handling
try {
  const response = await fetch('/api/ai/titles', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ manuscript: '' }) // Empty manuscript
  });
  if (!response.ok) {
    const error = await response.json();
    console.error('AI API Error:', error.error); // "manuscript field is required and cannot be empty"
  }
} catch (error) {
  console.error('AI Generation Error:', error);
}
```

### Field Types and Formats

The API supports several field types with specific formatting requirements:

#### Date Fields
- **Format**: `YYYY-MM-DDTHH:MM` (e.g., "2024-01-15T14:30")
- **Input Type**: `datetime` for UTC compatibility
- **Validation**: Must match Go time format `2006-01-02T15:04`

#### Text Fields
- **Multi-line**: Use `textarea` input type with configurable rows
- **Single-line**: Use `text` input type
- **Examples**: Description, Tags, Timecodes

#### Boolean Fields
- **Input Type**: `checkbox`
- **Completion**: Often uses `true_only` criteria
- **Examples**: Code completion, Upload status

#### String Fields
- **Input Type**: `text`
- **Validation**: Optional pattern matching
- **Examples**: Title, Project Name

**Benefits:**
- **93% smaller payload** for overview endpoint (~1KB vs ~15KB)
- **Dynamic form generation** from field metadata
- **Consistent field ordering** and validation rules
- **Rich UI hints** for optimal user experience
- **NEW**: **Real-time progress tracking** with completion counts
- **NEW**: **Backend consistency** - uses same logic as CLI progress calculations
- **NEW**: **Graceful error handling** for missing videos

For complete API documentation and testing examples, see [docs/api-manual-testing.md](docs/api-manual-testing.md).

## Development

For development guidelines, project structure, and contribution information, please refer to [docs/development.md](docs/development.md).

## Contributing

(TODO: Add contribution guidelines)

## License

(TODO: Add license information)

<!-- Test comment for release automation -->
