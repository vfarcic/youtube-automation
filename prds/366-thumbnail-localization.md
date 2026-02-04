# PRD: AI-Powered Thumbnail Localization with Google Nano Banana

**Issue**: #366
**Status**: In Progress
**Priority**: Medium
**Created**: 2026-01-21
**Last Updated**: 2026-02-04
**Depends On**: #363 (Video Dubbing - Complete)

---

## Problem Statement

When dubbing videos to other languages, thumbnails still contain English text. Currently, translating thumbnail text requires:
- Manual work in Google AI Studio with Nano Banana
- Going back to the agency for each language variant (costly, slow)
- No automation possible since agency delivers finished images (no templates/source files)

The `Tagline` field in video YAML contains the exact text that appears in thumbnails, but this information isn't being leveraged for automation.

## Proposed Solution

Integrate Google's Nano Banana API to automatically generate localized thumbnails:

1. **Thumbnail Generation**: Use Nano Banana to replace English tagline text with translated version
2. **Leverage Existing Data**: Read `Tagline` from video YAML as the known text to replace
3. **Consistent Naming**: Save localized thumbnails as `[ORIGINAL_NAME]-[lang].[EXT]`
4. **YouTube Upload**: Add ability to upload language-specific thumbnails to dubbed videos
5. **Model Selection**: Test both Nano Banana (cheaper) and Nano Banana Pro (better text) to determine quality threshold

### User Journey

**Current State (Manual)**:
1. Creator receives thumbnail from agency (English text)
2. Opens Google AI Studio manually
3. Uploads thumbnail and prompts for Spanish translation
4. Downloads result, verifies quality
5. Manually uploads to Spanish YouTube video
6. Repeats for each language

**After (With This Feature)**:
1. Creator opens Dubbing menu, selects "Generate Thumbnail"
2. System reads English thumbnail path and Tagline from YAML
3. System calls Nano Banana API with specific replacement instruction
4. Localized thumbnail saved automatically with language suffix
5. **Thumbnail auto-opens in default image viewer** for instant visual verification
6. Creator verifies result, then selects "Upload Thumbnail" to push to dubbed video on YouTube
7. Process works for all supported dubbing languages

## Success Criteria

### Must Have (MVP)
- [x] Google Gemini API client for image generation (`internal/thumbnail/gemini.go`)
- [x] Generate localized thumbnail from English original + tagline
- [x] Support same languages as dubbing (Spanish initially, extensible)
- [x] Save generated thumbnail as `[ORIGINAL_NAME]-[lang].[EXT]`
- [ ] Store localized thumbnail path in `DubbingInfo` struct
- [ ] CLI option "Generate Thumbnail" in Dubbing menu (auto-opens result for preview)
- [ ] CLI option "Upload Thumbnail" for dubbed videos
- [ ] Upload thumbnail to correct YouTube video using existing `Thumbnails.Set` API
- [x] Configuration for Gemini API key and model selection
- [x] Unit tests with mocked Gemini API

### Nice to Have (Future)
- [ ] Batch generation for all languages at once
- [ ] Side-by-side preview comparison (English vs localized)
- [ ] Automatic quality validation (detect if text is readable)
- [ ] Support for thumbnail variants (if video has multiple thumbnails)
- [ ] API endpoint for programmatic thumbnail generation

## Technical Scope

### Core Components

#### 1. Gemini Client (`internal/thumbnail/gemini.go`) ✅ IMPLEMENTED
```go
// Config holds Gemini API configuration
type Config struct {
    APIKey string
    Model  string // "gemini-3-pro-image-preview" (default) or "gemini-2.5-flash-image"
}

// Client is the Google Gemini API client for image generation
type Client struct {
    config     Config
    httpClient *http.Client
    baseURL    string
}

// NewClient creates a new Gemini client (reads GEMINI_API_KEY env var)
func NewClient() (*Client, error)

// NewClientWithHTTPClient creates a client with injectable HTTP client (for testing)
func NewClientWithHTTPClient(config Config, httpClient *http.Client, baseURL string) *Client

// GenerateLocalizedThumbnail generates a thumbnail with translated text
// Takes original image, source text (tagline), and target language code
// Gemini handles translation + image generation in a single call
// Returns the generated image bytes
func (c *Client) GenerateLocalizedThumbnail(ctx context.Context, imagePath, tagline, targetLang string) ([]byte, error)

// Helper functions
func GetLanguageName(langCode string) string      // Returns full language name
func IsSupportedLanguage(langCode string) bool    // Checks if language is supported
```

#### 2. Thumbnail Service (`internal/thumbnail/service.go`) ✅ IMPLEMENTED
```go
// LocalizeThumbnail generates and saves a localized thumbnail
// Returns the path to the saved thumbnail
func LocalizeThumbnail(ctx context.Context, generator ThumbnailGenerator, video *storage.Video, langCode string) (string, error)

// GetLocalizedThumbnailPath constructs the output path for a localized thumbnail
// e.g., "/path/to/thumbnail.png" + "es" -> "/path/to/thumbnail-es.png"
func GetLocalizedThumbnailPath(originalPath, langCode string) string

// GetOriginalThumbnailPath extracts thumbnail path from video (handles ThumbnailVariants and legacy field)
func GetOriginalThumbnailPath(video *storage.Video) (string, error)

// OpenInDefaultViewer opens a file in the OS default application
// Cross-platform: macOS (open), Linux (xdg-open), Windows (start)
func OpenInDefaultViewer(filePath string) error

// ThumbnailGenerator interface for mocking in tests
type ThumbnailGenerator interface {
    GenerateLocalizedThumbnail(ctx context.Context, imagePath, tagline, targetLang string) ([]byte, error)
}
```

#### 3. YouTube Thumbnail Upload (`internal/publishing/youtube.go`)
```go
// UploadThumbnailForDubbedVideo uploads a thumbnail to a dubbed video
// Uses the video ID from DubbingInfo.UploadedVideoID
func UploadThumbnailForDubbedVideo(video storage.Video, langCode string) error
```

#### 4. Storage Updates (`internal/storage/yaml.go`)
```go
// DubbingInfo - add new field:
type DubbingInfo struct {
    // ... existing fields ...
    ThumbnailPath string `yaml:"thumbnailPath,omitempty" json:"thumbnailPath,omitempty"` // Path to localized thumbnail
}
```

### Prompt Strategy

The key insight: we have the **exact tagline text** from YAML, so the prompt can be very specific. Gemini handles both translation and image generation in a single call - no need for a separate translation step.

```
You are given a YouTube thumbnail image. The image contains the text: "[ENGLISH_TAGLINE]"

Translate that text to [TARGET_LANGUAGE] and replace it in the image.

Keep everything else exactly the same:
- Same colors, fonts, and styling
- Same positioning and layout
- Same background and all other elements
- Only the specified text should change

Generate the modified image.
```

**Why single-step translation**: Gemini can translate the text as part of image generation, eliminating the need for a separate AI call to pre-translate the tagline. This simplifies the architecture and reduces API costs.

### Configuration ✅ IMPLEMENTED

**settings.yaml additions:**
```yaml
gemini:
  # apiKey: use GEMINI_API_KEY env var
  model: "gemini-3-pro-image-preview"  # default (pro for better text quality), or "gemini-2.5-flash-image" for speed
```

### CLI Integration

Add to Dubbing phase menu (after dubbing actions):

```
== Dubbing: My Video ==
Progress: 7/8 complete

[ ] Generate Thumbnail (es)     <- New
[ ] Upload Thumbnail (es)       <- New (only shown if thumbnail exists)
[x] Start Dubbing
[x] Check Status
[x] Translate Metadata
[x] Upload to YouTube
```

### Implementation Phases

**Phase 1: Gemini API Integration**
- Create `internal/thumbnail/` package
- Implement Gemini client with image generation
- Support both model variants (flash vs pro)
- Unit tests with mock HTTP server

**Phase 2: Thumbnail Generation Logic**
- Implement `LocalizeThumbnail()` service function
- Path construction for localized thumbnails
- Read tagline from video YAML
- Save generated image to disk
- Implement `OpenInDefaultViewer()` for cross-platform preview (macOS: `open`, Linux: `xdg-open`, Windows: `start`)
- Unit tests for service layer

**Phase 3: Storage Updates**
- Add `ThumbnailPath` field to `DubbingInfo`
- Update YAML serialization
- Unit tests for storage changes

**Phase 4: CLI Integration**
- Add "Generate Thumbnail" menu option
- Add "Upload Thumbnail" menu option
- Progress counter updates
- Error handling and user feedback

**Phase 5: YouTube Upload Integration**
- Implement `UploadThumbnailForDubbedVideo()`
- Use Spanish channel client for Spanish thumbnails
- Handle authentication for target channel
- Unit tests with mocked YouTube API

**Phase 6: Model Comparison & Validation**
- Test with real thumbnails using both models
- Compare quality: Nano Banana vs Nano Banana Pro
- Document findings and set recommended default
- Validate end-to-end workflow

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| AI doesn't match original exactly | Medium | High | Accept "good enough" for secondary channels; manual override option |
| Text placement/font differs | Medium | Medium | Specific prompt with exact text; test and iterate on prompt |
| Longer translations don't fit | Medium | Medium | AI should handle layout; worst case: manual adjustment |
| Gemini API costs | Low | Low | ~$0.04/image is acceptable; cache results |
| API rate limits | Low | Low | Single thumbnail per video; not high volume |
| Quality varies by thumbnail style | Medium | Medium | Test with variety of real thumbnails; document limitations |

## Dependencies

### Internal
- Video dubbing system (`internal/dubbing/`) - for language codes and DubbingInfo
- YouTube upload (`internal/publishing/youtube.go`) - for thumbnail upload
- Video storage (`internal/storage/yaml.go`) - for Tagline field
- Configuration system (`internal/configuration/`) - for API keys

### External
- Google Gemini API (new)
- YouTube Data API v3 (existing)

## Out of Scope

- Template-based thumbnail generation (no source files available)
- Automatic text detection/OCR (we have tagline in YAML)
- Thumbnail design or creation (agency handles this)
- Multi-text thumbnails (only tagline is tracked)
- ~~Real-time preview in CLI~~ (Replaced by auto-open in default image viewer)

## Validation Strategy

### Testing Approach
- Unit tests for Gemini client (mock HTTP server)
- Unit tests for thumbnail service (mock Gemini client)
- Unit tests for path construction
- Integration tests for YouTube upload (mock API)
- Table-driven tests following project patterns

### Manual Validation
- Test with 5+ real thumbnails of varying styles
- Compare Nano Banana vs Nano Banana Pro quality
- Verify text is legible in generated thumbnails
- Confirm YouTube upload works correctly
- Test with Spanish (primary) and one other language

## Milestones

- [x] **Gemini API Integration Working**: Can generate images via API
- [x] **Thumbnail Generation Functional**: Localized thumbnails saved correctly
- [ ] **Storage Integration Complete**: ThumbnailPath persisted in YAML
- [ ] **CLI Menu Integration**: Generate and Upload options available
- [ ] **YouTube Upload Working**: Thumbnails uploaded to dubbed videos
- [ ] **Model Comparison Complete**: Recommendation documented for default model
- [ ] **End-to-End Workflow Validated**: Full flow tested with real thumbnails

## Progress Log

### 2026-02-04
- **Phase 1 Complete**: Gemini API Integration
  - Created `internal/thumbnail/gemini.go` - Gemini API client (165 lines)
  - Created `internal/thumbnail/gemini_test.go` - Unit tests (36 tests, 92.3% coverage)
  - Added `SettingsGemini` to `internal/configuration/cli.go`
  - Default model: `gemini-3-pro-image-preview` (pro for better text quality)
  - API key via `GEMINI_API_KEY` environment variable
  - Supported languages: es, pt, de, fr, it, ja, ko, zh

- **Design Decision**: Let Gemini handle translation directly
  - **Rationale**: Gemini can translate text as part of image generation, eliminating the need for a separate AI translation call
  - **Impact**: Simpler architecture (one API call instead of two), reduced costs, Gemini has visual context for appropriate translation
  - **Code Impact**: `GenerateLocalizedThumbnail()` signature simplified - no translated tagline parameter needed
  - Updated prompt strategy to use single-step translation approach

- **Design Decision**: Auto-open generated thumbnails for visual verification
  - **Rationale**: User needs to visually inspect generated thumbnail quality before uploading; eliminates manual file navigation
  - **Impact**: Better UX - instant feedback loop after generation
  - **Code Impact**: Added `OpenInDefaultViewer()` function to service layer; cross-platform support (macOS/Linux/Windows)
  - Updated user journey step 5 and "Out of Scope" section (real-time CLI preview no longer needed)

- **Phase 2 Complete**: Thumbnail Generation Logic
  - Created `internal/thumbnail/service.go` - service layer (137 lines)
  - Created `internal/thumbnail/service_test.go` - unit tests (415 lines, 35+ test cases)
  - `LocalizeThumbnail()` orchestrates generation, saving, and returns output path
  - `GetLocalizedThumbnailPath()` constructs language-suffixed paths (e.g., `thumbnail-es.png`)
  - `GetOriginalThumbnailPath()` handles ThumbnailVariants (new) and Thumbnail (legacy) fields
  - `OpenInDefaultViewer()` cross-platform preview (macOS: open, Linux: xdg-open, Windows: cmd /c start)
  - `ThumbnailGenerator` interface enables mocking for tests
  - Test coverage: 90.4% for thumbnail package (57 tests total)

### 2026-01-21
- PRD created
- GitHub issue #366 opened
- Key decisions made:
  - Use Google Gemini API (Nano Banana) for image generation
  - Leverage existing Tagline field for precise text replacement
  - Same language support as dubbing feature
  - Naming convention: `[ORIGINAL]-[lang].[ext]`
  - Test both model variants to determine quality threshold
- Integration points identified:
  - Dubbing menu for CLI integration
  - DubbingInfo struct for storage
  - Existing YouTube thumbnail upload API

---

## Notes

- User is already successfully using Nano Banana manually in AI Studio
- No template/source files available - agency delivers finished images only
- Tagline is the only text tracked in YAML; complex multi-text thumbnails not supported
- "Good enough" quality acceptable for secondary language channels
- ~~Consider caching translations to avoid re-generating same text~~ (Not needed - Gemini translates inline)
- Gemini API supports image input + text prompt for editing workflows
