# YouTube Automation API Testing Guide

This document provides step-by-step instructions for manually testing the YouTube Automation REST API.

## Prerequisites

- YouTube Automation CLI built and available locally
- Basic HTTP client such as [curl](https://curl.se/) or [Postman](https://www.postman.com/)
- Required configuration settings (same as for the CLI)

## Starting the API Server

1. Start the API server by running the YouTube Automation CLI with the `--api` flag:

   ```bash
   ./youtube-automation --api --api-port=8080 \
     --email-from=your-email@example.com \
     --email-thumbnail-to=thumbnail@example.com \
     --email-edit-to=edit@example.com \
     --email-finance-to=finance@example.com \
     --email-****** \
     --ai-endpoint=your-ai-endpoint \
     --ai-key=your-ai-key \
     --ai-deployment=your-ai-deployment \
     --youtube-api-key=your-youtube-api-key \
     --hugo-path=/path/to/hugo
   ```

   Alternatively, you can set these values in your `settings.yaml` file or environment variables.

2. The server will start on the specified port (defaults to 8080 if not specified).

3. Verify the server is running by checking the health endpoint:

   ```bash
   curl http://localhost:8080/health
   ```

   Expected response:
   ```json
   {"status":"ok"}
   ```

## Testing API Endpoints

Below are examples for testing the main API endpoints using curl. For the complete API reference, see [API Reference](api_reference.md).

### 1. List Video Categories

```bash
curl -X GET http://localhost:8080/api/categories
```

### 2. List Video Phases

```bash
curl -X GET http://localhost:8080/api/videos/phases
```

### 3. List Videos in a Phase

```bash
# List videos in the "Initial" phase (phase_id = 0)
curl -X GET http://localhost:8080/api/videos?phase=0
```

### 4. Create a New Video

```bash
curl -X POST http://localhost:8080/api/videos \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Video","category":"test-category"}'
```

### 5. Get a Specific Video

```bash
# Replace "Test Video,test-category" with your video_id (format: "name,category")
curl -X GET http://localhost:8080/api/videos/Test%20Video,test-category
```

### 6. Update a Video's Initial Phase

```bash
# Replace "Test Video,test-category" with your video_id
curl -X PUT http://localhost:8080/api/videos/Test%20Video,test-category/initial \
  -H "Content-Type: application/json" \
  -d '{
    "projectName": "Test Project",
    "projectURL": "https://example.com",
    "sponsorshipAmount": "$1000",
    "sponsorshipEmails": "sponsor@example.com",
    "publishDate": "2023-12-31"
  }'
```

### 7. Update a Video's Work Phase

```bash
# Replace "Test Video,test-category" with your video_id
curl -X PUT http://localhost:8080/api/videos/Test%20Video,test-category/work \
  -H "Content-Type: application/json" \
  -d '{
    "codeDone": true,
    "talkingHeadDone": true,
    "screenRecordingDone": false,
    "thumbnailsDone": false,
    "tagline": "Example Tagline"
  }'
```

### 8. Update a Video's Definition Phase

```bash
# Replace "Test Video,test-category" with your video_id
curl -X PUT http://localhost:8080/api/videos/Test%20Video,test-category/definition \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Test Video Title",
    "description": "Test Video Description",
    "tags": "test,api,demo",
    "descriptionTags": "#test #api #demo",
    "tweetText": "Check out my new test video!"
  }'
```

### 9. Delete a Video

```bash
# Replace "Test Video,test-category" with your video_id
curl -X DELETE http://localhost:8080/api/videos/Test%20Video,test-category
```

## Testing a Complete Workflow

Here's a sequence to test a complete video workflow through all phases:

1. Create a new video
2. Update the Initial phase
3. Update the Work phase
4. Update the Definition phase
5. Update the Post-Production phase
6. Update the Publishing phase
7. Update the Post-Publish phase

## API Error Testing

Test error handling by sending invalid requests:

1. Request a non-existent video:
   ```bash
   curl -X GET http://localhost:8080/api/videos/NonExistent,unknown
   ```

2. Send malformed JSON:
   ```bash
   curl -X POST http://localhost:8080/api/videos \
     -H "Content-Type: application/json" \
     -d '{"name":"Test Video" "category":"invalid-json"}'
   ```

3. Send a request with missing required fields:
   ```bash
   curl -X POST http://localhost:8080/api/videos \
     -H "Content-Type: application/json" \
     -d '{}'
   ```

## Using Postman

If you prefer using Postman:

1. Import the API endpoints as a collection:
   - Create a new collection named "YouTube Automation API"
   - Add requests for each endpoint described above
   - Set the base URL to `http://localhost:8080`

2. Create environment variables for common values:
   - `baseUrl`: http://localhost:8080
   - `videoId`: Test Video,test-category (after creating a video)

3. Use the Postman interface to send requests and view responses.

## Troubleshooting

- **Connection refused**: Ensure the API server is running and you're using the correct port.
- **404 Not Found**: Check that your URL is correct and you're using the correct endpoint path.
- **400 Bad Request**: Verify your request body is properly formatted JSON and includes all required fields.
- **500 Internal Server Error**: Check the server logs for more information about the error.

## Next Steps

After manually testing the API, you might want to:

- Create automated tests using tools like Postman Collections or a test framework
- Integrate the API with your application using HTTP clients
- Set up monitoring for the API endpoints
