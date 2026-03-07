# PRD: Title Analysis API & UI

**Issue**: #375
**Status**: Draft
**Priority**: Medium
**Created**: 2026-03-07

---

## Problem Statement

AI-powered title performance analysis — identifying which title patterns drive views and engagement — is only available through the CLI's `Analyze → Titles` menu. The analysis results (high/low-performing patterns, length analysis, recommendations, prompt suggestions) are displayed as terminal output and saved to local files, with no way to view them from the web UI.

## Proposed Solution

Add a `POST /api/analytics/titles` endpoint that triggers the existing `AnalyzeTitles()` AI function from `internal/ai/analyze_titles.go`, and build a frontend view displaying the structured analysis results: title patterns, recommendations, and prompt engineering suggestions.

### What Already Exists

- `AnalyzeTitles(ctx, analytics)` — sends video analytics to AI, returns `TitleAnalysisResult` with patterns, length analysis, content type analysis, engagement patterns, recommendations, and prompt suggestions
- `FormatTitleAnalysisMarkdown()` — formats results as readable markdown
- `SaveCompleteAnalysis()` — saves 4-file audit trail (analytics JSON, prompt, AI response, formatted result)
- AI template at `internal/ai/templates/analyze-titles.md`

### What Needs to Be Built

1. **API endpoint**: `POST /api/analytics/titles` — fetches analytics, runs AI analysis, returns structured `TitleAnalysisResult`
2. **Frontend view**: Rendered analysis with sections for patterns, recommendations, and prompt suggestions

### User Journey

**Before**: CLI → Analyze → Titles → wait ~30s → read terminal output → check `tmp/` for saved files

**After**: Open browser → Analytics → Title Analysis → click "Run Analysis" → see loading spinner → view structured results with pattern cards, recommendation list, and prompt suggestions

## Success Criteria

- [ ] `POST /api/analytics/titles` returns `TitleAnalysisResult` JSON
- [ ] Audit trail saved to `tmp/` (same as CLI behavior)
- [ ] Frontend displays all analysis sections: high/low patterns, length, content type, engagement, recommendations, prompt suggestions
- [ ] Loading state during analysis (~30s for API calls + AI)
- [ ] Error handling for YouTube API and AI failures
- [ ] Tests passing on API handler

## Technical Scope

### API Endpoint

```
POST /api/analytics/titles
```

Response: `TitleAnalysisResult` JSON (same structure the AI returns)

Flow: Fetch analytics → `AnalyzeTitles()` → save audit trail → return result.

### Frontend

- New route: `/analytics/titles`
- "Run Analysis" button (not auto-triggered — expensive operation)
- Sections: High-Performing Patterns (cards with pattern, description, impact, examples), Low-Performing Patterns, Title Length Analysis, Content Type Analysis, Engagement Patterns, Recommendations, Prompt Suggestions
- TanStack Query mutation (not a query — triggered on demand)

### Dependencies

- YouTube Analytics API credentials
- AI provider credentials (Azure OpenAI)
- Existing `internal/ai/analyze_titles.go` and `internal/publishing/youtube_analytics.go`

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Long operation (~30s) | Medium | Clear loading state, consider SSE for progress |
| AI quota/cost | Low | On-demand only, not auto-triggered |
| AI response format changes | Low | Existing JSON parsing with validation |

## Milestones

- [ ] **API endpoint**: `POST /api/analytics/titles` handler using existing `AnalyzeTitles()`, audit trail saving, wired in server. Tests passing.
- [ ] **Frontend title analysis view**: Route, "Run Analysis" button, structured display of all result sections, loading/error states. Tests passing.

## Progress Log

### 2026-03-07
- PRD created
- GitHub issue #375 opened
