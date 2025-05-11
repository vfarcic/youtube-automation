<context>
# Overview  
This document outlines the requirements for automating LinkedIn posting functionality in the YouTube Automation tool. Currently, posting to LinkedIn is manually marked as completed with a "TODO: Automate" comment in the code.

# Core Features  
- Automated LinkedIn posting
- LinkedIn API integration
- LinkedIn post configuration
- Status tracking and error handling
</context>
<PRD>
# YouTube Automation Project - LinkedIn Posting Automation Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process, including promotion on LinkedIn. Currently, marking LinkedIn posting as completed requires manual intervention, with a "TODO: Automate" comment in the code. This PRD outlines requirements for automating LinkedIn posting.

## Automation Objectives
- Eliminate manual steps for LinkedIn posting tasks
- Ensure consistent promotion on LinkedIn
- Reduce the time required to publish and promote videos
- Track posting status accurately
- Provide fallback mechanisms for API failures

## Technical Implementation Requirements

### 1. LinkedIn API Integration

#### 1.1 Authentication & Authorization
- Implement API integration with LinkedIn using OAuth 2.0
- Securely store and manage LinkedIn API credentials
- Handle token refresh workflows
- Implement proper error handling for authentication failures

#### 1.2 Post Creation
- Create posts that include video title, description, and YouTube link
- Support video thumbnails in posts
- Handle API rate limiting with exponential backoff
- Implement proper error handling for failed posts

### 2. LinkedIn Post Configuration

#### 2.1 Content Formatting
- Allow customization of post format and content
- Support hashtags and @mentions
- Create templates for different types of video content
- Format code snippets appropriately for technical videos

#### 2.2 Post Management
- Enable scheduling of posts for optimal timing
- Provide preview of post content before publishing
- Support editing of scheduled posts
- Track analytics and engagement through LinkedIn API

### 3. Infrastructure Components

#### 3.1 Authentication Management
- Implement secure credential storage for LinkedIn
- Support OAuth refresh token workflows
- Handle expired tokens gracefully
- Provide clear error messages for authentication failures

#### 3.2 Status Tracking
- Update video metadata with LinkedIn posting status
- Record timestamps of successful posts
- Store post URLs for reference
- Track failed attempts with reason codes

#### 3.3 Error Handling
- Implement graceful failure modes for LinkedIn API interactions
- Provide retry capabilities with backoff
- Log detailed error information for troubleshooting
- Allow manual intervention when automation fails

## Implementation Details

```go
// Example implementation of LinkedIn posting function
func postLinkedIn(video Video) error {
    // Initialize LinkedIn client
    client := linkedin.NewClient(settings.LinkedIn.AccessToken)
    
    // Create post content
    postContent := formatLinkedInPost(video.Title, video.Description, video.VideoId)
    
    // Upload thumbnail if available
    var imageURN string
    if len(video.ThumbnailPath) > 0 {
        imageData, err := os.ReadFile(video.ThumbnailPath)
        if err != nil {
            return fmt.Errorf("failed to read thumbnail: %w", err)
        }
        
        imageURN, err = client.UploadImage(imageData)
        if err != nil {
            // Continue without image rather than failing completely
            log.Printf("Warning: Failed to upload thumbnail to LinkedIn: %s", err)
        }
    }
    
    // Create and publish post
    postURN, err := client.CreatePost(postContent, imageURN)
    if err != nil {
        return fmt.Errorf("failed to create LinkedIn post: %w", err)
    }
    
    // Update video metadata
    video.LinkedInPosted = true
    video.LinkedInPostURL = fmt.Sprintf("https://www.linkedin.com/feed/update/%s", postURN)
    video.LinkedInPostTimestamp = time.Now().Format(time.RFC3339)
    
    // Save updated video metadata
    yaml := YAML{}
    return yaml.WriteVideo(video, video.Path)
}

// Format LinkedIn post with appropriate content
func formatLinkedInPost(title, description, videoId string) string {
    youtubeURL := fmt.Sprintf("https://youtu.be/%s", videoId)
    
    // Create post with hashtags based on content
    hashtags := generateHashtags(title, description)
    
    return fmt.Sprintf(
        "🎬 New Video: %s\n\n%s\n\n%s\n\n%s",
        title,
        getLinkedInSummary(description),
        youtubeURL,
        hashtags,
    )
}
```

## Implementation Strategy

### Phase 1: API Integration
1. Create authentication module for LinkedIn
2. Implement basic API client functionality
3. Test connectivity and validation
4. Establish error handling patterns

### Phase 2: Content Generation
1. Develop templates for LinkedIn posts
2. Implement content generation logic
3. Create preview capabilities
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
- The "TODO: Automate" comment for LinkedIn is removed from the code
- LinkedIn posts can be created without manual intervention
- Posting status is accurately reflected in the video metadata
- Error handling provides clear guidance for failures
- LinkedIn posting process is significantly faster than manual approach

## Dependencies
- LinkedIn API access and developer account
- Authentication credentials for LinkedIn API
- Rate limit considerations for LinkedIn API
- Understanding of LinkedIn content policies
</PRD> 