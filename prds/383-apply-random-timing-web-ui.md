# PRD: Apply Random Timing in Web UI

**Issue**: #383
**Status**: In Progress
**Priority**: Medium
**Created**: 2026-03-12

---

## Problem Statement

The CLI has an "Apply Random Timing" feature that lets users apply AI-generated timing recommendations to a video's publish date. When a user selects "Apply Random Timing? Yes" in the CLI form, the system picks a random recommendation from `settings.yaml` and calculates a new date within the same week. The Web UI has no equivalent — it only offers a plain `datetime-local` input for the date field, leaving users no way to leverage timing recommendations without switching to the CLI.

## Proposed Solution

1. **Backend API endpoint**: `POST /api/videos/{videoName}/apply-random-timing` — loads timing recommendations from `settings.yaml`, calls the existing `ApplyRandomTiming()` logic, and returns the new date plus the selected recommendation's reasoning
2. **Frontend button**: Add an "Apply Random Timing" button next to the date field in the Web UI that calls the endpoint and updates the date value in the form
3. **Testing**: Backend handler tests and frontend component tests maintaining 80% coverage

## Success Criteria

### Must Have
- [x] Backend endpoint applies random timing and returns new date + reasoning
- [ ] Web UI shows an "Apply Random Timing" button next to the date field
- [ ] Clicking the button updates the date field with the AI-recommended timing
- [ ] User sees which recommendation was applied (day, time, reasoning)
- [ ] Button is disabled or hidden when no timing recommendations exist in settings
- [ ] 80% test coverage maintained on new/modified code
- [x] All existing tests continue to pass

### Nice to Have
- [ ] Let user pick from all recommendations instead of random selection
- [ ] Show before/after date comparison

## Technical Scope

### What Already Exists

| Layer | Component | Location | Status |
|-------|-----------|----------|--------|
| Backend | `ApplyRandomTiming()` | `internal/app/timing_logic.go` | Done |
| Backend | `GetWeekBoundaries()` | `internal/app/timing_logic.go` | Done |
| Backend | `LoadTimingRecommendations()` | `internal/configuration/settings.go` | Done |
| Backend | `TimingRecommendation` struct | `internal/configuration/cli.go` | Done |
| Backend | Timing config in `Settings.Timing` | `internal/configuration/cli.go` | Done |
| Frontend | `DateInput` component | `web/src/components/forms/DateInput.tsx` | Done |
| Frontend | `DynamicForm` rendering for date fields | `web/src/components/forms/DynamicForm.tsx` | Done |
| Frontend | Form dirty tracking + save flow | `web/src/components/forms/DynamicForm.tsx` | Done |

### What Needs to Be Built

| Layer | Component | Details |
|-------|-----------|---------|
| Backend | API handler | New handler `handleApplyRandomTiming` in `internal/api/` |
| Backend | Route | `POST /api/videos/{videoName}/apply-random-timing` in `server.go` |
| Backend | Handler tests | Test handler with mocked timing recommendations |
| Frontend | API hook | TanStack Query mutation hook `useApplyRandomTiming` in `web/src/api/hooks.ts` |
| Frontend | Types | Response type for timing endpoint in `web/src/api/types.ts` |
| Frontend | Button component | "Apply Random Timing" button next to `DateInput` |
| Frontend | DynamicForm update | Render the button for date fields when timing is available |
| Frontend | Component tests | Vitest + Testing Library tests for new UI |

### Key Files to Modify

- `internal/api/server.go` — Add route
- `internal/api/handlers_timing.go` (new) — Handler implementation
- `web/src/api/hooks.ts` — Add mutation hook
- `web/src/api/types.ts` — Add response type
- `web/src/components/forms/DateInput.tsx` or `DynamicForm.tsx` — Add button next to date field

### Patterns to Follow

- **Backend handler pattern**: Follow existing handlers in `internal/api/` (e.g., `handleAnalyzeTitles`) — read video, perform action, return JSON response
- **Frontend mutation pattern**: Follow `useApplyTitlesTemplate` in `hooks.ts` for POST mutations with TanStack Query
- **Button pattern**: Follow `ActionButton.tsx` for buttons that trigger API calls and update form state

## Implementation Milestones

- [x] **Backend endpoint**: Create `POST /api/videos/{videoName}/apply-random-timing` handler that loads recommendations from settings, calls `ApplyRandomTiming()` with the video's current date, and returns `{ newDate, recommendation: { day, time, reasoning } }`. Include handler tests.
- [ ] **Frontend hook + types**: Add `useApplyRandomTiming` mutation hook and response types. Add MSW test handler.
- [ ] **UI integration**: Add "Apply Random Timing" button next to the date field. Clicking it calls the mutation, updates the date form value, and shows the applied recommendation's reasoning. Button disabled when no recommendations configured.
- [ ] **Testing + validation**: Component tests for the button. End-to-end validation. 80% coverage maintained.

## Dependencies

- Timing recommendations must already exist in `settings.yaml` (generated via CLI's "Analyze > Timing" flow)
- `internal/app/timing_logic.go` — existing logic to reuse
- `internal/configuration/settings.go` — `LoadTimingRecommendations()` to load config

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| No timing recommendations in settings.yaml | Medium | Low | Disable button and show helpful message when no recommendations exist |
| `ApplyRandomTiming` is in `internal/app/` (CLI package) | Low | Medium | Either import directly or extract to shared package if import cycle occurs |
| Video has no date set yet | Low | Low | Require date field to have a value before applying timing (same as CLI behavior) |

## Progress Log

### 2026-03-12
- PRD created
- GitHub issue #383 opened
- Backend endpoint implemented: `handlers_timing.go` with handler, response type, route, and 5 test cases
