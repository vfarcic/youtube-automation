# PRD: AI Title Generation Improvement via A/B Share Analysis

| Metadata | Details |
|:---|:---|
| **PRD ID** | 344 |
| **Issue** | [#344](https://github.com/vfarcic/youtube-automation/issues/344) |
| **Feature Name** | AI Title Generation Improvement |
| **Status** | In Progress |
| **Priority** | Medium |
| **Author** | @vfarcic |
| **Created** | 2025-11-18 |
| **Updated** | 2026-03-10 |

## 1. Problem Statement

The AI title generator uses static patterns and generic best practices hardcoded in `internal/ai/templates/titles.md`. These patterns (e.g., "Provocative Opinion + Technical Specificity gets 3-5x avg views") are generic guesses, not derived from actual channel data.

Meanwhile, we have the cleanest possible signal sitting unused: **YouTube A/B test share data**. When YouTube tests multiple title variants on the same video, the share percentages directly show which title style kept viewers watching longer — eliminating topic, timing, and age bias entirely.

Combined with first-week YouTube Analytics (views, likes, comments, engagement), the AI can discover which title patterns actually work on this specific channel and update the title generation prompt accordingly.

## 2. Proposed Solution

Replace the current "Analyze > Titles" CLI menu item with a new analysis flow that:

1. **Collects A/B share data** from local video YAML files (current year + previous year index).
2. **Enriches with first-week YouTube Analytics** for absolute performance context.
3. **Sends to AI** for pattern analysis — which title styles win A/B tests and correlate with high first-week performance.
4. **Outputs**:
   - A report saved to `./tmp/title-analysis-prompt.md` (prompt) and `./tmp/title-analysis-{date}/` (results).
   - Recommended `titles.md` content with data-driven patterns.
5. **Applies update** to `titles.md` in the working directory after user approval.

### Why A/B shares are the primary signal

A/B shares compare titles head-to-head on the **same video**. A title with 55% share beat the alternatives regardless of whether the video topic was popular or niche. This is the purest signal for title quality. First-week metrics add absolute context (did the winning title also drive high views/CTR?).

### Why two years of data

The current year's `index.yaml` may have limited data (especially early in the year). Including the previous year via `index/{YEAR-1}.yaml` ensures sufficient sample size while keeping data recent enough to reflect current audience preferences.

## 3. User Stories

* **As a** content creator,
* **I want** the AI to analyze my channel's A/B test results and first-week performance data,
* **So that** it discovers which title styles actually work on my channel, updates the title generation prompt with data-driven patterns, and generates better titles for future videos.

## 4. Data Model

### 4.1. Per-Video Data Point

Each video in the dataset contributes:

| Field | Source | Notes |
|-------|--------|-------|
| Title variants | Video YAML | Each: text + A/B share % |
| Category | Video YAML | e.g., "ai", "kubernetes" |
| Date | Video YAML | Publish date |
| First-week views | YouTube Analytics | Views in days 0-7 |
| First-week likes | YouTube Analytics | Likes in first week |
| First-week comments | YouTube Analytics | Comments in first week |
| First-week engagement | Derived | (likes + comments) / views × 100 |
| Publish day of week | Derived from date | e.g., "Monday" |

### 4.2. Data Filtering

- **Only include videos with A/B data**: Videos must have 2+ title variants where at least one has a non-zero share value.
- Videos without A/B data are excluded — they provide no share signal.
- Videos must have a `videoId` (published videos only).

### 4.3. Data Format in Prompt

Grouped by video, with a legend:

```markdown
## Data Legend
- **A/B test share**: Watch-time share percentage per title variant. Higher share = that title kept viewers watching longer vs other variants in the same test. This is the primary quality signal.
- **First-week metrics** (days 0-7 after publish, eliminates age bias):
  - **views**: Total views in first week
  - **likes**: Total likes in first week
  - **comments**: Total comments in first week
  - **engagement**: (likes + comments) / views × 100

## A/B Test Results

### Video: ai | Monday
First-week: views=15230 | likes=890 | comments=145 | engagement=6.8%
Titles:
- "Why I Changed My Mind About Cursor" (share: 42.1%)
- "Top 10 AI Coding Tools in 2025" (share: 35.5%)
- "AI Coding Is Broken (Here's the Fix)" (share: 22.4%)

### Video: kubernetes | Thursday
First-week: views=9800 | likes=420 | comments=67 | engagement=5.0%
Titles:
- "Stop Using Helm Charts!" (share: 51.2%)
- "Why Helm Is Dead in 2025" (share: 48.8%)
```

### 4.4. Data Sources

- **Current year videos**: `index.yaml`
- **Previous year videos**: `index/{CURRENT_YEAR-1}.yaml`
- **Video details**: Individual YAML files at `manuscript/{category}/{name}.yaml`
- **YouTube Analytics**: `GetVideoAnalytics()` + `EnrichWithFirstWeekMetrics()` + `EnrichWithTimingData()`

### 4.5. Audit Trail

- Prompt saved to `./tmp/title-analysis-prompt.md` before the LLM call (already implemented).
- Results saved via `SaveCompleteAnalysis()` to `./tmp/title-analysis-{date}/`.

## 5. Functional Requirements

### 5.1. Data Collection

- Load video index entries from current year (`index.yaml`) and previous year (`index/{YEAR-1}.yaml`).
- For each video with a `videoId`, read its YAML to extract titles with shares, category, and date.
- **Filter to only videos with A/B data** (2+ title variants, at least one with share > 0).
- Fetch first-week YouTube Analytics and join by video ID.

### 5.2. AI Analysis

- Send the dataset to AI with instructions to:
  - Identify which title styles/patterns consistently win A/B tests (higher share).
  - Cross-reference with first-week metrics to find styles that also drive absolute performance.
  - Produce two outputs:
    1. **Updated `titles.md` content**: Replace the current static patterns with data-driven ones. Keep the same format (numbered patterns with examples, AVOID section, character length guidance). Use actual channel data as evidence.
    2. **Title patterns for `settings.yaml`**: A structured list of high-performing and low-performing patterns with evidence, suitable for storage.

### 5.3. User Approval Flow

After AI analysis:
1. Display the recommended `titles.md` content preview.
2. Ask user: "Update titles.md template?" (Yes/No).
3. Write approved `titles.md` to working directory (data repo).

### 5.4. Graceful Degradation

- If no A/B data is available, inform user and skip analysis.
- If YouTube Analytics are unavailable, analyze with shares only (no first-week metrics).
- If previous year index doesn't exist, use only current year.

## 6. Technical Implementation

### 6.1. New Components

- `internal/ai/title_context.go`: Functions to load videos from multiple index files, filter to A/B data, format the dataset.
- `internal/ai/title_context_test.go`: Tests for data collection and formatting.
- `internal/ai/templates/analyze-titles.md`: Updated template focused on A/B share analysis with structured output (updated `titles.md` content + settings patterns).

### 6.2. Modified Components

- `internal/app/menu_analyze.go`: Replace `HandleAnalyzeTitles()` with new A/B share analysis flow including approval prompts.
- `internal/ai/analyze_titles.go`: Update `AnalyzeTitles()` to accept enriched data (videos + analytics), use new template, return structured result with `titles.md` update and settings patterns.
- `internal/ai/templates/titles.md`: Kept as embedded default template; runtime reads `titles.md` from working directory.
- `internal/ai/titles.go`: `SuggestTitles()` loads template from working directory via `LoadTitlesTemplate()`.

### 6.3. Index File Loading

- Current year: `index.yaml` (relative to data repo root / CWD)
- Previous year: `index/{currentYear - 1}.yaml`
- If previous year file doesn't exist, use only current year.
- Use existing `readArchiveIndex()` pattern from `video_service.go`.

## 7. Milestones

- [x] **Milestone 1: Data Collection**
    - Implement index loading for current + previous year.
    - Implement video loading and filtering to A/B data only.
    - Format dataset with shares + first-week metrics.
    - Unit tests covering: videos with/without A/B data, missing index files, analytics join.

- [x] **Milestone 2: AI Analysis + Template**
    - Update `analyze-titles.md` template to focus on A/B share patterns.
    - AI output: recommended `titles.md` content + structured title patterns.
    - Update `AnalyzeTitles()` to use enriched data and new template.
    - Unit tests for template rendering and response parsing.

- [x] **Milestone 3: Approval Flow + Runtime Template**
    - `SuggestTitles()` reads `titles.md` from working directory at runtime (no longer embedded at compile time).
    - If `titles.md` missing, error with instructions to run analysis or create manually (shows default content).
    - `HandleAnalyzeTitles()` shows proposed `titles.md` preview and prompts user to save.
    - Writes approved `titles.md` to working directory (data repo, not source repo).
    - Analysis template updated: generates 10 titles as JSON array, no rule numbers, pattern-diverse.
    - Unit tests for `LoadTitlesTemplate()` and existing `SuggestTitles()` tests updated.

- [ ] **Milestone 4: Web UI Analyze Section**
    - Add "ANALYZE" section to sidebar below phases with links: Titles, Timing, Sponsor Page.
    - Add API endpoints for each analyze feature (title analysis, timing recommendations, sponsor page update).
    - Build frontend views with "Run Analysis" buttons, loading states, and result display.
    - Title analysis view: show patterns, recommendations, and offer to save `titles.md`.
    - Timing view: show current recommendations, generate new ones, save to `settings.yaml`.
    - Sponsor page view: trigger update, show success/error with file path or PR URL.
    - Related PRDs: #375 (Title Analysis API & UI), #376 (Timing Recommendations API & UI), #378 (Sponsor Page Analytics Update API & UI).

## 8. Success Criteria

- **Data quality**: Only videos with actual A/B share data are included in analysis.
- **Actionable output**: AI produces a concrete `titles.md` replacement, not generic advice.
- **Runtime template**: `titles.md` read from data repo at runtime, not baked into the binary.
- **User control**: Changes to `titles.md` only applied after explicit approval.
- **Audit trail**: Full prompt and response saved to `./tmp/` for inspection.
- **Web UI access**: All three Analyze features accessible from sidebar in the Web UI.

## 9. Dependencies

- `storage.Video` struct with `Titles []TitleVariant` (exists).
- `publishing.GetVideoAnalytics()` and `EnrichWithFirstWeekMetrics()` (exist).
- YouTube Analytics API OAuth credentials (exist in deployment).
- Index file structure: `index.yaml` + `index/{year}.yaml` (exists).
- Timing recommendations pattern in `settings.yaml` (exists as reference).

## 10. Risk Assessment

- **Risk**: Insufficient A/B data (few videos have share values).
    - *Mitigation*: Include previous year. Inform user if sample is too small for meaningful analysis.
- **Risk**: AI produces poor `titles.md` replacement.
    - *Mitigation*: User approval required. Show diff before applying. Original file can be restored from git.
- **Risk**: YouTube Analytics API rate limits during first-week enrichment.
    - *Mitigation*: EnrichWithFirstWeekMetrics makes one call per video — acceptable for ~60 videos. Falls back to shares-only if API fails.
