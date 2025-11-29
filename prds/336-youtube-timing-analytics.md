# PRD: YouTube Publishing Timing Analytics & Optimization

**Issue**: [#336](https://github.com/vfarcic/youtube-automation/issues/336)
**Status**: Design Complete
**Created**: 2025-11-10
**Last Updated**: 2025-11-29
**Priority**: Medium

---

## Problem Statement

The YouTube automation tool currently publishes all videos at the same day of week and time, resulting in **zero variation** in the historical publishing schedule. This makes it impossible to:

1. **Identify optimal publishing windows** - Can't determine if current timing maximizes views/engagement
2. **Adapt to audience behavior** - No data on when target audience is most receptive
3. **Improve video performance** - Missing potential views by not optimizing publish timing
4. **Make data-driven scheduling decisions** - Relying on guesswork rather than evidence

Unlike the title analytics feature which analyzes existing variation in titles, timing analysis faces a **chicken-and-egg problem**: we need varied data to analyze, but haven't been publishing at varied times.

---

## Proposed Solution

A **unified analysis system** with iterative improvement:

### Core Workflow

**Every time user runs analysis** (Menu → Analyze → Timing):

1. **Fetch YouTube Analytics**
   - Audience demographics (locations, viewing patterns)
   - All video performance data (views, CTR, engagement)
   - **Publish dates and times for all videos** (critical for pattern analysis)

2. **AI Analyzes Performance Patterns**
   - Group videos by publish day/time
   - Identify top performers (high views, engagement for their day/time)
   - Identify poor performers (low views, engagement for their day/time)
   - Consider audience timezone and content niche patterns

3. **Generate 6-8 Timing Recommendations** (UTC)
   - **Keep** times that perform well (data-driven)
   - **Replace** poor performers with new alternatives
   - **Add** new times to test if no variation exists yet
   - All times in **UTC format** (e.g., "16:00" = 5pm CET winter)

4. **User Reviews and Saves to settings.yaml**
   - Show recommendations with reasoning
   - User approves → saves to `settings.yaml`
   - Can re-run periodically (quarterly, biannually) to evolve recommendations

5. **Apply Recommendations to Videos**
   - Button in "Initial Details" form: "Apply Random Timing"
   - Picks random recommendation from settings.yaml
   - Applies to **same week** as current date (Monday-Sunday)
   - User sees new date/time and approves before saving

### Key Design Principles

- **Same logic every run**: AI always analyzes performance and generates recommendations
- **Iterative improvement**: Build on success, evolve based on data
- **User control**: Manual review and application, not automatic
- **UTC consistency**: All times in UTC for YouTube API compatibility
- **Same-week scheduling**: Preserve weekly planning boundaries

---

## Goals & Non-Goals

### Goals
✅ **Analyze audience behavior** - Understand when target audience is most active
✅ **Generate timing recommendations** - 6-8 day/time combinations in UTC based on data
✅ **Store recommendations in settings.yaml** - Persistent, reusable timing library
✅ **Easy application to videos** - One-click button to apply random timing from library
✅ **Iterative improvement** - Keep successful times, replace poor performers over time
✅ **Same-week scheduling** - Apply recommendations within current week boundary
✅ **Support re-analysis** - Re-run periodically as more performance data accumulates

### Non-Goals
❌ **Automatic date/time modification** - User manually clicks button and approves
❌ **Real-time scheduling** - No calendar system integration
❌ **A/B testing infrastructure** - Simple performance comparison, not statistical framework
❌ **Multi-channel comparison** - Single channel analysis only
❌ **Timezone conversion UI** - Keep UTC, defer timezone handling to future PRD
❌ **Automatic recommendation generation** - User triggers analysis manually

---

## User Experience

### CLI Workflow: Generate Timing Recommendations

```
Main Menu
├── Analyze
    └── Timing [NEW]
```

**User selects "Analyze → Timing":**

```
Fetching video analytics from YouTube...
✓ Successfully fetched analytics for 127 videos
  - Publish dates/times: Extracted
  - Performance metrics: Views, CTR, engagement
  - Audience data: Top locations, viewing patterns

Analyzing timing patterns with AI...
This may take a moment.

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 Current Publishing Pattern
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  - Monday 16:00 UTC: 121 videos (95%)
  - Tuesday 10:00 UTC: 4 videos (3%)
  - Wednesday 14:00 UTC: 2 videos (2%)

Primary audience: Europe (60%), North America (25%)
Content type: DevOps, cloud-native tutorials

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ Recommended Publish Times (UTC)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. Monday 16:00 UTC
   Reasoning: Current baseline, no negative performance data
   Status: Keep for comparison

2. Tuesday 09:00 UTC
   Reasoning: European workday morning (10-11am CET), high engagement window
   Status: New - test early-week morning

3. Tuesday 15:00 UTC
   Reasoning: US East Coast morning (10-11am EST), workday start
   Status: New - test transatlantic window

4. Thursday 13:00 UTC
   Reasoning: Mid-week afternoon, typically strong B2B engagement
   Status: New - test mid-week slot

5. Thursday 16:00 UTC
   Reasoning: End-of-workday Europe + mid-day US overlap
   Status: New - test overlap period

6. Wednesday 10:00 UTC
   Reasoning: Mid-week morning for global audience
   Status: New - test Wednesday pattern

💾 Save these recommendations to settings.yaml? (y/N): _
```

**If user confirms:**

```
✓ Recommendations saved to settings.yaml
  - 6 timing recommendations stored
  - Use "Apply Random Timing" button when editing videos
  - Re-run this analysis in 3-6 months to evolve recommendations

✓ Files saved:
  - ./tmp/timing-analytics-2025-11-29.json (raw data)
  - ./tmp/timing-recommendations-2025-11-29.md (full analysis)
```

### Video Edit Workflow: Apply Timing

**In "Initial Details" form:**

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Video: Kubernetes Best Practices 2025
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Project Name: kubernetes-demo
Project URL: https://github.com/user/k8s-demo

📅 Date: 2025-12-02T16:00  (Monday 16:00 UTC)

[Apply Random Timing]  ← NEW BUTTON

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**When user clicks "Apply Random Timing":**

```
🎲 Randomly selected: Thursday 13:00 UTC
   (Mid-week afternoon, typically strong B2B engagement)

📅 Original date: Monday 2025-12-02 16:00 UTC
📅 New date:      Thursday 2025-12-05 13:00 UTC
   (Same week: Monday Dec 2 - Sunday Dec 8)

Apply this timing? (y/N): _
```

**If user confirms:**
```
✓ Date updated to 2025-12-05T13:00
  Field updated, ready to save video
```

---

## Technical Design

### Architecture Overview

**Simplified architecture** following title analytics pattern:

```
internal/
├── publishing/
│   ├── youtube_analytics.go         [EXTEND] Add timing data extraction
│   └── youtube_analytics_test.go    [EXTEND]
├── ai/
│   ├── analyze_timing.go            [NEW] Generate timing recommendations
│   ├── analyze_timing_test.go       [NEW]
│   └── templates/
│       └── analyze-timing.md        [NEW] Analysis prompt (UTC-aware)
├── app/
│   ├── menu_handler.go              [EXTEND] Add HandleAnalyzeTiming
│   ├── aspect_forms.go              [EXTEND] Add "Apply Random Timing" button
│   ├── timing_logic.go              [NEW] Apply timing logic
│   └── timing_logic_test.go         [NEW]
└── configuration/
    └── settings.go                  [EXTEND] Add timing recommendations struct

settings.yaml                        [EXTEND] Store recommendations
```

### Settings.yaml Structure

**Add timing recommendations section:**

```yaml
timing:
  recommendations:
    - day: "Monday"
      time: "16:00"  # UTC
      reasoning: "Current baseline, no negative performance data"
    - day: "Tuesday"
      time: "09:00"  # UTC
      reasoning: "European workday morning (10-11am CET), high engagement window"
    - day: "Tuesday"
      time: "15:00"  # UTC
      reasoning: "US East Coast morning (10-11am EST), workday start"
    - day: "Thursday"
      time: "13:00"  # UTC
      reasoning: "Mid-week afternoon, typically strong B2B engagement"
    - day: "Thursday"
      time: "16:00"  # UTC
      reasoning: "End-of-workday Europe + mid-day US overlap"
    - day: "Wednesday"
      time: "10:00"  # UTC
      reasoning: "Mid-week morning for global audience"
```

### Data Structures

**Settings configuration:**

```go
// internal/configuration/settings.go
type TimingRecommendation struct {
    Day       string `yaml:"day" json:"day"`             // "Monday", "Tuesday", etc.
    Time      string `yaml:"time" json:"time"`           // "16:00", "09:00", etc. (UTC)
    Reasoning string `yaml:"reasoning" json:"reasoning"` // Why this slot recommended
}

type TimingConfig struct {
    Recommendations []TimingRecommendation `yaml:"recommendations" json:"recommendations"`
}

type Settings struct {
    // ... existing fields
    Timing TimingConfig `yaml:"timing" json:"timing"`
}
```

**Analytics data (extend existing):**

```go
// internal/publishing/youtube_analytics.go
type VideoAnalytics struct {
    // ... existing fields
    VideoID             string
    Title               string
    Views               int64
    CTR                 float64
    Likes               int64
    Comments            int64
    PublishedAt         time.Time  // Already exists, extract day/time from this

    // NEW: Computed timing fields
    DayOfWeek           string    // "Monday", "Tuesday", etc.
    TimeOfDay           string    // "16:00", "09:00", etc. (UTC hour:minute)
    ViewsPerDay         float64   // Normalized: TotalViews / DaysSincePublish
    EngagementRate      float64   // (Likes + Comments) / Views
}
```

### AI Prompt Template

**Single Unified Template** (`internal/ai/templates/analyze-timing.md`)

```markdown
You are analyzing a YouTube channel's publishing schedule and performance data to generate timing recommendations.

**CRITICAL: All times must be in UTC timezone format (HH:MM, 24-hour).**

## Current Publishing Pattern
{{range .CurrentPattern}}
- {{.DayOfWeek}} {{.TimeOfDay}} UTC: {{.Count}} videos ({{.Percentage}}%)
{{end}}

## Performance Data by Time Slot
{{range .PerformanceBySlot}}
**{{.DayOfWeek}} {{.TimeOfDay}} UTC** ({{.VideoCount}} videos)
- Avg Views: {{.AvgViews}}
- Avg Views/Day: {{.AvgViewsPerDay}}
- Avg CTR: {{.AvgCTR}}%
- Avg Engagement: {{.AvgEngagement}}%
- Performance: {{.Rating}} (excellent/good/average/poor)
{{end}}

## Channel Context
- **Total Videos**: {{.TotalVideos}}
- **Content Type**: DevOps, cloud-native, Kubernetes tutorials
- **Primary Audience Locations**: {{.TopLocations}}
- **Audience Timezone Distribution**: {{.TimezoneBreakdown}}

## Task: Generate 6-8 Timing Recommendations

Your goal is **iterative improvement**: keep what works, replace what doesn't.

### Strategy:
1. **If time slot has data**:
   - Performance excellent/good → **KEEP** in recommendations
   - Performance poor → **REPLACE** with new alternative

2. **If time slot has no data**:
   - Current time is neutral → **KEEP** as baseline
   - Add new times to test based on audience patterns

3. **Output 6-8 total recommendations** with mix of proven and new times

### Audience Behavior Patterns (DevOps/Tech):
- Professional audience checks YouTube during:
  - Work breaks (mid-morning, lunch, mid-afternoon)
  - Commute times (if viewing on mobile)
  - Evening learning sessions
- Workdays (Mon-Fri) typically outperform weekends for B2B content
- First 1-2 hours after publish critical for YouTube algorithm momentum
- European audience (CET = UTC+1): Active 08:00-18:00 CET = 07:00-17:00 UTC
- US East Coast (EST = UTC-5): Active 08:00-18:00 EST = 13:00-23:00 UTC
- Overlap window: 13:00-17:00 UTC hits both audiences' workdays

### Constraints:
- **All times MUST be UTC** (e.g., "16:00", not "5pm CET")
- Focus on workday times (Mon-Fri) unless data shows otherwise
- Spread across different days for variation
- Consider global audience timezone overlaps
- Include times proven to work + new times to test

### Output Format:
For each recommendation (6-8 total), provide:
```json
{
  "day": "Monday",
  "time": "16:00",
  "reasoning": "Current baseline. European end-of-workday (5pm CET) + US mid-day. No negative performance data, keep for comparison."
}
```

Return ONLY valid JSON array of recommendations, no markdown formatting.
```

### Core Functions

**Analysis and Recommendation Generation:**

```go
// internal/ai/analyze_timing.go
func GenerateTimingRecommendations(ctx context.Context, analytics []VideoAnalytics) ([]TimingRecommendation, error)

// internal/publishing/youtube_analytics.go (extend existing)
func EnrichWithTimingData(analytics []VideoAnalytics) []VideoAnalytics {
    // Extract DayOfWeek, TimeOfDay from PublishedAt
    // Calculate ViewsPerDay, EngagementRate
}

func GroupByTimeSlot(analytics []VideoAnalytics) map[string][]VideoAnalytics {
    // Group videos by "Monday 16:00", "Tuesday 09:00", etc.
}
```

**Settings Management:**

```go
// internal/configuration/settings.go
func SaveTimingRecommendations(recommendations []TimingRecommendation) error {
    // Update settings.yaml with new recommendations
}

func LoadTimingRecommendations() ([]TimingRecommendation, error) {
    // Read recommendations from settings.yaml
}
```

**Apply Timing Logic:**

```go
// internal/app/timing_logic.go
func ApplyRandomTiming(currentDate time.Time, recommendations []TimingRecommendation) (time.Time, TimingRecommendation, error) {
    // 1. Pick random recommendation
    // 2. Calculate next occurrence within same week (Mon-Sun)
    // 3. Return new date + selected recommendation
}

func GetWeekBoundaries(date time.Time) (monday, sunday time.Time) {
    // Return Monday and Sunday of the week containing date
}
```

**Menu Handler:**

```go
// internal/app/menu_handler.go
func (h *MenuHandler) HandleAnalyzeTiming(ctx context.Context) error {
    // 1. Fetch YouTube Analytics
    // 2. Enrich with timing data
    // 3. Call AI to generate recommendations
    // 4. Display to user
    // 5. Save to settings.yaml if approved
    // 6. Save analysis files to ./tmp/
}
```

### Video Date Format

**Current format in YAML:**

```yaml
# Example: manuscript/category-02/video.yaml
date: "2025-12-02T16:00"  # UTC (no timezone suffix)
```

**YouTube API interpretation:**
- Treats as UTC when no timezone specified
- `16:00` UTC = 5pm CET (winter) / 6pm CEST (summer)
- Passed directly to YouTube API's `PublishAt` field

**Apply Random Timing:**
- Button updates this field in-place
- Format remains `YYYY-MM-DDTHH:MM`
- User sees change in form, saves video normally

### Menu Integration

**Add "Timing" to Analyze menu:**

```go
// internal/app/menu_handler.go

func (h *MenuHandler) HandleAnalyzeMenu(ctx context.Context) error {
    for {
        choice, err := huh.NewSelect[string]().
            Title("Select Analysis Type").
            Options(
                huh.NewOption("Titles", "titles"),
                huh.NewOption("Timing", "timing"),  // NEW
                huh.NewOption("Back", "back"),
            ).
            Run()

        switch choice {
        case "titles":
            return h.HandleAnalyzeTitles(ctx)
        case "timing":
            return h.HandleAnalyzeTiming(ctx)  // NEW
        case "back":
            return nil
        }
    }
}

func (h *MenuHandler) HandleAnalyzeTiming(ctx context.Context) error {
    // 1. Fetch analytics from YouTube
    // 2. Enrich with timing data (day/time extraction)
    // 3. Group by time slot and calculate performance metrics
    // 4. Call AI to generate 6-8 recommendations
    // 5. Display to user
    // 6. Prompt to save to settings.yaml
    // 7. Save analysis JSON/markdown to ./tmp/
}
```

### Initial Details Form Integration

**Add button below date field:**

```go
// internal/app/aspect_forms.go (where InitialDetails form is defined)

func CreateInitialDetailsForm(video *storage.Video) *huh.Form {
    return huh.NewForm(
        huh.NewGroup(
            // ... existing fields (ProjectName, ProjectURL)

            huh.NewInput().
                Title("Date (UTC)").
                Value(&video.Date).
                Validate(/* date validation */),

            // NEW: Apply Random Timing button
            huh.NewConfirm().
                Title("Apply Random Timing?").
                Description("Pick a random timing recommendation from settings.yaml").
                Affirmative("Apply").
                Negative("Skip").
                Value(&applyTiming),
        ),
    )

    // After form runs, if applyTiming == true:
    // - Call timing_logic.ApplyRandomTiming(video.Date, recommendations)
    // - Update video.Date with result
    // - Show confirmation
}
```

---

## Success Metrics

### Initial Release
- **Recommendation Generation**: AI successfully generates 6-8 timing recommendations based on analytics
- **Settings Storage**: Recommendations persist in settings.yaml correctly
- **Button Functionality**: "Apply Random Timing" button works reliably in Initial Details form
- **Same-Week Logic**: Date calculations respect Monday-Sunday boundaries
- **UTC Consistency**: All times stored and applied in UTC format

### User Adoption
- **Feature Usage**: 50%+ of videos use "Apply Random Timing" button
- **Schedule Diversity**: Achieve 4+ different day/time combinations within 3 months
- **Re-analysis**: User re-runs timing analysis quarterly to evolve recommendations

### Performance Impact
- **Measurable Variation**: Sufficient data to compare performance across time slots (after 3-6 months)
- **Optimization**: 10-15%+ improvement in views/engagement for best-performing slots
- **Data-Driven Decisions**: Recommendations evolve based on actual performance data

---

## Implementation Milestones

### Milestone 1: Settings & Configuration ✅
- [x] Add `TimingRecommendation` and `TimingConfig` structs to `internal/configuration/settings.go`
- [x] Implement `LoadTimingRecommendations()` and `SaveTimingRecommendations()` functions
- [x] Add settings.yaml schema for `timing.recommendations` array
- [x] Add comprehensive unit tests (80% coverage target)

**Validation**: Can read/write timing recommendations from/to settings.yaml ✅
**Completed**: 2025-11-29

### Milestone 2: Analytics Data Extraction ✅
- [x] Verify existing fetcher accepts flexible date ranges (already exists - `GetVideoAnalytics(ctx, startDate, endDate)`)
- [x] Verify `VideoAnalytics` struct has `PublishedAt` field (confirmed - field exists and populated)
- [x] **Implement first-week metrics system** (`GetFirstWeekMetrics()`, `EnrichWithFirstWeekMetrics()`) - Critical for eliminating age bias
- [x] Extend `VideoAnalytics` struct with first-week fields (`FirstWeekViews`, `FirstWeekLikes`, `FirstWeekComments`, `FirstWeekCTR`) and timing fields (`DayOfWeek`, `TimeOfDay`, `FirstWeekEngagement`)
- [x] Implement `EnrichWithTimingData()` to extract day/time from `PublishedAt` and calculate engagement rate
- [x] Implement `GroupByTimeSlot()` with `TimeSlot` struct for video aggregation
- [x] Implement `CalculateTimeSlotPerformance()` for metric aggregation per time slot
- [x] Add comprehensive tests for all timing functions (10+ test functions, all passing)

**Validation**: Can extract timing patterns and first-week performance metrics from YouTube Analytics data ✅
**Completed**: 2025-11-29

**Implementation Notes**: Added first-week metrics system (not in original plan) to eliminate age bias. Uses per-video API queries with `filters="video==ID"` - simple, maintainable approach prioritizing accuracy over optimization. Quarterly analysis frequency makes N API calls acceptable.

### Milestone 3: Generate Recommendations (Backend + CLI) ✅
**Vertical slice: Complete recommendation generation workflow**

- [x] Create `analyze-timing.md` AI prompt template (assumption-free, iterative improvement strategy)
- [x] Implement `GenerateTimingRecommendations()` function
- [x] Parse JSON response from AI into `[]TimingRecommendation`
- [x] Handle edge cases (no data, AI errors, invalid JSON)
- [x] Add comprehensive tests for recommendation generation (80%+ coverage)
- [x] Add "Timing" option to Analyze menu
- [x] Implement `HandleAnalyzeTiming()` menu handler
- [x] Save analysis files to `./tmp/` (complete audit trail: analytics, prompt, response, result)

**Validation**: User can run tool → Analyze → Timing → generate and save 6-8 recommendations to settings.yaml ✅
**Completed**: 2025-11-29 (morning)
**Enhanced**: 2025-11-29 (evening) - Critical bug fixes for production readiness (see work log)

**Implementation Notes**:
- AI prompt uses assumption-free approach (no timezone targeting, no behavioral assumptions)
- Robust JSON parsing with markdown code block extraction
- Strict validation: 6-8 recommendations, valid days, HH:MM format, substantive reasoning
- Test coverage: 89.3% GenerateTimingRecommendations, 100% validation functions
- **Major Refactoring**: Unified analytics save pattern across all analyses (title + timing)
  - Created shared JSON parsing utilities (`json_utils.go`)
  - Refactored both analyses to return (result, prompt, rawResponse, error)
  - Created `SaveCompleteAnalysis()` for audit trail (4 files: analytics, prompt, response, result)
  - Updated title analysis to return structured JSON (TitleAnalysisResult)
  - Both analyses now save: `01-analytics.json`, `02-prompt.md`, `03-ai-response.txt`, `04-result.md`

### Milestone 4: Apply Timing (Backend + Button) ✅
**Vertical slice: Complete timing application workflow**

- [x] Implement `ApplyRandomTiming(currentDate, recommendations)` in `internal/app/timing_logic.go`
- [x] Implement `GetWeekBoundaries()` helper (Monday-Sunday calculation)
- [x] Handle date format conversion (`YYYY-MM-DDTHH:MM`)
- [x] Add tests for week boundary logic and date calculation (80%+ coverage)
- [x] Add "Apply Random Timing" button to Initial Details form
- [x] Wire button to `ApplyRandomTiming()` logic
- [x] Test complete button workflow

**Validation**: User can edit video → Initial Details → click "Apply Random Timing" button → see date change → save ✅
**Completed**: 2025-11-29

### Milestone 5: Documentation
- [ ] Update CLAUDE.md with timing feature architecture
- [ ] **Add "Analytics Integration Pattern" section to CLAUDE.md** documenting shared fetcher approach:
  - Pattern: `fetch once → enrich differently → different AI prompts`
  - Code example showing `GetVideoAnalytics()` reused by multiple analyses
  - Guidelines for adding new analyses (reuse fetcher, add enrichment, create AI prompt)
- [ ] Document settings.yaml timing configuration
- [ ] Add usage examples and best practices
- [ ] Document UTC timezone handling

**Validation**: Users and developers understand how to use timing recommendations and add new analyses following established patterns

---

## Dependencies

### Existing Infrastructure

- **YouTube Analytics API Integration** (`internal/publishing/youtube_analytics.go`)
  - **REUSES existing analytics fetcher** (`GetVideoAnalyticsForLastYear()` or similar)
  - **All analyses share the same fetcher function** - title analysis and timing analysis both use it
  - Fetcher should accept period parameter (e.g., `period: "last_year"`, `"last_quarter"`)
  - Returns comprehensive `VideoAnalytics` struct with all fields
  - **Timing analysis adds NO new fetching logic** - only enrichment functions

- **AI Provider System** (`internal/ai/`)
  - Azure OpenAI and Anthropic already configured
  - Template system supports new prompts
  - Each analysis type has its own AI module with specific prompts

- **Video Storage Layer** (`internal/storage/`)
  - YAML-based video metadata with date fields (UTC format)
  - Phase system (0-7) for workflow tracking

- **CLI Menu System** (`internal/app/menu_handler.go`)
  - Existing Analyze menu to extend
  - Pattern: fetch analytics once → multiple analyses can use same data

### Shared Analytics Pattern

**Key Principle**: One fetcher, multiple enrichers

```go
// Shared by all analyses (already exists)
analytics := publishing.GetVideoAnalytics(ctx, period)

// Title analysis
titleData := ExtractTitleData(analytics)
ai.AnalyzeTitles(titleData)

// Timing analysis
timingData := EnrichWithTimingData(analytics)
ai.GenerateTimingRecommendations(timingData)
```

**What's New for Timing**:
- `EnrichWithTimingData()` - extract day/time from PublishedAt
- `GroupByTimeSlot()` - group videos by publish time
- Timing-specific AI prompt template
- Everything else is reused

### External APIs
- **YouTube Analytics API v2** (already authenticated)
  - Same API calls as title analysis
  - No additional queries needed

- **YouTube Data API v3** (already authenticated)
  - No changes needed

### New Dependencies
- None - uses existing infrastructure

---

## Risks & Mitigations

### Risk 1: User Non-Compliance with Varied Schedule
**Risk**: User doesn't apply suggested date changes, experiment never completes
**Impact**: HIGH - Feature provides no value without data variation
**Mitigation**:
- Make suggestions extremely easy to apply (copy-paste dates)
- Show progress tracking ("5/28 slots tested") to gamify completion
- Integrate into publish workflow so user sees suggestions at the right time
- Provide clear "why this matters" messaging with estimated impact

### Risk 2: Insufficient Sample Size for Statistical Significance
**Risk**: 3 videos per slot may not be enough for meaningful conclusions
**Impact**: MEDIUM - Analysis may produce unreliable recommendations
**Mitigation**:
- AI prompt explicitly mentions sample size limitations
- Display confidence levels with recommendations ("Medium confidence - 3 videos")
- Allow configurable threshold (default 3, can increase to 5+)
- Recommend continued monitoring and re-analysis over time

### Risk 3: Seasonal or External Factors Skewing Results
**Risk**: Holiday periods, news events, or algorithm changes affect performance unrelated to timing
**Impact**: MEDIUM - False conclusions about optimal timing
**Mitigation**:
- AI prompt considers seasonal effects in analysis
- Analysis report notes date range and potential confounding factors
- Recommend long-term monitoring (3-6 months) for stable patterns
- Allow user to exclude outlier periods from analysis

### Risk 4: First-Week Metrics Not Available via API
**Risk**: YouTube Analytics API may not provide time-boxed metrics (first 7 days)
**Impact**: MEDIUM - Have to use cumulative metrics (age bias issue)
**Mitigation**:
- Calculate normalized metrics (views-per-day) as fallback
- Focus on videos published within similar timeframes
- Explicitly note age bias in AI prompt to compensate
- Research if API supports date-range queries per video

### Risk 5: Timezone Confusion
**Risk**: User in different timezone than audience, leading to scheduling errors
**Impact**: LOW - Recommendations may be off by hours
**Mitigation**:
- Always display timezone (EST) explicitly in all suggestions
- Consider fetching channel's primary audience timezone from analytics
- Document timezone handling in user guide
- Allow timezone configuration in settings

---

## Open Questions

### Resolved ✅

1. **Settings Integration**: ✅ RESOLVED - Store in `settings.yaml` for persistent, reusable recommendations
2. **Timezone Handling**: ✅ RESOLVED - Use UTC, defer timezone conversion to future PRD
3. **Two-Phase System**: ✅ RESOLVED - Unified approach, same logic every run
4. **Application Method**: ✅ RESOLVED - Button in Initial Details form, not automatic
5. **Week Boundaries**: ✅ RESOLVED - Monday-Sunday, same week as current date
6. **Recommendation Count**: ✅ RESOLVED - 6-8 timing recommendations

### Remaining

1. **First-Week Metrics API**: Can YouTube Analytics API provide views/CTR for first 7 days after publish, or only cumulative? May need to calculate normalized metrics (views-per-day) as fallback.

2. **API Endpoint**: Should we add `/api/analyze/timing` REST endpoint initially, or CLI-only? Suggest CLI-only for v1, add REST later if needed.

3. **Competitor Analysis**: Feasible to check when similar channels publish? Would require additional YouTube Data API queries. Defer to future enhancement.

---

## Design Decisions

### 2025-11-29: Major Simplification

**Decision**: Simplify from two-phase system to unified approach
- **Rationale**: Two-phase design (experimentation vs. analysis) was overcomplicating the feature. User insight: "I don't think it should matter whether it is the first or the second or the third run"
- **Impact**: Significantly simplified architecture, removed stateful tracking, same logic runs every time
- **Code Impact**: Removed `ExperimentPlan`, `ScheduleSuggestion`, `DetectTimingVariation`, phase-specific handlers

**Decision**: Store recommendations in settings.yaml
- **Rationale**: Persistent, reusable library of timing recommendations that can be applied repeatedly
- **Impact**: Added `TimingConfig` to settings.go, recommendations persist across sessions
- **Code Impact**: New structs in `internal/configuration/settings.go`

**Decision**: Apply via button in Initial Details form
- **Rationale**: User-controlled application, not automatic. Picks random recommendation from settings.yaml
- **Impact**: Non-intrusive, user explicitly chooses when to vary timing
- **Code Impact**: New button in `aspect_forms.go`, calls `ApplyRandomTiming()`

**Decision**: Same-week constraint (Monday-Sunday)
- **Rationale**: Preserve weekly planning, user sets Monday date and knows what day/time it'll publish that week
- **Impact**: Predictable scheduling, no cross-week changes
- **Code Impact**: `GetWeekBoundaries()` function, date calculation logic

**Decision**: Use UTC exclusively
- **Rationale**: Video YAML dates are stored as `T16:00` (no timezone suffix), YouTube API treats as UTC. Keep consistent.
- **Impact**: AI must generate times in UTC, no timezone conversion in this PRD
- **Code Impact**: AI prompt explicitly requires UTC format, all recommendations in UTC

**Decision**: Include publish dates/times in analytics
- **Rationale**: AI needs to see when videos were published to analyze performance patterns
- **Impact**: Critical data for pattern detection
- **Code Impact**: Ensure `PublishedAt` is extracted and passed to AI

**Decision**: Iterative improvement (keep good, replace bad)
- **Rationale**: Build on success rather than complete replacement. User insight: "keep some of the current dates (those that were more successful)"
- **Impact**: Recommendations evolve over time, not thrown away
- **Code Impact**: AI prompt strategy: keep excellent/good performers, replace poor performers

**Decision**: Shared analytics fetcher across all analyses
- **Rationale**: Title analysis and timing analysis need the same base data (views, CTR, publishedAt, etc.). User insight: "all options in Analyze should use the same function... and the major difference is in data we embed in prompts"
- **Impact**: Single source of truth, consistency across analyses, easier to add new analyses
- **Code Impact**:
  - Refactor existing fetcher to accept `period` parameter
  - Timing analysis only adds enrichment functions (`EnrichWithTimingData`, `GroupByTimeSlot`)
  - NO new YouTube API fetching logic needed
  - Pattern: `fetch once → enrich differently → different AI prompts`

---

## Progress Log

### 2025-11-10
- ✅ PRD created with comprehensive two-phase design
- ✅ Defined 5 major implementation milestones
- ✅ Analyzed title analytics architecture
- 📝 Pending: User review

### 2025-11-29 (Morning)
- ✅ Major design simplification based on user feedback
- ✅ Resolved all major design questions
- ✅ Unified approach (same logic every run)
- ✅ Settings.yaml integration designed
- ✅ Button-based application workflow
- ✅ UTC timezone handling clarified
- ✅ Milestones reduced from 5 to 6 (simpler structure)
- ✅ PRD status updated to "Design Complete"
- ✅ **Milestone 1 completed**: Settings & Configuration infrastructure implemented with 83.3% test coverage

### 2025-11-29 (Afternoon): Milestone 2 - Analytics Data Extraction Complete
**Duration**: ~3 hours
**Commits**: Multiple implementation commits
**Primary Focus**: First-week metrics system and timing data extraction

**Completed PRD Items**:
- [x] Extended `VideoAnalytics` struct with first-week performance fields - Evidence: `internal/publishing/youtube_analytics.go:27-37`
- [x] Implemented first-week metrics fetching system - Evidence: `GetFirstWeekMetrics()` (line 202-285), `EnrichWithFirstWeekMetrics()` (line 287-322)
- [x] Implemented timing data extraction - Evidence: `EnrichWithTimingData()` (line 324-357)
- [x] Implemented grouping and aggregation - Evidence: `TimeSlot` struct (line 359-368), `GroupByTimeSlot()` (line 370-392), `TimeSlotPerformance` struct (line 394-402), `CalculateTimeSlotPerformance()` (line 404-445)
- [x] Comprehensive test coverage - Evidence: `internal/publishing/youtube_analytics_test.go` (+330 lines, 10+ test functions, all passing)

**Critical Design Decision Made**:
Implemented first-week metrics system instead of cumulative `ViewsPerDay` to eliminate age bias. Original plan was to use `Views / DaysSincePublish` as normalized metric, but this incorrectly favors newer videos. First-week metrics provide accurate apples-to-apples comparison across all videos regardless of publication date, which is critical for timing analysis since YouTube's algorithm prioritizes early performance.

**Implementation Approach**:
- Uses standard YouTube Analytics API with `filters="video==ID"` parameter (per-video queries)
- Makes N API calls (one per video) - simple, maintainable approach
- Prioritized accuracy and code simplicity over batching optimization
- Quarterly analysis frequency makes performance impact acceptable (user can wait)

**Files Modified**:
- `internal/publishing/youtube_analytics.go`: +286 lines
- `internal/publishing/youtube_analytics_test.go`: +330 lines

**Next Session Priority**: Milestone 3 - AI Recommendation Generation

### 2025-11-29 (Evening): Milestone 3 Complete + Major Refactoring
**Duration**: ~4 hours
**Primary Focus**: Complete timing recommendations + unified analytics pattern

**Completed PRD Items**:
- [x] AI recommendation generation (timing analysis)
- [x] Menu integration (Analyze → Timing)
- [x] HandleAnalyzeTiming() implementation
- [x] Complete audit trail file saving

**Major Refactoring Work** (not in original plan):
- **Unified Analytics Pattern**: Created consistent approach for all analyses
  - `internal/ai/json_utils.go`: Shared JSON parsing utilities
  - Both `AnalyzeTitles()` and `GenerateTimingRecommendations()` now return `(result, prompt, rawResponse, error)`
  - `internal/app/SaveCompleteAnalysis()`: Unified save function for all analyses
  - `internal/app/format_analysis.go`: Markdown formatting helpers
- **Title Analysis Refactor**: Updated to return structured JSON (TitleAnalysisResult)
  - Updated template to request JSON output
  - Created comprehensive struct types for all analysis components
  - Both title and timing analyses now follow identical patterns
- **Complete Audit Trail**: All analyses save 4 files:
  - `01-analytics.json` - Raw YouTube API data
  - `02-prompt.md` - AI prompt sent
  - `03-ai-response.txt` - Raw AI response
  - `04-result.md` - Formatted user-friendly result

**Files Created/Modified**:
- Created: `internal/ai/json_utils.go` (shared JSON parsing)
- Created: `internal/app/format_analysis.go` (markdown formatting)
- Modified: `internal/ai/analyze_titles.go` (refactored to return structured JSON)
- Modified: `internal/ai/analyze_timing.go` (refactored to return prompt/response)
- Modified: `internal/ai/templates/analyze-titles.md` (request JSON output)
- Modified: `internal/app/analytics_files.go` (added SaveCompleteAnalysis)
- Modified: `internal/app/menu_handler.go` (added HandleAnalyzeTiming, updated HandleAnalyzeTitles)
- Modified: Tests (partially updated, need completion)

**Known Issues**:
- Title analysis tests need updating for new JSON return format
- Need end-to-end manual testing

**Next Session Priority**:
1. Fix remaining test failures (title analysis tests)
2. Manual end-to-end testing of both analyses
3. Start Milestone 4 (Apply Random Timing button)

### 2025-11-29 (Late Evening): Critical Bug Fixes & Data Quality Improvements
**Duration**: ~4-5 hours
**Primary Focus**: Post-Milestone 3 bug fixes and real-world validation

**Context**: After Milestone 3 completion, user testing revealed critical data quality issues producing incorrect recommendations with zero metrics.

**Issues Fixed**:
1. ⭐ **Critical: Zero metrics bug**
   - First-week metrics never fetched, all performance data showed 0 in AI prompt
   - Root cause: `EnrichWithFirstWeekMetrics()` not called before timing enrichment
   - Fix: Added metrics fetching in `analyze_timing.go:55-59` with smart skipping for pre-populated test data
   - Impact: Real performance metrics now drive recommendations
   - Evidence: `internal/ai/analyze_timing.go`, `internal/publishing/youtube_analytics.go:307-309`

2. ⭐ **Critical: Test suite blocked**
   - Title analysis tests failing after JSON refactoring (expected string, got struct)
   - Fixed validation functions to expect `TitleAnalysisResult` struct
   - Updated mock responses from markdown to valid JSON
   - Impact: All tests passing, CI unblocked
   - Evidence: `internal/ai/analyze_titles_test.go:148, 159, 267-341`

3. ⭐ **Data quality: Live stream contamination**
   - Live streams have fundamentally different performance (live spike, different CTR)
   - Added `LiveBroadcastContent` field fetching from YouTube API
   - Filter: Skip videos where `LiveBroadcastContent != "none"`
   - Impact: Only regular videos included in analysis
   - Evidence: `internal/publishing/youtube_analytics.go:117, 142, 173-175`

4. ⭐ **Data quality: Historical video contamination**
   - Videos from 2021 appearing in 365-day analysis
   - Root cause: YouTube Analytics API filters by view dates, not publish dates
   - Fix: Manual publish date filtering `metadata.PublishedAt.Before(startDate)`
   - Impact: Only videos published in last 365 days included
   - Evidence: `internal/publishing/youtube_analytics.go:180-184`

5. ⭐ **Data quality: YouTube Shorts contamination**
   - Shorts (≤60s) have completely different audience behavior and algorithm treatment
   - Implemented `isShort()` function with ISO 8601 duration parsing (PT1M30S format)
   - Added `ContentDetails.Duration` fetching from YouTube API
   - Comprehensive test suite (15 test cases covering edge cases)
   - Impact: Clean dataset of regular long-form videos only
   - Evidence: `internal/publishing/youtube_analytics.go:129, 146, 187-189, 476-512`
   - Tests: `internal/publishing/youtube_analytics_test.go:534-570`

6. ⭐ **Analysis quality: Insufficient sample sizes**
   - Time slots with 1-2 videos lack statistical significance
   - Added `filterTimeSlotsByMinVideos()` requiring minimum 3 videos per slot
   - AI only sees statistically meaningful time slots
   - Impact: More reliable recommendations, no noise from sparse data
   - Evidence: `internal/ai/analyze_timing.go:68-70, 140-159`

7. ⭐ **Recommendation diversity: 16:00 UTC clustering**
   - AI recommending 4-5 slots at Monday 16:00 (the "safe" high-data slot)
   - Strengthened prompt constraints with explicit limits:
     - Max 2 recommendations same day
     - Max 2 recommendations same time
     - Min 4 different days
     - Min 12-hour time spread
   - Impact: Better experimental coverage, more actionable insights
   - Evidence: `internal/ai/templates/analyze-timing.md:44-49, 74-76`

**Validation Completed**:
- ✅ All tests passing (`go test ./...` - 25 packages)
- ✅ Build successful (`make build-local`)
- ✅ Real-world user testing with actual YouTube channel data
- ✅ Analytics files show real first-week metrics (no more zeros)
- ✅ Appropriate video count (~50-60 vs 200, matching weekly schedule)
- ✅ Recommendations show good day/time diversity
- ✅ No live streams, Shorts, or old videos in dataset

**Files Modified**:
- `internal/ai/analyze_timing.go` (55-62, 68-70, 140-159)
- `internal/ai/analyze_titles_test.go` (148-225, 267-341)
- `internal/ai/templates/analyze-timing.md` (44-49, 74-76)
- `internal/publishing/youtube_analytics.go` (117, 129, 142, 146, 173-175, 180-184, 187-189, 476-512)
- `internal/publishing/youtube_analytics_test.go` (534-570)

**Impact**: Milestone 3 now produces accurate, actionable recommendations with high-quality data and proper experimental design. Feature is production-ready for real-world use.

**Next Session Priority**: Begin Milestone 4 (Apply Random Timing button implementation)

### 2025-11-29 (Late Evening): Milestone 4 Complete - Apply Random Timing Button
**Duration**: ~2 hours
**Primary Focus**: Complete timing application workflow with UI integration

**Completed PRD Items**:
- [x] Implement `ApplyRandomTiming(currentDate, recommendations)` in `internal/app/timing_logic.go` - Evidence: Core function with random selection, date calculation, same-week logic
- [x] Implement `GetWeekBoundaries()` helper (Monday-Sunday calculation) - Evidence: Week boundary calculation with proper Monday=start, Sunday=end handling
- [x] Handle date format conversion (`YYYY-MM-DDTHH:MM`) - Evidence: Parsing and formatting throughout timing_logic.go
- [x] Add tests for week boundary logic and date calculation (80%+ coverage) - Evidence: `timing_logic_test.go` with 5 test functions covering all scenarios, **96.9% coverage achieved**
- [x] Add "Apply Random Timing" button to Initial Details form - Evidence: `menu_handler.go:847-852` (huh.NewConfirm field)
- [x] Wire button to `ApplyRandomTiming()` logic - Evidence: `menu_handler.go:869-916` (application logic with confirmation display)
- [x] Test complete button workflow - Evidence: Build successful, all tests passing, user tested feature

**User Testing & Enhancements**:
- User tested feature end-to-end and provided feedback
- User requested enhancement: display day of week alongside dates for better clarity
- Enhancement implemented: Output now shows "Monday, 2025-12-02T16:00" instead of just "2025-12-02T16:00"
- User confirmed feature works as expected

**Implementation Highlights**:
- **Clean UX Flow**: User selects "Apply Random Timing? → Yes" → fills form → clicks Save → timing applied → confirmation shown → all changes saved together
- **Smart Fallback**: Shows helpful message if no recommendations exist in settings.yaml ("Run 'Analyze → Timing' to generate recommendations first")
- **Visual Feedback**: Color-coded output with reasoning, week boundaries, and clear before/after comparison
- **Edge Case Handling**: Invalid dates, empty recommendations, parsing errors all handled gracefully
- **Week Boundary Logic**: Robust Monday-Sunday calculation supporting any starting day of week
- **Test Coverage Excellence**: 96.9% coverage exceeds 80% target, includes randomness verification and edge case testing

**Example Output**:
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
🎲 Random timing applied:
   Thursday 13:00 UTC
   Reasoning: Mid-week afternoon, typically strong B2B engagement

📅 Original date: Monday, 2025-12-02T16:00
📅 New date:      Thursday, 2025-12-05T13:00
   (Same week: Monday Dec 2 - Sunday Dec 8, 2025)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

**Files Created**:
- `internal/app/timing_logic.go` (123 lines) - Core timing application logic
- `internal/app/timing_logic_test.go` (306 lines) - Comprehensive test suite

**Files Modified**:
- `internal/app/menu_handler.go` - Added "Apply Random Timing?" button and application logic

**Test Results**:
- ✅ All tests passing (25 packages, go test ./...)
- ✅ Build successful (make build-local)
- ✅ 96.9% test coverage on timing_logic.go
- ✅ User acceptance testing complete

**Next Session Priority**: Begin Milestone 5 (Documentation)

---

## Notes

**Key Innovation**: Iterative improvement model - AI keeps what works and evolves recommendations over time rather than complete replacement.

**Simplified Design Principle**: Same logic every run. No phases, no state tracking, no complexity. Just: analyze → recommend → store → apply.

**⚠️ IMPORTANT - Post-Implementation**:
After completing implementation, update CLAUDE.md with "Analytics Integration Pattern" section documenting the shared fetcher approach. This establishes a reusable pattern for future analyses (thumbnails, descriptions, etc.). Include:
- Code example of shared `GetVideoAnalytics()` function
- Pattern diagram: `fetch once → enrich differently → different AI prompts`
- Guidelines for developers adding new analyses
