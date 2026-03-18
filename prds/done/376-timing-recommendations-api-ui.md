# PRD: Timing Recommendations API & UI

**Issue**: #376
**Status**: Complete
**Priority**: Medium
**Created**: 2026-03-07

---

## Problem Statement

Publish timing optimization — analyzing historical data to determine optimal days/times for uploading videos — is CLI-only. The generated timing recommendations are saved to `settings.yaml` and used by the "Apply Random Timing" button in CLI forms, but none of this is accessible from the web UI. Users can't generate new recommendations, view current ones, or apply random timing from the browser.

## Proposed Solution

Add timing recommendation endpoints using the existing `GenerateTimingRecommendations()` AI function and `settings.yaml` persistence, plus a frontend view for generating, viewing, and managing timing recommendations. Integrate "Apply Random Timing" into the video editing form's date field.

### What Already Exists

- `GenerateTimingRecommendations(ctx, analytics)` — fetches analytics, enriches with timing data, sends to AI, returns 6-8 `TimingRecommendation` structs
- `SaveTimingRecommendations()` — persists to `settings.yaml`
- `ApplyRandomTiming(currentDate, recommendations)` — picks random recommendation, adjusts date within same week
- `TimingRecommendation` struct: `Day`, `Time`, `Reasoning`
- `GetVideoAnalyticsForLastYear()`, `EnrichWithFirstWeekMetrics()`, `EnrichWithTimingData()`, `GroupByTimeSlot()`, `CalculateTimeSlotPerformance()`
- AI template at `internal/ai/templates/analyze-timing.md`

### What Needs to Be Built

1. **API endpoints**:
   - `POST /api/analyze/timing/generate` — generate new timing recommendations (expensive, AI-powered)
   - `GET /api/analyze/timing` — read current recommendations from settings
   - `PUT /api/analyze/timing` — save/update recommendations
2. **Frontend view**: Display current recommendations, button to generate new ones, save confirmation
3. **Apply Random Timing integration**: Button next to date field in video editing form

### User Journey

**Before**: CLI → Analyze → Timing → wait ~60s → review recommendations → confirm save → manually apply via "Apply Random Timing" button in video form

**After**: Browser → Analytics → Timing → see current recommendations → optionally generate new ones → save → in video edit form, click "Random Timing" next to date field to auto-apply

## Success Criteria

- [x] `POST /api/analyze/timing/generate` generates and returns timing recommendations
- [x] `GET /api/analyze/timing` returns current recommendations from settings
- [x] `PUT /api/analyze/timing` saves recommendations to settings.yaml
- [x] Frontend displays recommendations with day, time (UTC), and reasoning
- [x] "Generate New Recommendations" button with loading state
- [x] "Apply Random Timing" button next to date field in video editing form (PRD-383)
- [x] Tests passing on all endpoints (11 backend + 5 frontend)

## Technical Scope

### API Endpoints

```text
POST /api/analyze/timing/generate
→ Response: { "recommendations": TimingRecommendation[], "videoCount": int, "syncWarning"?: string }

GET /api/analyze/timing
→ Response: { "recommendations": TimingRecommendation[] }

PUT /api/analyze/timing
→ Body: { "recommendations": TimingRecommendation[] }
→ Response: { "saved": true, "syncWarning"?: string }
```

### Frontend

- New route: `/analyze/timing`
- Table/cards showing current recommendations: Day, Time (UTC), Reasoning
- "Generate New" button (triggers `POST`, shows loading ~60s)
- Save confirmation after generation
- In `DateInput` component: optional "Random Timing" button that calls `ApplyRandomTiming` logic (can be client-side since it's just date math + random pick from known recommendations)

### Dependencies

- YouTube Analytics API credentials
- AI provider credentials
- Existing timing analysis infrastructure
- `settings.yaml` read/write access

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Very long operation (~60s, N+1 API calls for first-week enrichment) | High | Clear loading state with progress indication, consider SSE |
| Settings.yaml write conflicts with git sync | Medium | Trigger git commit/push after save (existing onMutate pattern) |
| AI recommendation quality | Low | User reviews before saving, can regenerate |

## Milestones

- [x] **Config endpoints**: `GET /PUT /api/analyze/timing` for reading/saving recommendations. Tests passing.
- [x] **Generate endpoint**: `POST /api/analyze/timing/generate` using existing `GenerateTimingRecommendations()`. Tests passing.
- [x] **Frontend timing view**: Route `/analyze/timing`, recommendation table, generate button, loading/error/success states. Tests passing.
- [x] **Apply Random Timing in date field**: Button next to date input in video editing form, server-side via `POST /api/videos/{name}/apply-random-timing`. Tests passing. (PRD-383)

## Progress Log

### 2026-03-07
- PRD created
- GitHub issue #376 opened

### 2026-03-18
- Implemented all 3 API endpoints (GET/PUT/POST) in `internal/api/handlers_analyze_timing.go`
- Added `GenerateTimingRecommendations` to `AnalyzeService` interface for testability
- Registered routes under `/api/analyze/timing` in `server.go`
- Created `AnalyzeTiming.tsx` page with recommendations table, generate button, loading/error states
- Added sidebar navigation link (teal dot) and route in `App.tsx`
- Added TypeScript types, React Query hooks, MSW mock handlers
- 11 backend tests + 5 frontend tests, all passing (165 total frontend tests)
- PRD marked complete
