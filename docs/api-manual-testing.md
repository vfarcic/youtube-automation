# YouTube Automation REST API - Manual Testing Guide

This guide provides step-by-step manual testing scenarios for the YouTube Automation REST API using `curl` commands.

## Breaking Changes - String-Based IDs

**Important:** As of the latest version, all video responses now include a string-based `id` field in the format `category/name` (e.g., `"tutorials/my-video"`). This replaces the previous numeric ID system and provides better consistency across the API.

### Migration Guide for Frontend Applications

If you're updating a frontend application that previously used numeric IDs, update your interfaces:

```typescript
// Before
interface Video {
  id: number;
  name: string;
  category: string;
  // ... other fields
}

// After  
interface Video {
  id: string;        // Now string-based: "category/name"
  name: string;
  category: string;
  // ... other fields
}
```

### Key Changes:
- **List endpoint** (`/api/videos/list`): Now includes `id` field with string values and `name` field for easy access
- **Individual video endpoint** (`/api/videos/{videoName}`): Now includes `id` field in response
- **ID format**: `category/name` (e.g., `"tutorials/kubernetes-guide"`)
- **Name field**: Provides direct access to filename without parsing the ID
- **URL paths**: Still use filename-based identifiers (unchanged)

## Prerequisites

1. Start the API server:
   ```bash
   ./youtube-automation --api-enabled
   ```

2. Ensure the server is running on the configured port (default: 8080)

3. Verify the server is healthy:
   ```bash
   curl http://localhost:8080/health
   ```
   Expected response:
   ```json
   {
     "status": "ok",
     "time": "2023-XX-XXTXX:XX:XXZ"
   }
   ```

## Test Scenarios

### 1. Categories Management

#### Get Available Categories
```bash
curl -X GET http://localhost:8080/api/categories
```
Expected response:
```json
{
  "categories": [
    {
      "name": "Category Name",
      "path": "manuscript/category-name"
    }
  ]
}
```

### 2. Video Creation

#### Create a New Video
```bash
curl -X POST http://localhost:8080/api/videos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-api-video",
    "category": "test-category"
  }'
```
Expected response (201 Created):
```json
{
  "video": {
    "name": "test-api-video",
    "category": "test-category"
  }
}
```

#### Test Error Cases for Video Creation
```bash
# Missing name
curl -X POST http://localhost:8080/api/videos \
  -H "Content-Type: application/json" \
  -d '{
    "name": "",
    "category": "test-category"
  }'
```
Expected response (400 Bad Request):
```json
{
  "error": "name and category are required"
}
```

### 3. Video Phase Management

#### Get Video Phases Summary
```bash
curl -X GET http://localhost:8080/api/videos/phases
```
Expected response:
```json
{
  "phases": [
    {
      "id": 7,
      "name": "Ideas",
      "count": 1
    }
  ]
}
```

#### List Videos in a Specific Phase
```bash
# List videos in Ideas phase (id: 7)
curl -X GET "http://localhost:8080/api/videos?phase=7"
```
Expected response:
```json
{
  "videos": [
    {
      "name": "test-api-video",
      "category": "test-category",
      "path": "manuscript/test-category/test-api-video.yaml",
      ...
    }
  ]
}
```

#### **NEW**: Optimized Video List for Frontend Grids
```bash
# Get lightweight video list optimized for frontend display (87% smaller payload)
curl -X GET "http://localhost:8080/api/videos/list?phase=7"
```
Expected response:
```json
{
  "videos": [
    {
      "id": "development/rest-api-testing",
      "name": "rest-api-testing",
      "title": "Complete Guide to REST API Testing",
      "date": "2025-01-06T16:00",
      "thumbnail": "material/api-testing/thumbnail.jpg",
      "category": "development",
      "status": "published",
      "phase": 0,
      "progress": {
        "completed": 10,
        "total": 10
      }
    },
    {
      "id": "devops/kubernetes-deployments",
      "name": "kubernetes-deployments",
      "title": "Advanced Kubernetes Deployments",
      "date": "2025-01-08T14:30",
      "thumbnail": "material/kubernetes/deployment-thumb.jpg",
      "category": "devops",
      "status": "draft",
      "phase": 4,
      "progress": {
        "completed": 7,
        "total": 12
      }
    }
  ]
}
```

**Performance Benefits:**
- **87% smaller payload** (~3.6KB vs ~28KB for full video objects)
- **~200 bytes per video** vs ~8.8KB in the full endpoint
- **Optimized for video grid components** - contains only essential display fields
- **Sub-millisecond response times** for large video lists

**Status Values:**
- `"published"` - Video is fully completed and published
- `"draft"` - Video is in progress or not yet published

**Phase Values (0-7):**
- `0` - Published: Video is completed and live on YouTube
- `1` - Publish Pending: Video ready for upload and publishing
- `2` - Edit Requested: Video needs editing or revisions
- `3` - Material Done: All materials completed, ready for post-production
- `4` - Started: Video creation has begun
- `5` - Delayed: Video is postponed
- `6` - Sponsored Blocked: Video blocked by sponsorship requirements
- `7` - Ideas: Video is still in planning/idea phase

**Phase Calculation Logic:**
The phase value is automatically calculated based on video state:
- Delayed videos → Phase 5
- Videos with sponsorship blocks → Phase 6  
- Published videos → Phase 0
- Videos with upload + tweet → Phase 1
- Videos requiring edits → Phase 2
- Videos with all materials (code, screen, head, diagrams) → Phase 3
- Videos with any materials started → Phase 4
- All other videos → Phase 7 (Ideas)

**Error Cases:**
```bash
# Missing phase parameter
curl -X GET "http://localhost:8080/api/videos/list"
# Response: 400 Bad Request
{
  "error": "phase parameter is required"
}

# Invalid phase parameter
curl -X GET "http://localhost:8080/api/videos/list?phase=invalid"
# Response: 400 Bad Request
{
  "error": "Invalid phase parameter"
}
```

### 4. Individual Video Operations

#### Get Specific Video Details
```bash
curl -X GET "http://localhost:8080/api/videos/my-video-filename?category=test-category"
```

**Note on Video IDs and Names:** 
- The `{videoName}` in the URL path (e.g., `my-video-filename`) is the video's filename-based identifier and must match its YAML filename (e.g., `my-video-filename.yaml`). This is used for all API lookups.
- The `id` field in API responses is a string-based identifier in the format `category/name` (e.g., `"test-category/my-video-filename"`). This provides a unique, human-readable identifier for frontend applications.
- The `name` field in API responses provides direct access to the filename portion (e.g., `"my-video-filename"`) without requiring ID parsing.
- The `title` field contains the video's display name (e.g., "My Video Display Name"), which is read from the file and is independent of both the filename and the string ID.

Expected response:
```json
{
  "video": {
    "id": "test-category/my-video-filename",
    "name": "My Video Display Name",
    "category": "test-category",
    "path": "manuscript/test-category/my-video-filename.yaml",
    "init": {"completed": 0, "total": 8},
    "work": {"completed": 0, "total": 11},
    "define": {"completed": 0, "total": 9},
    "edit": {"completed": 0, "total": 6},
    "publish": {"completed": 0, "total": 2},
    "postPublish": {"completed": 0, "total": 9}
  }
}
```

#### Update Video - Initial Details Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/initial-details?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "projectName": "Test API Project",
    "projectURL": "https://github.com/example/test-project",
    "date": "2023-12-25T10:00",
    "gist": "manuscript/test-category/test-api-video.md"
  }'
```
Expected response:
```json
{
  "video": {
    "name": "test-api-video",
    "projectName": "Test API Project",
    "projectURL": "https://github.com/example/test-project",
    "date": "2023-12-25T10:00",
    "gist": "manuscript/test-category/test-api-video.md",
    "init": {"completed": 4, "total": 8}
  }
}
```

#### Update Video - Work Progress Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/work-progress?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "code": true,
    "head": true,
    "screen": true,
    "thumbnails": true,
    "diagrams": true,
    "screenshots": true,
    "location": "https://drive.google.com/folder/example",
    "tagline": "Learn API testing with this comprehensive guide",
    "relatedVideos": "Previous API tutorial series\nREST fundamentals video",
    "taglineIdeas": "Master API testing in 30 minutes\nComplete REST API testing guide\nAPI testing made simple",
    "otherLogos": "Postman, Insomnia"
  }'
```

#### Update Video - Definition Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/definition?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete Guide to REST API Testing",
    "description": "In this comprehensive video, we explore the fundamentals of REST API testing...",
    "highlight": "Master API testing in 30 minutes",
    "tags": "api, testing, rest, development, tutorial",
    "descriptionTags": "#API #Testing #REST #Development",
    "tweet": "Just released a comprehensive guide to REST API testing! Perfect for developers looking to improve their testing skills. 🚀 #API #Testing",
    "animations": "API request flow diagram\nHTTP status code visualization"
  }'
```

#### Update Video - Post-Production Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/post-production?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "thumbnail": "/path/to/thumbnail.jpg",
    "members": "John Doe, Jane Smith",
    "requestEdit": true,
    "timecodes": "00:00 - Introduction\n05:00 - Setup\n15:00 - Testing\n25:00 - Conclusion",
    "movie": true,
    "slides": true
  }'
```

#### Update Video - Publishing Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/publishing?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "uploadVideo": "/path/to/final-video.mp4",
    "videoId": "dQw4w9WgXcQ",
    "hugoPath": "/content/posts/api-testing-guide.md"
  }'
```

#### Update Video - Post-Publish Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/post-publish?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "dotPosted": true,
    "blueSkyPosted": true,
    "linkedInPosted": true,
    "slackPosted": true,
    "youTubeHighlight": true,
    "youTubeComment": true,
    "youTubeCommentReply": true,
    "gde": false,
    "repo": "https://github.com/example/api-testing-guide",
    "notifiedSponsors": false
  }'
```

### 5. Video Management Operations

#### Move Video to Different Category
```bash
curl -X POST "http://localhost:8080/api/videos/test-api-video/move?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "target_directory_path": "manuscript/tutorials"
  }'
```
Expected response:
```json
{
  "message": "Video moved successfully"
}
```

#### Delete Video
```bash
curl -X DELETE "http://localhost:8080/api/videos/test-api-video?category=tutorials"
```
Expected response: 204 No Content (empty body)

### 6. AI Content Generation

#### Generate Video Titles
```bash
curl -X POST http://localhost:8080/api/ai/titles \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "This tutorial covers advanced Kubernetes deployment strategies including blue-green deployments, canary releases, and rolling updates. We will explore how to implement these patterns using kubectl and Helm charts."
  }'
```
Expected response:
```json
{
  "titles": [
    "Advanced Kubernetes Deployment Strategies: Blue-Green, Canary & Rolling Updates",
    "Master Kubernetes Deployments: Complete Guide to Blue-Green and Canary Releases",
    "Kubernetes Deployment Patterns: From Rolling Updates to Advanced Strategies"
  ]
}
```

#### Generate Video Description
```bash
curl -X POST http://localhost:8080/api/ai/description \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "This tutorial covers advanced Kubernetes deployment strategies including blue-green deployments, canary releases, and rolling updates. We will explore how to implement these patterns using kubectl and Helm charts."
  }'
```
Expected response:
```json
{
  "description": "In this comprehensive tutorial, we dive deep into advanced Kubernetes deployment strategies that every DevOps engineer should master. Learn how to implement blue-green deployments for zero-downtime releases, set up canary deployments for safe feature rollouts, and optimize rolling updates for your applications.\n\nWe'll cover practical implementation using kubectl and Helm charts, with real-world examples and best practices. Perfect for intermediate to advanced Kubernetes users looking to enhance their deployment workflows.\n\nTopics covered:\n- Blue-green deployment implementation\n- Canary release strategies\n- Rolling update optimization\n- kubectl and Helm integration\n- Production deployment best practices"
  }
}
```

#### Generate Video Tags
```bash
curl -X POST http://localhost:8080/api/ai/tags \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "This tutorial covers advanced Kubernetes deployment strategies including blue-green deployments, canary releases, and rolling updates. We will explore how to implement these patterns using kubectl and Helm charts."
  }'
```
Expected response:
```json
{
  "tags": [
    "kubernetes",
    "deployment",
    "devops",
    "blue-green",
    "canary",
    "rolling-updates",
    "kubectl",
    "helm",
    "container-orchestration",
    "cloud-native"
  ]
}
```

#### Generate Social Media Tweets
```bash
curl -X POST http://localhost:8080/api/ai/tweets \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "This tutorial covers advanced Kubernetes deployment strategies including blue-green deployments, canary releases, and rolling updates. We will explore how to implement these patterns using kubectl and Helm charts."
  }'
```
Expected response:
```json
{
  "tweets": [
    "🚀 New tutorial: Master advanced Kubernetes deployment strategies! Learn blue-green deployments, canary releases, and rolling updates with practical kubectl and Helm examples. Perfect for leveling up your DevOps skills! #Kubernetes #DevOps #CloudNative",
    "Zero-downtime deployments made easy! 📹 Just dropped a comprehensive guide to Kubernetes deployment patterns including blue-green and canary strategies. Check it out! #Kubernetes #Deployment #DevOps",
    "Ready to take your Kubernetes deployments to the next level? 🎯 New video covers advanced strategies every DevOps engineer should know. Blue-green, canary, rolling updates + more! #Kubernetes #DevOps"
  ]
}
```

#### Generate Video Highlights
```bash
curl -X POST http://localhost:8080/api/ai/highlights \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "This tutorial covers advanced Kubernetes deployment strategies including blue-green deployments, canary releases, and rolling updates. We will explore how to implement these patterns using kubectl and Helm charts."
  }'
```
Expected response:
```json
{
  "highlights": [
    "Master zero-downtime deployments with blue-green strategies",
    "Implement safe feature rollouts using canary releases",
    "Optimize rolling updates for production workloads",
    "Practical kubectl and Helm chart examples included"
  ]
}
```

#### Generate Description with Tags
```bash
curl -X POST http://localhost:8080/api/ai/description-tags \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "This tutorial covers advanced Kubernetes deployment strategies including blue-green deployments, canary releases, and rolling updates. We will explore how to implement these patterns using kubectl and Helm charts."
  }'
```
Expected response:
```json
{
  "description": "In this comprehensive tutorial, we dive deep into advanced Kubernetes deployment strategies that every DevOps engineer should master. Learn how to implement blue-green deployments for zero-downtime releases, set up canary deployments for safe feature rollouts, and optimize rolling updates for your applications.\n\nWe'll cover practical implementation using kubectl and Helm charts, with real-world examples and best practices. Perfect for intermediate to advanced Kubernetes users looking to enhance their deployment workflows.\n\nTopics covered:\n- Blue-green deployment implementation\n- Canary release strategies\n- Rolling update optimization\n- kubectl and Helm integration\n- Production deployment best practices",
  "tags": "#Kubernetes #DevOps #BlueGreen #CanaryDeployment #RollingUpdates #kubectl #Helm #CloudNative #ContainerOrchestration #DeploymentStrategies"
}
```

#### AI API Error Cases

##### Invalid JSON
```bash
curl -X POST http://localhost:8080/api/ai/titles \
  -H "Content-Type: application/json" \
  -d '{invalid json}'
```
Expected response (400 Bad Request):
```json
{
  "error": "Invalid JSON format"
}
```

##### Missing Manuscript
```bash
curl -X POST http://localhost:8080/api/ai/titles \
  -H "Content-Type: application/json" \
  -d '{}'
```
Expected response (400 Bad Request):
```json
{
  "error": "manuscript field is required and cannot be empty"
}
```

##### Empty Manuscript
```bash
curl -X POST http://localhost:8080/api/ai/titles \
  -H "Content-Type: application/json" \
  -d '{"manuscript": ""}'
```
Expected response (400 Bad Request):
```json
{
  "error": "manuscript field is required and cannot be empty"
}
```

##### Wrong HTTP Method
```bash
curl -X GET http://localhost:8080/api/ai/titles
```
Expected response (405 Method Not Allowed):
```json
{
  "error": "Method not allowed"
}
```

##### Invalid Content-Type
```bash
curl -X POST http://localhost:8080/api/ai/titles \
  -H "Content-Type: text/plain" \
  -d '{"manuscript": "test content"}'
```
Expected response (400 Bad Request):
```json
{
  "error": "Content-Type must be application/json"
}
```

#### AI API Performance Characteristics

**Response Times:**
- Typical response time: 2-8 seconds (depends on AI provider)
- Timeout: 30 seconds maximum
- Concurrent requests: Supported (up to 100 simultaneous)

**Content Limits:**
- Minimum manuscript length: 10 characters
- Maximum manuscript length: 100,000 characters
- Optimal manuscript length: 500-5,000 characters for best results

**Rate Limiting:**
- No built-in rate limiting (depends on AI provider limits)
- Recommended: Max 10 requests per minute per client

#### Test AI API with Different Content Types

##### Technical Tutorial Content
```bash
curl -X POST http://localhost:8080/api/ai/titles \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "In this comprehensive guide, we will explore Infrastructure as Code using Terraform. We will cover provider configuration, resource management, state management, and best practices for team collaboration. The tutorial includes hands-on examples with AWS, Azure, and Google Cloud Platform."
  }'
```

##### Entertainment/Gaming Content
```bash
curl -X POST http://localhost:8080/api/ai/description \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "Join me for an epic gaming session where we tackle the hardest boss fights in Elden Ring. I will share my strategies, weapon recommendations, and the mistakes you should avoid. This playthrough includes commentary and tips for both beginners and experienced players."
  }'
```

##### Educational Content
```bash
curl -X POST http://localhost:8080/api/ai/tags \
  -H "Content-Type: application/json" \
  -d '{
    "manuscript": "This mathematics lesson covers advanced calculus concepts including derivatives, integrals, and their real-world applications. We will work through step-by-step examples and solve practical problems from physics and engineering."
  }'
```

### 7. Editing Aspects Management

#### Get Editing Aspects Overview (Basic)
```bash
curl -X GET http://localhost:8080/api/editing/aspects
```
Expected response:
```json
{
  "aspects": [
    {
      "key": "initial-details",
      "title": "Initial Details",
      "description": "Basic video information and setup",
      "endpoint": "/api/videos/{videoName}/initial-details",
      "icon": "info",
      "order": 1,
      "fieldCount": 8,
      "completedFieldCount": 0
    },
    {
      "key": "work-progress",
      "title": "Work Progress",
      "description": "Video creation and production tracking",
      "endpoint": "/api/videos/{videoName}/work-progress",
      "icon": "work", 
      "order": 2,
      "fieldCount": 11,
      "completedFieldCount": 0
    },
    {
      "key": "definition",
      "title": "Definition",
      "description": "Video content definition and planning",
      "endpoint": "/api/videos/{videoName}/definition",
      "icon": "definition",
      "order": 3,
      "fieldCount": 9,
      "completedFieldCount": 0
    },
    {
      "key": "post-production",
      "title": "Post Production",
      "description": "Video editing and post-production tasks",
      "endpoint": "/api/videos/{videoName}/post-production",
      "icon": "edit",
      "order": 4,
      "fieldCount": 6,
      "completedFieldCount": 0
    },
    {
      "key": "publishing",
      "title": "Publishing",
      "description": "Video publishing and upload",
      "endpoint": "/api/videos/{videoName}/publishing",
      "icon": "publish",
      "order": 5,
      "fieldCount": 3,
      "completedFieldCount": 0
    },
    {
      "key": "post-publish",
      "title": "Post Publish",
      "description": "Post-publication tasks and follow-up activities",
      "endpoint": "/api/videos/{videoName}/post-publish",
      "icon": "post-publish",
      "order": 6,
      "fieldCount": 10,
      "completedFieldCount": 0
    }
  ]
}
```

#### **NEW**: Get Editing Aspects with Progress Tracking
```bash
# Get aspects overview with completion counts for a specific video
curl -X GET "http://localhost:8080/api/editing/aspects?videoName=test-api-video&category=test-category"
```
Expected response with completion tracking:
```json
{
  "aspects": [
    {
      "key": "initial-details",
      "title": "Initial Details", 
      "description": "Basic video information and setup",
      "endpoint": "/api/videos/{videoName}/initial-details",
      "icon": "info",
      "order": 1,
      "fieldCount": 8,
      "completedFieldCount": 6
    },
    {
      "key": "work-progress",
      "title": "Work Progress",
      "description": "Video creation and production tracking", 
      "endpoint": "/api/videos/{videoName}/work-progress",
      "icon": "work",
      "order": 2,
      "fieldCount": 11,
      "completedFieldCount": 8
    },
    {
      "key": "definition",
      "title": "Definition",
      "description": "Video content definition and planning",
      "endpoint": "/api/videos/{videoName}/definition", 
      "icon": "definition",
      "order": 3,
      "fieldCount": 9,
      "completedFieldCount": 3
    },
    {
      "key": "post-production",
      "title": "Post Production",
      "description": "Video editing and post-production tasks",
      "endpoint": "/api/videos/{videoName}/post-production",
      "icon": "edit", 
      "order": 4,
      "fieldCount": 6,
      "completedFieldCount": 2
    },
    {
      "key": "publishing",
      "title": "Publishing",
      "description": "Video publishing and upload",
      "endpoint": "/api/videos/{videoName}/publishing",
      "icon": "publish",
      "order": 5,
      "fieldCount": 3,
      "completedFieldCount": 0
    },
    {
      "key": "post-publish", 
      "title": "Post Publish",
      "description": "Post-publication tasks and follow-up activities",
      "endpoint": "/api/videos/{videoName}/post-publish",
      "icon": "post-publish",
      "order": 6,
      "fieldCount": 10,
      "completedFieldCount": 0
    }
  ]
}
```

**Progress Tracking Features:**
- **Completion Calculation**: Shows actual completed field counts (e.g., "6/8 fields completed")
- **Backend Consistency**: Uses same calculation logic as CLI progress tracking
- **Dynamic Updates**: Completion counts reflect current video state
- **Backward Compatibility**: Works with or without video context
- **Error Handling**: Graceful fallback to 0 counts if video not found

**Completion Criteria Used:**
- **String/Text/Date/Select fields**: Not empty string
- **Boolean fields**: Always complete (both true/false count as complete)
- **Number fields**: Not nil/zero value

#### Test Progress Tracking Error Cases
```bash
# Test with missing category parameter
curl -X GET "http://localhost:8080/api/editing/aspects?videoName=test-video"
```
Expected response (400 Bad Request):
```json
{
  "error": "When videoName is provided, category is also required"
}
```

```bash
# Test with non-existent video (should fallback gracefully to 0 counts)
curl -X GET "http://localhost:8080/api/editing/aspects?videoName=nonexistent&category=test-category"
```
Expected response (200 OK with 0 completion counts):
```json
{
  "aspects": [
    {
      "key": "initial-details",
      "title": "Initial Details",
      "description": "Basic video information and setup",
      "endpoint": "/api/videos/{videoName}/initial-details", 
      "icon": "info",
      "order": 1,
      "fieldCount": 8,
      "completedFieldCount": 0
    }
    // ... other aspects with completedFieldCount: 0
  ]
}
```

#### **Legacy**: Get Editing Aspects Overview (Old Format)
```bash
curl -X GET http://localhost:8080/api/editing/aspects
```
Expected response:
```json
{
  "aspects": [
    {
      "key": "initial-details",
      "title": "Initial Details",
      "description": "Basic video information and setup",
      "order": 1,
      "endpoint": "/api/videos/{videoName}/initial-details",
      "summary": {
        "fieldCount": 8,
        "requiredFields": 0,
        "hasRequiredFields": false
      }
    },
    {
      "key": "work-progress",
      "title": "Work Progress",
      "description": "Content creation and material preparation tracking",
      "order": 2,
      "endpoint": "/api/videos/{videoName}/work-progress",
      "summary": {
        "fieldCount": 11,
        "requiredFields": 0,
        "hasRequiredFields": false
      }
    },
    {
      "key": "definition",
      "title": "Definition",
      "description": "Video content definition and structure",
      "order": 3,
      "endpoint": "/api/videos/{videoName}/definition",
      "summary": {
        "fieldCount": 9,
        "requiredFields": 0,
        "hasRequiredFields": false
      }
    },
    {
      "key": "post-production",
      "title": "Post Production",
      "description": "Video editing and post-production",
      "order": 4,
      "endpoint": "/api/videos/{videoName}/post-production",
      "summary": {
        "fieldCount": 6,
        "requiredFields": 0,
        "hasRequiredFields": false
      }
    },
    {
      "key": "publishing",
      "title": "Publishing Details",
      "description": "Video publishing and distribution",
      "order": 5,
      "endpoint": "/api/videos/{videoName}/publishing",
      "summary": {
        "fieldCount": 3,
        "requiredFields": 0,
        "hasRequiredFields": false
      }
    },
    {
      "key": "post-publish",
      "title": "Post Publish",
      "description": "Post-publication tasks and follow-up activities",
      "order": 6,
      "endpoint": "/api/videos/{videoName}/post-publish",
      "summary": {
        "fieldCount": 10,
        "requiredFields": 0,
        "hasRequiredFields": false
      }
    }
  ]
}
```

**Performance Benefits:**
- **93% smaller than full metadata** (~1KB vs ~15KB)
- **Perfect for navigation menus** and phase selection UIs
- **Quick summary statistics** for each editing phase
- **Ordered by workflow sequence** (1-6)

#### Get Detailed Fields for Specific Aspect
```bash
# Get detailed field information for work-progress phase
curl -X GET http://localhost:8080/api/editing/aspects/work-progress/fields
```
Expected response:
```json
{
  "aspect": {
    "key": "work-progress",
    "title": "Work Progress",
    "description": "Content creation and material preparation tracking",
    "order": 2,
    "endpoint": "/api/videos/{videoName}/work-progress"
  },
  "fields": [
    {
      "name": "Code",
      "fieldName": "code",
      "type": "boolean",
      "required": false,
      "order": 1,
      "description": "Field description",
      "uiHints": {
        "inputType": "checkbox",
        "placeholder": "",
        "helpText": "",
        "multiline": false
      },
      "validationHints": {
        "required": false
      },
      "defaultValue": false,
      "completionCriteria": "filled_only"
    },
    {
      "name": "Head",
      "fieldName": "head",
      "type": "boolean",
      "required": false,
      "order": 2,
      "description": "Field description",
      "uiHints": {
        "inputType": "checkbox",
        "placeholder": "",
        "helpText": "",
        "multiline": false
      },
      "validationHints": {
        "required": false
      },
      "defaultValue": false,
      "completionCriteria": "filled_only"
    },
    {
      "name": "Screen",
      "fieldName": "screen",
      "type": "boolean",
      "required": false,
      "order": 3,
      "description": "Field description",
      "uiHints": {
        "inputType": "checkbox",
        "placeholder": "",
        "helpText": "",
        "multiline": false
      },
      "validationHints": {
        "required": false
      },
      "defaultValue": false,
      "completionCriteria": "filled_only"
    },
    {
      "name": "Related Videos",
      "fieldName": "relatedVideos",
      "type": "text",
      "required": false,
      "order": 4,
      "description": "List of related videos for reference",
      "uiHints": {
        "inputType": "textarea",
        "placeholder": "",
        "helpText": "",
        "rows": 3,
        "multiline": true
      },
      "validationHints": {
        "required": false
      },
      "defaultValue": "",
      "completionCriteria": "filled_only"
    }
    // ... additional fields (11 total in work-progress aspect)
  ]
}
```

**NEW: Field Name Mapping**

Each field now includes a `fieldName` property that contains the actual camelCase property name used in the video data API. This enables frontend applications to map between field metadata and video data without hardcoded conversion logic.

- `name` - Display name for UI (e.g., "Project Name")
- `fieldName` - Actual property name in video data (e.g., "projectName", "sponsorship.amount")

**NEW: Enhanced Field Structure**

Each field now includes comprehensive metadata for UI generation and validation:

- `type` - Field data type: "string", "boolean", "text", "date"
- `uiHints` - UI rendering information:
  - `inputType` - HTML input type: "text", "checkbox", "textarea", "datetime"
  - `placeholder` - User-friendly placeholder text
  - `multiline` - Whether field supports multiple lines
  - `rows` - Number of rows for textarea fields
- `validationHints` - Validation rules and requirements
- `defaultValue` - Default value for new videos
- `description` - Field purpose and usage description

**Date Field Format**

Date fields use UTC datetime format with specific UI hints:
```json
{
  "name": "Date",
  "fieldName": "date",
  "type": "date",
  "uiHints": {
    "inputType": "datetime",
    "placeholder": "YYYY-MM-DDTHH:MM",
    "helpText": "",
    "multiline": false
  }
}
```
- **Format**: `YYYY-MM-DDTHH:MM` (e.g., "2024-01-15T14:30")
- **Input Type**: `datetime` for UTC compatibility
- **Validation**: Must match Go time format `2006-01-02T15:04`

**Text Field Format**

Multi-line text fields use textarea input with specific UI hints:
```json
{
  "name": "Description",
  "fieldName": "description",
  "type": "text",
  "uiHints": {
    "inputType": "textarea",
    "placeholder": "",
    "helpText": "",
    "rows": 3,
    "multiline": true
  }
}
```
- **Input Type**: `textarea` for multi-line content
- **Rows**: Default 3 rows for textarea height
- **Multiline**: Always `true` for text type fields

**NEW: Completion Criteria Field**

Each field now includes a `completionCriteria` property that defines when the field should be considered "complete" for progress tracking and UI styling:

- `"filled_only"` - Complete when field has any non-empty value (not "-" or empty string)
- `"empty_or_filled"` - Always considered complete regardless of value
- `"conditional"` - Complex logic based on other field values (e.g., sponsored emails only required when sponsor field is set)
- `"true_only"` - Complete only when boolean field is `true`
- `"false_only"` - Complete only when boolean field is `false`
- `"filled_required"` - Must be filled (similar to filled_only but stricter validation)

**Example completion criteria by field type:**
```bash
# Get initial-details fields to see different completion criteria
curl -X GET http://localhost:8080/api/editing/aspects/initial-details/fields
```
Expected different completion criteria:
```json
{
  "fields": [
    {
      "name": "Project Name",
      "fieldName": "projectName",
      "type": "string",
      "required": false,
      "order": 1,
      "description": "Name of the related project",
      "uiHints": {
        "inputType": "text",
        "placeholder": "",
        "helpText": "",
        "multiline": false
      },
      "validationHints": {
        "required": false
      },
      "defaultValue": "",
      "completionCriteria": "filled_only"
    },
    {
      "name": "Project URL",
      "fieldName": "projectURL",
      "type": "string",
      "required": false,
      "order": 2,
      "description": "URL to the project repository or documentation",
      "uiHints": {
        "inputType": "text",
        "placeholder": "",
        "helpText": "",
        "multiline": false
      },
      "validationHints": {
        "required": false
      },
      "defaultValue": "",
      "completionCriteria": "filled_only"
    },
    {
      "name": "Sponsorship Amount",
      "fieldName": "sponsorship.amount",
      "type": "string",
      "required": false,
      "order": 3,
      "description": "Field description",
      "uiHints": {
        "inputType": "text",
        "placeholder": "",
        "helpText": "",
        "multiline": false
      },
      "validationHints": {
        "required": false
      },
      "defaultValue": "",
      "completionCriteria": "filled_only"
    }
  ]
}
```

**Use Cases for Completion Criteria:**
- **Progress Calculation**: Determine field completion for progress bars
- **UI Styling**: Apply different colors/styles based on completion status
- **Validation**: Implement consistent validation logic between CLI and web interfaces
- **Workflow Logic**: Determine when phases or aspects can be considered complete

#### Test Different Aspect Keys
```bash
# Initial details phase
curl -X GET http://localhost:8080/api/editing/aspects/initial-details/fields

# Definition phase  
curl -X GET http://localhost:8080/api/editing/aspects/definition/fields

# Post-production phase
curl -X GET http://localhost:8080/api/editing/aspects/post-production/fields

# Publishing phase
curl -X GET http://localhost:8080/api/editing/aspects/publishing/fields

# Post-publish phase
curl -X GET http://localhost:8080/api/editing/aspects/post-publish/fields
```

#### Error Cases for Editing Aspects

##### Invalid Aspect Key
```bash
curl -X GET http://localhost:8080/api/editing/aspects/invalid-key/fields
```
Expected response (400 Bad Request):
```json
{
  "error": "Invalid aspect key: invalid-key"
}
```

##### Wrong HTTP Method
```bash
curl -X POST http://localhost:8080/api/editing/aspects
```
Expected response (405 Method Not Allowed):
```json
{
  "error": "Method not allowed"
}
```

**Use Cases for Frontend Development:**
- **Dynamic Form Generation**: Use field metadata to build editing forms
- **Navigation Menus**: Display aspects with field counts and completion status
- **Progress Tracking**: Show summary stats for each editing phase (NEW: with completedFieldCount)
- **Field Validation**: Use required/type information for client-side validation
- **Help Text**: Display field-specific guidance to users
- **Workflow Ordering**: Present editing phases in the correct sequence (1-6)

### 8. Error Testing Scenarios

#### Test Invalid Phase Parameter
```bash
curl -X GET "http://localhost:8080/api/videos?phase=invalid"
```
Expected response (400 Bad Request):
```json
{
  "error": "Invalid phase parameter",
  "message": "strconv.Atoi: parsing \"invalid\": invalid syntax"
}
```

#### Test Missing Required Parameters
```bash
curl -X GET "http://localhost:8080/api/videos/nonexistent?category=test-category"
```
Expected response (404 Not Found):
```json
{
  "error": "Video not found",
  "message": "failed to get video nonexistent: ..."
}
```

#### Test Invalid JSON
```bash
curl -X POST http://localhost:8080/api/videos \
  -H "Content-Type: application/json" \
  -d '{invalid json}'
```
Expected response (400 Bad Request):
```json
{
  "error": "Invalid JSON",
  "message": "invalid character 'i' looking for beginning of object key string"
}
```

## Complete Workflow Test

Follow these steps to test a complete video workflow:

1. **Create Video**: Use the create video endpoint
2. **Check Phases**: Verify the video appears in "Ideas" phase
3. **Update Initial Details**: Add project information and publish date
4. **Update Work Progress**: Mark content creation tasks as complete
5. **Update Definition**: Add title, description, and metadata
6. **Test AI Content Generation**: Use AI endpoints to generate titles, descriptions, tags, tweets, and highlights
7. **Update Post-Production**: Add thumbnail and editing information
8. **Update Publishing**: Simulate video upload and Hugo post creation
9. **Update Post-Publish**: Mark social media and follow-up tasks as complete
10. **Verify Phase Progression**: Check that the video has moved through phases
11. **Move Video**: Test moving the video to a different category
12. **Delete Video**: Clean up by deleting the test video

**NEW**: Test the enhanced editing aspects endpoint during the workflow:
13. **Check Progress**: Use `GET /api/editing/aspects?videoName=test-api-video&category=test-category` after each phase update to verify completion counts

## Notes

- Replace `localhost:8080` with your actual server address and port
- Ensure you have the necessary permissions to read/write files in the manuscript directory
- The API uses JSON for all request/response bodies
- Some operations (like YouTube upload) are placeholders in this implementation
- Check the server logs for detailed error information if requests fail
- **NEW**: The enhanced editing aspects endpoint with completion tracking provides real-time progress updates for better UI development
