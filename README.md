# YouTube Automation

This project automates various aspects of managing a YouTube channel. It provides an **HTTP API with a React web UI** for browser-based management of your YouTube workflow.

## Demo Manifests and Code Used in DevOps Toolkit Videos

[![My Workflow With AI: How I Code, Test, and Deploy Faster Than Ever](https://img.youtube.com/vi/2E610yzqQwg/0.jpg)](https://youtu.be/2E610yzqQwg)
[![How I Fixed My Lazy Vibe Coding Habits with Taskmaster](https://img.youtube.com/vi/0WtCBbIHoKE/0.jpg)](https://youtu.be/0WtCBbIHoKE)

## Features

* **Video Lifecycle Management**: Complete workflow from ideas to publishing and post-publish activities across 8 phases
* **HTTP API**: RESTful API exposing all backend functionality (see [API documentation](#http-api))
* **React Web UI**: Browser-based interface with dynamic form rendering from backend aspect metadata
* **AI Content Generation**: AI-powered generation of titles, descriptions, tags, tweets, shorts, thumbnails, and translations
* **Video Uploads**: Automated YouTube video and Short uploads
* **Google Drive Integration**: Upload/download video files and thumbnails via Google Drive
* **Thumbnail Management**: Thumbnail creation, upload, and variant tracking
* **Social Media Integration**: BlueSky, LinkedIn, Slack, Hacker News posting
* **Hugo Integration**: Blog post generation with deterministic URL construction
* **Sponsorship Management**: Sponsor tracking and notification system

## Getting Started

### Prerequisites

* Go 1.20 or higher
* Node.js 18+ (for frontend development)
* Google Cloud Project with YouTube Data API v3 enabled
* OAuth 2.0 Client ID credentials (`client_secret.json`)

### Installation & Setup

1. Clone the repository
2. Place your `client_secret.json` in the root directory
3. Build the executable: `go build -o youtube-release ./cmd/youtube-automation`

### Configuration

Global settings are managed via `settings.yaml` and command-line flags. See [docs/configuration.md](docs/configuration.md) for details.

#### Environment Variables

| Variable | Description |
|---|---|
| `API_TOKEN` | Bearer token for API authentication |
| `DATA_DIR` | Data directory for video YAML files (default: `./tmp`) |
| `EMAIL_PASSWORD` | Email account password |
| `AI_KEY` | Azure OpenAI API key |
| `YOUTUBE_API_KEY` | YouTube Data API key |
| `BLUESKY_PASSWORD` | BlueSky account password |

## Usage

### Starting the Server

Start the API server:

```bash
./youtube-release
```

#### Server Configuration

The server is configured via environment variables:

| Variable | Default | Description |
|---|---|---|
| `SERVER_HOST` | `localhost` | Host to listen on |
| `SERVER_PORT` | `8080` | Port to listen on |
| `API_TOKEN` | - | Bearer token for API authentication |
| `DATA_DIR` | `./tmp` | Data directory for video YAML files |

#### Authentication

All `/api/*` endpoints require a Bearer token in the `Authorization` header:

```bash
curl -H "Authorization: Bearer $API_TOKEN" http://localhost:8080/api/videos/phases
```

The `/health` endpoint is always public. If `API_TOKEN` is empty, authentication is disabled.

#### API Endpoints Overview

| Group | Endpoints | Description |
|---|---|---|
| Health | `GET /health` | Server health check |
| Videos | `GET/POST /api/videos`, `GET/PUT/PATCH/DELETE /api/videos/{name}` | Video CRUD and lifecycle |
| Phases | `GET /api/videos/phases` | Phase list with video counts |
| Progress | `GET /api/videos/{name}/progress[/{aspect}]` | Completion tracking |
| Categories | `GET /api/categories` | Available video categories |
| Aspects | `GET /api/aspects[/overview]`, `GET /api/aspects/{key}/fields` | Form field metadata |
| AI | `POST /api/ai/{type}/{category}/{name}` | AI content generation (titles, description, tags, tweets, shorts) |
| AI (body) | `POST /api/ai/thumbnails`, `POST /api/ai/translate`, `POST /api/ai/ama/*` | Thumbnail analysis, translation, AMA content |
| Drive | `POST /api/drive/upload/{type}/{name}`, `GET /api/drive/download/video/{name}` | Google Drive file operations |
| Actions | `POST /api/actions/request-{thumbnail,edit}/{name}` | Request thumbnail/edit with email notifications |
| Publishing | `POST /api/publish/youtube/{name}[/thumbnail,/shorts/{id}]` | YouTube uploads |
| Publishing | `POST /api/publish/hugo/{name}` | Hugo blog post creation |
| Publishing | `GET /api/publish/transcript/{videoId}`, `GET /api/publish/metadata/{videoId}` | YouTube data retrieval |
| Social | `POST /api/social/{platform}/{name}` | Social media posting |

For full API documentation see [openapi.yaml](openapi.yaml).

### Web UI

When the API server is running, the React frontend is served at the root URL (e.g., `http://localhost:8080`). The frontend provides:

* Phase dashboard with video counts and navigation
* Dynamic form rendering based on backend aspect metadata
* Inline AI content generation with apply-to-field UX
* Google Drive file upload/download for videos and thumbnails
* Action buttons for requesting thumbnails and edits

### Docker

Build and run with Docker:

```bash
docker build -t youtube-automation .
docker run -p 8080:8080 \
  -e API_TOKEN=your-token \
  -e AI_KEY=your-ai-key \
  -e SERVER_HOST=0.0.0.0 \
  youtube-automation
```

## Development

For development guidelines, project structure, and contribution information, see [docs/development.md](docs/development.md).

## License

MIT
