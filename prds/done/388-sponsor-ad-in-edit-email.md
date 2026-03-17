# PRD #388: Include Sponsor Ad Content in Edit Request Emails

**Status**: Complete
**Completed**: 2026-03-17
**Priority**: Medium
**Created**: 2026-03-14
**GitHub Issue**: #388

## Problem

When a sponsored video's edit is requested via the Web UI or CLI, the editor receives an email with animations, project info, and the manuscript — but zero information about the sponsor. The editor has no way to know the ad script, required illustrations, or placement details without manually asking. This causes delays and miscommunication.

## Solution

Add an `AdFile` field to the `Sponsorship` struct that stores the filename of a markdown file in `manuscript/ads/` (e.g., `kilo.md`). When the edit request email is generated and this field is set, read the file content and include it in the email body under a clear "Sponsor Information" heading.

## User Journey

1. User creates a video and sets `sponsorship.adFile` to `kilo.md` (the ad definition file)
2. The file lives at `manuscript/ads/kilo.md` in the data repo
3. User clicks "Request Edit" in the Web UI (or triggers it from CLI)
4. The editor receives the usual edit email, now with an additional section containing the full ad file content
5. The editor has all sponsor context needed — pitch script, illustrations, description text — without any back-and-forth

## Technical Scope

### Data Model Change
- Add `AdFile string` field to `Sponsorship` struct in `internal/storage/yaml.go`
- YAML key: `adFile`, JSON key: `adFile`
- Value is just the filename (e.g., `kilo.md`), not a full path

### Email Generation Change
- Modify `generateEditEmailContent()` in `internal/notification/email.go`
- When `video.Sponsorship.AdFile` is non-empty, resolve the file path within the data directory (`manuscript/ads/<adFile>`)
- Read the file content and append it to the email body inside an HTML section with a heading like **"Sponsor Information"**
- The file content is markdown — convert to basic HTML or include as preformatted text

### File Resolution
- In CLI mode: resolve relative to the manuscript directory
- In server/API mode: resolve relative to the data directory (the cloned git repo)
- The `generateEditEmailContent` function needs access to a base path or a file reader to load the ad file

### Form/UI
- Add `AdFile` field to the sponsorship section in CLI forms and Web UI
- Completion tag: `empty_or_filled` (not required for non-sponsored videos)

## Success Criteria

- Editor receives sponsor ad content in edit request emails when `adFile` is set
- No change to edit emails for non-sponsored videos
- Ad file content is clearly separated from the rest of the email with a visible heading
- Works in both CLI and Web UI email flows

## Milestones

- [x] Add `AdFile` field to `Sponsorship` struct with proper YAML/JSON tags
- [x] Update `generateEditEmailContent()` to read and include ad file content when set
- [x] Handle file resolution for both CLI and server modes
- [x] Add `AdFile` to CLI forms and Web UI sponsorship fields
- [x] Tests covering: sponsored video with ad file, non-sponsored video, missing ad file error handling
- [x] End-to-end validation: request edit for a sponsored video and verify email content

## Risks & Considerations

- **File not found**: If `adFile` is set but the file doesn't exist, the email should still send — include a warning in the email rather than failing the entire request
- **Large files**: Ad files are small markdown (~50 lines), so no size concerns expected
- **HTML rendering**: The ad file is markdown; including it as preformatted text (`<pre>`) is simplest and preserves structure without needing a markdown-to-HTML converter
