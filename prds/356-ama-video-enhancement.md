# PRD: AMA Video Enhancement

**Issue**: #356
**Status**: In Progress
**Created**: 2025-12-06
**Last Updated**: 2025-12-06

## Problem Statement

Weekly YouTube live AMA (Ask Me Anything) sessions have several issues that reduce their value to viewers:

1. **No Timestamps**: Viewers cannot jump to specific questions they're interested in, forcing them to watch the entire video or scrub through manually
2. **Generic Descriptions**: Current descriptions are not content-specific, missing an opportunity to improve discoverability and SEO
3. **Generic Tags**: Tags are not tailored to the actual topics discussed, limiting search visibility
4. **Manual Effort**: Creating timestamps manually is time-consuming and often skipped

Unlike regular videos managed through the existing workflow, AMA sessions are live streams that need post-processing to add metadata based on the actual content discussed.

## Solution Overview

Add a new "Ask Me Anything" section to the CLI that:

1. **Fetches YouTube auto-generated captions** for a given video ID
2. **Uses AI to generate** four pieces of content:
   - **Title**: Content-specific title based on topics discussed
   - **Timecodes**: Q&A segments with timestamps (00:00 = intro/music, rest = questions)
   - **Description**: Content-specific description based on transcript
   - **Tags**: Relevant tags based on topics discussed

   > **Note**: AMA streams have intro music and animation, so 00:00 timecode should be static text like "Intro" or "Stream Starting" so viewers can skip to the first question.
3. **Displays all four for user review and editing**
4. **Applies changes** to the YouTube video via API

### UI Flow

```
Main Menu
├── Create Video
├── List Videos
├── Analyze
└── Ask Me Anything (NEW)
    └── Time Codes
        ├── Video ID: [input field]
        ├── [Generate] button
        │   ↓
        │   Fetches captions → AI generates content
        │   ↓
        │   Shows editable:
        │   - Title (replaces video title)
        │   - Description (replaces content before ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬)
        │   - Tags
        │   - Timecodes (appended at bottom)
        └── [Apply] button → Updates YouTube video
```

### Output Format

**Timecodes**:
```
▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬
00:00 Intro (skip to first question)
02:03 How do you handle secrets in GitOps?
08:13 What's your opinion on Kubernetes vs Nomad?
12:26 Best practices for multi-cluster management
...
```

**Description Structure**:
```
[AI-generated description based on transcript]
▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
[Existing boilerplate content preserved]
...
▬▬▬▬▬▬ ⏱ Timecodes ⏱ ▬▬▬▬▬▬
[Generated timecodes]
```

## User Stories

### Primary User Story
**As a** YouTube content creator with weekly AMA sessions
**I want** to automatically generate timestamps, descriptions, and tags from my live stream transcripts
**So that** viewers can easily find and navigate to questions they're interested in

### Secondary User Stories

**As a** viewer
**I want** to see timestamped questions in the video description
**So that** I can jump directly to topics that interest me

**As a** content creator
**I want** my AMA videos to have content-specific tags and descriptions
**So that** they appear in relevant search results

## Success Criteria

### Must Have
- [x] Fetch YouTube auto-generated captions via API
- [x] Generate content-specific title from transcript
- [x] Generate timecodes identifying Q&A segments (00:00 = "Intro" for music/animation)
- [x] Generate content-specific description from transcript
- [x] Generate relevant tags from transcript (max 450 chars)
- [x] Display all four outputs for user editing before applying
- [x] Apply changes to YouTube video (title, description, and tags)
- [x] New "Ask Me Anything" root menu section
- [ ] All new code has comprehensive tests (80% coverage)

### Should Have
- [x] Reuse existing `SuggestTitles`, `SuggestDescription` and `SuggestTags` patterns
- [x] Preserve existing description boilerplate (after ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬)
- [x] Handle videos without captions gracefully
- [x] Progress feedback during generation

### Could Have
- [x] Save generated content locally before applying (audit trail)
- [ ] Support for multiple languages (caption language selection)
- [ ] Undo/rollback capability

### Won't Have (This Release)
- Management of AMA videos through the regular video workflow
- Automatic scheduling of AMA timestamp generation
- Integration with regular video timecodes feature

## Technical Approach

### Code Reuse Strategy

**Existing Code to Reuse**:

1. **AI Provider System** (`internal/ai/provider.go`)
   - Already abstracted for Azure OpenAI and Anthropic
   - Use `GetAIProvider()` and `GenerateContent()` pattern

2. **Title Generation Pattern** (`internal/ai/titles.go`)
   - Adapt `SuggestTitles()` for transcript input instead of manuscript
   - Similar prompt structure, different input source

3. **Description Generation Pattern** (`internal/ai/descriptions.go`)
   - Adapt `SuggestDescription()` for transcript input instead of manuscript
   - Similar prompt structure, different input source

4. **Tags Generation Pattern** (`internal/ai/tags.go`)
   - Adapt `SuggestTags()` for transcript input
   - Keep 450 character limit and truncation logic

5. **YouTube API Client** (`internal/publishing/youtube.go`)
   - Reuse `getClient()` for OAuth authentication
   - Add new caption fetching and video update functions

6. **Menu System Pattern** (`internal/app/menu_analyze.go`)
   - Follow same structure for new AMA menu
   - Use `huh` forms for input and display

### New Components Required

1. **Caption Fetching** (`internal/publishing/youtube_captions.go`)
   ```go
   // ListCaptions returns available caption tracks for a video
   func ListCaptions(videoID string) ([]CaptionTrack, error)

   // DownloadCaption downloads a caption track as text
   func DownloadCaption(videoID, captionID string) (string, error)

   // GetTranscript is a convenience function that fetches the default caption
   func GetTranscript(videoID string) (string, error)
   ```

2. **AMA Content Generation** (`internal/ai/ama.go`)
   ```go
   // AMAContent holds all generated content for an AMA video
   type AMAContent struct {
       Title       string
       Timecodes   string
       Description string
       Tags        string
   }

   // GenerateAMAContent generates all content from a transcript
   func GenerateAMAContent(ctx context.Context, transcript string) (*AMAContent, error)

   // GenerateAMATitle generates a title from transcript
   func GenerateAMATitle(ctx context.Context, transcript string) (string, error)

   // GenerateAMATimecodes generates timestamped Q&A segments (00:00 = Intro)
   func GenerateAMATimecodes(ctx context.Context, transcript string) (string, error)

   // GenerateAMADescription generates a description from transcript
   func GenerateAMADescription(ctx context.Context, transcript string) (string, error)

   // GenerateAMATags generates tags from transcript
   func GenerateAMATags(ctx context.Context, transcript string) (string, error)
   ```

3. **YouTube Video Update** (`internal/publishing/youtube_update.go`)
   ```go
   // UpdateVideoMetadata updates a video's title, description and tags
   func UpdateVideoMetadata(videoID, title, description, tags string) error

   // GetVideoMetadata fetches current video title, description and tags
   func GetVideoMetadata(videoID string) (*VideoMetadata, error)
   ```

4. **AMA Menu Handler** (`internal/app/menu_ama.go`)
   ```go
   // HandleAMAMenu displays the AMA submenu
   func (m *MenuHandler) HandleAMAMenu() error

   // HandleAMATimecodes handles the timecode generation workflow
   func (m *MenuHandler) HandleAMATimecodes() error
   ```

### AI Prompt Strategy

**Title Prompt** (template in `internal/ai/templates/ama-title.md`):
- Input: Full transcript with timestamps
- Task: Generate a descriptive title based on main topics discussed
- Output: Single line title (max 100 chars)
- Format: "DevOps Q&A: [Main Topics]" or similar

**Timecodes Prompt** (template in `internal/ai/templates/ama-timecodes.md`):
- Input: Full transcript with timestamps
- Task: Identify distinct Q&A segments
- Output: Formatted timecodes (00:00 = "Intro (skip to first question)", rest = questions)
- Format: Plain text, one entry per line
- Note: First entry must be "00:00 Intro (skip to first question)" for intro music/animation

**Description Prompt**:
- Adapt existing description prompt for transcript input
- Focus on summarizing key topics and questions discussed
- Keep concise (1-2 paragraphs)

**Tags Prompt**:
- Adapt existing tags prompt for transcript input
- Focus on specific technologies, concepts, and terms mentioned
- Maintain 450 character limit

### Description Update Logic

```go
// When applying:
// 1. Fetch current description
// 2. Find first occurrence of ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬
// 3. Replace everything before it with new description
// 4. Append timecodes section at the end
// 5. Update video via API
```

## Milestones

### 1. YouTube Caption Fetching ✅
- [x] ~~Implement `ListCaptions()` to get available caption tracks~~ (Not needed - using library)
- [x] ~~Implement `DownloadCaption()` to fetch caption content~~ (Not needed - using library)
- [x] Implement `GetTranscript()` convenience function (using `youtube-transcript-api-go` library)
- [x] Write comprehensive tests for caption functions
- **Validation**: ✅ Can fetch transcript for any public video with captions

### 2. AMA Content Generation ✅
- [x] Create `internal/ai/ama.go` with generation functions
- [x] Create prompt template for title generation
- [x] Create prompt template for timecode generation (00:00 = Intro)
- [x] Adapt description generation for transcript input
- [x] Adapt tags generation for transcript input
- [x] Write tests for all generation functions
- **Validation**: ✅ Given a transcript, generates valid title, timecodes, description, and tags

### 3. YouTube Video Update Capability ✅
- [x] Implement `GetVideoMetadata()` to fetch current title/description/tags/publishedAt
- [x] Implement `UpdateAMAVideo()` to update video with merged description
- [x] Implement description merging logic (preserve boilerplate after ▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬)
- [x] Write tests for update functions (`youtube_update_test.go`)
- **Validation**: ✅ Can update a video's title, description, and tags

### 4. AMA Menu and UI ✅
- [x] Add "Ask Me Anything" option to main menu (direct access, no submenu)
- [x] Create `menu_ama.go` with single-screen form handler
- [x] Implement video ID input field
- [x] Implement "Generate with AI" button (fetches transcript + generates all content)
- [x] Implement editable fields for title, description, tags, timecodes
- [x] Implement "Apply to YouTube" button (updates video + saves local files)
- [x] Local file saving to `manuscript/ama/` using existing `storage.Video` struct
- **Validation**: ✅ Full workflow works end-to-end in CLI

### 5. Integration and Testing
- [x] Run full test suite (all tests pass)
- [x] Manual end-to-end testing with real AMA video
- [ ] Verify test coverage >= 80%
- [ ] Update CLAUDE.md with AMA feature documentation
- **Validation**: Feature complete and documented

## Dependencies

### Internal Dependencies
- AI provider system (`internal/ai/provider.go`)
- YouTube API client (`internal/publishing/youtube.go`)
- Menu system (`internal/app/`)
- OAuth authentication (existing YouTube scopes include caption access)

### External Dependencies
- YouTube Data API v3 (captions.list, captions.download, videos.update)
- Existing AI provider (Azure OpenAI or Anthropic)

### Blocking Dependencies
None - all required APIs and patterns already exist

## Risks & Mitigations

### Risk 1: Caption Quality/Availability
**Impact**: Medium
**Likelihood**: Medium
**Description**: Auto-generated captions may have errors or not exist for all videos
**Mitigation**:
- Handle missing captions gracefully with clear error message
- AI can work with imperfect transcripts
- User can edit generated content before applying

### Risk 2: AI Timestamp Accuracy
**Impact**: Medium
**Likelihood**: Medium
**Description**: AI may not correctly identify question boundaries
**Mitigation**:
- User reviews and edits before applying
- Include clear instructions in AI prompt
- 00:00 entry provides fallback for intro

### Risk 3: YouTube API Rate Limits
**Impact**: Low
**Likelihood**: Low
**Description**: Hitting rate limits during caption fetch or video update
**Mitigation**:
- Single video per operation (not batch)
- Existing OAuth flow handles token refresh
- Can add retry logic if needed

### Risk 4: Description Format Changes
**Impact**: Medium
**Likelihood**: Low
**Description**: Existing AMA videos may have different description formats
**Mitigation**:
- Use `▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬▬` as reliable delimiter
- Preview before applying allows user to verify
- Can manually adjust if format differs

## Open Questions

1. **Should we store generated content locally before applying?**
   - Pro: Audit trail, can review later
   - Con: Additional complexity
   - Decision: Defer to "Could Have"

2. **Should timecodes use MM:SS or HH:MM:SS format?**
   - Most AMAs are < 1 hour, so MM:SS is cleaner
   - YouTube accepts both formats
   - Decision: Use MM:SS for < 1 hour, HH:MM:SS for longer

3. **How to handle videos without auto-generated captions?**
   - Display clear error message
   - Suggest enabling auto-captions in YouTube Studio
   - Decision: Fail gracefully with helpful message

## Out of Scope

- Managing AMA videos through regular video workflow
- Automatic scheduling of timestamp generation
- Batch processing multiple AMA videos
- Manual transcript input (only YouTube captions)
- Caption language selection (use default/first available)

## Progress Log

### 2025-12-06 (Session 4)
- **Milestones 3 & 4: YouTube Video Update + AMA Menu/UI - COMPLETE**
- Implemented `GetVideoMetadata()` with `PublishedAt` field in `youtube_update.go`
- Implemented `UpdateAMAVideo()` with description merging logic
- Implemented `buildAMADescription()` to preserve boilerplate and append timecodes
- Created comprehensive tests in `youtube_update_test.go`
- Redesigned `menu_ama.go`:
  - Single-screen form (removed submenu)
  - Video ID, Title, Description, Tags, Timecodes fields
  - "Generate with AI" button fetches transcript + metadata, generates all content
  - "Apply to YouTube" button updates video and saves local files
- Local file saving to `manuscript/ama/YYYY-MM-DD-videoID.yaml` and `.md`
  - Reuses existing `storage.Video` struct and `WriteVideo()` function
  - Date extracted from video's YouTube publish date
- AI prompt improvements:
  - Added "Viktor (with a K)" instruction to all 4 prompts
  - Timecodes prompt now skips non-questions (sponsor mentions, intro chatter)
  - Fixed empty line before timecodes header in description
- Created `menu_ama_test.go` with tests for `extractDateFromISO()`
- All tests pass

### 2025-12-06 (Session 3)
- **Milestone 2: AMA Content Generation - COMPLETE**
- Created `internal/ai/ama.go` with 5 generation functions:
  - `GenerateAMATitle()` - generates AMA-specific title from transcript
  - `GenerateAMATimecodes()` - extracts Q&A segments with timestamps
  - `GenerateAMADescription()` - generates description from transcript
  - `GenerateAMATags()` - generates tags with 450 char limit
  - `GenerateAMAContent()` - convenience function for all 4 outputs
- Created prompt templates:
  - `internal/ai/templates/ama-title.md`
  - `internal/ai/templates/ama-timecodes.md`
- Refactored existing AI modules to use template files:
  - `descriptions.go` now uses `templates/description.md`
  - `tags.go` now uses `templates/tags.md`
- Created comprehensive tests in `ama_test.go` (78.3% coverage)
- All tests pass

### 2025-12-06 (Session 2)
- Updated PRD to include title generation from transcript
- Added note about intro music/animation (00:00 = "Intro" so viewers can skip)
- Updated all sections to reflect 4 outputs instead of 3

### 2025-12-06 (Session 1)
- **Milestone 1: YouTube Caption Fetching - COMPLETE**
- Implemented `GetTranscript()` using `github.com/horiagug/youtube-transcript-api-go` library
- Added "Ask Me Anything" menu to main CLI
- Created `menu_ama.go` with `HandleAMAMenu()` and `HandleAMAFetchTranscript()`
- Fixed broken `TestMain` in youtube_test.go (tests weren't running)
- Fixed pre-existing bugs exposed by TestMain fix:
  - `TestGetAdditionalInfo` - was calling wrong function
  - `TestUploadVideo` / `TestUpdateVideoLanguage` - wrong language fallback expectations
  - `isShort()` - ISO 8601 duration parsing bug
- Removed useless integration tests that required real credentials
- All tests pass (100%)
- Successfully tested transcript fetch with real AMA video

### 2025-12-06 (Initial)
- PRD created
- GitHub issue #356 opened
- Identified code reuse opportunities (descriptions.go, tags.go, youtube.go)
- Defined 5 major milestones

---

**Next Steps**: Complete Milestone 5 (verify test coverage, update CLAUDE.md documentation)
