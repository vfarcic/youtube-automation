<context>
# Overview  
This document outlines the requirements for implementing a video file management feature in the YouTube Automation tool. Currently, the ability to move video files to different directories is marked as a TODO in the codebase.

# Core Features  
- Move video files between directories
- Maintain proper metadata synchronization
- Update references in the YAML file
- Provide user-friendly interface for directory selection
</context>
<PRD>
# YouTube Automation Project - Video File Management Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process. Currently, the tool lacks the capability to move video files between directories, which has been marked as a TODO in the codebase. This feature would enhance the file organization capabilities and improve workflow management.

## Feature Objectives
- Enable moving video files to different directories
- Maintain integrity of metadata and references
- Provide a simple, intuitive interface for directory selection
- Ensure proper error handling during file operations
- Update all references in configuration files

## Technical Implementation Requirements

### 1. File Movement Core Functionality

#### 1.1 File System Operations
- Implement safe file moving operations that handle potential errors
- Support moving both video files and associated metadata
- Maintain file permissions during moves
- Handle filename collisions with appropriate strategies (rename, prompt, abort)
- Verify file integrity after move operations

#### 1.2 Directory Structure Management
- Maintain organized directory structures for video files
- Create target directories if they don't exist (with confirmation)
- Support predefined directory templates for common organization patterns
- Preserve relative directory relationships when possible

### 2. Metadata Synchronization

#### 2.1 YAML File Updates
- Automatically update file paths in YAML metadata
- Adjust all references to the moved files
- Maintain history of file movements in metadata
- Update any directory-specific configuration

#### 2.2 Index Updates
- Update index entries to reflect new file locations
- Maintain proper sorting and categorization after moves
- Adjust category assignments if moving between category directories
- Verify index integrity after updates

### 3. User Interface

#### 3.1 Directory Selection
- Add "Move Files" option to the action menu in `getActionOptions()`
- Provide directory browser for selecting target location
- Support both predefined directories and custom paths
- Include recently used directories for quick selection

#### 3.2 Batch Operations
- Allow moving multiple video files at once
- Support filtering and selection patterns
- Provide confirmation for batch operations
- Display progress during batch moves

### 4. Validation and Error Handling

#### 4.1 Pre-move Validation
- Verify write permissions on target directory
- Check available disk space before moving
- Validate that all references can be updated
- Identify potential conflicts before executing move

#### 4.2 Error Recovery
- Implement rollback capability for failed moves
- Keep backup of metadata before updating
- Log all operations for audit purposes
- Provide clear error messages with suggested resolutions

## Implementation Details

### File Movement Implementation
```go
// Add to choices.go in getActionOptions()
func (c *Choices) getActionOptions() []huh.Option[int] {
    return []huh.Option[int]{
        huh.NewOption("Edit", actionEdit),
        huh.NewOption("Delete", actionDelete),
        huh.NewOption("Move Files", actionMoveFiles), // New option
        huh.NewOption("Return", actionReturn),
    }
}

// New function to handle moving files
func (c *Choices) MoveVideoFiles(video Video) error {
    // Get target directory
    targetDir, err := c.selectTargetDirectory()
    if err != nil {
        return err
    }
    
    // Move video file
    oldPath := c.GetFilePath(video.Category, video.Name, "md")
    newPath := filepath.Join(targetDir, fmt.Sprintf("%s.md", video.Name))
    
    if err := moveFile(oldPath, newPath); err != nil {
        return fmt.Errorf("failed to move markdown file: %w", err)
    }
    
    // Move YAML file
    oldYamlPath := video.Path
    newYamlPath := filepath.Join(targetDir, fmt.Sprintf("%s.yaml", video.Name))
    
    if err := moveFile(oldYamlPath, newYamlPath); err != nil {
        // Try to roll back the first move
        moveFile(newPath, oldPath)
        return fmt.Errorf("failed to move YAML file: %w", err)
    }
    
    // Update YAML file content with new paths
    yaml := YAML{}
    video.Path = newYamlPath
    video.Category = filepath.Base(targetDir)
    
    if err := yaml.WriteVideo(video, newYamlPath); err != nil {
        // Try rollback
        moveFile(newPath, oldPath)
        moveFile(newYamlPath, oldYamlPath)
        return fmt.Errorf("failed to update YAML content: %w", err)
    }
    
    // Update index
    if err := c.updateVideoIndex(video, oldPath, newYamlPath); err != nil {
        return fmt.Errorf("failed to update index: %w", err)
    }
    
    return nil
}

// Helper function to move a file
func moveFile(src, dst string) error {
    // Create target directory if it doesn't exist
    if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
        return err
    }
    
    // Check if destination file already exists
    if _, err := os.Stat(dst); err == nil {
        return fmt.Errorf("destination file already exists: %s", dst)
    }
    
    // Move the file
    return os.Rename(src, dst)
}
```

### Directory Selection UI
```go
func (c *Choices) selectTargetDirectory() (string, error) {
    // Get available directories
    dirs, err := c.getAvailableDirectories()
    if err != nil {
        return "", err
    }
    
    // Add option for custom directory
    options := huh.NewOptions[string]()
    for _, dir := range dirs {
        options = append(options, huh.NewOption(dir.Name, dir.Path))
    }
    options = append(options, huh.NewOption("Custom directory...", "custom"))
    
    // Create selection form
    var selectedDir string
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Select target directory").
                Options(options...).
                Value(&selectedDir),
        ),
    )
    
    if err := form.Run(); err != nil {
        return "", err
    }
    
    // Handle custom directory option
    if selectedDir == "custom" {
        var customPath string
        customForm := huh.NewForm(
            huh.NewGroup(
                huh.NewInput().
                    Title("Enter custom directory path").
                    Value(&customPath),
            ),
        )
        
        if err := customForm.Run(); err != nil {
            return "", err
        }
        
        return customPath, nil
    }
    
    return selectedDir, nil
}
```

## Implementation Strategy

### Phase 1: Core Functionality
1. Add the "Move Files" option to the action menu
2. Implement basic file movement functionality
3. Update YAML file references
4. Add directory selection interface

### Phase 2: Metadata Synchronization
1. Implement index updating
2. Add validation for moves
3. Create rollback mechanisms
4. Test with various file types and structures

### Phase 3: Enhanced Features
1. Add batch movement capabilities
2. Implement directory templates
3. Create move history tracking
4. Refine user interface

## Success Criteria
- The "TODO: Add the option to move video files to a different directory" comment is removed from the code
- Users can successfully move video files between directories
- All metadata and references are properly updated after moves
- The feature is accessible through the existing UI
- Error handling properly manages edge cases

## Dependencies
- Access to the file system with appropriate permissions
- Understanding of the current directory structure and naming conventions
- Compatibility with existing metadata formats
</PRD> 