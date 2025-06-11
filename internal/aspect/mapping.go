package aspect

import (
	"devopstoolkit/youtube-automation/internal/constants"
	"devopstoolkit/youtube-automation/internal/storage"
)

// FieldMapping defines how a Video property maps to an aspect field
type FieldMapping struct {
	VideoProperty string   `json:"videoProperty"`     // The property name in storage.Video
	FieldKey      string   `json:"fieldKey"`          // The field key in the aspect
	FieldType     string   `json:"fieldType"`         // The field type (string, boolean, etc.)
	Title         string   `json:"title"`             // Display title (reused from CLI constants)
	Required      bool     `json:"required"`          // Whether the field is required
	Order         int      `json:"order"`             // Display order (matches CLI form order)
	Options       []string `json:"options,omitempty"` // Options for select fields

	// Enhanced metadata for frontend rendering (Task 3)
	UIHints         UIHints         `json:"uiHints"`         // UI rendering hints
	ValidationHints ValidationHints `json:"validationHints"` // Validation rule metadata
	DefaultValue    interface{}     `json:"defaultValue"`    // Default value for the field
}

// AspectMapping defines the complete mapping configuration for an editing aspect
type AspectMapping struct {
	AspectKey   string         `json:"aspectKey"`   // The aspect identifier
	Title       string         `json:"title"`       // Display title for the aspect
	Description string         `json:"description"` // Description of the aspect
	Fields      []FieldMapping `json:"fields"`      // Field mappings for this aspect
}

// createFieldMapping creates a FieldMapping with enhanced metadata using the field type system
func createFieldMapping(videoProperty, fieldKey, fieldType, title string, required bool, order int, options []string) FieldMapping {
	// Create the appropriate field type instance
	var fieldTypeInstance FieldType

	switch fieldType {
	case FieldTypeString:
		fieldTypeInstance = &StringFieldType{}
	case FieldTypeText:
		fieldTypeInstance = &TextFieldType{}
	case FieldTypeBoolean:
		fieldTypeInstance = &BooleanFieldType{}
	case FieldTypeDate:
		fieldTypeInstance = &DateFieldType{}
	case FieldTypeNumber:
		fieldTypeInstance = &NumberFieldType{}
	case FieldTypeSelect:
		selectOptions := make([]SelectOption, len(options))
		for i, opt := range options {
			selectOptions[i] = SelectOption{Value: opt, Label: opt}
		}
		fieldTypeInstance = &SelectFieldType{Options: selectOptions}
	default:
		fieldTypeInstance = &StringFieldType{} // Default fallback
	}

	return FieldMapping{
		VideoProperty:   videoProperty,
		FieldKey:        fieldKey,
		FieldType:       fieldType,
		Title:           title,
		Required:        required,
		Order:           order,
		Options:         options,
		UIHints:         fieldTypeInstance.GetUIHints(),
		ValidationHints: fieldTypeInstance.GetValidationHints(),
		DefaultValue:    getDefaultValueForField(fieldType, required),
	}
}

// getDefaultValueForField returns appropriate default values based on field type
func getDefaultValueForField(fieldType string, required bool) interface{} {
	switch fieldType {
	case FieldTypeBoolean:
		return false
	case FieldTypeNumber:
		return 0
	case FieldTypeString, FieldTypeText, FieldTypeDate, FieldTypeSelect:
		if required {
			return ""
		}
		return nil
	default:
		return nil
	}
}

// GetVideoAspectMappings returns the complete mapping configuration
// This maps Video object properties to editing aspect fields
// Field orders match the exact order they appear in CLI forms
func GetVideoAspectMappings() []AspectMapping {
	return []AspectMapping{
		{
			AspectKey:   AspectKeyInitialDetails,
			Title:       constants.PhaseTitleInitialDetails,
			Description: "Initial video details and project information",
			Fields: []FieldMapping{
				createFieldMapping("ProjectName", "projectName", FieldTypeString, constants.FieldTitleProjectName, false, 1, nil),
				createFieldMapping("ProjectURL", "projectURL", FieldTypeString, constants.FieldTitleProjectURL, false, 2, nil),
				createFieldMapping("Sponsorship.Amount", "sponsorshipAmount", FieldTypeString, constants.FieldTitleSponsorshipAmount, false, 3, nil),
				createFieldMapping("Sponsorship.Emails", "sponsorshipEmails", FieldTypeString, constants.FieldTitleSponsorshipEmails, false, 4, nil),
				createFieldMapping("Sponsorship.Blocked", "sponsorshipBlocked", FieldTypeString, constants.FieldTitleSponsorshipBlocked, false, 5, nil),
				createFieldMapping("Date", "publishDate", FieldTypeDate, constants.FieldTitlePublishDate, false, 6, nil),
				createFieldMapping("Delayed", "delayed", FieldTypeBoolean, constants.FieldTitleDelayed, false, 7, nil),
				createFieldMapping("Gist", "gistPath", FieldTypeString, constants.FieldTitleGistPath, false, 8, nil),
			},
		},
		{
			AspectKey:   AspectKeyWorkProgress,
			Title:       constants.PhaseTitleWorkProgress,
			Description: "Work progress and content creation status",
			Fields: []FieldMapping{
				createFieldMapping("Code", "codeDone", FieldTypeBoolean, constants.FieldTitleCodeDone, false, 1, nil),
				createFieldMapping("Head", "talkingHeadDone", FieldTypeBoolean, constants.FieldTitleTalkingHeadDone, false, 2, nil),
				createFieldMapping("Screen", "screenRecordingDone", FieldTypeBoolean, constants.FieldTitleScreenRecordingDone, false, 3, nil),
				createFieldMapping("RelatedVideos", "relatedVideos", FieldTypeText, constants.FieldTitleRelatedVideos, false, 4, nil),
				createFieldMapping("Thumbnails", "thumbnailsDone", FieldTypeBoolean, constants.FieldTitleThumbnailsDone, false, 5, nil),
				createFieldMapping("Diagrams", "diagramsDone", FieldTypeBoolean, constants.FieldTitleDiagramsDone, false, 6, nil),
				createFieldMapping("Screenshots", "screenshotsDone", FieldTypeBoolean, constants.FieldTitleScreenshotsDone, false, 7, nil),
				createFieldMapping("Location", "filesLocation", FieldTypeString, constants.FieldTitleFilesLocation, false, 8, nil),
				createFieldMapping("Tagline", "tagline", FieldTypeString, constants.FieldTitleTagline, false, 9, nil),
				createFieldMapping("TaglineIdeas", "taglineIdeas", FieldTypeText, constants.FieldTitleTaglineIdeas, false, 10, nil),
				createFieldMapping("OtherLogos", "otherLogos", FieldTypeString, constants.FieldTitleOtherLogos, false, 11, nil),
			},
		},
		{
			AspectKey:   AspectKeyDefinition,
			Title:       constants.PhaseTitleDefinition,
			Description: "Video content definition and metadata",
			Fields: []FieldMapping{
				createFieldMapping("Title", "title", FieldTypeString, constants.FieldTitleTitle, true, 1, nil),
				createFieldMapping("Description", "description", FieldTypeText, constants.FieldTitleDescription, false, 2, nil),
				createFieldMapping("Highlight", "highlight", FieldTypeString, constants.FieldTitleHighlight, false, 3, nil),
				createFieldMapping("Tags", "tags", FieldTypeString, constants.FieldTitleTags, false, 4, nil),
				createFieldMapping("DescriptionTags", "descriptionTags", FieldTypeText, constants.FieldTitleDescriptionTags, false, 5, nil),
				createFieldMapping("Tweet", "tweet", FieldTypeString, constants.FieldTitleTweet, false, 6, nil),
				createFieldMapping("Animations", "animationsScript", FieldTypeText, constants.FieldTitleAnimationsScript, false, 7, nil),
			},
		},
		{
			AspectKey:   AspectKeyPostProduction,
			Title:       constants.PhaseTitlePostProduction,
			Description: "Post-production editing and review tasks",
			Fields: []FieldMapping{
				createFieldMapping("Thumbnail", "thumbnailPath", FieldTypeString, constants.FieldTitleThumbnailPath, false, 1, nil),
				createFieldMapping("Members", "members", FieldTypeString, constants.FieldTitleMembers, false, 2, nil),
				createFieldMapping("RequestEdit", "requestEdit", FieldTypeBoolean, constants.FieldTitleRequestEdit, false, 3, nil),
				createFieldMapping("Timecodes", "timecodes", FieldTypeText, constants.FieldTitleTimecodes, false, 4, nil),
				createFieldMapping("Movie", "movieDone", FieldTypeBoolean, constants.FieldTitleMovieDone, false, 5, nil),
				createFieldMapping("Slides", "slidesDone", FieldTypeBoolean, constants.FieldTitleSlidesDone, false, 6, nil),
			},
		},
		{
			AspectKey:   AspectKeyPublishing,
			Title:       constants.PhaseTitlePublishingDetails,
			Description: "Publishing settings and video upload",
			Fields: []FieldMapping{
				createFieldMapping("UploadVideo", "videoFilePath", FieldTypeString, constants.FieldTitleVideoFilePath, false, 1, nil),
				createFieldMapping("VideoId", "youTubeVideoId", FieldTypeString, constants.FieldTitleCurrentVideoID, false, 2, nil),
				createFieldMapping("HugoPath", "hugoPostPath", FieldTypeString, constants.FieldTitleCreateHugo, false, 3, nil),
			},
		},
		{
			AspectKey:   AspectKeyPostPublish,
			Title:       constants.PhaseTitlePostPublish,
			Description: "Post-publication tasks and social media",
			Fields: []FieldMapping{
				createFieldMapping("DOTPosted", "dotPosted", FieldTypeBoolean, constants.FieldTitleDOTPosted, false, 1, nil),
				createFieldMapping("BlueSkyPosted", "blueSkyPosted", FieldTypeBoolean, constants.FieldTitleBlueSkyPosted, false, 2, nil),
				createFieldMapping("LinkedInPosted", "linkedInPosted", FieldTypeBoolean, constants.FieldTitleLinkedInPosted, false, 3, nil),
				createFieldMapping("SlackPosted", "slackPosted", FieldTypeBoolean, constants.FieldTitleSlackPosted, false, 4, nil),
				createFieldMapping("YouTubeHighlight", "youTubeHighlight", FieldTypeBoolean, constants.FieldTitleYouTubeHighlight, false, 5, nil),
				createFieldMapping("YouTubeComment", "youTubeComment", FieldTypeBoolean, constants.FieldTitleYouTubeComment, false, 6, nil),
				createFieldMapping("YouTubeCommentReply", "youTubeCommentReply", FieldTypeBoolean, constants.FieldTitleYouTubeCommentReply, false, 7, nil),
				createFieldMapping("GDE", "gdePosted", FieldTypeBoolean, constants.FieldTitleGDEPosted, false, 8, nil),
				createFieldMapping("Repo", "codeRepository", FieldTypeString, constants.FieldTitleCodeRepository, false, 9, nil),
				createFieldMapping("NotifiedSponsors", "notifySponsors", FieldTypeBoolean, constants.FieldTitleNotifySponsors, false, 10, nil),
			},
		},
	}
}

// GetVideoPropertyValue extracts a property value from a Video object using the property path
func GetVideoPropertyValue(video storage.Video, propertyPath string) interface{} {
	switch propertyPath {
	// Initial Details
	case "ProjectName":
		return video.ProjectName
	case "ProjectURL":
		return video.ProjectURL
	case "Sponsorship.Amount":
		return video.Sponsorship.Amount
	case "Sponsorship.Emails":
		return video.Sponsorship.Emails
	case "Sponsorship.Blocked":
		return video.Sponsorship.Blocked
	case "Date":
		return video.Date
	case "Delayed":
		return video.Delayed
	case "Gist":
		return video.Gist

	// Work Progress
	case "Code":
		return video.Code
	case "Head":
		return video.Head
	case "Screen":
		return video.Screen
	case "RelatedVideos":
		return video.RelatedVideos
	case "Thumbnails":
		return video.Thumbnails
	case "Diagrams":
		return video.Diagrams
	case "Screenshots":
		return video.Screenshots
	case "Location":
		return video.Location
	case "Tagline":
		return video.Tagline
	case "TaglineIdeas":
		return video.TaglineIdeas
	case "OtherLogos":
		return video.OtherLogos

	// Definition
	case "Title":
		return video.Title
	case "Description":
		return video.Description
	case "Highlight":
		return video.Highlight
	case "Tags":
		return video.Tags
	case "DescriptionTags":
		return video.DescriptionTags
	case "Tweet":
		return video.Tweet
	case "Animations":
		return video.Animations

	// Post-Production
	case "Thumbnail":
		return video.Thumbnail
	case "Members":
		return video.Members
	case "RequestEdit":
		return video.RequestEdit
	case "Timecodes":
		return video.Timecodes
	case "Movie":
		return video.Movie
	case "Slides":
		return video.Slides

	// Publishing
	case "UploadVideo":
		return video.UploadVideo
	case "VideoId":
		return video.VideoId
	case "HugoPath":
		return video.HugoPath

	// Post-Publish
	case "DOTPosted":
		return video.DOTPosted
	case "BlueSkyPosted":
		return video.BlueSkyPosted
	case "LinkedInPosted":
		return video.LinkedInPosted
	case "SlackPosted":
		return video.SlackPosted
	case "YouTubeHighlight":
		return video.YouTubeHighlight
	case "YouTubeComment":
		return video.YouTubeComment
	case "YouTubeCommentReply":
		return video.YouTubeCommentReply
	case "GDE":
		return video.GDE
	case "Repo":
		return video.Repo
	case "NotifiedSponsors":
		return video.NotifiedSponsors

	default:
		return nil
	}
}
