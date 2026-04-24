# PRD: AI-Powered Thumbnail Generation with Multi-Provider Support

**Issue**: #396
**Status**: In Progress
**Priority**: High
**Created**: 2026-04-24

---

## Problem Statement

Thumbnails are currently created manually outside the app and uploaded to Google Drive via the Web UI's "Upload to Drive" button. This workflow is time-consuming and doesn't leverage modern AI image generation capabilities. The channel uses a specific posterized stencil-art style with bold text that could be programmatically described and generated.

## Proposed Solution

Add a Web UI workflow where users can generate styled thumbnails using multiple AI image generation providers (Gemini, GPT Image 2, etc.), pick the best result, and have it automatically uploaded to Google Drive as a thumbnail variant. The existing manual "Upload to Drive" button remains for cases where the user has their own thumbnail.

### What Already Exists

- `ThumbnailVariant` struct with `Index`, `Path`, `DriveFileID`, `Share` fields (`internal/storage/yaml.go`)
- `ResolveThumbnail()` and `WithThumbnailFile()` for resolving variants to files (`internal/thumbnail/resolve.go`)
- Google Drive upload via `POST /api/drive/upload/thumbnail/{videoName}` (`internal/api/handlers_drive.go`)
- YouTube thumbnail upload during video publish (`internal/publishing/youtube.go`)
- AI variation prompt generation (text prompts, not images) via `GenerateThumbnailVariations()` (`internal/ai/thumbnails.go`)
- Frontend `ArrayInput` with `FileUploadInput` for manual Drive uploads
- `AIService` interface with text-based AI providers (Anthropic, Azure) (`internal/api/ai_service.go`)
- Video `Tagline` field already populated per video (`internal/storage/yaml.go:79`)
- Reference implementation: `manuscript-thumbnail` skill in `../devops-catalog` using Gemini API with the channel's visual style specification

### What Needs to Be Built

1. **Helm chart ConfigMap**: Move from env-var-only config to structured `settings.yaml` via ConfigMap for new thumbnail generation settings (providers list, photo directory). Existing env var config untouched.
2. **Go configuration**: `ThumbnailGeneration` settings struct with providers list and photo directory
3. **Multi-provider image generation**: `ImageGenerator` interface with Gemini and GPT Image 2 implementations
4. **Prompt builder**: Randomized color/placement selection matching the channel's stencil-art style
5. **Illustration suggestions**: Text AI suggests illustration ideas from manuscript + tagline
6. **In-memory image store**: Temporary storage for generated images before user selection
7. **API endpoints**: Illustration suggestions, thumbnail generation, image download, selection (upload to Drive)
8. **Frontend component**: Two-step UI (suggest illustrations, then generate thumbnails), image grid display, "Use This" selection

### User Journey

**Before**: Create thumbnail externally → Open Web UI → Click "Upload to Drive" → Pick file from laptop → Thumbnail uploaded

**After**:
1. Open Web UI → Navigate to video's thumbnail variants section
2. Click "Suggest Illustrations" → See 3-4 AI-suggested illustration ideas + "None" option → Pick one
3. Click "Generate Thumbnails" → Wait ~30-90s → See grid of generated thumbnails (2 per provider: with/without illustration)
4. Click "Use This" on preferred thumbnail → Automatically uploaded to Google Drive as a variant
5. (Alternative: still use "Upload to Drive" button for manual uploads when preferred)

## Success Criteria

- [ ] Multiple image generation providers configurable via Helm chart values
- [ ] Each provider generates 2 thumbnails (with and without illustration) per generation request
- [ ] All provider calls run concurrently to minimize total wait time
- [ ] Generated thumbnails displayed in Web UI with provider name and style info
- [ ] Selected thumbnail uploaded to Google Drive and saved as a `ThumbnailVariant`
- [ ] Existing manual "Upload to Drive" button remains functional alongside generation
- [ ] Illustration suggestions generated from manuscript + tagline via existing text AI
- [ ] Helm chart uses ConfigMap for structured settings (not env vars for provider list)
- [ ] API keys remain in K8s Secrets (not ConfigMap)
- [ ] Tests passing with 80%+ coverage on new code

## Technical Scope

### Helm Chart Changes

**New `settings` section in `values.yaml`:**
```yaml
settings:
  thumbnailGeneration:
    photoDir: /data/photos
    providers:
      - name: gemini
        model: gemini-2.0-flash-preview-image-generation
      - name: gpt-image
        model: gpt-image-1
```

**New template `configmap-settings.yaml`**: Renders `settings` values into a `settings.yaml` ConfigMap.

**Deployment template update**: Add volume + volumeMount for the ConfigMap at `/app/settings.yaml`.

API keys (`GEMINI_API_KEY`, `OPENAI_API_KEY`) stay as env vars from K8s Secrets via `envFrom`.

### Go Configuration

New structs in `internal/configuration/settings.go`:
```go
type SettingsThumbnailGeneration struct {
    PhotoDir  string                     `yaml:"photoDir"`
    Providers []SettingsThumbnailProvider `yaml:"providers"`
}
type SettingsThumbnailProvider struct {
    Name  string `yaml:"name"`   // "gemini", "gpt-image"
    Model string `yaml:"model"`
}
```

Env var overrides: `GEMINI_API_KEY`, `OPENAI_API_KEY`.

### Image Generation Providers

`ImageGenerator` interface in `internal/thumbnail/generate.go`:
```go
type ImageGenerator interface {
    Name() string
    GenerateImage(ctx context.Context, prompt string, photos [][]byte) ([]byte, error)
}
```

Implementations:
- `GeminiClient` (`internal/thumbnail/gemini.go`) - HTTP client for Gemini REST API
- `GPTImageClient` (`internal/thumbnail/gpt_image.go`) - HTTP client for OpenAI image generation API

### Prompt Builder

`internal/thumbnail/prompt_builder.go` with:
- Channel's color palette (5 backgrounds, text color rules per background)
- 6 person placement options with face direction logic
- Embedded prompt template from the reference skill's style specification
- Random selection of colors/placement per prompt, with illustration section toggled

### API Endpoints

```
POST /api/ai/illustrations/{category}/{name}     → Suggest illustration ideas from manuscript
POST /api/thumbnails/generate                     → Generate thumbnails across all providers
GET  /api/thumbnails/generated/{id}               → Download a generated thumbnail image
POST /api/thumbnails/generated/{id}/select        → Upload selected thumbnail to Drive
```

### Frontend Component

`ThumbnailGenerateButton.tsx` rendered in `DynamicForm.tsx` when `field.fieldName === 'thumbnailVariants'`:
- Step 1: "Suggest Illustrations" → radio buttons with illustration ideas + "None"
- Step 2: "Generate Thumbnails" → loading state ("Generating... may take up to 2 minutes"), button disabled during generation
- Step 3: Image grid grouped by provider, labeled with/without illustration
- "Use This" button per thumbnail → calls select endpoint → uploads to Drive → refreshes video data

### Dependencies

- Gemini API key (for Gemini provider)
- OpenAI API key (for GPT Image 2 provider)
- Existing Google Drive service (for uploading selected thumbnails)
- Existing text AI provider (for illustration suggestions)
- Creator photos placed in configured `photoDir`

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Image generation takes 30-90s per call | Medium | All provider calls run concurrently via goroutines; clear loading UI with time estimate |
| Cost per generation (multiple providers x 2 images) | Medium | Button disabled during generation to prevent double-clicks; generation is on-demand only |
| Generated thumbnails don't match channel style | Medium | Detailed prompt template based on proven reference implementation; iterative prompt refinement |
| Gemini/OpenAI API changes or outages | Low | Provider interface allows easy addition/swapping; individual provider failures don't block others |
| Large image payloads in memory | Low | In-memory store with TTL cleanup; images are transient (selected or discarded within minutes) |

## Milestones

- [x] **Helm chart ConfigMap**: New `settings` section in values.yaml, ConfigMap template, deployment volume mount. Existing env var config untouched. Chart lints and templates correctly.
- [x] **Go configuration**: `ThumbnailGeneration` structs, env var overrides for API keys, loaded from settings.yaml. Tests passing.
- [ ] **Image generation interface and Gemini provider**: `ImageGenerator` interface, `GeminiClient` implementation with HTTP calls, prompt builder with randomized color/placement, embedded prompt template. Tests passing with httptest mock.
- [ ] **GPT Image 2 provider**: `GPTImageClient` implementation. Tests passing with httptest mock.
- [ ] **Illustration suggestions**: `SuggestIllustrations()` in AI package using existing text AI provider, added to `AIService` interface, API endpoint wired. Tests passing.
- [ ] **Thumbnail generation orchestrator and store**: `GenerateThumbnails()` with concurrent multi-provider execution, in-memory `GeneratedImageStore` with cleanup. Tests passing.
- [ ] **API endpoints**: All 4 endpoints (illustrations, generate, download, select) wired in server, including Drive upload on selection. Tests passing.
- [ ] **Server wiring**: Generators initialized from config in `main.go`, image store created and set on server.
- [ ] **Frontend component**: `ThumbnailGenerateButton` with two-step flow (illustrations then generation), image grid, "Use This" selection, loading/error states. Integrated into `DynamicForm` alongside existing upload button.
- [ ] **End-to-end validation**: Full flow works: suggest illustrations → generate thumbnails → pick one → uploaded to Drive → appears as thumbnail variant.

## Progress Log

### 2026-04-25
- Milestone 1 complete: Helm chart ConfigMap — settings.thumbnailGeneration in values.yaml, configmap-settings.yaml template, deployment volume+volumeMount, SETTINGS_FILE env var
- Milestone 2 complete: Go configuration — SettingsThumbnailGeneration/SettingsThumbnailProvider structs, SETTINGS_FILE env var support in InitGlobalSettings(), 8 table-driven tests passing
- Reviewed: integration issue found (settings path mismatch) and fixed via SETTINGS_FILE env var
- Audited: no critical security issues

### 2026-04-24
- PRD created
- GitHub issue #396 opened
