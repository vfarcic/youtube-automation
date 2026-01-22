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
	"gopkg.in/yaml.v3"
)

// VideoService provides data operations for videos in CLI
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

// ArchiveVideo moves a video from index.yaml to index/[YEAR].yaml
// The year is extracted from the provided date string (format: "2006-01-02T15:04")
func (s *VideoService) ArchiveVideo(name, category, date string) error {
	if name == "" || category == "" {
		return fmt.Errorf("name and category are required")
	}

	// Extract year from date (format: "2006-01-02T15:04")
	year := extractYearFromDate(date)
	if year == "" {
		return fmt.Errorf("video has no valid date, cannot archive")
	}

	// Ensure index/ directory exists
	indexDir := "index"
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return fmt.Errorf("failed to create index directory: %w", err)
	}

	// Get or create archive index file path
	archiveIndexPath := filepath.Join(indexDir, year+".yaml")

	// Read existing archive index (or create empty if doesn't exist)
	archivedIndex, err := s.readArchiveIndex(archiveIndexPath)
	if err != nil {
		return fmt.Errorf("failed to read archive index: %w", err)
	}

	// Add video to archived index
	archivedIndex = append(archivedIndex, storage.VideoIndex{
		Name:     name,
		Category: category,
	})

	// Write archived index
	if err := s.writeArchiveIndex(archiveIndexPath, archivedIndex); err != nil {
		return fmt.Errorf("failed to write archive index: %w", err)
	}

	// Remove from main index.yaml
	index, err := s.yamlStorage.GetIndex()
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}

	// Sanitize the name for comparison since index may have unsanitized legacy names
	sanitizedName := s.filesystem.SanitizeName(name)
	var updatedIndex []storage.VideoIndex
	for _, vi := range index {
		if !(s.filesystem.SanitizeName(vi.Name) == sanitizedName && vi.Category == category) {
			updatedIndex = append(updatedIndex, vi)
		}
	}

	return s.yamlStorage.WriteIndex(updatedIndex)
}

// extractYearFromDate extracts the year from a date string in format "2006-01-02T15:04"
func extractYearFromDate(dateStr string) string {
	if len(dateStr) < 4 {
		return ""
	}
	return dateStr[:4]
}

// readArchiveIndex reads an archive index file, returning empty slice if file doesn't exist
func (s *VideoService) readArchiveIndex(path string) ([]storage.VideoIndex, error) {
	var index []storage.VideoIndex
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return index, nil // New file, return empty slice
		}
		return nil, fmt.Errorf("failed to read archive index file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal archive index from %s: %w", path, err)
	}

	return index, nil
}

// writeArchiveIndex writes an archive index to a file
func (s *VideoService) writeArchiveIndex(path string, vi []storage.VideoIndex) error {
	data, err := yaml.Marshal(&vi)
	if err != nil {
		return fmt.Errorf("failed to marshal archive index: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write archive index to file %s: %w", path, err)
	}
	return nil
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

// Category represents a video category
type Category struct {
	Name string `json:"name"`
	Path string `json:"path"`
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
