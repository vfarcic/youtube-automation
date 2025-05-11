# Product Requirements Document: Fix Bluesky Post Creation Error

## Overview
The application is currently experiencing an error when attempting to post content to Bluesky. The specific error occurs when a post contains an external embed with a thumbnail. The error message indicates that the thumbnail reference is not being properly formatted as a "blob ref" as required by Bluesky's API.

Error: "Invalid app.bsky.feed.post record: Record/embed/external/thumb should be a blob ref"

## Objectives
1. Diagnose the exact cause of the blob reference issue in the Bluesky posting functionality
2. Implement a fix that properly handles thumbnail images according to Bluesky's API requirements
3. Add appropriate error handling for Bluesky post creation
4. Add tests to verify the fix works correctly
5. Document the solution for future reference

## Requirements

### Technical Requirements
1. Investigate the current implementation of the Bluesky posting functionality to identify how thumbnails are currently being handled
2. Research Bluesky's API documentation to understand the correct format for "blob ref" in post embeds
3. Fix the code to properly upload thumbnail images to Bluesky before referencing them in posts
4. Ensure proper blob references are used in the post creation request
5. Implement validation to catch invalid image formats before attempting to post
6. Add comprehensive error handling to provide clear feedback on posting failures
7. Add unit tests to verify the fix works across different posting scenarios

### User Experience Requirements
1. Users should receive clear error messages when their posts fail
2. The application should attempt to automatically fix common image format issues when possible
3. The posting flow should remain simple and not require additional steps from users

### Non-Functional Requirements
1. The fix should not significantly increase post creation time
2. The solution should be backward compatible with existing post data
3. The implementation should follow best practices for the codebase

## Constraints
1. Changes should be focused on the Bluesky posting functionality only
2. The fix must adhere to Bluesky's API specifications and requirements
3. Implementation should be done with minimal disruption to other parts of the codebase

## Success Criteria
1. Posts with external content and thumbnails successfully upload to Bluesky
2. No more "Invalid blob ref" errors are reported by users
3. All tests pass successfully
4. The posting process remains user-friendly and efficient

## Out of Scope
1. Redesigning the entire posting system
2. Adding new social media platforms
3. Changing core application functionality not related to Bluesky posting 