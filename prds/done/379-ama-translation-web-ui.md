# PRD: AMA + Translation Web UI Integration

**Issue**: #379
**Status**: Complete
**Completed**: 2026-03-13
**Priority**: Medium
**Created**: 2026-03-07
**Parent**: Split from PRD #372 (HTTP API + React Frontend), Milestone 13

---

## Problem Statement

The AMA (Ask Me Anything) and Translation features have full backend API endpoints and frontend hooks/types already implemented, but lack frontend UI integration. Users cannot trigger AMA content generation or video translation from the web UI — these features are only accessible via the CLI.

Specifically:
- 5 API endpoints exist and are routed (`/api/ai/ama/content`, `/api/ai/ama/title`, `/api/ai/ama/description`, `/api/ai/ama/timecodes`, `/api/ai/translate`)
- TanStack Query hooks exist (`useAIAMAContent`, `useAIAMATitle`, `useAIAMADescription`, `useAIAMATimecodes`, `useAITranslate`)
- TypeScript types and MSW test handlers exist
- But none of this is wired into the dynamic form system (`aiFields.ts`) or exposed as UI components

## Proposed Solution

1. **AMA Integration**: Wire AMA generation into `aiFields.ts` so AMA-specific fields show AI generate buttons in the Definition aspect, following the existing `AIGenerateButton` pattern
2. **Translation UI**: Add a Translation panel or button that triggers `/api/ai/translate` for a video and displays/applies the translated metadata
3. **Testing**: End-to-end validation through the web UI with maintained 80% test coverage

## Success Criteria

### Must Have
- [ ] AMA generate buttons appear on relevant fields in the Definition aspect
- [ ] Users can generate AMA content (title, description, timecodes, full content) from the web UI
- [ ] Users can trigger translation for a video from the web UI
- [ ] Translation results are displayed and can be applied to video fields
- [ ] 80% test coverage maintained on new/modified code
- [ ] All existing tests continue to pass

### Nice to Have
- [ ] Language selector for translation (currently uses a default target language)
- [ ] Translation preview before applying
- [ ] Batch AMA generation (generate all AMA fields at once)

## Technical Scope

### What Already Exists

| Layer | Component | Location | Status |
|-------|-----------|----------|--------|
| Backend | AMA content endpoint | `internal/api/handlers_ai.go` | Done |
| Backend | AMA title endpoint | `internal/api/handlers_ai.go` | Done |
| Backend | AMA description endpoint | `internal/api/handlers_ai.go` | Done |
| Backend | AMA timecodes endpoint | `internal/api/handlers_ai.go` | Done |
| Backend | Translate endpoint | `internal/api/handlers_ai.go` | Done |
| Backend | Route wiring | `internal/api/server.go` | Done |
| Frontend | TanStack Query hooks | `web/src/api/hooks.ts` | Done |
| Frontend | TypeScript types | `web/src/api/types.ts` | Done |
| Frontend | MSW test handlers | `web/src/test/handlers.ts` | Done |
| Frontend | `aiFields.ts` integration | `web/src/lib/aiFields.ts` | **Missing** |
| Frontend | AMA UI buttons | `web/src/components/forms/` | **Missing** |
| Frontend | Translation UI panel | `web/src/components/` | **Missing** |

### What Needs to Be Built

1. **`aiFields.ts` updates**: Add AMA field mappings so `AIGenerateButton` renders for AMA-eligible fields
2. **AMA button integration**: May need AMA-specific button behavior (e.g., generate-all option) following existing `AIGenerateButton` patterns
3. **Translation component**: New UI component — translation is not field-level (it translates entire video metadata), so it needs a different pattern than `AIGenerateButton`
4. **Tests**: Vitest + Testing Library + MSW tests for new components

### Key Files to Modify

- `web/src/lib/aiFields.ts` — Add AMA field-to-endpoint mappings
- `web/src/components/forms/AIGenerateButton.tsx` — Extend if needed for AMA-specific behavior
- `web/src/components/forms/DynamicForm.tsx` — May need updates for translation panel placement
- New: Translation UI component

### Patterns to Follow

- **AI field config**: `web/src/lib/aiFields.ts` maps field names to AI endpoints and UI behavior
- **AI generate button**: `web/src/components/forms/AIGenerateButton.tsx` handles checkboxes (titles/shorts), radio (tweets), and simple generation
- **Testing**: Vitest + Testing Library + MSW, following patterns in `web/src/test/`

## Implementation Milestones

- [ ] **AMA Field Integration**: Add AMA fields to `aiFields.ts`, wire generate buttons into Definition aspect forms. AMA title/description/timecodes/content generation accessible from web UI.
- [ ] **Translation UI Component**: Create translation panel/button component. User can trigger translation, see results, and apply translated metadata to video fields.
- [ ] **Testing + Validation**: Component tests for AMA buttons and Translation UI. End-to-end validation with real data. 80% test coverage maintained.

## Dependencies

- PRD #372 (HTTP API + React Frontend) — provides the foundation (API layer, frontend framework, dynamic forms)
- Existing AI modules in `internal/ai/` — AMA and translation logic
- Azure OpenAI configuration — required for AI generation to work

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| AMA fields don't map cleanly to `aiFields.ts` pattern | Low | Medium | AMA endpoints already follow same request/response pattern as other AI endpoints |
| Translation UX unclear (whole-video vs field-level) | Medium | Medium | Start with simple "Translate" button at video level, iterate based on usage |
| Test coverage regression | Low | High | Run coverage check after each milestone |

## Progress Log

### 2026-03-07
- PRD created (split from PRD #372, Milestone 13)
- GitHub issue #379 opened
- All backend endpoints and frontend hooks already exist from PRD #372 work
