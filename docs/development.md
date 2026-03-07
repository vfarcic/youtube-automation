# Development Guide

This guide covers setting up your development environment, building, running, testing, and contributing to the YouTube Automation Tool.

## Prerequisites

- Go 1.20+
- Node.js 18+ (for frontend)
- YouTube API credentials (`client_secret.json`)
- Azure OpenAI API key
- Email account for notifications (optional)
- Hugo site repository (optional)
- Bluesky account (optional)

## Project Structure

```
├── cmd/youtube-automation/     # Application entry point
├── internal/
│   ├── ai/                     # AI content generation (titles, descriptions, tags, etc.)
│   ├── api/                    # HTTP API server, handlers, middleware
│   ├── app/                    # CLI interactive interface (huh forms)
│   ├── aspect/                 # Aspect system (dynamic form generation, completion tracking)
│   ├── configuration/          # Settings, CLI flags, YAML config
│   ├── filesystem/             # File operations, path resolution
│   ├── gdrive/                 # Google Drive integration
│   ├── platform/               # Social media platform integrations
│   ├── publishing/             # YouTube upload, Hugo blog generation
│   ├── service/                # Business logic layer (VideoService)
│   ├── slack/                  # Slack integration
│   ├── storage/                # YAML-based data persistence
│   ├── thumbnail/              # Thumbnail resolution and handling
│   ├── video/                  # Phase calculation, workflow management
│   └── workflow/               # Phase constants and definitions
├── web/                        # React + TypeScript frontend
│   ├── src/
│   │   ├── components/         # UI components (forms, dashboard, navigation)
│   │   ├── hooks/              # React hooks
│   │   ├── services/           # API client
│   │   └── types/              # TypeScript type definitions
│   ├── package.json
│   └── vite.config.ts
├── pkg/
│   ├── mocks/                  # Test mocks
│   └── testutil/               # Test utilities and fixtures
├── scripts/                    # Build, coverage, and utility scripts
├── openapi.yaml                # OpenAPI 3.1 specification
├── settings.yaml               # Global configuration
└── Dockerfile                  # Multi-stage container build
```

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

Environment variables for sensitive information:
- `EMAIL_PASSWORD`
- `AI_KEY`
- `YOUTUBE_API_KEY`
- `BLUESKY_PASSWORD`

## Building

```bash
# Build CLI binary
make build-local
# or
go build -o youtube-release ./cmd/youtube-automation

# Build for all platforms
make build

# Build frontend
cd web && npm install && npm run build

# Build Docker image
docker build -t youtube-automation .

# Clean build artifacts
make clean
```

## Running

### CLI Mode

```bash
./youtube-release [flags]
```

### API Server Mode

```bash
# Start with defaults (localhost:8080)
./youtube-release serve

# Custom host/port with authentication
./youtube-release serve --host 0.0.0.0 --port 9090 --api-token my-secret

# Using environment variables
API_TOKEN=my-secret DATA_DIR=./data ./youtube-release serve
```

The API server serves both the REST API (`/api/*`) and the React frontend (all other paths).

### Frontend Development

For frontend development with hot reload:

```bash
# Terminal 1: Start the Go API server
./youtube-release serve --api-token dev-token

# Terminal 2: Start the Vite dev server with API proxy
cd web
npm install
npm run dev
```

The Vite dev server (port 5173) proxies `/api` and `/health` requests to the Go server (port 8080).

## Testing

```bash
# Run all Go tests
go test ./...

# Run Go tests with coverage
go test ./... -cover

# Generate detailed coverage report
./scripts/coverage.sh

# Run specific package tests
go test ./internal/api/...

# Run specific test function with verbose output
go test -v -run TestHandleGetPhases ./internal/api/

# Run frontend tests
cd web && npm test

# Run frontend tests in watch mode
cd web && npm run test:watch

# Check for brittle tests
./scripts/find_brittle_tests.sh
```

### Coverage Goal

The project targets 80% test coverage. API handlers are currently at 82.3%.

## API Development

### Adding a New Endpoint

1. Define the handler in the appropriate `handlers_*.go` file in `internal/api/`
2. Register the route in `server.go` `setupRoutes()`
3. Add request/response types as needed
4. Write tests using the test helper in `testhelper_test.go`
5. Update `openapi.yaml` with the new endpoint

### Handler Pattern

All handlers follow a consistent pattern:

```go
func (s *Server) handleMyEndpoint(w http.ResponseWriter, r *http.Request) {
    // 1. Extract and validate parameters
    videoName := chi.URLParam(r, "videoName")
    category := r.URL.Query().Get("category")
    if category == "" {
        respondError(w, http.StatusBadRequest, "Missing category", "")
        return
    }

    // 2. Business logic via service layer
    result, err := s.videoService.DoSomething(videoName, category)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed", err.Error())
        return
    }

    // 3. Return JSON response
    respondJSON(w, http.StatusOK, result)
}
```

### Authentication

- All `/api/*` routes use bearer token middleware
- `/health` is always public
- Set `API_TOKEN` env var or `--api-token` flag
- Empty token disables authentication

## Frontend Development

### Tech Stack

- **React 19** with TypeScript
- **Vite** for build and dev server
- **TanStack Query** for data fetching
- **Zustand** for state management
- **Tailwind CSS** for styling
- **Vitest** + React Testing Library for tests

### Key Patterns

- **Dynamic form rendering**: Frontend reads aspect metadata from the API and renders forms without hardcoding field names. New fields added to the Go backend automatically appear in the UI.
- **Dual-mode file storage**: CLI uses local file paths, Web UI uses Google Drive file IDs. Both coexist via the `DriveFileID` field pattern.
- **Action buttons**: `RequestThumbnail` and `RequestEdit` are rendered as buttons (not checkboxes) to communicate irreversible intent.

## Field Completion System

The project uses a reflection-based completion system via struct tags in `internal/storage/yaml.go`:

```go
type Video struct {
    Date string `json:"date" completion:"filled_only"`
    Code bool   `json:"code" completion:"true_only"`
}
```

### Completion Criteria

- `filled_only` - Complete when field has any non-empty value
- `true_only` - Complete only when boolean field is `true`
- `false_only` - Complete only when boolean field is `false`
- `conditional_sponsorship` - Special logic for sponsorship emails
- `conditional_sponsors` - Special logic for sponsor notifications
- `empty_or_filled` - Always considered complete
- `no_fixme` - Complete when field doesn't contain "FIXME"

## Contributing

1. Create a feature branch from `main`
2. Write tests for all new functionality (80% coverage target)
3. Ensure `go test ./...` and `cd web && npm test` pass
4. Update `openapi.yaml` if API endpoints change
5. Submit a pull request

## Version Management

```bash
make bump-patch    # Bump patch version
make bump-minor    # Bump minor version
make bump-major    # Bump major version
```
