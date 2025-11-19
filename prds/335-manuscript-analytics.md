# PRD: Manuscript & Narration Analytics

**Status**: Draft
**Priority**: High
**GitHub Issue**: [#335](https://github.com/vfarcic/youtube-automation/issues/335)
**Created**: 2025-11-09
**Last Updated**: 2025-11-19

---

## Problem Statement

Currently, video manuscripts are written without data-driven insights into which writing patterns, structures, and phrasing lead to better video performance. We don't know:
- Which manuscript lengths correlate with better watch time
- What introduction patterns drive engagement
- Which explanation styles work best for technical content
- How demo-to-explanation ratio affects retention
- Which phrasing patterns resonate with viewers

**Result**: Manuscript-writing slash commands generate content based on generic best practices rather than patterns proven to work with our specific audience.

## Solution Overview

Build an analytics feature that:
1.  **Data Collection**: Fetches historical video performance (views, retention, CTR) and corresponding manuscript text for the last ~50 videos.
2.  **All-In Context Analysis**: Uses a large-context AI model (e.g., Claude 3.5 Sonnet, 200k+ tokens) to process *all* manuscripts and their performance data simultaneously.
3.  **Pattern Discovery**: Identifies correlations between specific writing choices (structure, tone, pacing) and high/low performance without pre-defined biases.
4.  **Actionable Guidelines**: Generates a `manuscript-guidelines.md` file with specific "Do's and Don'ts" to update our writing process.

**Key Principle**: This is a **workflow improvement activity**. We periodically analyze what works, update our writing patterns and slash commands, and all future manuscripts benefit from these insights.

## User Journey

### Primary Flow: Running Manuscript Analysis

1. User launches app and selects new menu option: **Analyze → Manuscripts**
2. App authenticates with YouTube (OAuth for analytics data)
3. App collects data for the last ~50 videos:
    *   Reads `Video` YAML to get `VideoId` and `Gist` (manuscript path).
    *   Reads and **sanitizes** manuscript content (removes `TODO`, `FIXME`, `NOTE`).
    *   Fetches performance metrics (Views, Watch Time, Retention).
4. App constructs a single "Mega-Prompt" containing all 50 manuscripts + their stats.
5. AI analyzes the entire dataset to find patterns separating winners from losers.
6. App saves output to `./tmp/manuscript-guidelines-YYYY-MM-DD.md`.
7. App displays summary in terminal.
8. User exits app.

### Secondary Flow: Post-Upload Manuscript Mapping (Simplified)

*   **Logic**: The system relies on the existing `gist` field in the `Video` YAML file.
*   **Action**: When a video is uploaded, we ensure the `gist` field points to the correct Markdown file. No file renaming is required.

### Tertiary Flow: Improving Slash Commands

1. User reviews `manuscript-guidelines-YYYY-MM-DD.md`
2. Guidelines include: "Intros under 200 words → 15% better retention"
3. User updates manuscript-generating slash command prompts.

## Success Criteria

### Must Have
- [ ] **Data Collector**: Iterate videos, read `gist` path, read `videoid`, fetch YouTube Analytics.
- [ ] **Sanitizer**: Strip `TODO`, `FIXME`, `NOTE` lines from manuscripts before analysis.
- [ ] **Mega-Prompt Builder**: robustly construct a ~150k token prompt with clear delimiters between videos.
- [ ] **AI Integration**: Send the massive context to the AI provider and handle the response.
- [ ] **Output**: Generate a clear, Markdown-formatted guidelines file.
- [ ] **CLI Integration**: "Analyze -> Manuscripts" menu item.

### Nice to Have
- [ ] Frame extraction + AI vision analysis.
- [ ] Automated slash command update suggestions.
- [ ] "Experiment" marking support (manually tagging sections to see if they correlate).

### Success Metrics
- Guidelines include specific, data-driven recommendations.
- The process runs successfully on a batch of 50 videos without context overflow errors.
- Analysis identifies at least 3 distinct patterns correlating with performance.

## Technical Architecture

### New Components

```
internal/analysis/
├── collector.go         # Orchestrates gathering Video structs + Analytics
├── processor.go         # Reads MD files, sanitizes content (removes TODOs)
└── types.go             # Structs for AnalysisInput (Video + content + stats)

internal/ai/
└── manuscript_analysis.go # Builds the Mega-Prompt and calls AI
```

### Data Flow

```
User: Analyze → Manuscripts
         ↓
Collector:
  1. Load all Video YAMLs
  2. Filter for last 50 with valid Gist & VideoID
  3. Fetch YouTube Analytics (Views, Retention) for these IDs
  4. Read & Sanitize Manuscript Files (Processor)
         ↓
AI Service:
  1. Build Prompt:
     "Analyze these 50 scripts.
      Video 1 [Stats: High]: ...text...
      Video 2 [Stats: Low]: ...text..."
  2. Call LLM (high context model)
         ↓
Save: ./tmp/manuscript-guidelines-YYYY-MM-DD.md
```

## Decision Log

### [2025-11-19] Architectural Simplification
*   **Decision**: Use "All-In Context Analysis" instead of extracting specific metrics.
    *   **Rationale**: Claude 3.5 Sonnet has a 200k context window. Feeding raw text allows unbiased pattern discovery (e.g., finding that "humor" works) which feature extraction would miss.
    *   **Impact**: Removes need for complex "Manuscript Parser" with hardcoded metrics. Replaces it with a simple "Read & Sanitize" processor.
*   **Decision**: Simplify Mapping Strategy.
    *   **Rationale**: The `Video` YAML already contains `gist` (path) and `videoid`. Renaming files is unnecessary work.
    *   **Impact**: Dropped "File Renaming" milestone.
*   **Decision**: Data Sanitization.
    *   **Rationale**: Editor notes (`TODO`, `FIXME`) are noise. Titles are signal.
    *   **Impact**: Added sanitization step and Title injection.

## Implementation Milestones

### Milestone 1: Core Logic & Sanitization
**Goal**: Build the processor that prepares data for the AI.
- Create `internal/ai/manuscript_analysis.go`
- Implement `SanitizeManuscript(content string) string` (Strip TODO/FIXME/NOTE)
- Unit tests for sanitization.

### Milestone 2: Data Collection
**Goal**: Gather the raw materials (Stats + Text).
- Create `internal/analysis/collector.go`
- Implement logic to find last N videos.
- Integrate with YouTube Analytics to fetch metrics.
- Read and sanitize manuscript files.

### Milestone 3: AI Analysis
**Goal**: The "Brain" of the operation.
- Implement `AnalyzeManuscripts(inputs []AnalysisInput)`
- Construct the prompt.
- Wire up the AI provider call.
- Save output to file.

### Milestone 4: CLI Integration
**Goal**: Make it usable.
- Add to `internal/app` menu.
- End-to-end verification.

---

## Progress Log

### [2025-11-19] - PRD Updated
**Status**: Architecture pivoted to High-Context AI Analysis.