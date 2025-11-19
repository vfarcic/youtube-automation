# PRD: Thumbnail A/B Test Variation Generator

**Status**: Implementation Complete
**Priority**: High
**GitHub Issue**: [#334](https://github.com/vfarcic/youtube-automation/issues/334)
**Created**: 2025-11-09
**Last Updated**: 2025-11-19

---

## Problem Statement

Thumbnails from the design agency follow a consistent pattern, creating limited variability in performance data. Without testing variations with real audience feedback, we can't learn what works specifically for our viewers. Current process:
- Agency provides one thumbnail per video
- We publish and hope it works
- No systematic way to test alternatives
- Can't validate which visual patterns resonate with our audience

**Result**: Generic guidelines based on industry best practices instead of audience-specific insights.

## Solution Overview

Integrated workflow within the CLI application that:
1. Analyzes a provided thumbnail image (from agency) using AI vision (Anthropic).
2. Generates two strategic variation prompts:
   - **Subtle Refinement**: A/B test of visual hierarchy (shift focus, lighting, layout).
   - **Bold Subject Variation**: Same style, but drastically different subject depiction (e.g., photo vs. drawing).
3. User uses these prompts with an image generator (Midjourney/DALL-E) to create files.
4. User saves file paths in the application.
5. Application manages original + 2 variations for YouTube A/B testing.

**Key Principle**: This creates a **learning feedback loop**. Each video becomes an experiment that teaches us what works for our specific audience.

## User Journey

### Primary Flow: Generate & Save Variations

1. User enters "Edit Video" -> "Post-Production" in the CLI.
2. User inputs the original thumbnail path.
3. User selects **[Generate Variation Prompts (AI)]**.
4. AI analyzes the thumbnail and outputs two prompts:
   - **Subtle**: Describes a variation shifting visual emphasis.
   - **Bold**: Describes the subject in detail for blending with a creator photo (to maintain style but change subject).
5. User copies prompts to image generation AI.
6. User generates images and saves them locally.
7. User inputs the paths for the new "Subtle" and "Bold" thumbnails into the CLI form.
8. CLI saves all three paths to the video's metadata (`.yaml`).

### Secondary Flow: Track Results and Learn

1. YouTube runs A/B test for 7-14 days
2. YouTube declares winner based on CTR performance
3. User notes winning pattern
4. User runs PRD #333 analysis periodically to incorporate learnings
5. Updated guidelines reflect validated patterns
6. Agency receives data-driven instructions based on real tests

### Ongoing Usage

- Run for every new video (or selectively for important videos)
- Each test validates/refutes hypotheses
- Accumulate evidence over time about what works
- Guidelines evolve from "best practices" to "proven with our audience"

## Success Criteria

### Must Have
- [x] CLI accepts thumbnail file path as input
- [x] AI vision analyzes thumbnail and identifies visual characteristics
- [x] Generates 2 distinct, strategic variation prompts (Subtle vs Bold)
- [x] Variations are testable hypotheses (not random changes)
- [x] Output is formatted for easy copy-paste to image AI
- [x] Includes rationale explaining why these variations matter
- [x] Graceful error handling for missing files, unsupported formats, AI failures
- [x] Data persistence for variation file paths

### Nice to Have
- [ ] Support for URL input (not just local file path)
- [ ] Interactive mode: suggest 3-4 variations, user picks 2
- [ ] Reference historical test results (which variations won previously)
- [ ] Generate image AI prompts in multiple formats (Midjourney syntax vs. DALL-E syntax)
- [ ] Save variation prompts + results to `./tmp/ab-tests/` for tracking

### Success Metrics
- Variations are meaningfully different from original (visually testable)
- Variations are grounded in best practices or data (not arbitrary)
- User can execute workflow in <5 minutes per video
- A/B tests produce clear winners (statistically significant CTR differences)

## Technical Architecture

### New Components

```
internal/app/menu_handler.go
└── HandleThumbnailVariations() & editPhasePostProduction updates
    - Orchestrates the interactive workflow

internal/ai/thumbnails.go
├── GenerateThumbnailVariations(ctx, imagePath) → AI vision analysis & prompt generation
└── parseVariationResponse() → Extracts JSON prompt data

internal/ai/templates/thumbnail_variations.md
└── Prompt template defining "Subtle" and "Bold" strategies

internal/storage/yaml.go
└── ThumbnailVariants struct & migration logic for data persistence
```

### Data Flow

```
User (CLI): Selects "Generate Variations"
         ↓
Read image file from path
         ↓
Send to Anthropic Vision:
  - Analyze visual characteristics
  - Apply strategies from template (Subtle Hierarchy Shift vs. Bold Subject Swap)
         ↓
AI returns JSON:
  - Variation 1 Prompt (Subtle)
  - Variation 2 Prompt (Bold)
         ↓
CLI formats and displays prompts
         ↓
User generates images externally
         ↓
User inputs paths into CLI
         ↓
CLI saves paths to video.yaml
```

### Integration Points

1. **AI Provider with Vision** (`internal/ai/provider.go`)
   - Uses Anthropic SDK (Claude 3.5 Sonnet) for vision analysis.

2. **CLI Interface** (`internal/app/`)
   - Integrated into the existing `huh` form workflow.

3. **Storage Layer** (`internal/storage/`)
   - Persists variation paths in `ThumbnailVariants` slice.

### CLI Design

**Location:** Video Edit Menu -> Post-Production Phase

**Workflow:**
1. **Thumbnail Management**: Dedicated sub-loop within the form.
2. **Action Menu**: `[Save & Continue]`, `[Generate Variation Prompts]`.
3. **Output**: Displays prompts clearly with copy-paste support.

## Implementation Milestones

### Milestone 1: CLI Integration (Completed)
**Goal**: Integrate workflow into `editPhasePostProduction`
- [x] Update `menu_handler.go`
- [x] Implement interactive loop for thumbnail management
- [x] Add inputs for Original, Subtle, and Bold paths

### Milestone 2: AI Vision Integration (Completed)
**Goal**: Send thumbnail to AI and get analysis
- [x] Create `thumbnails.go` module
- [x] Implement `GenerateThumbnailVariations` using Anthropic SDK
- [x] Handle base64 image encoding and API response parsing

### Milestone 3: Variation Prompt Generation (Completed)
**Goal**: AI generates 2 strategic variation prompts
- [x] Create `templates/thumbnail_variations.md`
- [x] Define "Subtle" strategy (Visual Hierarchy Shift)
- [x] Define "Bold" strategy (Subject Variation / Photo Blending)

### Milestone 4: Data Persistence (Completed)
**Goal**: Save variation paths
- [x] Update `Video` struct in `yaml.go`
- [x] Add `ThumbnailVariants` slice
- [x] Add auto-migration for legacy `Thumbnail` field

### Milestone 5: Output Formatting (Completed)
**Goal**: Terminal output is clear and actionable
- [x] Use `lipgloss` styles for prompt display
- [x] Clear instructions for next steps (copy-paste)

## Risks & Mitigation

### Risk: Variation Quality
**Impact**: High
**Mitigation**:
- Refined prompt template to ensure "Subtle" changes are meaningful A/B tests.
- Refined "Bold" prompt to support photo-blending for brand consistency.

### Risk: Image Generation AI Compatibility
**Impact**: Low
**Mitigation**:
- Prompts are descriptive and platform-agnostic.
- "Bold" prompt includes full physical description to support diverse generation workflows.

## Future Enhancements

**Phase 2: Result Tracking**
- Save variation prompts to `./tmp/ab-tests/video-{id}.json`
- User inputs A/B test results (winning variation, CTR difference)

**Phase 3: Learning System**
- Analyze historical A/B test results
- Identify consistently winning patterns

**Phase 4: Automation**
- Direct integration with image generation APIs (auto-generate images)
- Direct integration with YouTube API (auto-upload A/B test)

---

## Progress Log

### 2025-11-19 - Implementation Complete
**Status**: 100% Complete

#### ✅ Milestone 1-5: Full Implementation
**Files Created/Modified:**
- `internal/app/menu_handler.go`
- `internal/ai/thumbnails.go`
- `internal/ai/templates/thumbnail_variations.md`
- `internal/storage/yaml.go`

**Implementation Details:**
- Fully integrated into the CLI application structure.
- Leveraged Anthropic SDK for vision capabilities.
- Refined prompts based on user feedback to support photo-blending workflows.