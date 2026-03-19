# PRD: Channel Stats API & UI

**Issue**: #377
**Status**: No Longer Needed
**Priority**: Medium
**Created**: 2026-03-07
**Last Updated**: 2026-03-19
**Closed**: 2026-03-19

---

## Problem Statement

Channel-level statistics — subscriber count, total views, demographics (age/gender), geographic distribution, and engagement metrics — are only fetched internally for the sponsor page update. There's no way to view these metrics directly in the web UI, limiting visibility into audience composition and channel health.

## Proposed Solution

Add a `GET /api/analytics/channel` endpoint that aggregates channel statistics from the existing functions (`GetChannelStatistics`, `GetChannelDemographics`, `GetGeographicDistribution`, `GetEngagementMetrics`), and build a frontend dashboard view displaying these metrics with charts.

### What Already Exists

- `GetChannelStatistics(ctx)` — subscribers, total views, video count
- `GetChannelDemographics(ctx)` — age group and gender distribution (last 90 days)
- `GetGeographicDistribution(ctx)` — top 10 countries by views (last 90 days)
- `GetEngagementMetrics(ctx)` — avg view duration, likes, comments, shares, engagement rate (last 90 days)
- Mermaid chart generators for all of the above (used by sponsor page)

### What Needs to Be Built

1. **API endpoint**: `GET /api/analytics/channel` — calls all 4 functions, returns aggregated result
2. **Frontend view**: Dashboard cards/charts showing subscribers, demographics, geography, engagement

### User Journey

**Before**: No direct way to view channel stats (only visible indirectly via sponsor page update)

**After**: Browser → Analytics → Channel → see subscriber count, video count, age/gender distribution charts, top countries, engagement metrics — all on one dashboard page

## Success Criteria

- [ ] `GET /api/analytics/channel` returns aggregated channel statistics JSON
- [ ] Frontend displays: subscriber count, total views, video count
- [ ] Frontend displays: age group distribution (bar chart or table)
- [ ] Frontend displays: gender distribution (pie chart or table)
- [ ] Frontend displays: top countries by views
- [ ] Frontend displays: engagement metrics (avg views/video, watch time, likes, comments, engagement rate)
- [ ] Loading state while fetching (4 parallel API calls)
- [ ] Tests passing on API handler

## Technical Scope

### API Endpoint

```
GET /api/analytics/channel
```

Response:
```json
{
  "statistics": { "subscriberCount": 123000, "totalViews": 45000000, "videoCount": 500 },
  "demographics": { "ageGroups": [...], "gender": [...] },
  "geography": { "countries": [...] },
  "engagement": { "averageViewDuration": 450.5, "likes": 12000, "comments": 3000, "shares": 800, "views": 500000, "videoCount": 50 },
  "fetchedAt": "ISO8601"
}
```

### Frontend

- New route: `/analytics/channel`
- Stats cards: Subscribers, Total Views, Video Count
- Demographics section: Age groups (horizontal bars), Gender (pie or donut)
- Geography section: Top 10 countries with view counts and percentages
- Engagement section: Metrics table with avg views/video, watch time, engagement rate
- All data fetched in one API call, displayed together

### Dependencies

- YouTube Analytics API credentials
- YouTube Data API credentials
- Existing `internal/publishing/youtube_analytics.go` functions

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Multiple API calls (4 functions) | Low | Run in parallel server-side, single response |
| YouTube API quota | Low | Don't auto-refresh, manual trigger or long staleTime |
| Demographics data availability | Low | Handle missing data gracefully (new channels may have sparse data) |

## Milestones

- [ ] **API endpoint**: `GET /api/analytics/channel` aggregating all 4 data sources, parallel fetching. Tests passing.
- [ ] **Frontend channel dashboard**: Route, stats cards, demographics display, geography display, engagement metrics. Tests passing.

## Progress Log

### 2026-03-07
- PRD created
- GitHub issue #377 opened
