package app

import (
	"reflect"
	"strings"

	"devopstoolkit/youtube-automation/internal/constants"
)

// Helper function to count completed tasks based on old logic
func (m *MenuHandler) countCompletedTasks(fields []interface{}) (completed int, total int) {
	for _, field := range fields {
		valueType := reflect.TypeOf(field)
		if valueType == nil { // Handle cases where a field might be nil unexpectedly
			total++
			continue
		}
		switch valueType.Kind() {
		case reflect.String:
			if len(field.(string)) > 0 && field.(string) != "-" { // Field is complete if not empty and not just a dash
				completed++
			}
		case reflect.Bool:
			if field.(bool) {
				completed++
			}
		case reflect.Slice:
			// Assuming non-empty slice means task related to it is done
			if reflect.ValueOf(field).Len() > 0 {
				completed++
			}
		}
		total++
	}
	return completed, total
}

// Helper function that uses shared completion criteria logic from aspect service
func (m *MenuHandler) colorTitleWithSharedLogic(aspectKey, fieldTitle, fieldValue string, boolValue *bool, sponsorshipAmount string) string {
	// Map field titles to their corresponding field keys used in the completion service
	fieldKey := m.getFieldKeyFromTitle(fieldTitle)

	// Get completion criteria from the shared service
	completionCriteria := m.aspectService.GetFieldCompletionCriteria(aspectKey, fieldKey)

	// Apply completion logic based on the criteria
	isComplete := false

	switch completionCriteria {
	case "filled_only":
		// Complete when not empty and not "-"
		isComplete = len(fieldValue) > 0 && fieldValue != "-"

	case "filled_required":
		// Must be filled (similar to filled_only for now)
		isComplete = len(fieldValue) > 0 && fieldValue != "-"

	case "empty_or_filled":
		// Complete when empty OR filled (always green)
		isComplete = true

	case "true_only":
		// Complete when boolean is true
		if boolValue != nil {
			isComplete = *boolValue
		}

	case "false_only":
		// Complete when boolean is false
		if boolValue != nil {
			isComplete = !(*boolValue)
		}

	case "conditional_sponsorship":
		// Special logic for sponsorship emails field
		if fieldKey == "sponsorshipEmails" {
			// Complete if sponsorship amount is empty/N/A/- OR emails have content
			isComplete = (len(sponsorshipAmount) == 0 || sponsorshipAmount == "N/A" || sponsorshipAmount == "-") || len(fieldValue) > 0
		} else {
			// For other conditional fields, default to filled_only logic
			isComplete = len(fieldValue) > 0 && fieldValue != "-"
		}

	case "conditional_sponsors":
		// Special logic for notify sponsors field
		if fieldKey == "notifySponsors" || fieldKey == "notifiedSponsors" {
			// Complete if sponsorship amount is empty/N/A/- OR notification is done
			isComplete = (len(sponsorshipAmount) == 0 || sponsorshipAmount == "N/A" || sponsorshipAmount == "-")
			if !isComplete && boolValue != nil {
				isComplete = *boolValue
			}
		} else {
			// For other conditional fields, default to filled_only logic
			isComplete = len(fieldValue) > 0 && fieldValue != "-"
		}

	case "no_fixme":
		// Complete when content doesn't contain FIXME
		isComplete = len(fieldValue) > 0 && !strings.Contains(fieldValue, "FIXME")

	default:
		// Default to filled_only logic
		isComplete = len(fieldValue) > 0 && fieldValue != "-"
	}

	if isComplete {
		return m.greenStyle.Render(fieldTitle)
	}
	return m.orangeStyle.Render(fieldTitle)
}

// getFieldKeyFromTitle maps field titles to their corresponding field keys used in completion service
func (m *MenuHandler) getFieldKeyFromTitle(fieldTitle string) string {
	titleToKeyMap := map[string]string{
		// Initial Details
		constants.FieldTitleProjectName:        "projectName",
		constants.FieldTitleProjectURL:         "projectURL",
		constants.FieldTitleSponsorshipAmount:  "sponsorshipAmount",
		constants.FieldTitleSponsorshipEmails:  "sponsorshipEmails",
		constants.FieldTitleSponsorshipBlocked: "sponsorshipBlockedReason",
		constants.FieldTitleSponsorshipName:    "sponsorshipName",
		constants.FieldTitleSponsorshipURL:     "sponsorshipURL",
		constants.FieldTitlePublishDate:        "date",
		constants.FieldTitleDelayed:            "delayed",
		constants.FieldTitleGistPath:           "gist",

		// Work Progress
		constants.FieldTitleCodeDone:            "code",
		constants.FieldTitleTalkingHeadDone:     "head",
		constants.FieldTitleScreenRecordingDone: "screen",
		constants.FieldTitleRelatedVideos:       "relatedVideos",
		constants.FieldTitleThumbnailsDone:      "thumbnails",
		constants.FieldTitleDiagramsDone:        "diagrams",
		constants.FieldTitleScreenshotsDone:     "screenshots",
		constants.FieldTitleFilesLocation:       "location",
		constants.FieldTitleTagline:             "tagline",
		constants.FieldTitleTaglineIdeas:        "taglineIdeas",
		constants.FieldTitleOtherLogos:          "otherLogos",

		// Definition
		constants.FieldTitleTitle:            "title",
		constants.FieldTitleDescription:      "description",
		constants.FieldTitleTags:             "tags",
		constants.FieldTitleDescriptionTags:  "descriptionTags",
		constants.FieldTitleTweet:            "tweet",
		constants.FieldTitleAnimationsScript: "animationsScript",

		// Post-Production
		constants.FieldTitleThumbnailPath: "thumbnailPath",
		constants.FieldTitleMembers:       "members",
		constants.FieldTitleRequestEdit:   "requestEdit",
		constants.FieldTitleTimecodes:     "timecodes",
		constants.FieldTitleMovieDone:     "movieDone",
		constants.FieldTitleSlidesDone:    "slidesDone",

		// Publishing
		constants.FieldTitleVideoFilePath:  "videoFilePath",
		constants.FieldTitleCurrentVideoID: "youTubeVideoId",
		constants.FieldTitleCreateHugo:     "hugoPostPath",

		// Post-Publish
		constants.FieldTitleDOTPosted:           "dotPosted",
		constants.FieldTitleBlueSkyPosted:       "blueSkyPosted",
		constants.FieldTitleLinkedInPosted:      "linkedInPosted",
		constants.FieldTitleSlackPosted:         "slackPosted",
		constants.FieldTitleYouTubeHighlight:    "youTubeHighlight",
		constants.FieldTitleYouTubeComment:      "youTubeComment",
		constants.FieldTitleYouTubeCommentReply: "youTubeCommentReply",
		constants.FieldTitleGDEPosted:           "gdePosted",
		constants.FieldTitleCodeRepository:      "codeRepository",
		constants.FieldTitleNotifySponsors:      "notifySponsors",
	}

	if fieldKey, exists := titleToKeyMap[fieldTitle]; exists {
		return fieldKey
	}

	// If no mapping found, return the title as-is as fallback
	return fieldTitle
}

// Helper for form titles based on string value - now uses shared logic
func (m *MenuHandler) colorTitleString(title, value string) string {
	return m.colorTitleWithSharedLogic("initial-details", title, value, nil, "")
}

// Helper for form titles based on boolean value - now uses shared logic
func (m *MenuHandler) colorTitleBool(title string, value bool) string {
	return m.colorTitleWithSharedLogic("work-progress", title, "", &value, "")
}

// Helper for form titles for Sponsorship Amount - now uses shared logic
func (m *MenuHandler) colorTitleSponsorshipAmount(title, value string) string {
	return m.colorTitleWithSharedLogic("initial-details", title, value, nil, "")
}

// Helper for form titles for sponsored emails - now uses shared logic
func (m *MenuHandler) colorTitleSponsoredEmails(title, sponsoredAmount, sponsoredEmails string) string {
	return m.colorTitleWithSharedLogic("initial-details", title, sponsoredEmails, nil, sponsoredAmount)
}

// Helper for form titles based on string value (inverse logic) - now uses shared logic
func (m *MenuHandler) colorTitleStringInverse(title, value string) string {
	return m.colorTitleWithSharedLogic("initial-details", title, value, nil, "")
}

// Helper for form titles based on boolean value (inverse logic) - now uses shared logic
func (m *MenuHandler) colorTitleBoolInverse(title string, value bool) string {
	return m.colorTitleWithSharedLogic("initial-details", title, "", &value, "")
}
