# PRD: AI-Powered YouTube Shorts Candidate Identification from Manuscripts

**Issue**: #339
**Status**: In Progress
**Priority**: Medium
**Created**: 2025-11-11
**Last Updated**: 2025-12-16

---

## Problem Statement

Currently, identifying segments from a manuscript that would make good YouTube Shorts requires manual review and subjective judgment. Content creators must:
- Manually read through entire manuscripts looking for self-contained segments
- Estimate if segments fit within the 60-second Short format
- Judge which moments are impactful enough to work as standalone content
- Track which segments they've already extracted to avoid duplication

This manual process is time-consuming, inconsistent, and risks missing valuable Short opportunities that could drive traffic to main videos.

## Proposed Solution

Add AI-powered analysis to automatically identify manuscript segments suitable for YouTube Shorts. The feature will:

1. **Analyze manuscripts** using AI to identify self-contained, high-impact segments
2. **Estimate duration** based on configurable word count limits
3. **Present 10 candidates** for creator to review and select from
4. **Insert TODO markers** in manuscript with unique IDs for editor extraction instructions
5. **Calculate scheduled publish dates** with 1-day intervals and randomized times
6. **Batch upload Shorts** with scheduled publishing (not immediate)
7. **Link Shorts to main videos** in descriptions for cross-promotion

### User Journey

**Before (Current State)**:
1. Editor sends completed video to content creator
2. Creator manually reviews manuscript looking for Short candidates
3. Creator messages editor with timestamps or descriptions of segments
4. Editor extracts and uploads Shorts manually
5. No systematic tracking of which Shorts came from which videos

**After (With This Feature)**:
1. After manuscript is complete, creator opens "Publishing Details"
2. Selects "Analyze for Shorts" option
3. AI presents 10 candidate segments with rationale
4. Creator reviews suggestions and selects which to pursue (e.g., selects 4 out of 10)
5. System inserts TODO markers in manuscript: `TODO: Short (id: short1) (start)` / `(end)`
6. System calculates scheduled publish dates (1 day intervals, randomized times)
7. System stores selected Shorts metadata in video YAML
8. Creator shares manuscript with editor who extracts segments using TODO markers
9. Editor provides extracted video files back to creator
10. Creator uses "Upload Shorts" option (batch upload all at once)
11. Provides file paths for each Short during upload
12. System uploads all Shorts with scheduled publish dates
13. Short descriptions automatically link to main video
14. Shorts go live on schedule automatically

## Success Criteria

### Must Have (MVP)
- [x] AI successfully identifies 10 Short candidates from a manuscript
- [x] Candidates meet word count limit (configurable in settings.yaml)
- [x] System presents candidates with extracted text and AI rationale (during selection only)
- [x] Selected segments get TODO markers inserted in manuscript with unique IDs
- [x] TODO markers include (start) and (end) tags for clear boundaries
- [ ] Scheduled publish dates calculated (1 day intervals, randomized times)
- [x] Selected Shorts metadata stored in video YAML
- [ ] "Upload Shorts" option accepts batch upload with file paths
- [ ] Shorts uploaded with scheduled publish dates (not immediate)
- [ ] Short descriptions include link to main video
- [ ] System tracks all Shorts linked to main video via YouTubeID

### Nice to Have (Future)
- [ ] AI learns from user selections to improve recommendations over time
- [ ] Support for different Short types (tutorial, insight, controversial take)
- [ ] Automated thumbnail generation suggestions
- [ ] Analytics tracking: which Shorts drive most traffic to main videos
- [ ] Batch analysis across multiple videos to find patterns

## Technical Scope

### Core Components

#### 1. AI Analysis Module (`internal/ai/shorts.go`)
- New AI module for analyzing manuscripts
- Prompt engineering to identify self-contained, high-impact segments
- Word count validation against configurable limit
- Return structured results with text and rationale (rationale not stored, only shown during selection)

#### 2. Storage Changes (`internal/storage/yaml.go`)
- Add `Shorts` field to `Video` struct:
  ```go
  type Video struct {
      // ... existing fields
      Shorts []Short `json:"shorts,omitempty" yaml:"shorts,omitempty"`
  }

  type Short struct {
      ID            string   `json:"id" yaml:"id"`                       // Unique identifier (short1, short2, etc.)
      Title         string   `json:"title" yaml:"title"`                 // Short title
      Text          string   `json:"text" yaml:"text"`                   // Extracted manuscript segment
      ScheduledDate string   `json:"scheduled_date" yaml:"scheduled_date"` // ISO format publish timestamp
      YouTubeID     string   `json:"youtube_id,omitempty" yaml:"youtube_id,omitempty"` // Short's YouTube video ID
  }
  ```

**Key Design Decisions:**
- **No LineStart/LineEnd**: Manuscript TODO markers with IDs are source of truth for segment boundaries
- **No Status field**: Derived from data (no YouTubeID = pending, has YouTubeID + future date = uploaded, past date = published)
- **No Rationale field**: Only shown during selection, not persisted
- **No FilePath field**: File paths are transient CLI input during upload, not stored
- **YouTubeID**: Refers to the Short's video ID, not the main video (main video is implicit from YAML location)

#### 3. Manuscript Modifier (`internal/manuscript/shorts.go` - new module)
- Insert TODO markers in manuscript at identified segment locations
- Marker format: `TODO: Short (id: short1) - "Title" (start)` and `TODO: Short (id: short1) (end)`
- Parse manuscript to extract segment text by ID when needed
- Handles markdown formatting preservation

#### 4. Scheduling Calculator (`internal/publishing/scheduler.go` - new or extend existing)
- Calculate publish dates for Shorts based on main video publish date
- Algorithm: 1 day intervals with randomized times (0-23 hours, 0-59 minutes)
- Returns ISO format timestamps for YouTube API scheduling

#### 5. Configuration (`settings.yaml`)
- Add `shorts_max_words` setting (default: 150)
- Add `shorts_candidate_count` setting (default: 10)

**Removed Configuration** (design decision: hard-code for simplicity):
- ~~`shorts_publish_start_hour`~~ - Random times cover full 24 hours
- ~~`shorts_publish_end_hour`~~ - No time restrictions
- ~~`shorts_interval_days`~~ - Hard-coded to 1 day

#### 6. CLI Interface (`internal/app/`)
- Add "Analyze for Shorts" option to Publishing Details menu
- Form for reviewing 10 AI suggestions and selecting candidates (multi-select)
- Add "Upload Shorts" option to Publishing Details menu (batch upload)
- Form for providing file paths for each pending Short
- Display scheduled publish dates for each Short
- Allow skipping individual Shorts during upload

#### 7. API Interface (`internal/api/`)
- `POST /videos/{id}/analyze-shorts` - Trigger analysis, returns 10 candidates
- `POST /videos/{id}/shorts/select` - Select Short candidates, inserts TODO markers
- `GET /videos/{id}/shorts` - Get all Shorts for a video
- `POST /videos/{id}/shorts/upload` - Batch upload Shorts with file paths

#### 8. Publishing Integration (`internal/publishing/youtube.go`)
- New `UploadShorts()` function for batch upload
- Each Short uploaded with scheduled publish date (not immediate)
- Video description includes link to main video: "Watch the full video: [URL]\n\n#Shorts"
- Handle YouTube Shorts-specific metadata (aspect ratio auto-detected by YouTube)
- Store YouTubeID in YAML after successful upload

### Implementation Phases

**Phase 1: Analysis & Manuscript Modification** (Week 1)
- AI module for manuscript analysis (identify 10 candidates)
- Manuscript modifier module (insert TODO markers with IDs)
- Storage schema updates (simplified Short struct)
- Configuration settings

**Phase 2: Scheduling & CLI Selection** (Week 2)
- Scheduling calculator (1 day intervals, randomized times)
- Publishing Details "Analyze for Shorts" option
- Review and selection form (choose from 10 candidates)
- TODO markers inserted in manuscript after selection

**Phase 3: YouTube Integration** (Week 3)
- Batch upload function with scheduled publishing
- Description linking to main video
- Metadata handling (Shorts format detection)

**Phase 4: CLI Upload & API** (Week 4)
- "Upload Shorts" CLI workflow (file path input)
- REST API endpoints
- API documentation

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| AI identifies poor Short candidates | High | Medium | Present 10 candidates (up from 5); user selects best ones |
| Word count estimation inaccurate | Medium | Medium | Make limit configurable; conservative by default |
| YouTube API limits on Shorts | High | Low | Use standard video upload API (Shorts auto-detected by ratio) |
| Editor can't find segments in video | High | Medium | TODO markers with (start)/(end) tags; store segment text in YAML |
| TODO markers get removed/modified | Medium | Low | Manuscript is shared as-is; editor follows instructions |
| Scheduled publish times conflict | Low | Low | Out of scope for MVP; randomization reduces likelihood |
| File paths invalid during upload | Medium | Medium | CLI validates paths exist before upload; allow skipping |

## Dependencies

### Internal
- Existing AI integration (`internal/ai/`)
- YouTube upload functionality (`internal/publishing/youtube.go`)
- Video YAML storage (`internal/storage/yaml.go`)
- Publishing Details workflow

### External
- Azure OpenAI API (existing)
- YouTube Data API v3 (existing)
- No new external dependencies required

## Resolved Questions (Design Decisions)

1. **Line numbers vs TODO markers**: How should editor find segments?
   - **Decision (2025-11-28)**: Use TODO markers with unique IDs in manuscript, not line numbers
   - **Rationale**: Line numbers become stale if manuscript edited; markers remain valid

2. **How many candidates to present**: 5 or more?
   - **Decision (2025-11-28)**: Present 10 candidates for larger selection pool
   - **Rationale**: More choices increases likelihood of finding high-quality Shorts

3. **Store rationale in YAML?**: Should AI explanation be persisted?
   - **Decision (2025-11-28)**: No, only show during selection phase
   - **Rationale**: No long-term value; adds clutter to YAML

4. **Status tracking**: How to track Short lifecycle?
   - **Decision (2025-11-28)**: Derive from data, don't store separate field
   - **Rationale**: Redundant; can compute from YouTubeID and ScheduledDate

5. **Upload workflow**: Individual or batch?
   - **Decision (2025-11-28)**: Batch upload all Shorts at once with scheduled publish
   - **Rationale**: Matches existing long-form workflow; more efficient

6. **Scheduling configuration**: How much control?
   - **Decision (2025-11-28)**: Hard-code 1 day intervals, randomized times (no config)
   - **Rationale**: Simple is better; can enhance later if needed

7. **Pinned comments**: Should we add CTAs as comments?
   - **Decision (2025-11-28)**: No pinned comments
   - **Rationale**: Users swipe through Shorts quickly without reading comments

## Open Questions

1. **Thumbnail handling**: Should we support custom Short thumbnails or use auto-generated?
   - *Decision pending user feedback*

2. **Multi-video analysis**: Should we support "find Shorts across all videos in category"?
   - *Future enhancement, not MVP*

3. **Analytics integration**: Should we track Short → main video conversion?
   - *Future enhancement, requires YouTube Analytics API*

## Out of Scope

- Automated video editing or segment extraction (editor does this manually)
- Automated Short thumbnail generation
- YouTube Shorts-specific analytics (deferred to future PRD)
- Cross-platform Short publishing (TikTok, Instagram Reels)
- Pinned comments on Shorts (users don't read them)
- A/B testing different Shorts from same video
- Intelligent scheduling conflict resolution (randomization is sufficient)
- Configurable scheduling parameters (hard-coded intervals work fine)

## Documentation Impact

### New Documentation
- How to use Shorts analysis feature
- TODO marker format specification
- Best practices for selecting Short candidates (from 10 options)
- Editor workflow for extracting Shorts using TODO markers
- Batch upload workflow with file paths

### Updated Documentation
- Publishing workflow documentation (add Shorts analysis and upload steps)
- Video YAML schema reference (new Short struct)
- Manuscript format guide (TODO markers for Shorts)
- Settings.yaml configuration options (shorts_candidate_count: 10)
- API documentation (if API mode enabled)

## Validation Strategy

### Testing Approach
- Unit tests for AI analysis module (10 candidates generation)
- Unit tests for manuscript modifier (TODO marker insertion/parsing)
- Unit tests for scheduling calculator (randomization, 1-day intervals)
- Integration tests for YAML storage of Shorts (simplified struct)
- Mock YouTube API for scheduled upload testing
- End-to-end test: analyze → select → TODO markers → upload → verify link

### Manual Validation
- Test with diverse manuscript types (tutorial, discussion, demo)
- Verify TODO markers inserted correctly with (start)/(end) tags
- Verify word count estimates match actual video timing
- Confirm editor can locate segments using TODO markers
- Test batch upload with multiple file paths
- Validate scheduled publish dates (1 day intervals, random times)
- Confirm YouTube properly recognizes Shorts format
- Verify Short descriptions link to main video

### Success Metrics
- AI identifies at least 10 viable candidates per manuscript
- Creator selects 3-5 candidates on average from 10 options
- TODO markers inserted correctly 100% of time
- 80%+ of user-selected Shorts successfully upload
- Editors can locate segments within 1 minute using TODO markers
- Scheduled publish dates accurate (1 day intervals)
- Short descriptions correctly link to main videos

## Milestones

- [x] **AI Analysis Working**: Manuscript analysis identifies 10 Short candidates with rationale
- [x] **TODO Marker System Complete**: Markers inserted/parsed correctly with unique IDs and (start)/(end) tags
- [x] **Storage Schema Complete**: Video YAML stores simplified Shorts metadata (5 fields only)
- [ ] **Scheduling Calculator Working**: Calculates 1-day intervals with randomized times
- [x] **CLI Selection Workflow Functional**: Users can analyze, select from 10 candidates, TODO markers inserted
- [ ] **CLI Upload Workflow Functional**: Batch upload with file paths, scheduled publishing
- [ ] **YouTube Integration Live**: Shorts upload successfully with scheduled dates and main video links
- [ ] **API Endpoints Deployed**: RESTful API supports full Shorts workflow (if applicable)
- [ ] **Documentation Published**: User and editor guides with TODO marker examples
- [ ] **Feature Tested & Validated**: End-to-end testing confirms reliable operation with real manuscripts
- [ ] **Feature Launched**: Available in production for all video categories

## Progress Log

### 2025-11-11
- PRD created
- GitHub issue #339 opened
- Initial architecture defined
- User requirements gathered

### 2025-11-28
- **Design decisions finalized** through user collaboration
- **Decision**: TODO markers with IDs instead of line numbers (manuscript = source of truth)
- **Decision**: Increase candidate count from 5 to 10 (larger selection pool)
- **Decision**: Remove Status field (derived from data)
- **Decision**: Remove Rationale storage (only needed during selection)
- **Decision**: Batch upload workflow with transient file paths
- **Decision**: Hard-coded scheduling (1 day intervals, random times, no config)
- **Decision**: No pinned comments (users swipe through Shorts too quickly)
- **Simplified Short struct** from 7 fields to 5 fields
- **Updated user journey** with TODO marker workflow
- **Updated technical scope** with new Manuscript Modifier module
- **Updated success criteria** and validation strategy
- **PRD ready for implementation**

### 2025-12-16
- **Phase 1 Started: Storage Schema & Configuration**
- ✅ Added `Short` struct to `internal/storage/yaml.go` with 5 fields (ID, Title, Text, ScheduledDate, YouTubeID)
- ✅ Added `Shorts` field to `Video` struct (optional slice)
- ✅ Added `ShortsConfig` struct to `internal/configuration/cli.go`
- ✅ Added `Shorts` field to `Settings` struct
- ✅ Added defaults: `MaxWords: 150`, `CandidateCount: 10`
- ✅ Updated `settings.yaml` with shorts configuration section
- ✅ Added comprehensive tests for Short struct serialization (JSON/YAML)
- ✅ Added tests for Video with Shorts persistence
- ✅ Added tests for ShortsConfig defaults and serialization
- ✅ All tests pass, build successful
- **Milestone Complete**: Storage Schema Complete

### 2025-12-16 (Session 2)
- **Phase 1 & 2: AI Analysis & CLI Selection**
- ✅ Created `internal/ai/shorts.go` - AI module for manuscript analysis
- ✅ Created `internal/ai/templates/shorts.md` - prompt template for identifying Short candidates
- ✅ Created `internal/ai/shorts_test.go` - comprehensive unit tests
- ✅ Created `internal/app/menu_shorts.go` - CLI handler for Shorts analysis
- ✅ Created `internal/app/menu_shorts_test.go` - menu handler tests
- ✅ Modified `internal/app/menu_phase_editor.go` - added Shorts section to Post-Production phase
- ✅ User tested feature with real manuscript - AI successfully identified candidates
- ✅ All tests pass, build successful
- **Milestones Complete**: AI Analysis Working, CLI Selection Workflow Functional
- **Next**: TODO marker system for manuscript modification

### 2025-12-16 (Session 3)
- **TODO Marker System Implementation**
- ✅ Created `internal/manuscript/shorts.go` - manuscript modifier module
- ✅ Implemented `InsertShortMarkers()` - inserts start/end markers around selected segments
- ✅ Implemented text matching with whitespace normalization fallback
- ✅ Implemented `RemoveShortMarkers()` - utility for cleanup/re-analysis
- ✅ Implemented `ExtractShortText()` - extracts text between markers by ID
- ✅ Created `internal/manuscript/shorts_test.go` - 14 comprehensive unit tests
- ✅ Integrated marker insertion into CLI workflow (auto-inserts after selection)
- ✅ Marker format: `TODO: Short (id: short1) (start)` / `TODO: Short (id: short1) (end)`
- ✅ All tests pass, build successful
- **Milestone Complete**: TODO Marker System Complete
- **Next**: Scheduling calculator for publish dates

---

## Notes

- This feature enables a new content multiplication strategy: one long video → multiple Shorts → increased reach
- Consider this as foundation for future content repurposing features (blog snippets, social posts, etc.)
- Word count setting in settings.yaml allows tuning based on actual Short performance
- TODO marker approach makes manuscript the single source of truth for segment boundaries
- Randomized scheduling lays groundwork for future timing analytics (separate PRD)
