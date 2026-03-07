# PRD: Ship-Ready Hugo Blog Posts

**Issue**: #373
**Status**: Draft
**Priority**: Medium
**Created**: 2026-03-07

---

## Problem Statement

The current Hugo post generation (`Hugo.Post()`) produces a post that requires manual edits before it can be published. The generated content includes a `FIXME:` placeholder and the raw manuscript content, which needs manual curation (e.g., adding a summary/excerpt before `<!--more-->`, cleaning up manuscript formatting, adding/removing sections). This creates friction in the publishing workflow — every post needs hands-on editing before it's ready to ship.

## Proposed Solution

Improve the Hugo post generation so that the output is publish-ready without manual modifications. The post should include a proper summary/excerpt, well-formatted content, and all necessary frontmatter — ready to merge the PR (or commit locally) as-is.

## Current Behavior

The generated post (`internal/publishing/hugo.go:getPost()`) contains:
- TOML frontmatter (title, date, draft=false)
- A `FIXME:` placeholder where a summary should go
- `<!--more-->` separator
- YouTube embed shortcode
- Raw manuscript content

## Desired Behavior

The generated post should be complete and publish-ready:
- Proper summary/excerpt replacing `FIXME:`
- Well-formatted content derived from the manuscript
- No manual editing required before publishing

## Success Criteria

- [ ] Hugo posts generated without `FIXME:` placeholders
- [ ] Posts include a meaningful summary/excerpt
- [ ] Generated posts can be published as-is (merge PR or push)
- [ ] Existing manuscript content is properly formatted for Hugo
- [ ] Tests passing

## Technical Scope

**Key file**: `internal/publishing/hugo.go` — `getPost()` function (lines 84-111)

Details to be determined when implementation begins. May involve AI-assisted summary generation, manuscript-to-Hugo content transformation, or leveraging existing video metadata (description, tags) to enrich the post.

## Dependencies

- Existing AI infrastructure (`internal/ai/`)
- Video metadata (description, tags, etc.) available at post creation time
- Manuscript content available via `gist` path

## Out of Scope

- Hugo repo PR workflow (covered by PRD #372 milestone "Hugo Post PR Workflow")
- Hugo theme or layout changes
- Changes to the manuscript format itself
