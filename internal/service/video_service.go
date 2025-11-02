package service

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"
	"devopstoolkit/youtube-automation/pkg/utils"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// VideoService provides unified data operations for videos across CLI and API
type VideoService struct {
	indexPath    string
	yamlStorage  *storage.YAML
	filesystem   *filesystem.Operations
	videoManager *video.Manager
}

// NewVideoService creates a new video service
func NewVideoService(indexPath string, filesystem *filesystem.Operations, videoManager *video.Manager) *VideoService {
	return &VideoService{
		indexPath:    indexPath,
		yamlStorage:  storage.NewYAML(indexPath),
		filesystem:   filesystem,
		videoManager: videoManager,
	}
}

// CreateVideo creates a new video entry
func (s *VideoService) CreateVideo(name, category, date string) (storage.VideoIndex, error) {
	if name == "" || category == "" {
		return storage.VideoIndex{}, fmt.Errorf("name and category are required")
	}

	// Sanitize the name to match the filename that will be created
	sanitizedName := s.filesystem.SanitizeName(name)

	vi := storage.VideoIndex{
		Name:     sanitizedName,
		Category: category,
	}

	// Create directory if it doesn't exist
	dirPath := s.filesystem.GetDirPath(vi.Category)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if mkDirErr := os.Mkdir(dirPath, 0755); mkDirErr != nil {
			return storage.VideoIndex{}, fmt.Errorf("failed to create directory %s: %w", dirPath, mkDirErr)
		}
	}

	// Create markdown script file with template content
	scriptContent := `## Intro

FIXME: Shock

FIXME: Establish expectations

FIXME: What's the ending?

## Setup

FIXME:

## FIXME:

FIXME:

## FIXME: Pros and Cons

FIXME: Header: Cons; Items: FIXME:

FIXME: Header: Pros; Items: FIXME:

## Destroy

FIXME:
`
	filePath := s.filesystem.GetFilePath(vi.Category, vi.Name, "md")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		f, errCreate := os.Create(filePath)
		if errCreate != nil {
			return storage.VideoIndex{}, fmt.Errorf("failed to create script file %s: %w", filePath, errCreate)
		}
		defer f.Close()
		if _, writeErr := f.Write([]byte(scriptContent)); writeErr != nil {
			return storage.VideoIndex{}, fmt.Errorf("failed to write to script file %s: %w", filePath, writeErr)
		}
	}

	// Create the default video YAML file
	videoPath := s.filesystem.GetFilePath(vi.Category, vi.Name, "yaml")

	defaultVideo := storage.Video{
		Name:     sanitizedName,
		Category: category,
		Path:     videoPath,
		// Initialize sponsorship
		Sponsorship: storage.Sponsorship{
			Amount:  "",
			Emails:  "",
			Blocked: "",
		},
		// Initialize other fields with default values
		Date:                date,
		Delayed:             false,
		Screen:              false,
		Head:                false,
		Thumbnails:          false,
		Diagrams:            false,
		Screenshots:         false,
		RequestThumbnail:    false,
		Movie:               false,
		Slides:              false,
		RequestEdit:         false,
		LinkedInPosted:      false,
		SlackPosted:         false,
		HNPosted:            false,
		DOTPosted:           false,
		BlueSkyPosted:       false,
		YouTubeHighlight:    false,
		YouTubeComment:      false,
		YouTubeCommentReply: false,
		GDE:                 false,
		NotifiedSponsors:    false,
		Code:                false,
	}

	// Write the video YAML file
	if err := s.yamlStorage.WriteVideo(defaultVideo, videoPath); err != nil {
		return storage.VideoIndex{}, fmt.Errorf("failed to create video file %s: %w", videoPath, err)
	}

	// Add to index
	index, err := s.yamlStorage.GetIndex()
	if err != nil {
		return storage.VideoIndex{}, fmt.Errorf("failed to get video index: %w", err)
	}

	index = append(index, vi)
	if err := s.yamlStorage.WriteIndex(index); err != nil {
		return storage.VideoIndex{}, fmt.Errorf("failed to write index: %w", err)
	}

	return vi, nil
}

// GetVideosByPhase returns videos in a specific phase
func (s *VideoService) GetVideosByPhase(phase int) ([]storage.Video, error) {
	index, err := s.yamlStorage.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get video index: %w", err)
	}

	var videosInPhase []storage.Video
	for _, videoIndex := range index {
		// Sanitize the name from index to match the actual filename
		sanitizedName := s.filesystem.SanitizeName(videoIndex.Name)
		videoPath := s.filesystem.GetFilePath(videoIndex.Category, sanitizedName, "yaml")
		fullVideo, err := s.yamlStorage.GetVideo(videoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get video details for %s: %w", videoIndex.Name, err)
		}
		// Always use sanitized name to ensure consistency with filenames
		fullVideo.Name = s.filesystem.SanitizeName(fullVideo.Name)
		fullVideo.Category = videoIndex.Category
		fullVideo.Path = videoPath

		// Use CalculateVideoPhase since we already have the full video data loaded
		// This avoids the file I/O overhead of GetVideoPhase which would reload the video
		currentPhase := video.CalculateVideoPhase(fullVideo)
		if currentPhase == phase {
			videosInPhase = append(videosInPhase, fullVideo)
		}
	}

	// Apply appropriate ordering based on phase
	if phase == workflow.PhaseIdeas {
		// Randomize videos in Ideas phase (phase 7)
		rand.Shuffle(len(videosInPhase), func(i, j int) {
			videosInPhase[i], videosInPhase[j] = videosInPhase[j], videosInPhase[i]
		})
	} else {
		// Sort videos by date for all other phases (maintain existing behavior)
		sort.Slice(videosInPhase, func(i, j int) bool {
			date1, _ := time.Parse("2006-01-02T15:04", videosInPhase[i].Date)
			date2, _ := time.Parse("2006-01-02T15:04", videosInPhase[j].Date)
			return date1.Before(date2)
		})
	}

	return videosInPhase, nil
}

// GetAllVideos returns all videos from all phases (0-7)
// This method is used when the phase parameter is omitted from API requests
func (s *VideoService) GetAllVideos() ([]storage.Video, error) {
	index, err := s.yamlStorage.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get video index: %w", err)
	}

	var allVideos []storage.Video
	for _, video := range index {
		// Sanitize the name from index to match the actual filename
		sanitizedName := s.filesystem.SanitizeName(video.Name)
		videoPath := s.filesystem.GetFilePath(video.Category, sanitizedName, "yaml")
		fullVideo, err := s.yamlStorage.GetVideo(videoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get video details for %s: %w", video.Name, err)
		}
		// Always use sanitized name to ensure consistency with filenames
		fullVideo.Name = s.filesystem.SanitizeName(fullVideo.Name)
		fullVideo.Category = video.Category
		fullVideo.Path = videoPath
		allVideos = append(allVideos, fullVideo)
	}

	// Sort videos by date
	sort.Slice(allVideos, func(i, j int) bool {
		date1, _ := time.Parse("2006-01-02T15:04", allVideos[i].Date)
		date2, _ := time.Parse("2006-01-02T15:04", allVideos[j].Date)
		return date1.Before(date2)
	})

	return allVideos, nil
}

// GetVideoPhases returns the count of videos in each phase
func (s *VideoService) GetVideoPhases() (map[int]int, error) {
	index, err := s.yamlStorage.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get video index: %w", err)
	}

	phases := map[int]int{
		workflow.PhaseIdeas:            0,
		workflow.PhaseStarted:          0,
		workflow.PhaseMaterialDone:     0,
		workflow.PhaseEditRequested:    0,
		workflow.PhasePublishPending:   0,
		workflow.PhasePublished:        0,
		workflow.PhaseDelayed:          0,
		workflow.PhaseSponsoredBlocked: 0,
	}

	for _, videoIndex := range index {
		// Load video data once and use CalculateVideoPhase directly
		// This avoids the double file I/O that GetVideoPhase does (load + reload)
		// Sanitize the name from index to match the actual filename
		sanitizedName := s.filesystem.SanitizeName(videoIndex.Name)
		videoPath := s.filesystem.GetFilePath(videoIndex.Category, sanitizedName, "yaml")
		fullVideo, err := s.yamlStorage.GetVideo(videoPath)
		if err != nil {
			// Log error but continue counting other videos
			continue
		}

		currentPhase := video.CalculateVideoPhase(fullVideo)
		phases[currentPhase]++
	}

	return phases, nil
}

// GetVideo retrieves a specific video by name and category
func (s *VideoService) GetVideo(name, category string) (storage.Video, error) {
	if name == "" || category == "" {
		return storage.Video{}, fmt.Errorf("name and category are required")
	}

	// Sanitize the name to match the actual filename
	sanitizedName := s.filesystem.SanitizeName(name)
	videoPath := s.filesystem.GetFilePath(category, sanitizedName, "yaml")
	video, err := s.yamlStorage.GetVideo(videoPath)
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to get video %s: %w", name, err)
	}

	// Always use the sanitized filename to ensure consistency
	// The filename is the source of truth, not the YAML content
	video.Name = s.filesystem.SanitizeName(video.Name)
	video.Category = category
	video.Path = videoPath

	return video, nil
}

// UpdateVideo updates a video's data
func (s *VideoService) UpdateVideo(video storage.Video) error {
	if video.Path == "" {
		return fmt.Errorf("video path is required")
	}

	return s.yamlStorage.WriteVideo(video, video.Path)
}

// DeleteVideo deletes a video and its associated files
func (s *VideoService) DeleteVideo(name, category string) error {
	if name == "" || category == "" {
		return fmt.Errorf("name and category are required")
	}

	// Sanitize the name to match the actual filename
	sanitizedName := s.filesystem.SanitizeName(name)
	videoPath := s.filesystem.GetFilePath(category, sanitizedName, "yaml")
	mdPath := s.filesystem.GetFilePath(category, sanitizedName, "md")

	// Delete both files
	yamlErr := os.Remove(videoPath)
	mdErr := os.Remove(mdPath)

	var deletionErrors []string
	if yamlErr != nil && !os.IsNotExist(yamlErr) {
		deletionErrors = append(deletionErrors, fmt.Sprintf("YAML file (%s): %v", videoPath, yamlErr))
	}
	if mdErr != nil && !os.IsNotExist(mdErr) {
		deletionErrors = append(deletionErrors, fmt.Sprintf("MD file (%s): %v", mdPath, mdErr))
	}

	if len(deletionErrors) > 0 {
		return fmt.Errorf("errors during file deletion: %s", strings.Join(deletionErrors, "; "))
	}

	// Remove from index
	index, err := s.yamlStorage.GetIndex()
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	var updatedIndex []storage.VideoIndex
	for _, vi := range index {
		if !(vi.Name == name && vi.Category == category) {
			updatedIndex = append(updatedIndex, vi)
		}
	}

	return s.yamlStorage.WriteIndex(updatedIndex)
}

// GetCategories returns available video categories
func (s *VideoService) GetCategories() ([]Category, error) {
	var availableDirs []Category
	manuscriptPath := "manuscript"

	files, err := os.ReadDir(manuscriptPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []Category{}, nil
		}
		return nil, fmt.Errorf("failed to read manuscript directory '%s': %w", manuscriptPath, err)
	}

	caser := cases.Title(language.AmericanEnglish)
	for _, file := range files {
		if file.IsDir() {
			displayName := caser.String(strings.ReplaceAll(file.Name(), "-", " "))
			dirPath := filepath.Join(manuscriptPath, file.Name())
			availableDirs = append(availableDirs, Category{Name: displayName, Path: dirPath})
		}
	}

	sort.Slice(availableDirs, func(i, j int) bool {
		return availableDirs[i].Name < availableDirs[j].Name
	})

	return availableDirs, nil
}

// MoveVideo moves video files to a new directory using the robust utils.MoveVideoFiles
func (s *VideoService) MoveVideo(name, category, targetDir string) error {
	if name == "" || category == "" || targetDir == "" {
		return fmt.Errorf("name, category, and target directory are required")
	}

	// Sanitize the name to match the actual filename
	sanitizedName := s.filesystem.SanitizeName(name)
	currentYAMLPath := s.filesystem.GetFilePath(category, sanitizedName, "yaml")
	currentMDPath := s.filesystem.GetFilePath(category, sanitizedName, "md")

	// Use the robust file moving utility instead of the simpler implementation
	_, _, err := utils.MoveVideoFiles(currentYAMLPath, currentMDPath, targetDir, name)
	if err != nil {
		return fmt.Errorf("failed to move video files: %w", err)
	}

	// Update index
	index, err := s.yamlStorage.GetIndex()
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	for i, vi := range index {
		if vi.Name == name && vi.Category == category {
			index[i].Category = filepath.Base(targetDir)
			break
		}
	}

	if err := s.yamlStorage.WriteIndex(index); err != nil {
		return fmt.Errorf("failed to update index: %w", err)
	}

	return nil
}

// UpdateVideoPhase updates a video with phase-specific changes and recalculates completion.
// It now takes a pointer to a storage.Video object directly.
func (s *VideoService) UpdateVideoPhase(video *storage.Video, phase string, updateData map[string]interface{}) (*storage.Video, error) {
	if video == nil {
		return nil, fmt.Errorf("video to update cannot be nil")
	}

	// Apply the updates based on the phase
	if err := s.applyPhaseUpdates(video, phase, updateData); err != nil {
		return nil, fmt.Errorf("failed to apply phase updates: %w", err)
	}

	// Save the updated video
	if err := s.UpdateVideo(*video); err != nil {
		return nil, fmt.Errorf("failed to save video: %w", err)
	}

	// No longer update completion counts - using real-time calculations from video manager
	return video, nil
}

// Category represents a video category
type Category struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// applyPhaseUpdates applies updates to a video based on its phase
func (s *VideoService) applyPhaseUpdates(video *storage.Video, phase string, updateData map[string]interface{}) error {
	// Validate phase
	validPhases := map[string]bool{
		"initial-details": true,
		"work-progress":   true,
		"definition":      true,
		"post-production": true,
		"publishing":      true,
		"post-publish":    true,
	}

	if !validPhases[phase] {
		return fmt.Errorf("unknown phase: %s", phase)
	}

	// Use generic field mapping for all phases
	return s.updateVideoFields(video, updateData)
}

// updateVideoFields uses reflection to map JSON field names to struct fields
// This eliminates the need for hard-coded field mappings and prevents field name mismatches
func (s *VideoService) updateVideoFields(video *storage.Video, updateData map[string]interface{}) error {
	if video == nil {
		return fmt.Errorf("video cannot be nil")
	}

	videoValue := reflect.ValueOf(video).Elem() // Get the underlying value that the pointer points to
	videoType := videoValue.Type()

	// Create a map of JSON field names to struct field indices for fast lookup
	jsonToFieldMap := make(map[string]int)
	for i := 0; i < videoType.NumField(); i++ {
		field := videoType.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Handle json tags like "fieldName,omitempty"
			jsonFieldName := strings.Split(jsonTag, ",")[0]
			jsonToFieldMap[jsonFieldName] = i
		}
	}

	// Apply updates from the updateData map
	for jsonFieldName, newValue := range updateData {
		fieldIndex, exists := jsonToFieldMap[jsonFieldName]
		if !exists {
			// Handle nested fields (like sponsorship.amount -> sponsorshipAmount)
			if err := s.updateNestedField(video, jsonFieldName, newValue); err != nil {
				// If it's not a nested field either, skip it (could be a frontend-only field)
				continue
			}
			continue
		}

		fieldValue := videoValue.Field(fieldIndex)
		if !fieldValue.CanSet() {
			continue // Skip unexported fields
		}

		// Type-safe assignment based on the field's actual type
		if err := s.setFieldValue(fieldValue, newValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", jsonFieldName, err)
		}
	}

	return nil
}

// updateNestedField handles special cases like sponsorship fields and field name mappings
func (s *VideoService) updateNestedField(video *storage.Video, fieldName string, newValue interface{}) error {
	// Handle sponsorship fields
	switch fieldName {
	case "sponsorshipAmount":
		if str, ok := newValue.(string); ok {
			video.Sponsorship.Amount = str
			return nil
		}
	case "sponsorshipEmails":
		if str, ok := newValue.(string); ok {
			video.Sponsorship.Emails = str
			return nil
		}
	case "sponsorshipBlockedReason":
		if str, ok := newValue.(string); ok {
			video.Sponsorship.Blocked = str
			return nil
		}

	// Handle field name mappings that don't match JSON tags
	case "codeDone":
		if b, ok := newValue.(bool); ok {
			video.Code = b
			return nil
		}
	case "talkingHeadDone":
		if b, ok := newValue.(bool); ok {
			video.Head = b
			return nil
		}
	case "screenRecordingDone":
		if b, ok := newValue.(bool); ok {
			video.Screen = b
			return nil
		}
	case "thumbnailsDone":
		if b, ok := newValue.(bool); ok {
			video.Thumbnails = b
			return nil
		}
	case "diagramsDone":
		if b, ok := newValue.(bool); ok {
			video.Diagrams = b
			return nil
		}
	case "screenshotsDone":
		if b, ok := newValue.(bool); ok {
			video.Screenshots = b
			return nil
		}
	case "filesLocation":
		if str, ok := newValue.(string); ok {
			video.Location = str
			return nil
		}
	case "otherLogosAssets":
		if str, ok := newValue.(string); ok {
			video.OtherLogos = str
			return nil
		}
	case "tweetText":
		if str, ok := newValue.(string); ok {
			video.Tweet = str
			return nil
		}
	case "animationsScript":
		if str, ok := newValue.(string); ok {
			video.Animations = str
			return nil
		}
	case "requestThumbnailGeneration":
		if b, ok := newValue.(bool); ok {
			video.RequestThumbnail = b
			return nil
		}

	case "thumbnailPath":
		if str, ok := newValue.(string); ok {
			video.Thumbnail = str
			return nil
		}
	case "movieDone":
		if b, ok := newValue.(bool); ok {
			video.Movie = b
			return nil
		}
	case "slidesDone":
		if b, ok := newValue.(bool); ok {
			video.Slides = b
			return nil
		}
	case "videoFilePath":
		if str, ok := newValue.(string); ok {
			video.UploadVideo = str
			return nil
		}

	// Handle special publishing actions
	case "uploadToYouTube":
		if b, ok := newValue.(bool); ok && b {
			// TODO: Implement YouTube upload logic
			// For now, just mark as completed
			video.VideoId = "placeholder-youtube-id"
			return nil
		}
	case "createHugoPost":
		if b, ok := newValue.(bool); ok && b {
			// TODO: Implement Hugo post creation logic
			// For now, just mark as completed
			video.HugoPath = "placeholder-hugo-path"
			return nil
		}

	// Handle post-publish field mappings
	case "dotPosted":
		if b, ok := newValue.(bool); ok {
			video.DOTPosted = b
			return nil
		}
	case "blueSkyPostSent":
		if b, ok := newValue.(bool); ok {
			video.BlueSkyPosted = b
			return nil
		}
	case "linkedInPostSent":
		if b, ok := newValue.(bool); ok {
			video.LinkedInPosted = b
			return nil
		}
	case "slackPostSent":
		if b, ok := newValue.(bool); ok {
			video.SlackPosted = b
			return nil
		}
	case "youTubeHighlightCreated":
		if b, ok := newValue.(bool); ok {
			video.YouTubeHighlight = b
			return nil
		}
	case "youTubePinnedCommentAdded":
		if b, ok := newValue.(bool); ok {
			video.YouTubeComment = b
			return nil
		}
	case "repliedToYouTubeComments":
		if b, ok := newValue.(bool); ok {
			video.YouTubeCommentReply = b
			return nil
		}
	case "gdeAdvocuPostSent":
		if b, ok := newValue.(bool); ok {
			video.GDE = b
			return nil
		}
	case "codeRepositoryURL":
		if str, ok := newValue.(string); ok {
			video.Repo = str
			return nil
		}
	}

	return fmt.Errorf("unknown nested field: %s", fieldName)
}

// setFieldValue safely sets a field value with proper type conversion
func (s *VideoService) setFieldValue(fieldValue reflect.Value, newValue interface{}) error {
	if newValue == nil {
		return nil // Skip nil values
	}

	fieldType := fieldValue.Type()
	newValueType := reflect.TypeOf(newValue)

	// Handle type conversions
	switch fieldType.Kind() {
	case reflect.String:
		if str, ok := newValue.(string); ok {
			fieldValue.SetString(str)
		} else {
			return fmt.Errorf("expected string, got %v", newValueType)
		}
	case reflect.Bool:
		if b, ok := newValue.(bool); ok {
			fieldValue.SetBool(b)
		} else {
			return fmt.Errorf("expected bool, got %v", newValueType)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, ok := newValue.(int64); ok {
			fieldValue.SetInt(i)
		} else if i, ok := newValue.(int); ok {
			fieldValue.SetInt(int64(i))
		} else if f, ok := newValue.(float64); ok {
			fieldValue.SetInt(int64(f))
		} else {
			return fmt.Errorf("expected integer, got %v", newValueType)
		}
	case reflect.Float32, reflect.Float64:
		if f, ok := newValue.(float64); ok {
			fieldValue.SetFloat(f)
		} else if i, ok := newValue.(int); ok {
			fieldValue.SetFloat(float64(i))
		} else {
			return fmt.Errorf("expected float, got %v", newValueType)
		}
	default:
		return fmt.Errorf("unsupported field type: %v", fieldType)
	}

	return nil
}

// GetVideoManuscript retrieves and reads the manuscript content for a video
// This extracts the common pattern used throughout the CLI for reading manuscript files
func (s *VideoService) GetVideoManuscript(name, category string) (string, error) {
	// First get the video to access its Gist field
	video, err := s.GetVideo(name, category)
	if err != nil {
		return "", err
	}

	// Check if Gist field is empty
	if video.Gist == "" {
		return "", fmt.Errorf("gist field is empty for video %s in category %s", name, category)
	}

	// Read the manuscript file
	manuscriptContent, readErr := os.ReadFile(video.Gist)
	if readErr != nil {
		return "", fmt.Errorf("failed to read manuscript file %s: %w", video.Gist, readErr)
	}

	return string(manuscriptContent), nil
}

// GetManuscriptPath returns the expected path for a video's manuscript file
// This is useful for tests and other scenarios where you need to know the path without reading the content
func (s *VideoService) GetManuscriptPath(name, category string) string {
	sanitizedName := s.filesystem.SanitizeName(name)
	return s.filesystem.GetFilePath(category, sanitizedName, "md")
}
