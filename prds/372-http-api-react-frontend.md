# PRD: HTTP API and React Frontend for Video Management

**Issue**: #372
**Status**: In Progress
**Priority**: Medium
**Created**: 2026-03-05

---

## Problem Statement

All features are locked behind a CLI interface, which:
- Limits accessibility to terminal-proficient users
- Makes it harder to visualize video workflows and phase progress
- Prevents browser-based access or collaboration through a shared UI
- Requires memorizing menu navigation for 100+ features across 8 phases

## Proposed Solution

Add two layers to the existing system:

1. **HTTP API layer** (`internal/api/`): Exposes all existing functionality — video lifecycle, AI content generation, publishing, social media, analytics — as RESTful endpoints. Uses the existing service layer, aspect system, and video manager directly.

2. **React + TypeScript frontend** (`web/`): A single-page application that dynamically renders forms from the backend's aspect metadata. The frontend is **smart about presentation, dumb about business rules** — it owns layout, UX, and interactions, but defers all business logic (validation, phase calculation, completion tracking) to the API.

### Key Architectural Principle

The existing **aspect system** (`internal/aspect/`) already generates typed field metadata with UI hints, validation hints, and completion criteria. The API exposes this metadata as-is. The frontend consumes it to dynamically render forms. When a new field is added to the Go `Video` struct with appropriate tags, it automatically appears in the frontend — no frontend code changes needed for new fields.

### What Does NOT Change
- CLI continues to work exactly as today
- All business logic stays in Go (service layer, video manager, aspect system)
- YAML storage layer unchanged
- AI modules, publishing, social media modules unchanged

### User Journey

**Before**: Launch CLI → navigate menus → fill huh forms → repeat

**After**: Open browser → see phase dashboard with video counts → click a phase → see video list with progress bars → click a video → edit through tabbed aspect forms → trigger AI generation → upload to YouTube — all in a visual interface

## Success Criteria

### Must Have (MVP)
- [ ] HTTP API serves all video lifecycle operations (CRUD, phase, progress)
- [ ] API exposes aspect metadata for dynamic form rendering
- [ ] API serves AI content generation (titles, description, tags, tweets)
- [ ] API serves publishing operations (YouTube upload, Hugo blog post)
- [x] React frontend renders phase dashboard with video counts
- [ ] Frontend dynamically renders aspect-based editing forms from API metadata
- [ ] Frontend shows progress tracking per aspect and overall
- [ ] Frontend supports AI content generation with apply-to-field UX
- [x] API protected by bearer token auth (env var, disabled when unset)
- [x] Go server embeds and serves the built frontend (single binary deployment)
- [ ] Helm chart deploys backend + frontend to Kubernetes
- [ ] GHA builds and pushes container images to ghcr.io
- [ ] 80% test coverage on API handlers

### Nice to Have (Future)
- [ ] WebSocket/SSE for real-time progress on long-running operations
- [ ] Analytics dashboard with charts
- [ ] Drag-and-drop for video phase transitions
- [ ] Keyboard shortcuts matching CLI muscle memory
- [ ] Dark mode
- [ ] Mobile-responsive layout

## Technical Scope

### Existing Foundation

The codebase already has:
- **`openapi.yaml`** (2420 lines): OpenAPI 3.1.0 spec covering ~26 endpoints (video CRUD, phases, lightweight list, aspect metadata, aspect-specific PATCH, categories, animations, AI generation)
- **`go-chi/chi/v5`**: Already in `go.mod` — the router for the API
- **Service layer** (`internal/service/video_service.go`): 11 clean, stateless methods
- **Aspect system** (`internal/aspect/`): JSON-serializable form metadata with field types, UI hints, validation hints, completion criteria
- **Video manager** (`internal/video/manager.go`): Pure functions for phase calculation and progress tracking

### 1. HTTP API Layer (`internal/api/`)

**Router**: chi (already a dependency)

**Handler organization by domain**:
```
internal/api/
  server.go              -- Server struct, chi router setup, embedded frontend
  middleware.go           -- CORS, logging, recovery, request ID
  errors.go              -- Standardized error response helpers
  handlers_video.go      -- Video CRUD, phase listing, lifecycle
  handlers_aspect.go     -- Aspect metadata, field completion
  handlers_ai.go         -- AI content generation endpoints
  handlers_publish.go    -- YouTube upload, Hugo blog, shorts
  handlers_social.go     -- BlueSky, LinkedIn, Slack, HN posting
  handlers_analytics.go  -- YouTube analytics, timing, title analysis
  handlers_config.go     -- Settings, categories, timing recommendations
  handlers_ama.go        -- AMA content generation
  handlers_dubbing.go    -- Dubbing operations
  sse.go                 -- Server-Sent Events for long-running operations
```

**Authentication**: Bearer token via `API_TOKEN` environment variable. Empty/unset = auth disabled (local dev). In Kubernetes, delivered via Secret (not committed to git). `/health` is always public (K8s probes). Uses `crypto/subtle.ConstantTimeCompare` for timing-safe comparison.

**Long-running operations** (YouTube upload, AI generation, analytics): SSE (`text/event-stream`) for progress updates. Start with synchronous for simple operations, add SSE incrementally.

**Error pattern**:
```json
{"status": 400, "message": "Invalid phase", "detail": "Phase must be 0-7"}
```

**Configuration**: Add `--serve` flag (or `serve` subcommand) to existing CLI. The `init()` in `internal/configuration/cli.go` loads settings from `settings.yaml` and env vars — this works for the server too. The cobra `MarkFlagRequired` calls only enforce on CLI execution, not on server mode. A new `cmd/youtube-automation/serve.go` adds a `serve` cobra subcommand.

### 2. API Endpoints (~62 total)

**Video Lifecycle** (8):
- `POST /api/videos` — Create video
- `GET /api/videos?phase={n}` — List by phase
- `GET /api/videos/list?phase={n}` — Lightweight list
- `GET /api/videos/phases` — Phase counts
- `GET /api/videos/{category}/{name}` — Get video
- `PUT /api/videos/{category}/{name}` — Update video
- `DELETE /api/videos/{category}/{name}` — Delete video
- `POST /api/videos/{category}/{name}/archive` — Archive

**Aspect Metadata** (4):
- `GET /api/aspects` — All aspects with fields
- `GET /api/aspects/overview` — Lightweight summary
- `GET /api/aspects/{key}/fields` — Single aspect fields
- `GET /api/aspects/{key}/fields/{field}/completion` — Completion criteria

**Aspect-Specific Updates** (7):
- `PATCH /api/videos/{category}/{name}/{aspectKey}` — Partial update per aspect

**Progress** (2):
- `GET /api/videos/{category}/{name}/progress` — Overall
- `GET /api/videos/{category}/{name}/progress/{aspect}` — Per-aspect

**AI Generation** (12):
- `POST /api/ai/titles/{category}/{name}` — Title suggestions
- `POST /api/ai/description/{category}/{name}` — Description
- `POST /api/ai/tags/{category}/{name}` — Tags
- `POST /api/ai/tweets/{category}/{name}` — Tweets
- `POST /api/ai/description-tags/{category}/{name}` — Hashtags
- `POST /api/ai/shorts/{category}/{name}` — Shorts analysis
- `POST /api/ai/thumbnails` — Thumbnail variations
- `POST /api/ai/translate` — Video metadata translation
- `POST /api/ai/ama/content` — AMA all-in-one
- `POST /api/ai/ama/title` — AMA title
- `POST /api/ai/ama/timecodes` — AMA timecodes
- `POST /api/ai/ama/description` — AMA description

**Publishing** (6):
- `POST /api/publish/youtube/{category}/{name}` — Upload video
- `POST /api/publish/youtube/{category}/{name}/shorts/{shortId}` — Upload short
- `POST /api/publish/hugo/{category}/{name}` — Create blog post
- `POST /api/publish/dubbed/{category}/{name}` — Upload dubbed video
- `GET /api/publish/transcript/{videoId}` — Get transcript
- `GET /api/publish/metadata/{videoId}` — Get YouTube metadata

**Social Media** (5):
- `POST /api/social/{platform}/{category}/{name}` — Post to BlueSky, LinkedIn, Slack, DOT, HN

**Analytics** (5):
- `GET /api/analytics/videos` — Video analytics
- `POST /api/analytics/titles` — Title analysis
- `POST /api/analytics/timing` — Timing recommendations
- `POST /api/analytics/sponsor-page` — Update sponsor page
- `GET /api/analytics/channel` — Channel demographics/stats

**Configuration** (4):
- `GET /api/categories` — List categories
- `GET /api/config/timing` — Timing recommendations
- `PUT /api/config/timing` — Save timing recommendations
- `GET /health` — Health check

**Manuscript & Animations** (2):
- `GET /api/videos/{category}/{name}/manuscript` — Manuscript content
- `GET /api/videos/{category}/{name}/animations` — Available animations

### 3. React Frontend (`web/`)

**Build**: Vite + React 18 + TypeScript

**State management**:
- **TanStack Query (React Query)**: Server state — API fetching, caching, invalidation, optimistic updates
- **Zustand**: UI state — selected phase, active aspect tab, sidebar state

**Key components**:
```
web/src/
  api/           -- Typed API client layer
  components/
    layout/      -- AppLayout, Sidebar, Header
    phases/      -- PhaseOverview (dashboard), PhaseVideoList
    videos/      -- VideoCard, VideoDetail, VideoCreateForm
    forms/       -- DynamicForm, FieldRenderer, field-type components
    ai/          -- AIPanel, SuggestionsDisplay, OperationProgress
    publishing/  -- PublishPanel, UploadProgress
  hooks/         -- useApi, useSSE, useAspects, useVideoForm
  stores/        -- videoStore, uiStore
```

**Dynamic form rendering flow**:
1. Fetch aspect metadata: `GET /api/aspects` (cached)
2. For each aspect tab, render `DynamicForm` with field definitions
3. Map field types to renderers: string → text input, text → textarea, boolean → toggle, date → date picker
4. Show completion badges from `completionCriteria` metadata
5. PATCH changed fields to aspect-specific endpoint

**Serving**: Go server embeds built frontend via `//go:embed web/dist`. API under `/api/`, SPA fallback for everything else.

### 4. File Locking

Add `sync.RWMutex` in storage layer for index operations and per-video writes. Acceptable for single-user; avoids corruption from concurrent API requests.

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Dynamic form rendering complexity | High | Start with basic fields (string, bool); add complex types iteratively |
| Long-running operations (upload, AI) | Medium | Start synchronous; add SSE for operations > 5 seconds |
| OpenAPI spec drift | Medium | Validate API against `openapi.yaml` in tests |
| Frontend bundle size | Low | Vite tree-shaking; lazy-load analytics charts |
| Concurrent YAML writes | Medium | Add RWMutex in storage layer |
| Configuration init() pattern | Low | Use cobra subcommand; init() loads settings fine for server mode |

## Dependencies

### Internal
- Service layer (`internal/service/`)
- Aspect system (`internal/aspect/`)
- Video manager (`internal/video/`)
- AI modules (`internal/ai/`)
- Publishing modules (`internal/publishing/`)
- Platform modules (`internal/platform/`)

### New External Dependencies
- **Go**: `go-chi/cors` (CORS middleware)
- **Frontend**: react, react-dom, typescript, vite, @tanstack/react-query, zustand, tailwindcss (or similar)

## Out of Scope

- Multi-user authentication/authorization (single shared token is sufficient)
- Database migration (stays YAML)
- Real-time collaborative editing
- Mobile app
- Replacing the CLI (it continues to work)

## Milestones

- [x] **API Foundation + Video CRUD**: chi router, middleware, error handling, all video lifecycle endpoints, categories, health check. Tests passing.
- [x] **Aspect Metadata + Video Editing API**: Aspect metadata endpoints, 7 aspect-specific PATCH endpoints, progress endpoints, manuscript/animations endpoints. Tests passing.
- [x] **Bearer Token Authentication**: `API_TOKEN` env var middleware, constant-time comparison, `/health` always public, empty token = disabled. Tests passing.
- [x] **Frontend Foundation + Phase Dashboard**: Vite + React + TypeScript project, API client layer, app layout with sidebar, phase overview dashboard, video list per phase, video detail (read-only). Go server serves embedded frontend. Auth screen mandatory on load.
- [ ] **Git Sync for YAML Data**: Server clones/pulls a configured Git repo on startup, auto-commits and pushes on data mutations. Required because YAML data lives in a GitHub repo.
- [ ] **Dynamic Form Rendering + Video Editing UI**: DynamicForm component, all field renderers, aspect tab navigation, PATCH updates, completion badges, progress bars, video create/delete/archive actions.
- [ ] **AI Content Generation**: All 12 AI API endpoints, SSE infrastructure for long-running operations, frontend AI panel with suggestion display and "apply" action.
- [ ] **Publishing + Social Media**: YouTube upload, Hugo blog, shorts upload, dubbed upload, transcript fetch endpoints. Social media posting endpoints. Frontend publishing panel with upload progress.
- [ ] **Analytics Dashboard**: Video analytics, title analysis, timing recommendations, channel stats endpoints. Frontend analytics views.
- [ ] **AMA, Dubbing, Translation**: Remaining specialized feature endpoints and frontend panels. Full feature parity with CLI.
- [ ] **Containerization + Kubernetes Deployment**: Dockerfile (backend + frontend), Helm chart, GHA workflow to build/push images to ghcr.io on main push, PR-only test workflow. K8s Secret for `API_TOKEN`.
- [ ] **Documentation + Polish**: OpenAPI spec updated to cover all endpoints, README updated, build/deployment documentation.
- [ ] **Feature Tested & Validated**: End-to-end testing, 80% test coverage on API handlers, frontend tested with real data.

## Progress Log

### 2026-03-05
- PRD created
- GitHub issue #372 opened
- Architecture designed based on existing codebase analysis
- Discovered existing `openapi.yaml` with ~26 endpoints already specified
- Confirmed chi router already in `go.mod`
- Confirmed service layer, aspect system, and video manager are cleanly separated from CLI
- **Milestone 1 complete**: API Foundation + Video CRUD
  - Created `internal/api/` package: server, middleware (slog, CORS, RequestID, Recoverer), error helpers
  - Implemented 9 endpoints: health, video CRUD (GET/POST/PUT/DELETE), phases, categories, lightweight list
  - Added `serve` cobra subcommand with `--host`/`--port` flags
  - Fixed CLI required-flag validation via `PersistentPreRunE` (skips for subcommands)
  - Added `sync.RWMutex` to storage YAML struct for concurrent safety
  - Wired serve mode into `main.go`
  - 83.1% test coverage on API handlers, all tests pass with `-race`
  - Note: URL pattern uses `?category=X` query param instead of PRD's `/{category}/{name}` path params
- **Milestone 2 complete**: Aspect Metadata + Video Editing API
  - Created `internal/aspect/setter.go`: reflection-based `SetFieldValueByJSONPath` with type coercion (float64→int, JSON round-trip for slices)
  - Created `internal/api/handlers_aspects.go`: 4 aspect metadata endpoints (GET aspects, overview, fields, completion)
  - Created `internal/api/handlers_video_patch.go`: PATCH handler with aspect-scoped field validation, progress (overall + per-aspect), manuscript, animations endpoints
  - Updated `server.go` with `aspectService` and `filesystem` dependencies, registered 9 new routes
  - Added PATCH to CORS allowed methods
  - Updated `main.go` and test helper for new `NewServer` signature
  - 82.1% API test coverage, 84.8% aspect test coverage, all tests pass with `-race`
  - PATCH uses query params `?category=X&aspect=Y` (consistent with M1 pattern)
- **Design decisions**:
  - **Authentication added**: Bearer token via `API_TOKEN` env var. Empty = disabled (local dev). In K8s, delivered via Secret (not in git). `/health` always public for probes. Constant-time comparison. Replaces original "no auth" decision.
  - **Kubernetes deployment**: Will deploy via Helm chart to K8s. Container images pushed to ghcr.io (public). GHA workflow extended to build images on push to main, run tests on PRs only.
  - **Deferred packaging**: Dockerfile, Helm chart, and GHA image workflow deferred until after React frontend is built (both backend + frontend containerized together).
- **Milestone 3 complete**: Bearer Token Authentication
  - Created `internal/api/middleware_auth.go`: `bearerTokenAuth()` middleware with `crypto/subtle.ConstantTimeCompare`
  - Added `--api-token` flag to serve command with `API_TOKEN` env var fallback
  - `/health` always public (K8s probes), `/api/*` routes protected when token set
  - Auth disabled when token empty (backwards compatible, local dev friendly)
  - 9 tests (7 unit + 2 integration): valid/invalid/missing/malformed tokens, auth disabled, health public, CORS preflight
  - All tests pass with `-race`
- **Milestone 4 complete**: Frontend Foundation + Phase Dashboard
  - Scaffolded `web/` with Vite + React + TypeScript + Tailwind CSS v4
  - API client layer: `ApiError` class, Bearer token from localStorage, TanStack Query hooks (`usePhases`, `useVideosList`, `useVideo`, `useVideoProgress`)
  - Zustand UI store for sidebar/phase state
  - Layout: fixed sidebar (240px) with phase navigation + main content area
  - Phase Dashboard: responsive grid (4/2/1 cols) of phase cards with counts and color accents
  - Video List: table with name, category, date, progress bars; click navigates to detail
  - Video Detail: read-only fields grouped by aspect (Init, Work, Define, Edit, Publish, Post-Publish) with per-aspect and overall progress bars
  - Auth screen: mandatory on load when no token in localStorage, re-triggered on 401
  - Routing: `/` → Dashboard, `/phases/:id` → Video List, `/videos/:category/:name` → Detail
  - SPA fallback handler in Go: serves static files from embedded FS, falls back to index.html for client-side routes
  - `internal/frontend/embed.go` with `//go:embed all:dist` for single-binary deployment
  - Build targets: `make frontend-build`, `make build-local-full` / `just frontend-build`, `just build-full`
  - 16 frontend tests (Vitest + Testing Library + MSW): dashboard, video list, video detail, API client, Zustand store
  - 4 Go SPA handler tests: static files, client-side route fallback, API routes unaffected, nil FS
  - All tests pass
- **Design decisions**:
  - **Git sync needed**: YAML data lives in a GitHub repo. Server must clone/pull on startup and auto-commit/push on data mutations. New `git` section in `settings.yaml` for repo URL, branch, credentials. This is the next milestone before Dynamic Form Rendering.
  - **API token in settings.yaml**: Added `SettingsAPI` struct with `api.token` field as fallback. Precedence: `--api-token` flag > `API_TOKEN` env var > `settings.yaml`.
  - **Auth screen mandatory**: Frontend always shows token input on first load when no token in localStorage. Also re-shows on 401 responses.
