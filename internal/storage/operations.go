package storage

import (
	"fmt"
	"os"
	"strings"
)

// Operations defines the interface for storage operations
type Operations interface {
	GetVideo(path string) (Video, error)
	WriteVideo(video Video, path string) error
	GetIndex() ([]VideoIndex, error)
	WriteIndex(videos []VideoIndex) error
	GetVideosByPhase(phaseID int) ([]Video, error)
	GetVideoPath(name, category string) string
}

// YAMLOperations implements the Operations interface using YAML storage
type YAMLOperations struct {
	yaml      YAML
	indexPath string
}

// NewOperations creates a new YAMLOperations instance
func NewOperations(indexPath string) *YAMLOperations {
	return &YAMLOperations{
		yaml:      YAML{IndexPath: indexPath},
		indexPath: indexPath,
	}
}

// GetVideo retrieves a video from a YAML file
func (o *YAMLOperations) GetVideo(path string) (Video, error) {
	return o.yaml.GetVideo(path)
}

// WriteVideo writes a video to a YAML file
func (o *YAMLOperations) WriteVideo(video Video, path string) error {
	return o.yaml.WriteVideo(video, path)
}

// GetIndex retrieves the video index
func (o *YAMLOperations) GetIndex() ([]VideoIndex, error) {
	return o.yaml.GetIndex()
}

// WriteIndex writes the video index
func (o *YAMLOperations) WriteIndex(videos []VideoIndex) error {
	return o.yaml.WriteIndex(videos)
}

// GetVideoPath generates a path for a video file based on name and category
func (o *YAMLOperations) GetVideoPath(name, category string) string {
	// Convert to lower case and replace spaces with hyphens for both parts
	sanitizedCategory := strings.ReplaceAll(strings.ToLower(category), " ", "-")
	sanitizedName := strings.ReplaceAll(strings.ToLower(name), " ", "-")
	
	// Sanitize the name for file system use
	sanitizedName = sanitizeFileName(sanitizedName)
	
	return fmt.Sprintf("manuscript/%s/%s.yaml", sanitizedCategory, sanitizedName)
}

// GetVideosByPhase returns videos filtered by their phase
func (o *YAMLOperations) GetVideosByPhase(phaseID int) ([]Video, error) {
	index, err := o.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	var videos []Video
	for _, videoIdx := range index {
		videoPath := o.GetVideoPath(videoIdx.Name, videoIdx.Category)
		video, err := o.GetVideo(videoPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Skip videos that don't exist
			}
			return nil, fmt.Errorf("failed to read video %s: %w", videoPath, err)
		}
		
		// Check if the video belongs to the requested phase
		if getVideoPhaseID(video) == phaseID {
			videos = append(videos, video)
		}
	}
	
	return videos, nil
}

// Helper function to determine which phase a video is in
func getVideoPhaseID(video Video) int {
	if !video.Init.Done {
		return 0 // editPhaseInitial
	} else if !video.Work.Done {
		return 1 // editPhaseWork
	} else if !video.Define.Done {
		return 2 // editPhaseDefinition
	} else if !video.Edit.Done {
		return 3 // editPhasePostProduction
	} else if !video.Publish.Done {
		return 4 // editPhasePublishing
	} else {
		return 5 // editPhasePostPublish
	}
}

// sanitizeFileName removes or replaces characters that are typically invalid in file names.
func sanitizeFileName(name string) string {
	name = strings.ReplaceAll(name, ":", "-")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-") // For Windows paths if ever relevant
	name = strings.ReplaceAll(name, "?", "")
	name = strings.ReplaceAll(name, "*", "")
	name = strings.ReplaceAll(name, "<", "")
	name = strings.ReplaceAll(name, ">", "")
	name = strings.ReplaceAll(name, "|", "")
	name = strings.ReplaceAll(name, "\"", "")
	
	// Remove any resulting double hyphens
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	
	return name
}