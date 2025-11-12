# PRD: AI-Powered YouTube Shorts Candidate Identification from Manuscripts

**Issue**: #339
**Status**: Planning
**Priority**: Medium
**Created**: 2025-11-11
**Last Updated**: 2025-11-11

---

## Problem Statement

Currently, identifying segments from a manuscript that would make good YouTube Shorts requires manual review and subjective judgment. Content creators must:
- Manually read through entire manuscripts looking for self-contained segments
- Estimate if segments fit within the 60-second Short format
- Judge which moments are impactful enough to work as standalone content
- Track which segments they've already extracted to avoid duplication

This manual process is time-consuming, inconsistent, and risks missing valuable Short opportunities that could drive traffic to main videos.

## Proposed Solution

Add AI-powered analysis to automatically identify manuscript segments suitable for YouTube Shorts. The feature will:

1. **Analyze manuscripts** using AI to identify self-contained, high-impact segments
2. **Estimate duration** based on configurable word count limits
3. **Present candidates** with both line number references and extracted text
4. **Store metadata** in video YAML for tracking and editor communication
5. **Support multiple uploads** allowing iterative Short creation from one video
6. **Link Shorts to main videos** in descriptions for cross-promotion

### User Journey

**Before (Current State)**:
1. Editor sends completed video to content creator
2. Creator manually reviews manuscript looking for Short candidates
3. Creator messages editor with timestamps or descriptions of segments
4. Editor extracts and uploads Shorts manually
5. No systematic tracking of which Shorts came from which videos

**After (With This Feature)**:
1. After manuscript is complete, creator opens "Publishing Details"
2. Selects "Analyze for Shorts" option
3. AI presents 3-5 candidate segments with rationale
4. Creator reviews suggestions and selects which to pursue
5. System stores selected segments in video YAML
6. Creator shares YAML with editor who extracts segments
7. Creator uses "Upload Short" option to publish each Short
8. System tracks all Shorts linked to main video
9. Short descriptions automatically link to main video

## Success Criteria

### Must Have (MVP)
- [ ] AI successfully identifies 3-5 Short candidates from a manuscript
- [ ] Candidates meet word count limit (configurable in settings.yaml)
- [ ] System presents both line numbers and extracted text
- [ ] Selected segments are stored in video YAML
- [ ] "Upload Short" option allows multiple uploads per video
- [ ] Uploaded Shorts store reference to main video
- [ ] Short descriptions include link to main video

### Nice to Have (Future)
- [ ] AI learns from user selections to improve recommendations over time
- [ ] Support for different Short types (tutorial, insight, controversial take)
- [ ] Automated thumbnail generation suggestions
- [ ] Analytics tracking: which Shorts drive most traffic to main videos
- [ ] Batch analysis across multiple videos to find patterns

## Technical Scope

### Core Components

#### 1. AI Analysis Module (`internal/ai/shorts.go`)
- New AI module for analyzing manuscripts
- Prompt engineering to identify self-contained, high-impact segments
- Word count validation against configurable limit
- Return structured results with line numbers, text, rationale

#### 2. Storage Changes (`internal/storage/yaml.go`)
- Add `Shorts` field to `Video` struct:
  ```go
  type Video struct {
      // ... existing fields
      Shorts []Short `json:"shorts,omitempty" yaml:"shorts,omitempty"`
  }

  type Short struct {
      ID          string   `json:"id" yaml:"id"`                     // YouTube video ID after upload
      Title       string   `json:"title" yaml:"title"`               // Short title
      Text        string   `json:"text" yaml:"text"`                 // Extracted manuscript text
      LineStart   int      `json:"line_start" yaml:"line_start"`     // Manuscript line start
      LineEnd     int      `json:"line_end" yaml:"line_end"`         // Manuscript line end
      Rationale   string   `json:"rationale" yaml:"rationale"`       // AI explanation
      UploadDate  string   `json:"upload_date,omitempty" yaml:"upload_date,omitempty"` // When uploaded
  }
  ```

#### 3. Configuration (`settings.yaml`)
- Add `shorts_max_words` setting (default: 150)
- Add `shorts_candidate_count` setting (default: 5)

#### 4. CLI Interface (`internal/app/`)
- Add "Analyze for Shorts" option to Publishing Details menu
- Add "Upload Short" option to Publishing Details menu
- Form for reviewing AI suggestions and selecting candidates
- Form for uploading individual Shorts (title, description override)

#### 5. API Interface (`internal/api/`)
- `POST /videos/{id}/analyze-shorts` - Trigger analysis
- `GET /videos/{id}/shorts` - Get suggested/uploaded Shorts
- `POST /videos/{id}/shorts` - Select a Short candidate
- `POST /videos/{id}/shorts/{short_id}/upload` - Upload a Short

#### 6. Publishing Integration (`internal/publishing/youtube.go`)
- New `UploadShort()` function
- Video description includes link to main video
- Handle YouTube Shorts-specific metadata

### Implementation Phases

**Phase 1: Analysis & Storage** (Week 1)
- AI module for manuscript analysis
- Storage schema updates
- Configuration settings

**Phase 2: CLI Interface** (Week 2)
- Publishing Details menu options
- Review and selection forms
- Upload workflow

**Phase 3: YouTube Integration** (Week 3)
- Short upload function
- Description linking
- Metadata handling

**Phase 4: API Interface** (Week 4)
- REST endpoints
- API documentation

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| AI identifies poor Short candidates | High | Medium | Include rationale in output; user selects which to use |
| Word count estimation inaccurate | Medium | Medium | Make limit configurable; conservative by default |
| YouTube API limits on Shorts | High | Low | Use standard video upload API (Shorts auto-detected by ratio) |
| Editor can't find segments in video | High | Medium | Store both line numbers and full text; clear formatting |
| Duplicate Shorts uploaded | Low | Low | Track uploaded Shorts in YAML; show upload status |

## Dependencies

### Internal
- Existing AI integration (`internal/ai/`)
- YouTube upload functionality (`internal/publishing/youtube.go`)
- Video YAML storage (`internal/storage/yaml.go`)
- Publishing Details workflow

### External
- Azure OpenAI API (existing)
- YouTube Data API v3 (existing)
- No new external dependencies required

## Open Questions

1. **Thumbnail handling**: Should we support custom Short thumbnails or use auto-generated?
   - *Decision pending user feedback*

2. **Title generation**: Should AI suggest titles for each Short, or user creates them?
   - *Lean toward AI suggestions with user override*

3. **Multi-video analysis**: Should we support "find Shorts across all videos in category"?
   - *Future enhancement, not MVP*

4. **Analytics integration**: Should we track Short → main video conversion?
   - *Future enhancement, requires YouTube Analytics API*

## Out of Scope

- Automated video editing or segment extraction (editor does this manually)
- Automated Short thumbnail generation
- YouTube Shorts-specific analytics
- Cross-platform Short publishing (TikTok, Instagram Reels)
- Automated scheduling of Short uploads
- A/B testing different Shorts from same video

## Documentation Impact

### New Documentation
- How to use Shorts analysis feature
- Best practices for selecting Short candidates
- Editor workflow for extracting Shorts from manuscripts

### Updated Documentation
- Publishing workflow documentation
- Video YAML schema reference
- Settings.yaml configuration options
- API documentation (if API mode enabled)

## Validation Strategy

### Testing Approach
- Unit tests for AI analysis module
- Integration tests for YAML storage of Shorts
- Mock YouTube API for upload testing
- End-to-end test: analyze → select → upload → verify link

### Manual Validation
- Test with diverse manuscript types (tutorial, discussion, demo)
- Verify word count estimates match actual video timing
- Confirm editor can locate segments using stored text
- Validate YouTube properly recognizes Shorts format

### Success Metrics
- AI identifies at least 3 viable candidates per manuscript
- 80%+ of user-selected Shorts successfully upload
- Editors can locate segments within 1 minute using stored text
- Short descriptions correctly link to main videos

## Milestones

- [ ] **AI Analysis Working**: Manuscript analysis identifies 3-5 Short candidates with rationale
- [ ] **Storage Schema Complete**: Video YAML properly stores Shorts metadata with all required fields
- [ ] **CLI Workflow Functional**: Users can analyze, select, and upload Shorts through Publishing Details
- [ ] **YouTube Integration Live**: Shorts upload successfully with correct metadata and main video links
- [ ] **API Endpoints Deployed**: RESTful API supports full Shorts workflow (if applicable)
- [ ] **Documentation Published**: User and editor guides available with workflow examples
- [ ] **Feature Tested & Validated**: End-to-end testing confirms reliable operation with real manuscripts
- [ ] **Feature Launched**: Available in production for all video categories

## Progress Log

### 2025-11-11
- PRD created
- GitHub issue #339 opened
- Initial architecture defined
- User requirements gathered

---

## Notes

- This feature enables a new content multiplication strategy: one long video → multiple Shorts → increased reach
- Consider this as foundation for future content repurposing features (blog snippets, social posts, etc.)
- Word count setting in settings.yaml allows tuning based on actual Short performance
