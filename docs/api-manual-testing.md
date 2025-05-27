# YouTube Automation REST API - Manual Testing Guide

This guide provides step-by-step manual testing scenarios for the YouTube Automation REST API using `curl` commands.

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

### 4. Individual Video Operations

#### Get Specific Video Details
```bash
curl -X GET "http://localhost:8080/api/videos/test-api-video?category=test-category"
```
Expected response:
```json
{
  "video": {
    "name": "test-api-video",
    "category": "test-category",
    "path": "manuscript/test-category/test-api-video.yaml",
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
    "publishDate": "2023-12-25T10:00",
    "gistPath": "manuscript/test-category/test-api-video.md"
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
    "codeDone": true,
    "talkingHeadDone": true,
    "screenRecordingDone": true,
    "thumbnailsDone": true,
    "diagramsDone": true,
    "screenshotsDone": true,
    "filesLocation": "https://drive.google.com/folder/example",
    "tagline": "Learn API testing with this comprehensive guide"
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
    "tweetText": "Just released a comprehensive guide to REST API testing! Perfect for developers looking to improve their testing skills. 🚀 #API #Testing",
    "requestThumbnailGeneration": true
  }'
```

#### Update Video - Post-Production Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/post-production?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "thumbnailPath": "/path/to/thumbnail.jpg",
    "members": "John Doe, Jane Smith",
    "requestEdit": true,
    "timecodes": "00:00 - Introduction\n05:00 - Setup\n15:00 - Testing\n25:00 - Conclusion",
    "movieDone": true,
    "slidesDone": true
  }'
```

#### Update Video - Publishing Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/publishing?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "videoFilePath": "/path/to/final-video.mp4",
    "uploadToYouTube": true,
    "createHugoPost": true
  }'
```

#### Update Video - Post-Publish Phase
```bash
curl -X PUT "http://localhost:8080/api/videos/test-api-video/post-publish?category=test-category" \
  -H "Content-Type: application/json" \
  -d '{
    "blueSkyPostSent": true,
    "linkedInPostSent": true,
    "slackPostSent": true,
    "youTubeHighlightCreated": true,
    "youTubePinnedCommentAdded": true,
    "repliedToYouTubeComments": true,
    "gdeAdvocuPostSent": false,
    "codeRepositoryURL": "https://github.com/example/api-testing-guide",
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

### 6. Error Testing Scenarios

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
6. **Update Post-Production**: Add thumbnail and editing information
7. **Update Publishing**: Simulate video upload and Hugo post creation
8. **Update Post-Publish**: Mark social media and follow-up tasks as complete
9. **Verify Phase Progression**: Check that the video has moved through phases
10. **Move Video**: Test moving the video to a different category
11. **Delete Video**: Clean up by deleting the test video

## Notes

- Replace `localhost:8080` with your actual server address and port
- Ensure you have the necessary permissions to read/write files in the manuscript directory
- The API uses JSON for all request/response bodies
- Some operations (like YouTube upload) are placeholders in this implementation
- Check the server logs for detailed error information if requests fail