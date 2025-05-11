<context>
# Overview  
This document outlines the requirements for automating GDE posting functionality in the YouTube Automation tool. Currently, posting to the GDE platform is manually marked as completed with a "TODO: Automate" comment in the code.

# Core Features  
- Automated GDE platform posting
- GDE API integration
- GDE content strategy
- Status tracking and error handling
</context>
<PRD>
# YouTube Automation Project - GDE Posting Automation Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process, including promotion on the Google Developer Experts (GDE) platform. Currently, marking GDE posting as completed requires manual intervention, with a "TODO: Automate" comment in the code. This PRD outlines requirements for automating GDE platform posting.

## Automation Objectives
- Eliminate manual steps for GDE posting tasks
- Ensure appropriate content for the GDE platform
- Reduce the time required to publish and promote videos
- Track posting status accurately
- Provide fallback mechanisms for API failures

## Technical Implementation Requirements

### 1. GDE Platform Integration

#### 1.1 Authentication & Authorization
- Implement integration with https://gde.advocu.com
- Handle OAuth authentication flow securely
- Store credentials with proper encryption
- Implement proper error handling for authentication failures

#### 1.2 Post Creation
- Support post creation with proper formatting
- Handle metadata requirements specific to GDE platform
- Track submission status through the API
- Implement retry mechanism for submission failures

### 2. GDE Content Strategy

#### 2.1 Content Formatting
- Generate appropriate content for the GDE platform
- Support GDE-specific content guidelines
- Include proper attribution and links
- Format technical content according to platform standards

#### 2.2 Platform Engagement
- Track post analytics when available
- Support community guidelines compliance
- Implement appropriate tagging and categorization
- Follow GDE platform best practices

### 3. Infrastructure Components

#### 3.1 Authentication Management
- Implement secure credential storage for GDE platform
- Support token refresh workflows
- Handle authentication failures gracefully
- Provide clear error messages for authentication issues

#### 3.2 Status Tracking
- Update video metadata with GDE posting status
- Record timestamps of successful posts
- Store post URLs and IDs for reference
- Track failed attempts with reason codes

#### 3.3 Error Handling
- Implement graceful failure modes for API interactions
- Provide retry capabilities with backoff
- Log detailed error information for troubleshooting
- Allow manual intervention when automation fails

## Implementation Details

```go
// Example implementation of GDE posting function
func postGDE(video Video) error {
    // Initialize GDE client
    client := gde.NewClient(
        settings.GDE.APIEndpoint,
        settings.GDE.ClientID,
        settings.GDE.ClientSecret,
    )
    
    // Authenticate with the API
    token, err := client.Authenticate()
    if err != nil {
        return fmt.Errorf("failed to authenticate with GDE platform: %w", err)
    }
    
    // Set the token for subsequent requests
    client.SetToken(token)
    
    // Prepare post content
    content := formatGDEContent(video)
    
    // Get categories for the post
    categories := getGDECategories(video.Category, video.Tags)
    
    // Create post request
    postReq := &gde.Post{
        Title:       video.Title,
        Content:     content,
        Categories:  categories,
        VideoURL:    fmt.Sprintf("https://youtu.be/%s", video.VideoId),
        ThumbnailURL: video.ThumbnailURL,
        Published:   true,
    }
    
    // Submit the post with retries
    var post *gde.PostResponse
    err = withRetry(3, func() error {
        var reqErr error
        post, reqErr = client.CreatePost(postReq)
        return reqErr
    })
    
    if err != nil {
        return fmt.Errorf("failed to create GDE post: %w", err)
    }
    
    // Update video metadata
    video.GDEPosted = true
    video.GDEPostURL = post.URL
    video.GDEPostID = post.ID
    video.GDEPostTimestamp = time.Now().Format(time.RFC3339)
    
    // Save updated video metadata
    yaml := YAML{}
    if err := yaml.WriteVideo(video, video.Path); err != nil {
        return fmt.Errorf("failed to update video metadata: %w", err)
    }
    
    log.Printf("Successfully created GDE post: %s", post.URL)
    
    return nil
}

// Format content for GDE platform
func formatGDEContent(video Video) string {
    var content strings.Builder
    
    // Add introduction
    content.WriteString(fmt.Sprintf("## %s\n\n", video.Title))
    
    // Add description
    description := video.Description
    // Remove any YouTube-specific content
    description = cleanYouTubeSpecificContent(description)
    content.WriteString(fmt.Sprintf("%s\n\n", description))
    
    // Add technical details section if exists
    if len(video.TechnicalDetails) > 0 {
        content.WriteString("## Technical Details\n\n")
        content.WriteString(fmt.Sprintf("%s\n\n", video.TechnicalDetails))
    }
    
    // Add resources section
    content.WriteString("## Resources\n\n")
    content.WriteString(fmt.Sprintf("- [YouTube Video](https://youtu.be/%s)\n", video.VideoId))
    
    // Add other links if available
    if len(video.ProjectURL) > 0 {
        content.WriteString(fmt.Sprintf("- [Project Repository](%s)\n", video.ProjectURL))
    }
    
    if len(video.SlideURL) > 0 {
        content.WriteString(fmt.Sprintf("- [Presentation Slides](%s)\n", video.SlideURL))
    }
    
    // Add GitHub URL if it exists
    if len(video.GitHubURL) > 0 {
        content.WriteString(fmt.Sprintf("- [GitHub Repository](%s)\n", video.GitHubURL))
    }
    
    // Add contribution note
    content.WriteString("\n## About the Author\n\n")
    content.WriteString(fmt.Sprintf("This content was created as part of my contribution to the Google Developer Experts program, focusing on %s.", getGDEExpertiseArea(video.Category)))
    
    return content.String()
}

// Get GDE categories based on video category and tags
func getGDECategories(category string, tags []string) []string {
    categories := []string{}
    
    // Map video category to GDE categories
    categoryMap := map[string]string{
        "kubernetes":   "Cloud",
        "docker":       "Cloud",
        "cloud":        "Cloud",
        "golang":       "Web Technologies",
        "web":          "Web Technologies",
        "javascript":   "Web Technologies",
        "devops":       "DevOps",
        "ci/cd":        "DevOps",
        "gitops":       "DevOps",
    }
    
    // Add mapped category if exists
    for key, gdeCategory := range categoryMap {
        if strings.Contains(strings.ToLower(category), key) {
            categories = append(categories, gdeCategory)
            break
        }
    }
    
    // Also check tags
    for _, tag := range tags {
        for key, gdeCategory := range categoryMap {
            if strings.Contains(strings.ToLower(tag), key) && !contains(categories, gdeCategory) {
                categories = append(categories, gdeCategory)
                break
            }
        }
    }
    
    // If no categories were matched, use a default
    if len(categories) == 0 {
        categories = append(categories, "DevOps")
    }
    
    return categories
}

// Helper function to check if slice contains a string
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

// Get expertise area based on category
func getGDEExpertiseArea(category string) string {
    category = strings.ToLower(category)
    
    if strings.Contains(category, "kubernetes") || strings.Contains(category, "docker") || strings.Contains(category, "cloud") {
        return "Cloud technologies and DevOps"
    } else if strings.Contains(category, "golang") || strings.Contains(category, "web") {
        return "Web Technologies"
    } else {
        return "DevOps and software engineering"
    }
}

// Implement retry mechanism with exponential backoff
func withRetry(maxRetries int, f func() error) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = f()
        if err == nil {
            return nil
        }
        
        // Sleep with exponential backoff before retrying
        waitTime := time.Duration(math.Pow(2, float64(i))) * time.Second
        log.Printf("Request failed, waiting %v before retry %d/%d: %v", waitTime, i+1, maxRetries, err)
        time.Sleep(waitTime)
    }
    
    return fmt.Errorf("all %d attempts failed: %w", maxRetries, err)
}
```

## Implementation Strategy

### Phase 1: API Integration
1. Create authentication module for GDE platform
2. Implement basic API client functionality
3. Test post creation and submission
4. Establish error handling patterns

### Phase 2: Content Generation
1. Develop content formatting functions
2. Implement category mapping logic
3. Create templates for GDE posts
4. Test with various video types

### Phase 3: Workflow Integration
1. Replace manual checkbox UI with automation triggers
2. Update status tracking in the YAML files
3. Implement progress indicators during posting
4. Create manual override capabilities

### Phase 4: Testing and Refinement
1. Test with various video types and categories
2. Refine error handling and recovery
3. Optimize submission process
4. Create comprehensive logging

## Success Criteria
- The "TODO: Automate" comment for GDE is removed from the code
- GDE posts can be created without manual intervention
- Content is properly formatted according to platform guidelines
- Categories are appropriately assigned based on video content
- Posting status is accurately reflected in the video metadata
- Error handling provides clear guidance for failures
- GDE posting process is significantly faster than manual approach

## Dependencies
- GDE platform API access (https://gde.advocu.com)
- Authentication credentials for the platform
- Understanding of GDE content guidelines and community standards
- Google Developer Experts program participation
</PRD> 