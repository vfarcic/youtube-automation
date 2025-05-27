package api

import (
	"reflect"

	"devopstoolkit/youtube-automation/internal/storage"
)

// applyInitialDetailsUpdates applies updates to the initial details phase
func (s *Server) applyInitialDetailsUpdates(video *storage.Video, updateData map[string]interface{}) error {
	if val, ok := updateData["projectName"]; ok {
		if str, ok := val.(string); ok {
			video.ProjectName = str
		}
	}
	if val, ok := updateData["projectURL"]; ok {
		if str, ok := val.(string); ok {
			video.ProjectURL = str
		}
	}
	if val, ok := updateData["sponsorshipAmount"]; ok {
		if str, ok := val.(string); ok {
			video.Sponsorship.Amount = str
		}
	}
	if val, ok := updateData["sponsorshipEmails"]; ok {
		if str, ok := val.(string); ok {
			video.Sponsorship.Emails = str
		}
	}
	if val, ok := updateData["sponsorshipBlockedReason"]; ok {
		if str, ok := val.(string); ok {
			video.Sponsorship.Blocked = str
		}
	}
	if val, ok := updateData["publishDate"]; ok {
		if str, ok := val.(string); ok {
			video.Date = str
		}
	}
	if val, ok := updateData["delayed"]; ok {
		if b, ok := val.(bool); ok {
			video.Delayed = b
		}
	}
	if val, ok := updateData["gistPath"]; ok {
		if str, ok := val.(string); ok {
			video.Gist = str
		}
	}

	// Update completion counts
	s.updateInitialDetailsCompletion(video)
	return nil
}

// applyWorkProgressUpdates applies updates to the work progress phase
func (s *Server) applyWorkProgressUpdates(video *storage.Video, updateData map[string]interface{}) error {
	if val, ok := updateData["codeDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Code = b
		}
	}
	if val, ok := updateData["talkingHeadDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Head = b
		}
	}
	if val, ok := updateData["screenRecordingDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Screen = b
		}
	}
	if val, ok := updateData["relatedVideos"]; ok {
		if str, ok := val.(string); ok {
			video.RelatedVideos = str
		}
	}
	if val, ok := updateData["thumbnailsDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Thumbnails = b
		}
	}
	if val, ok := updateData["diagramsDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Diagrams = b
		}
	}
	if val, ok := updateData["screenshotsDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Screenshots = b
		}
	}
	if val, ok := updateData["filesLocation"]; ok {
		if str, ok := val.(string); ok {
			video.Location = str
		}
	}
	if val, ok := updateData["tagline"]; ok {
		if str, ok := val.(string); ok {
			video.Tagline = str
		}
	}
	if val, ok := updateData["taglineIdeas"]; ok {
		if str, ok := val.(string); ok {
			video.TaglineIdeas = str
		}
	}
	if val, ok := updateData["otherLogosAssets"]; ok {
		if str, ok := val.(string); ok {
			video.OtherLogos = str
		}
	}

	// Update completion counts
	s.updateWorkProgressCompletion(video)
	return nil
}

// applyDefinitionUpdates applies updates to the definition phase
func (s *Server) applyDefinitionUpdates(video *storage.Video, updateData map[string]interface{}) error {
	if val, ok := updateData["title"]; ok {
		if str, ok := val.(string); ok {
			video.Title = str
		}
	}
	if val, ok := updateData["description"]; ok {
		if str, ok := val.(string); ok {
			video.Description = str
		}
	}
	if val, ok := updateData["highlight"]; ok {
		if str, ok := val.(string); ok {
			video.Highlight = str
		}
	}
	if val, ok := updateData["tags"]; ok {
		if str, ok := val.(string); ok {
			video.Tags = str
		}
	}
	if val, ok := updateData["descriptionTags"]; ok {
		if str, ok := val.(string); ok {
			video.DescriptionTags = str
		}
	}
	if val, ok := updateData["tweetText"]; ok {
		if str, ok := val.(string); ok {
			video.Tweet = str
		}
	}
	if val, ok := updateData["animationsScript"]; ok {
		if str, ok := val.(string); ok {
			video.Animations = str
		}
	}
	if val, ok := updateData["requestThumbnailGeneration"]; ok {
		if b, ok := val.(bool); ok {
			video.RequestThumbnail = b
		}
	}

	// Update completion counts
	s.updateDefinitionCompletion(video)
	return nil
}

// applyPostProductionUpdates applies updates to the post-production phase
func (s *Server) applyPostProductionUpdates(video *storage.Video, updateData map[string]interface{}) error {
	if val, ok := updateData["thumbnailPath"]; ok {
		if str, ok := val.(string); ok {
			video.Thumbnail = str
		}
	}
	if val, ok := updateData["members"]; ok {
		if str, ok := val.(string); ok {
			video.Members = str
		}
	}
	if val, ok := updateData["requestEdit"]; ok {
		if b, ok := val.(bool); ok {
			video.RequestEdit = b
		}
	}
	if val, ok := updateData["timecodes"]; ok {
		if str, ok := val.(string); ok {
			video.Timecodes = str
		}
	}
	if val, ok := updateData["movieDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Movie = b
		}
	}
	if val, ok := updateData["slidesDone"]; ok {
		if b, ok := val.(bool); ok {
			video.Slides = b
		}
	}

	// Update completion counts
	s.updatePostProductionCompletion(video)
	return nil
}

// applyPublishingUpdates applies updates to the publishing phase
func (s *Server) applyPublishingUpdates(video *storage.Video, updateData map[string]interface{}) error {
	if val, ok := updateData["videoFilePath"]; ok {
		if str, ok := val.(string); ok {
			video.UploadVideo = str
		}
	}
	if val, ok := updateData["uploadToYouTube"]; ok {
		if b, ok := val.(bool); ok && b {
			// TODO: Implement YouTube upload logic
			// For now, just mark as completed
			video.VideoId = "placeholder-youtube-id"
		}
	}
	if val, ok := updateData["createHugoPost"]; ok {
		if b, ok := val.(bool); ok && b {
			// TODO: Implement Hugo post creation logic
			// For now, just mark as completed
			video.HugoPath = "placeholder-hugo-path"
		}
	}

	// Update completion counts
	s.updatePublishingCompletion(video)
	return nil
}

// applyPostPublishUpdates applies updates to the post-publish phase
func (s *Server) applyPostPublishUpdates(video *storage.Video, updateData map[string]interface{}) error {
	if val, ok := updateData["blueSkyPostSent"]; ok {
		if b, ok := val.(bool); ok {
			video.BlueSkyPosted = b
		}
	}
	if val, ok := updateData["linkedInPostSent"]; ok {
		if b, ok := val.(bool); ok {
			video.LinkedInPosted = b
		}
	}
	if val, ok := updateData["slackPostSent"]; ok {
		if b, ok := val.(bool); ok {
			video.SlackPosted = b
		}
	}
	if val, ok := updateData["youTubeHighlightCreated"]; ok {
		if b, ok := val.(bool); ok {
			video.YouTubeHighlight = b
		}
	}
	if val, ok := updateData["youTubePinnedCommentAdded"]; ok {
		if b, ok := val.(bool); ok {
			video.YouTubeComment = b
		}
	}
	if val, ok := updateData["repliedToYouTubeComments"]; ok {
		if b, ok := val.(bool); ok {
			video.YouTubeCommentReply = b
		}
	}
	if val, ok := updateData["gdeAdvocuPostSent"]; ok {
		if b, ok := val.(bool); ok {
			video.GDE = b
		}
	}
	if val, ok := updateData["codeRepositoryURL"]; ok {
		if str, ok := val.(string); ok {
			video.Repo = str
		}
	}
	if val, ok := updateData["notifiedSponsors"]; ok {
		if b, ok := val.(bool); ok {
			video.NotifiedSponsors = b
		}
	}

	// Update completion counts
	s.updatePostPublishCompletion(video)
	return nil
}

// Completion calculation helpers
func (s *Server) updateInitialDetailsCompletion(video *storage.Video) {
	var completedCount, totalCount int

	// General fields
	generalFields := []interface{}{
		video.ProjectName,
		video.ProjectURL,
		video.Gist,
		video.Date,
	}
	c, t := s.countCompletedTasks(generalFields)
	completedCount += c
	totalCount += t

	// Sponsorship.Amount
	totalCount++
	if len(video.Sponsorship.Amount) > 0 {
		completedCount++
	}

	// Special conditions
	totalCount += 3
	
	// Condition 1: Sponsorship Emails
	if len(video.Sponsorship.Amount) == 0 || video.Sponsorship.Amount == "N/A" || video.Sponsorship.Amount == "-" || len(video.Sponsorship.Emails) > 0 {
		completedCount++
	}
	
	// Condition 2: Sponsorship Blocked
	if len(video.Sponsorship.Blocked) == 0 {
		completedCount++
	}
	
	// Condition 3: Delayed
	if !video.Delayed {
		completedCount++
	}

	video.Init.Completed = completedCount
	video.Init.Total = totalCount
}

func (s *Server) updateWorkProgressCompletion(video *storage.Video) {
	fields := []interface{}{
		video.Code,
		video.Head,
		video.Screen,
		video.RelatedVideos,
		video.Thumbnails,
		video.Diagrams,
		video.Screenshots,
		video.Location,
		video.Tagline,
		video.TaglineIdeas,
		video.OtherLogos,
	}
	video.Work.Completed, video.Work.Total = s.countCompletedTasks(fields)
}

func (s *Server) updateDefinitionCompletion(video *storage.Video) {
	fields := []interface{}{
		video.Title,
		video.Description,
		video.Highlight,
		video.Tags,
		video.DescriptionTags,
		video.Tweet,
		video.Animations,
		video.RequestThumbnail,
		video.Gist,
	}
	video.Define.Completed, video.Define.Total = s.countCompletedTasks(fields)
}

func (s *Server) updatePostProductionCompletion(video *storage.Video) {
	fields := []interface{}{
		video.Thumbnail,
		video.Members,
		video.RequestEdit,
		video.Movie,
		video.Slides,
	}
	video.Edit.Completed, video.Edit.Total = s.countCompletedTasks(fields)
	
	// Special handling for Timecodes
	video.Edit.Total++
	if video.Timecodes != "" && !containsString(video.Timecodes, "FIXME:") {
		video.Edit.Completed++
	}
}

func (s *Server) updatePublishingCompletion(video *storage.Video) {
	fields := []interface{}{
		video.UploadVideo,
		video.HugoPath,
	}
	video.Publish.Completed, video.Publish.Total = s.countCompletedTasks(fields)
}

func (s *Server) updatePostPublishCompletion(video *storage.Video) {
	fields := []interface{}{
		video.BlueSkyPosted,
		video.LinkedInPosted,
		video.SlackPosted,
		video.YouTubeHighlight,
		video.YouTubeComment,
		video.YouTubeCommentReply,
		video.GDE,
		video.Repo,
	}
	video.PostPublish.Completed, video.PostPublish.Total = s.countCompletedTasks(fields)
	
	// Special logic for NotifiedSponsors
	video.PostPublish.Total++
	if video.NotifiedSponsors || len(video.Sponsorship.Amount) == 0 || video.Sponsorship.Amount == "N/A" || video.Sponsorship.Amount == "-" {
		video.PostPublish.Completed++
	}
}

// countCompletedTasks counts completed tasks based on field values
func (s *Server) countCompletedTasks(fields []interface{}) (completed int, total int) {
	for _, field := range fields {
		valueType := reflect.TypeOf(field)
		if valueType == nil {
			total++
			continue
		}
		switch valueType.Kind() {
		case reflect.String:
			if len(field.(string)) > 0 && field.(string) != "-" {
				completed++
			}
		case reflect.Bool:
			if field.(bool) {
				completed++
			}
		case reflect.Slice:
			if reflect.ValueOf(field).Len() > 0 {
				completed++
			}
		}
		total++
	}
	return completed, total
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || func() bool {
		for i := 0; i <= len(s)-len(substr); i++ {
			if s[i:i+len(substr)] == substr {
				return true
			}
		}
		return false
	}())
}