# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

### Building and Running
```bash
# Build for local testing
make build-local
# or
just build-local
# or
go build -o youtube-release ./cmd/youtube-automation

# Build for all platforms
make build

# Run CLI mode (default)
./youtube-release

# Run API server mode
./youtube-release --api-enabled --api-port 8080

# Clean build artifacts
make clean
# or
just clean
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
just test
# or
go test ./... -cover

# Generate detailed coverage report
./scripts/coverage.sh

# Run specific package tests
go test ./internal/publishing/...

# Run specific test function with verbose output
go test -v -run TestUploadVideo ./internal/publishing/
```

### Development Tools
```bash
# Check for brittle tests
./scripts/find_brittle_tests.sh

# Version management
make bump-patch    # Bump patch version
make bump-minor    # Bump minor version
make bump-major    # Bump major version
```

## Architecture Overview

### Core System Design
The YouTube Automation Tool is built around a **phase-based video lifecycle management system** with both CLI and REST API interfaces. Videos progress through 6 distinct phases from idea to post-publish activities.

### Key Architectural Components

#### 1. Video Lifecycle Phases (0-7)
- **Phase 0-6**: Active workflow phases (Ideas → Started → Material Done → Edit Requested → Publish Pending → Published → Delayed)
- **Phase 7**: Sponsored/Blocked videos (special handling)
- Each phase has specific completion criteria and field requirements

#### 2. Multi-Interface Architecture
- **CLI Mode**: Interactive terminal interface with huh forms (`internal/app/`)
- **API Mode**: REST server for programmatic access (`internal/api/`)
- **Shared Core**: Common business logic in `internal/service/` and `internal/storage/`

#### 3. Field Completion System
Uses reflection-based completion tracking via struct tags in `internal/storage/yaml.go`:
```go
type Video struct {
    Date string `json:"date" completion:"filled_only"`
    Code bool   `json:"code" completion:"true_only"`
    // ... other fields
}
```

Completion criteria: `filled_only`, `true_only`, `false_only`, `conditional_sponsorship`, `conditional_sponsors`, `empty_or_filled`, `no_fixme`

#### 4. AI Content Generation
- Integrated Azure OpenAI for titles, descriptions, tags, and tweets
- Two API approaches: traditional (JSON payload) and optimized (URL parameters)
- Located in `internal/ai/` with individual modules per content type

#### 5. Publishing Integration
- **YouTube API**: Automated video uploads with OAuth2 (`internal/publishing/youtube.go`)
- **Hugo Integration**: Blog post generation (`internal/publishing/hugo.go`)
- **Solved Chicken-and-Egg Problem**: Video descriptions now include correct Hugo URLs upfront by pre-constructing them during upload

#### 6. Social Media Distribution
- BlueSky, LinkedIn, Slack posting capabilities
- Platform-specific modules in `internal/platform/`

### Data Flow Architecture

#### 1. Storage Layer (`internal/storage/`)
- YAML-based persistence for video metadata
- Index file tracks all videos across categories
- Individual YAML files per video with complete metadata

#### 2. Service Layer (`internal/service/`)
- `VideoService`: Unified data operations for CLI and API
- Handles CRUD operations, phase transitions, manuscript processing
- Abstracts storage details from interfaces

#### 3. Business Logic Layer
- **Aspect System** (`internal/aspect/`): Dynamic form generation and completion tracking
- **Video Manager** (`internal/video/`): Phase calculation and workflow management
- **Workflow** (`internal/workflow/`): Constants and phase definitions

#### 4. Interface Layer
- **CLI**: Menu-driven interface with phase-specific forms
- **API**: RESTful endpoints mirroring CLI functionality
- **Shared Validation**: Both interfaces use identical business rules

### Configuration System
- `settings.yaml` for global configuration
- Environment variables for sensitive data (API keys, passwords)
- CLI flags with YAML path mapping (`internal/configuration/`)

### Hugo Integration Details
The system generates Hugo blog posts with deterministic URL construction:
- **URL Pattern**: `https://devopstoolkit.live/{category}/{sanitized-title}`
- **Title Sanitization**: Spaces→hyphens, remove special chars, lowercase
- **Path Mapping**: Manuscript categories map to Hugo content directories
- **Cross-Referencing**: Videos include blog post URLs, blog posts include YouTube embeds

### Testing Architecture
- **Mock-Based Testing**: External services (YouTube API, email) are mocked
- **Coverage Goal**: 80% test coverage with `./scripts/coverage.sh`
- **Test Organization**: Tests alongside source code, mocks in `pkg/mocks/`
- **Fixture System**: Test data in `pkg/testutil/testdata/`

### Key Patterns to Follow

#### 1. Phase-Based Development
When adding new functionality, consider which phase(s) it affects and update:
- Field completion criteria in struct tags
- Aspect definitions for dynamic forms
- Phase transition logic in video manager

#### 2. Interface Consistency
Both CLI and API should provide identical functionality:
- Share business logic through service layer
- Use same validation rules and error handling
- Maintain feature parity between interfaces

#### 3. AI Integration Pattern
When adding new AI features:
- Create dedicated module in `internal/ai/`
- Provide both traditional (JSON) and optimized (URL params) API endpoints
- Include proper error handling and retry logic

#### 4. Hugo URL Construction
For any Hugo-related functionality, use the established pattern:
- Extract category from manuscript path using `GetCategoryFromFilePath()`
- Sanitize titles using `SanitizeTitle()`
- Construct URLs using `ConstructHugoURL()`

#### 5. Test-First Development (MANDATORY)
**⚠️ CRITICAL: Tests are NOT optional. Every code change MUST include tests.**

When writing ANY new functionality or modifying existing code:
1. **ALWAYS write tests** - No exceptions
2. **Test BEFORE considering the work complete** - Untested code is incomplete code
3. **Update existing tests** if behavior changes
4. **Cover edge cases**: success paths, error paths, boundary conditions
5. **Use table-driven tests** for multiple scenarios (Go best practice)

**Test Checklist** (use this for every PR/feature):
- [ ] New functions have corresponding test functions
- [ ] Success cases are tested
- [ ] Error cases are tested (nil checks, invalid inputs, file I/O errors, etc.)
- [ ] Edge cases are covered (empty inputs, large inputs, etc.)
- [ ] External dependencies are mocked (YouTube API, AI providers, file system when appropriate)
- [ ] Tests run successfully with `go test ./...`
- [ ] Coverage meets 80% threshold (verify with `./scripts/coverage.sh`)

**Common Testing Patterns:**
```go
// Table-driven tests (preferred)
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {name: "valid input", input: validInput, want: expectedOutput, wantErr: false},
        {name: "empty input", input: emptyInput, want: nil, wantErr: true},
        {name: "error case", input: badInput, want: nil, wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

**File I/O Testing:**
- Use `t.TempDir()` for temporary directories
- Clean up resources in tests
- Test both successful file writes and error conditions

**Why This Matters:**
- Tests catch bugs before they reach users
- Tests document expected behavior
- Tests enable confident refactoring
- 80% coverage goal maintains code quality
- CI/CD pipeline depends on passing tests

This architecture supports the tool's evolution from simple video management to comprehensive content creation workflow automation while maintaining clean separation of concerns and interface consistency.