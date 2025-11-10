# PRD: YouTube Title Analytics & Optimization

**Status**: Draft
**Priority**: High
**GitHub Issue**: [#331](https://github.com/vfarcic/youtube-automation/issues/331)
**Created**: 2025-11-09
**Last Updated**: 2025-11-09

---

## Problem Statement

Currently, the YouTube automation tool generates video titles without data-driven insights into what actually works. We don't know:
- Which title patterns generate the highest CTR (click-through rate)
- What types of titles drive more views and engagement
- Which title structures perform better or worse
- How our titles compare to channel averages

This lack of insights means we're generating titles blindly, potentially missing opportunities to optimize for better performance and channel growth.

## Solution Overview

Add an analytics feature that:
1. Fetches historical video performance data from YouTube Analytics API
2. Uses AI to analyze title patterns and identify what works
3. Saves raw data and analysis to files for review
4. Enables iterative improvement of title generation prompts based on real data

**Key Principle**: Analysis is a **development activity**, not a runtime feature. We periodically analyze data, update the title generation prompt in code, and all future titles benefit automatically.

## User Journey

### Primary Flow: Running Analysis

1. User launches app and selects new menu option: **Analyze → Titles**
2. App authenticates with YouTube (OAuth, may require re-auth for analytics scope)
3. App fetches last 365 days of video performance from YouTube Analytics API
4. AI analyzes the data for patterns:
   - High-performing title structures
   - Power words that increase CTR
   - Optimal title length
   - Topic/keyword performance
5. App saves two files to `./tmp`:
   - `youtube-analytics-YYYY-MM-DD.json` (raw data)
   - `title-analysis-YYYY-MM-DD.md` (AI recommendations)
6. App displays summary in terminal
7. User exits app

### Secondary Flow: Improving Title Generation

1. Developer opens `./tmp/title-analysis-YYYY-MM-DD.md` in Claude Code
2. Developer (optionally) runs slash command `/analyze-titles` for guided review
3. Developer reviews AI recommendations (e.g., "Add more number-based titles")
4. Developer updates prompt in `internal/ai/titles.go` to incorporate learnings
5. Developer commits the improved prompt
6. All future title generations automatically use improved patterns

### Ongoing Usage

- User creates new videos → Generate titles → Uses latest improved prompt
- No need to re-run analysis every time
- Analysis run periodically (monthly/quarterly) to refine further

## Success Criteria

### Must Have
- [x] Successfully fetch video analytics from YouTube API (last 365 days)
- [x] AI generates actionable recommendations about title patterns
- [x] Raw data saved as JSON in `./tmp`
- [x] Analysis saved as Markdown in `./tmp`
- [x] New "Analyze" menu option with "Titles" sub-menu works
- [x] Graceful error handling for API failures/quota limits

### Nice to Have
- [x] Slash command to review and suggest prompt changes
- [ ] Comparison over time (track how improvements affect performance)
- [ ] Analysis of thumbnail/title combinations

### Success Metrics
- Can identify high-performing vs low-performing title patterns
- Recommendations are specific and actionable
- Title generation quality improves after applying insights

## Technical Architecture

### New Components

```
internal/publishing/youtube_analytics.go
├── GetVideoAnalytics(startDate, endDate) → Fetch from API
├── VideoAnalytics struct (videoID, title, views, CTR, etc.)
└── GetVideoAnalyticsForLastYear() → Convenience method

internal/ai/analyze_titles.go
├── AnalyzeTitles(videoData) → AI analysis
└── FormatAnalysisPrompt() → Structure data for AI

internal/app/
├── Add "Analyze" menu option
└── Add "Titles" sub-menu handler

.claude/commands/analyze-titles
└── Slash command for reviewing saved analysis
```

### Data Flow

```
User: Analyze → Titles
         ↓
YouTube Analytics API
         ↓
Parse & Structure Data (VideoAnalytics[])
         ↓
AI Analysis (similar to tag generation)
         ↓
Save Files:
  - ./tmp/youtube-analytics-2025-11-09.json
  - ./tmp/title-analysis-2025-11-09.md
         ↓
Display Summary in Terminal
```

### Integration Points

1. **YouTube API Client** (`internal/publishing/youtube.go`)
   - Extend existing OAuth flow
   - Add new scope: `yt-analytics.readonly`
   - Reuse token caching mechanism

2. **AI Provider** (`internal/ai/provider.go`)
   - Use existing AI provider (Azure/Anthropic)
   - Similar pattern to tag/title generation

3. **App Menu** (`internal/app/app.go`)
   - Add new root menu option "Analyze"
   - Sub-menu for "Titles" (extensible for future analytics)

4. **File System** (`./tmp` directory)
   - Already gitignored
   - Store analysis files with timestamps

## Implementation Milestones

### Milestone 1: YouTube Analytics API Integration
**Goal**: Fetch video performance data from YouTube

- Extend OAuth client to support analytics scope
- Implement `GetVideoAnalytics()` function
- Parse API response into structured `VideoAnalytics` data
- Handle API errors and quota limits gracefully
- Add unit tests with mocked API responses

**Validation**: Can fetch and parse analytics data for last 365 days

---

### Milestone 2: AI Title Analysis Engine
**Goal**: AI can analyze patterns and generate recommendations

- Create `AnalyzeTitles()` function in `internal/ai/`
- Design prompt to extract meaningful patterns
- Structure AI response as actionable recommendations
- Test with sample data from real channel
- Ensure recommendations are specific (not generic advice)

**Validation**: AI generates concrete, data-backed recommendations

---

### Milestone 3: Menu Integration & UX
**Goal**: User can trigger analysis from app menu

- Add "Analyze" root menu option
- Add "Titles" sub-menu under Analyze
- Wire analytics flow through app layer
- Display progress indicators during API fetch
- Show summary after analysis completes
- Handle OAuth re-authentication flow

**Validation**: User can run analysis end-to-end from menu

---

### Milestone 4: File Output & Persistence
**Goal**: Analysis results saved to files for review

- Save raw JSON data to `./tmp/youtube-analytics-{date}.json`
- Save Markdown analysis to `./tmp/title-analysis-{date}.md`
- Format Markdown for readability (headers, lists, examples)
- Include metadata (date, video count, date range)
- Display file paths in terminal after saving

**Validation**: Files created with correct format and content

---

### Milestone 5: Slash Command for Review
**Goal**: Claude Code command helps review and apply insights

- Create `.claude/commands/analyze-titles` command
- Command reads latest analysis files from `./tmp`
- Suggests specific prompt improvements
- Shows before/after examples
- Guides developer through updating `titles.go`

**Validation**: Slash command successfully guides prompt improvement

---

### Milestone 6: Documentation & Testing
**Goal**: Feature is documented and tested

- Update `CLAUDE.md` with analysis workflow
- Add example analysis output to docs
- Integration test for full analysis flow
- Document OAuth re-auth requirement
- Add troubleshooting section for common errors

**Validation**: New users can successfully run analysis

---

### Milestone 7: Production Ready
**Goal**: Feature is stable and ready for regular use

- Error handling for all edge cases
- Logging for debugging
- Performance optimization (API calls, AI tokens)
- Rate limiting considerations
- Final end-to-end testing with real channel data

**Validation**: Feature works reliably in production

---

## Dependencies

### External
- YouTube Analytics API v2
- YouTube Data API v3 (already integrated)
- AI Provider (Azure OpenAI or Anthropic - already integrated)

### Internal
- Existing OAuth implementation in `internal/publishing/youtube.go`
- Existing AI provider in `internal/ai/provider.go`
- Existing menu system in `internal/app/app.go`

### New Scopes Required
- `https://www.googleapis.com/auth/yt-analytics.readonly` (one-time re-auth)

## Risks & Mitigation

### Risk: API Quota Limits
**Impact**: Medium
**Probability**: Low
**Mitigation**:
- Fetch only once per day (analysis is periodic, not real-time)
- Fail gracefully with clear error message
- Document quota limits in troubleshooting guide

### Risk: OAuth Re-authentication Friction
**Impact**: Low
**Probability**: High
**Mitigation**:
- Clear messaging about why re-auth is needed
- One-time inconvenience (token cached afterward)
- Document the process

### Risk: AI Analysis Quality
**Impact**: High
**Probability**: Medium
**Mitigation**:
- Iterate on prompt design with real data
- Validate recommendations against known patterns
- Include raw data so human can verify AI insights

### Risk: Stale Analysis
**Impact**: Low
**Probability**: Medium
**Mitigation**:
- Analysis is intentionally periodic (not real-time)
- Timestamp files to show freshness
- Document recommended analysis frequency (monthly/quarterly)

## Open Questions

1. **Date range configurability**: Should we support custom date ranges, or hardcode 365 days?
   - **Decision**: Hardcode 365 days for v1 (simplicity)

2. **Video metadata**: Should we fetch video titles/details from YouTube Data API?
   - **Decision**: Yes, minimal fetch to enrich analytics with titles

3. **Multiple analysis types**: Should "Analyze" menu support future analytics (thumbnails, descriptions)?
   - **Decision**: Yes, design for extensibility but implement Titles only for v1

4. **Analysis frequency**: How often should users run analysis?
   - **Recommendation**: Monthly or after publishing ~10-20 new videos

## Future Enhancements

**Phase 2: Enhanced Analytics**
- Thumbnail analysis (correlation with CTR)
- Description length and structure analysis
- Publishing time optimization
- Topic/keyword trend analysis

**Phase 3: Automated Insights**
- Periodic analysis runner (e.g., monthly cron)
- Email/notification with insights
- Dashboard view of trends over time

**Phase 4: Real-time Feedback**
- Show predicted CTR when generating titles
- A/B test title suggestions
- Live performance tracking

---

## Progress Log

### 2025-11-09 - Session 1: Milestones 1 & 3 Complete
**Duration**: ~2 hours
**Status**: 2 of 7 milestones complete (29%)

#### ✅ Milestone 1: YouTube Analytics API Integration (100%)
**Files Created:**
- `internal/publishing/youtube_analytics.go` - Core analytics implementation
- `internal/publishing/youtube_analytics_test.go` - Comprehensive test coverage

**Implementation Details:**
- `VideoAnalytics` struct with fields: VideoID, Title, Views, CTR, AverageViewDuration, Likes, Comments, PublishedAt
- `GetVideoAnalytics(ctx, startDate, endDate)` - Fetches analytics for custom date ranges
- `GetVideoAnalyticsForLastYear(ctx)` - Convenience wrapper for 365-day fetch
- OAuth client extended with `yt-analytics.readonly` scope
- Uses `configuration.GlobalSettings.YouTube.ChannelId` from settings.yaml (not hardcoded)
- Fetches video metadata (titles, publish dates) from YouTube Data API
- Handles up to 200 videos per request
- Comprehensive error handling for API failures and quota limits

**Dependencies Added:**
- `google.golang.org/api/youtubeanalytics/v2`

**Testing & Validation:**
- Unit tests created with struct validation and edge case coverage
- Manual testing successful with real YouTube channel (200+ videos)
- OAuth flow validated with brand account authentication
- All tests passing, build succeeds

#### ✅ Milestone 3: Menu Integration & UX (100%)
**Files Modified:**
- `internal/app/menu_handler.go` - Added Analyze menu functionality

**Implementation Details:**
- Added "Analyze" as root menu option (index 2)
- Created `HandleAnalyzeMenu()` - Sub-menu with "Titles" option
- Created `HandleAnalyzeTitles()` - Fetches analytics and displays success message
- Progress indicators: "Fetching video analytics from YouTube..."
- Success message: "✓ Successfully fetched analytics for X videos from the last 365 days"
- Simplified terminal output (removed detailed statistics - will be in files per Milestone 4)
- OAuth re-authentication flow handled seamlessly

**User Experience:**
- User selects: Main Menu → Analyze → Titles
- App fetches analytics (with progress messages)
- Displays success with video count
- Returns to menu (no manual "Press Enter" required)

#### Technical Decisions Made:
1. **Channel ID from settings**: Use `settings.yaml` instead of hardcoded constant for flexibility
2. **Terminal output simplified**: Removed detailed statistics display in favor of file output (Milestone 4)
3. **Menu structure**: Designed "Analyze" menu to be extensible for future analytics types (thumbnails, descriptions)
4. **Brand account authentication**: User must authenticate as brand account, not personal account

#### Issues Resolved:
- **403 Forbidden error**: Resolved by enabling YouTube Analytics API in Google Cloud Console
- **Wrong channel data**: Fixed by using brand account authentication instead of personal account
- **Channel ID mismatch**: Updated code to read from `settings.yaml` instead of hardcoded value

#### Next Session Priorities:
- **Milestone 2**: AI Title Analysis Engine - Send analytics data to AI, get recommendations
- **Milestone 4**: File Output - Save JSON data and markdown analysis to `./tmp/`

---

### 2025-11-09 - Session 2: Milestone 2 Complete (AI Title Analysis Engine)
**Duration**: ~2 hours
**Status**: 3 of 7 milestones complete (43%)

#### ✅ Milestone 2: AI Title Analysis Engine (100%)
**Files Created:**
- `internal/ai/analyze_titles.go` - Core analysis function with template support (85 lines)
- `internal/ai/templates/analyze-titles.md` - Go template-based prompt for AI (96 lines)
- `internal/ai/analyze_titles_test.go` - Comprehensive test coverage (228 lines, 7 test cases)

**Implementation Details:**
- `AnalyzeTitles(ctx, analytics)` function processes raw `[]VideoAnalytics` data
- Template embedded in binary using `//go:embed` (no external file dependencies)
- Go's `text/template` package for standard, maintainable templating
- Raw data passed to AI (no pre-processing) for maximum pattern discovery capability
- AI instructed to account for video age bias (older videos naturally accumulate more views)
- Comprehensive error handling: empty data, AI failures, template execution errors

**Testing & Validation:**
- Unit tests cover: valid data, empty data, AI errors, single video, large datasets (100 videos)
- Template execution validated with special characters and date formatting
- All 52 AI package tests passing (including 7 new tests for analyze_titles)
- Build succeeds with embedded template
- Integration into menu handler tested

**Technical Decisions Made:**
1. **Go templates over string replacement**: Standard, well-understood approach preferred by user
2. **Raw data to AI**: Let AI discover patterns without lossy summarization - no pre-calculation of statistics
3. **Embedded templates**: Moved from `prompts/` to `internal/ai/templates/` and embedded in binary for distribution
4. **Removed unused `prompts/` directory**: Cleaned up repository, all templates now embedded
5. **Temporary output display**: Added console output for testing before file saving (Milestone 4)

**Files Modified:**
- `internal/app/menu_handler.go` - Updated `HandleAnalyzeTitles()` to call `ai.AnalyzeTitles()` with progress messages and temporary output display

**Files Deleted:**
- `prompts/` directory - No longer needed, templates embedded in packages

#### Next Session Priorities:
- **Milestone 4**: File Output & Persistence - Save JSON data and markdown analysis to `./tmp/`
- Test with real YouTube channel data to validate analysis quality
- Remove temporary console output once file saving is implemented

---

### 2025-11-09 - Session 3: Milestone 4 Complete (File Output & Persistence)
**Duration**: ~3 hours
**Status**: 4 of 7 milestones complete (57%)

#### ✅ Milestone 4: File Output & Persistence (100%)
**Files Created:**
- `internal/app/analytics_files.go` (85 lines) - Pure, testable file-saving function
- `internal/app/analytics_files_test.go` (275 lines, 9 tests) - Comprehensive test coverage

**Implementation Details:**
- `SaveAnalysisFiles()` function - Pure function with no external dependencies
- JSON file saved to `./tmp/youtube-analytics-{date}.json` with pretty-printed formatting
- Markdown file saved to `./tmp/title-analysis-{date}.md` with metadata header
- Metadata includes: generation timestamp, video count, date range, channel ID
- Error handling for empty data, invalid paths, file I/O failures
- Files overwrite on same-day re-runs (idempotent behavior)

**Testing & Validation:**
- 9 comprehensive test cases covering success and error scenarios
- All tests passing (SaveAnalysisFiles: 9/9, app package: all tests, full suite: 24 packages)
- Test coverage includes: JSON structure validation, Markdown metadata, filename format, overwrite behavior

**Code Quality Improvements:**
- Fixed 12 linter errors in `menu_handler.go` (`log.Printf` → `log.Print`)
- Fixed 3 linter errors in `internal/slack/service.go`
- Fixed API test field count (44 → 46) in `handlers_test.go`
- Updated `CLAUDE.md` with mandatory testing guidelines and checklist
- Refactored file-saving logic into pure, testable function (no test pollution in production code)

**Bug Fixes:**
- Fixed CTR (Click-Through Rate) bug: Was hardcoded to 0.0
- Added `cardClickRate` metric to YouTube Analytics API query (line 73 in `youtube_analytics.go`)
- Updated response parsing to extract actual CTR data from API (line 146)
- Now fetches real CTR percentages from YouTube

**Files Modified:**
- `internal/app/menu_handler.go` - Integrated `SaveAnalysisFiles()`, removed temporary debug output, fixed linter errors
- `internal/publishing/youtube_analytics.go` - Added cardClickRate metric, fixed CTR parsing
- `internal/slack/service.go` - Fixed logging linter errors
- `internal/api/handlers_test.go` - Updated field count expectations
- `CLAUDE.md` - Added "Test-First Development (MANDATORY)" section with testing checklist

**Technical Decisions Made:**
1. **Extracted pure function**: Separated file-saving logic from MenuHandler for testability
2. **Test-first approach**: Wrote comprehensive tests before considering feature complete
3. **Conservative completion criteria**: Only mark checkboxes when direct evidence exists
4. **Fixed pre-existing issues**: Addressed linter errors throughout codebase, not just new code

#### Next Session Priorities:
- **Milestone 5**: Slash Command for Review - Create `.claude/commands/analyze-titles` command
- Test file generation with real YouTube data (CTR now populated with actual values!)
- Validate AI analysis quality with real channel data

---

### 2025-11-09 - Session 4: Milestone 5 Complete (Slash Command for Review)
**Duration**: ~2 hours
**Status**: 5 of 7 milestones complete (71%)

#### Pre-Work: Template Refactoring
**Files Created:**
- `internal/ai/templates/titles.md` - Extracted prompt template from inline code

**Files Modified:**
- `internal/ai/titles.go` - Refactored to use `//go:embed` template pattern
  - Added `titlesTemplateData` struct for template data
  - Replaced `fmt.Sprintf` with `template.Parse()` and `Execute()`
- `internal/ai/titles_test.go` - Added `TestTitlesTemplateExecution` (4 test cases)

**Testing:**
- All existing tests pass (5/5 for titles)
- New template tests pass (4/4 covering normal, special chars, empty, and large manuscripts)
- Full test suite passes (24 packages)
- AI package coverage: 85.1%

**Rationale**: Separating prompt from code enables:
1. Easier iteration on prompt improvements
2. Clearer git diffs for prompt changes
3. Direct slash command editing of template files

---

#### ✅ Milestone 5: Slash Command for Review (100%)

**Files Created:**
- `.claude/commands/analyze-titles.md` - Comprehensive 7-step workflow specification (245 lines)

**Command Design:**
The command implements:
1. Find latest analysis file in `./tmp/`
2. Read AI recommendations about patterns, length, performance
3. Review current template in `internal/ai/templates/titles.md`
4. Suggest data-driven improvements with before/after examples
5. Present structured recommendations by priority
6. Guide implementation (4 options: apply all, selective, preview, manual)
7. Validate applied changes

**Key Features:**
- **Data-Driven**: Every suggestion references specific analytics findings
- **Actionable**: Concrete before/after examples, not generic advice
- **Traceable**: Links recommendations to performance metrics
- **Iterative**: Designed for periodic refinement (monthly/quarterly)

---

#### Validation: Slash Command Execution

**Test Scenario**: User executed `/analyze-titles` in separate Claude Code session (no context)

**Analysis File Used**: `./tmp/title-analysis-2025-11-09.md` (200 videos, last 365 days)

**Results**: Command successfully transformed template with data-driven improvements:

**Character Count Optimization:**
- ❌ Before: "Each title must be 70 characters or less"
- ✅ After: "Target 56-65 characters for optimal performance (acceptable range: 55-75 characters maximum)"
- **Data source**: Analysis showed 56-65 char titles averaged 8,200 views vs 4,100 for under 50 chars

**Front-Loading Strategy Added:**
- ✅ New: "CRITICAL: Front-load the hook in the first 60 characters (mobile truncation point)"
- **Data source**: Front-loaded hooks outperform by 23% when titles exceed 70 chars

**5 High-Performing Patterns Added (with metrics):**
1. **Provocative Opinion** (3-5x avg views)
2. **Top N Lists** (highest view counts - 48,858 views example)
3. **Direct Comparisons** (35% higher watch time)
4. **Challenge/Disruptive** (2.8% engagement vs 1.9% avg)
5. **Personal Workflow** (65% higher comment rate)

**AVOID Section Added (anti-patterns with penalties):**
- Generic "How To" openings (52% fewer likes per view)
- "Explained" or "Tutorial" suffix (62% lower views)
- Question-only without controversy (54% lower watch time)
- Overly technical without context (80% lower comment rate)
- Titles over 75 characters (8-12% CTR drop per 10 extra chars)

**All recommendations traceable to analysis data** with specific examples from actual videos.

---

#### Technical Decisions Made

1. **Template-First Design**: Refactored titles.go before implementing slash command (consistency, easier iteration)
2. **Comprehensive Command Spec**: Detailed 7-step workflow ensures predictable, valuable guidance
3. **Data Traceability**: Every recommendation links to specific analytics findings
4. **User Options**: Command offers 4 implementation paths
5. **Iterative Process**: Designed for periodic execution as more data accumulates

---

#### Impact Assessment

**Before Template Update:**
- Generic prompt with no specific guidance
- 70-character max (not optimized)
- No pattern recommendations
- No anti-patterns identified

**After Template Update:**
- Data-driven character count (56-65 optimal)
- 5 proven high-performing patterns with metrics
- 5 anti-patterns to avoid with performance penalties
- Front-loading strategy for mobile truncation
- All guidance backed by real channel data

**Expected Impact:**
- Higher CTR from optimized title lengths
- More engaging titles using proven patterns
- Fewer low-performing "How To" and "Explained" titles
- Better mobile visibility with front-loaded hooks

**Validation Method**: Generate titles for next 5-10 videos, compare performance to historical averages after 30-60 days

---

#### Next Session Priorities

**Milestone 6**: Documentation & Testing
- Update `CLAUDE.md` with analytics workflow
- Add example analysis output to docs
- Document OAuth re-auth requirement
- Add troubleshooting section

**Milestone 7**: Production Ready
- Error handling review
- Performance optimization
- Final end-to-end testing
- Feature stability validation

---

## References

- [YouTube Analytics API Documentation](https://developers.google.com/youtube/analytics)
- [Existing Analysis Document](docs/youtube-analytics-recommendations.md)
- [GitHub Issue #331](https://github.com/vfarcic/youtube-automation/issues/331)
