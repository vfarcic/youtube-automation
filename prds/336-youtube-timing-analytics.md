# PRD: YouTube Publishing Timing Analytics & Optimization

**Issue**: [#336](https://github.com/vfarcic/youtube-automation/issues/336)
**Status**: Draft
**Created**: 2025-11-10
**Last Updated**: 2025-11-10
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

A **two-phase system** that first generates experimentation plans, then analyzes results:

### Phase 1: Experimentation Mode (Insufficient Data)
When timing variation is insufficient (< 3 videos per time slot):

1. **Detect lack of variation** - Analyze existing publish times, identify under-tested slots
2. **AI-generated schedule recommendations** - Suggest varied publish times for upcoming videos based on:
   - Target audience timezone/location (from YouTube demographics)
   - Content niche patterns (DevOps/tech publishing norms)
   - YouTube best practices research
   - Competitor analysis (if feasible)
3. **Suggest date modifications** - Recommend specific date/time changes for scheduled-but-unpublished videos
4. **Track progress** - Monitor which time slots have sufficient data (3+ videos minimum)
5. **Integrate with publish workflow** - Show next recommended time slot when publishing

### Phase 2: Analysis Mode (Sufficient Data)
Once each targeted time slot has 3+ videos:

1. **Fetch performance data** - Use YouTube Analytics API (extend existing `GetVideoAnalyticsForLastYear()`)
2. **AI-powered pattern analysis** - Identify optimal publish windows considering:
   - Initial performance velocity (first 7/14 days to avoid age bias)
   - Normalized metrics (views-per-day, engagement rate)
   - Day-of-week patterns
   - Time-of-day patterns (hourly or period-based)
   - CTR, watch time, engagement correlations
3. **Generate recommendations** - Specific, actionable guidance on best publish times
4. **Save analysis files** - JSON data + markdown report (same pattern as title analytics)
5. **Slash command review** - `/analyze-timing` for guided review workflow

---

## Goals & Non-Goals

### Goals
✅ **Detect timing variation deficiency** and alert user
✅ **Generate AI-driven experimentation plans** with 6-8 priority time slots to test
✅ **Track experimentation progress** via video metadata (which slots have 3+ videos)
✅ **Integrate with publish workflow** to show next recommended slot
✅ **Analyze timing patterns** once sufficient data exists
✅ **Provide baseline performance metrics** even with no variation
✅ **Support iterative refinement** - re-run analysis as more data accumulates

### Non-Goals
❌ **Automatic date/time modification** - User manually reviews and applies suggestions
❌ **Real-time scheduling** - No integration with calendar systems
❌ **A/B testing infrastructure** - Not building statistical testing framework
❌ **Multi-channel comparison** - Single channel analysis only
❌ **Audience timezone detection** - Uses existing YouTube Analytics demographics

---

## User Experience

### CLI Workflow: Experimentation Mode

```
Main Menu
├── Analyze
    └── Timing [NEW]
```

**User selects "Analyze → Timing":**

```
Fetching video analytics from YouTube...
✓ Successfully fetched analytics for 127 videos

Analyzing publishing timing patterns...

⚠️  Insufficient Timing Variation Detected
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Current Schedule: 95% of videos published Monday 4:00pm EST
  - Monday 4pm: 121 videos
  - Tuesday 10am: 4 videos
  - Wednesday 2pm: 2 videos

Need 3+ videos per time slot for meaningful analysis.

Generating Experimentation Plan...
✓ AI recommendations complete!

📋 Recommended Time Slots to Test (Next 12 weeks):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Based on DevOps audience patterns and timezone analysis:

Priority 1: Tuesday 10:00am EST (Workday morning, 2/3 videos)
Priority 2: Thursday 2:00pm EST (Mid-week afternoon, 0/3 videos)
Priority 3: Wednesday 9:00am EST (Workday start, 0/3 videos)
Priority 4: Friday 11:00am EST (Pre-weekend, 0/3 videos)
Priority 5: Monday 10:00am EST (Week start, 0/3 videos)
Priority 6: Tuesday 3:00pm EST (Afternoon slot, 0/3 videos)

📝 Suggested Schedule Modifications:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Found 18 scheduled videos (Phase 0-4) that could be rescheduled:

1. "kubernetes-best-practices-2025.yaml" (Phase 2)
   Current: Monday 2025-11-17 16:00:00
   Suggest: Tuesday 2025-11-18 10:00:00 (Priority slot 1)

2. "terraform-vs-pulumi-comparison.yaml" (Phase 1)
   Current: Monday 2025-11-24 16:00:00
   Suggest: Thursday 2025-11-27 14:00:00 (Priority slot 2)

[...16 more suggestions...]

💡 Next Steps:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Review suggested schedule in: ./tmp/timing-experiment-2025-11-10.md
2. Manually update video YAML files with new dates/times
3. As videos publish, run this analysis again to track progress
4. Once 3+ videos per slot, meaningful analysis will be available

✓ Files saved:
  - ./tmp/timing-analytics-2025-11-10.json (baseline data)
  - ./tmp/timing-experiment-2025-11-10.md (schedule recommendations)
```

### CLI Workflow: Analysis Mode (After Experimentation)

```
Fetching video analytics from YouTube...
✓ Successfully fetched analytics for 145 videos

Analyzing publishing timing patterns...
✓ Sufficient variation detected! (6/6 priority slots have 3+ videos)

Analyzing performance patterns with AI...
This may take a moment.
✓ Analysis complete!

📊 Key Findings:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Best Overall: Tuesday 10:00am EST
  - Avg first-week views: 8,247 (vs baseline 5,201)
  - Avg CTR: 8.9% (vs baseline 6.2%)
  - Engagement rate: 12.3% (vs baseline 9.1%)

Top 3 Time Slots:
  1. Tuesday 10:00am EST (+58% views vs baseline)
  2. Wednesday 9:00am EST (+41% views vs baseline)
  3. Thursday 2:00pm EST (+23% views vs baseline)

Worst Performers:
  - Friday 11:00am EST (-18% views vs baseline)
  - Monday 10:00am EST (-12% views vs baseline)

✓ Files saved:
  - ./tmp/timing-analytics-2025-11-10.json (full data)
  - ./tmp/timing-analysis-2025-11-10.md (detailed insights)

💡 Next Steps:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Review detailed analysis: ./tmp/timing-analysis-2025-11-10.md
2. Run /analyze-timing for guided recommendations
3. Update publishing schedule to favor Tuesday 10am slot
4. Continue monitoring with monthly analysis runs
```

### Publish Workflow Integration

When user publishes a video during experimentation phase:

```
Publishing Video to YouTube
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Title: Building Cloud-Native Apps with Kubernetes
Scheduled Publish: Monday 2025-11-17 16:00:00

💡 Timing Experiment Suggestion:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Recommended slot: Thursday 2025-11-20 14:00:00 (Priority 2)
Reason: Testing mid-week afternoon slot (0/3 videos)

Would you like to use the recommended time? (y/N): _
```

---

## Technical Design

### Architecture Overview

Following the **title analytics pattern** with extensions for experimentation:

```
internal/
├── publishing/
│   ├── youtube_analytics.go         [EXTEND] Add timing-specific queries
│   └── youtube_analytics_test.go    [EXTEND]
├── ai/
│   ├── analyze_timing.go            [NEW] Phase 2: Performance analysis
│   ├── analyze_timing_test.go       [NEW]
│   ├── generate_schedule.go         [NEW] Phase 1: Experiment plan
│   ├── generate_schedule_test.go    [NEW]
│   └── templates/
│       ├── analyze-timing.md        [NEW] Analysis prompt
│       └── generate-schedule.md     [NEW] Experiment prompt
├── app/
│   ├── menu_handler.go              [EXTEND] Add HandleAnalyzeTiming
│   ├── analytics_files.go           [EXTEND] Add SaveTimingFiles
│   ├── analytics_files_test.go      [EXTEND]
│   └── timing_suggestions.go        [NEW] Video date modification logic
└── video/
    └── manager.go                   [EXTEND] Add GetScheduledVideos helper

.claude/commands/
└── analyze-timing.md                [NEW] Slash command for review
```

### Data Structures

**Extend existing `VideoAnalytics` struct:**

```go
// internal/publishing/youtube_analytics.go
type VideoAnalytics struct {
    VideoID             string
    Title               string
    Views               int64
    CTR                 float64
    AverageViewDuration float64
    Likes               int64
    Comments            int64
    PublishedAt         time.Time

    // NEW: Timing-specific fields
    DayOfWeek           string    // "Monday", "Tuesday", etc.
    TimeOfDay           string    // "09:00", "14:00", etc. (hour:minute)
    TimePeriod          string    // "morning", "afternoon", "evening", "night"
    FirstWeekViews      int64     // Views in first 7 days
    FirstWeekCTR        float64   // CTR in first 7 days (if available)
    ViewsPerDay         float64   // Normalized: TotalViews / DaysSincePublish
    EngagementRate      float64   // (Likes + Comments) / Views
}
```

**New structs for experimentation:**

```go
// internal/app/timing_suggestions.go
type TimeSlot struct {
    DayOfWeek   string    // "Monday", "Tuesday", etc.
    TimeOfDay   string    // "10:00", "14:00", etc.
    Priority    int       // 1-8 (AI-ranked priority)
    Reasoning   string    // Why this slot was chosen
    VideosCount int       // Current count of videos in this slot
    TargetCount int       // Minimum needed (default 3)
    Status      string    // "needs_testing", "sufficient_data"
}

type ScheduleSuggestion struct {
    VideoPath      string    // Path to video YAML file
    VideoTitle     string    // For display
    CurrentPhase   int       // 0-4 (only suggest for unpublished)
    CurrentDate    time.Time // Existing scheduled date
    SuggestedDate  time.Time // Recommended new date
    RecommendedSlot TimeSlot // Which slot this fills
}

type ExperimentPlan struct {
    GeneratedAt       time.Time
    PrioritySlots     []TimeSlot
    Suggestions       []ScheduleSuggestion
    CurrentCoverage   string              // "3/7 days tested, 2/4 periods tested"
    EstimatedWeeks    int                 // Weeks to complete experiment
}
```

### AI Prompt Templates

**Template 1: Schedule Generation** (`internal/ai/templates/generate-schedule.md`)

```markdown
You are analyzing a YouTube channel's publishing schedule to recommend an experimentation plan for testing varied publish times.

## Current Publishing Pattern
{{range .CurrentPattern}}
- {{.DayOfWeek}} {{.TimeOfDay}}: {{.Count}} videos ({{.Percentage}}%)
{{end}}

## Channel Context
- **Total Videos**: {{.TotalVideos}}
- **Content Type**: {{.ContentType}} (DevOps, cloud-native, tutorials)
- **Target Audience**: {{.AudienceDescription}}
- **Top Viewer Locations**: {{.TopLocations}}
- **Typical Video Length**: {{.AvgDuration}} minutes

## Task
Generate 6-8 priority time slots to test over the next 12 weeks. For each slot, provide:

1. **Day and Time** (EST timezone)
2. **Priority** (1=highest)
3. **Reasoning** (why this slot is promising based on audience/niche)
4. **Expected Impact** (hypothesis about performance)

### Considerations:
- **Audience Timezone**: Most viewers in {{.PrimaryTimezone}}
- **Content Niche**: DevOps professionals typically check YouTube during work breaks, mornings, or evenings
- **YouTube Algorithm**: First 1-2 hours after publish are critical for momentum
- **Workday Patterns**: B2B tech content performs differently than entertainment
- **Weekend vs Weekday**: Consider professional audience behavior

### Constraints:
- Focus on **workday times** (Mon-Fri) unless data suggests otherwise
- Avoid very early (before 8am) or very late (after 8pm) EST
- Spread across different days and times for maximum variation
- Prioritize slots likely to outperform current baseline ({{.BaselineDayTime}})

### Output Format:
Return a prioritized list of time slots with clear, data-driven reasoning.
```

**Template 2: Performance Analysis** (`internal/ai/templates/analyze-timing.md`)

```markdown
You are analyzing YouTube video performance data to identify optimal publishing times.

## Dataset
{{.VideoCount}} videos published between {{.StartDate}} and {{.EndDate}}

{{range .Videos}}
- **{{.Title}}**
  - Published: {{.DayOfWeek}} {{.TimeOfDay}}
  - First Week Views: {{.FirstWeekViews}}
  - Total Views: {{.Views}}
  - CTR: {{.CTR}}%
  - Engagement Rate: {{.EngagementRate}}%
  - Views/Day: {{.ViewsPerDay}}
{{end}}

## Analysis Tasks

### 1. Day-of-Week Patterns
Analyze performance by day of week. Account for:
- **Sample size per day** (statistical significance)
- **Video age bias** (focus on first-week metrics)
- **Content type variations** (if detectable)

### 2. Time-of-Day Patterns
Analyze performance by publish time. Consider:
- **Hour-level granularity** (if sufficient data)
- **Grouped periods** (morning 8-11am, afternoon 12-4pm, evening 5-8pm)
- **Interaction with day-of-week** (e.g., Tuesday morning vs Tuesday afternoon)

### 3. Best vs Worst Performers
Identify:
- **Top 3 time slots** with highest avg first-week views
- **Bottom 3 time slots** to avoid
- **Statistical confidence** (mention sample sizes)

### 4. Engagement Correlations
Check if timing affects:
- **CTR** (does publish time impact click-through?)
- **Watch time** (do certain times get more engaged viewers?)
- **Engagement rate** (likes/comments per view)

### 5. Baseline Comparison
Compare experimental slots against original baseline ({{.BaselineDayTime}}):
- **Percentage improvement/decline**
- **Absolute differences** in key metrics

### 6. Actionable Recommendations
Provide 5-7 specific recommendations:
- **Optimal publish day/time** (primary recommendation)
- **Secondary options** (backup time slots)
- **Times to avoid** (underperformers)
- **Confidence levels** (high/medium/low based on sample size)
- **Next steps** (continue testing, adjust schedule, etc.)

### Important Considerations:
- **Statistical Significance**: Note when sample sizes are too small for strong conclusions
- **Seasonal Effects**: Mention if data spans multiple seasons
- **Content-Specific Patterns**: Highlight if certain content types perform better at specific times
- **Audience Behavior**: Explain findings in context of DevOps/tech professional audience

### Output Format:
Structured markdown with clear sections, data-backed insights, and specific, actionable guidance.
```

### Core Functions

**Experimentation Phase:**

```go
// internal/ai/generate_schedule.go
func GenerateExperimentSchedule(ctx context.Context, analytics []VideoAnalytics, audienceData AudienceContext) (*ExperimentPlan, error)

// internal/app/timing_suggestions.go
func DetectTimingVariation(analytics []VideoAnalytics) (hasVariation bool, coverage TimingCoverage)
func GetScheduledVideos(minPhase, maxPhase int) ([]Video, error)
func GenerateScheduleSuggestions(plan ExperimentPlan, scheduledVideos []Video) ([]ScheduleSuggestion, error)
func FormatExperimentReport(plan ExperimentPlan, suggestions []ScheduleSuggestion) string
```

**Analysis Phase:**

```go
// internal/ai/analyze_timing.go
func AnalyzeTiming(ctx context.Context, analytics []VideoAnalytics) (string, error)

// internal/publishing/youtube_analytics.go (extend existing)
func EnrichWithTimingData(analytics []VideoAnalytics) []VideoAnalytics
func CalculateFirstWeekMetrics(ctx context.Context, videoID string, publishedAt time.Time) (FirstWeekMetrics, error)
```

**Shared Functions:**

```go
// internal/app/analytics_files.go (extend existing)
func SaveTimingAnalysisFiles(analytics []VideoAnalytics, analysis string, isExperiment bool) (AnalysisFiles, error)

// internal/app/menu_handler.go
func (h *MenuHandler) HandleAnalyzeTiming(ctx context.Context) error
```

### Video YAML Integration

**Accessing video date fields:**

```yaml
# Example: data/devops-toolkit/kubernetes-best-practices.yaml
title: "Kubernetes Best Practices 2025"
date: "2025-11-17T16:00:00Z"  # Current scheduled time
phase: 2  # Material Done (unpublished)
# ... other fields
```

**Reading/suggesting modifications:**

1. Find all videos in phases 0-4 (unpublished) using existing `storage` package
2. Parse `date` field from each video's YAML
3. Generate new date/time recommendations based on experiment plan
4. Display suggestions to user (don't auto-modify)
5. User manually edits YAML files with new dates

**Progress tracking approach:**

Instead of maintaining separate state file, calculate coverage dynamically:
1. Fetch all historical video analytics
2. Group by time slot (day + hour)
3. Count videos per slot
4. Identify slots with < 3 videos (need testing)
5. Prioritize these slots in recommendations

This is **stateless** - no need to track "which slots we've tested" separately since video metadata contains publish dates.

### Menu Integration

**Add to existing analyze menu:**

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
    // 1. Fetch analytics
    fmt.Println("Fetching video analytics from YouTube...")
    analytics, err := publishing.GetVideoAnalyticsForLastYear(ctx)
    if err != nil {
        return fmt.Errorf("failed to fetch analytics: %w", err)
    }
    fmt.Printf("✓ Successfully fetched analytics for %d videos\n\n", len(analytics))

    // 2. Enrich with timing data
    analytics = publishing.EnrichWithTimingData(analytics)

    // 3. Check for variation
    hasVariation, coverage := timing_suggestions.DetectTimingVariation(analytics)

    if !hasVariation {
        // EXPERIMENTATION MODE
        return h.handleExperimentationMode(ctx, analytics, coverage)
    } else {
        // ANALYSIS MODE
        return h.handleAnalysisMode(ctx, analytics)
    }
}

func (h *MenuHandler) handleExperimentationMode(ctx context.Context, analytics []VideoAnalytics, coverage TimingCoverage) error {
    fmt.Println("⚠️  Insufficient Timing Variation Detected")
    // Display current pattern...
    // Generate experiment plan...
    // Get scheduled videos...
    // Generate suggestions...
    // Save files...
    // Display next steps...
}

func (h *MenuHandler) handleAnalysisMode(ctx context.Context, analytics []VideoAnalytics) error {
    fmt.Println("✓ Sufficient variation detected!")
    // Run AI analysis...
    // Save files...
    // Display key findings...
}
```

### Integration with Publish Workflow

**Display timing suggestion during publish:**

```go
// internal/app/menu_handler.go (in HandlePublishVideo or similar)

func (h *MenuHandler) showTimingSuggestion(video Video) {
    // Check if experimentation is active
    plan, err := timing_suggestions.GetActiveExperimentPlan()
    if err != nil || plan == nil {
        return // No active experiment
    }

    // Find next priority slot needing videos
    nextSlot := plan.GetNextRecommendedSlot()
    if nextSlot == nil {
        return // All slots have sufficient data
    }

    // Calculate suggested publish date
    suggestedDate := timing_suggestions.CalculateNextDateForSlot(nextSlot, video.Date)

    // Display suggestion
    fmt.Printf("\n💡 Timing Experiment Suggestion:\n")
    fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
    fmt.Printf("Recommended slot: %s %s (Priority %d)\n",
        nextSlot.DayOfWeek, nextSlot.TimeOfDay, nextSlot.Priority)
    fmt.Printf("Reason: %s\n", nextSlot.Reasoning)
    fmt.Printf("\nCurrent: %s\n", video.Date.Format("Monday 2006-01-02 15:04:05"))
    fmt.Printf("Suggested: %s\n", suggestedDate.Format("Monday 2006-01-02 15:04:05"))
    fmt.Printf("\nWould you like to use the recommended time? (y/N): ")

    // Handle user input...
}
```

---

## Success Metrics

### Experimentation Phase
- **Schedule Diversity**: 6+ time slots tested with 3+ videos each within 12 weeks
- **User Adoption**: 70%+ of suggested schedule changes applied by user
- **Data Quality**: All experiment videos have complete analytics data

### Analysis Phase
- **Insight Quality**: AI identifies clear performance differences (>20%) between best/worst slots
- **Actionability**: Recommendations are specific enough to implement immediately
- **Statistical Confidence**: Findings based on sufficient sample sizes (3+ videos per slot minimum)

### Overall Feature Success
- **Performance Improvement**: 15%+ increase in first-week views after adopting optimal timing
- **User Engagement**: Feature used monthly to monitor timing effectiveness
- **Workflow Integration**: 50%+ of video publishes reference timing recommendations

---

## Implementation Milestones

### Milestone 1: Core Timing Analytics Infrastructure
- [ ] Extend `VideoAnalytics` struct with timing fields (day, time, first-week metrics)
- [ ] Implement `EnrichWithTimingData()` to parse publish times into structured fields
- [ ] Add `CalculateFirstWeekMetrics()` for age-normalized analysis
- [ ] Implement `DetectTimingVariation()` to assess data coverage
- [ ] Add comprehensive unit tests (80% coverage target)

**Validation**: Can fetch and enrich video analytics with timing data, detect variation status

### Milestone 2: Experimentation Plan Generation
- [ ] Create `generate-schedule.md` AI prompt template with audience/niche context
- [ ] Implement `GenerateExperimentSchedule()` using AI to suggest 6-8 priority slots
- [ ] Build `GetScheduledVideos()` helper to find unpublished videos (phases 0-4)
- [ ] Implement `GenerateScheduleSuggestions()` to map slots to specific videos
- [ ] Create `FormatExperimentReport()` for user-friendly output
- [ ] Add tests for schedule generation logic

**Validation**: System generates actionable experiment plans with specific video date suggestions

### Milestone 3: Performance Analysis Engine
- [ ] Create `analyze-timing.md` AI prompt template for pattern detection
- [ ] Implement `AnalyzeTiming()` function with statistical analysis requirements
- [ ] Build day-of-week and time-of-day grouping logic
- [ ] Calculate baseline comparisons and performance deltas
- [ ] Add tests for analysis accuracy and edge cases

**Validation**: Given varied timing data, system produces meaningful insights and recommendations

### Milestone 4: CLI Integration & User Experience
- [ ] Add "Timing" option to Analyze menu in `menu_handler.go`
- [ ] Implement `HandleAnalyzeTiming()` with experimentation vs analysis mode logic
- [ ] Create `SaveTimingAnalysisFiles()` for JSON + markdown output
- [ ] Build progress display (coverage stats, key findings summary)
- [ ] Add publish workflow integration to show timing suggestions
- [ ] Test complete CLI workflow end-to-end

**Validation**: Users can run timing analysis via CLI, receive experiment plans or performance insights

### Milestone 5: Documentation & Slash Command
- [ ] Create comprehensive user documentation for timing analysis feature
- [ ] Implement `/analyze-timing` slash command for guided review workflow
- [ ] Add usage examples and best practices guide
- [ ] Document expected experimentation timeline and sample sizes
- [ ] Update CLAUDE.md with timing analysis architecture notes

**Validation**: Users understand how to use feature, run experiments, and interpret results

---

## Dependencies

### Existing Infrastructure
- **YouTube Analytics API Integration** (`internal/publishing/youtube_analytics.go`)
  - Already fetches views, CTR, engagement metrics
  - Needs extension for first-week metrics calculation

- **AI Provider System** (`internal/ai/`)
  - Azure OpenAI and Anthropic already configured
  - Template system supports new prompts

- **Video Storage Layer** (`internal/storage/`)
  - YAML-based video metadata with date fields
  - Phase system (0-7) for workflow tracking

- **CLI Menu System** (`internal/app/menu_handler.go`)
  - Existing analyze menu to extend

### External APIs
- **YouTube Analytics API v2** (already authenticated)
  - May need additional queries for time-range-specific metrics

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

1. **First-Week Metrics API**: Can YouTube Analytics API provide views/CTR for first 7 days after publish, or only cumulative? Need to verify API capabilities.

2. **Audience Timezone Detection**: Should we fetch viewer geography from YouTube Analytics to auto-detect target timezone, or use fixed EST?

3. **Time Slot Granularity**: Start with hour-level (24 slots per day) or broader periods (morning/afternoon/evening/night)? AI can handle either, but user experience differs.

4. **Experiment Duration**: Default to 12 weeks for testing 6-8 slots with 1 video/week. Is this reasonable, or should we recommend 2 videos/week (6 weeks)?

5. **API Endpoint**: Should we add `/api/analyze/timing` REST endpoint initially, or CLI-only like early title analytics?

6. **Settings Integration**: Should optimal timing recommendations be stored in `settings.yaml` for reference, or only in `./tmp/` analysis files?

7. **Competitor Analysis**: Feasible to check when similar channels publish? Would require additional YouTube Data API queries and channel identification logic.

---

## Progress Log

### 2025-11-10
- ✅ PRD created with comprehensive two-phase design (experimentation + analysis)
- ✅ Defined 5 major implementation milestones
- ✅ Analyzed title analytics architecture for consistency patterns
- ✅ Addressed chicken-and-egg problem with experiment plan generation
- 📝 Pending: Verify YouTube Analytics API capabilities for first-week metrics
- 📝 Pending: User review and approval to begin implementation

---

## Notes

This feature represents a **significant evolution** beyond simple analytics reporting (like title analytics) by introducing:

1. **Active experimentation guidance** - System tells user what to test, not just what happened
2. **Stateful workflow awareness** - Integrates with video phases and scheduling
3. **Iterative refinement** - Transitions from "not enough data" to "actionable insights" over time

The chicken-and-egg problem (need variation to analyze, but no variation exists) is solved by making the feature **proactive** rather than purely reactive. This sets a pattern for future analytics features that may need similar bootstrapping.

**Key Design Principle**: Follow title analytics architecture patterns while extending for stateful experimentation workflow.
