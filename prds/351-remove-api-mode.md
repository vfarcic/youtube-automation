# PRD: Remove Unused API Mode

**Issue**: #351
**Status**: In Progress
**Created**: 2025-11-29
**Last Updated**: 2026-01-22

## Problem Statement

The YouTube Automation Tool currently supports two modes: CLI (interactive terminal) and API (REST server). The API mode was designed to provide programmatic access to the tool's functionality, but:

1. **Not Being Used**: No one is currently using the API mode, and it has never been used in production
2. **Maintenance Burden**: Maintaining two interfaces doubles the testing surface and complexity
3. **Code Overhead**: API-specific code adds unnecessary complexity to the codebase
4. **Documentation Debt**: API documentation requires maintenance alongside CLI docs
5. **Feature Parity**: Every new feature must be implemented in both interfaces

Since the API mode provides no value and only adds complexity, it should be removed entirely.

## Solution Overview

Completely remove all API-related code, configuration, documentation, and tests. Keep only the CLI interface, which is actively used. Simplify the service layer if it was only abstracting for API/CLI separation and can now be streamlined for CLI-only usage.

## User Stories

### Primary User Story
**As a** maintainer of the YouTube Automation Tool
**I want** to remove unused API code
**So that** the codebase is simpler, easier to maintain, and has less technical debt

### User Impact
- **End Users**: No impact - API mode was not being used
- **Developers**: Simplified codebase with reduced maintenance burden
- **Future Development**: Faster feature development without API parity requirements

## Success Criteria

### Must Have
- [x] Complete removal of `internal/api/` directory
- [x] Removal of API-related CLI flags (`--api-enabled`, `--api-port`)
- [x] Removal of API-related configuration from `settings.yaml` (if any)
- [x] Removal of all API documentation (`docs/api-manual-testing.md`, API sections in README, CLAUDE.md)
- [x] Cleanup of main.go to remove API server initialization
- [x] Service layer remains functional for CLI usage
- [x] All existing CLI functionality works after removal
- [x] All tests pass after removal
- [x] No broken references or imports remain

### Should Have
- [ ] Simplification of service layer if it was primarily for API/CLI abstraction
- [x] Code coverage maintained at 80% or higher
- [x] Git history preserved (no force pushes or history rewriting)

### Could Have
- [ ] Refactoring opportunities identified during removal
- [ ] Performance improvements from simplified architecture
- [ ] Migration notes for hypothetical future API needs

### Won't Have (This Release)
- Replacement API system
- Gradual deprecation (complete removal in one PR)
- API compatibility layer
- API usage analytics (none existed)

## Technical Approach

### Files to Remove Completely
1. **API Implementation**
   - `internal/api/server.go`
   - `internal/api/handlers.go`
   - `internal/api/ai_handlers_test.go`
   - `internal/api/ai_handlers_performance_test.go`
   - `internal/api/handlers_test.go`
   - Entire `internal/api/` directory

2. **API Documentation**
   - `docs/api-manual-testing.md`
   - API sections in `README.md`
   - API sections in `CLAUDE.md`

3. **Configuration Files** (if any API-specific settings exist)
   - API configuration in `settings.yaml`
   - API-related environment variables documentation

### Files to Modify

1. **`cmd/youtube-automation/main.go`**
   - Remove `internal/api` import
   - Remove API server initialization code
   - Remove API mode conditional logic
   - Simplify to CLI-only execution

2. **`internal/configuration/cli.go`**
   - Remove `--api-enabled` flag
   - Remove `--api-port` flag
   - Remove any API-related configuration structs

3. **`internal/service/video_service.go`** (if simplification opportunities exist)
   - Review if service layer can be simplified
   - Remove any API-specific logic
   - Keep all CLI-required functionality
   - Maintain interface stability for CLI usage

4. **Documentation Files**
   - `README.md`: Remove API mode documentation
   - `CLAUDE.md`: Remove API architecture sections
   - Update examples to show CLI-only usage

### Service Layer Strategy

**Current State**: `internal/service/` is used by both CLI (`internal/app/`) and API (`internal/api/`)

**After Removal**:
- Keep service layer if CLI actively uses it for business logic abstraction
- Review `video_service.go` for any API-specific methods or logic
- Simplify if service layer was primarily for API/CLI separation
- Consider inlining simple operations if abstraction is no longer needed

**Decision Point**: During implementation, evaluate whether service layer provides value for CLI-only usage or if it's over-engineering

### Testing Strategy

1. **Pre-Removal Validation**
   - Document current CLI functionality
   - Run full test suite to establish baseline
   - Verify all CLI features work as expected

2. **During Removal**
   - Remove API code incrementally
   - Run tests after each major deletion
   - Fix broken imports and references immediately

3. **Post-Removal Validation**
   - Run full test suite (should pass 100%)
   - Manual CLI testing of all major features
   - Verify test coverage remains ≥80%
   - Check for orphaned code or unused imports

## Milestones

### 1. API Code Removal
- [x] Remove entire `internal/api/` directory
- [x] Remove API imports from `main.go`
- [x] Remove API server initialization logic
- [x] Fix resulting compilation errors
- **Validation**: Code compiles without API references ✅

### 2. Configuration & Flag Cleanup
- [x] Remove API flags from CLI configuration
- [x] Remove API settings from `settings.yaml` (if any)
- [x] Update configuration documentation
- **Validation**: CLI runs without API flags; no API config remains ✅

### 3. Documentation Cleanup
- [x] Remove `docs/api-manual-testing.md`
- [x] Remove API sections from `README.md`
- [x] Remove API architecture sections from `CLAUDE.md`
- [x] Update all examples to show CLI-only usage
- **Validation**: No API documentation remains; docs are CLI-focused ✅

### 4. Service Layer Review & Simplification
- [ ] Review `internal/service/video_service.go` for API-specific code
- [ ] Remove or simplify API-specific logic
- [ ] Ensure all CLI functionality remains intact
- [ ] Update service tests if needed
- **Validation**: Service layer works correctly for CLI; no API artifacts remain

### 5. Testing & Validation
- [ ] Run full test suite (all tests pass)
- [ ] Manual CLI testing of all features
- [ ] Verify test coverage ≥80%
- [ ] Check for unused imports or orphaned code
- [ ] Update any broken tests
- **Validation**: All tests pass, CLI works perfectly, coverage maintained

## Dependencies

### Internal Dependencies
- CLI application (`internal/app/`)
- Service layer (`internal/service/`)
- Configuration system (`internal/configuration/`)

### External Dependencies
None - this is purely internal refactoring

### Blocking Dependencies
None - can be completed independently

## Risks & Mitigations

### Risk 1: Breaking CLI Functionality
**Impact**: High
**Likelihood**: Low
**Mitigation**: Comprehensive testing before and after removal; incremental deletion with test runs

### Risk 2: Accidentally Removing Shared Code
**Impact**: High
**Likelihood**: Low
**Mitigation**: Careful review of service layer; only remove API-specific code; keep all CLI-used code

### Risk 3: Test Coverage Degradation
**Impact**: Medium
**Likelihood**: Low
**Mitigation**: Run coverage reports before and after; ensure ≥80% maintained

### Risk 4: Undiscovered API Usage
**Impact**: Medium
**Likelihood**: Very Low
**Mitigation**: User confirmed no production usage; git history preserves code if needed

### Risk 5: Future API Needs
**Impact**: Low
**Likelihood**: Low
**Mitigation**: Code preserved in git history; can be restored if needed; simpler to rebuild than maintain unused code

## Open Questions

1. **Are there any API-specific settings in `settings.yaml`?**
   - Will verify during implementation
   - Remove if found

2. **Should service layer be completely removed or just simplified?**
   - Decision point during Milestone 4
   - Keep if CLI benefits from abstraction
   - Simplify/remove if it's only for API/CLI separation

3. **Are there any environment variables specific to API mode?**
   - Will check during implementation
   - Document and remove if found

## Out of Scope

- Building a new/different API system
- Gradual deprecation with warnings (complete removal)
- API usage analytics or tracking (never existed)
- Migration tooling (no users to migrate)
- Backwards compatibility (no API users)
- Future API planning (can be addressed later if needed)

## Future Enhancements

- N/A - This is a removal task
- If API mode is needed in the future, it can be rebuilt from scratch based on actual requirements
- Git history preserves all removed code for reference

## Progress Log

### 2026-01-22
- **Milestone 3 Complete**: Documentation Cleanup
  - Deleted `docs/api-manual-testing.md` (1,678 lines of API testing documentation)
  - Deleted `docs/api-optimization-deployment.md` (303 lines of API deployment docs)
  - Updated `README.md`: removed API mode section, endpoints, frontend integration (~300 lines)
  - Updated `CLAUDE.md`: removed API server commands, multi-interface architecture, API patterns
  - All documentation now CLI-focused with no API references

- **Milestone 2 Complete**: Configuration & Flag Cleanup
  - Removed `SettingsAPI` struct from `cli.go`
  - Removed `API` field from `Settings` struct
  - Removed `--api-port` and `--api-enabled` CLI flags
  - Removed default API settings block
  - Removed `api:` section from `settings.yaml`
  - Verified no dead code remains (go vet passes, no orphaned references)
  - All tests pass, CLI runs without API flags

### 2026-01-21
- **Milestone 1 Complete**: API Code Removal
  - Removed entire `internal/api/` directory (5 files)
  - Simplified `main.go`: removed API import, `startAPIServer()` function, and conditional logic
  - Removed 7 unused imports from main.go
  - All tests pass, code compiles successfully

### 2025-11-29
- PRD created
- GitHub issue #351 opened
- Identified all API-related files and references
- Defined 5 major milestones for complete removal

---

**Next Steps**: Milestone 4 - Service Layer Review & Simplification
