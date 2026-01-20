# PRD: AI-Powered Video Dubbing with ElevenLabs

**Issue**: #363
**Status**: Complete
**Priority**: High
**Created**: 2025-01-11
**Last Updated**: 2026-01-20
**Depends On**: None

---

## Problem Statement

Reaching Spanish-speaking audiences requires manually dubbing videos and translating metadata, which is:
- Time-consuming (hours per video for professional dubbing)
- Expensive (hiring voice actors or dubbing services)
- Limiting content reach to English-speaking audiences only

The YouTube Automation Tool currently has no support for creating multilingual versions of videos, forcing creators to choose between manual effort or ignoring non-English audiences entirely.

## Proposed Solution

Integrate AI-powered video dubbing using ElevenLabs API with automatic metadata translation via Claude AI:

1. **ElevenLabs Automatic Dubbing**: Use ElevenLabs' automatic dubbing which naturally preserves English pronunciation for technical terms (kubectl, docker, helm, etc.)
2. **Pronunciation Dictionary** (if needed): Optional pronunciation dictionary for problematic terms that aren't handled well by automatic dubbing
3. **Claude Translation**: Translate title, description, and tags to Spanish
4. **Spanish YouTube Channel**: Upload dubbed videos to a dedicated Spanish channel for better audience targeting
5. **Publishing Integration**: Add dubbing workflow to the existing Publishing Details phase

**Note**: ElevenLabs Multilingual v2 has a natural tendency to keep English/technical words pronounced correctly even in Spanish output. We start simple and only add pronunciation dictionary complexity if real-world testing shows issues.

### User Journey

**Current State (Manual)**:
1. Creator records and publishes English video
2. Manually pays for dubbing service or records Spanish voiceover
3. Manually translates title, description, tags
4. Manually uploads to Spanish channel or as alternative audio track
5. No tracking or automation

**After (With This Feature)**:
1. Creator publishes English video as normal
2. Opens Publishing Details, selects "Start Spanish Dubbing"
3. System shows dubbing options for long-form video AND any associated shorts
4. System initiates ElevenLabs automatic dubbing (async, can take minutes)
5. Creator checks status later, downloads dubbed video when ready
6. System translates title/description/tags using Claude AI
7. Creator uploads to Spanish channel with one click
8. Spanish video ID tracked in video YAML (for both long-form and shorts)

## Success Criteria

### Must Have (MVP)
- [x] ElevenLabs API integration: create dubbing job, poll status, download result
- [x] Spanish dubbing works for local video files using automatic dubbing
- [x] Smart video compression for files >1GB (auto-compress to fit ElevenLabs limit)
- [x] Test mode configuration (watermark + lower resolution + segment time control)
- [x] Claude AI translates title, description, tags, and timecodes to Spanish
- [x] Upload dubbed video to separate Spanish YouTube channel
- [x] OAuth2 authentication for Spanish channel (separate credentials)
- [x] Dubbing status persisted in video YAML (allows resumption)
- [x] CLI integration in Dubbing phase with options for both long-form and shorts
- [x] Configuration for ElevenLabs API key, test mode settings, and Spanish channel
- [x] Support dubbing associated shorts (discovered from video YAML)

### Nice to Have (Future)
- [ ] Pronunciation dictionary for technical terms (if automatic dubbing mispronounces them)
- [ ] Support for additional languages (Portuguese, German, French)
- [x] Dub directly from YouTube URL (fallback option - may have reliability issues)
- [ ] Background polling for dubbing status
- [ ] Batch dubbing for multiple videos
- [ ] API endpoints for programmatic dubbing

## Technical Scope

### Core Components

#### 1. ElevenLabs Client (`internal/dubbing/elevenlabs.go`)
```go
// NewClient creates a new ElevenLabs API client
func NewClient(apiKey string, config Config) *Client

// CreateDubFromFile initiates a dubbing job using a local video file
// POST /v1/dubbing with multipart file upload
// Automatically compresses video if >1GB using smart compression algorithm
func (c *Client) CreateDubFromFile(ctx, filePath, sourceLang, targetLang string) (*DubbingJob, error)

// CreateDubFromURL initiates a dubbing job using a YouTube URL (fallback option)
// POST /v1/dubbing with source_url and target language
// Note: May have reliability issues due to YouTube/ElevenLabs restrictions
func (c *Client) CreateDubFromURL(ctx, youtubeURL, sourceLang, targetLang string) (*DubbingJob, error)

// GetDubbingStatus checks job status
// GET /v1/dubbing/{dubbing_id}
func (c *Client) GetDubbingStatus(ctx, dubbingID string) (*DubbingJob, error)

// DownloadDubbedAudio downloads the dubbed video (auto-creates output directory)
// GET /v1/dubbing/{dubbing_id}/audio/{language_code}
func (c *Client) DownloadDubbedAudio(ctx, dubbingID, langCode, outputPath string) error
```

#### 1b. Video Compression (`internal/dubbing/compress.go`)
```go
// CompressForDubbing compresses video to fit under 1GB limit while maximizing quality
// Algorithm:
//   - If video ≤ 1GB → return original path (no compression needed)
//   - If video > 1GB and duration ≤ 25 min → 4K at dynamic CRF (calculated to target ~900MB)
//   - If video > 1GB and duration > 25 min → 1080p at CRF 26
func CompressForDubbing(ctx context.Context, inputPath string) (outputPath string, err error)

// GetVideoInfo returns video metadata (duration, size, resolution)
func GetVideoInfo(filePath string) (*VideoInfo, error)

// CalculateOptimalCRF determines the best CRF value to hit target size while maintaining quality
// CRF range: 23 (high quality) to 30 (more compression)
// If calculated CRF > 30, switches to 1080p instead
func CalculateOptimalCRF(duration float64, targetSizeMB int) (crf int, use1080p bool)
```

#### 2. Translation Functions (`internal/ai/translation.go`)
```go
// VideoMetadataInput holds the input fields for translation
type VideoMetadataInput struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Tags        string   `json:"tags"`
    Timecodes   string   `json:"timecodes"`
    ShortTitles []string `json:"shortTitles,omitempty"` // Titles of YouTube Shorts to translate
}

// VideoMetadataOutput holds the translated fields
type VideoMetadataOutput struct {
    Title       string   `json:"title"`
    Description string   `json:"description"`
    Tags        string   `json:"tags"`
    Timecodes   string   `json:"timecodes"`
    ShortTitles []string `json:"shortTitles,omitempty"` // Translated Short titles
}

// TranslateVideoMetadata translates all metadata in a single API call
func TranslateVideoMetadata(ctx context.Context, input VideoMetadataInput, targetLanguage string) (*VideoMetadataOutput, error)
```

#### 3. Spanish Channel Upload (`internal/publishing/youtube_spanish.go`)
```go
// UploadVideoToSpanishChannel uploads dubbed video with translated metadata
func UploadVideoToSpanishChannel(video *storage.Video, dubbingInfo *storage.DubbingInfo) (string, error)
```

#### 4. Storage Updates (`internal/storage/yaml.go`)
```go
// DubbingInfo tracks dubbing status for a specific language.
// The language code is the map key in Video.Dubbing (e.g., "es" for Spanish).
type DubbingInfo struct {
    DubbingID       string `yaml:"dubbingId,omitempty"`
    DubbedVideoPath string `yaml:"dubbedVideoPath,omitempty"`
    Title           string `yaml:"title,omitempty"`           // Translated title
    Description     string `yaml:"description,omitempty"`     // Translated description
    Tags            string `yaml:"tags,omitempty"`            // Translated tags
    Timecodes       string `yaml:"timecodes,omitempty"`       // Translated timecodes
    UploadedVideoID string `yaml:"uploadedVideoId,omitempty"` // YouTube ID on target channel
    DubbingStatus   string `yaml:"dubbingStatus,omitempty"`   // "", "dubbing", "dubbed", "failed"
    DubbingError    string `yaml:"dubbingError,omitempty"`
}

// Video struct includes:
Dubbing map[string]DubbingInfo `yaml:"dubbing,omitempty"` // Key = language code (e.g., "es")
```

### Configuration

**settings.yaml additions:**
```yaml
elevenLabs:
  # apiKey: use ELEVENLABS_API_KEY env var
  testMode: true      # true = watermark + lower resolution (saves credits)
  startTime: 0        # Start time in seconds (0 = beginning)
  endTime: 60         # End time in seconds (0 = full video)
  # When done testing, set: testMode: false, endTime: 0

spanishChannel:
  channelId: "UC_YOUR_SPANISH_CHANNEL_ID"
  credentialsFile: "client_secret_spanish.json"
  tokenFile: "youtube-go-spanish.json"
```

**Default API parameters (based on use case):**
- `num_speakers: 1` - Single speaker (just you)
- `drop_background_audio: false` - Preserve music/sound effects between sections
- `watermark: true/false` - Based on testMode setting
- `highest_resolution: true/false` - Based on testMode setting (inverse)
- `start_time/end_time` - From settings for segment testing

### Implementation Phases

**Phase 1: ElevenLabs API Integration** ✅
- [x] Create `internal/dubbing/` package
- [x] Implement API client with create, status, download
- [x] Unit tests with mock HTTP server (82.5% coverage)

**Phase 2: ElevenLabs Configuration** ✅
- [x] Add `SettingsElevenLabs` struct to configuration
- [x] Add `elevenLabs:` section to settings.yaml
- [x] Environment variable support (`ELEVENLABS_API_KEY`)
- [x] Unit tests for config loading

**Phase 3: CLI Integration** ✅
- [x] Add dubbing as separate phase (renamed "Publishing Details" to "Upload")
- [x] Context-sensitive menu (show relevant actions based on state)
- [x] Present dubbing options for both long-form video AND associated shorts (read from video YAML)
- [x] Handler functions for each action (start dubbing, check status, auto-download)
- [x] Progress feedback during operations
- [x] Progress counter shows X/Y (1 long-form + N shorts)

**Phase 4: Dubbing Validation** ✅
- [x] Test dubbing with real video file
- [x] Validate dubbed audio quality and sync
- [x] Verify technical terms are pronounced correctly
- [x] Confirm ElevenLabs API integration works end-to-end
- [x] Fixed MIME type issue for .mov files (video/quicktime)

**Phase 5: Translation Integration** ✅
- [x] Add `TranslateVideoMetadata()` function to `internal/ai/translation.go`
- [x] Create `translate_metadata.md` prompt template (single call for all fields)
- [x] Use existing Claude/Azure OpenAI provider
- [x] Unit tests with mock AI provider (17 test cases)
- [x] CLI integration: "Translate Metadata" option in Dubbing phase
- [x] Progress counter updated to include translation step

**Phase 6: Spanish Channel Setup** ✅
- [x] Add `SettingsSpanishChannel` struct to configuration
- [x] Add `spanishChannel` section to settings.yaml with defaults
- [x] Refactor OAuth flow to accept parameters (credentials file, token file, port)
- [x] Create `OAuthConfig` struct for parameterized OAuth
- [x] Create `GetSpanishChannelClient()` and `GetSpanishChannelID()` functions
- [x] Create `SpanishOAuthConfig()` with defaults (port 8091, separate credentials)
- [x] Unit tests for Spanish channel configuration (7 test cases)
- [x] Unit tests for OAuth config functions (4 test cases)
- [x] **MANUAL**: Create Spanish YouTube channel (ID: `UCM7ZVtFa6baCzPRMwtIt_gA`)
- [x] **MANUAL**: Generate OAuth credentials in Google Cloud Console
- [x] **MANUAL**: Add authorized redirect URI: `http://localhost:8091`
- [x] **MANUAL**: Download credentials as `client_secret_spanish.json`
- [x] **MANUAL**: Fill in `channelId` in settings.yaml

**Phase 7: Video Compression** ✅
- [x] Create `internal/dubbing/compress.go` module
- [x] Implement `GetVideoInfo()` using FFprobe
- [x] Implement `CalculateOptimalCRF()` algorithm
- [x] Implement `CompressForDubbing()` using FFmpeg
- [x] Add `CreateDubFromFile()` method with auto-compression
- [x] Unit tests for compression logic (81.4% coverage)
- [x] Integration test with real video files

**Phase 8: Upload Integration** ✅
- [x] Implement `UploadDubbedVideo()` with date-based scheduling
- [x] Implement `UploadDubbedShort()` with interval scheduling
- [x] Build Spanish descriptions with link to original
- [x] Store Spanish video ID in YAML
- [x] CLI "Upload All to YouTube" option
- [x] Progress counter includes upload step

**Phase 9: Final Testing & Validation** ✅
- [x] End-to-end testing of complete workflow
- [x] Test translation accuracy
- [x] Verify Spanish channel upload works
- [x] Test compression with various video sizes

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| ElevenLabs API costs | Medium | High | Start with shorter videos; monitor usage |
| Dubbing quality issues | High | Medium | Review results before uploading; allow re-dubbing |
| Long dubbing times | Medium | High | Async workflow; user checks status later |
| OAuth complexity (2 channels) | Medium | Medium | Separate credential files; different callback ports |
| Translation accuracy | Medium | Medium | Review translations before upload; use quality prompts |
| ElevenLabs API changes | Low | Low | Version-pin API; monitor changelog |
| FFmpeg not installed | Medium | Low | Check for FFmpeg at startup; provide clear error message |
| YouTube URL unreliable | Medium | High | Use local file upload as primary; YouTube URL as fallback |
| Large file compression time | Low | Medium | Show progress; compression is one-time per video |

## Dependencies

### Internal
- Existing YouTube upload functionality (`internal/publishing/youtube.go`)
- Existing AI provider (`internal/ai/provider.go`)
- Video YAML storage (`internal/storage/yaml.go`)
- CLI menu system (`internal/app/menu_phase_editor.go`)

### External
- ElevenLabs Dubbing API (new)
- YouTube Data API v3 (existing)
- Claude AI / Azure OpenAI (existing)
- FFmpeg (new - for video compression)

## Out of Scope

- API mode endpoints (CLI-only for MVP)
- Multiple languages (Spanish only for MVP)
- Custom voice selection (use ElevenLabs auto-detect)
- Background polling (user-triggered status checks)
- Batch dubbing multiple videos at once
- Thumbnail translation/recreation

## Validation Strategy

### Testing Approach
- Unit tests for ElevenLabs client (mock HTTP server)
- Unit tests for translation functions (mock AI provider)
- Unit tests for Spanish channel upload (mock YouTube API)
- Integration tests for complete workflow
- Table-driven tests following project patterns

### Manual Validation
- Test with real video file (short duration to minimize cost)
- Verify dubbed audio syncs with video
- Check translation quality for title/description/tags
- Confirm upload to Spanish channel works
- Validate video appears in Spanish channel dashboard

## Milestones

- [x] **ElevenLabs API Integration Working**: Create, poll, and download dubbing jobs
- [x] **Claude Translation Integration**: Title, description, tags, timecodes translation working
- [x] **Spanish YouTube Channel Configured**: Channel created, OAuth credentials set up
- [x] **Smart Video Compression Working**: Auto-compress large videos to fit 1GB limit
- [x] **Local File Upload Functional**: Dub from local files with compression support
- [x] **Spanish Channel Upload Functional**: Dubbed video uploads with translated metadata
- [x] **CLI Menu Integration Complete**: Dubbing workflow in Dubbing phase
- [x] **End-to-End Workflow Validated**: Full flow tested with real video

## Progress Log

### 2026-01-20 (Update 11) - PRD COMPLETE
- **Phase 8 Complete**: Upload Integration
  - Implemented `UploadDubbedVideo()` with date-based scheduling (future dates = scheduled private, past dates = public immediately)
  - Implemented `UploadDubbedShort()` with interval scheduling (Short 1 = +1 day, Short 2 = +2 days, etc.)
  - Added CLI "Upload All to YouTube" option (replaced individual upload buttons)
  - Progress counter now includes upload step (X/8 total)
  - Extended `TranslateVideoMetadata()` to include short titles in single API call
  - Comprehensive tests for upload validation, date parsing, and scheduling
- **Phase 9 Complete**: Final Testing & Validation
  - All tests passing
  - End-to-end workflow validated
- **All Milestones Achieved**: PRD marked as Complete

### 2026-01-20 (Update 10)
- **Phase 7 Fully Complete**: Integration tested with real video file
- **CLI Improvements**: Enhanced dubbing workflow UX
  - Menu now shows `[Local]` or `[YouTube]` indicator next to each video option
  - Shows source type before user selects, so they know what will happen
  - Local file is tried first if it exists, falls back to YouTube URL
  - Clear step-by-step progress messages:
    - "Step 1/2: Compressing video (file >1GB)..." for large files
    - "Step 1/1: Uploading to ElevenLabs..." for small files
  - Updated `menu_phase_editor.go` with `getSourceLabel()` helper function
- User tested end-to-end with real .mov file - confirmed working

### 2026-01-20 (Update 9)
- **Phase 7 Complete**: Video Compression Implementation
  - Created `internal/dubbing/compress.go` with smart compression algorithm
  - `VideoInfo` struct for video metadata (duration, size, resolution)
  - `GetVideoInfo()` extracts metadata using FFprobe JSON output
  - `CalculateOptimalCRF()` determines optimal CRF to hit target size:
    - Videos ≤1GB: no compression needed
    - Videos >1GB & ≤25min: 4K at dynamic CRF (target ~900MB)
    - Videos >1GB & >25min: 1080p at CRF 26
  - `CompressForDubbing()` orchestrates compression decision and FFmpeg execution
  - `CommandExecutor` interface for testability (mock FFprobe/FFmpeg in tests)
  - Added `CreateDubFromFile()` method to `elevenlabs.go` with auto-compression
  - `getMIMEType()` helper for proper content-type detection
  - Compressed files cleaned up after successful upload
  - Created `internal/dubbing/compress_test.go` with comprehensive tests
  - 81.4% test coverage for dubbing package
  - All tests passing, build verified

### 2026-01-20 (Update 8)
- **Decision**: Add local file upload with smart compression (YouTube URL kept as fallback)
  - **Rationale**: ElevenLabs has reliability issues fetching videos from YouTube due to platform restrictions
  - **Change**: Add `CreateDubFromFile()` method alongside existing `CreateDubFromURL()`
  - **Smart Compression Algorithm**:
    - If video ≤ 1GB → upload as-is, no compression needed
    - If video > 1GB and duration ≤ ~25 min → compress to 4K at dynamic CRF (maximize quality under 1GB)
    - If video > 1GB and duration > ~25 min → compress to 1080p at CRF 26 (clean quality)
  - **Testing Results** (with real videos):
    - 44.5 min 4K video (17GB) → 1080p @ CRF 26 → ~620MB (clean quality)
    - 17.6 min 4K video (4.9GB) → 4K @ CRF 24 → ~794MB (excellent quality)
  - **Impact**: New `internal/dubbing/compress.go` module needed
  - **Dependencies**: Requires FFmpeg installed on system

### 2026-01-12 (Update 7)
- **YouTube URL Dubbing**: Switched from local file upload to YouTube URL
  - Removed `CreateDub()` method (local file upload had issues with large files >1GB)
  - Added `CreateDubFromURL()` method - ElevenLabs fetches directly from YouTube
  - Videos must be published on YouTube before dubbing (public or unlisted)
  - Both long-form videos AND shorts now use YouTube URL
  - No file size limits, faster initiation (no upload wait)
- **Improved Error Handling**: Better capture of ElevenLabs error responses
  - Handles multiple error formats: `{"detail": {...}}`, `{"detail": "string"}`, `{"error": "..."}`, etc.
  - `GetMessage()` method extracts best available error message
- **Auto-create Output Directory**: `DownloadDubbedAudio()` now creates parent directories
- **Comprehensive Tests**: Added tests for URL dubbing, error handling, directory creation
- All tests passing, build verified

### 2026-01-12 (Update 6)
- **Phase 6 Complete (Code)**: Spanish Channel Setup
  - Added `SettingsSpanishChannel` struct to `internal/configuration/cli.go`
  - Added `spanishChannel` section to `settings.yaml` with configurable fields
  - Refactored OAuth flow in `internal/publishing/youtube.go` to be parameterized
  - Created `OAuthConfig` struct with `CredentialsFile`, `TokenFileName`, `CallbackPort`
  - Created `DefaultOAuthConfig()` for main channel (port 8090)
  - Created `SpanishOAuthConfig()` for Spanish channel (port 8091)
  - Created `GetSpanishChannelClient()` for Spanish channel authentication
  - Created `GetSpanishChannelID()` helper function
  - Added parameterized functions: `startWebServerWithPort()`, `tokenCacheFileWithName()`, `getTokenFromWebWithPort()`
  - All existing OAuth functions delegate to parameterized versions (backward compatible)
  - Comprehensive unit tests (11 new test cases across configuration and publishing)
  - All tests passing, build verified
- **Phase 6 Complete**: All manual setup steps finished
  - Created Spanish YouTube channel (ID: `UCM7ZVtFa6baCzPRMwtIt_gA`)
  - Generated OAuth credentials in Google Cloud Console
  - Configured redirect URI `http://localhost:8091`
  - Downloaded `client_secret_spanish.json`
  - Updated `channelId` in settings.yaml
  - Added `client_secret_spanish.json` to `.gitignore`

### 2025-01-12 (Update 5)
- **Phase 5 Complete**: Translation Integration
  - Created `internal/ai/translation.go` with `TranslateVideoMetadata()` function
  - Single API call translates title, description, tags, and timecodes together (consistency + efficiency)
  - Created `internal/ai/templates/translate_metadata.md` prompt template
  - Prompt uses general principle with examples for technical term preservation
  - Added `Timecodes` field to `DubbingInfo` struct
  - Refactored field names: `title`, `description`, `tags`, `timecodes` (not `translatedTitle`, etc.)
  - Comprehensive unit tests (17 test cases) covering special characters, code fences, errors
  - CLI integration: "Translate Metadata" option in Dubbing phase menu
  - Translation available anytime there's a title (not just after dubbing)
  - Updated `CalculateDubbingProgress()` to include translation as a step
  - User tested and confirmed working
  - All tests passing, build verified

### 2025-01-12 (Update 4)
- **Phase 3 & 4 Complete**: CLI Integration and Dubbing Validation
  - Added `DubbingInfo` struct and `Dubbing` map to Video storage
  - Renamed "Publishing Details" to "Upload", added "Dubbing" phase below it
  - Dubbing menu shows long-form video AND all associated shorts with status
  - Context-sensitive actions: Start Dubbing, Check Status (auto-downloads when complete)
  - Progress counter shows X/Y (1 long-form + N shorts)
  - Added `CalculateDubbingProgress()` function to video manager
  - Fixed MIME type detection for .mov files (video/quicktime)
  - Successfully tested dubbing with real video file via ElevenLabs API
  - All tests passing, build verified

### 2025-01-12 (Update 3)
- **Decision**: Reordered implementation phases
  - **Rationale**: Validate dubbing with real video before building translation on top
  - **Change**: Moved CLI Integration (was Phase 5) to Phase 3, added explicit Dubbing Validation phase
  - **New order**: API → Config → CLI → Validate Dubbing → Translation → Spanish Channel → Upload → Final Testing
  - **Impact**: Ensures foundation is solid before building more features

### 2025-01-12 (Update 2)
- **ElevenLabs Configuration Complete**:
  - Added `SettingsElevenLabs` struct to `internal/configuration/cli.go`
  - Fields: APIKey, TestMode, StartTime, EndTime, NumSpeakers, DropBackgroundAudio
  - Environment variable support: `ELEVENLABS_API_KEY`
  - Default `numSpeakers: 1` applied automatically
  - Added `elevenLabs:` section to `settings.yaml` with documentation
  - Comprehensive unit tests (8 test cases) for config loading and serialization
  - All tests passing, build verified

### 2025-01-12
- **Phase 1 Complete**: ElevenLabs API Integration
  - Created `internal/dubbing/` package with types and client
  - Implemented all API methods: `CreateDub`, `GetDubbingStatus`, `DownloadDubbedAudio`
  - Config struct supports: TestMode, StartTime/EndTime, NumSpeakers, DropBackgroundAudio
  - Comprehensive unit tests with mock HTTP server (82.5% coverage)
  - All tests passing, build verified
- Technical decisions confirmed:
  - Single base URL: `https://api.elevenlabs.io`
  - No fixed HTTP timeout - use context for cancellation (supports 45min videos)
  - No retry logic - return clear error messages

### 2025-01-11 (Update 2)
- **Decision**: Support dubbing for both long-form videos AND associated shorts
  - **Rationale**: Videos often have shorts derived from them; users need to dub both for consistent Spanish channel content
  - **Impact**: CLI must read video YAML to discover associated shorts and present dubbing options for all videos
  - **Affects**: Phase 5 (CLI Integration) - must show options for long-form + shorts

### 2025-01-11
- PRD created
- GitHub issue #363 opened
- Implementation plan designed during planning session
- Key decisions made:
  - Spanish first, extensible to other languages
  - Separate Spanish YouTube channel (not multi-audio track)
  - ElevenLabs automatic dubbing for audio (no Claude transcript translation)
  - Claude only for metadata translation (title, description, tags)
  - Integration in Publishing Details phase
- Research findings:
  - ElevenLabs Multilingual v2 naturally preserves English pronunciation for technical terms
  - Pronunciation dictionary available if needed (moved to "Nice to Have")
  - Start simple, iterate based on real-world testing

---

## Notes

- **Two dubbing methods available**: Local file upload (primary) and YouTube URL (fallback)
- **Local file upload**: Recommended method; videos >1GB are auto-compressed to fit ElevenLabs 1GB limit
- **YouTube URL**: May have reliability issues due to YouTube/ElevenLabs platform restrictions; kept as fallback
- **Smart compression**: Maximizes quality within 1GB limit (4K for short videos, 1080p for long videos)
- Dubbing can take several minutes; async workflow required
- Spanish channel should use port 8091 for OAuth callback (main channel uses 8090)
- Keep dubbed video path adjacent to original (e.g., `video_es.mp4`)
- Link to original English video in Spanish description for cross-promotion
- Output directory is auto-created when downloading dubbed video
- Compressed videos stored temporarily; cleaned up after dubbing completes
