# PRD: Thumbnail A/B Test Variation Generator

**Status**: Draft
**Priority**: High
**GitHub Issue**: [#334](https://github.com/vfarcic/youtube-automation/issues/334)
**Created**: 2025-11-09
**Last Updated**: 2025-11-09

---

## Problem Statement

Thumbnails from the design agency follow a consistent pattern, creating limited variability in performance data. Without testing variations with real audience feedback, we can't learn what works specifically for our viewers. Current process:
- Agency provides one thumbnail per video
- We publish and hope it works
- No systematic way to test alternatives
- Can't validate which visual patterns resonate with our audience

**Result**: Generic guidelines based on industry best practices instead of audience-specific insights.

## Solution Overview

Create a slash command that:
1. Analyzes a provided thumbnail image (from agency) using AI vision
2. Generates two variation prompts based on best practices and strategic hypotheses
3. User feeds prompts to image generation AI (Midjourney/DALL-E/Stable Diffusion)
4. User uploads original + 2 variations to YouTube A/B testing
5. YouTube tracks performance and declares winner
6. Results inform future agency guidelines (feeds into PRD #333)

**Key Principle**: This creates a **learning feedback loop**. Each video becomes an experiment that teaches us what works for our specific audience.

## User Journey

### Primary Flow: Generate Variation Prompts

1. User receives thumbnail from agency for upcoming video
2. User saves thumbnail locally (e.g., `./tmp/video-thumbnail.jpg`)
3. User runs slash command: `/thumbnail-variations ./tmp/video-thumbnail.jpg`
4. Claude Code:
   - Reads the thumbnail image
   - Analyzes visual characteristics (text, colors, composition)
   - References historical guidelines (from PRD #333 if available)
   - Generates 2 strategic variation prompts
5. Claude Code outputs:
   - **Analysis**: Description of original thumbnail characteristics
   - **Variation A Prompt**: First hypothesis to test (e.g., "Increase text contrast")
   - **Variation B Prompt**: Second hypothesis to test (e.g., "Simplify composition")
   - **Rationale**: Why these variations are worth testing
6. User copies prompts to image generation AI (Midjourney, DALL-E, etc.)
7. User generates 2 variations
8. User uploads all 3 thumbnails to YouTube A/B test feature

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
- [ ] Slash command accepts thumbnail file path as input
- [ ] AI vision analyzes thumbnail and identifies visual characteristics
- [ ] Generates 2 distinct, strategic variation prompts
- [ ] Variations are testable hypotheses (not random changes)
- [ ] Output is formatted for easy copy-paste to image AI
- [ ] Includes rationale explaining why these variations matter
- [ ] Graceful error handling for missing files, unsupported formats, AI failures

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
.claude/commands/thumbnail-variations.md
└── Slash command that orchestrates variation generation

internal/ai/thumbnail_variations.go
├── AnalyzeThumbnailForVariations(imagePath) → AI vision analysis
├── GenerateVariationPrompts(analysis, guidelines) → Create 2 prompts
└── FormatVariationOutput() → Structure output for user

internal/ai/templates/thumbnail-variations.md
└── Prompt template for AI to generate variation ideas
```

### Data Flow

```
User: /thumbnail-variations ./tmp/thumbnail.jpg
         ↓
Read image file from path
         ↓
Send to AI Vision (Claude/GPT-4V):
  - Analyze visual characteristics
  - Reference guidelines from PRD #333 (if available)
  - Generate strategic hypotheses
         ↓
AI returns:
  - Original analysis
  - Variation A prompt + rationale
  - Variation B prompt + rationale
         ↓
Format output for terminal display
         ↓
User copies prompts to Midjourney/DALL-E
         ↓
User generates variations
         ↓
User uploads to YouTube A/B test
         ↓
(Future: Track results in ./tmp/ab-tests/)
```

### Integration Points

1. **AI Provider with Vision** (`internal/ai/provider.go`)
   - Use Claude 3.5 Sonnet or GPT-4 Vision
   - Send thumbnail image for analysis
   - Similar pattern to PRD #333 vision analysis

2. **File System** (`./tmp` directory)
   - Read thumbnail images from user-specified path
   - (Future) Save variation prompts and results for tracking

3. **Guidelines Reference** (Optional)
   - If `./tmp/thumbnail-guidelines-*.md` exists (from PRD #333)
   - Include guidelines context in AI prompt
   - Variations aligned with strategic insights

### Slash Command Design

**Command Syntax:**
```bash
/thumbnail-variations <image-path>
```

**Examples:**
```bash
/thumbnail-variations ./tmp/video-123-thumbnail.jpg
/thumbnail-variations ~/Downloads/thumbnail.png
```

**Output Format:**
```markdown
## Thumbnail Variation Analysis

### Original Thumbnail Analysis
- **Text**: 4 words, high contrast, top-center positioning
- **Colors**: Dark blue background, yellow text, high contrast
- **Composition**: Face visible, simple layout
- **Complexity**: Medium (3 visual elements)

### Variation A: Increase Text Boldness
**Hypothesis**: Bolder, larger text may improve readability on mobile devices

**Image Generation Prompt:**
"Create a YouTube thumbnail with a dark blue background. Feature a professional headshot in the lower left corner. Add bold, extra-large yellow text reading '[YOUR TEXT]' positioned at the top center. High contrast, simple composition, professional DevOps aesthetic."

**Rationale**: Analysis shows top performers in DevOps niche use larger, bolder text. Testing if this improves mobile CTR.

### Variation B: Simplify Composition
**Hypothesis**: Removing secondary visual elements may improve focus and CTR

**Image Generation Prompt:**
"Create a minimalist YouTube thumbnail with a solid dark blue background. Feature only large, bold yellow text reading '[YOUR TEXT]' centered on the image. No other elements. High contrast, clean, professional aesthetic."

**Rationale**: Competitor analysis suggests simpler thumbnails outperform busy layouts. Testing minimal approach.

---

**Next Steps:**
1. Copy prompts to Midjourney/DALL-E
2. Generate 2 variations
3. Upload original + 2 variations to YouTube A/B test
4. Track winning pattern for future guidelines
```

## Implementation Milestones

### Milestone 1: Slash Command Foundation
**Goal**: Slash command can accept file path and read image

- Create `.claude/commands/thumbnail-variations.md`
- Parse command arguments (file path)
- Validate file exists and is supported format (jpg, png, webp)
- Read image file into memory
- Error handling for missing/invalid files
- Basic test with sample thumbnail

**Validation**: Command accepts file path and loads image successfully

---

### Milestone 2: AI Vision Integration
**Goal**: Send thumbnail to AI and get analysis

- Create `thumbnail_variations.go` module
- Implement image-to-AI pipeline (Claude Vision or GPT-4V)
- Extract visual characteristics: text, colors, composition, complexity
- Return structured analysis data
- Handle AI API failures gracefully
- Test with various thumbnail styles

**Validation**: AI correctly identifies visual characteristics of thumbnails

---

### Milestone 3: Variation Prompt Generation
**Goal**: AI generates 2 strategic variation prompts

- Create prompt template for variation generation
- Implement `GenerateVariationPrompts()` function
- Ensure variations are meaningfully different
- Ensure variations are testable hypotheses (grounded in strategy)
- Format prompts for image generation AI
- Test with multiple thumbnail styles

**Validation**: AI generates distinct, strategic variations with rationale

---

### Milestone 4: Guidelines Integration (Optional)
**Goal**: Reference existing guidelines from PRD #333

- Check if `./tmp/thumbnail-guidelines-*.md` exists
- Parse guidelines for key insights
- Include guidelines context in AI prompt
- Align variations with strategic insights
- Fallback gracefully if no guidelines exist

**Validation**: Variations reference specific guideline insights when available

---

### Milestone 5: Output Formatting
**Goal**: Terminal output is clear, actionable, copy-paste friendly

- Format analysis section (bullet points, clear labels)
- Format variation prompts (ready for Midjourney/DALL-E)
- Include rationale for each variation
- Add next steps instructions
- Test readability in terminal

**Validation**: User can quickly understand and use output

---

### Milestone 6: Testing & Refinement
**Goal**: Feature works reliably with various inputs

- Test with different image formats (jpg, png, webp)
- Test with different thumbnail styles (text-heavy, minimal, complex)
- Validate variation quality (meaningfully different, strategic)
- Error handling for edge cases
- Performance optimization (image loading, AI tokens)

**Validation**: Feature works reliably across different use cases

---

### Milestone 7: Production Ready
**Goal**: Feature is stable and ready for regular use

- Comprehensive error handling
- Clear error messages for common issues
- Documentation in slash command file
- (Optional) Example workflow in CLAUDE.md
- Final end-to-end testing

**Validation**: Feature works reliably for every video

---

## Dependencies

### External
- AI Provider with vision capabilities (Claude 3.5 Sonnet or GPT-4 Vision)
- YouTube A/B testing feature (user's responsibility to use)
- Image generation AI (Midjourney/DALL-E - user's responsibility)

### Internal
- Existing AI provider in `internal/ai/provider.go`
- File system utilities for reading images
- (Optional) Guidelines from PRD #333

### User Requirements
- Access to image generation AI (Midjourney, DALL-E, Stable Diffusion, etc.)
- YouTube channel with A/B testing enabled (available to most channels)
- Thumbnails from design agency (or ability to create originals)

## Risks & Mitigation

### Risk: Variation Quality
**Impact**: High
**Probability**: Medium
**Mitigation**:
- Iterate on prompt design to ensure strategic variations
- Test with real thumbnails before launch
- Include rationale so user can judge quality
- Fallback: User can re-run with refined guidance

### Risk: Image Generation AI Compatibility
**Impact**: Low
**Probability**: Low
**Mitigation**:
- Generate generic prompts that work across platforms
- (Nice to have) Support multiple prompt formats
- Document which AI tools work best

### Risk: File Format Support
**Impact**: Low
**Probability**: Low
**Mitigation**:
- Support common formats (jpg, png, webp)
- Clear error messages for unsupported formats
- Document supported formats in slash command

### Risk: A/B Test Interpretation
**Impact**: Medium
**Probability**: Medium
**Mitigation**:
- Include guidance on statistical significance
- Recommend minimum test duration (7-14 days)
- (Future) Track and summarize results in tool

## Open Questions

1. **Image generation AI preference**: Should we optimize prompts for specific platform (Midjourney vs. DALL-E)?
   - **Decision**: Generic prompts for v1, platform-specific as nice-to-have

2. **Number of variations**: Why 2 variations instead of 3-5?
   - **Decision**: 2 variations + original = 3 total (YouTube A/B test limit for most channels)

3. **Guidelines reference**: Should this be required or optional?
   - **Decision**: Optional. Works standalone, enhanced if PRD #333 analysis exists

4. **Result tracking**: Should we build tracking into this PRD?
   - **Decision**: Manual tracking for v1, automated tracking as future enhancement

5. **Interactive mode**: Should users pick from multiple suggested variations?
   - **Decision**: Auto-generate 2 strategic variations for v1, interactive as nice-to-have

## Future Enhancements

**Phase 2: Result Tracking**
- Save variation prompts to `./tmp/ab-tests/video-{id}.json`
- User inputs A/B test results (winning variation, CTR difference)
- Build database of validated patterns
- Feed results back into PRD #333 guidelines

**Phase 3: Learning System**
- Analyze historical A/B test results
- Identify consistently winning patterns
- Prioritize variations that have high success probability
- Recommend when to stop testing (pattern validated)

**Phase 4: Automation**
- Direct integration with image generation APIs (auto-generate variations)
- Direct integration with YouTube API (auto-upload A/B test)
- Automated result collection
- Closed-loop learning system

**Phase 5: Advanced Variations**
- Generate 3-5 variations, user picks best 2
- Test multiple hypotheses simultaneously
- Statistical analysis of cumulative results
- Confidence scoring for recommendations

---

## Progress Log

### [Date] - Session [N]: [Milestone] Complete
**Duration**: ~X hours
**Status**: X of 7 milestones complete (X%)

#### ✅ Milestone [N]: [Name] (100%)
**Files Created:**
- [List files]

**Implementation Details:**
- [Key implementation notes]

**Testing & Validation:**
- [Test results]

**Technical Decisions Made:**
- [Decisions and rationale]

---

## Cross-References

**Related PRDs:**
- **PRD #333**: Thumbnail Analytics & Competitive Benchmarking (in progress) - Provides strategic guidelines
- **PRD #331**: YouTube Title Analytics & Optimization (completed) - Similar slash command pattern

**Integration Flow:**
```
PRD #333 (Analytics) → Strategic Guidelines
         ↓
PRD #334 (Variation Generator) → A/B Testing
         ↓
YouTube A/B Test Results → Validation
         ↓
PRD #333 (Updated Guidelines) → Improved Instructions
```

---

## References

- [YouTube A/B Testing Feature](https://support.google.com/youtube/answer/11364882)
- [Claude Vision Capabilities](https://docs.anthropic.com/claude/docs/vision)
- [Midjourney Prompt Guide](https://docs.midjourney.com/docs/prompts)
- [DALL-E Prompt Guide](https://platform.openai.com/docs/guides/images)
- [PRD #333 - Thumbnail Analytics](prds/333-thumbnail-analytics.md) - Complementary feature
- [PRD #331 - Title Analytics](prds/done/331-youtube-title-analytics.md) - Reference slash command pattern
- [GitHub Issue #334](https://github.com/vfarcic/youtube-automation/issues/334)
