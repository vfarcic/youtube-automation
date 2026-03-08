# PRD: Video Analytics API & UI

**Issue**: #374
**Status**: Draft
**Priority**: Medium
**Created**: 2026-03-07

---

## Problem Statement

Video performance data (views, CTR, watch time, likes, comments) is only accessible through the CLI's `Analyze → Titles` menu flow. Users cannot view per-video analytics from the web UI, limiting visibility into how individual videos perform over time.

## Proposed Solution

Add a `GET /api/analytics/videos` endpoint that fetches YouTube video analytics using the existing `GetVideoAnalytics()` / `GetVideoAnalyticsForLastYear()` infrastructure in `internal/publishing/youtube_analytics.go`, and build a frontend analytics table displaying per-video performance metrics with sorting and filtering.

### What Already Exists

- `GetVideoAnalytics(ctx, startDate, endDate)` — fetches up to 200 videos with views, CTR, avgViewDuration, likes, comments, publishedAt
- `GetVideoAnalyticsForLastYear(ctx)` — convenience wrapper for last 365 days
- `GetFirstWeekMetrics(ctx, videoID, publishDate)` — first-week performance per video
- `EnrichWithFirstWeekMetrics(ctx, analytics)` — batch enrichment with first-week data
- `VideoAnalytics` struct with all required fields

### What Needs to Be Built

1. **API endpoint**: `GET /api/analytics/videos` — wraps `GetVideoAnalyticsForLastYear()`, optionally enriched with first-week metrics
2. **Analytics service interface**: `AnalyticsService` in `internal/api/` to abstract YouTube API calls (testable)
3. **Frontend view**: Sortable table showing video title, views, CTR, avg watch time, likes, comments, publish date, and optionally first-week metrics

### User Journey

**Before**: Launch CLI → Analyze → Titles → wait for API calls → see raw output in terminal

**After**: Open browser → navigate to Analytics tab → see sortable table of video performance → click column headers to sort by views/CTR/engagement → filter by date range

## Success Criteria

- [ ] `GET /api/analytics/videos` returns `VideoAnalytics[]` JSON
- [ ] Optional `?enrichFirstWeek=true` query param adds first-week metrics (slower, more API calls)
- [ ] Frontend analytics table with sortable columns
- [ ] Loading state while fetching (YouTube API calls take several seconds)
- [ ] Error handling for YouTube API failures (quota, auth)
- [ ] Tests passing on API handler

## Technical Scope

### API Endpoint

```
GET /api/analytics/videos?enrichFirstWeek=true
```

Response: `{ "videos": VideoAnalytics[], "fetchedAt": "ISO8601" }`

The handler creates a YouTube service client (same pattern as existing publishing handlers), calls `GetVideoAnalyticsForLastYear()`, optionally calls `EnrichWithFirstWeekMetrics()`, and returns the result.

### Frontend

- New route: `/analytics/videos`
- Navigation: Add "Analytics" section to sidebar
- Table columns: Title, Views, CTR%, Avg Watch Time, Likes, Comments, Published Date
- Optional first-week columns (toggled): First Week Views, First Week CTR, First Week Engagement
- Client-side sorting on all columns
- TanStack Query hook with appropriate staleTime (analytics data doesn't change rapidly)

### Dependencies

- YouTube Analytics API credentials (already configured for publishing)
- YouTube Data API credentials (already configured)
- Existing `internal/publishing/youtube_analytics.go` functions

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| YouTube API quota limits | Medium | Cache results, don't auto-fetch on page load |
| Slow first-week enrichment (N API calls) | Medium | Make it opt-in via query param, show loading indicator |
| OAuth token expiry | Low | Existing token refresh flow handles this |

## Milestones

- [ ] **Analytics service interface + API endpoint**: `AnalyticsService` interface, `GET /api/analytics/videos` handler, wired in server. Tests passing.
- [ ] **Frontend analytics table**: Route, sidebar nav entry, sortable table with all columns, loading/error states. Tests passing.
- [ ] **First-week enrichment toggle**: Optional `enrichFirstWeek` param, frontend toggle, additional columns. Tests passing.

## Progress Log

### 2026-03-07
- PRD created
- GitHub issue #374 opened
