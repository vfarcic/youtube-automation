# PRD: Ship-Ready Hugo Blog Posts

**Issue**: #373
**Status**: Complete
**Priority**: Medium
**Created**: 2026-03-07
**Updated**: 2026-03-12

---

## Problem Statement

The current Hugo post generation (`Hugo.Post()`) produces a post that requires 6 manual steps before it can be published as a PR. The generated content includes a `FIXME:` placeholder, raw manuscript with `TODO:`/`FIXME:` lines, no images, no thumbnail, and no home page listing. Every post needs hands-on editing before it's ready to ship.

## Proposed Solution

Automate all 6 manual post-creation steps so the generated PR is merge-ready without any manual modifications.

## Current Behavior

After `Hugo.Post()` generates a post via PR, the user must manually:

1. Remove the `## Intro` header and move intro content above `<!--more-->` as the excerpt
2. Remove all `FIXME:` and `TODO:` lines from the manuscript content
3. Copy image files referenced in the manuscript (`![](filename)`) from Google Drive into the post directory
4. Copy the video thumbnail into the post directory
5. Add the post to `content/_index.md` (Hugo home page) in the established entry format
6. Remove older entries from `content/_index.md` to keep it at 10 entries max

## Desired Behavior

All 6 steps are performed automatically between post file creation and PR submission. The resulting PR contains a complete, publish-ready post with:

- `## Intro` content extracted as the excerpt above `<!--more-->`
- No `TODO:` or `FIXME:` lines in the body
- All referenced images downloaded from Google Drive into the post directory
- Thumbnail saved as `thumbnail.jpg` in the post directory
- New entry added to `content/_index.md` with thumbnail, title, intro text, and "Full article" link
- `content/_index.md` trimmed to 10 entries maximum

## Success Criteria

- [ ] `## Intro` section extracted as excerpt above `<!--more-->`, removed from body
- [ ] All `TODO:` and `FIXME:` lines stripped from manuscript content
- [ ] Images referenced via `![](filename)` copied from Google Drive to post directory
- [ ] Thumbnail (variant index 0) downloaded to post directory as `thumbnail.jpg`
- [ ] New entry added to `content/_index.md` in established format
- [ ] `content/_index.md` trimmed to max 10 entries
- [ ] Hugo posts generated without any `FIXME:` placeholders
- [ ] Generated PRs can be merged as-is
- [ ] Graceful fallback when Google Drive is not configured (skip image/thumbnail steps)
- [ ] Tests passing, 80% coverage threshold met

## Technical Scope

### Content Processing
- **Key file**: `internal/publishing/hugo.go` — `getPost()` and `Post()` functions
- Extract intro from manuscript (content between `## Intro` and next `##` heading)
- Strip lines starting with `TODO:` or `FIXME:`
- Parse `![](filename)` references for image downloading

### Google Drive Integration
- **Key file**: `internal/gdrive/service.go` — extend `DriveService` interface with `ListFilesInFolder`
- Images are stored in the same Drive folder as the video (subfolder named after video name under root `gdrive.folderId`)
- Thumbnail uses `ThumbnailVariants[0].DriveFileID`

### Home Page Management
- **Key file**: new `internal/publishing/hugo_homepage.go`
- Entry format matches existing `content/_index.md` pattern (thumbnail image + title link + intro + "Full article >>" link + `---` separator)
- New entries inserted after `# Latest Posts` header
- Older entries removed when count exceeds 10

### API Change
- `Post()` signature changes to accept `*storage.Video` (has all needed metadata) and optional `*HugoPostOptions` (carries Drive service dependency)
- Callers updated: API handler passes Drive service from Server fields, CLI passes nil

## Dependencies

- Google Drive service (`internal/gdrive/`) for image and thumbnail downloading
- Video metadata (`*storage.Video`) including Name, ThumbnailVariants, Gist
- Manuscript content available via `Gist` path
- Hugo repo (local or remote) for `content/_index.md` management

## Out of Scope

- Hugo repo PR workflow mechanics (already working, covered by PRD #372)
- Hugo theme or layout changes
- Changes to the manuscript format itself
- AI-generated summaries (using manuscript's own `## Intro` section instead)
