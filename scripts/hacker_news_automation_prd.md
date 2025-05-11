<context>
# Overview  
This document outlines the requirements for automating Hacker News posting functionality in the YouTube Automation tool. Currently, posting to Hacker News is manually marked as completed with a "TODO: Automate" comment in the code.

# Core Features  
- Automated Hacker News posting
- Hacker News API integration
- Content strategy for Hacker News
- Status tracking and error handling
</context>
<PRD>
# YouTube Automation Project - Hacker News Posting Automation Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process, including promotion on Hacker News. Currently, marking Hacker News posting as completed requires manual intervention, with a "TODO: Automate" comment in the code. This PRD outlines requirements for automating Hacker News posting.

## Automation Objectives
- Eliminate manual steps for Hacker News posting tasks
- Ensure appropriate content for the Hacker News community
- Reduce the time required to publish and promote videos
- Track posting status accurately
- Provide fallback mechanisms for API failures

## Technical Implementation Requirements

### 1. Hacker News API Integration

#### 1.1 Authentication & Authorization
- Use Hacker News API to create new posts
- Handle authentication securely
- Store credentials with proper encryption
- Implement proper error handling for authentication failures

#### 1.2 Post Creation
- Create submissions with appropriate content and formatting
- Support both "Show HN" and regular submissions based on content
- Implement retry mechanism for submission failures
- Track post IDs for future reference

### 2. Hacker News Content Strategy

#### 2.1 Title Optimization
- Optimize post titles for Hacker News audience
- Follow Hacker News posting guidelines and community norms
- Generate titles that avoid clickbait but encourage interest
- Support customization per video category

#### 2.2 Content Formatting
- Create compelling descriptions that follow community guidelines
- Implement timing strategy to maximize visibility
- Monitor post performance through the API
- Adapt content based on historical performance

### 3. Infrastructure Components

#### 3.1 Authentication Management
- Implement secure credential storage for Hacker News
- Support token refresh workflows if applicable
- Handle authentication failures gracefully
- Provide clear error messages for authentication issues

#### 3.2 Status Tracking
- Update video metadata with Hacker News posting status
- Record timestamps of successful posts
- Store post URLs and IDs for reference
- Track failed attempts with reason codes

#### 3.3 Error Handling
- Implement graceful failure modes for Hacker News API interactions
- Provide retry capabilities with backoff
- Log detailed error information for troubleshooting
- Allow manual intervention when automation fails

## Implementation Details

```go
// Example implementation of Hacker News posting function
func postHackerNews(video Video) error {
    // Initialize Hacker News client
    client := hackernews.NewClient(settings.HackerNews.Username, settings.HackerNews.Password)
    
    // Generate appropriate title based on video content
    title := generateHNTitle(video.Title, video.Category)
    
    // Create YouTube URL
    youtubeURL := fmt.Sprintf("https://youtu.be/%s", video.VideoId)
    
    // Determine if this should be a "Show HN" post
    isShowHN := shouldUseShowHN(video.Category, video.Tags)
    if isShowHN {
        title = fmt.Sprintf("Show HN: %s", title)
    }
    
    // Submit the post
    postID, err := client.SubmitStory(title, youtubeURL, "")
    if err != nil {
        // Handle rate limiting with exponential backoff
        if isRateLimitError(err) {
            for retries := 0; retries < 3; retries++ {
                waitTime := time.Duration(math.Pow(2, float64(retries))) * time.Second
                log.Printf("Rate limited by Hacker News, waiting %v before retrying", waitTime)
                time.Sleep(waitTime)
                
                postID, err = client.SubmitStory(title, youtubeURL, "")
                if err == nil {
                    break
                }
            }
            
            if err != nil {
                return fmt.Errorf("failed to submit to Hacker News after retries: %w", err)
            }
        } else {
            return fmt.Errorf("failed to submit to Hacker News: %w", err)
        }
    }
    
    // Update video metadata
    video.HackerNewsPosted = true
    video.HackerNewsPostURL = fmt.Sprintf("https://news.ycombinator.com/item?id=%s", postID)
    video.HackerNewsPostTimestamp = time.Now().Format(time.RFC3339)
    video.HackerNewsID = postID
    
    // Save updated video metadata
    yaml := YAML{}
    if err := yaml.WriteVideo(video, video.Path); err != nil {
        return fmt.Errorf("failed to update video metadata: %w", err)
    }
    
    // Add a first comment to provide more context if needed
    if len(video.HackerNewsComment) > 0 {
        _, err := client.Comment(postID, video.HackerNewsComment)
        if err != nil {
            log.Printf("Warning: Failed to post initial comment to Hacker News: %s", err)
            // Non-critical error, don't fail the whole process
        }
    }
    
    return nil
}

// Generate an appropriate title for Hacker News audience
func generateHNTitle(title string, category string) string {
    // Remove any YouTube-specific formatting
    cleanTitle := strings.ReplaceAll(title, " | YouTube Tutorial", "")
    cleanTitle = strings.ReplaceAll(cleanTitle, " - YouTube", "")
    
    // For tutorials, use a more direct approach
    if strings.Contains(strings.ToLower(category), "tutorial") {
        return fmt.Sprintf("Tutorial: %s", cleanTitle)
    }
    
    // For technical demos, highlight the tech
    if strings.Contains(strings.ToLower(category), "demo") {
        return fmt.Sprintf("Demo: %s", cleanTitle)
    }
    
    return cleanTitle
}

// Determine whether to use "Show HN" prefix
func shouldUseShowHN(category string, tags []string) bool {
    // "Show HN" is for things you've created
    showHNCategories := []string{"project", "demo", "launch", "release"}
    
    for _, c := range showHNCategories {
        if strings.Contains(strings.ToLower(category), c) {
            return true
        }
    }
    
    showHNTags := []string{"release", "project", "demo", "launch"}
    for _, tag := range tags {
        for _, showTag := range showHNTags {
            if strings.Contains(strings.ToLower(tag), showTag) {
                return true
            }
        }
    }
    
    return false
}
```

## Implementation Strategy

### Phase 1: API Integration
1. Create authentication module for Hacker News
2. Implement basic API client functionality
3. Test submission and comment capability
4. Establish error handling patterns

### Phase 2: Content Generation
1. Develop title optimization logic
2. Implement "Show HN" detection
3. Create comment templates
4. Test with various video types

### Phase 3: Workflow Integration
1. Replace manual checkbox UI with automation triggers
2. Update status tracking in the YAML files
3. Implement progress indicators during posting
4. Create manual override capabilities

### Phase 4: Testing and Refinement
1. Test with various video types and categories
2. Refine error handling and recovery
3. Optimize posting timing
4. Create comprehensive logging

## Success Criteria
- The "TODO: Automate" comment for Hacker News is removed from the code
- Hacker News submissions can be created without manual intervention
- Posts follow community guidelines and best practices
- Posting status is accurately reflected in the video metadata
- Error handling provides clear guidance for failures
- Hacker News posting process is significantly faster than manual approach

## Dependencies
- Hacker News API or appropriate libraries
- Authentication credentials for Hacker News
- Understanding of Hacker News community guidelines
- Rate limit considerations for the Hacker News API
</PRD> 