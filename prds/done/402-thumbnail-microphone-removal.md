# PRD: Remove Microphone from Creator Photo in All Thumbnail Variants

**Status**: Complete
**Priority**: Medium
**GitHub Issue**: [#402](https://github.com/vfarcic/youtube-automation/issues/402)
**Created**: 2026-05-21
**Last Updated**: 2026-05-22
**Completed**: 2026-05-22

---

## Problem Statement

Creator photos fed into the thumbnail generation pipeline (`internal/thumbnail/`) often contain a microphone — handheld, on a stand, boom, or lapel — left over from the recording setup. The microphone is incidental to the recording, not part of the intended thumbnail composition, and its presence in the generated thumbnail:

- distracts from the creator's face and the tagline,
- looks unprofessional and inconsistent across videos depending on which photo was used,
- can occlude the tagline text or break the silhouette of the figure.

The existing `BuildPrompt` in `internal/thumbnail/prompt_builder.go` already contains a microphone-removal instruction for the B&W variants, but:

1. It is buried inside a long paragraph about photo treatment, which may reduce model adherence.
2. The new photo-realistic variant (PRD #401) is a separate prompt that does not yet include this rule.
3. We have no verification step — if the model leaves a microphone in, we ship it.

The result is inconsistent output across runs and across variants.

## Solution Overview

Treat microphone removal as a first-class, cross-variant rule:

1. **Make the instruction explicit and prominent in every variant prompt.** Hoist it into its own labeled section of the prompt (e.g., `**Microphone removal:**`) so the model can attend to it independently of the photo-treatment paragraph.
2. **Apply it to every variant**, including the existing two B&W variants and the new photo-realistic variant from PRD #401. As new variants are added, the rule travels with them.
3. **Centralize the wording** so the same canonical instruction is used everywhere — no drift between variants.
4. **Add a lightweight verification path (nice-to-have)** that, if reliably feasible, detects a residual microphone and triggers a single regeneration before returning the image to the user.

## User Journey

There is no new UI surface. The improvement is invisible to the user except in the output:

1. User clicks "Generate thumbnails" in the Web UI (existing flow).
2. System runs every configured provider for every variant.
3. Each variant's prompt includes the canonical microphone-removal instruction.
4. (Optional, nice-to-have) The orchestrator verifies the output and, if a microphone is detected, regenerates once.
5. User sees thumbnails that consistently exclude any microphone from the source photo.

## Success Criteria

### Must have

- [x] Every variant prompt (existing B&W "with illustration", B&W "without illustration", and the new photo-realistic variant from PRD #401) includes the canonical microphone-removal instruction.
- [x] The instruction lives in a single shared constant or helper so all variants pull from the same source — no copy-paste drift.
- [x] The instruction is prominent: hoisted into its own labeled section of the prompt, not buried mid-paragraph.
- [x] Unit tests assert that every prompt produced by the prompt builder(s) contains the canonical microphone-removal text.
- [x] When a new variant is added in the future, the test suite forces it to include the same instruction (e.g., by iterating over all known prompt builders).

### Nice to have

- [~] Post-generation verification step: a vision check that flags a residual microphone, with a single automatic regeneration before the image is stored. _Out of scope: dropped together with M5 — prompt-only fix in M1+M2 is sufficient; revisit only if M4 manual validation shows otherwise._
- [~] If verification is added: telemetry counter for "microphone detected and regenerated" so we can monitor model adherence over time. _Out of scope: dropped together with M5 — prompt-only fix in M1+M2 is sufficient; revisit only if M4 manual validation shows otherwise._
- [ ] Spike: try alternative phrasings (e.g., "the photo I attach may contain a microphone — treat it as if it is not there") and pick the wording with the highest empirical adherence on a sample set.

### Success metrics

- On a manual review of N≥10 freshly generated thumbnails across variants, **zero** contain a visible microphone.
- No regression in generation latency or cost for the must-have path (the verification step, if added, is opt-in or runs only when needed).

## Technical Scope

### Affected packages (high-level)

- `internal/thumbnail/prompt_builder.go` —
  - Extract the microphone-removal sentence into an exported constant (e.g., `MicrophoneRemovalInstruction`).
  - Move it into its own labeled section in `BuildPrompt` (the B&W variants) rather than burying it in the photo-treatment paragraph.
  - If PRD #401 lands first, the new photo-realistic prompt builder must consume the same constant.
- Any new prompt builder added by PRD #401 must include the same constant — enforced by tests.
- `internal/thumbnail/orchestrator.go` — only touched if the nice-to-have verification path is implemented (would add a post-generation check before storing `GeneratedImage`).
- (Nice-to-have) New module under `internal/thumbnail/` (e.g., `verify.go`) that performs a vision check on the generated bytes.

### Open questions

1. Should the microphone-removal text be in `BuildPrompt`'s "Rules" footer (where the most critical bullet rules already live) as well as in its own section, for redundancy? Default proposal: yes — repeating critical instructions in the closing rules list is consistent with current style (the tagline rule is repeated this way).
2. Is the verification step worth the cost? Each verification adds one vision-API call per generated image, and a regeneration doubles the image-gen cost for that variant. Default proposal: gate it behind a setting (default off) and revisit after the must-have prompt change has been evaluated on real output.
3. If verification is implemented, what model does it use? Default proposal: reuse the existing AI provider (`internal/ai/provider.go`) with a simple yes/no prompt — no new dependency.
4. Coordination with PRD #401 — should this PRD be implemented before, after, or in parallel? Default proposal: in parallel; both PRDs touch `prompt_builder.go`, so whichever lands first introduces the shared constant and the other adopts it.

## Milestones

- [x] **M1 — Centralize the microphone-removal instruction.** Extract the sentence into a shared constant; refactor `BuildPrompt` to reference it. No behavior change for the model beyond placement.
- [x] **M2 — Hoist the instruction into its own prompt section.** Move it out of the photo-treatment paragraph into a dedicated labeled section, and also list it in the closing "Rules" footer. Update unit tests to assert the new structure.
- [x] **M3 — Cross-variant enforcement.** Add a test helper that iterates over every known prompt builder (current and any added by PRD #401) and asserts each prompt contains the canonical microphone-removal text. This is the guardrail for future variants.
- [x] **M4 — Manual validation on real output.** Generate thumbnails for at least 5 videos whose source photos contain a microphone; verify all variants exclude it. Capture any prompt-tuning follow-ups. (Confirmed by user 2026-05-23: zero microphones across generated variants on real videos.)
- [~] **M5 — (Out of scope for this PRD) Verification + regenerate path.** Add a vision check + single regeneration when a microphone is detected, gated behind a setting. Tests cover both code paths (detected → regenerate, not detected → pass-through). _Out of scope: M1+M2 already explicitly instruct the image-gen AI to exclude microphones via a hoisted prompt section + Rules footer bullet — the post-hoc vision check + regen is unnecessary unless M4 manual validation shows the prompt-only fix is insufficient._
- [x] **M6 — Test coverage and docs.** `./scripts/coverage.sh` confirms ≥80% coverage on changed packages. Update relevant docs (CLAUDE.md or thumbnail section) to describe the cross-variant rule and the shared constant.

## Risks and mitigations

- **Model non-adherence.** Image models may still leave a microphone visible despite explicit instruction. Mitigation: prompt restructuring (M2) is the first lever; verification + regenerate (M5) is the fallback if M2 is insufficient.
- **Cost of verification.** Adding a vision call per image is non-trivial at scale. Mitigation: gate behind a setting; only enable after M4 shows M2 is not enough.
- **Coupling with PRD #401.** Both PRDs touch the same prompt builder. Mitigation: whichever lands first introduces the shared constant; the other adopts it. M3's cross-variant test forces consistency.

## Dependencies

- Existing `internal/thumbnail` pipeline and providers.
- PRD #401 (coordination only — not a hard blocker; they can land in either order).

No new external services required for M1–M4. M5 (verification) would reuse the existing AI provider.

## Validation Strategy

- Unit tests assert (a) the constant exists and is non-empty, (b) every prompt builder's output contains it, (c) the instruction appears in the dedicated section.
- A cross-variant test iterates over every registered prompt builder so new variants cannot silently skip the rule.
- Manual e2e validation on real photos with visible microphones (M4).
- Coverage verified via `./scripts/coverage.sh`.
