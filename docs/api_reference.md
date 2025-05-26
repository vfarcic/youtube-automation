# YouTube Automation API Reference

This document provides documentation for the REST API endpoints available in the YouTube Automation tool.

## Base URL

All API endpoints are prefixed with `/api`.

## Authentication

Currently, the API does not implement authentication. It should be run in a trusted environment.

## Videos

### List Video Phases

Returns all video phases with counts of videos in each phase.

**Endpoint**: `GET /api/videos/phases`

**Response**:

```json
[
  {
    "name": "Initial",
    "count": 5,
    "id": 0
  },
  {
    "name": "Work",
    "count": 3,
    "id": 1
  },
  /* ... other phases ... */
]
```

### List Videos by Phase

Returns all videos in a specific phase.

**Endpoint**: `GET /api/videos?phase={phase_id}`

**Parameters**:
- `phase_id` - The ID of the phase (0 = Initial, 1 = Work, 2 = Definition, 3 = Post-Production, 4 = Publishing, 5 = Post-Publish)

**Response**:
```json
[
  {
    "Name": "Example Video",
    "Category": "example-category",
    "Path": "manuscript/example-category/example-video.yaml",
    /* ... other video fields ... */
  },
  /* ... other videos ... */
]
```

### Get Video by ID

Returns a specific video.

**Endpoint**: `GET /api/videos/{video_id}`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Response**:
```json
{
  "Name": "Example Video",
  "Category": "example-category",
  "Path": "manuscript/example-category/example-video.yaml",
  /* ... other video fields ... */
}
```

### Create Video

Creates a new video.

**Endpoint**: `POST /api/videos`

**Request Body**:
```json
{
  "name": "New Video",
  "category": "example-category"
}
```

**Response**:
```json
{
  "Name": "New Video",
  "Category": "example-category",
  "Path": "manuscript/example-category/new-video.yaml",
  /* ... other video fields ... */
}
```

### Update Video

Updates a specific video.

**Endpoint**: `PUT /api/videos/{video_id}`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "Name": "Updated Video",
  "Category": "example-category",
  /* ... other video fields ... */
}
```

**Response**:
```json
{
  "Name": "Updated Video",
  "Category": "example-category",
  "Path": "manuscript/example-category/updated-video.yaml",
  /* ... other video fields ... */
}
```

### Delete Video

Deletes a specific video.

**Endpoint**: `DELETE /api/videos/{video_id}`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Response**:
- `204 No Content` on success

### Move Video Files

Moves video files to a new directory.

**Endpoint**: `POST /api/videos/{video_id}/move`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "target_directory_path": "/path/to/target/directory"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Video files moved successfully"
}
```

## Video Phases

Each phase of a video has its own update endpoint:

### Update Initial Phase

**Endpoint**: `PUT /api/videos/{video_id}/initial`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "projectName": "Project Name",
  "projectURL": "https://example.com",
  "sponsorshipAmount": "$1000",
  "sponsorshipEmails": "sponsor@example.com",
  "sponsorshipBlockedReason": "",
  "publishDate": "2023-01-01",
  "delayed": false,
  "gistPath": "https://gist.github.com/example"
}
```

### Update Work Phase

**Endpoint**: `PUT /api/videos/{video_id}/work`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "codeDone": true,
  "talkingHeadDone": true,
  "screenRecordingDone": false,
  "relatedVideos": "Video1, Video2",
  "thumbnailsDone": false,
  "diagramsDone": true,
  "screenshotsDone": true,
  "filesLocation": "/path/to/files",
  "tagline": "Example Tagline",
  "taglineIdeas": "Idea 1, Idea 2",
  "otherLogosAssets": "Logo1, Logo2"
}
```

### Update Definition Phase

**Endpoint**: `PUT /api/videos/{video_id}/definition`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "title": "Video Title",
  "description": "Video Description",
  "highlight": "Video Highlight",
  "tags": "tag1,tag2",
  "descriptionTags": "#tag1 #tag2",
  "tweetText": "Check out my new video!",
  "animationsScript": "Animation script...",
  "requestThumbnailGeneration": true
}
```

### Update Post-Production Phase

**Endpoint**: `PUT /api/videos/{video_id}/post-production`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "thumbnailPath": "/path/to/thumbnail.jpg",
  "members": "Member content...",
  "requestEdit": true,
  "timecodes": "00:00 Intro, 01:23 Topic",
  "movieDone": false,
  "slidesDone": true
}
```

### Update Publishing Phase

**Endpoint**: `PUT /api/videos/{video_id}/publishing`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "videoFilePath": "/path/to/video.mp4",
  "uploadToYouTube": true,
  "createHugoPost": true
}
```

### Update Post-Publish Phase

**Endpoint**: `PUT /api/videos/{video_id}/post-publish`

**Parameters**:
- `video_id` - The ID of the video in the format `name,category`

**Request Body**:
```json
{
  "blueSkyPostSent": true,
  "linkedInPostSent": false,
  "slackPostSent": true,
  "youTubeHighlightCreated": false,
  "youTubePinnedCommentAdded": true,
  "repliedToYouTubeComments": false,
  "gdeAdvocuPostSent": false,
  "codeRepositoryURL": "https://github.com/example/repo",
  "notifiedSponsors": true
}
```

## Categories

### List Categories

Returns all available video categories.

**Endpoint**: `GET /api/categories`

**Response**:
```json
{
  "categories": [
    "example-category",
    "another-category"
  ]
}
```

## Health Check

### Check API Health

Returns the health status of the API.

**Endpoint**: `GET /health`

**Response**:
```json
{
  "status": "ok"
}
```

## Error Responses

All API endpoints return appropriate HTTP status codes:

- `200 OK` - Request succeeded
- `201 Created` - Resource was created
- `204 No Content` - Request succeeded with no content to return
- `400 Bad Request` - Invalid request parameters or body
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server-side error

Error responses follow this format:
```json
{
  "status": 400,
  "message": "Error message"
}
```