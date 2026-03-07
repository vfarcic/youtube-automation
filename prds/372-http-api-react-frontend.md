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
- [x] HTTP API serves all video lifecycle operations (CRUD, phase, progress)
- [x] API exposes aspect metadata for dynamic form rendering
- [x] API serves AI content generation (titles, description, tags, tweets)
- [x] API serves publishing operations (YouTube upload, Hugo blog post)
- [x] React frontend renders phase dashboard with video counts
- [x] Frontend dynamically renders aspect-based editing forms from API metadata
- [x] Frontend shows progress tracking per aspect and overall
- [x] Frontend supports AI content generation with apply-to-field UX
- [x] API protected by bearer token auth (env var, disabled when unset)
- [x] Go server embeds and serves the built frontend (single binary deployment)
- [x] Helm chart deploys backend + frontend to Kubernetes
- [x] GHA builds and pushes container images to ghcr.io
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
- [x] **Git Sync for YAML Data**: Server clones/pulls a configured Git repo on startup, auto-commits and pushes on data mutations. Required because YAML data lives in a GitHub repo.
- [x] **Dynamic Form Rendering + Video Editing UI**: DynamicForm component, all field renderers, aspect tab navigation, PATCH updates, completion badges, progress bars, video create/delete actions.
- [x] **Array Field Type Support**: Add `"array"` field type to aspect system with item schema. Backend: new field type in aspect metadata with `itemFields` describing each sub-field. Frontend: `ArrayInput` component rendering items as sub-forms with add/remove. PATCH: verify reflection-based setter handles typed slices/maps correctly. Affected fields: `titles` ([]TitleVariant), `thumbnailVariants` ([]ThumbnailVariant), `shorts` ([]Short).
- [x] **AI Content Generation**: All 12 AI API endpoints, frontend inline AI generation with apply-to-field UX. SSE deferred to future milestone.
- [x] **Google Drive Thumbnail Upload**: New `internal/gdrive/` package with Drive API client, upload, and folder management (auto-creates per-video subfolders). Extract OAuth into shared `internal/auth/` package (reused by YouTube + Drive). Add `DriveFileID` field to `ThumbnailVariant` and `ThumbnailDriveFileID` to `DubbingInfo`. API endpoint `POST /api/drive/upload/thumbnail/{videoName}` (multipart, returns file ID). Frontend: `FileUploadInput` component with upload/replace UX, Drive ID display, sync warnings. Supports Shared Drives via `SupportsAllDrives`. Config: `gdrive.folderId` in settings.yaml. Dual mode: CLI keeps local paths, Web UI uses Drive file IDs. Tests passing.
- [x] **Aspect Reorganization + Action Buttons**: Reorganize fields across Definition and Post Production aspects to separate "define and request" from "receive deliverables." Move `Shorts`, `Members`, `RequestEdit` from Post Production to Definition. In the web UI, render `RequestThumbnail` and `RequestEdit` as action buttons (click sends email notification, sets bool to `true`, button shows "Requested" when done) instead of checkboxes — CLI keeps checkbox behavior unchanged. Backend: add `POST /api/actions/request-thumbnail/{videoName}` and `POST /api/actions/request-edit/{videoName}` endpoints that send emails and update the bool field. Frontend: new `ActionButton` component that calls the action endpoint, shows loading/success state, and is disabled when already `true`. Update `internal/aspect/mapping.go`: Definition gains `Shorts`, `Members`, `RequestEdit`; Post Production loses them (keeps `ThumbnailVariants`, `Timecodes`, `Slides`). Update tests and field counts. Tests passing.
- [x] **Google Drive Video Upload + Post Production Restructure**: Replace `Movie` (bool) in Post Production with `VideoFile` (string, local path for CLI) and `VideoDriveFileID` (string, Drive file ID for Web UI, `ui:"auto"` hidden from CLI). Add upload endpoint `POST /api/drive/upload/video/{videoName}` (multipart, reuses existing Drive infrastructure and per-video subfolders). Add download/preview endpoint `GET /api/drive/download/video/{videoName}` for review before publishing. Frontend: reuse `FileUploadInput` pattern from thumbnails with video-specific size/progress handling. In Publishing, replace `UploadVideo` (path string) with a trigger-only field — the publish action reads the video from Post Production (`VideoDriveFileID` first, falls back to `VideoFile` path). Update aspect mapping: Post Production has `ThumbnailVariants`, `VideoFile`, `Timecodes`, `Slides`; Publishing keeps `VideoId`/`HugoPath` and the upload trigger. Completion criteria: `VideoFile` uses `filled_only` (path or Drive ID must be set). Tests passing.
- [x] **Google Drive Thumbnail Consumption**: Add `WithThumbnailFile` temp-file helper and `ResolveThumbnail` fallback (DriveFileID first, Path for legacy). Update all 5 consumers to download from Drive on-demand: YouTube upload, dubbed upload, BlueSky posting, AI thumbnail analysis, Gemini localization. Temp files deleted immediately after use. Tests passing.
- [x] **Publishing + Social Media**: YouTube upload, Hugo blog, shorts upload, dubbed upload, transcript fetch endpoints. Social media posting endpoints. Frontend publish buttons (YouTube, Hugo) and social post buttons (BlueSky, Slack automated; LinkedIn, HN, DOT copy-paste). `PublishingService` interface with `DefaultPublishingService`. `UploadVideo` refactored to return errors. Tag sanitization for YouTube 500-char limit. `UploadVideo` field hidden from web UI. Tests passing.
- [x] **Hugo Post PR Workflow**: On a remote server, `Hugo.Post()` can't write to a local clone. Extend `SettingsHugo` with `repoURL`, `branch`, `token` fields. When `repoURL` is configured, clone the Hugo repo to a temp dir, create a branch, write the post, push, and create a GitHub PR via REST API. When only `path` is set, keep current local filesystem behavior (backward compatible for CLI). Extract shared `AuthenticatedURL` helper from `SyncManager`. `hugo.token` falls back to `GITHUB_TOKEN` env var. Tests passing.
- [ ] **AMA + Translation**: Remaining specialized feature endpoints and frontend panels. Full feature parity with CLI.
- [x] **Containerization + Kubernetes Deployment**: Dockerfile (multi-stage: node→golang→distroless, 16.5MB image), Helm chart (deployment, service, ingress, secret), GHA release workflow builds/pushes multi-platform images to ghcr.io, test workflow includes frontend tests. K8s Secret for `API_TOKEN`. Locally verified: image builds, health endpoint responds, frontend served.
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
- **Milestone 5 complete**: Git Sync for YAML Data
  - Created `internal/git/sync.go`: `SyncManager` with `CommandExecutor` interface, `InitialSync()` (clone or pull), `CommitAndPush()` (add → status → commit → pull --rebase → push), token injection into HTTPS URLs, mutex serialization, token redaction in error output
  - Created `internal/git/sync_test.go`: 10 tests (clone vs pull, skip-when-clean, push failure, token URL injection, output sanitization)
  - Updated `internal/filesystem/operations.go`: configurable `baseDir` field, `NewOperationsWithBaseDir()`, `GetBaseDir()` getter, replaced hardcoded `"manuscript"` references
  - Updated `internal/configuration/serve.go`: `--data-dir` flag with `DATA_DIR` env var fallback (default `./tmp`), `GetDataDir()` getter
  - Updated `internal/configuration/cli.go`: `SettingsGit` struct (`RepoURL`, `Branch`, `Token`), env var overrides (`GIT_REPO_URL`, `GIT_BRANCH`, `GIT_TOKEN`), default branch `"main"`
  - Updated `internal/service/video_service.go`: `onMutate` callback field, `SetOnMutate()`, `notifyMutation()` (logs errors, doesn't fail requests), called after `CreateVideo`/`UpdateVideo`/`DeleteVideo`/`ArchiveVideo`/`MoveVideo`; fixed `GetCategories()` hardcoded `"manuscript"` path
  - Updated `cmd/youtube-automation/main.go`: serve mode wiring — reads `dataDir`, creates `SyncManager` if git configured (fatal on clone/pull failure), `NewOperationsWithBaseDir(dataDir/manuscript)`, index path `dataDir/index.yaml`, registers `CommitAndPush` as onMutate callback
  - 6 new onMutate tests in `video_service_test.go`, 4 new filesystem tests
  - Verified with real data: cloned `vfarcic/devops-catalog`, served phase dashboard with correct counts
  - All tests pass (`go test ./...`)
- **Bug fix**: Frontend phase ID→name mapping was inverted in `web/src/lib/constants.ts` (e.g., id=0 showed "Ideas" instead of "Published"). Corrected to match backend `workflow/constants.go`. Updated mock test data in `web/src/test/handlers.ts`.
- **Milestone 6 complete**: Dynamic Form Rendering + Video Editing UI
  - **Bug fix**: `ASPECT_LABELS` in `web/src/lib/constants.ts` used camelCase keys (`initialDetails`) but backend returns kebab-case (`initial-details`). Fixed to match backend. Also fixed mock data in `web/src/test/handlers.ts`.
  - Added `patch()` to `web/src/api/client.ts`
  - Added 8 TypeScript interfaces in `web/src/api/types.ts`: `SelectOption`, `AspectFieldUIHints`, `AspectFieldValidationHints`, `FieldOptions`, `AspectField`, `AspectMetadata`, `AspectsResponse`, `CreateVideoRequest`
  - Added 4 hooks in `web/src/api/hooks.ts`: `useAspects()` (5min staleTime), `usePatchVideo()`, `useCreateVideo()`, `useDeleteVideo()` — all invalidate relevant queries on success
  - Created `web/src/components/forms/` with 8 components: `FieldLabel`, `TextInput`, `TextArea`, `Toggle`, `DateInput`, `NumberInput`, `SelectInput`, `DynamicForm` + barrel export
  - `DynamicForm`: renders fields from aspect metadata sorted by order, tracks dirty state via diff, sends only changed fields on save, supports dot-notation paths (e.g., `sponsorship.amount`), `key={aspect.key}` resets state when switching tabs
  - Rewrote `web/src/pages/VideoDetail.tsx`: tab bar for 7 aspects with progress badges (completed/total), active tab renders `DynamicForm`, save/error feedback, delete button with inline confirmation dialog
  - Created `web/src/components/CreateVideoDialog.tsx`: modal with name, category (required), date (optional), navigates to new video on success
  - Updated `web/src/pages/VideoList.tsx`: added "Create Video" button that opens dialog
  - Updated `web/src/test/handlers.ts`: fixed mock aspect keys to kebab-case, added `mockAspects` data, added MSW handlers for GET aspects, PATCH/POST/DELETE videos
  - Created `web/src/test/FieldRenderers.test.tsx` (9 tests): each input type renders correctly
  - Created `web/src/test/DynamicForm.test.tsx` (6 tests): field rendering, dirty tracking, save sends only changed fields, reset, dot-notation
  - Updated `web/src/test/VideoDetail.test.tsx` (9 tests): tabs from metadata, tab switching, save triggers PATCH, delete confirmation, delete navigates away
  - 37 frontend tests pass, all backend tests pass

### 2026-03-06
- **UX polish**: Added colored completion indicators to aspect tabs (green=complete, yellow=partial, gray=none) and field labels (red dot for incomplete fields). Helps quickly spot what's missing.
- **Bug fix**: Array/object field values (e.g., thumbnail variants) rendered as `[object Object]`. Fixed with JSON serialization for non-primitive values.
- **Frontend-evaluated completion**: Fields now evaluate `completionCriteria` (`filled_only`, `true_only`, `false_only`, `no_fixme`, `empty_or_filled`) client-side to show live completion status as users edit.
- Updated PRD checkboxes: marked 4 more Must Have items complete (API CRUD, aspect metadata, dynamic forms, progress tracking). 8/13 Must Have items done (62%).
- **Milestone 7 complete**: Array Field Type Support
  - Backend: Added `FieldTypeArray`/`FieldTypeMap` constants, `ItemField` struct, `ArrayFieldType`/`MapFieldType` implementations
  - Updated `determineFieldType` for `reflect.Slice` → `"array"`, `reflect.Map` → `"map"`
  - Added `generateItemFields` helper: introspects struct fields via reflection, reads JSON tags, skips `ui:"auto"`-tagged fields (e.g., auto-assigned `index`, analytics-populated `share`)
  - Added `reflect.Map` and `reflect.Struct` cases to `setFieldValue` (JSON round-trip pattern)
  - Frontend: `ArrayInput` component with compact single-field mode (inline inputs) and multi-field card mode (bordered sub-forms)
  - Frontend: `MapInput` component with key input + sub-form value cards
  - Updated `DynamicForm`: array/map dispatch, `JSON.stringify` deep comparison for dirty detection, array/map-aware `isFieldComplete`
  - Storage: Added `ui:"auto"` struct tag to `TitleVariant.Index`, `TitleVariant.Share`, `ThumbnailVariant.Index`, `ThumbnailVariant.Share` to exclude auto-managed fields from UI
  - 11 new backend tests, 13 new frontend tests, all existing tests updated and passing
  - Verified with real data: Titles renders as inline text inputs, Shorts/ThumbnailVariants as multi-field cards
- **Design decision**: Array/complex field type support
  - **Problem**: Fields like `titles`, `thumbnailVariants`, `shorts`, `dubbing` are arrays/maps of objects. The aspect system currently sends them as `type: "string"` or `type: "text"`, causing the frontend to render raw JSON strings (e.g., `[{"index":1,"text":"..."}]`). The CLI shows these as structured multi-field lists.
  - **Decision**: Add a new `"array"` field type to the backend aspect system with `itemFields` metadata describing each sub-field's name, type, and order. The frontend renders a generic `ArrayInput` component (list of sub-forms with add/remove). This keeps the frontend dumb about specific field names — any future array-of-objects field gets proper rendering automatically.
  - **Rationale**: Option 1 (frontend hardcodes known field names) would be faster but breaks the core architectural principle that "when a new field is added to the Go Video struct, it automatically appears in the frontend." Option 2 (backend metadata-driven) maintains that principle.
  - **Scope**: Backend aspect types + metadata generation, frontend `ArrayInput` component, PATCH handler verification for typed slices. `dubbing` (map[string]DubbingInfo) may need separate `"map"` type or special handling.
  - **Impact**: New milestone inserted before AI Content Generation. Affects `internal/aspect/` (types, field type registry, metadata builder) and `web/src/components/forms/` (new component).

- **Milestone 8 complete**: AI Content Generation
  - Backend: Created `internal/api/ai_service.go` (AIService interface + DefaultAIService), `internal/api/handlers_ai.go` (12 endpoints: titles, description, tags, tweets, description-tags, shorts, thumbnails, translate, 4x AMA), `internal/api/handlers_ai_test.go`
  - Updated `server.go`: added `aiService` dependency, registered `/api/ai/` route group with all 12 endpoints
  - Updated `cmd/youtube-automation/main.go`: wired `&api.DefaultAIService{}` into server
  - Frontend: Created `web/src/lib/aiFields.ts` (field-to-AI config map), `web/src/components/forms/AIGenerateButton.tsx` (inline generate button with field-specific UX: checkboxes for titles/shorts, radio for tweets, direct apply for strings)
  - Integrated into `DynamicForm.tsx`: added `category`/`videoName` props, renders `AIGenerateButton` next to AI-eligible fields
  - Updated `VideoDetail.tsx`: passes video identifiers to DynamicForm, removed separate "AI Assist" tab and `AIPanel` component
  - Added 12 AI mutation hooks in `web/src/api/hooks.ts`, 9 AI response types in `web/src/api/types.ts`
  - 7 new frontend tests in `web/src/test/AIGenerateButton.test.tsx`, MSW handlers for all AI endpoints
  - **Bug fix**: `filesystem.ResolvePath()` — manuscript paths stored in YAML are relative to data dir, but `os.ReadFile` in serve mode runs from a different CWD. Added `ResolvePath()` to prepend `rootDir` when set. Fixed `GetVideoManuscript` and `handleGetVideoAnimations`.
  - 3 new `ResolvePath` tests in `operations_test.go`
  - **Design decision**: Inline AI over separate tab — user feedback that switching between AI tab and form tab was clunky. Generate buttons render next to each AI-eligible field, results appear inline, Apply populates the field and dismisses results. Thumbnails, translation, and AMA remain API-only (not field-level).
  - Verified with real data: titles, description, tags generation working inline
  - 57 frontend tests pass, all backend tests pass

- **Design decision**: ThumbnailVariant `Type` field removed
  - **Problem**: `ThumbnailVariant` had a `Type` field ("original", "subtle", "bold") but it wasn't meaningfully used. The only consumer (`GetOriginalThumbnailPath`) had a fallback that made the Type check redundant.
  - **Decision**: Remove `Type` from the struct entirely. Thumbnails are identified by index/position only.
  - **Impact**: Simplified struct, simplified `GetOriginalThumbnailPath()`, CLI thumbnail form updated to index-based paths.

- **Design decision**: Short struct fields hidden from editing UI
  - **Problem**: `FilePath`, `ScheduledDate`, and `YouTubeID` fields on `Short` were showing in frontend editing forms, but they're populated programmatically by the publishing workflow (not manual editing). The CLI never shows them in editing context.
  - **Decision**: Added `ui:"auto"` tags to these three fields so they're excluded from aspect metadata and form rendering. Only `ID`, `Title`, `Text` appear in editing forms.
  - **Impact**: Cleaner editing UI, consistent with CLI behavior.

- **Design decision**: Thumbnails migrating from local paths to Google Drive file IDs
  - **Problem**: Thumbnails are stored as local filesystem paths. The user has symlinks to Google Drive locally, but the app will run remotely in a Kubernetes cluster where no mounted drive exists.
  - **Decision**: Store Google Drive file IDs instead of local paths. New `internal/gdrive/` package wraps Drive API. Central `WithThumbnailFile` helper downloads to temp file, executes operation, deletes temp file immediately. `ResolveThumbnail` checks `DriveFileID` first, falls back to `Path` for backward compatibility. OAuth refactored to shared `internal/auth/` package (reused by YouTube + Drive).
  - **Rationale**: File IDs are unambiguous, work anywhere, and avoid filesystem dependencies. Temp-file-only pattern prevents server disk from filling up.
  - **Scope**: Adds `DriveFileID` field to `ThumbnailVariant` and `DubbingInfo`. All 5 thumbnail consumers updated (YouTube upload, dubbed upload, BlueSky, AI analysis, Gemini localization). New milestone before Publishing + Social Media.
  - **Impact**: New `internal/gdrive/` and `internal/auth/` packages. Users must re-authenticate after adding Drive OAuth scope.

### How to Run for Manual Testing

The server reads `settings.yaml` from the **current working directory**. To run with real data from `devops-catalog`:

```bash
# Build with frontend embedded
make build-local-full

# Run from the devops-catalog directory (where settings.yaml lives)
cd ../devops-catalog && /path/to/youtube-automation/youtube-release serve
```

The server starts at `http://localhost:8080`. The frontend is embedded in the binary and served automatically.

### Post-Task Validation Protocol

**After completing each task**, always rebuild and restart the backend server so the user can manually validate:

```bash
# 1. Stop the running server (Ctrl+C or kill the process)
# 2. Rebuild with frontend embedded
make build-local-full
# 3. Restart from the devops-catalog directory
cd ../devops-catalog && /path/to/youtube-automation/youtube-release serve
```

**Why**: The Go binary embeds the frontend at build time, so backend changes require a full rebuild + restart. The frontend dev server (`npm run dev`) hot-reloads automatically, but for integrated testing the embedded build is used. Always restart the server after any backend or frontend change to ensure the user can validate against real data.

- **Milestone 9 complete**: Google Drive Thumbnail Upload
  - Created `internal/auth/oauth.go`: shared OAuth package extracted from `publishing/youtube.go`. Exports `OAuthConfig` (with `Scopes` field), `GetClient()`, `TokenFromFile()`, `SaveToken()`. Returns errors instead of `log.Fatalf`. 5 tests.
  - Updated `internal/publishing/youtube.go`: `getClientWithConfig()` delegates to `auth.GetClient()` with YouTube scopes. Removed 8 extracted helper functions.
  - Created `internal/gdrive/service.go`: `DriveService` interface with `UploadFile()` and `FindOrCreateFolder()`. Uses `SupportsAllDrives(true)` for Shared Drive support. Auto-creates per-video subfolders. Files named `thumbnail-N.ext`.
  - Updated `internal/storage/yaml.go`: Added `DriveFileID` (with `ui:"auto"`) to `ThumbnailVariant`, `ThumbnailDriveFileID` to `DubbingInfo`.
  - Updated `internal/configuration/cli.go`: `SettingsGDrive` struct with `credentialsFile`, `tokenFile`, `callbackPort`, `folderId`.
  - Created `internal/api/handlers_drive.go`: `POST /api/drive/upload/thumbnail/{videoName}?category=X&variantIndex=N`. Multipart upload, creates subfolder per video, returns `driveFileId`. 501 when Drive not configured. 8 tests.
  - Updated `internal/api/server.go`: `SetDriveService()` setter, `/api/drive/` route group.
  - Updated `cmd/youtube-automation/main.go`: wires Drive service from settings, separate token file (`gdrive-go.json`, port 8092), non-fatal on auth failure.
  - Frontend: `FileUploadInput.tsx` with upload/replace button, Drive ID display, sync warning support. `uploadFile()` in client for multipart FormData. `useUploadThumbnailToDrive()` hook. `ArrayInput` renders upload control for thumbnail variants. 4 tests.
  - Sync warning propagation: `VideoService.LastSyncError()` + `IsSyncConfigured()`. Responses include `syncWarning` field. UI shows green (synced), yellow (warning), red (error).
  - **UX polish**: Shorts candidate display shows full text (white) + rationale (gray italic).
  - Tested with real Google Drive: uploads to Shared Drive with per-video subfolders.
  - 61 frontend tests pass, all Go tests pass.

- **Design decision**: Dual-mode thumbnails (CLI local paths vs Web UI Drive)
  - **Problem**: Original milestone said "CLI updated to accept Drive file IDs" but CLI works fine with local paths (symlinked to Drive). Only the Web UI needs Drive file IDs for remote/K8s deployment.
  - **Decision**: Dual mode — CLI keeps using `Path` (local filesystem), Web UI writes `DriveFileID`. Both fields coexist on `ThumbnailVariant`. Future consumption milestone will add `ResolveThumbnail` that checks `DriveFileID` first, falls back to `Path`.
  - **Impact**: No CLI changes needed. Web UI uploads to Drive, CLI continues working unchanged.

- **Design decision**: Google Drive folder structure
  - **Problem**: Thumbnails uploaded to Drive root are hard to organize.
  - **Decision**: `gdrive.folderId` in settings.yaml sets root folder. Upload handler auto-creates a subfolder per video name (e.g., `web-ui-vs-agents-for-ai/thumbnail-0.png`). Uses `FindOrCreateFolder` to avoid duplicates.
  - **Impact**: Clean folder hierarchy, thumbnails grouped by video.

- **Design decision**: Video file moved from Publishing to Post Production
  - **Problem**: The `UploadVideo` field (a local file path) sits in the Publishing aspect, but the actual workflow has a gap between "video is ready" (editor delivers it) and "publish to YouTube" (user reviews and triggers). In a remote/K8s deployment, local paths don't work. The Publishing section conflates "where's the file" with "do the upload."
  - **Decision**: Replace `Movie` (bool, "is the movie done?") in Post Production with a proper video file field. Dual mode like thumbnails: `VideoFile` (local path for CLI) + `VideoDriveFileID` (Drive file ID for Web UI). Editor uploads through the web UI to Drive, user downloads/previews to review, then clicks publish. In Publishing, `UploadVideo` becomes a trigger action (not a path field) — it reads the video from Post Production. The `VideoFile` field uses `filled_only` completion (path or Drive ID must be set), maintaining the phase gate that `Movie` provided.
  - **Rationale**: Separates "material is ready" (Post Production) from "publish it" (Publishing). Matches the thumbnail pattern. The editor-to-reviewer handoff needs durable accessible storage, which Drive provides.
  - **Impact**: New milestone inserted after Google Drive Thumbnail Consumption. Changes `internal/storage/yaml.go` (struct fields), `internal/aspect/mapping.go` (field reassignment), `internal/api/handlers_drive.go` (new upload endpoint), Publishing handler (reads from Post Production instead of expecting a path). CLI continues working with local paths unchanged.

- **Design decision**: Aspect reorganization — Definition as "define and request", Post Production as "receive deliverables"
  - **Problem**: Post Production mixed "things to define" (Shorts, Members, RequestEdit) with "finished deliverables" (ThumbnailVariants, Timecodes, Movie, Slides). The boundary between Definition and Post Production was unclear.
  - **Decision**: Move `Shorts`, `Members`, `RequestEdit` from Post Production to Definition. Definition becomes the aspect where all content is defined and work is requested from collaborators. Post Production becomes purely about receiving finished material.
  - **New aspect layout**:
    - **Definition**: Titles, Description, Tags, DescriptionTags, Tweet, Animations, Shorts, Members, RequestThumbnail (button), RequestEdit (button)
    - **Post Production**: ThumbnailVariants, VideoFile/VideoDriveFileID, Timecodes, Slides
  - **Rationale**: Clean workflow — define everything, request work via buttons, then Post Production is where deliverables land. No mixing of intent with output.
  - **Impact**: `internal/aspect/mapping.go` field reassignment, field count changes in tests. No struct changes needed (fields stay on the same Video struct, just grouped differently in the UI).

- **Design decision**: Action buttons for RequestThumbnail and RequestEdit in web UI
  - **Problem**: `RequestThumbnail` and `RequestEdit` are booleans that trigger email notifications when flipped to `true` in the CLI. The CLI uses a checkbox because that's all the `huh` form framework supports. In the web UI, a checkbox is misleading — unchecking doesn't "un-request" the work.
  - **Decision**: In the web UI, render these as action buttons ("Request Thumbnail", "Request Edit"). Click sends email via a dedicated API endpoint (`POST /api/actions/request-thumbnail/{videoName}`, `POST /api/actions/request-edit/{videoName}`), sets the bool to `true`, and shows "Requested" (disabled). CLI behavior unchanged — still uses checkboxes with email side effect on toggle.
  - **Rationale**: Buttons communicate the irreversible, action-oriented nature of these fields. The bool still exists for completion tracking (`completion:"true_only"`).
  - **Impact**: New API endpoints, new `ActionButton` frontend component, field renderer dispatch updated to detect action fields. No struct changes.

- **Milestone 10 complete**: Aspect Reorganization + Action Buttons
  - Backend: Moved `Shorts`, `Members`, `RequestEdit` from Post Production to Definition in `internal/aspect/mapping.go`. Definition: 7→10 fields, Post Production: 7→4 fields.
  - Created `internal/api/handlers_actions.go`: `EmailService` interface, `SetEmailService()` setter, two action endpoints (`POST /api/actions/request-thumbnail/{videoName}`, `POST /api/actions/request-edit/{videoName}`). Idempotent (returns `alreadyRequested: true`), non-fatal email (sets field even if email fails, reports `emailError`), sync warning propagation.
  - Updated `internal/api/server.go`: added `emailService`/`emailSettings` fields, registered `/api/actions` route group.
  - Updated `cmd/youtube-automation/main.go`: wires `notification.Email` when email password configured.
  - Frontend: Created `web/src/components/forms/ActionButton.tsx` with `isActionField()` helper. States: enabled button → "Sending..." → "Requested" (disabled green badge). `useRequestThumbnail()` and `useRequestEdit()` hooks. `ActionResponse` type.
  - Updated `DynamicForm.tsx`: `isActionField()` check in boolean branch renders `ActionButton` instead of `Toggle`.
  - 10 backend tests (`handlers_actions_test.go`), 9 frontend tests (`ActionButton.test.tsx`), updated aspect mapping and service tests.
  - 70 frontend tests pass, all Go packages pass.
  - **Bug fix**: Progress counts derived dynamically from aspect mappings instead of hardcoded field lists. When fields moved between aspects (Definition 7→10, Post Production 7→4, Publishing +VideoId), the hardcoded `Calculate*Progress` functions in `manager.go` weren't updated. Fix: added `CalculateAspectProgress` to `CompletionService` that reads fields from aspect mappings, extracts values via `GetFieldValueByJSONPath`, evaluates completion via `IsFieldComplete`. Manager delegates via `ProgressCalculator` interface. Fixed `isFilledOnly` to handle slices. Special cases preserved: Titles in Definition (at least one non-empty), Analysis (per-title Share tracking), Shorts in Publishing (per-short YouTubeID tracking). `CalculateDubbingProgress` kept hardcoded (complex state machine). 11 files changed, all tests pass.

### 2026-03-07
- **Milestone 11 complete**: Google Drive Video Upload + Post Production Restructure
  - Backend: Added `POST /api/drive/upload/video/{videoName}` endpoint (multipart, reuses Drive infrastructure with per-video subfolders), `GET /api/drive/download/video/{videoName}` for download/preview. Replaced `Movie` (bool) with `VideoFile` (string, `ui:"label"`) and `VideoDriveFileID` (string, `ui:"auto"`). `VideoFile` stores `drive://<id>` for web uploads (satisfies `completion:"filled_only"`), local paths for CLI.
  - Added `FieldTypeLabel` to aspect system: new constant, `LabelFieldType` struct in `fieldtypes.go`, `ui:"label"` struct tag override in `generateFieldMapping()`. Frontend renders labels as read-only `<code>` elements — no hardcoded field name checks needed.
  - Frontend: Created `VideoUploadInput.tsx` with upload/replace/download buttons and progress bar. Thumbnails now upload-only (no path text input) — `ThumbnailVariant.Path` tagged `ui:"auto"`, `ArrayInput` handles zero visible `itemFields` with upload-only layout showing variant index and Drive file ID inline.
  - `FileUploadInput.tsx` cleaned up: removed Drive ID badge, added inline Download button (Google Drive direct link). `VideoUploadInput.tsx`: removed Drive ID badge, inline Download via backend proxy endpoint.
  - 9 new frontend tests (`VideoUploadInput.test.tsx`), updated `FileUploadInput.test.tsx`, updated aspect mapping/service/completion tests for `FieldTypeLabel` and `ui:"auto"` changes.
  - 79 frontend tests pass, all Go tests pass.

- **Milestone 12 complete**: Google Drive Thumbnail Consumption
  - Created `internal/thumbnail/resolve.go` with `ThumbnailRef` struct, `ResolveThumbnail()` (variants DriveFileID → Path → deprecated field), `ResolveDubbingThumbnail()` (DriveFileID → Path), and `WithThumbnailFile()` callback pattern (downloads Drive file to temp, calls fn, defers cleanup).
  - Consumer 1 (YouTube upload): Simplified `UploadThumbnail(video)` → `UploadThumbnail(videoId, thumbnailPath)`. Call site uses `ResolveThumbnail` + `WithThumbnailFile`.
  - Consumer 2 (Dubbed upload): `UploadDubbedVideo` takes `gdrive.DriveService` param. Thumbnail section uses `ResolveDubbingThumbnail` + `WithThumbnailFile`.
  - Consumer 3 (BlueSky): Call site uses `ResolveThumbnail` instead of deprecated `video.Thumbnail`.
  - Consumer 4 (Gemini localization): `LocalizeThumbnail` takes `gdrive.DriveService` param, uses `ResolveThumbnail` + `WithThumbnailFile` to support Drive-hosted originals.
  - Consumer 5 (AI analysis): `AIThumbnailsRequest` accepts `driveFileId` field. Handler uses `WithThumbnailFile` with server's `driveService`.
  - 18 new tests in `resolve_test.go`, updated existing tests in `service_test.go`, `youtube_dubbed_test.go`, `handlers_ai_test.go`. All Go and frontend tests pass.

- **Design decision**: Hugo Post PR workflow for remote servers
  - **Problem**: `Hugo.Post()` writes directly to a local filesystem path (`../devopstoolkit-live/`). On a remote server / Kubernetes cluster, there's no local clone of the Hugo repo available.
  - **Decision**: Extend `SettingsHugo` with `repoURL`, `branch`, `token`. When `repoURL` is configured, clone to temp dir → create branch → write post → push → create PR via GitHub REST API → cleanup. When only `path` is set, keep current local filesystem behavior. Extract `AuthenticatedURL` from `SyncManager` as shared utility in `internal/git/auth.go`.
  - **Rationale**: PR workflow lets the user review the Hugo post before it goes live (merge = publish). Local mode preserved for CLI backward compatibility. `hugo.token` field falls back to `GITHUB_TOKEN` env var (consistent with rest of infrastructure).
  - **Impact**: `internal/configuration/cli.go` (extend `SettingsHugo`), `internal/publishing/hugo.go` (refactor `Hugo` struct + PR workflow), `internal/git/auth.go` (shared helper), `cmd/youtube-automation/main.go` (pass config to `NewHugo`). No frontend changes needed — `Post()` returns a path (local) or PR URL (remote), stored in `video.HugoPath` either way.

- **Bug fix**: Analysis aspect title rendering
  - **Problem**: The Analysis tab reused the same `Titles` field metadata as Definition. Title texts rendered as editable inputs (should be read-only labels), and share percentages were hidden (`ui:"auto"` on `TitleVariant.Share`).
  - **Fix**: Added aspect-specific item field overrides in `GetVideoAspectMappings()`. When generating the Analysis aspect's `Titles` field, item fields are overridden to: `text` as `FieldTypeLabel` (read-only) + `share` as `FieldTypeNumber` (visible). Frontend `ArrayInput` updated to handle `label` sub-field type, rendering as `<code>` element. New test `TestAnalysisTitlesItemFieldOverrides`. All tests pass.

- **Dubbing feature removed**: Removed the entire dubbing subsystem (6,020 lines deleted across 34 files). Deleted `internal/dubbing/` directory (ElevenLabs client, compression, types), `internal/publishing/youtube_dubbed.go`, `DubbingInfo` struct from storage, ElevenLabs/SpanishChannel configuration, `CalculateDubbingProgress()`, `ResolveDubbingThumbnail()`, Spanish OAuth functions, API handler/route for dubbed uploads, ~648 lines of dubbing CLI code, and all corresponding tests. Frontend types, hooks, and test fixtures updated. Dubbing has been split into separate PRDs (#374-378 for analytics). All Go and frontend tests pass.
- **PRD milestones updated**: Removed "Analytics Dashboard" milestone (analytics work split into PRDs #374-378). Updated "AMA, Dubbing, Translation" milestone to "AMA + Translation".
