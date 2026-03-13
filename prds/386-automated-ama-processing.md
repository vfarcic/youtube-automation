# PRD: Automated Weekly AMA Stream Processing

**Issue**: #386
**Status**: Not Started
**Priority**: Medium
**Created**: 2026-03-13
**Last Updated**: 2026-03-13
**Depends On**: #379 (AMA Web UI — completed)

---

## Problem Statement

After each AMA livestream, the creator must manually:
1. Open the CLI or Web UI
2. Enter the YouTube Video ID
3. Click "Generate with AI" to fetch the transcript and generate content
4. Review the generated title, description, tags, and timecodes
5. Click "Apply to YouTube" to push changes

This is a repetitive weekly task that's easy to forget, and the generated content is trusted enough to apply automatically.

## Proposed Solution

An **in-app scheduler** running inside the existing server process that:

1. Runs on a configurable schedule (default: every Friday at 10:00 UTC)
2. Lists videos from the "Ask Me Anything" YouTube playlist
3. Identifies the most recent video
4. Checks if it has already been processed (has a `manuscript/ama/` entry)
5. If unprocessed: fetches transcript, generates AI content, applies to YouTube, saves locally
6. Sends a Slack notification with the video link (success or failure)

### Why In-App Scheduler (Not K8s CronJob)

- Server is already running 24/7 in Kubernetes
- No additional K8s manifests, secrets, or RBAC to maintain
- Direct access to all existing services (YouTube API, AI, Slack, storage)
- Configuration lives in `settings.yaml` alongside everything else
- Can also be triggered manually via API endpoint

### User Journey

**Current (Manual)**:
1. Remember it's Friday and the AMA needs processing
2. Find the video ID from YouTube
3. Open CLI/Web UI and run the AMA workflow
4. Check YouTube to verify changes applied

**After (Automated)**:
1. Receive Slack notification on Friday with a link to the processed video
2. Click the link to verify (optional)

## Technical Design

### New Configuration (`settings.yaml`)

```yaml
ama:
  enabled: true
  playlistId: "PLxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
  schedule: "0 10 * * 5"  # Cron expression: Fridays at 10:00 UTC
  slackNotify: true
```

### Component Overview

1. **Playlist Service** (`internal/publishing/`): New function to list videos from a YouTube playlist, returning video IDs ordered by publish date
2. **AMA Scheduler** (`internal/scheduler/`): Goroutine-based scheduler using a cron library (e.g., `robfig/cron`) that triggers the auto-process workflow on schedule
3. **Auto-Process Logic** (`internal/api/` or `internal/service/`): Orchestrates the full workflow — find latest unprocessed video, generate content, apply, notify. Reuses existing `PublishingService.GetTranscript()`, `AIService.GenerateAMAContent()`, `PublishingService.UpdateAMAVideo()`, and Slack notification
4. **Manual Trigger Endpoint** (`POST /api/ama/auto-process`): Allows triggering the same workflow from the Web UI or manually via API
5. **Slack Notification**: On success, sends message with video link. On failure, sends error details

### Detection of Unprocessed Videos

Check whether a `manuscript/ama/` YAML file exists whose `videoId` field matches the latest playlist video. If no match exists, the video needs processing.

### Key Files to Create/Modify

| File | Action | Purpose |
|------|--------|---------|
| `internal/configuration/cli.go` | MODIFY | Add `SettingsAMA` struct |
| `internal/publishing/youtube_playlist.go` | CREATE | YouTube playlist listing |
| `internal/scheduler/scheduler.go` | CREATE | Cron-based scheduler |
| `internal/scheduler/ama_job.go` | CREATE | AMA auto-process job logic |
| `internal/api/server.go` | MODIFY | Start scheduler, add manual trigger route |
| `internal/api/handlers_publish.go` | MODIFY | Add `POST /api/ama/auto-process` handler |
| `web/src/pages/AskMeAnything.tsx` | MODIFY | Add "Run Now" button for manual trigger |
| `web/src/api/hooks.ts` | MODIFY | Add `useAMAAutoProcess` hook |

### Reused Existing Code

- `PublishingService.GetTranscript()` — fetch YouTube transcript
- `AIService.GenerateAMAContent()` — generate title/description/tags/timecodes
- `PublishingService.UpdateAMAVideo()` — apply to YouTube
- `menu_ama.go:saveAMAFiles()` — save to `manuscript/ama/` (extract to shared function)
- Slack client — send notification message

## Success Criteria

### Must Have
- [ ] Scheduler runs on configurable cron schedule inside the server process
- [ ] Lists videos from configured YouTube playlist
- [ ] Detects whether the latest video has already been processed
- [ ] Generates and applies AI content for unprocessed videos
- [ ] Sends Slack notification with video link on success
- [ ] Sends Slack notification with error details on failure
- [ ] Configurable via `settings.yaml` (enabled, playlistId, schedule)
- [ ] Manual trigger via `POST /api/ama/auto-process` endpoint

### Nice to Have
- [ ] "Run Now" button on the Web UI AMA page
- [ ] Process multiple unprocessed videos (not just the latest)
- [ ] Dry-run mode that generates but doesn't apply

## Milestones

- [ ] **Milestone 1: YouTube Playlist Integration** — Add function to list videos from a YouTube playlist by playlist ID, with tests. This is the only net-new YouTube API capability needed.

- [ ] **Milestone 2: Auto-Process Orchestration** — Extract `saveAMAFiles` to a shared function. Create the orchestration logic: find latest unprocessed video → fetch transcript → generate AI content → apply to YouTube → save locally. Expose as `POST /api/ama/auto-process` endpoint. Tests for the full workflow.

- [ ] **Milestone 3: Slack Notification** — Send a Slack message after processing with the video link (success) or error details (failure). Uses existing Slack client infrastructure.

- [ ] **Milestone 4: In-App Scheduler** — Add cron-based scheduler that starts with the server. Reads schedule from `settings.yaml`. Triggers the auto-process workflow on schedule. Graceful shutdown.

- [ ] **Milestone 5: Configuration & Settings** — Add `SettingsAMA` to configuration with `enabled`, `playlistId`, `schedule`, `slackNotify` fields. Environment variable overrides. Validate on startup.

- [ ] **Milestone 6: Web UI Integration** — Add "Run Now" button to the AMA page that calls the auto-process endpoint. Show last run status/result.

- [ ] **Milestone 7: Testing & Validation** — Full test coverage for playlist listing, orchestration, scheduler, and endpoint. Manual end-to-end validation.

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| YouTube API quota limits | Playlist listing is cheap (1 API call). Full workflow only runs for unprocessed videos. |
| Transcript not available yet | Auto-generated captions may take hours. If transcript fetch fails, skip and retry next scheduled run. |
| AI generation produces poor content | Same AI pipeline as CLI/Web UI which is already trusted. |
| Server restart loses scheduler state | Scheduler is stateless — on startup it checks if the latest video needs processing regardless of history. |
