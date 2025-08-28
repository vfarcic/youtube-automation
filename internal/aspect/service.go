package aspect

import (
	"sort"
)

// Service provides access to editing aspect metadata
type Service struct {
	completionService *CompletionService
}

// NewService creates a new aspect service
func NewService() *Service {
	return &Service{
		completionService: NewCompletionService(),
	}
}

// GetAspects returns complete aspect metadata including all fields
// This method is kept for backward compatibility if needed
func (s *Service) GetAspects() AspectMetadata {
	mappings := GetVideoAspectMappings()
	aspects := make([]Aspect, len(mappings))

	for i, mapping := range mappings {
		fields := make([]Field, len(mapping.Fields))
		for j, fieldMapping := range mapping.Fields {
			completionFieldName := mapFieldNameForCompletion(fieldMapping.FieldName)
			fields[j] = Field{
				Name:               fieldMapping.Title,
				FieldName:          fieldMapping.FieldName,
				Type:               fieldMapping.FieldType,
				Required:           fieldMapping.Required,
				Order:              fieldMapping.Order,
				Description:        getFieldDescription(fieldMapping.FieldName),
				Options:            FieldOptions{Values: fieldMapping.Options},
				UIHints:            fieldMapping.UIHints,
				ValidationHints:    fieldMapping.ValidationHints,
				DefaultValue:       fieldMapping.DefaultValue,
				CompletionCriteria: s.completionService.GetFieldCompletionCriteria(mapping.AspectKey, completionFieldName),
			}
		}

		aspects[i] = Aspect{
			Key:         mapping.AspectKey,
			Title:       mapping.Title,
			Description: mapping.Description,
			Endpoint:    getEndpointForAspect(mapping.AspectKey),
			Icon:        getIconForAspect(mapping.AspectKey),
			Order:       mapping.Order,
			Fields:      fields,
		}
	}

	// Sort aspects by order
	sort.Slice(aspects, func(i, j int) bool {
		return aspects[i].Order < aspects[j].Order
	})

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
			Order:               mapping.Order,
			FieldCount:          len(mapping.Fields),
			CompletedFieldCount: 0, // Will be calculated in handler
		}
	}

	// Sort aspects by order
	sort.Slice(aspects, func(i, j int) bool {
		return aspects[i].Order < aspects[j].Order
	})

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
				completionFieldName := mapFieldNameForCompletion(fieldMapping.FieldName)
				fields[i] = Field{
					Name:               fieldMapping.Title,
					FieldName:          fieldMapping.FieldName,
					Type:               fieldMapping.FieldType,
					Required:           fieldMapping.Required,
					Order:              fieldMapping.Order,
					Description:        getFieldDescription(fieldMapping.FieldName),
					Options:            FieldOptions{Values: fieldMapping.Options},
					UIHints:            fieldMapping.UIHints,
					ValidationHints:    fieldMapping.ValidationHints,
					DefaultValue:       fieldMapping.DefaultValue,
					CompletionCriteria: s.completionService.GetFieldCompletionCriteria(mapping.AspectKey, completionFieldName),
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

// GetFieldCompletionCriteria returns completion criteria for a specific field
func (s *Service) GetFieldCompletionCriteria(aspectKey, fieldKey string) string {
	return s.completionService.GetFieldCompletionCriteria(aspectKey, fieldKey)
}

// mapFieldNameForCompletion maps JSON field names to completion service field names
func mapFieldNameForCompletion(jsonFieldName string) string {
	// Map JSON field names to completion service field names
	mappings := map[string]string{
		"sponsorship.amount":  "sponsorshipAmount",
		"sponsorship.emails":  "sponsorshipEmails",
		"sponsorship.blocked": "sponsorshipBlockedReason",
		"gist":                "gist",
		"code":                "code",
		"head":                "head",
		"screen":              "screen",
		"thumbnails":          "thumbnails",
		"diagrams":            "diagrams",
		"screenshots":         "screenshots",
		"location":            "location",
		"otherLogos":          "otherLogos",
		"requestThumbnail":    "requestThumbnail",
		"thumbnail":           "thumbnail",
		"movie":               "movie",
		"slides":              "slides",
		"uploadVideo":         "uploadVideo",
		"videoId":             "videoId",
		"hugoPath":            "hugoPath",
		"dotPosted":           "dotPosted",
		"blueSkyPosted":       "blueSkyPosted",
		"linkedInPosted":      "linkedInPosted",
		"slackPosted":         "slackPosted",
		"youTubeHighlight":    "youTubeHighlight",
		"youTubeComment":      "youTubeComment",
		"youTubeCommentReply": "youTubeCommentReply",
		"gde":                 "gde",
		"repo":                "repo",
		"notifiedSponsors":    "notifiedSponsors",
	}

	if mapped, exists := mappings[jsonFieldName]; exists {
		return mapped
	}

	// For fields that don't need mapping, return as-is
	return jsonFieldName
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
		"projectName":              "Name of the related project",
		"projectURL":               "URL to the project repository or documentation",
		"date":                     "Scheduled publication date and time",
		"gist":                     "Path to the manuscript/gist file",
		"sponsorshipAmount":        "Sponsorship amount if applicable",
		"sponsorshipEmails":        "Sponsor contact emails",
		"sponsorshipBlockedReason": "Reason for sponsorship blocking if applicable",
		"delayed":                  "Whether the video is delayed",

		// Work Progress
		"code":          "Code/demonstration completed",
		"head":          "Talking head video recorded",
		"screen":        "Screen recording completed",
		"relatedVideos": "List of related videos for reference",
		"thumbnails":    "Thumbnail images prepared",
		"diagrams":      "Diagrams and visual aids created",
		"screenshots":   "Screenshots captured",
		"location":      "File storage location or Google Drive link",
		"tagline":       "Video tagline or subtitle",
		"taglineIdeas":  "Alternative tagline options",
		"otherLogos":    "Additional logos or assets needed",

		// Definition
		"title":            "Video title",
		"description":      "Video description text",
		"tags":             "Video tags for categorization",
		"descriptionTags":  "Tags for video description",
		"tweet":            "Social media tweet text",
		"animationsScript": "Animation instructions or script",
		"requestThumbnail": "Request custom thumbnail creation",

		// Post-Production
		"thumbnailPath": "Path to thumbnail image file",
		"members":       "Team members involved",
		"requestEdit":   "Special editing requests or notes",
		"timecodes":     "Important timestamp markers",
		"movieDone":     "Video editing completed",
		"slidesDone":    "Presentation slides finalized",

		// Publishing
		"videoFilePath":  "Path to final video file",
		"youTubeVideoId": "YouTube video ID after upload",
		"hugoPostPath":   "Path to Hugo blog post",

		// Post-Publish
		"dotPosted":           "Posted to DevOpsToolkit",
		"blueSkyPosted":       "Posted to BlueSky social media",
		"linkedInPosted":      "Posted to LinkedIn",
		"slackPosted":         "Posted to Slack channels",
		"youTubeHighlight":    "YouTube highlight reel created",
		"youTubeComment":      "Pinned comment added to YouTube",
		"youTubeCommentReply": "Replied to YouTube comments",
		"gdePosted":           "Posted to GDE Advocu",
		"codeRepository":      "Link to associated code repository",
		"notifySponsors":      "Notify sponsors of publication",
	}

	description, exists := descriptionMap[fieldKey]
	if !exists {
		return "Field description"
	}
	return description
}
