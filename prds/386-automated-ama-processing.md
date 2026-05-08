# PRD: Automated Daily AMA Stream Processing

**Issue**: #386
**Status**: In Progress
**Priority**: Medium
**Created**: 2026-03-13
**Last Updated**: 2026-05-09
**Depends On**: #379 (AMA Web UI — completed), #356 (AMA Video Enhancement — completed)

---

## Problem Statement

After each AMA livestream, the creator must manually:
1. Open the Web UI
2. Enter the YouTube Video ID
3. Click "Generate with AI" to fetch the transcript and generate content
4. Review the generated title, description, tags, and timecodes
5. Click "Apply to YouTube" to push changes

This is a repetitive task that's easy to forget, and the generated content is trusted enough to apply automatically.

## Proposed Solution

An **in-app scheduler** running inside the existing server process that:

1. Runs daily on a configurable schedule (default: 10:00 UTC every day)
2. Lists videos from the configured "Ask Me Anything" YouTube playlist
3. Picks the most recent video
4. Reads its YouTube description and checks for the timecodes marker (`▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬`)
5. **If the marker is present** → the video is already processed → silent exit (no notification)
6. **If the marker is absent** → fetches transcript, generates AI content, applies to YouTube, sends an email with the result (success or failure)
7. **If a pre-decision step fails** (e.g., playlist API unreachable) → sends an email so the operator knows the scheduler is broken

### Why Daily (Not Weekly + Retry)

A daily cadence with marker-based idempotency replaces the original weekly-cron + explicit-retry-counter design:

- **Transcript not ready today?** Tomorrow's run picks it up automatically.
- **Already processed (manually or by yesterday's run)?** Marker present → skip.
- **Multiple AMAs in a week?** The manual Web UI buttons remain available as the escape hatch for older videos.

No retry counter, no scheduled retry jobs, no state to persist across restarts.

### Why In-App Scheduler (Not K8s CronJob)

- Server is already running 24/7 in Kubernetes
- No additional K8s manifests, secrets, or RBAC to maintain
- Direct access to all existing services (YouTube API, AI, email)
- Configuration lives in `settings.yaml` alongside everything else

### Why Email-Only (No Slack)

- Reduces config surface (no Slack channel/token requirements)
- Email infrastructure already exists for upload notifications
- Inbox cadence is acceptable: ~1 email per week (the day after the AMA airs)

### User Journey

**Current (Manual)**:
1. Remember the AMA needs processing
2. Find the video ID from YouTube
3. Open the Web UI and run the AMA workflow
4. Verify changes applied

**After (Automated)**:
1. Receive an email the day after the AMA aired ("processed: <link>" or "failed: <error>")
2. Click the link to verify (success case) or intervene manually via the existing Web UI buttons (failure case)

## Technical Design

### Detection: Marker-Based Idempotency

The scheduler reads the YouTube video's current description and looks for the timecodes header constant (`timecodesHeader` in `internal/publishing/youtube_update.go`). Presence of this marker means the video has already been processed — by yesterday's scheduler run, by a manual click on the existing "Apply to YouTube" button, or by any other path. No external state is needed.

### Notification Rules

| Condition | Notify? |
|---|---|
| Marker present (already processed) | No — silent skip |
| Marker absent → processing succeeds | Yes — "AMA processed: <video link>" |
| Marker absent → processing fails (transcript, AI, or apply error) | Yes — "AMA processing failed: <error>" |
| Playlist fetch fails before the marker check | Yes — "Scheduler error: <error>" (soft exception so silent-broken state is detectable) |

### Configuration (`settings.yaml`)

```yaml
ama:
  enabled: true
  playlistId: "PLxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  schedule: "0 10 * * *"   # Cron expression: daily at 10:00 UTC
  emailTo: ""               # Email address for processing & error notifications
```

Environment variable overrides follow the existing pattern in `internal/configuration/`.

### Component Overview

1. **Playlist Service** (`internal/publishing/youtube_playlist.go`): New function to list videos from a YouTube playlist by playlist ID, ordered by publish date (most recent first). This is the only net-new YouTube API capability.
2. **AMA Job** (`internal/scheduler/ama_job.go`): Orchestration logic — fetch latest playlist video → read its description → check for marker → if absent, call `GetTranscript()` + `GenerateAMAContent()` + `UpdateAMAVideo()` → send email.
3. **Scheduler** (`internal/scheduler/scheduler.go`): Goroutine-based cron scheduler (e.g., `robfig/cron`) that triggers the AMA job on schedule. Graceful shutdown.
4. **Server Wiring** (`internal/api/server.go`): Start the scheduler when the server starts, stop on shutdown.
5. **Email Notification**: Reuses the existing email infrastructure used by upload notifications.

### Reused Existing Code

- `PublishingService.GetTranscript()` — fetch YouTube transcript
- `AIService.GenerateAMAContent()` — generate title/description/tags/timecodes
- `PublishingService.UpdateAMAVideo()` — apply to YouTube
- `timecodesHeader` constant from `internal/publishing/youtube_update.go` — marker for idempotency check
- Email service — reuse existing notification infrastructure
- The existing `POST /api/ama/generate` and `POST /api/ama/apply` endpoints remain available for manual triggering via the Web UI (no changes required)

### Key Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/configuration/cli.go` | MODIFY | Add `SettingsAMA` struct (`Enabled`, `PlaylistID`, `Schedule`, `EmailTo`) |
| `internal/publishing/youtube_playlist.go` | CREATE | YouTube playlist listing + read latest video description |
| `internal/scheduler/scheduler.go` | CREATE | Cron-based scheduler with graceful shutdown |
| `internal/scheduler/ama_job.go` | CREATE | AMA orchestration job |
| `internal/api/server.go` | MODIFY | Start/stop scheduler with the server lifecycle |
| `helm/youtube-automation/values.yaml` | MODIFY | Add AMA scheduler config values |

**Not changing**: `handlers_publish.go` (no new endpoint — manual trigger uses existing `/api/ama/generate` + `/api/ama/apply`), Web UI files (manual buttons already exist).

## Success Criteria

### Must Have
- [ ] Scheduler runs daily on a configurable cron schedule inside the server process
- [ ] Lists videos from the configured YouTube playlist
- [ ] Reads the latest video's description and detects the timecodes marker
- [ ] Skips silently when the marker is present
- [ ] When the marker is absent: generates AI content, applies to YouTube, sends success email with video link
- [ ] When processing fails after a marker-absent decision: sends failure email with error details
- [ ] When the playlist fetch (or any pre-decision step) fails: sends a scheduler-error email
- [ ] Configurable via `settings.yaml` (`enabled`, `playlistId`, `schedule`, `emailTo`)
- [ ] Helm chart exposes corresponding values
- [ ] Graceful shutdown when the server stops

### Nice to Have
- [ ] Process more than just the latest playlist entry (iterate until a processed one is found)
- [ ] Dry-run mode that generates but doesn't apply

## Milestones

- [x] **Milestone 1: YouTube Playlist Integration** — Add a function to list videos from a YouTube playlist by playlist ID, ordered by publish date. Add a function to read a video's current description (used for marker detection). Tests with mocked YouTube client. ✅ Implemented in `internal/publishing/youtube_playlist.go` (`ListPlaylistVideos`, `GetVideoDescription`) with 14 test cases, 84.62% coverage.

- [ ] **Milestone 2: AMA Job Orchestration** — Create the orchestration logic in `internal/scheduler/ama_job.go`: list playlist → read latest video description → check for `timecodesHeader` marker → if absent, run `GetTranscript()` → `GenerateAMAContent()` → `UpdateAMAVideo()`. Returns a typed result (skipped / processed / failed / scheduler-error). Full test coverage.

- [ ] **Milestone 3: Email Notifications** — Send email after each non-skipped outcome (processed, failed, scheduler-error). Reuses the existing email infrastructure. Subject + body templated by outcome type. Tests with mocked email client.

- [ ] **Milestone 4: In-App Scheduler** — Cron-based scheduler that starts with the server. Reads schedule from `settings.yaml`. Triggers the AMA job on schedule. Graceful shutdown. Tests for start/stop and scheduling.

- [ ] **Milestone 5: Configuration & Helm** — Add `SettingsAMA` struct (`Enabled`, `PlaylistID`, `Schedule`, `EmailTo`) with env-var overrides and startup validation. Update `helm/youtube-automation/values.yaml` with the corresponding values.

- [ ] **Milestone 6: Testing & Validation** — Full test coverage across all packages (≥80%). Manual end-to-end validation against a real AMA video.

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| YouTube API quota | Daily run = 1 playlist list + 1 video read + (only when processing) the existing generate/apply quota. Negligible. |
| Transcript not yet available (auto-captions can lag) | Daily idempotent re-run replaces explicit retry logic — tomorrow's run picks it up. |
| AI produces poor content | Same AI pipeline as the existing manual flow, which is already trusted. |
| Server restart loses scheduler state | Scheduler is stateless — marker check determines work on every run. |
| Stream is currently live when the scheduler runs | Marker absent + transcript unavailable → failure email → operator can wait and let tomorrow's run handle it. |
| Email infrastructure unavailable | Logged, scheduler continues. Same behavior as existing upload-notification path. |
| Silent-broken scheduler (e.g., playlist API outage that lasts days) | Soft-exception email on pre-decision failures restores "no email = working correctly" as a useful signal. |
