# PRD: Photo-realistic Thumbnail Variation with Contextual Illustration

**Status**: Draft
**Priority**: Medium
**GitHub Issue**: [#401](https://github.com/vfarcic/youtube-automation/issues/401)
**Created**: 2026-05-21
**Last Updated**: 2026-05-21

---

## Problem Statement

The current thumbnail generation pipeline (`internal/thumbnail/`) produces two variants per provider, both built from the same `BuildPrompt` template:

- "with illustration"
- "without illustration"

Both variants share a single visual identity: a high-contrast black-and-white stencil treatment of the creator, a solid flat background, and a massive tagline that fills the canvas. The only stylistic axis we A/B test today is the presence/absence of a small illustration accent.

This means:

1. We can't compare our B&W house style against a fundamentally different aesthetic (photo-realistic).
2. We have no thumbnail style where the topic itself is the visual hook — videos about specific named tools (e.g., CodeRabbit, Kubernetes, Crossplane) cannot show a recognizable, topic-evocative subject (e.g., a rabbit, a ship's wheel) rendered in a realistic style.
3. Every variant carries large overlay text, so we cannot test whether text-free thumbnails perform differently in the YouTube feed.

## Solution Overview

Add a **third thumbnail variant** generated alongside the existing two, with these characteristics:

1. **Photo-realistic creator depiction** — uses the attached creator photo as-is (not threshold/stencil-processed). No flat B&W treatment.
2. **Photo-realistic contextual illustration** — an additional element (e.g., a rabbit, a server rack, a robot) rendered in a photo-realistic style, integrated into the composition.
3. **No text overlay** — no tagline, no title, no captions. The image stands on its own.
4. **Subject source** — the illustration subject is either:
   - Inferred by AI from the video manuscript/title (extending the existing `SuggestTaglineAndIllustrations` flow in `internal/ai/tagline_and_illustrations.go` with a new field), OR
   - Typed manually by the user (always available, takes precedence over AI inference).

The variant integrates into the existing pipeline as a third style flowing through `GenerateThumbnails` / `runProvider` — not a parallel system.

## User Journey

### Primary flow

1. User opens the Web UI thumbnail surface for a video.
2. User sees three subject inputs:
   - Tagline (existing, used by B&W variants)
   - Illustration text (existing, used by B&W "with illustration" variant)
   - **Photo-realistic subject (new)** — pre-filled by AI suggestion, editable.
3. User clicks "Generate thumbnails".
4. System produces three variants per configured provider:
   - B&W with illustration (existing)
   - B&W without illustration (existing)
   - **Photo-realistic, contextual subject, no text (new)**
5. User picks the variant they want and selects it (existing select flow, with an extra slot for the third variant).

### Manual-override flow

User types their own subject (e.g., "a small white rabbit holding a code review checklist"). AI suggestion is bypassed for this variant.

### AI-suggestion flow

User leaves the subject empty. On generate, the system uses the AI-suggested photo-realistic subject derived from the manuscript. If AI suggestion is unavailable (no manuscript, AI failure), the variant is skipped with a clear UI message — generation does not block on this.

## Success Criteria

### Must have

- [ ] A third variant is generated for every thumbnail-generation run when the photo-realistic subject is resolvable (manual or AI).
- [ ] Photo-realistic variant uses the creator photo without B&W/stencil processing.
- [ ] Photo-realistic variant includes the topic-related illustration as a realistic element, not flat/stylized.
- [ ] Photo-realistic variant contains no text overlay.
- [ ] Manual subject override is honored over AI inference.
- [ ] AI inference reads from manuscript/title/tags and returns a subject suitable for photo-realistic rendering (concrete noun phrase, not abstract).
- [ ] If AI inference fails and no manual subject is set, generation still produces the two B&W variants — only the third variant is skipped, with a sanitized error returned to the UI.
- [ ] Variant is selectable and uploadable via the existing select endpoint (`POST /api/thumbnails/generated/{id}/select`) and persisted as a `ThumbnailVariant` on the video.
- [ ] Prompt-injection protections (`SanitizePromptInput`) apply to the new subject input.

### Nice to have

- [ ] AI suggests 2–3 alternative photo-realistic subjects (parallel to existing tagline/illustration suggestion shape), letting the user pick rather than retype.
- [ ] Per-variant analytics / labeling so we can compare CTR of the photo-realistic variant against the two B&W variants over time.
- [ ] User setting to disable the third variant globally (for users who don't want photo-realistic output).

### Success metrics

- Third variant is produced successfully on the majority of generate runs where AI/manual subject is resolvable.
- The variant is visually and stylistically distinct from the existing two (verifiable by eye on a sample of videos).
- No regression in generation latency for the existing two B&W variants (they continue to run concurrently).

## Technical Scope

### Affected packages (high-level)

- `internal/thumbnail/prompt_builder.go` — add a new prompt builder for the photo-realistic + contextual-subject + no-text variant. Either a new `BuildPhotoRealisticPrompt(cfg)` function or generalize `BuildPrompt` to take a "style" parameter. The new prompt must explicitly:
  - Instruct the model to keep the creator photo photo-realistic (no threshold/stencil).
  - Describe the contextual subject as a photo-realistic element integrated with the creator.
  - Forbid any text rendering.
- `internal/thumbnail/orchestrator.go` — `runProvider` currently hard-codes two styles ("with illustration" / "without illustration"). Generalize to N styles, or extend `GenerateRequest` with a third prompt field (`PromptPhotoRealistic`).
- `internal/api/handlers_thumbnail.go` —
  - `ThumbnailConfigRequest` and `handleSaveThumbnailConfig` gain a new field (e.g., `PhotoRealisticSubject`).
  - `handleGenerateThumbnails` builds the third prompt and passes it through.
- `internal/storage/yaml.go` — extend the `Video` struct with the new subject field (with appropriate `completion` tag). Existing `ThumbnailVariants` already supports an arbitrary number of slots, so no schema change there.
- `internal/ai/tagline_and_illustrations.go` (or a sibling file `internal/ai/photo_realistic_subject.go`) — add an AI-suggestion function that returns a subject string suitable for photo-realistic rendering. Update the template in `internal/ai/templates/` accordingly (or add a new template).
- `web/src/components/forms/ThumbnailGenerateButton.tsx` and surrounding form code — add the photo-realistic subject input, surface AI suggestion, render the third variant alongside the existing two, route selection through the existing select endpoint.
- `internal/aspect/` — if thumbnail config is exposed via the dynamic-form aspect system, register the new field there too.

### Open questions

1. Should the photo-realistic variant be generated by the same providers (Gemini + GPT-Image) or only the one that handles photo-realism best? Default proposal: same providers, single prompt change.
2. If AI inference returns nothing usable, do we (a) silently skip the variant, (b) fall back to the existing illustration text, or (c) prompt the user inline? Default proposal: (a) silently skip with a sanitized error message in the response `errors[]`.
3. Do we want a separate aspect-system slot for "photo-realistic subject" in the completion tracker, or treat it as opt-in metadata? Default proposal: opt-in metadata with `completion:"empty_or_filled"` so it doesn't block phase progression.
4. Naming for the style label persisted on `GeneratedImage.Style` — proposed: `"photorealistic"` (lowercase, single word).

## Milestones

- [ ] **M1 — Prompt + orchestrator support for a third style.** New photo-realistic prompt builder; `runProvider` generalized; `GenerateRequest` extended. Unit tests cover the new prompt content (no text, photo-realistic instruction, subject included) and the orchestrator producing three images per provider when configured.
- [ ] **M2 — Storage + API field for the photo-realistic subject.** New `Video` field, save endpoint accepts and persists it, generate endpoint reads it. Tests cover save/load round-trip, sanitization, and that an empty subject does not block the two B&W variants.
- [ ] **M3 — AI inference for the photo-realistic subject.** New AI function (or extension of `SuggestTaglineAndIllustrations`) that returns a photo-realistic subject string derived from the manuscript. Tests mock the AI provider and verify a non-empty string is returned on valid input and the call surfaces a clear error on invalid input.
- [ ] **M4 — Web UI input + variant rendering.** Subject input field with AI-suggest action; third variant rendered in the generated-thumbnails grid; selection routes through the existing select endpoint; component tests cover happy path, AI-suggestion failure, and manual override.
- [ ] **M5 — End-to-end validation.** Manually generate thumbnails for a sample of videos covering different topics; verify the third variant is visually distinct, contains no text, and uses the contextual subject. Capture any prompt-tuning follow-ups.
- [ ] **M6 — Test coverage threshold and docs.** `./scripts/coverage.sh` confirms ≥80% coverage on changed packages. Update `CLAUDE.md` Thumbnail section (or equivalent docs) to describe the three-variant model and the new subject field.

## Risks and mitigations

- **Prompt brittleness** — image models may still render text or refuse photo-realism with the creator photo. Mitigation: iterate on prompt wording in M1 and M5; keep a smoke-test sample of generated images per provider.
- **Subject inference quality** — AI may suggest abstract or unsuitable subjects (e.g., "the concept of trust"). Mitigation: constrain the template to a "concrete noun phrase, photographable" rule; allow manual override; allow regeneration of suggestion.
- **Cost** — adding a third image per provider per generate run increases per-video image-gen cost roughly 50%. Mitigation: opt-out setting (nice-to-have), and acknowledge cost in M5 review.
- **UI complexity** — three variants instead of two changes the layout. Mitigation: reuse the existing thumbnail grid; treat the third as just another tile.

## Dependencies

- Existing `internal/thumbnail` orchestrator and providers (Gemini, GPT-Image).
- Existing `SuggestTaglineAndIllustrations` AI infrastructure.
- Existing thumbnail select / Drive upload flow.

No external service additions required.

## Validation Strategy

- Unit tests on prompt builder, orchestrator generalization, storage, and AI function (table-driven where applicable, per `CLAUDE.md` Test-First section).
- Component tests on the new Web UI input and variant rendering.
- Manual end-to-end run on at least 3 distinct video topics to verify the photo-realistic + contextual + no-text output is what we expect.
- Coverage verified via `./scripts/coverage.sh`.
