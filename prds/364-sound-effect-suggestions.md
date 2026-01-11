# PRD: AI-Powered Sound Effect Suggestions from Manuscript

**Issue**: #364
**Status**: Not Started
**Priority**: Medium
**Created**: 2025-01-11
**Last Updated**: 2025-01-11
**Depends On**: None

---

## Problem Statement

Adding sound effects to videos is manual and time-consuming. The editor must:
- Decide where sound effects would enhance the video
- Find or create appropriate sounds
- Manually place them in the timeline

Meanwhile, the manuscript already contains structural hints that could inform sound effect placement:
- Code blocks (`\`\`\`sh`) → typing sounds
- Output blocks → terminal/processing sounds
- Mermaid diagrams → reveal/transition sounds
- Section headings → transition effects
- Narrative sentiment → dramatic, warning, success sounds

This information is lost, requiring the editor to re-analyze content that's already structured.

## Proposed Solution

Claude analyzes manuscripts to automatically suggest sound effects:

1. **Effect Library**: Maintain a reusable library of pre-generated sound effects
2. **Manuscript Analysis**: Claude parses structure and sentiment to identify effect opportunities
3. **Smart Selection**: Use existing effects from library, or generate new ones via ElevenLabs API
4. **Editor Instructions**: Insert `TODO: effect [EFFECT_FILE_NAME]` markers in manuscript
5. **Effect Generation**: New effects generated via ElevenLabs Sound Effects API and added to library

### User Journey

**Current State (Manual)**:
1. Editor receives manuscript and raw video
2. Editor manually identifies moments that need effects
3. Editor searches for appropriate sound effects
4. Editor manually places effects in timeline
5. No consistency across videos

**After (With This Feature)**:
1. Creator runs "Analyze for Sound Effects" on manuscript
2. Claude identifies moments based on structure and sentiment
3. System selects from library or generates new effects via ElevenLabs
4. Manuscript updated with `TODO: effect [filename]` markers
5. Editor sees clear instructions and pre-generated effect files
6. Consistent sound design across videos

## Success Criteria

### Must Have (MVP)
- [ ] Effect library file storing reusable effects with metadata
- [ ] Claude manuscript analysis identifying effect opportunities
- [ ] Detection of code blocks, outputs, diagrams, headings
- [ ] Basic sentiment analysis for dramatic/warning/success moments
- [ ] ElevenLabs API integration for generating new effects
- [ ] Manuscript updated with `TODO: effect [filename]` markers
- [ ] New effects automatically added to library

### Nice to Have (Future)
- [ ] Effect preview/playback in CLI
- [ ] Effect intensity/volume suggestions
- [ ] Learning from editor feedback (which effects were kept/removed)
- [ ] Batch processing multiple manuscripts
- [ ] Effect timing duration suggestions

## Technical Scope

### Core Components

#### 1. Effect Library (`internal/effects/library.go`)
```go
type SoundEffect struct {
    Name        string   `yaml:"name"`        // e.g., "keyboard-typing"
    FilePath    string   `yaml:"filePath"`    // e.g., "effects/keyboard-typing.mp3"
    Category    string   `yaml:"category"`    // e.g., "typing", "transition", "alert"
    Description string   `yaml:"description"` // When to use this effect
    Tags        []string `yaml:"tags"`        // Searchable tags
    Duration    float64  `yaml:"duration"`    // Duration in seconds
}

// LoadLibrary loads effects from library file
func LoadLibrary(path string) (*Library, error)

// FindEffect finds best matching effect for a context
func (l *Library) FindEffect(category string, context string) *SoundEffect

// AddEffect adds a new effect to the library
func (l *Library) AddEffect(effect SoundEffect) error
```

#### 2. Manuscript Analyzer (`internal/effects/analyzer.go`)
```go
type EffectSuggestion struct {
    LineNumber  int    // Where in manuscript
    TriggerType string // "code_block", "output", "diagram", "heading", "sentiment"
    Context     string // The text/content that triggered this
    Category    string // Effect category needed
    EffectName  string // Chosen effect from library (or "GENERATE_NEW")
}

// AnalyzeManuscript identifies opportunities for sound effects
func AnalyzeManuscript(ctx context.Context, content string) ([]EffectSuggestion, error)
```

#### 3. ElevenLabs Sound Effects Client (`internal/effects/elevenlabs.go`)
```go
// GenerateEffect creates a new sound effect via ElevenLabs API
// POST /v1/sound-generation
func (c *Client) GenerateEffect(ctx context.Context, prompt string, duration float64) ([]byte, error)

// SaveEffect saves generated effect to file and adds to library
func (c *Client) SaveEffect(audioData []byte, name string, category string) (*SoundEffect, error)
```

#### 4. Manuscript Updater (`internal/effects/updater.go`)
```go
// InsertEffectMarkers adds TODO markers to manuscript
func InsertEffectMarkers(content string, suggestions []EffectSuggestion) (string, error)
```

### Effect Categories

| Category | Trigger | Example Effects |
|----------|---------|-----------------|
| `typing` | `\`\`\`sh`, `\`\`\`bash` | keyboard-typing, mechanical-keyboard |
| `output` | `\`\`\`` (no language) | terminal-beep, processing-sound |
| `diagram` | `\`\`\`mermaid` | reveal-whoosh, diagram-appear |
| `transition` | `## Heading` | section-transition, whoosh |
| `warning` | "careful", "warning", "don't", "danger" | alert-sound, caution-tone |
| `success` | "works", "done", "success", "perfect" | success-chime, positive-ding |
| `error` | "error", "failed", "broken" | error-buzz, failure-tone |
| `dramatic` | Sentiment analysis | tension-build, revelation-sting |

### Configuration

**settings.yaml additions:**
```yaml
effects:
  libraryPath: "effects/library.yaml"    # Path to effect library
  outputDir: "effects/generated/"        # Where to save new effects
  enabled: true                          # Enable/disable feature
```

**effects/library.yaml example:**
```yaml
effects:
  - name: keyboard-typing
    filePath: effects/keyboard-typing.mp3
    category: typing
    description: Mechanical keyboard typing sound
    tags: [typing, keyboard, code]
    duration: 2.0

  - name: section-transition
    filePath: effects/section-transition.mp3
    category: transition
    description: Whoosh sound for section changes
    tags: [transition, whoosh, section]
    duration: 0.5

  - name: warning-alert
    filePath: effects/warning-alert.mp3
    category: warning
    description: Subtle alert for warning messages
    tags: [warning, alert, caution]
    duration: 1.0
```

### Implementation Phases

**Phase 1: Effect Library System**
- Define library data structure
- Load/save library from YAML
- Search/match effects by category and context
- Unit tests for library operations

**Phase 2: Manuscript Analysis**
- Parse markdown structure (code blocks, headings, etc.)
- Identify effect opportunities from structure
- Basic sentiment analysis for narrative moments
- Generate EffectSuggestion list

**Phase 3: ElevenLabs Integration**
- Implement Sound Effects API client
- Generate effects from text prompts
- Save to file and update library
- Test mode configuration (like dubbing PRD)

**Phase 4: Manuscript Update**
- Insert `TODO: effect [filename]` markers
- Preserve original manuscript formatting
- Handle edge cases (nested blocks, etc.)

**Phase 5: CLI Integration**
- Add "Analyze for Sound Effects" menu option
- Display suggestions before applying
- Allow user to accept/reject/modify suggestions

**Phase 6: Testing & Validation**
- End-to-end testing with real manuscripts
- Validate effect quality from ElevenLabs
- Test editor workflow with TODO markers

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Over-suggesting effects | Medium | High | Conservative defaults; user review before applying |
| Poor effect quality from ElevenLabs | Medium | Medium | Build library of tested effects; regenerate if poor |
| Sentiment analysis misses context | Low | Medium | Focus on structural triggers; sentiment as enhancement |
| ElevenLabs API costs | Low | Medium | Reuse library effects; generate sparingly |
| Editor ignores TODO markers | Low | Low | Clear, consistent format; documentation |

## Dependencies

### Internal
- Manuscript storage/processing (existing)
- AI provider for sentiment analysis (`internal/ai/`)
- Configuration system (`internal/configuration/`)

### External
- ElevenLabs Sound Effects API (new)
- Claude AI for manuscript analysis (existing)

## Out of Scope

- Automatic placement in video timeline (editor's job)
- Volume/mixing adjustments
- Effect playback in CLI
- Video analysis for timing
- Real-time effect generation during recording

## Validation Strategy

### Testing Approach
- Unit tests for library operations
- Unit tests for manuscript parsing
- Integration tests with mock ElevenLabs API
- Table-driven tests for effect matching

### Manual Validation
- Test with real manuscripts from past videos
- Verify TODO markers are clear for editor
- Validate generated effect quality
- Confirm library grows appropriately over time

## Milestones

- [ ] **Effect Library System Working**: Load, search, save effects
- [ ] **Manuscript Analysis Functional**: Identifies effect opportunities from structure
- [ ] **ElevenLabs Integration Complete**: Can generate new effects via API
- [ ] **Manuscript Markers Working**: Inserts `TODO: effect [filename]` correctly
- [ ] **CLI Integration Complete**: Menu option to analyze manuscripts
- [ ] **End-to-End Workflow Validated**: Full flow tested with real manuscript

## Progress Log

### 2025-01-11
- PRD created
- GitHub issue #364 opened
- Key decisions made:
  - Effect library stored in YAML file
  - Reuse existing effects when possible, generate new via ElevenLabs
  - Editor instructions via `TODO: effect [filename]` format
  - Integration point TBD during implementation
- Identified effect categories based on manuscript structure

---

## Notes

- Start with a small library of ~10-15 core effects
- Effects should be subtle (not distracting from narration)
- Typing effects should be low volume, under narration
- Consider effect duration relative to content length
- Library can grow organically as new needs arise
