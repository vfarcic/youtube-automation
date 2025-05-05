<context>
# Overview  
This document outlines the requirements for automating devopstoolkit.live posting functionality in the YouTube Automation tool. Currently, posting to devopstoolkit.live is manually marked as completed with a "TODO: Automate" comment in the code.

# Core Features  
- Automated devopstoolkit.live posting
- Website API integration
- Content formatting for the platform
- Status tracking and error handling
</context>
<PRD>
# YouTube Automation Project - DevOpsToolkit.live Posting Automation Requirements

## Project Overview
YouTube Automation is a CLI tool for managing the YouTube video creation process, including promotion on the devopstoolkit.live website. Currently, marking website posting as completed requires manual intervention, with a "TODO: Automate" comment in the code. This PRD outlines requirements for automating devopstoolkit.live posting.

## Automation Objectives
- Eliminate manual steps for devopstoolkit.live posting tasks
- Ensure consistent content formatting on the website
- Reduce the time required to publish and promote videos
- Track posting status accurately
- Provide fallback mechanisms for API failures

## Technical Implementation Requirements

### 1. Website API Integration

#### 1.1 Authentication & Authorization
- Implement API client for devopstoolkit.live
- Handle authentication securely
- Support session management for the API
- Implement proper error handling for authentication failures

#### 1.2 Post Creation
- Create posts with appropriate metadata
- Support featured image upload from video thumbnail
- Set proper categories and tags based on video content
- Implement retry mechanism for submission failures

### 2. Content Formatting

#### 2.1 Post Structure
- Generate post content from video description
- Format code snippets appropriately
- Create proper hyperlinking to YouTube video
- Maintain consistent styling with existing content

#### 2.2 Metadata Management
- Support category and tag assignment
- Set appropriate featured images
- Create SEO-friendly permalinks
- Include appropriate metadata for social sharing

### 3. Infrastructure Components

#### 3.1 Authentication Management
- Implement secure credential storage for website API
- Support token refresh workflows if applicable
- Handle authentication failures gracefully
- Provide clear error messages for authentication issues

#### 3.2 Status Tracking
- Update video metadata with website posting status
- Record timestamps of successful posts
- Store post URLs for reference
- Track failed attempts with reason codes

#### 3.3 Error Handling
- Implement graceful failure modes for API interactions
- Provide retry capabilities with backoff
- Log detailed error information for troubleshooting
- Allow manual intervention when automation fails

## Implementation Details

```go
// Example implementation of devopstoolkit.live posting function
func postDevOpsToolkit(video Video) error {
    // Initialize devopstoolkit.live client
    client := devopstoolkit.NewClient(
        settings.DevOpsToolkit.URL,
        settings.DevOpsToolkit.Username,
        settings.DevOpsToolkit.Password,
    )
    
    // Authenticate with the API
    if err := client.Authenticate(); err != nil {
        return fmt.Errorf("failed to authenticate with devopstoolkit.live: %w", err)
    }
    
    // Create post content
    content := formatDevOpsToolkitContent(video)
    
    // Prepare post data
    post := devopstoolkit.Post{
        Title:       video.Title,
        Content:     content,
        Excerpt:     getExcerpt(video.Description, 150),
        Status:      "publish",
        Categories:  getCategoriesForVideo(video),
        Tags:        getTagsForVideo(video),
        FeaturedMedia: 0, // Will be set if thumbnail upload succeeds
    }
    
    // Upload thumbnail as featured image if available
    if len(video.ThumbnailPath) > 0 {
        mediaID, err := client.UploadMedia(video.ThumbnailPath, video.Title)
        if err != nil {
            log.Printf("Warning: Failed to upload thumbnail to devopstoolkit.live: %s", err)
            // Continue without featured image rather than failing completely
        } else {
            post.FeaturedMedia = mediaID
        }
    }
    
    // Create the post
    postID, err := client.CreatePost(post)
    if err != nil {
        return fmt.Errorf("failed to create post on devopstoolkit.live: %w", err)
    }
    
    // Get the post URL
    postURL, err := client.GetPostURL(postID)
    if err != nil {
        log.Printf("Warning: Failed to get post URL: %s", err)
        // Create a reasonable fallback URL
        postURL = fmt.Sprintf("%s/?p=%d", settings.DevOpsToolkit.URL, postID)
    }
    
    // Update video metadata
    video.DevOpsToolkitPosted = true
    video.DevOpsToolkitPostURL = postURL
    video.DevOpsToolkitPostTimestamp = time.Now().Format(time.RFC3339)
    video.DevOpsToolkitPostID = postID
    
    // Save updated video metadata
    yaml := YAML{}
    if err := yaml.WriteVideo(video, video.Path); err != nil {
        return fmt.Errorf("failed to update video metadata: %w", err)
    }
    
    return nil
}

// Format content for devopstoolkit.live post
func formatDevOpsToolkitContent(video Video) string {
    youtubeURL := fmt.Sprintf("https://youtu.be/%s", video.VideoId)
    embedCode := fmt.Sprintf(`<iframe width="560" height="315" src="https://www.youtube.com/embed/%s" frameborder="0" allow="accelerometer; autoplay; encrypted-media; gyroscope; picture-in-picture" allowfullscreen></iframe>`, video.VideoId)
    
    // Process the description to create proper formatting
    description := video.Description
    description = formatCodeBlocks(description)
    description = formatMarkdownLinks(description)
    description = addParagraphTags(description)
    
    // Assemble the full post content
    content := strings.Builder{}
    content.WriteString(fmt.Sprintf("<h2>%s</h2>\n\n", video.Title))
    content.WriteString(fmt.Sprintf("%s\n\n", embedCode))
    content.WriteString("<h3>Video Description</h3>\n\n")
    content.WriteString(fmt.Sprintf("%s\n\n", description))
    
    // Add a call to action
    content.WriteString(`<h3>Watch on YouTube</h3>
<p>Click <a href="`)
    content.WriteString(youtubeURL)
    content.WriteString(`" target="_blank" rel="noopener">here</a> to watch the video on YouTube.</p>`)
    
    return content.String()
}

// Helper functions
func formatCodeBlocks(text string) string {
    // Convert markdown code blocks to HTML
    codeBlockRegex := regexp.MustCompile("```([^`]*?)```")
    return codeBlockRegex.ReplaceAllStringFunc(text, func(match string) string {
        // Remove the backticks
        code := strings.TrimPrefix(match, "```")
        code = strings.TrimSuffix(code, "```")
        
        // Wrap in pre and code tags
        return fmt.Sprintf("<pre><code>%s</code></pre>", html.EscapeString(code))
    })
}

func formatMarkdownLinks(text string) string {
    // Convert markdown links to HTML
    linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
    return linkRegex.ReplaceAllString(text, `<a href="$2" target="_blank" rel="noopener">$1</a>`)
}

func addParagraphTags(text string) string {
    // Add paragraph tags to text blocks
    lines := strings.Split(text, "\n\n")
    for i, line := range lines {
        if !strings.HasPrefix(line, "<") && line != "" {
            lines[i] = fmt.Sprintf("<p>%s</p>", line)
        }
    }
    return strings.Join(lines, "\n\n")
}
```

## Implementation Strategy

### Phase 1: API Integration
1. Create authentication module for devopstoolkit.live
2. Implement basic API client functionality
3. Test post creation and media upload
4. Establish error handling patterns

### Phase 2: Content Generation
1. Develop content formatting functions
2. Implement code block handling
3. Create category and tag mapping
4. Test with various video types

### Phase 3: Workflow Integration
1. Replace manual checkbox UI with automation triggers
2. Update status tracking in the YAML files
3. Implement progress indicators during posting
4. Create manual override capabilities

### Phase 4: Testing and Refinement
1. Test with various video types and categories
2. Refine error handling and recovery
3. Optimize media handling
4. Create comprehensive logging

## Success Criteria
- The "TODO: Automate" comment for devopstoolkit.live is removed from the code
- Website posts can be created without manual intervention
- Content is properly formatted with correct media
- Categories and tags are appropriately assigned
- Posting status is accurately reflected in the video metadata
- Error handling provides clear guidance for failures
- Website posting process is significantly faster than manual approach

## Dependencies
- API access to devopstoolkit.live website
- Authentication credentials for the website
- WordPress API compatibility (if applicable)
- Understanding of the website's content structure and taxonomy
</PRD> 