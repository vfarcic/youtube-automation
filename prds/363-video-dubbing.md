# PRD: AI-Powered Video Dubbing with ElevenLabs

**Issue**: #363
**Status**: In Progress
**Priority**: High
**Created**: 2025-01-11
**Last Updated**: 2025-01-12
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
- [ ] Spanish dubbing works for local video files using automatic dubbing
- [x] Test mode configuration (watermark + lower resolution + segment time control)
- [ ] Claude AI translates title, description, and tags to Spanish
- [ ] Upload dubbed video to separate Spanish YouTube channel
- [ ] OAuth2 authentication for Spanish channel (separate credentials)
- [ ] Dubbing status persisted in video YAML (allows resumption)
- [ ] CLI integration in Publishing Details phase with options for both long-form and shorts
- [ ] Configuration for ElevenLabs API key, test mode settings, and Spanish channel
- [ ] Support dubbing associated shorts (discovered from video YAML)

### Nice to Have (Future)
- [ ] Pronunciation dictionary for technical terms (if automatic dubbing mispronounces them)
- [ ] Support for additional languages (Portuguese, German, French)
- [ ] Dub directly from YouTube video ID (not just local files)
- [ ] Background polling for dubbing status
- [ ] Batch dubbing for multiple videos
- [ ] API endpoints for programmatic dubbing

## Technical Scope

### Core Components

#### 1. ElevenLabs Client (`internal/dubbing/elevenlabs.go`)
```go
// NewClient creates a new ElevenLabs API client
func NewClient(apiKey string) *Client

// CreateDub initiates a dubbing job
// POST /v1/dubbing with video file and target language
func (c *Client) CreateDub(ctx, videoPath, sourceLang, targetLang string) (*DubbingJob, error)

// GetDubbingStatus checks job status
// GET /v1/dubbing/{dubbing_id}
func (c *Client) GetDubbingStatus(ctx, dubbingID string) (*DubbingJob, error)

// DownloadDubbedAudio downloads the dubbed video
// GET /v1/dubbing/{dubbing_id}/audio/{language_code}
func (c *Client) DownloadDubbedAudio(ctx, dubbingID, langCode, outputPath string) error
```

#### 2. Translation Functions (`internal/ai/translation.go`)
```go
// TranslateTitle translates a video title to target language
func TranslateTitle(ctx, title, targetLang, langName string) (string, error)

// TranslateDescription translates video description
func TranslateDescription(ctx, description, targetLang, langName string) (string, error)

// TranslateTags translates comma-separated tags
func TranslateTags(ctx, tags, targetLang, langName string) (string, error)
```

#### 3. Spanish Channel Upload (`internal/publishing/youtube_spanish.go`)
```go
// UploadVideoToSpanishChannel uploads dubbed video with translated metadata
func UploadVideoToSpanishChannel(video *storage.Video, dubbingInfo *storage.DubbingInfo) (string, error)
```

#### 4. Storage Updates (`internal/storage/yaml.go`)
```go
type DubbingInfo struct {
    LanguageCode    string `yaml:"languageCode"`
    DubbingID       string `yaml:"dubbingId,omitempty"`
    DubbedVideoPath string `yaml:"dubbedVideoPath,omitempty"`
    TranslatedTitle string `yaml:"translatedTitle,omitempty"`
    TranslatedDesc  string `yaml:"translatedDesc,omitempty"`
    TranslatedTags  string `yaml:"translatedTags,omitempty"`
    SpanishVideoId  string `yaml:"spanishVideoId,omitempty"`
    DubbingStatus   string `yaml:"dubbingStatus,omitempty"`
    DubbingError    string `yaml:"dubbingError,omitempty"`
}

// Add to Video struct:
Dubbing []DubbingInfo `yaml:"dubbing,omitempty"`
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

**Phase 3: CLI Integration** ← Moved up to enable testing
- Add dubbing section to Publishing Details phase
- Context-sensitive menu (show relevant actions based on state)
- Present dubbing options for both long-form video AND associated shorts (read from video YAML)
- Handler functions for each action
- Progress feedback during operations

**Phase 4: Dubbing Validation** ← New phase
- Test dubbing with real video file
- Validate dubbed audio quality and sync
- Verify technical terms are pronounced correctly
- Confirm ElevenLabs API integration works end-to-end

**Phase 5: Translation Integration** ← Was Phase 2
- Add translation functions to `internal/ai/`
- Create prompt templates for title, description, tags
- Use existing Claude provider
- Unit tests with mock AI provider

**Phase 6: Spanish Channel Setup** ← Was Phase 3
- Create Spanish YouTube channel
- Generate separate OAuth credentials
- Add Spanish channel config to settings.yaml
- Implement separate OAuth flow (port 8091)

**Phase 7: Upload Integration** ← Was Phase 4
- Implement `UploadVideoToSpanishChannel()`
- Build Spanish descriptions with link to original
- Store Spanish video ID in YAML

**Phase 8: Final Testing & Validation** ← Was Phase 6
- End-to-end testing of complete workflow
- Test translation accuracy
- Verify Spanish channel upload works

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| ElevenLabs API costs | Medium | High | Start with shorter videos; monitor usage |
| Dubbing quality issues | High | Medium | Review results before uploading; allow re-dubbing |
| Long dubbing times | Medium | High | Async workflow; user checks status later |
| OAuth complexity (2 channels) | Medium | Medium | Separate credential files; different callback ports |
| Translation accuracy | Medium | Medium | Review translations before upload; use quality prompts |
| ElevenLabs API changes | Low | Low | Version-pin API; monitor changelog |

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

## Out of Scope

- API mode endpoints (CLI-only for MVP)
- Dubbing from YouTube video ID (local files only)
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
- [ ] **Claude Translation Integration**: Title, description, tags translation working
- [ ] **Spanish YouTube Channel Configured**: Channel created, OAuth credentials set up
- [ ] **Spanish Channel Upload Functional**: Dubbed video uploads with translated metadata
- [ ] **CLI Menu Integration Complete**: Dubbing workflow in Publishing Details phase
- [ ] **End-to-End Workflow Validated**: Full flow tested with real video

## Progress Log

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

- ElevenLabs API supports files up to 1GB and 2.5 hours
- Dubbing can take several minutes; async workflow required
- Spanish channel should use port 8091 for OAuth callback (main channel uses 8090)
- Keep dubbed video path adjacent to original (e.g., `video_es.mp4`)
- Link to original English video in Spanish description for cross-promotion
