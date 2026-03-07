# PRD: Sponsor Page Analytics Update API & UI

**Issue**: #378
**Status**: Draft
**Priority**: Medium
**Created**: 2026-03-07

---

## Problem Statement

Updating the Hugo sponsor page with latest channel analytics (demographics charts, geographic distribution, engagement metrics, channel stats) requires CLI access and manual triggering via `Analyze → Sponsor Page`. There's no way to trigger this update or preview the result from the web UI.

## Proposed Solution

Add a `POST /api/analytics/sponsor-page` endpoint using the existing `GenerateSponsorAnalyticsSection()` and `UpdateSponsorPageAnalytics()` functions, with a frontend button to trigger the update and show the result status.

### What Already Exists

- `GetChannelDemographics(ctx)`, `GetGeographicDistribution(ctx)`, `GetChannelStatistics(ctx)`, `GetEngagementMetrics(ctx)` — data fetching
- `GenerateAgeDistributionChart()`, `GenerateGenderDistributionChart()`, `GenerateGeographyChart()`, `GenerateChannelStatsTable()`, `GenerateEngagementTable()` — Mermaid chart/table generators
- `GenerateSponsorAnalyticsSection(demographics, geography, stats, engagement)` — combines all into markdown section with `<!-- SPONSOR_ANALYTICS_START/END -->` markers
- `UpdateSponsorPageAnalytics(section)` — reads Hugo sponsor page, replaces content between markers, writes back
- Hugo Post PR workflow — when `hugo.repoURL` is configured, changes go through a PR instead of direct file writes

### What Needs to Be Built

1. **API endpoint**: `POST /api/analytics/sponsor-page` — fetches all data, generates section, updates page (local or PR depending on config)
2. **Frontend button**: Trigger button in analytics section with status/result display

### User Journey

**Before**: CLI → Analyze → Sponsor Page → wait → see "Updated sponsor page at path" message

**After**: Browser → Analytics → click "Update Sponsor Page" → see loading → see success message with file path or PR URL

## Success Criteria

- [ ] `POST /api/analytics/sponsor-page` fetches analytics, generates charts, updates sponsor page
- [ ] Works in both modes: local filesystem (CLI/dev) and PR workflow (remote/K8s)
- [ ] Returns result indicating what was updated (file path or PR URL)
- [ ] Frontend button with loading and success/error states
- [ ] Tests passing on API handler

## Technical Scope

### API Endpoint

```
POST /api/analytics/sponsor-page
```

Response:
```json
{
  "updated": true,
  "mode": "local" | "pr",
  "path": "/path/to/sponsor/_index.md",
  "prUrl": "https://github.com/.../pull/123"
}
```

Flow:
1. Fetch demographics, geography, stats, engagement (parallel)
2. Generate analytics section via `GenerateSponsorAnalyticsSection()`
3. Update sponsor page via `UpdateSponsorPageAnalytics()` (respects hugo config for local vs PR mode)
4. Return result

### Frontend

- Button in analytics navigation or channel stats page: "Update Sponsor Page"
- Loading state during update
- Success: show file path (local mode) or link to PR (remote mode)
- Error: show error message
- TanStack Query mutation

### Dependencies

- YouTube Analytics API credentials
- Hugo configuration (`hugo.path` for local, `hugo.repoURL` + `hugo.token` for PR mode)
- Existing chart generation and page update functions

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Hugo config not set | Low | Return 501 with clear message |
| PR creation failure (token, permissions) | Medium | Return error with details, don't leave partial state |
| File write conflicts | Low | Marker-based replacement is idempotent |

## Milestones

- [ ] **API endpoint**: `POST /api/analytics/sponsor-page` handler, both local and PR modes, wired in server. Tests passing.
- [ ] **Frontend trigger button**: Button with loading/success/error states, displays result (path or PR URL). Tests passing.

## Progress Log

### 2026-03-07
- PRD created
- GitHub issue #378 opened
