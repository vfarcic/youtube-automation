package service

import (
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
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
func (s *VideoService) CreateVideo(name, category string) (storage.VideoIndex, error) {
	if name == "" || category == "" {
		return storage.VideoIndex{}, fmt.Errorf("name and category are required")
	}

	vi := storage.VideoIndex{
		Name:     name,
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
		Name:     name,
		Category: category,
		Path:     videoPath,
		// Initialize sponsorship
		Sponsorship: storage.Sponsorship{
			Amount:  "",
			Emails:  "",
			Blocked: "",
		},
		// Initialize other fields with default values
		Date:                "",
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
	for i, videoIndex := range index {
		videoPath := s.filesystem.GetFilePath(videoIndex.Category, videoIndex.Name, "yaml")
		fullVideo, err := s.yamlStorage.GetVideo(videoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get video details for %s: %w", videoIndex.Name, err)
		}
		fullVideo.Index = i
		fullVideo.Name = videoIndex.Name
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
	for i, video := range index {
		videoPath := s.filesystem.GetFilePath(video.Category, video.Name, "yaml")
		fullVideo, err := s.yamlStorage.GetVideo(videoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get video details for %s: %w", video.Name, err)
		}
		fullVideo.Index = i
		fullVideo.Name = video.Name
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
		videoPath := s.filesystem.GetFilePath(videoIndex.Category, videoIndex.Name, "yaml")
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

	videoPath := s.filesystem.GetFilePath(category, name, "yaml")
	video, err := s.yamlStorage.GetVideo(videoPath)
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to get video %s: %w", name, err)
	}

	if video.Name == "" {
		video.Name = name
	}
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

	videoPath := s.filesystem.GetFilePath(category, name, "yaml")
	mdPath := s.filesystem.GetFilePath(category, name, "md")

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

	currentYAMLPath := s.filesystem.GetFilePath(category, name, "yaml")
	currentMDPath := s.filesystem.GetFilePath(category, name, "md")

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
	switch phase {
	case "initial-details":
		return s.applyInitialDetailsUpdates(video, updateData)
	case "work-progress":
		return s.applyWorkProgressUpdates(video, updateData)
	case "definition":
		return s.applyDefinitionUpdates(video, updateData)
	case "post-production":
		return s.applyPostProductionUpdates(video, updateData)
	case "publishing":
		return s.applyPublishingUpdates(video, updateData)
	case "post-publish":
		return s.applyPostPublishUpdates(video, updateData)
	default:
		return fmt.Errorf("unknown phase: %s", phase)
	}
}

// applyInitialDetailsUpdates applies updates to the initial details phase
func (s *VideoService) applyInitialDetailsUpdates(video *storage.Video, updateData map[string]interface{}) error {
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

	return nil
}

// applyWorkProgressUpdates applies updates to the work progress phase
func (s *VideoService) applyWorkProgressUpdates(video *storage.Video, updateData map[string]interface{}) error {
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

	return nil
}

// applyDefinitionUpdates applies updates to the definition phase
func (s *VideoService) applyDefinitionUpdates(video *storage.Video, updateData map[string]interface{}) error {
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
	if val, ok := updateData["gistPath"]; ok {
		if str, ok := val.(string); ok {
			video.Gist = str
		}
	}

	return nil
}

// applyPostProductionUpdates applies updates to the post-production phase
func (s *VideoService) applyPostProductionUpdates(video *storage.Video, updateData map[string]interface{}) error {
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

	return nil
}

// applyPublishingUpdates applies updates to the publishing phase
func (s *VideoService) applyPublishingUpdates(video *storage.Video, updateData map[string]interface{}) error {
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

	return nil
}

// applyPostPublishUpdates applies updates to the post-publish phase
func (s *VideoService) applyPostPublishUpdates(video *storage.Video, updateData map[string]interface{}) error {
	if val, ok := updateData["dotPosted"]; ok {
		if b, ok := val.(bool); ok {
			video.DOTPosted = b
		}
	}
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

	return nil
}
