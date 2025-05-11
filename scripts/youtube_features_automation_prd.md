<context>
# Overview  
This document outlines the requirements for automating YouTube-specific features in the YouTube Automation tool. Currently, creating YouTube highlights and posting comments are manually marked as completed with "TODO: Automate" comments in the code.

# Core Features  
- Automated YouTube highlight creation
- Automated YouTube comment posting
- YouTube Data API integration
- Status tracking and error handling
</context>
<PRD>
# YouTube Automation Project - YouTube Features Automation Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process, including additional YouTube-specific features like highlights and comments. Currently, creating highlights and posting comments are manually marked as completed, with "TODO: Automate" comments in the code. This PRD outlines requirements for automating these YouTube features.

## Automation Objectives
- Eliminate manual steps for YouTube highlight creation and comment posting
- Ensure consistent quality and formatting for highlights and comments
- Reduce the time required to enhance videos after publication
- Track features status accurately
- Provide fallback mechanisms for API failures

## Technical Implementation Requirements

### 1. YouTube Highlight Creation

#### 1.1 API Integration
- Use YouTube Data API to create video highlights
- Handle authentication with OAuth 2.0
- Manage API quota limitations
- Implement proper error handling for API failures

#### 1.2 Highlight Management
- Support timestamp selection for highlights
- Generate appropriate titles and descriptions
- Enable custom thumbnail selection
- Create highlight playlists when appropriate

### 2. YouTube Comment Management

#### 2.1 Comment Creation
- Automate pinned comment creation with timestamp links
- Generate comment content from video outline
- Support markdown/formatting allowed by YouTube
- Enable mentions and references as needed

#### 2.2 Comment Engagement
- Implement comment reply monitoring
- Create engagement-focused comment templates
- Support threading for multiple comments
- Track comment engagement metrics

### 3. Infrastructure Components

#### 3.1 Authentication Management
- Implement secure credential storage for YouTube API
- Support OAuth refresh token workflows
- Handle expired tokens gracefully
- Provide clear error messages for authentication failures

#### 3.2 Status Tracking
- Update video metadata with highlight and comment status
- Record timestamps of successful operations
- Store highlight IDs and comment IDs for reference
- Track failed attempts with reason codes

#### 3.3 Error Handling
- Implement graceful failure modes for API interactions
- Provide retry capabilities with backoff
- Log detailed error information for troubleshooting
- Allow manual intervention when automation fails

## Implementation Details

```go
// Example implementation of YouTube highlight creation function
func createYouTubeHighlight(video Video) error {
    // Initialize YouTube client
    client, err := youtube.NewClient(settings.YouTube.OAuthToken)
    if err != nil {
        return fmt.Errorf("failed to create YouTube client: %w", err)
    }
    
    // Prepare highlight details
    startTimeStr := video.HighlightStart // Format: "1:30"
    endTimeStr := video.HighlightEnd     // Format: "2:45"
    
    // Convert time strings to seconds
    startTimeSec, err := timeStringToSeconds(startTimeStr)
    if err != nil {
        return fmt.Errorf("invalid highlight start time format: %w", err)
    }
    
    endTimeSec, err := timeStringToSeconds(endTimeStr)
    if err != nil {
        return fmt.Errorf("invalid highlight end time format: %w", err)
    }
    
    // Create highlight title
    highlightTitle := fmt.Sprintf("%s - Highlight", video.Title)
    if len(video.HighlightTitle) > 0 {
        highlightTitle = video.HighlightTitle
    }
    
    // Create highlight description
    highlightDescription := fmt.Sprintf(
        "Highlight from \"%s\"\n\nOriginal video: https://youtu.be/%s",
        video.Title,
        video.VideoId,
    )
    if len(video.HighlightDescription) > 0 {
        highlightDescription = video.HighlightDescription
    }
    
    // Create the highlight clip
    highlightReq := &youtube.ClipInsertRequest{
        Snippet: &youtube.ClipSnippet{
            Title:       highlightTitle,
            Description: highlightDescription,
            OriginalVideoId: video.VideoId,
            TimeRange: &youtube.ClipTimeRange{
                StartTimeMs: int64(startTimeSec * 1000),
                EndTimeMs:   int64(endTimeSec * 1000),
            },
        },
    }
    
    // Execute the request with quota awareness and retries
    var highlight *youtube.Clip
    err = withRetry(3, func() error {
        var reqErr error
        highlight, reqErr = client.Clips.Insert("snippet", highlightReq).Do()
        return reqErr
    })
    
    if err != nil {
        return fmt.Errorf("failed to create YouTube highlight: %w", err)
    }
    
    // Update video metadata
    video.HighlightCreated = true
    video.HighlightId = highlight.Id
    video.HighlightURL = fmt.Sprintf("https://youtu.be/%s", highlight.Id)
    video.HighlightCreatedTimestamp = time.Now().Format(time.RFC3339)
    
    // Save updated video metadata
    yaml := YAML{}
    if err := yaml.WriteVideo(video, video.Path); err != nil {
        return fmt.Errorf("failed to update video metadata: %w", err)
    }
    
    return nil
}

// Example implementation of YouTube comment posting function
func postYouTubeComment(video Video) error {
    // Initialize YouTube client
    client, err := youtube.NewClient(settings.YouTube.OAuthToken)
    if err != nil {
        return fmt.Errorf("failed to create YouTube client: %w", err)
    }
    
    // Generate comment content
    commentText := generateCommentWithTimestamps(video)
    
    // Create the comment
    commentReq := &youtube.CommentThread{
        Snippet: &youtube.CommentThreadSnippet{
            VideoId: video.VideoId,
            TopLevelComment: &youtube.Comment{
                Snippet: &youtube.CommentSnippet{
                    TextOriginal: commentText,
                },
            },
        },
    }
    
    // Execute the request with quota awareness and retries
    var comment *youtube.CommentThread
    err = withRetry(3, func() error {
        var reqErr error
        comment, reqErr = client.CommentThreads.Insert("snippet", commentReq).Do()
        return reqErr
    })
    
    if err != nil {
        return fmt.Errorf("failed to post YouTube comment: %w", err)
    }
    
    // Pin the comment if requested
    if video.PinComment {
        modReq := &youtube.Comment{
            Id: comment.Snippet.TopLevelComment.Id,
            Snippet: &youtube.CommentSnippet{
                ModerationStatus: "published",
                IsPinned:         true,
            },
        }
        
        err = withRetry(2, func() error {
            _, reqErr := client.Comments.Update("snippet", modReq).Do()
            return reqErr
        })
        
        if err != nil {
            log.Printf("Warning: Failed to pin comment: %s", err)
            // Continue without pinning rather than failing completely
        }
    }
    
    // Update video metadata
    video.CommentPosted = true
    video.CommentId = comment.Id
    video.CommentTimestamp = time.Now().Format(time.RFC3339)
    
    // Save updated video metadata
    yaml := YAML{}
    if err := yaml.WriteVideo(video, video.Path); err != nil {
        return fmt.Errorf("failed to update video metadata: %w", err)
    }
    
    return nil
}

// Helper function to generate comment with timestamps
func generateCommentWithTimestamps(video Video) string {
    // Generate a comment with timestamps
    var comment strings.Builder
    
    // Add intro text
    comment.WriteString("🔍 Video Chapters:\n\n")
    
    // Add timestamps if available
    if len(video.Timestamps) > 0 {
        for _, ts := range video.Timestamps {
            comment.WriteString(fmt.Sprintf("%s - %s\n", ts.Time, ts.Title))
        }
    } else {
        // Extract timestamps from description if not explicitly provided
        timestamps := extractTimestampsFromDescription(video.Description)
        for _, ts := range timestamps {
            comment.WriteString(fmt.Sprintf("%s\n", ts))
        }
    }
    
    // Add a call to action
    comment.WriteString("\nLet me know in the comments if you have any questions!")
    
    return comment.String()
}

// Helper function for retries with exponential backoff
func withRetry(maxRetries int, f func() error) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = f()
        if err == nil {
            return nil
        }
        
        // Check if it's a quota error
        if isQuotaError(err) {
            log.Printf("YouTube API quota exceeded, waiting before retry %d/%d", i+1, maxRetries)
            time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
            continue
        }
        
        // For non-quota errors, only retry server errors
        if isServerError(err) {
            log.Printf("YouTube API server error, waiting before retry %d/%d", i+1, maxRetries)
            time.Sleep(time.Duration(math.Pow(2, float64(i))) * time.Second)
            continue
        }
        
        // Don't retry client errors
        return err
    }
    
    return err
}
```

## Implementation Strategy

### Phase 1: API Integration
1. Create authentication module for YouTube Data API
2. Implement basic API client functionality
3. Test highlight creation and comment posting
4. Establish quota management and error handling patterns

### Phase 2: Content Generation
1. Develop timestamp extraction logic
2. Implement highlight selection algorithms
3. Create comment templates with timestamps
4. Test with various video types

### Phase 3: Workflow Integration
1. Replace manual checkbox UI with automation triggers
2. Update status tracking in the YAML files
3. Implement progress indicators during API operations
4. Create manual override capabilities

### Phase 4: Testing and Refinement
1. Test with various video types and categories
2. Refine error handling and recovery
3. Optimize API usage to respect quota limitations
4. Create comprehensive logging

## Success Criteria
- The "TODO: Automate" comments for YouTube features are removed from the code
- YouTube highlights can be created without manual intervention
- Comments with timestamps can be posted and pinned automatically
- Feature status is accurately reflected in the video metadata
- Error handling provides clear guidance for failures
- API quota usage is optimized and monitored
- YouTube feature automation is significantly faster than manual approach

## Dependencies
- YouTube Data API access
- OAuth 2.0 authentication credentials
- Understanding of API quota limitations
- Proper API scopes for highlight creation and comment management
</PRD> 