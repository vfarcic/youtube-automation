# PRD: Shorts Upload Automation

**Issue**: #360
**Status**: Not Started
**Priority**: Medium
**Created**: 2025-12-16
**Last Updated**: 2025-12-16
**Depends On**: PRD #339 (AI Shorts Identification) - MVP Complete

---

## Problem Statement

After using PRD #339's AI-powered Shorts identification feature, creators have successfully:
- Identified Short candidates via AI analysis
- Selected the best segments
- Had TODO markers inserted in manuscripts for editors
- Received extracted video files from editors

However, the upload process remains manual and repetitive:
- Each Short must be individually uploaded to YouTube Studio
- Scheduled publish dates must be manually calculated and set
- Main video links must be manually added to each description
- No tracking of which Shorts have been uploaded

This manual process is time-consuming and error-prone, especially when uploading 3-5 Shorts per video.

## Proposed Solution

Automate the Shorts upload workflow with:

1. **Scheduling Calculator**: Automatically calculate publish dates (1-day intervals with randomized times)
2. **Batch Upload CLI**: Single workflow to upload all pending Shorts with file paths
3. **Auto-Generated Descriptions**: Include link to main video automatically
4. **Upload Tracking**: Store YouTubeID after successful upload

### User Journey

**Current State (Manual)**:
1. Creator receives extracted Short video files from editor
2. Opens YouTube Studio
3. Uploads first Short manually
4. Sets scheduled publish date manually
5. Writes description, manually adds main video link
6. Repeats steps 3-5 for each Short (3-5 times)
7. No tracking in automation tool

**After (With This Feature)**:
1. Creator receives extracted Short video files from editor
2. Opens "Upload Shorts" in Publishing Details
3. System shows pending Shorts with calculated schedule dates
4. Creator provides file path for each Short (or skips)
5. System batch uploads all Shorts with:
   - Scheduled publish dates (1-day intervals, random times)
   - Auto-generated descriptions with main video link
6. YouTubeIDs stored in video YAML for tracking
7. Shorts go live on schedule automatically

## Success Criteria

### Must Have (MVP)
- [ ] Scheduling calculator computes publish dates (1-day intervals from main video date)
- [ ] Each Short gets randomized publish time (0-23 hours, 0-59 minutes)
- [ ] CLI "Upload Shorts" workflow shows pending Shorts
- [ ] User can provide file path for each pending Short
- [ ] User can skip individual Shorts during upload
- [ ] Shorts uploaded to YouTube with scheduled publish date
- [ ] Short descriptions include link to main video
- [ ] YouTubeID stored in YAML after successful upload
- [ ] Validation: file paths exist before starting upload

### Nice to Have (Future)
- [ ] API endpoints for programmatic upload
- [ ] Retry failed uploads
- [ ] Progress indicator during batch upload
- [ ] Custom description template per Short

## Technical Scope

### Core Components

#### 1. Scheduling Calculator (`internal/publishing/scheduler.go`)
```go
// CalculateShortsSchedule returns scheduled publish times for Shorts
// starting from the day after the main video's publish date
func CalculateShortsSchedule(mainVideoDate time.Time, count int) []time.Time {
    // Algorithm:
    // - Start from mainVideoDate + 1 day
    // - Each subsequent Short adds 1 day
    // - Randomize hour (0-23) and minute (0-59) for each
    // - Return ISO format timestamps
}
```

#### 2. CLI Upload Workflow (`internal/app/menu_shorts_upload.go`)
- New menu option in Publishing Details: "Upload Shorts"
- Display pending Shorts (those without YouTubeID)
- Show calculated scheduled date for each
- Form to input file path for each Short
- Skip button for individual Shorts
- Validate file paths before upload
- Progress feedback during upload

#### 3. YouTube Upload Integration (`internal/publishing/youtube.go`)
```go
// UploadShort uploads a single Short with scheduled publishing
func UploadShort(filePath string, short storage.Short, mainVideoURL string) (string, error) {
    // - Upload video file
    // - Set title from short.Title
    // - Set description: "Watch the full video: {mainVideoURL}\n\n#Shorts"
    // - Set scheduled publish time
    // - Return YouTube video ID
}

// UploadShorts batch uploads multiple Shorts
func UploadShorts(shorts []ShortUpload, mainVideoURL string) ([]UploadResult, error) {
    // - Iterate through shorts
    // - Upload each with UploadShort
    // - Collect results (success/failure per Short)
    // - Return results for YAML update
}
```

#### 4. Storage Updates
- No schema changes needed (Short struct already has YouTubeID field from PRD #339)
- Update YAML after each successful upload with YouTubeID

### Configuration
No new configuration needed. Hard-coded values from PRD #339 design decisions:
- 1-day intervals between Shorts
- Random times (full 24-hour range)

### Implementation Phases

**Phase 1: Scheduling Calculator**
- Implement `CalculateShortsSchedule()` function
- Unit tests for date calculations and randomization
- Integration with existing Short struct

**Phase 2: CLI Upload Workflow**
- "Upload Shorts" menu option
- File path input form
- Skip functionality
- Path validation

**Phase 3: YouTube Integration**
- `UploadShort()` function with scheduling
- Description generation with main video link
- YouTubeID storage after upload
- Error handling for failed uploads

**Phase 4: Testing & Polish**
- End-to-end testing with mock YouTube API
- Error recovery (partial upload failures)
- User feedback improvements

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| YouTube API rate limits | High | Low | Sequential uploads with delays; batch size limits |
| Invalid file paths | Medium | Medium | Validate all paths before starting upload |
| Partial upload failure | Medium | Medium | Track success/failure per Short; allow retry |
| Scheduled time conflicts | Low | Low | Randomization reduces likelihood |
| OAuth token expiry mid-batch | Medium | Low | Refresh token before batch; handle mid-upload refresh |

## Dependencies

### Internal
- PRD #339 complete (Shorts stored in YAML with TODO markers) - **Done**
- Existing YouTube upload functionality (`internal/publishing/youtube.go`)
- Video YAML storage (`internal/storage/yaml.go`)

### External
- YouTube Data API v3 (existing)
- No new external dependencies

## Out of Scope

- API endpoints (CLI-only for MVP)
- Custom scheduling intervals (hard-coded to 1 day)
- Custom description templates
- Thumbnail handling for Shorts
- Analytics tracking
- Retry UI for failed uploads (manual re-run instead)

## Validation Strategy

### Testing Approach
- Unit tests for scheduling calculator (date math, randomization bounds)
- Unit tests for description generation
- Integration tests with mock YouTube API
- End-to-end test: pending Shorts -> upload -> YouTubeID stored

### Manual Validation
- Test with real Short video files
- Verify scheduled dates appear correctly in YouTube Studio
- Confirm descriptions include main video link
- Validate Shorts format detected by YouTube

## Milestones

- [ ] **Scheduling Calculator Working**: Computes 1-day intervals with randomized times
- [ ] **CLI Upload Workflow Functional**: File path input, skip support, validation
- [ ] **YouTube Upload Integration**: Shorts upload with scheduling and descriptions
- [ ] **YouTubeID Tracking**: IDs stored in YAML after successful upload
- [ ] **Feature Tested & Validated**: End-to-end testing complete
- [ ] **Feature Launched**: Available in production

## Progress Log

### 2025-12-16
- PRD created
- GitHub issue #360 opened
- Scope defined based on pending items from PRD #339
- Waiting for validation of PRD #339 MVP before implementation

---

## Notes

- This PRD intentionally waits for real-world validation of PRD #339's AI identification feature
- Learnings from manual uploads will inform design decisions
- Keep scope minimal for MVP; enhance based on actual usage patterns
