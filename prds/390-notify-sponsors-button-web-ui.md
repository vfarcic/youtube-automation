# PRD #390 - Notify Sponsors Button in Web UI

**Status**: Draft
**Priority**: Medium
**Created**: 2026-03-15
**GitHub Issue**: [#390](https://github.com/vfarcic/youtube-automation/issues/390)

---

## Problem

The Web UI renders "Notified Sponsors" as a simple toggle that only saves the boolean field without actually sending the sponsor notification email. The CLI correctly sends the email when the field is toggled from false to true, creating an inconsistency where Web UI users can mark sponsors as "notified" without actually notifying them.

## Solution

Replace the `notifiedSponsors` toggle with a "Notify Sponsors" action button backed by a dedicated API endpoint that sends the sponsor notification email and marks the field as true on success. Follow the established `ActionButton` pattern already used for "Request Thumbnail" and "Request Edit".

## User Journey

1. User opens a published video in the Web UI's Post Publish aspect
2. If sponsors have NOT been notified: a "Notify Sponsors" button is displayed
3. User clicks the button; it shows "Sending..." while the request is in-flight
4. The API sends the notification email to all sponsor email addresses (+ finance CC)
5. On success: the button is replaced with a green "Sponsors Notified" badge
6. On failure: an error message is shown below the button; the field remains false

## Technical Design

### Existing Patterns to Follow

| Component | Pattern File | Purpose |
|-----------|-------------|---------|
| `ActionButton` | `web/src/components/forms/ActionButton.tsx` | Button + badge UI for action fields |
| `isActionField()` | Same file | Registry of fields rendered as buttons |
| `useRequestEdit()` | `web/src/api/hooks.ts` | React Query mutation hook pattern |
| `handleRequestEdit()` | `internal/api/handlers_actions.go` | Go handler for action endpoints |
| `EmailService` interface | Same file | Abstraction for email sending |
| `Email.SendSponsors()` | `internal/notification/email.go:150` | Existing sponsor email logic |

### Changes Required

**Backend (`internal/api/`):**

1. **Extend `EmailService` interface** (`handlers_actions.go:29-32`): Add `SendSponsors(from, to string, videoID, sponsorshipPrice, videoTitle string) error`
2. **New handler `handleNotifySponsors`** (`handlers_actions.go`): `POST /api/actions/notify-sponsors/{videoName}?category=X`
   - Load video, check if `NotifiedSponsors` is already true (return `alreadyRequested`)
   - Validate sponsorship exists (amount not empty/"N/A"/"-") and emails are set
   - Set `NotifiedSponsors = true`, persist via `UpdateVideo`
   - Send email via `EmailService.SendSponsors()` using `emailSettings.From`, video's `Sponsorship.Emails`, plus `emailSettings.FinanceTo`
   - Return `ActionResponse` (same shape as request-edit/request-thumbnail)
3. **Register route** (`server.go:133-136`): Add `r.Post("/notify-sponsors/{videoName}", s.handleNotifySponsors)` in the `/api/actions` group

**Frontend (`web/src/`):**

4. **Add `notifiedSponsors` to `ACTION_FIELDS`** (`components/forms/ActionButton.tsx:5-8`):
   ```ts
   notifiedSponsors: { label: 'Notify Sponsors', sentLabel: 'Sponsors Notified' },
   ```
5. **Add `useNotifySponsors()` mutation hook** (`api/hooks.ts`): Following `useRequestEdit()` pattern, POST to `/api/actions/notify-sponsors/{videoName}`
6. **Update `ActionButton` to use the new hook**: Add conditional for `notifiedSponsors` field alongside `requestThumbnail`/`requestEdit`

**Tests:**

7. **Go tests** (`internal/api/handlers_actions_test.go`): Test the new handler — success, already-notified, no-sponsorship, email-not-configured, email-error cases
8. **Frontend tests** (`web/src/test/ActionButton.test.tsx`): Add `notifiedSponsors` to `isActionField` tests and button render/badge tests

---

## Milestones

- [ ] **M1: API endpoint** — `POST /api/actions/notify-sponsors/{videoName}` handler + route + extended EmailService interface + Go tests
- [ ] **M2: Frontend button** — Register `notifiedSponsors` in ActionButton, add mutation hook, update ActionButton to use it, add frontend tests
- [ ] **M3: Validation** — End-to-end verification: button renders in Post Publish aspect, click sends email, badge shown after success, error displayed on failure

---

## Success Criteria

- Clicking "Notify Sponsors" in the Web UI sends the same email that the CLI sends
- The button only appears when `notifiedSponsors` is false and sponsorship data exists
- After successful notification, a green "Sponsors Notified" badge is shown
- Email failures are surfaced to the user without marking the field as complete
- All existing tests continue to pass; new functionality has test coverage

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Email password not configured in serve mode | Return clear `emailError` message (existing pattern from `emailNotConfiguredMessage()`) |
| Accidental double-send | `alreadyRequested` guard returns early if field is already true |
| No sponsorship data on video | Validate sponsorship amount + emails before attempting send |

## Out of Scope

- Changing the CLI behavior (CLI checkbox + email trigger remains as-is)
- Adding a "resend" capability (if sponsors were already notified, the badge stays)
- Editing sponsor email addresses from the Web UI
