<context>
# Overview  
This document outlines the requirements for refactoring the video animation extraction functionality in the YouTube Automation tool. Currently, there is a deprecated method `getAnimationsFromScript` in repo.go that is marked with "TODO: Remove".

# Core Features  
- Remove deprecated code
- Ensure all animation extraction uses the preferred method
- Maintain backward compatibility where needed
- Update any code that depends on the deprecated function
</context>
<PRD>
# YouTube Automation Project - Animation Extraction Refactoring Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process. The codebase currently contains a deprecated method `getAnimationsFromScript` in repo.go that is marked with "TODO: Remove". This PRD outlines the requirements for safely removing this method and ensuring all animation extraction uses the preferred approach.

## Refactoring Objectives
- Remove the deprecated `getAnimationsFromScript` method
- Ensure all animation extraction uses the preferred `getAnimationsFromMarkdown` method
- Maintain compatibility with existing data formats
- Improve code maintainability and reduce technical debt
- Update tests to reflect the changes

## Technical Implementation Requirements

### 1. Code Analysis

#### 1.1 Usage Assessment
- Identify all places where `getAnimationsFromScript` is called
- Document format differences between script and markdown animation extraction
- Determine if any unique functionality exists in the deprecated method
- Assess the impact of removal on existing videos and scripts

#### 1.2 Dependency Analysis
- Check if any interfaces rely on the deprecated method
- Identify any external components expecting script-specific behavior
- Verify if any file formats depend specifically on the script animation format
- Document any differences in parsing logic that might affect output

### 2. Refactoring Implementation

#### 2.1 Method Consolidation
- Remove the `getAnimationsFromScript` method
- Update the `GetAnimations` method to handle both script and markdown files with a single approach
- Ensure proper error handling during file format detection
- Maintain compatibility with existing file types

#### 2.2 Format Standardization
- Standardize on a single animation format for all file types
- Update the file suffix detection in the `GetAnimations` method
- Implement consistent parsing rules for all animation types
- Document the standardized format for future reference

### 3. Testing and Validation

#### 3.1 Test Updates
- Update existing tests to reflect the removal of the deprecated method
- Add tests to verify compatible behavior with script files
- Create tests for edge cases in format detection
- Validate output consistency between old and new implementations

#### 3.2 Regression Testing
- Test with a variety of existing script and markdown files
- Verify that animation extraction works correctly for all file types
- Ensure section detection remains accurate
- Validate that TODO handling is consistent

### 4. Documentation Updates

#### 4.1 Code Documentation
- Update comments to reflect the consolidated approach
- Add explanations for format detection logic
- Document the standardized animation format
- Add migration notes for any dependent code

#### 4.2 User Documentation
- Update any user-facing documentation that references animation extraction
- Provide guidance on preferred file formats
- Document the supported animation markup formats
- Update examples to reflect the current implementation

## Implementation Details

### Current Code Analysis
```go
// Current implementation in repo.go
func (r *Repo) GetAnimations(filePath string) (animations, sections []string, err error) {
    if strings.HasSuffix(filePath, ".sh") {
        return r.getAnimationsFromScript(filePath)
    }
    return r.getAnimationsFromMarkdown(filePath)
}

// TODO: Remove
func (r *Repo) getAnimationsFromScript(filePath string) (animations, sections []string, err error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, nil, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        line = strings.TrimSpace(line)
        line = strings.ReplaceAll(line, " ", " ")
        if strings.HasPrefix(line, "#") && strings.HasSuffix(line, "#") && !strings.HasPrefix(line, "##") {
            foundIt := false
            for _, value := range []string{"# [[title]] #", "# Intro #", "# Setup #", "# Destroy #"} {
                if line == value {
                    foundIt = true
                    break
                }
            }
            if !foundIt {
                line = strings.ReplaceAll(line, "#", "")
                line = strings.TrimSpace(line)
                line = fmt.Sprintf("Section: %s", line)
                animations = append(animations, line)
                sections = append(sections, line)
            }
        } else if strings.HasPrefix(line, "# TODO:") {
            line = strings.ReplaceAll(line, "# TODO:", "")
            line = strings.TrimSpace(line)
            animations = append(animations, line)
        }
    }
    if err := scanner.Err(); err != nil {
        return nil, nil, err
    }

    return animations, sections, nil
}

func (r *Repo) getAnimationsFromMarkdown(filePath string) (animations, sections []string, err error) {
    // Similar implementation with different parsing rules
    // ...
}
```

### Refactored Implementation
```go
// Refactored implementation
func (r *Repo) GetAnimations(filePath string) (animations, sections []string, err error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, nil, err
    }
    defer file.Close()

    isScript := strings.HasSuffix(filePath, ".sh")
    
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        line = strings.TrimSpace(line)
        line = strings.ReplaceAll(line, " ", " ")
        
        // Handle section headers
        if isScript && strings.HasPrefix(line, "#") && strings.HasSuffix(line, "#") && !strings.HasPrefix(line, "##") {
            // Process script-style section headers
            foundIt := false
            for _, value := range []string{"# [[title]] #", "# Intro #", "# Setup #", "# Destroy #"} {
                if line == value {
                    foundIt = true
                    break
                }
            }
            if !foundIt {
                line = strings.ReplaceAll(line, "#", "")
                line = strings.TrimSpace(line)
                line = fmt.Sprintf("Section: %s", line)
                animations = append(animations, line)
                sections = append(sections, line)
            }
        } else if !isScript && strings.HasPrefix(line, "## ") {
            // Process markdown-style section headers
            containsAny := false
            for _, value := range []string{"## Intro", "## Setup", "## Destroy"} {
                if line == value {
                    containsAny = true
                    break
                }
            }
            if !containsAny {
                line = strings.Replace(line, "## ", "", 1)
                line = strings.TrimSpace(line)
                line = fmt.Sprintf("Section: %s", line)
                animations = append(animations, line)
                sections = append(sections, line)
            }
        }
        
        // Handle TODOs with unified approach
        todoPrefix := isScript ? "# TODO:" : "TODO:"
        if strings.HasPrefix(line, todoPrefix) {
            line = strings.ReplaceAll(line, todoPrefix, "")
            line = strings.TrimSpace(line)
            animations = append(animations, line)
        }
    }
    
    if err := scanner.Err(); err != nil {
        return nil, nil, err
    }

    return animations, sections, nil
}
```

## Implementation Strategy

### Phase 1: Analysis
1. Review all calls to `GetAnimations` and assess impact
2. Document differences between script and markdown parsing
3. Create test cases that cover all supported formats

### Phase 2: Implementation
1. Refactor `GetAnimations` to handle both formats directly
2. Remove the deprecated `getAnimationsFromScript` method
3. Update any code that depends on the specific method
4. Ensure format detection is robust

### Phase 3: Testing
1. Run tests with various file types
2. Verify that output matches expected formats
3. Check for any regressions in animation extraction
4. Validate performance characteristics

### Phase 4: Documentation
1. Update code comments to reflect changes
2. Document the unified approach to animation extraction
3. Update any external documentation

## Success Criteria
- The "TODO: Remove" comment and the corresponding deprecated method are removed from the codebase
- All animation extraction uses a single, unified approach
- Existing script and markdown files continue to work correctly
- Tests pass for all supported file formats
- Code is more maintainable with reduced duplication

## Dependencies
- Current implementation of `getAnimationsFromMarkdown`
- Understanding of the animation format needs for both script and markdown files
- Test coverage for existing functionality
</PRD> 