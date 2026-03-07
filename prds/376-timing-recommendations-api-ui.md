# PRD: Timing Recommendations API & UI

**Issue**: #376
**Status**: Draft
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
   - `POST /api/analytics/timing` — generate new timing recommendations (expensive, AI-powered)
   - `GET /api/config/timing` — read current recommendations from settings
   - `PUT /api/config/timing` — save/update recommendations
2. **Frontend view**: Display current recommendations, button to generate new ones, save confirmation
3. **Apply Random Timing integration**: Button next to date field in video editing form

### User Journey

**Before**: CLI → Analyze → Timing → wait ~60s → review recommendations → confirm save → manually apply via "Apply Random Timing" button in video form

**After**: Browser → Analytics → Timing → see current recommendations → optionally generate new ones → save → in video edit form, click "Random Timing" next to date field to auto-apply

## Success Criteria

- [ ] `POST /api/analytics/timing` generates and returns timing recommendations
- [ ] `GET /api/config/timing` returns current recommendations from settings
- [ ] `PUT /api/config/timing` saves recommendations to settings.yaml
- [ ] Frontend displays recommendations with day, time (UTC), and reasoning
- [ ] "Generate New Recommendations" button with loading state
- [ ] "Apply Random Timing" button next to date field in video editing form
- [ ] Tests passing on all endpoints

## Technical Scope

### API Endpoints

```
POST /api/analytics/timing
→ Response: { "recommendations": TimingRecommendation[], "saved": bool }

GET /api/config/timing
→ Response: { "recommendations": TimingRecommendation[] }

PUT /api/config/timing
→ Body: { "recommendations": TimingRecommendation[] }
→ Response: { "saved": true }
```

### Frontend

- New route: `/analytics/timing`
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

- [ ] **Config endpoints**: `GET /PUT /api/config/timing` for reading/saving recommendations. Tests passing.
- [ ] **Generate endpoint**: `POST /api/analytics/timing` using existing `GenerateTimingRecommendations()`. Tests passing.
- [ ] **Frontend timing view**: Route, recommendation display, generate/save buttons, loading states. Tests passing.
- [ ] **Apply Random Timing in date field**: Button next to date input in video editing form, client-side date calculation. Tests passing.

## Progress Log

### 2026-03-07
- PRD created
- GitHub issue #376 opened
