package data

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/video"
	"devopstoolkit/youtube-automation/internal/workflow"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// VideoService provides data operations for videos
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

	// Create markdown script file
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
		// Load video data once and use CalculateVideoPhase directly
		// This avoids the double file I/O that GetVideoPhase does (load + reload)
		videoPath := s.filesystem.GetFilePath(videoIndex.Category, videoIndex.Name, "yaml")
		fullVideo, err := s.yamlStorage.GetVideo(videoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get video details for %s: %w", videoIndex.Name, err)
		}
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

	// Sort videos by date
	sort.Slice(videosInPhase, func(i, j int) bool {
		date1, _ := time.Parse("2006-01-02T15:04", videosInPhase[i].Date)
		date2, _ := time.Parse("2006-01-02T15:04", videosInPhase[j].Date)
		return date1.Before(date2)
	})

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

	video.Name = name
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

// MoveVideo moves video files to a new directory
func (s *VideoService) MoveVideo(name, category, targetDir string) error {
	if name == "" || category == "" || targetDir == "" {
		return fmt.Errorf("name, category, and target directory are required")
	}

	currentYAMLPath := s.filesystem.GetFilePath(category, name, "yaml")
	currentMDPath := s.filesystem.GetFilePath(category, name, "md")

	// Move files using utility function
	_, _, err := s.moveVideoFiles(currentYAMLPath, currentMDPath, targetDir, name)
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

// moveVideoFiles is a helper function to move video files
func (s *VideoService) moveVideoFiles(yamlPath, mdPath, targetDir, baseName string) (string, string, error) {
	// Create target directory if it doesn't exist
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	// Determine new file paths
	newYAMLPath := filepath.Join(targetDir, baseName+".yaml")
	newMDPath := filepath.Join(targetDir, baseName+".md")

	// Move YAML file
	if err := os.Rename(yamlPath, newYAMLPath); err != nil {
		return "", "", fmt.Errorf("failed to move YAML file: %w", err)
	}

	// Move MD file
	if err := os.Rename(mdPath, newMDPath); err != nil {
		// Try to rollback YAML file move
		os.Rename(newYAMLPath, yamlPath)
		return "", "", fmt.Errorf("failed to move MD file: %w", err)
	}

	return newYAMLPath, newMDPath, nil
}
