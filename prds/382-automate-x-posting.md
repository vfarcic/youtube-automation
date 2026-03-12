# PRD: Automate X (Twitter) Posting

**Issue**: #382
**Status**: Not Started
**Priority**: Medium
**Created**: 2026-03-12

---

## Problem Statement

X (Twitter) posting is entirely manual. After publishing a video, the user must copy the tweet text and post it to X themselves. Every other major social platform in the system either has full automated posting (BlueSky, Slack) or a structured copy-paste flow (LinkedIn, HackerNews, DevOpsToolkit). X has neither — it only appears as a reminder in calendar event descriptions.

This creates friction in the post-publish workflow and is the only social platform without any integration.

## Proposed Solution

Add automated X posting following the same pattern as the existing BlueSky integration:

1. **Platform module** (`internal/platform/x/`) — OAuth authentication, tweet creation with media attachment, URL card support
2. **Video struct field** — `XPosted bool` to track posting status
3. **Service/API wiring** — `PostX()` method on `PublishingService`, new `"x"` case in the social posting API handler
4. **CLI integration** — X posting option in the post-publish phase menu
5. **Configuration** — X API credentials in `settings.yaml` and environment variables

## Success Criteria

### Must Have
- [ ] Automated posting to X with tweet text and YouTube link
- [ ] Image/thumbnail attachment support (like BlueSky)
- [ ] `XPosted` field tracked in video YAML
- [ ] Available via both CLI menu and HTTP API (`POST /api/publish/{name}/social/x`)
- [ ] Configuration via `settings.yaml` and environment variables for secrets
- [ ] Error handling with clear messages for auth failures, rate limits, etc.
- [ ] 80% test coverage on new code

### Nice to Have
- [ ] URL card preview (Twitter Cards) — may happen automatically via YouTube URL
- [ ] Character count validation (280 chars for X vs 300 for BlueSky)
- [ ] Rate limit awareness and retry logic

## Technical Scope

### Reference Implementation: BlueSky

The BlueSky integration serves as the direct template. Key files:

| Component | BlueSky Location | X Equivalent |
|-----------|-----------------|--------------|
| Platform module | `internal/platform/bluesky/bluesky.go` | `internal/platform/x/x.go` |
| Config struct | `configuration.SettingsBluesky` | `configuration.SettingsX` |
| Service interface | `PublishingService.PostBlueSky()` | `PublishingService.PostX()` |
| API handler | `handlers_publish.go` case `"bluesky"` | case `"x"` |
| Video field | `BlueSkyPosted bool` | `XPosted bool` |
| Env var | `BLUESKY_PASSWORD` | `X_API_KEY_SECRET` (or similar) |

### X API Requirements

- **API Version**: X API v2 (https://developer.x.com)
- **Auth**: OAuth 1.0a (User Context) for posting on behalf of a user
- **Endpoints needed**:
  - `POST /2/tweets` — create a tweet
  - `POST /1.1/media/upload.json` — upload thumbnail image (media upload is still v1.1)
- **Rate limits**: Free tier allows 1,500 tweets/month (sufficient for this use case)
- **Developer account**: Required — must register app at developer.x.com

### Configuration Addition

```yaml
# settings.yaml
x:
  apiKey: ""           # Consumer API key
  apiKeySecret: ""     # Consumer API key secret
  accessToken: ""      # User access token
  accessTokenSecret: "" # User access token secret
```

Environment variable overrides: `X_API_KEY`, `X_API_KEY_SECRET`, `X_ACCESS_TOKEN`, `X_ACCESS_TOKEN_SECRET`

### Files to Create/Modify

**New files:**
- `internal/platform/x/x.go` — Core X posting logic (auth, media upload, tweet creation)
- `internal/platform/x/x_test.go` — Tests with mocked HTTP

**Modified files:**
- `internal/storage/yaml.go` — Add `XPosted bool` field to `Video` struct
- `internal/configuration/cli.go` — Add `SettingsX` struct and wire into `Settings`
- `internal/api/publishing_service.go` — Add `PostX()` to interface and default implementation
- `internal/api/handlers_publish.go` — Add `"x"` case in social posting switch
- `internal/api/handlers_publish_test.go` — Add test cases for X posting
- `internal/app/menu_phase_editor.go` — Add X posting option in CLI post-publish menu
- `internal/constants/fields.go` — Add `FieldTitleXPosted`
- `internal/aspect/` — Wire `XPosted` into phase completion tracking

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| X API access approval delays | Blocks implementation | Apply for developer account early; code can be written against API docs before credentials arrive |
| X API pricing changes | Could increase cost | Free tier (1,500 tweets/month) is sufficient; monitor announcements |
| OAuth 1.0a complexity | More complex than BlueSky's simple auth | Use well-maintained Go library (e.g., `dghubble/oauth1`) |
| Media upload uses v1.1 API | Different auth flow than v2 tweets | Both use same OAuth 1.0a credentials; just different endpoints |

## Milestones

- [ ] **M1: X developer account and API credentials obtained** — Register app, get API keys, verify access level supports posting
- [ ] **M2: Core platform module with authentication** — `internal/platform/x/x.go` with OAuth 1.0a auth, config validation, and basic tweet creation (text only)
- [ ] **M3: Media upload and thumbnail attachment** — Image upload via v1.1 media endpoint, attach thumbnail to tweets
- [ ] **M4: Video struct and configuration wiring** — `XPosted` field, `SettingsX` config, environment variable overrides
- [ ] **M5: API and CLI integration** — `PostX()` on `PublishingService`, API handler case, CLI menu option, phase completion tracking
- [ ] **M6: Tests and validation** — Unit tests for platform module, API handler tests, integration test with real credentials, 80% coverage verified

## Dependencies

- X developer account with posting permissions (Free tier sufficient)
- OAuth 1.0a library for Go (recommend `dghubble/oauth1`)
