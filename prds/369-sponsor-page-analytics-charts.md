# PRD #369: Sponsor Page Analytics Charts

## Problem Statement

Potential sponsors evaluating the DevOps Toolkit YouTube channel need current audience analytics to make informed sponsorship decisions. Currently, the sponsor page (`devopstoolkit.live/sponsor`) only contains pricing information without any data about audience reach, demographics, or engagement. This makes it harder for sponsors to assess the value proposition and requires manual communication to share this information.

## Solution Overview

Build a feature that:
1. Fetches YouTube Analytics data (demographics, geographic distribution, channel statistics)
2. Generates Mermaid charts (pie charts for demographics, bar charts for geography)
3. Automatically updates the Hugo sponsor page with the generated charts
4. Supports interactive execution via CLI menu

## User Stories

1. **As a potential sponsor**, I want to see audience demographics on the sponsor page so I can assess if the channel's audience matches my target market.

2. **As a channel owner**, I want to automatically update sponsor page analytics so the information stays current without manual effort.


## Success Criteria

- [x] Demographics data (age groups, gender) displayed as Mermaid pie charts
- [x] Geographic distribution (top countries) displayed as Mermaid bar chart
- [x] Channel statistics (subscribers, total views, avg views/video) displayed in a table
- [x] Sponsor page preserves existing pricing content when updated
- [x] CLI menu option available under "Analyze" menu
- [ ] 80% test coverage maintained for new code

## Technical Approach

### Data Sources

**YouTube Analytics API v2:**
- Demographics: `Dimensions("ageGroup,gender").Metrics("viewerPercentage")`
- Geography: `Dimensions("country").Metrics("views").Sort("-views").MaxResults(10)`

**YouTube Data API v3:**
- Channel stats: `Channels.List(["statistics"])` for subscriber count, total views, video count

### Output Format

Mermaid charts embedded in markdown with marker-based section replacement:

```markdown
<!-- SPONSOR_ANALYTICS_START -->
## Channel Analytics
...charts...
<!-- SPONSOR_ANALYTICS_END -->
```

### Files to Create/Modify

| File | Action |
|------|--------|
| `internal/publishing/youtube_analytics.go` | Add demographic structs and fetching functions |
| `internal/publishing/sponsor_charts.go` | **New** - Mermaid chart generation |
| `internal/publishing/sponsor_page.go` | **New** - Sponsor page update logic |
| `internal/app/menu_analyze.go` | Add menu option and handler |
| Test files | Coverage for all new code |

## Milestones

- [x] **M1: Data fetching** - YouTube Analytics API calls for demographics, geography, and channel stats working
- [x] **M2: Chart generation** - Mermaid pie/bar chart generation functions implemented
- [x] **M3: Sponsor page update** - Marker-based section replacement working
- [x] **M4: CLI menu integration** - "Analyze → Sponsor Page" menu option functional
- [ ] **M5: Tests complete** - 80% coverage with unit tests for all new functions
- [x] **M6: End-to-end validation** - Full workflow tested with real API credentials

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| YouTube API quota limits | Feature runs periodically (monthly), minimal API calls per run |
| Subscriber count hidden | Handle gracefully, display "N/A" or omit if unavailable |
| Hugo theme Mermaid compatibility | Verified: Relearn theme supports Mermaid natively |

## Dependencies

- Existing YouTube OAuth2 authentication (already implemented)
- Hugo site path configured in `settings.yaml` (already exists as `hugo.path`)
- Mermaid support in Hugo theme (already available)

## Out of Scope

- Real-time analytics dashboard
- Historical trend analysis
- Video-level demographic breakdown
- Custom chart styling beyond Mermaid defaults
