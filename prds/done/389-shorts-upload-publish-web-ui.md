# PRD #389: Shorts Upload & Publish in Web UI

**Status**: Complete
**Priority**: High
**Created**: 2026-03-15
**GitHub Issue**: #389

## Problem

The Web UI can generate shorts via AI and display them as editable fields, but the workflow dead-ends there. Users cannot upload short video files to Google Drive or publish them to YouTube from the Web UI. The backend endpoint (`POST /api/publish/youtube/{videoName}/shorts/{shortId}`) and React hook (`usePublishShort`) already exist but have no UI surface. This forces users to rely on external tools or CLI to complete the shorts lifecycle.

## Solution

Add per-short Drive upload and YouTube publish buttons to the Web UI, completing the end-to-end shorts workflow entirely within the browser: generate (AI) → upload video file (Drive) → publish (YouTube).

## User Journey

1. User generates shorts via the existing "Generate Shorts" AI button in the Definition tab
2. User reviews and edits generated shorts (title, text) — already works
3. User sets `scheduledDate` for each short — already works via form fields
4. **NEW**: User uploads a short video file for each short via a per-short "Upload to Drive" button
5. **NEW**: User clicks "Publish to YouTube" on each short to schedule it for publishing
6. Short's `youtubeId` field updates automatically after successful publish

## Technical Scope

### Backend Changes

#### 1. Add Drive file ID field to Short struct
**File**: `internal/storage/yaml.go`

Add `DriveFileID` field to the `Short` struct:
```go
DriveFileID string `yaml:"drive_file_id,omitempty" json:"drive_file_id,omitempty" ui:"auto"`
```

#### 2. Add Drive upload endpoint for shorts
**File**: `internal/api/handlers_drive.go`

New endpoint: `POST /api/drive/upload/short/{videoName}/{shortId}?category=X`

Follow the existing `handleDriveUploadVideo` pattern:
- Accept multipart/form-data with `video` field
- Find the short by ID within the video's shorts array
- Upload to Drive (reuse existing Drive service) in a `shorts/` subfolder under the video folder
- Store `DriveFileID` on the short
- Set `FilePath` to `drive://<driveFileId>` (same pattern as main video)

#### 3. Add Drive download endpoint for shorts
**File**: `internal/api/handlers_drive.go`

New endpoint: `GET /api/drive/download/short/{videoName}/{shortId}?category=X`

Follow `handleDriveDownloadVideo` pattern — stream the short file from Drive for download.

#### 4. Resolve short file path from Drive during publish
**File**: `internal/api/handlers_publish.go`

Update `handlePublishShort` to handle `drive://` file paths — download the short from Drive to a temp file before calling `UploadShort`, same as `handlePublishYouTube` does for the main video.

#### 5. Register new routes
**File**: `internal/api/server.go`

Add to the Drive route group:
```go
r.Post("/upload/short/{videoName}/{shortId}", s.handleDriveUploadShort)
r.Get("/download/short/{videoName}/{shortId}", s.handleDriveDownloadShort)
```

#### 6. Insert TODO markers into manuscript on shorts generation
**File**: `internal/api/handlers_ai.go`

Update `handleAIShorts` to call `manuscript.InsertShortMarkers(video.Gist, shorts)` after returning candidates, matching the CLI behavior in `internal/app/menu_shorts.go`. This inserts `TODO: Short (id: shortX) (start/end)` markers into the manuscript file so the text segments are visually annotated for reference.

### Frontend Changes

#### 6. Add API hooks for short Drive upload
**File**: `web/src/api/hooks.ts`

- `useUploadShortToDrive()` — multipart upload with progress, mirrors `useUploadVideoToDrive`
- `useDownloadShortFromDrive()` — optional, for download button

#### 7. Add per-short action buttons in ArrayInput
**File**: `web/src/components/forms/ArrayInput.tsx`

Extend the shorts array rendering to show per-item action buttons when `fieldName === 'shorts'`:
- **Upload to Drive** button per short (uses `useUploadShortToDrive`)
- **Publish to YouTube** button per short (uses existing `usePublishShort`)
- Show Drive file status (uploaded indicator, download link)
- Show YouTube ID when published (green badge, link to short)

Follow existing patterns:
- `VideoUploadInput` for upload UI with progress bar
- `PublishButton` for publish button with prerequisites check and status display
- ThumbnailVariants in `ArrayInput` already show per-item `FileUploadInput` — extend this pattern

#### 8. Prerequisites and validation
- "Upload to Drive" is always available (user selects a video file)
- "Publish to YouTube" requires: `filePath` or `driveFileId` set, AND `scheduledDate` set
- Show disabled state with tooltip when prerequisites are missing

## File Resolution

- **Server mode**: Short files uploaded to Drive are resolved via `drive://<id>` pattern, downloaded to temp during publish (same as main video)
- **CLI mode**: Short files use local filesystem paths directly

## Success Criteria

- User can upload a short video file to Google Drive per-short from the Web UI
- User can publish each short to YouTube from the Web UI
- Published shorts show their YouTube ID in the UI
- Upload shows progress indication
- Publish button is disabled with clear messaging when prerequisites (file + scheduled date) are missing
- No change to the workflow for non-short videos
- Existing `usePublishShort` hook is reused (not duplicated)

## Milestones

- [x] Add `DriveFileID` to Short struct and Drive upload/download endpoints for shorts
- [x] Update `handlePublishShort` to resolve `drive://` file paths (download from Drive before YouTube upload)
- [x] Insert TODO markers into manuscript when shorts are generated via Web UI (parity with CLI `manuscript.InsertShortMarkers`)
- [x] Add `useUploadShortToDrive` hook and per-short upload UI with progress in ArrayInput
- [x] Add per-short "Publish to YouTube" button using existing `usePublishShort` hook
- [x] Tests covering: Drive upload for shorts, Drive-to-YouTube publish flow, UI prerequisites validation
- [x] End-to-end validation: generate shorts → upload files → publish to YouTube entirely from Web UI

## Risks & Considerations

- **Large files**: Short videos are typically small (< 60 seconds) but could still be hundreds of MB; reuse the existing progress-tracking upload pattern
- **Drive folder structure**: Shorts should go in a `shorts/` subfolder under the video's Drive folder to keep things organized
- **Scheduling**: `scheduledDate` must be in the future for YouTube to accept it; consider frontend validation
- **ArrayInput complexity**: Adding per-item action buttons to ArrayInput needs care to avoid breaking other array fields (relatedVideos, thumbnailVariants, etc.)
