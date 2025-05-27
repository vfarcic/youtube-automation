package service

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"devopstoolkit/youtube-automation/internal/storage"
	"devopstoolkit/youtube-automation/internal/workflow"
)

var (
	ErrVideoNotFound = errors.New("video not found")
	ErrInvalidRequest = errors.New("invalid request")
)

// VideoService handles business logic for video operations
type VideoService struct {
	storageOps storage.Operations
}

// VideoPhase represents a phase in the video lifecycle
type VideoPhase struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
	ID    int    `json:"id"`
}

// VideoCreateRequest represents the data needed to create a new video
type VideoCreateRequest struct {
	Name     string `json:"name"`
	Category string `json:"category"`
}

// NewVideoService creates a new VideoService instance
func NewVideoService(storageOps storage.Operations) *VideoService {
	return &VideoService{
		storageOps: storageOps,
	}
}

// GetVideoPhases returns available video phases and counts of videos in each
func (s *VideoService) GetVideoPhases() ([]VideoPhase, error) {
	// Get all videos
	index, err := s.storageOps.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}

	// Define phases using the shared workflow constants
	phases := []VideoPhase{
		{Name: workflow.PhaseNames[workflow.PhasePublished], ID: workflow.PhasePublished, Count: 0},
		{Name: workflow.PhaseNames[workflow.PhasePublishPending], ID: workflow.PhasePublishPending, Count: 0},
		{Name: workflow.PhaseNames[workflow.PhaseEditRequested], ID: workflow.PhaseEditRequested, Count: 0},
		{Name: workflow.PhaseNames[workflow.PhaseMaterialDone], ID: workflow.PhaseMaterialDone, Count: 0},
		{Name: workflow.PhaseNames[workflow.PhaseStarted], ID: workflow.PhaseStarted, Count: 0},
		{Name: workflow.PhaseNames[workflow.PhaseDelayed], ID: workflow.PhaseDelayed, Count: 0},
		{Name: workflow.PhaseNames[workflow.PhaseSponsoredBlocked], ID: workflow.PhaseSponsoredBlocked, Count: 0},
		{Name: workflow.PhaseNames[workflow.PhaseIdeas], ID: workflow.PhaseIdeas, Count: 0},
	}

	// Count videos in each phase
	for _, videoIdx := range index {
		videoPath := s.storageOps.GetVideoPath(videoIdx.Name, videoIdx.Category)
		video, err := s.storageOps.GetVideo(videoPath)
		if err != nil {
			// Skip videos that can't be read
			continue
		}

		// Determine which phase the video is in
		phaseID := getVideoPhaseID(video)
		phases[phaseID].Count++
	}

	return phases, nil
}

// GetVideosByPhase returns videos in a specific phase
func (s *VideoService) GetVideosByPhase(phaseID int) ([]storage.Video, error) {
	return s.storageOps.GetVideosByPhase(phaseID)
}

// GetVideo returns a specific video by ID
func (s *VideoService) GetVideo(videoID string) (storage.Video, error) {
	// Parse videoID which should be in the format "name,category"
	parts := strings.Split(videoID, ",")
	if len(parts) != 2 {
		return storage.Video{}, ErrInvalidRequest
	}
	
	name := parts[0]
	category := parts[1]
	
	videoPath := s.storageOps.GetVideoPath(name, category)
	video, err := s.storageOps.GetVideo(videoPath)
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to get video: %w", err)
	}
	
	return video, nil
}

// CreateVideo creates a new video
func (s *VideoService) CreateVideo(req VideoCreateRequest) (storage.Video, error) {
	// Validate request
	if req.Name == "" || req.Category == "" {
		return storage.Video{}, ErrInvalidRequest
	}
	
	// Create a new video
	newVideo := storage.Video{
		Name:     req.Name,
		Category: req.Category,
		Date:     time.Now().Format("2006-01-02"),
	}
	
	// Get the path where the video will be saved
	videoPath := s.storageOps.GetVideoPath(req.Name, req.Category)
	
	// Ensure the directory exists
	dir := strings.Split(videoPath, "/")
	if len(dir) > 1 {
		dirPath := strings.Join(dir[:len(dir)-1], "/")
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return storage.Video{}, fmt.Errorf("failed to create directory for video: %w", err)
		}
	}
	
	// Save the video
	err := s.storageOps.WriteVideo(newVideo, videoPath)
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to write video: %w", err)
	}
	
	// Update the index
	index, err := s.storageOps.GetIndex()
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to get index: %w", err)
	}
	
	index = append(index, storage.VideoIndex{
		Name:     req.Name,
		Category: req.Category,
	})
	
	err = s.storageOps.WriteIndex(index)
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to write index: %w", err)
	}
	
	return newVideo, nil
}

// UpdateVideo updates an existing video
func (s *VideoService) UpdateVideo(videoID string, updatedVideo storage.Video) (storage.Video, error) {
	// Parse videoID which should be in the format "name,category"
	parts := strings.Split(videoID, ",")
	if len(parts) != 2 {
		return storage.Video{}, ErrInvalidRequest
	}
	
	name := parts[0]
	category := parts[1]
	
	// Get the path where the video is saved
	videoPath := s.storageOps.GetVideoPath(name, category)
	
	// Check if the video exists
	_, err := s.storageOps.GetVideo(videoPath)
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to get video: %w", err)
	}
	
	// Update the video's path to ensure it's preserved
	updatedVideo.Path = videoPath
	
	// Save the updated video
	err = s.storageOps.WriteVideo(updatedVideo, videoPath)
	if err != nil {
		return storage.Video{}, fmt.Errorf("failed to write video: %w", err)
	}
	
	return updatedVideo, nil
}

// DeleteVideo deletes a video
func (s *VideoService) DeleteVideo(videoID string) error {
	// Parse videoID which should be in the format "name,category"
	parts := strings.Split(videoID, ",")
	if len(parts) != 2 {
		return ErrInvalidRequest
	}
	
	name := parts[0]
	category := parts[1]
	
	// Get the path where the video is saved
	videoPath := s.storageOps.GetVideoPath(name, category)
	
	// Delete the video file
	err := deleteFile(videoPath)
	if err != nil {
		return fmt.Errorf("failed to delete video file: %w", err)
	}
	
	// Update the index
	index, err := s.storageOps.GetIndex()
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}
	
	// Remove the video from the index
	var updatedIndex []storage.VideoIndex
	for _, video := range index {
		if video.Name != name || video.Category != category {
			updatedIndex = append(updatedIndex, video)
		}
	}
	
	err = s.storageOps.WriteIndex(updatedIndex)
	if err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}
	
	return nil
}

// GetCategories returns available video categories
func (s *VideoService) GetCategories() ([]string, error) {
	// This would typically access the filesystem to get manuscript directories
	// Since we don't have direct file system access here, we'll need to get unique categories from the index
	index, err := s.storageOps.GetIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get index: %w", err)
	}
	
	// Create a map to store unique categories
	categoriesMap := make(map[string]bool)
	for _, video := range index {
		categoriesMap[video.Category] = true
	}
	
	// Convert map keys to a slice
	categories := make([]string, 0, len(categoriesMap))
	for category := range categoriesMap {
		categories = append(categories, category)
	}
	
	return categories, nil
}

// Helper function to determine which phase a video is in, matching logic in video/manager.go
func getVideoPhaseID(video storage.Video) int {
	if len(video.Sponsorship.Blocked) > 0 { // Check for sponsorship block first
		return workflow.PhaseSponsoredBlocked
	} else if video.Delayed { // Then check for delayed
		return workflow.PhaseDelayed
	} else if len(video.Repo) > 0 { // Assuming video.Repo is populated when published
		return workflow.PhasePublished
	} else if len(video.UploadVideo) > 0 && len(video.Tweet) > 0 { // Assuming these indicate pending publish
		return workflow.PhasePublishPending
	} else if video.RequestEdit {
		return workflow.PhaseEditRequested
	} else if video.Code && video.Screen && video.Head && video.Diagrams { // Assuming these are key for material done
		return workflow.PhaseMaterialDone
	} else if len(video.Date) > 0 { // Date implies started
		return workflow.PhaseStarted
	} else {
		return workflow.PhaseIdeas
	}
}

// Helper function to delete a file
func deleteFile(path string) error {
	return os.Remove(path)
}