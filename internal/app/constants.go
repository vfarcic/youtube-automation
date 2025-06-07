package app

// Phase title constants for display in UI
// These match the exact strings used in menu_handler.go
const (
	PhaseTitleInitialDetails    = "Initial Details"
	PhaseTitleWorkProgress      = "Work In Progress"
	PhaseTitleDefinition        = "Definition"
	PhaseTitlePostProduction    = "Post-Production"
	PhaseTitlePublishingDetails = "Publishing Details"
	PhaseTitlePostPublish       = "Post-Publish Details"
)

// Phase message constants for consistent messaging
const (
	// Edit cancelled messages
	MessageInitialDetailsEditCancelled = "Initial details edit cancelled."
	MessageWorkProgressEditCancelled   = "Work progress edit cancelled."
	MessagePostProductionEditCancelled = "Post-production edit cancelled."
	MessageDefinitionPhaseAborted      = "Definition phase aborted."

	// Error messages
	ErrorRunInitialDetailsForm     = "failed to run initial details edit form"
	ErrorRunWorkProgressForm       = "failed to run work progress edit form"
	ErrorRunPostProductionForm     = "failed to run post-production edit form"
	ErrorSaveInitialDetails        = "failed to save initial details"
	ErrorSaveWorkProgress          = "failed to save work progress"
	ErrorSavePostProductionDetails = "failed to save post-production details"
	ErrorDefinitionPhase           = "error during definition phase"

	// Success messages
	MessageInitialDetailsUpdated = "initial details updated"
	MessageWorkProgressUpdated   = "work progress updated"
	MessagePostProductionUpdated = "post-production details updated"

	// Changes not saved messages
	MessageChangesNotSavedInitialDetails = "Changes not saved for initial details."
	MessageChangesNotSavedWorkProgress   = "Changes not saved for work progress."
	MessageChangesNotSavedPostProduction = "Changes not saved for post-production."

	// Other messages
	MessageDefinitionPhaseComplete = "--- Definition Phase Complete ---"
)

// Field titles that match EXACTLY the CLI form titles
// These are used in menu_handler.go and should be reused by API metadata
// to eliminate duplication between CLI and API
const (
	// Work Progress Phase Fields
	FieldTitleCodeDone            = "Code Done"
	FieldTitleTalkingHeadDone     = "Talking Head Done"
	FieldTitleScreenRecordingDone = "Screen Recording Done"
	FieldTitleRelatedVideos       = "Related Videos (comma separated)"
	FieldTitleThumbnailsDone      = "Thumbnails Done"
	FieldTitleDiagramsDone        = "Diagrams Done"
	FieldTitleScreenshotsDone     = "Screenshots Done"
	FieldTitleFilesLocation       = "Files Location (e.g., Google Drive link)"
	FieldTitleTagline             = "Tagline"
	FieldTitleTaglineIdeas        = "Tagline Ideas"
	FieldTitleOtherLogos          = "Other Logos/Assets"

	// Post-Publish Phase Fields
	FieldTitleDOTPosted           = "DevOpsToolkit Post Sent (manual)"
	FieldTitleBlueSkyPosted       = "BlueSky Post Sent"
	FieldTitleLinkedInPosted      = "LinkedIn Post Sent (manual)"
	FieldTitleSlackPosted         = "Slack Post Sent"
	FieldTitleYouTubeHighlight    = "YouTube Highlight Created (manual)"
	FieldTitleYouTubeComment      = "YouTube Pinned Comment Added (manual)"
	FieldTitleYouTubeCommentReply = "Replied to YouTube Comments (manual)"
	FieldTitleGDEPosted           = "GDE Advocu Post Sent (manual)"
	FieldTitleCodeRepository      = "Code Repository URL"
	FieldTitleNotifySponsors      = "Notify Sponsors"

	// Post-Production Phase Fields
	FieldTitleThumbnailPath = "Thumbnail Path"
	FieldTitleMembers       = "Members (comma separated)"
	FieldTitleRequestEdit   = "Edit Request"
	FieldTitleTimecodes     = "Timecodes"
	FieldTitleMovieDone     = "Movie Done"
	FieldTitleSlidesDone    = "Slides Done"

	// Initial Details Phase Fields
	FieldTitleProjectName        = "Project Name"
	FieldTitleProjectURL         = "Project URL"
	FieldTitleSponsorshipAmount  = "Sponsorship Amount"
	FieldTitleSponsorshipEmails  = "Sponsorship Emails (comma separated)"
	FieldTitleSponsorshipBlocked = "Sponsorship Blocked Reason"
	FieldTitlePublishDate        = "Publish Date (YYYY-MM-DDTHH:MM)"
	FieldTitleDelayed            = "Delayed"
	FieldTitleGistPath           = "Gist Path (.md file)"

	// Publishing Phase Fields
	FieldTitleVideoFilePath   = "Video File Path"
	FieldTitleUploadToYouTube = "Upload Video to YouTube?"
	FieldTitleCurrentVideoID  = "Current YouTube Video ID"
	FieldTitleCreateHugo      = "Create/Update Hugo Post"

	// Definition Phase Fields
	FieldTitleTitle            = "Title"
	FieldTitleDescription      = "Description"
	FieldTitleHighlight        = "Highlight"
	FieldTitleTags             = "Tags"
	FieldTitleDescriptionTags  = "Description Tags"
	FieldTitleTweet            = "Tweet"
	FieldTitleAnimationsScript = "Animations Script"
)
