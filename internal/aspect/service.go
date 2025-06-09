package aspect

// Service provides access to editing aspect metadata
type Service struct{}

// NewService creates a new aspect service
func NewService() *Service {
	return &Service{}
}

// GetAspects returns complete aspect metadata including all fields
// This method is kept for backward compatibility if needed
func (s *Service) GetAspects() AspectMetadata {
	mappings := GetVideoAspectMappings()
	aspects := make([]Aspect, len(mappings))

	for i, mapping := range mappings {
		fields := make([]Field, len(mapping.Fields))
		for j, fieldMapping := range mapping.Fields {
			fields[j] = Field{
				Name:            fieldMapping.Title,
				Type:            fieldMapping.FieldType,
				Required:        fieldMapping.Required,
				Order:           fieldMapping.Order,
				Description:     getFieldDescription(fieldMapping.FieldKey),
				Options:         FieldOptions{Values: fieldMapping.Options},
				UIHints:         fieldMapping.UIHints,
				ValidationHints: fieldMapping.ValidationHints,
				DefaultValue:    fieldMapping.DefaultValue,
			}
		}

		aspects[i] = Aspect{
			Key:         mapping.AspectKey,
			Title:       mapping.Title,
			Description: mapping.Description,
			Endpoint:    getEndpointForAspect(mapping.AspectKey),
			Icon:        getIconForAspect(mapping.AspectKey),
			Order:       i + 1,
			Fields:      fields,
		}
	}

	return AspectMetadata{
		Aspects: aspects,
	}
}

// GetAspectsOverview returns lightweight aspect metadata without fields
func (s *Service) GetAspectsOverview() AspectOverview {
	mappings := GetVideoAspectMappings()
	aspects := make([]AspectSummary, len(mappings))

	for i, mapping := range mappings {
		aspects[i] = AspectSummary{
			Key:                 mapping.AspectKey,
			Title:               mapping.Title,
			Description:         mapping.Description,
			Endpoint:            getEndpointForAspect(mapping.AspectKey),
			Icon:                getIconForAspect(mapping.AspectKey),
			Order:               i + 1,
			FieldCount:          len(mapping.Fields),
			CompletedFieldCount: 0, // Will be calculated in handler
		}
	}

	return AspectOverview{
		Aspects: aspects,
	}
}

// GetAspectFields returns detailed field information for a specific aspect
func (s *Service) GetAspectFields(aspectKey string) (*AspectFields, error) {
	mappings := GetVideoAspectMappings()

	for _, mapping := range mappings {
		if mapping.AspectKey == aspectKey {
			fields := make([]Field, len(mapping.Fields))
			for i, fieldMapping := range mapping.Fields {
				fields[i] = Field{
					Name:            fieldMapping.Title,
					Type:            fieldMapping.FieldType,
					Required:        fieldMapping.Required,
					Order:           fieldMapping.Order,
					Description:     getFieldDescription(fieldMapping.FieldKey),
					Options:         FieldOptions{Values: fieldMapping.Options},
					UIHints:         fieldMapping.UIHints,
					ValidationHints: fieldMapping.ValidationHints,
					DefaultValue:    fieldMapping.DefaultValue,
				}
			}

			return &AspectFields{
				AspectKey:   aspectKey,
				AspectTitle: mapping.Title,
				Fields:      fields,
			}, nil
		}
	}

	return nil, ErrAspectNotFound
}

// Helper functions for aspect metadata

func getEndpointForAspect(aspectKey string) string {
	endpointMap := map[string]string{
		AspectKeyInitialDetails: "/api/videos/{videoName}/initial-details",
		AspectKeyWorkProgress:   "/api/videos/{videoName}/work-progress",
		AspectKeyDefinition:     "/api/videos/{videoName}/definition",
		AspectKeyPostProduction: "/api/videos/{videoName}/post-production",
		AspectKeyPublishing:     "/api/videos/{videoName}/publishing",
		AspectKeyPostPublish:    "/api/videos/{videoName}/post-publish",
	}
	return endpointMap[aspectKey]
}

func getIconForAspect(aspectKey string) string {
	iconMap := map[string]string{
		AspectKeyInitialDetails: "info",
		AspectKeyWorkProgress:   "video",
		AspectKeyDefinition:     "edit",
		AspectKeyPostProduction: "scissors",
		AspectKeyPublishing:     "upload",
		AspectKeyPostPublish:    "share",
	}
	return iconMap[aspectKey]
}

func getFieldDescription(fieldKey string) string {
	descriptionMap := map[string]string{
		// Initial Details
		"projectName":       "Name of the related project",
		"projectURL":        "URL to the project repository or documentation",
		"publishDate":       "Scheduled publication date and time",
		"gistPath":          "Path to the manuscript/gist file",
		"sponsorshipAmount": "Sponsorship amount if applicable",
		"sponsorshipEmails": "Sponsor contact emails",
		"delayed":           "Whether the video is delayed",

		// Work Progress
		"codeDone":            "Code/demonstration completed",
		"talkingHeadDone":     "Talking head video recorded",
		"screenRecordingDone": "Screen recording completed",
		"relatedVideos":       "List of related videos for reference",
		"thumbnailsDone":      "Thumbnail images prepared",
		"diagramsDone":        "Diagrams and visual aids created",
		"screenshotsDone":     "Screenshots captured",
		"filesLocation":       "File storage location or Google Drive link",
		"tagline":             "Video tagline or subtitle",
		"taglineIdeas":        "Alternative tagline options",
		"otherLogos":          "Additional logos or assets needed",

		// Definition
		"title":            "Video title",
		"description":      "Video description text",
		"highlight":        "Key highlight or main point",
		"tags":             "Video tags for categorization",
		"descriptionTags":  "Tags for video description",
		"tweet":            "Social media tweet text",
		"animationsScript": "Animation instructions or script",
		"requestThumbnail": "Request custom thumbnail creation",

		// Post-Production
		"thumbnailPath": "Path to thumbnail image file",
		"members":       "Team members involved",
		"editRequest":   "Special editing requests or notes",
		"timecodes":     "Important timestamp markers",
		"movieDone":     "Video editing completed",
		"slidesDone":    "Presentation slides finalized",

		// Publishing
		"videoFilePath":   "Path to final video file",
		"uploadToYoutube": "Upload video to YouTube",
		"createHugoPost":  "Create blog post with Hugo",

		// Post-Publish
		"devOpstoolkitPosted":     "Posted to DevOpsToolkit",
		"blueskyPosted":           "Posted to BlueSky social media",
		"linkedinPosted":          "Posted to LinkedIn",
		"slackPosted":             "Posted to Slack channels",
		"youtubeHighlightCreated": "YouTube highlight reel created",
		"youtubePinnedComment":    "Pinned comment added to YouTube",
		"youtubeCommentsReplied":  "Replied to YouTube comments",
		"gdeAdvocuPosted":         "Posted to GDE Advocu",
		"codeRepositoryURL":       "Link to associated code repository",
		"notifySponsors":          "Notify sponsors of publication",
	}

	description, exists := descriptionMap[fieldKey]
	if !exists {
		return "Field description"
	}
	return description
}
