# PRD: Google Calendar Integration for Video Release Reminders

**Issue**: #350
**Status**: Draft
**Created**: 2025-11-29
**Last Updated**: 2025-11-29

## Problem Statement

After uploading videos to YouTube, creators need to be present when videos go live to perform manual tasks such as:
- Posting on X (Twitter)
- Monitoring and responding to early comments
- Engaging with early viewers
- Sharing on additional platforms

Currently, there's no automated reminder system to ensure the creator is available during these critical 30-minute windows before and after video release.

## Solution Overview

Automatically create Google Calendar events after video uploads that block time from 30 minutes before to 30 minutes after the scheduled publish time. These calendar events serve as reminders and time blocks for video release activities.

## User Stories

### Primary User Story
**As a** content creator
**I want** automatic calendar reminders when my videos go live
**So that** I can be present to perform manual tasks (posting on X, engaging with viewers, monitoring comments)

### User Journey
1. Creator uploads a video through the YouTube Automation Tool (CLI or API mode)
2. Video is successfully uploaded to YouTube with scheduled publish time
3. System automatically creates Google Calendar event:
   - Start time: 30 minutes before video publish time
   - End time: 30 minutes after video publish time
   - Title: "📺 Video Release: [video-title]"
   - Description: Includes YouTube link and relevant details
4. Creator receives calendar notification at appropriate times
5. Creator performs manual tasks during the blocked time window

## Success Criteria

### Must Have
- [ ] Google Calendar API integration with OAuth2 authentication
- [ ] Automatic calendar event creation after successful video upload
- [ ] Calendar events include correct timing (publish time ± 30 minutes)
- [ ] Event title includes video title with clear visual indicator (📺)
- [ ] Event description includes YouTube video URL
- [ ] Works in both CLI and API modes
- [ ] Proper error handling if calendar creation fails (doesn't block upload process)
- [ ] Configuration in settings.yaml for Google Calendar credentials

### Should Have
- [ ] Clear success/failure messages when calendar events are created
- [ ] Event description includes additional context (category, tags, etc.)
- [ ] Logging of calendar event creation for debugging

### Could Have
- [ ] Calendar event color coding based on video category
- [ ] Option to specify which Google Calendar to use (if user has multiple)
- [ ] Different time windows based on video type (could be configurable)

### Won't Have (This Release)
- Updating calendar events if publish time changes (requirement: publish time doesn't change after upload)
- Manual calendar event creation/deletion from the tool
- Calendar event creation for videos not yet uploaded
- Integration with other calendar systems (Outlook, Apple Calendar, etc.)

## Technical Approach

### Google Calendar API Integration
- Use Google Calendar API v3
- OAuth2 authentication flow (similar to existing YouTube API auth)
- Store credentials in environment variables or secure configuration
- Create events using `Events.insert()` API method

### Implementation Points
1. **Authentication Module** (`internal/platform/calendar.go` or similar)
   - OAuth2 setup and credential management
   - Token storage and refresh logic
   - Reuse patterns from existing YouTube OAuth implementation

2. **Calendar Event Creation**
   - Triggered after successful video upload in `internal/publishing/youtube.go`
   - Calculate event times based on video publish time
   - Format event data (title, description, start/end times)
   - Handle API errors gracefully

3. **Configuration Updates**
   - Add Google Calendar settings to `settings.yaml`
   - Environment variables for client ID, client secret, tokens
   - Optional: enable/disable calendar integration flag

4. **Error Handling**
   - Calendar creation failures should not block video upload
   - Log errors but allow upload process to complete
   - Return warnings to user if calendar event fails

### Data Flow
```
Video Upload Success
    ↓
Extract publish time & video details
    ↓
Calculate event times (publish ± 30 min)
    ↓
Format calendar event data
    ↓
Call Google Calendar API
    ↓
Log result (success/failure)
    ↓
Continue with normal upload flow
```

## Milestones

### 1. Google Calendar API Setup & Authentication
- [ ] Research and configure Google Calendar API credentials
- [ ] Implement OAuth2 authentication flow
- [ ] Create credential storage and token management
- [ ] Test authentication flow end-to-end
- **Validation**: Successfully authenticate and access user's calendar

### 2. Core Calendar Event Creation
- [ ] Implement calendar event creation logic
- [ ] Calculate correct event times (publish time ± 30 minutes)
- [ ] Format event title and description with video details
- [ ] Integrate with video upload success flow
- **Validation**: Calendar events created after video uploads with correct timing and details

### 3. Error Handling & Resilience
- [ ] Implement graceful error handling for API failures
- [ ] Ensure upload process continues if calendar creation fails
- [ ] Add logging and user feedback for calendar operations
- [ ] Handle edge cases (timezone issues, missing publish times, etc.)
- **Validation**: Upload succeeds even when calendar API fails; clear error messages

### 4. Configuration & Settings
- [ ] Add calendar settings to settings.yaml
- [ ] Document environment variable setup
- [ ] Add enable/disable flag for calendar integration
- [ ] Update configuration documentation
- **Validation**: Calendar integration configurable through settings

### 5. Testing & Documentation
- [ ] Write unit tests for calendar event creation logic
- [ ] Write integration tests with mocked Google Calendar API
- [ ] Update user documentation with setup instructions
- [ ] Test both CLI and API modes
- [ ] Achieve 80% test coverage
- **Validation**: All tests pass, documentation complete, feature ready for use

## Dependencies

### External Services
- **Google Calendar API v3**: Primary dependency for calendar operations
- **Google OAuth2**: Required for API authentication

### Internal Dependencies
- YouTube upload flow in `internal/publishing/youtube.go`
- Configuration system in `settings.yaml`
- OAuth pattern from existing YouTube integration

### Development Dependencies
- Google API Go client library (`google.golang.org/api/calendar/v3`)
- OAuth2 library (`golang.org/x/oauth2`)

## Risks & Mitigations

### Risk 1: Google Calendar API Authentication Complexity
**Impact**: High
**Likelihood**: Medium
**Mitigation**: Reuse existing YouTube OAuth patterns; extensive testing of auth flow

### Risk 2: Calendar Creation Failures Block Upload Process
**Impact**: High
**Likelihood**: Low
**Mitigation**: Implement graceful error handling; calendar creation is non-blocking

### Risk 3: Timezone Handling Issues
**Impact**: Medium
**Likelihood**: Medium
**Mitigation**: Use UTC consistently; convert to user's timezone only for display; test with various timezone scenarios

### Risk 4: API Rate Limits
**Impact**: Low
**Likelihood**: Low
**Mitigation**: Monitor API usage; implement retry logic if needed

## Open Questions

1. **Should calendar events be created for all videos or only certain phases/categories?**
   - Current assumption: All uploaded videos
   - Could add filtering based on category or phase

2. **What happens if a video is uploaded but publish time is not set?**
   - Skip calendar event creation?
   - Create event with placeholder time?

3. **Should there be a way to manually create calendar events for existing videos?**
   - Current scope: Only automatic creation after upload
   - Could be future enhancement

4. **Should we support multiple Google Calendar accounts?**
   - Current scope: Single account
   - Could be future enhancement

## Out of Scope

- Updating calendar events after creation (publish time doesn't change after upload per requirements)
- Deleting calendar events when videos are deleted
- Calendar event creation for videos in earlier phases (pre-upload)
- Integration with non-Google calendar systems
- Calendar sync/two-way updates
- Mobile app notifications (relies on user's existing calendar app)
- Custom time window configuration (fixed at ± 30 minutes)

## Future Enhancements

- Configurable time windows (e.g., ± 15 min, ± 60 min)
- Calendar event updates if publish time changes
- Different event types based on video category
- Calendar event templates
- Bulk calendar event creation for existing videos
- Integration with other calendar systems
- Calendar event deletion when videos are removed

## Progress Log

### 2025-11-29
- PRD created
- GitHub issue #350 opened
- Core requirements documented
- 5 major milestones defined

---

**Next Steps**: Begin implementation with Milestone 1 (Google Calendar API Setup & Authentication)
