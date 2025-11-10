# PRD: Manuscript & Narration Analytics

**Status**: Draft
**Priority**: High
**GitHub Issue**: [#335](https://github.com/vfarcic/youtube-automation/issues/335)
**Created**: 2025-11-09
**Last Updated**: 2025-11-09

---

## Problem Statement

Currently, video manuscripts are written without data-driven insights into which writing patterns, structures, and phrasing lead to better video performance. We don't know:
- Which manuscript lengths correlate with better watch time
- What introduction patterns drive engagement
- Which explanation styles work best for technical content
- How demo-to-explanation ratio affects retention
- Which phrasing patterns resonate with viewers

**Result**: Manuscript-writing slash commands generate content based on generic best practices rather than patterns proven to work with our specific audience.

## Solution Overview

Build an analytics feature that:
1. Analyzes historical manuscripts (Markdown files) and identifies writing patterns
2. Correlates manuscript characteristics with video performance metrics
3. Implements video-manuscript mapping system (track which manuscript → which video)
4. Supports experiment marking in manuscripts (test specific patterns)
5. Generates data-driven writing guidelines
6. Provides specific improvements for manuscript-generating slash commands

**Key Principle**: This is a **workflow improvement activity**. We periodically analyze what works, update our writing patterns and slash commands, and all future manuscripts benefit from these insights.

## User Journey

### Primary Flow: Running Manuscript Analysis

1. User launches app and selects new menu option: **Analyze → Manuscripts**
2. App authenticates with YouTube (OAuth for analytics data)
3. App locates manuscript files based on user-configured directory
4. App fetches video performance data (last 365 days)
5. App maps manuscripts to videos using filename conventions (video ID in filename)
6. App analyzes manuscripts:
   - Word count and estimated duration
   - Structure: intro length, section count, transition patterns
   - Technical depth: jargon frequency, explanation complexity
   - Demo ratio: command/code blocks vs prose
   - Experiment markers: detect HTML comments marking experiments
   - Phrasing patterns: questions, imperatives, storytelling elements
7. AI correlates manuscript patterns with performance metrics
8. App saves files to `./tmp`:
   - `manuscript-data-YYYY-MM-DD.json` (raw data)
   - `manuscript-guidelines-YYYY-MM-DD.md` (writing guidelines)
9. App displays summary in terminal with file paths
10. User exits app

### Secondary Flow: Post-Upload Manuscript Mapping

**Current Workflow:**
1. User writes manuscript (e.g., `kubernetes-networking.md`)
2. User uploads video through app
3. Video ID received from YouTube (e.g., `dQw4w9WgXcQ`)

**Enhanced Workflow:**
4. App prompts: "Rename manuscript file to include video ID?"
5. If yes: `kubernetes-networking.md` → `kubernetes-networking-dQw4w9WgXcQ.md`
6. Mapping established for future analysis

**Alternative Approach (Metadata File):**
- Create `manuscript-video-mapping.json` with entries: `{"manuscript": "file.md", "videoId": "dQw4w9WgXcQ"}`
- No filename changes needed
- Centralized tracking

### Tertiary Flow: Experiment Marking

**In Manuscript Writing:**
```markdown
<!-- EXPERIMENT: short-intro -->
Let me show you why Kubernetes networking is broken.
<!-- /EXPERIMENT -->

## Introduction

<!-- EXPERIMENT: question-hook -->
Have you ever wondered why your pods can't communicate?
<!-- /EXPERIMENT -->

<!-- EXPERIMENT: demo-first -->
```bash
kubectl get pods -A
```

Now let's understand what's happening...
<!-- /EXPERIMENT -->
```

**During Analysis:**
- Detect experiment markers in manuscripts
- Correlate experiments with video performance
- Report: "short-intro experiments → 85% watch time (20% above avg)"
- Validate which experiments work, which don't

### Quaternary Flow: Improving Slash Commands

1. User reviews `manuscript-guidelines-YYYY-MM-DD.md`
2. Guidelines include: "Intros under 200 words → 15% better retention"
3. User opens manuscript-generating slash command files
4. User updates prompts to encode proven patterns
5. Future manuscripts automatically follow optimized patterns

### Ongoing Usage

- Write manuscripts using current slash commands
- Upload videos, track manuscript-video mapping
- Mark experiments to test new patterns
- Run analysis periodically (quarterly)
- Update slash commands based on validated patterns
- Iterative improvement cycle

## Success Criteria

### Must Have
- [ ] Fetch video analytics data (views, watch time, CTR, engagement)
- [ ] Locate and parse manuscript files (Markdown)
- [ ] Implement video-manuscript mapping system (filename or metadata)
- [ ] Analyze manuscript characteristics: length, structure, phrasing
- [ ] Detect and parse experiment markers (HTML comments)
- [ ] Correlate manuscript patterns with performance metrics
- [ ] Generate writing guidelines document with actionable recommendations
- [ ] Post-upload workflow: prompt to map manuscript to video
- [ ] New "Analyze" menu option with "Manuscripts" sub-menu works
- [ ] Graceful error handling for missing files, unmapped videos, parse errors

### Nice to Have
- [ ] Frame extraction + AI vision analysis (validate manuscript execution)
- [ ] Automated slash command update suggestions
- [ ] Track guideline effectiveness over time (before/after analysis)
- [ ] Manuscript template generator based on proven patterns
- [ ] Real-time manuscript scoring (predict performance before filming)

### Success Metrics
- Guidelines include specific, data-driven recommendations (e.g., "200-word intros → 15% better retention")
- Experiment analysis identifies which patterns work (statistical significance)
- Guidelines are actionable for improving slash commands
- Writing patterns improve over time (tracked via subsequent analyses)

## Technical Architecture

### New Components

```
internal/publishing/manuscript_mapping.go
├── MapManuscriptToVideo(manuscriptPath, videoID) → Create mapping
├── GetMappingForVideo(videoID) → Return manuscript path
├── RenameManuscriptWithVideoID(path, videoID) → Update filename
└── ManuscriptMapping struct (manuscript, videoID, mappedDate)

internal/manuscript/parser.go
├── ParseManuscript(filepath) → Extract characteristics
├── AnalyzeStructure() → Intro length, section count, transitions
├── DetectExperiments() → Find HTML comment markers
├── CalculateMetrics() → Word count, code ratio, jargon density
└── ManuscriptAnalysis struct (all characteristics)

internal/ai/analyze_manuscripts.go
├── AnalyzeManuscripts(manuscripts, performanceData) → AI analysis
├── CorrelatePatterns() → Match characteristics with performance
├── FormatManuscriptGuidelines() → Generate recommendations
└── ExperimentResult struct (experiment type, performance impact)

internal/app/
├── Add "Manuscripts" sub-menu under "Analyze"
├── HandleAnalyzeManuscripts() → Orchestrate analysis workflow
└── HandlePostUploadMapping() → Prompt for manuscript mapping after upload
```

### Data Flow

```
User: Analyze → Manuscripts
         ↓
Locate Manuscript Files (user-configured directory)
         ↓
YouTube Analytics API (video performance data)
         ↓
Map Manuscripts to Videos (filename convention or metadata)
         ↓
Parse Each Manuscript:
  - Structure analysis
  - Pattern detection
  - Experiment markers
         ↓
AI Correlation Analysis:
  - Manuscript characteristics → Performance metrics
  - Experiment results
  - Pattern identification
         ↓
Save Files:
  - ./tmp/manuscript-data-2025-11-09.json
  - ./tmp/manuscript-guidelines-2025-11-09.md
         ↓
Display Summary in Terminal
```

### Integration Points

1. **YouTube API Client** (`internal/publishing/youtube.go`)
   - Extend existing OAuth flow (already has analytics scope)
   - Use YouTube Analytics API for performance metrics
   - Reuse video data fetching from PRD #331

2. **AI Provider** (`internal/ai/provider.go`)
   - Use existing AI provider (Azure/Anthropic)
   - Similar pattern to title/thumbnail analysis
   - Analyze text patterns and correlations

3. **App Menu** (`internal/app/app.go`)
   - Add "Manuscripts" sub-menu under existing "Analyze" menu
   - Extensible pattern for future analytics types

4. **File System** (`./tmp` directory)
   - Store analysis files with timestamps
   - Already gitignored

5. **Video Upload Workflow** (`internal/publishing/youtube.go`)
   - Hook into upload completion
   - Prompt for manuscript mapping
   - Update filename or metadata

### Manuscript-Video Mapping Options

**Option A: Filename Convention** (Simpler)
- Original: `kubernetes-networking.md`
- After upload: `kubernetes-networking-dQw4w9WgXcQ.md`
- Pros: Self-contained, no extra files, easy to see mapping
- Cons: Filename changes, requires file renaming

**Option B: Metadata File** (More Flexible)
- Create `manuscript-video-mapping.json`:
```json
{
  "mappings": [
    {
      "manuscript": "manuscripts/kubernetes-networking.md",
      "videoId": "dQw4w9WgXcQ",
      "uploadDate": "2025-11-09",
      "title": "Kubernetes Networking Explained"
    }
  ]
}
```
- Pros: No filename changes, supports multiple mappings per manuscript, metadata rich
- Cons: Extra file to manage

**Recommendation**: Start with **Option A (filename)** for v1, add Option B as enhancement if needed.

### Experiment Marking System

**Syntax:**
```markdown
<!-- EXPERIMENT: experiment-id -->
Content being tested
<!-- /EXPERIMENT -->
```

**Supported Experiment Types (Examples):**
- `short-intro` - Introduction under 200 words
- `question-hook` - Start with engaging question
- `demo-first` - Demo before explanation
- `storytelling` - Narrative/story-based explanation
- `technical-deep-dive` - High jargon, advanced concepts
- `conversational-tone` - Casual, friendly phrasing

**Detection:**
- Regex pattern: `<!-- EXPERIMENT: ([\w-]+) -->(.*?)<!-- /EXPERIMENT -->`
- Extract experiment ID and content
- Track position in manuscript (intro, middle, end)
- Correlate with video performance

**Analysis Output:**
```markdown
## Experiment Results

### short-intro (tested in 12 videos)
- **Performance**: 85% avg watch time (20% above baseline)
- **Recommendation**: ✅ Keep using, proven winner

### question-hook (tested in 8 videos)
- **Performance**: 70% avg watch time (5% below baseline)
- **Recommendation**: ⚠️ Needs refinement or avoid

### demo-first (tested in 15 videos)
- **Performance**: 90% retention in first 2 minutes (30% above baseline)
- **Recommendation**: ✅ Highly effective, use more often
```

## Implementation Milestones

### Milestone 1: Video-Manuscript Mapping System
**Goal**: Establish reliable mapping between manuscripts and videos

- Implement filename-based mapping (add video ID to filename)
- Create mapping data structures
- Add post-upload prompt: "Map manuscript to this video?"
- Implement file renaming function
- Store mappings for future analysis
- Handle edge cases (existing videos, manual mapping)
- Add unit tests

**Validation**: Can reliably map manuscripts to videos after upload

---

### Milestone 2: Manuscript Parser
**Goal**: Extract characteristics from Markdown manuscripts

- Implement Markdown file reader
- Parse structure: headings, sections, paragraphs
- Calculate metrics: word count, code block ratio, list frequency
- Identify patterns: intro length, transition phrases, question usage
- Detect experiment markers (HTML comments)
- Handle various Markdown styles and edge cases
- Add unit tests with sample manuscripts

**Validation**: Parser accurately extracts all manuscript characteristics

---

### Milestone 3: YouTube Analytics Integration
**Goal**: Fetch video performance data and match with manuscripts

- Reuse YouTube Analytics API integration from PRD #331
- Fetch performance data for mapped videos
- Match manuscripts to performance metrics via video ID
- Handle unmapped videos gracefully (skip or warn)
- Filter out videos without manuscripts
- Add unit tests with mocked API responses

**Validation**: Can fetch and correlate manuscript-video performance data

---

### Milestone 4: Manuscript Analysis Engine
**Goal**: AI analyzes manuscript patterns and correlates with performance

- Create `analyze_manuscripts.go` module
- Implement pattern identification (structure, phrasing, style)
- Correlate manuscript characteristics with performance metrics
- Identify high-performing patterns (e.g., "short intros → better retention")
- Identify anti-patterns (e.g., "long prose blocks → lower watch time")
- Account for video age bias and outliers
- Add unit tests

**Validation**: AI generates specific, data-driven insights about writing patterns

---

### Milestone 5: Experiment Analysis
**Goal**: Analyze marked experiments and report effectiveness

- Detect experiment markers in manuscripts
- Group videos by experiment type
- Calculate performance by experiment (avg watch time, retention, engagement)
- Compare against baseline (non-experiment videos)
- Determine statistical significance
- Report experiment results with recommendations
- Add unit tests

**Validation**: Experiment analysis identifies winning and losing patterns

---

### Milestone 6: Guidelines Generation
**Goal**: Output actionable writing guidelines document

- Create guidelines template (structured, scannable format)
- Include: optimal characteristics, high-performing patterns, anti-patterns
- Add experiment results section
- Format for easy consumption (headers, bullet points, examples)
- Save as `manuscript-guidelines-{date}.md`
- Include metadata (date, video count, experiments analyzed)
- Add specific slash command improvement suggestions

**Validation**: Guidelines document is actionable and ready to improve workflow

---

### Milestone 7: Menu Integration & UX
**Goal**: User can run analysis from app menu

- Add "Manuscripts" sub-menu under "Analyze"
- Wire manuscript analysis workflow through app layer
- Display progress indicators (locating files, analyzing, generating)
- Show summary after completion (file paths, key insights)
- Handle configuration (manuscript directory path)
- Handle OAuth re-authentication if needed

**Validation**: User can run analysis end-to-end from menu

---

### Milestone 8: Post-Upload Mapping Workflow
**Goal**: Prompt user to map manuscript after video upload

- Hook into video upload completion
- Prompt: "Map a manuscript to this video?"
- If yes: Ask for manuscript path or show list
- Rename file with video ID (or update metadata)
- Confirm mapping created
- Skip option for videos without manuscripts

**Validation**: Post-upload mapping workflow is smooth and reliable

---

### Milestone 9: Production Ready
**Goal**: Feature is stable and ready for regular use

- Comprehensive error handling (missing files, parse errors, API failures)
- Logging for debugging
- Performance optimization (parallel parsing, AI token usage)
- Configuration management (manuscript directory setting)
- Final end-to-end testing with real manuscripts
- Documentation in CLAUDE.md (optional)

**Validation**: Feature works reliably with production data

---

## Dependencies

### External
- YouTube Analytics API v2 (already integrated in PRD #331)
- YouTube Data API v3 (already integrated)
- AI Provider (Azure OpenAI or Anthropic - already integrated)

### Internal
- Existing OAuth implementation in `internal/publishing/youtube.go`
- Existing AI provider in `internal/ai/provider.go`
- Existing menu system in `internal/app/app.go`
- Video analytics fetching from PRD #331 (reusable)

### New Capabilities Required
- Markdown file parsing
- Pattern detection (text analysis)
- Experiment marker detection (HTML comment parsing)
- Manuscript-video mapping system

### Configuration Required
- Manuscript directory path (e.g., `~/manuscripts/` or `./content/scripts/`)
- Mapping storage location (filename convention or metadata file)

## Risks & Mitigation

### Risk: Unmapped Historical Videos
**Impact**: Medium
**Probability**: High
**Mitigation**:
- Manual mapping tool for historical manuscripts
- Gracefully skip unmapped videos in analysis
- Warn user about unmapped videos, suggest mapping
- Provide bulk mapping utility (match by date/title similarity)

### Risk: Manuscript Format Variability
**Impact**: Medium
**Probability**: Medium
**Mitigation**:
- Support multiple Markdown styles
- Robust parsing with error handling
- Fallback to basic metrics if advanced parsing fails
- Document recommended manuscript format

### Risk: Small Sample Size (Experiments)
**Impact**: Low
**Probability**: Medium
**Mitigation**:
- Require minimum sample size for experiment conclusions (e.g., 5 videos)
- Report confidence levels
- Warn when sample size is too small
- Encourage testing over multiple videos

### Risk: Correlation vs Causation
**Impact**: Medium
**Probability**: Medium
**Mitigation**:
- AI analysis notes correlation, not causation
- Encourage experimentation to validate patterns
- Include statistical significance in recommendations
- Provide raw data so user can verify insights

### Risk: Manuscript Directory Configuration
**Impact**: Low
**Probability**: Low
**Mitigation**:
- Prompt for directory during first run
- Store in settings.yaml
- Allow reconfiguration via settings menu
- Clear error messages if directory not found

## Open Questions

1. **Mapping approach**: Filename convention vs metadata file?
   - **Decision**: Filename convention for v1 (simpler), metadata file as future enhancement

2. **Historical manuscripts**: How to handle existing manuscripts without video IDs?
   - **Decision**: Provide manual mapping tool, match by date/title similarity

3. **Experiment marker syntax**: HTML comments vs other format?
   - **Decision**: HTML comments (invisible in rendered Markdown, easy to parse)

4. **Slash command improvements**: Auto-update vs manual?
   - **Decision**: Manual for v1 (guidelines suggest changes), auto-update as nice-to-have

5. **Frame extraction**: Include in v1 or defer?
   - **Decision**: Defer to future enhancement (experiment marking is sufficient)

## Future Enhancements

**Phase 2: Frame Extraction Validation**
- Extract key frames from videos (every 30s or scene changes)
- Use Claude Vision to analyze: screen time, demo execution, visual aids
- Compare manuscript intent vs video execution
- Identify where delivery diverges from script
- Recommendations for filming/editing approach

**Phase 3: Automated Slash Command Updates**
- AI generates updated prompts for slash commands
- User reviews and approves changes
- Direct integration: guidelines → prompt updates
- Version control for prompts (track changes)

**Phase 4: Real-Time Scoring**
- Analyze manuscript before filming
- Predict performance based on historical patterns
- Suggest improvements: "Consider shortening intro"
- Preview estimated watch time, retention

**Phase 5: Manuscript Template Generator**
- Generate optimal manuscript structure
- Based on topic type (tutorial, comparison, opinion)
- Encode all proven patterns
- Interactive: user answers questions, template adapts

**Phase 6: Cross-Feature Integration**
- Correlate manuscript patterns with title/thumbnail performance
- Identify synergies: "ArgoCD tutorials with bold thumbnails + short intros = highest performance"
- Holistic content optimization

---

## Progress Log

### [Date] - Session [N]: [Milestone] Complete
**Duration**: ~X hours
**Status**: X of 9 milestones complete (X%)

#### ✅ Milestone [N]: [Name] (100%)
**Files Created:**
- [List files]

**Implementation Details:**
- [Key implementation notes]

**Testing & Validation:**
- [Test results]

**Technical Decisions Made:**
- [Decisions and rationale]

---

## Cross-References

**Related PRDs:**
- **PRD #331**: YouTube Title Analytics & Optimization (completed) - Similar analytics pattern
- **PRD #333**: Thumbnail Analytics & Competitive Benchmarking (planned) - Complementary visual analysis
- **PRD #334**: Thumbnail A/B Test Variation Generator (planned) - Similar experiment approach

**Integration Opportunities:**
- Combine manuscript + title + thumbnail analysis for holistic insights
- Cross-reference: "Videos with pattern X in manuscript + pattern Y in title = best performance"

---

## References

- [YouTube Analytics API Documentation](https://developers.google.com/youtube/analytics)
- [Markdown Parsing in Go](https://github.com/yuin/goldmark)
- [PRD #331 - Title Analytics](prds/done/331-youtube-title-analytics.md) - Reference analytics pattern
- [GitHub Issue #335](https://github.com/vfarcic/youtube-automation/issues/335)
